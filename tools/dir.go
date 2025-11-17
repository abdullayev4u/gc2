package tools

import (
	"os"
	"path/filepath"
)

func EnsureParent(c *Gc2Cmd) error {

	parent := filepath.Dir(c.DestFullPath)

	return os.MkdirAll(parent, 0o755)
}
