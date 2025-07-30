//go:build linux
// +build linux

package playsound

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func play(sound []byte, vol float32) {
	tmpPath := filepath.Join(os.TempDir(), "notification.wav")
	os.WriteFile(tmpPath, sound, 0644)

	// Try paplay with vol
	volumeInt := int32(vol * 65536)
	cmd := exec.Command("paplay", "--volume", fmt.Sprintf("%d", volumeInt), tmpPath)
	err := cmd.Run()

	if err != nil {
		// Fallback to aplay (no volume support)
		exec.Command("aplay", tmpPath).Run()
	}
}
