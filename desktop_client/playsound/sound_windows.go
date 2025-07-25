//go:build windows
// +build windows

package playsound

import (
	"os"
	"os/exec"
	"path/filepath"
)

func play(sound []byte) {
	tmpPath := filepath.Join(os.TempDir(), "notification.wav")
	os.WriteFile(tmpPath, sound, 0644)
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden",
		"-Command", `Start-Sleep -Milliseconds 100; (New-Object Media.SoundPlayer "`+tmpPath+`").PlaySync()`)
	cmd.Run()
}
