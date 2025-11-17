package tools

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Gc2Cmd struct {
	RepoUrl string

	Repo_domain string
	Repo_author string
	Repo_name   string

	DestFullPath string

	Depth  int
	Editor string
}

func ParseCommand(args []string) (*Gc2Cmd, error) {
	c := new(Gc2Cmd)

	for i := 0; i < len(args); i++ {
		arg := args[i]

		{
			if strings.HasPrefix(arg, "-d") {
				val, ok := tryOneOption("-d", arg)
				if !ok {
					if arg != "-d" || i+1 == len(args) {
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
			}

			if strings.HasPrefix(arg, "--depth") {
				val, ok := tryOneOption("--depth", arg)
				if !ok {
					if arg != "--depth" || i+1 == len(args) {
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
			}
		}

		{
			if strings.HasPrefix(arg, "-e") {
				val, ok := tryOneOption("-e", arg)
				if !ok {
					if arg != "-e" || i+1 == len(args) {
						return nil, errNotSupporedOF(arg)
					}

					i++
					val = args[i]
				}

				if !isValidCmdName(val) {
					return nil, fmt.Errorf("invalid editor command name [%s]", val)
				}

				c.Editor = val
			}

			if strings.HasPrefix(arg, "--editor") {
				val, ok := tryOneOption("--editor", arg)
				if !ok {
					if arg != "--editor" || i+1 == len(args) {
						return nil, errNotSupporedOF(arg)
					}

					i++
					val = args[i]
				}

				if !isValidCmdName(val) {
					return nil, fmt.Errorf("invalid editor command name [%s]", val)
				}

				c.Editor = val
			}
		}

		if strings.HasPrefix(arg, "-") {
			return nil, errNotSupporedOF(arg)
		}

		err := parseRepoUrl(arg, c)
		if err != nil {
			return nil, err
		}
	}

	c.DestFullPath = filepath.Join(mustHomeDir(), c.Repo_domain, c.Repo_author, c.Repo_name)

	return c, nil
}

func parseRepoUrl(arg string, c *Gc2Cmd) error {
	c.RepoUrl = arg

	// Support generic Git host patterns:
	// 1) https://gitdomain.com/author/repo(.git)?
	// 2) http://gitdomain.com/author/repo(.git)?
	// 3) git@gitdomain.com:author/repo(.git)?
	// 4) ssh://git@gitdomain.com/author/repo(.git)?

	httpsPattern := regexp.MustCompile(`^https?://([^/]+)/([^/]+)/([^/]+?)(?:\.git)?/?$`)
	httpPattern := regexp.MustCompile(`^http?://([^/]+)/([^/]+)/([^/]+?)(?:\.git)?/?$`)
	sshPatternA := regexp.MustCompile(`^git@([^:]+):([^/]+)/([^/]+?)(?:\.git)?$`)
	sshPatternB := regexp.MustCompile(`^ssh://git@([^/]+)/([^/]+)/([^/]+?)(?:\.git)?/?$`)

	m := httpsPattern.FindStringSubmatch(arg)
	if m == nil {
		m = httpPattern.FindStringSubmatch(arg)
	}
	if m == nil {
		m = sshPatternA.FindStringSubmatch(arg)
	}
	if m == nil {
		m = sshPatternB.FindStringSubmatch(arg)
	}

	if m == nil {
		return fmt.Errorf("unsupported git URL[%s];\n	expected <scheme>://<domain>/<author>/<repo>[.git] or git@<domain>:<author>/<repo>[.git]", arg)
	}

	c.Repo_domain = m[1]
	c.Repo_author = m[2]
	c.Repo_name = strings.TrimSuffix(m[3], ".git")

	return nil
}

// var (
// 	// https://github.com/owner/repo(.git)?
// 	httpsPattern = regexp.MustCompile(`^https?://github\.com/([^/]+)/([^/]+?)(?:\.git)?/?$`)
// 	// git@github.com:owner/repo(.git)?
// 	sshPatternA = regexp.MustCompile(`^git@github\.com:([^/]+)/([^/]+?)(?:\.git)?$`)
// 	// ssh://git@github.com/owner/repo(.git)?
// 	sshPatternB = regexp.MustCompile(`^ssh://git@github\.com/([^/]+)/([^/]+?)(?:\.git)?/?$`)
// )
// func parseGitHubURL(u string) (owner string, repo string, err error) {
// 	if m := httpsPattern.FindStringSubmatch(u); m != nil {
// 		return m[1], trimGitSuffix(m[2]), nil
// 	}
// 	if m := sshPatternA.FindStringSubmatch(u); m != nil {
// 		return m[1], trimGitSuffix(m[2]), nil
// 	}
// 	if m := sshPatternB.FindStringSubmatch(u); m != nil {
// 		return m[1], trimGitSuffix(m[2]), nil
// 	}
// 	return "", "", errors.New("unsupported or non-GitHub URL format; expected github.com URL")
// }

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
