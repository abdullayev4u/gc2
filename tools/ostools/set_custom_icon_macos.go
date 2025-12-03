package ostools

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

// Objective-C helper function
int SetFolderIcon(const char* folderPath, const char* iconPath) {
    @autoreleasepool {
        NSString *fPath = [NSString stringWithUTF8String:folderPath];
        NSString *iPath = [NSString stringWithUTF8String:iconPath];

        // Load the image
        NSImage *image = [[NSImage alloc] initWithContentsOfFile:iPath];
        if (image == nil) {
            return -1; // Image load failed
        }

        // Set the icon
        BOOL result = [[NSWorkspace sharedWorkspace] setIcon:image forFile:fPath options:0];
        return result ? 1 : 0;
    }
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func setCustomIcon_macos(folderPath string, iconPath string) error {
	cFolder := C.CString(folderPath)
	cIcon := C.CString(iconPath)
	defer C.free(unsafe.Pointer(cFolder))
	defer C.free(unsafe.Pointer(cIcon))

	// Call the Objective-C function
	result := C.SetFolderIcon(cFolder, cIcon)

	if result == -1 {
		return fmt.Errorf("failed to load image at %s", iconPath)
	}
	if result == 0 {
		return fmt.Errorf("failed to set icon for folder %s (check permissions)", folderPath)
	}

	return nil
}
