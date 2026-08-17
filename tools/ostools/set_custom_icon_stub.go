// This condition is the exact negation of the one on set_custom_icon_macos.go,
// so precisely one definition of setCustomIcon_macos exists in every build.
// !cgo is part of it because CGO_ENABLED=0 drops the Cocoa file even on macOS.
//
//go:build !darwin || !cgo

package ostools

// Never reached at run time: SetCustomIcon only calls this when GOOS is
// darwin, and on darwin with cgo the real implementation is the one built.
func setCustomIcon_macos(folderPath string, iconPath string) error {
	return nil
}
