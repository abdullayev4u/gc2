package tools

import (
	"os"
	"regexp"
)

func mustHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic("cannot determine user home directory: " + err.Error())
	}

	return home
}

func isValidCmdName(name string) bool {
	// Must start with a letter (A-Z or a-z),
	// then can contain letters, digits, -, _, or .
	var re = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9._-]*$`)
	return re.MatchString(name)
}
