package tools

import (
	"os"
	"os/exec"

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

	cmd := exec.Command(editor, c.DestFullPath)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
