//go:build darwin || linux

package uninstall

import "os/exec"

func startCleanupScript(scriptPath string) error {
	cmd := exec.Command("sh", "-c", scriptPath+" &")
	return cmd.Start()
}
