//go:build windows

package uninstall

import (
	"os/exec"
	"syscall"
)

func startCleanupScript(scriptPath string) error {
	cmd := exec.Command(
		"powershell",
		"-NoProfile",
		"-ExecutionPolicy", "Bypass",
		"-WindowStyle", "Hidden",
		"-File", scriptPath,
	)
	// Windows-only fields – fine here
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
		// Optionally: CreationFlags: 0x08000000 | 0x00000008, // CREATE_NO_WINDOW | DETACHED_PROCESS
	}
	return cmd.Start()
}
