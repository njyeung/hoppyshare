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

	// Try paplay at full volume
	cmd := exec.Command("paplay", tmpPath)
	err := cmd.Run()

	if err != nil {
		// Fallback to aplay
		exec.Command("aplay", tmpPath).Run()
	}
}
