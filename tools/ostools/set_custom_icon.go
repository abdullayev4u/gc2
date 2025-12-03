package ostools

import "runtime"

func SetCustomIcon(folderPath string, iconPath string) error {
	
	if runtime.GOOS == `darwin` {
		return setCustomIcon_macos(folderPath, iconPath)
	}

	return nil
}
