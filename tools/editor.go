package tools

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/abdullayev4u/gc2/config"
)

func OpenEditor(c *Gc2Cmd) error {
	if c.Editor == "none" {
		return nil
	}

	editor := c.Editor
	if editor == "" {
		editor = config.DefaultEditor
	}

	destDir := filepath.Join(mustHomeDir(), c.Repo_domain, c.Repo_author, c.Repo_name)

	cmd := exec.Command(editor, destDir)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
