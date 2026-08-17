package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FileName is the config file gc2 looks for in the user's home directory.
const FileName = ".gc2.yaml"

// EnvVar overrides the config location, mostly so tests stay off the real $HOME.
const EnvVar = "GC2_CONFIG"

// Template is written verbatim when no config file exists yet.
//
// It is hand-written rather than marshalled so it can carry comments: this file
// is how most people will discover path mapping, so it has to read like
// documentation. The values in the "default" block must stay in step with the
// constants in config.go — TestTemplateMatchesDefaults enforces that.
//
// Note "~" is quoted: a bare ~ is null in YAML, not a home-directory shorthand.
const Template = `# gc2 configuration
#
# Settings under "default" apply to every git host. Add a section named after a
# host to override them for that host alone.

default:
  root: "~"                # base directory for everything gc2 clones
  domainFolder: true       # create <root>/<host>/... ; false drops the host level
  editor: code             # command used to open a repo; "gc2 -e none <url>" skips it
  depth: -1                # >0 clones with --depth N; anything else clones full history
  openExisting: true       # already cloned? open it instead of failing
  syncDomainIcon: false    # macOS only: paint the host's favicon onto its folder

# Per-host overrides. Uncomment and edit for your own git server:
#
# gitlab.mycompany.com:
#   root: ~/work           # this host's repos live somewhere else entirely
#   editor: goland
#   paths:
#     # Without a rule, the full group path is mirrored on disk:
#     #   gitlab.mycompany.com/a/b/c/d/repo -> ~/work/gitlab.mycompany.com/a/b/c/d/repo
#     #
#     # Rules collapse a group path onto a short local folder instead:
#     "/f1/f2/f3/f4": thatProject      # exact: only repos directly in that group
#     "/f1/f2/f3/f4/*": thatProject    # trailing *: that group and everything below it
#     "/company/*/backend": services   # interior *: matches exactly one segment
#
#     # A target starting with ~ or / is used as-is, ignoring root and domainFolder:
#     "/scratch/*": "~/tmp/scratch"
#
# Run "gc2 where <url>" to see where a repo would land without cloning it.
`

// Path returns the active config file location.
func Path() (string, error) {
	if p := os.Getenv(EnvVar); p != "" {
		return p, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}

	return filepath.Join(home, FileName), nil
}

// Load reads the config file, creating it from Template when it does not exist
// yet. A missing file is normal; a malformed one is not, and is reported with
// the offending line so it can be fixed rather than silently ignored.
func Load() (File, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)

	switch {
	case errors.Is(err, fs.ErrNotExist):
		// Creating the file is best-effort. The user asked gc2 to clone a
		// repository, not to manage a config file, so a read-only home
		// directory warns and carries on with the same values the file
		// would have contained.
		if werr := WriteTemplate(path); werr != nil {
			fmt.Fprintf(os.Stderr, "warning: could not create %s: %s\n", path, werr)
		}
		data = []byte(Template)

	case err != nil:
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	f, err := parse(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	return f, nil
}

// parse decodes a config document. Unknown keys are rejected so that a typo
// like "domainfolder" is reported instead of silently falling back to defaults.
func parse(data []byte) (File, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)

	var f File
	if err := dec.Decode(&f); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	if f == nil {
		f = File{}
	}

	return f, nil
}

// WriteTemplate writes the starter config to path.
//
// The write goes through a temporary file in the same directory and is then
// renamed into place, so an interrupted run cannot leave behind a half-written
// config that fails to parse on every subsequent run.
func WriteTemplate(path string) error {
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".gc2.yaml.tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	// Cleans up on any early return; a no-op once the rename has succeeded.
	defer os.Remove(tmpName)

	if _, err := tmp.WriteString(Template); err != nil {
		tmp.Close()
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.Chmod(tmpName, 0o644); err != nil {
		return err
	}

	return os.Rename(tmpName, path)
}

// Exists reports whether a config file is already present.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
