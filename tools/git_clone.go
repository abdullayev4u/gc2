package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/abdullayev4u/gc2/config"
)

func GitClone(c *Gc2Cmd) error {
	gitArgs := []string{"clone"}

	if d := c.Depth; d >= 0 {
		if d == 0 {
			d = config.DefaultDepth
		}
		gitArgs = append(gitArgs, "--depth", strconv.Itoa(d))
	}

	{
		dest := filepath.Join(mustHomeDir(), c.Repo_domain, c.Repo_author, c.Repo_name)

		gitArgs = append(gitArgs, c.RepoUrl, dest)
	}

	cmd := exec.Command("git", gitArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
