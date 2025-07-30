//go:build windows
// +build windows

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

	volume := int(vol * 100) // WMPlayer expects 0–100

	psScript := `
		$player = New-Object -ComObject WMPlayer.OCX
		$media = $player.newMedia("` + tmpPath + `")
		$player.settings.volume = ` + fmt.Sprintf("%d", volume) + `
		$player.controls.play()
		while ($player.playState -ne 1) { Start-Sleep -Milliseconds 100 }
		`

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden",
		"-Command", psScript)

	cmd.Run()
}
