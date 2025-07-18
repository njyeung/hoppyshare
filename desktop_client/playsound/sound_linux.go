//go:build linux
// +build linux

package playsound

import (
	"os"
	"os/exec"
	"path/filepath"
)

func play(sound []byte) {
	tmpPath := filepath.Join(os.TempDir(), "notification.wav")
	os.WriteFile(tmpPath, sound, 0644)

	// Try aplay, fallback to paplay if available
	if err := exec.Command("aplay", tmpPath).Run(); err != nil {
		exec.Command("paplay", tmpPath).Run()
	}
}
