package config

const Version = "1.1"

// Built-in defaults.
//
// These serve two purposes: they are the values written into a freshly
// generated ~/.gc2.yaml, and they are the fallback for any key the user later
// removes from it. The merge never depends on a key being physically present
// in the file.
const (
	DefaultRoot           = "~"
	DefaultDomainFolder   = true
	DefaultEditor         = "code"
	DefaultDepth          = -1
	DefaultOpenExisting   = true
	DefaultSyncDomainIcon = false
)

// Settings is one layer of configuration, as written under a root key of
// ~/.gc2.yaml. Every field is a pointer so the merge can tell "the user did
// not mention this" apart from "the user explicitly set it to false or 0".
type Settings struct {
	Root           *string           `yaml:"root,omitempty"`
	DomainFolder   *bool             `yaml:"domainFolder,omitempty"`
	Editor         *string           `yaml:"editor,omitempty"`
	Depth          *int              `yaml:"depth,omitempty"`
	OpenExisting   *bool             `yaml:"openExisting,omitempty"`
	SyncDomainIcon *bool             `yaml:"syncDomainIcon,omitempty"`
	Paths          map[string]string `yaml:"paths,omitempty"`
}

// File is the whole config document: one entry per root key. The key
// DefaultKey holds settings shared by every host, any other key is a hostname.
type File map[string]Settings

// DefaultKey is the root key holding host-independent settings.
const DefaultKey = "default"

// Resolved is a fully populated configuration for one specific host, after
// every layer has been merged. No pointers: every value is decided.
type Resolved struct {
	Root           string
	DomainFolder   bool
	Editor         string
	Depth          int
	OpenExisting   bool
	SyncDomainIcon bool

	// HostPaths are the mapping rules declared under the host's own key, and
	// GlobalPaths those under "default". Host rules are consulted first.
	HostPaths   map[string]string
	GlobalPaths map[string]string
}

// For merges the built-in defaults, the "default" block and the host's own
// block into a single Resolved. Later layers win.
func (f File) For(host string) Resolved {
	r := Resolved{
		Root:           DefaultRoot,
		DomainFolder:   DefaultDomainFolder,
		Editor:         DefaultEditor,
		Depth:          DefaultDepth,
		OpenExisting:   DefaultOpenExisting,
		SyncDomainIcon: DefaultSyncDomainIcon,
	}

	if d, ok := f[DefaultKey]; ok {
		r.apply(d)
		r.GlobalPaths = d.Paths
	}

	// A host key must never be matched against DefaultKey, otherwise a server
	// literally named "default" would silently pick up the global block twice.
	if host != DefaultKey {
		if h, ok := f[host]; ok {
			r.apply(h)
			r.HostPaths = h.Paths
		}
	}

	return r
}

// apply overlays the fields s actually sets onto r.
func (r *Resolved) apply(s Settings) {
	if s.Root != nil {
		r.Root = *s.Root
	}
	if s.DomainFolder != nil {
		r.DomainFolder = *s.DomainFolder
	}
	if s.Editor != nil {
		r.Editor = *s.Editor
	}
	if s.Depth != nil {
		r.Depth = *s.Depth
	}
	if s.OpenExisting != nil {
		r.OpenExisting = *s.OpenExisting
	}
	if s.SyncDomainIcon != nil {
		r.SyncDomainIcon = *s.SyncDomainIcon
	}
}

// Defaults returns the built-in defaults as a Settings, which is what gets
// serialised into a generated config file.
func Defaults() Settings {
	root := DefaultRoot
	domainFolder := DefaultDomainFolder
	editor := DefaultEditor
	depth := DefaultDepth
	openExisting := DefaultOpenExisting
	syncDomainIcon := DefaultSyncDomainIcon

	return Settings{
		Root:           &root,
		DomainFolder:   &domainFolder,
		Editor:         &editor,
		Depth:          &depth,
		OpenExisting:   &openExisting,
		SyncDomainIcon: &syncDomainIcon,
	}
}
