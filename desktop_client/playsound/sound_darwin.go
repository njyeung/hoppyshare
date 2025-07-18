//go:build darwin
// +build darwin

package playsound

import (
	"os"
	"os/exec"
	"path/filepath"
)

func play(sound []byte) {
	tmpPath := filepath.Join(os.TempDir(), "notification.wav")
	os.WriteFile(tmpPath, sound, 0644)
	exec.Command("afplay", tmpPath).Run()
}
