package tools

import (
	"os"
	"os/exec"
)

func OpenEditor(c *Gc2Cmd) error {
	editor := c.Cfg.Editor

	// "none" is the opt-out, from either the flag or the config file.
	if editor == "none" || editor == "" {
		return nil
	}

	cmd := exec.Command(editor, c.DestFullPath)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
