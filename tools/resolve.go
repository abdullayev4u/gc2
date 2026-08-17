package tools

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/abdullayev4u/gc2/config"
)

// Rule is a single mapping from a group path pattern to a local folder.
type Rule struct {
	Pattern string
	Target  string

	// literals is the number of non-wildcard segments in the pattern, and
	// wildcards the number of "*" segments. Both feed the precedence order.
	literals  int
	wildcards int

	// greedy is set when the pattern ends in "/*", meaning it also matches
	// every group path nested below it.
	greedy bool

	// segments is the pattern split on "/", with any trailing "*" removed.
	segments []string
}

// Destination is the outcome of resolving a repository URL against a config.
type Destination struct {
	// Path is the directory the repository should be cloned into.
	Path string

	// DomainFolder is the <root>/<host> directory, or "" when the repo does
	// not live under one — either because domainFolder is off or because a
	// rule pointed at an absolute location. Only used for the folder icon.
	DomainFolder string

	// Rule is the mapping rule that matched, or nil when the group path was
	// mirrored verbatim.
	Rule *Rule
}

// Resolve decides where a repository lands on disk.
//
// home is passed in rather than read from the environment so this stays a pure
// function and can be table-tested.
func Resolve(host string, groups []string, repo string, cfg config.Resolved, home string) Destination {
	root := expandTilde(cfg.Root, home)
	if root == "" {
		root = home
	}

	base := root
	domainFolder := ""
	if cfg.DomainFolder && host != "" {
		domainFolder = filepath.Join(root, host)
		base = domainFolder
	}

	// Rules on the host's own section take precedence over the shared ones.
	rule, ok := matchRule(groups, cfg.HostPaths)
	if !ok {
		rule, ok = matchRule(groups, cfg.GlobalPaths)
	}

	if !ok {
		// No rule: mirror the remote layout so the path stays unique and
		// round-trips back to the URL it came from.
		parts := append([]string{base}, groups...)
		return Destination{
			Path:         filepath.Join(append(parts, repo)...),
			DomainFolder: domainFolder,
		}
	}

	if isAbsoluteTarget(rule.Target) {
		// An absolute target opts out of root and domainFolder entirely.
		return Destination{
			Path: filepath.Join(expandTilde(rule.Target, home), repo),
			Rule: &rule,
		}
	}

	return Destination{
		Path:         filepath.Join(base, rule.Target, repo),
		DomainFolder: domainFolder,
		Rule:         &rule,
	}
}

// GroupPath renders group segments the way patterns are written in the config,
// e.g. ["f1", "f2"] -> "/f1/f2".
func GroupPath(groups []string) string {
	return "/" + strings.Join(groups, "/")
}

// ApplyConfig merges the config file and the command-line flags onto c, then
// works out where the repository belongs. Flags are the last layer, so they win
// over both the "default" block and the host's own section.
func ApplyConfig(c *Gc2Cmd, f config.File, home string) {
	cfg := f.For(c.Repo_domain)

	if c.Depth != 0 {
		cfg.Depth = c.Depth
	}
	if c.Editor != "" {
		cfg.Editor = c.Editor
	}

	dest := Resolve(c.Repo_domain, c.Repo_groups, c.Repo_name, cfg, home)

	c.Cfg = cfg
	c.DestFullPath = dest.Path
	c.DomainFolderPath = dest.DomainFolder
	c.MatchedRule = dest.Rule
}

// matchRule finds the best rule for a group path.
//
// Go randomises map iteration, so picking the first match would send the same
// URL to different folders on different runs. Every candidate is collected and
// ordered deterministically instead.
func matchRule(groups []string, patterns map[string]string) (Rule, bool) {
	if len(patterns) == 0 {
		return Rule{}, false
	}

	var candidates []Rule

	for pattern, target := range patterns {
		rule, ok := compileRule(pattern, target)
		if !ok {
			continue
		}
		if rule.matches(groups) {
			candidates = append(candidates, rule)
		}
	}

	if len(candidates) == 0 {
		return Rule{}, false
	}

	sort.Slice(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]

		// An exact rule always beats a subtree rule.
		if a.greedy != b.greedy {
			return !a.greedy
		}
		// Then the most specific: most literal segments matched.
		if a.literals != b.literals {
			return a.literals > b.literals
		}
		// Then the least fuzzy.
		if a.wildcards != b.wildcards {
			return a.wildcards < b.wildcards
		}
		// Never leave a tie to chance.
		return a.Pattern < b.Pattern
	})

	return candidates[0], true
}

// compileRule turns a config key into a Rule. Patterns that would match
// everything are rejected rather than silently swallowing every repository.
func compileRule(pattern, target string) (Rule, bool) {
	if strings.TrimSpace(target) == "" {
		return Rule{}, false
	}

	segs := splitPath(pattern)

	rule := Rule{Pattern: pattern, Target: target}

	if len(segs) > 0 && segs[len(segs)-1] == "*" {
		rule.greedy = true
		segs = segs[:len(segs)-1]
	}

	for _, s := range segs {
		if s == "*" {
			rule.wildcards++
		} else {
			rule.literals++
		}
	}

	// A bare "/*" has nothing literal to anchor on.
	if rule.literals == 0 {
		return Rule{}, false
	}

	rule.segments = segs

	return rule, true
}

// matches reports whether a group path satisfies the rule.
func (r Rule) matches(groups []string) bool {
	if r.greedy {
		// The literal prefix must match, and anything may follow.
		if len(groups) < len(r.segments) {
			return false
		}
		return segmentsMatch(r.segments, groups[:len(r.segments)])
	}

	if len(groups) != len(r.segments) {
		return false
	}

	return segmentsMatch(r.segments, groups)
}

func segmentsMatch(pattern, groups []string) bool {
	for i, p := range pattern {
		if p == "*" {
			continue
		}
		if p != groups[i] {
			return false
		}
	}
	return true
}

// splitPath breaks "/f1/f2/" into ["f1", "f2"], dropping empty segments so
// leading, trailing and doubled slashes are all forgiving.
func splitPath(p string) []string {
	var out []string
	for _, s := range strings.Split(p, "/") {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// isAbsoluteTarget reports whether a mapping target should bypass root and
// domainFolder and be used as written.
func isAbsoluteTarget(target string) bool {
	return strings.HasPrefix(target, "/") || target == "~" || strings.HasPrefix(target, "~/")
}

// expandTilde resolves a leading ~ against home. Only a bare "~" or a "~/"
// prefix count, so "~user" is left alone rather than being mangled.
func expandTilde(p, home string) string {
	if p == "~" {
		return home
	}
	if strings.HasPrefix(p, "~/") {
		return filepath.Join(home, p[2:])
	}
	return p
}
