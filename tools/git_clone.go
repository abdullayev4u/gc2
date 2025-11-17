package tools

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/abdullayev4u/gc2/config"
)

func GitClone(c *Gc2Cmd) error {
	gitArgs := []string{"clone"}

	{
		d := c.Depth
		if d == 0 {
			d = config.DefaultDepth
		}
		if d > 0 {
			gitArgs = append(gitArgs, "--depth", strconv.Itoa(d))
		}
	}

	{
		gitArgs = append(gitArgs, c.RepoUrl, c.DestFullPath)
	}

	var stderr bytes.Buffer

	cmd := exec.Command("git", gitArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)

	err := cmd.Run()

	if err != nil {
		val := stderr.String()
		alreadyExistsErr := fmt.Sprintf("fatal: destination path '%s' already exists and is not an empty directory.", c.DestFullPath)

		isAlreadyExistsErr := strings.Contains(val, alreadyExistsErr)

		if isAlreadyExistsErr && config.OpenEditorEvenAlredyExists {
			return nil
		}

		return err
	}

	return nil
}
