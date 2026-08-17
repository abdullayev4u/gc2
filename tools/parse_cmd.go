package tools

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/abdullayev4u/gc2/config"
)

type Gc2Cmd struct {
	RepoUrl string

	Repo_domain string
	Repo_groups []string
	Repo_name   string

	DestFullPath string

	// DomainFolderPath is the <root>/<domain> directory, or "" when the repo
	// does not live under one. Only the folder icon cares.
	DomainFolderPath string

	// MatchedRule is the mapping rule that decided DestFullPath, or nil when
	// the remote layout was mirrored.
	MatchedRule *Rule

	// Depth and Editor hold the command-line flags only. Cfg holds the
	// effective settings once those flags have been merged over the config
	// file; everything downstream reads Cfg.
	Depth  int
	Editor string

	Cfg config.Resolved
}

// scpPattern matches git's scp-like syntax: [user@]host:path/to/repo.git
var scpPattern = regexp.MustCompile(`^(?:([^@/]+)@)?([^:/]+):(.+)$`)

// ParseCommand reads the flags and repository URL from argv.
func ParseCommand(args []string) (*Gc2Cmd, error) {
	c := new(Gc2Cmd)

	seenUrl := false

	for i := 0; i < len(args); i++ {
		arg := args[i]

		{
			if strings.HasPrefix(arg, "-d") || strings.HasPrefix(arg, "--depth") {
				name := "-d"
				if strings.HasPrefix(arg, "--depth") {
					name = "--depth"
				}

				val, ok := tryOneOption(name, arg)
				if !ok {
					if arg != name || i+1 == len(args) {
						return nil, errNotSupporedOF(arg)
					}

					i++
					val = args[i]
				}

				depth, err := strconv.Atoi(val)
				if err != nil {
					return nil, fmt.Errorf("err parsing option [%s] with value [%s]", arg, val)
				}

				c.Depth = depth
				continue
			}
		}

		{
			if strings.HasPrefix(arg, "-e") || strings.HasPrefix(arg, "--editor") {
				name := "-e"
				if strings.HasPrefix(arg, "--editor") {
					name = "--editor"
				}

				val, ok := tryOneOption(name, arg)
				if !ok {
					if arg != name || i+1 == len(args) {
						return nil, errNotSupporedOF(arg)
					}

					i++
					val = args[i]
				}

				if !isValidCmdName(val) {
					return nil, fmt.Errorf("invalid editor command name [%s]", val)
				}

				c.Editor = val
				continue
			}
		}

		if strings.HasPrefix(arg, "-") {
			return nil, errNotSupporedOF(arg)
		}

		if seenUrl {
			return nil, fmt.Errorf("only one repository URL is supported, got a second one [%s]", arg)
		}

		err := parseRepoUrl(arg, c)
		if err != nil {
			return nil, err
		}

		seenUrl = true
	}

	// Without this the repo fields stay empty and the destination collapses to
	// the home directory itself, which would clone straight into $HOME.
	if !seenUrl {
		return nil, fmt.Errorf("no repository URL given\nusage: gc2 <repo-url>")
	}

	return c, nil
}

// parseRepoUrl splits a clone URL into host, group segments and repository
// name. Unlike a fixed host/author/repo shape, any number of group segments is
// accepted, which is what GitLab subgroups need.
func parseRepoUrl(arg string, c *Gc2Cmd) error {
	c.RepoUrl = arg

	host, path, err := splitRepoUrl(arg)
	if err != nil {
		return err
	}

	segments := splitPath(path)
	if len(segments) < 2 {
		return fmt.Errorf(
			"git URL [%s] has no group segment;\n	expected <scheme>://<domain>/<group>[/<subgroup>...]/<repo>[.git] or git@<domain>:<group>/<repo>[.git]",
			arg,
		)
	}

	c.Repo_domain = host
	c.Repo_groups = segments[:len(segments)-1]
	c.Repo_name = strings.TrimSuffix(segments[len(segments)-1], ".git")

	return nil
}

// splitRepoUrl extracts the bare hostname and path from any git URL form.
func splitRepoUrl(arg string) (host, path string, err error) {
	// Support:
	// 1) https://gitdomain.com/group[/subgroup...]/repo(.git)?
	// 2) http://gitdomain.com/group[/subgroup...]/repo(.git)?
	// 3) ssh://git@gitdomain.com/group[/subgroup...]/repo(.git)?
	// 4) git@gitdomain.com:group[/subgroup...]/repo(.git)?

	if strings.Contains(arg, "://") {
		u, perr := url.Parse(arg)
		if perr != nil {
			return "", "", unsupportedUrl(arg)
		}

		switch u.Scheme {
		case "http", "https", "ssh", "git":
		default:
			return "", "", unsupportedUrl(arg)
		}

		if u.Host == "" || strings.Trim(u.Path, "/") == "" {
			return "", "", unsupportedUrl(arg)
		}

		return bareHost(u.Host), u.Path, nil
	}

	if m := scpPattern.FindStringSubmatch(arg); m != nil {
		if m[2] == "" || strings.Trim(m[3], "/") == "" {
			return "", "", unsupportedUrl(arg)
		}

		return bareHost(m[2]), m[3], nil
	}

	return "", "", unsupportedUrl(arg)
}

// bareHost strips any port from a host so it can be used as a folder name.
// A colon is legal in a macOS path but shows up as "/" in Finder, and is
// outright illegal on Windows.
func bareHost(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil && h != "" {
		host = h
	}

	return strings.ToLower(strings.Trim(host, "[]"))
}

func unsupportedUrl(arg string) error {
	return fmt.Errorf(
		"unsupported git URL[%s];\n	expected <scheme>://<domain>/<group>[/<subgroup>...]/<repo>[.git] or git@<domain>:<group>/<repo>[.git]",
		arg,
	)
}

func tryOneOption(name, arg string) (string, bool) {
	nameEuals := name + "="
	if strings.HasPrefix(arg, nameEuals) {
		val := strings.TrimPrefix(arg, nameEuals)
		if val == "" {
			return "", false
		}
		return val, true
	}

	if strings.HasPrefix(arg, name) {
		val := strings.TrimPrefix(arg, name)
		if val == "" {
			return "", false
		}
		return val, true
	}

	return "", false
}

func errNotSupporedOF(arg string) error {
	return fmt.Errorf("option/flag [%s] is not supported", arg)
}
