package tools

import (
	"os"
	"path/filepath"
)

func EnsureParent(c *Gc2Cmd) error {

	destDir := filepath.Join(mustHomeDir(), c.Repo_domain, c.Repo_author, c.Repo_name)

	parent := filepath.Dir(destDir)

	return os.MkdirAll(parent, 0o755)
}
