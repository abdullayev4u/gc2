package tools

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/abdullayev4u/gc2/config"
)

func GitClone(c *Gc2Cmd) error {
	cloned, err := inspectDestination(c)
	if err != nil {
		return err
	}

	// Already the right repository sitting in the right place; nothing to
	// clone, and OpenEditor still gets to run.
	if cloned {
		fmt.Printf("already cloned: %s\n", c.DestFullPath)
		return nil
	}

	gitArgs := []string{"clone"}

	{
		if d := c.Cfg.Depth; d > 0 {
			gitArgs = append(gitArgs, "--depth", strconv.Itoa(d))
		}
	}

	{
		gitArgs = append(gitArgs, c.RepoUrl, c.DestFullPath)
	}

	cmd := exec.Command("git", gitArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// inspectDestination decides whether the destination is safe to clone into.
//
// Mapping rules deliberately collapse many remote paths onto one local folder,
// so two different repositories can end up wanting the same directory. Left
// unchecked that is silent and dangerous: the clone fails, openExisting
// swallows the failure, and the editor opens the wrong project. Comparing the
// origin remote turns that into a loud, specific error.
//
// It reports whether the destination already holds the requested repository.
func inspectDestination(c *Gc2Cmd) (cloned bool, err error) {
	info, err := os.Stat(c.DestFullPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if !info.IsDir() {
		return false, fmt.Errorf("destination [%s] exists and is not a directory", c.DestFullPath)
	}

	empty, err := isEmptyDir(c.DestFullPath)
	if err != nil {
		return false, err
	}
	if empty {
		// git is happy to clone into an existing empty directory.
		return false, nil
	}

	origin, ok := gitOrigin(c.DestFullPath)
	if !ok {
		return false, fmt.Errorf(
			"destination [%s] already exists and is not a git repository;\n	move it aside, or map this group somewhere else in %s",
			c.DestFullPath, configHint(),
		)
	}

	if !sameRepo(origin, c.RepoUrl) {
		return false, fmt.Errorf(
			"destination [%s] already holds a different repository;\n	there:     %s\n	requested: %s\n	add a more specific rule in %s so these two do not collapse onto one folder",
			c.DestFullPath, origin, c.RepoUrl, configHint(),
		)
	}

	if !c.Cfg.OpenExisting {
		return false, fmt.Errorf("destination [%s] already holds this repository", c.DestFullPath)
	}

	return true, nil
}

// gitOrigin returns the origin remote of a checkout, if it is one.
func gitOrigin(dir string) (string, bool) {
	out, err := exec.Command("git", "-C", dir, "remote", "get-url", "origin").Output()
	if err != nil {
		return "", false
	}

	origin := strings.TrimSpace(string(out))

	return origin, origin != ""
}

// sameRepo compares two clone URLs written in different forms, so that
// git@host:a/b.git and https://host/a/b are recognised as the same repository.
func sameRepo(a, b string) bool {
	return normalizeRepoUrl(a) == normalizeRepoUrl(b)
}

// normalizeRepoUrl reduces a clone URL to "host/path", dropping the scheme,
// any user info, the port, a .git suffix and trailing slashes.
func normalizeRepoUrl(raw string) string {
	host, path, err := splitRepoUrl(raw)
	if err != nil {
		return strings.ToLower(strings.TrimSpace(raw))
	}

	path = strings.TrimSuffix(strings.Trim(path, "/"), ".git")

	return strings.ToLower(host + "/" + path)
}

func isEmptyDir(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}

	return len(entries) == 0, nil
}

// configHint names the config file in error messages.
func configHint() string {
	path, err := config.Path()
	if err != nil {
		return "your gc2 config"
	}

	return path
}
