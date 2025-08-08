package notification

import (
	_ "embed"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gen2brain/beeep"
)

//go:embed assets/macOS/default.png
var defaultIconMacOS []byte

//go:embed assets/windows/default.ico
var defaultIconWindows []byte

func Notification(message string) error {
	var iconPath string
	var err error

	// Create temporary icon file based on OS
	if runtime.GOOS == "windows" {
		iconPath, err = createTempIcon("hoppyshare_icon.ico", defaultIconWindows)
	} else {
		iconPath, err = createTempIcon("hoppyshare_icon.png", defaultIconMacOS)
	}

	if err != nil {
		return beeep.Notify("HoppyShare", message, "")
	}
	defer os.Remove(iconPath)

	return beeep.Notify("HoppyShare", message, iconPath)
}

// createTempIcon creates a temporary file with the icon data
func createTempIcon(filename string, data []byte) (string, error) {
	tempDir := os.TempDir()
	iconPath := filepath.Join(tempDir, filename)

	err := os.WriteFile(iconPath, data, 0644)
	if err != nil {
		return "", err
	}

	return iconPath, nil
}
