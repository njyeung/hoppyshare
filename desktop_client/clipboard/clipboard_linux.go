//go:build linux

package clipboard

import (
	"bytes"
	"errors"
	"os/exec"
)

func readClipboard() ([]byte, string, error) {
	var out bytes.Buffer

	// Wayland
	_, err := exec.LookPath("wl-paste")
	if err == nil {
		cmd := exec.Command("wl-paste", "--no-newline")
		cmd.Stdout = &out
		err := cmd.Run()
		if err == nil {
			return out.Bytes(), "text/plain", nil
		}
	}

	// X11
	_, err = exec.LookPath("xclip")
	if err == nil {
		cmd := exec.Command("xclip", "-selection", "clipboard", "-o")
		cmd.Stdout = &out
		err := cmd.Run()
		if err == nil {
			return out.Bytes(), "text/plain", nil
		}
	}

	return nil, "", errors.New("clipboard read failed: Tried wl-paste and xclip")

}

func writeClipboard(data []byte, mimeType string) error {
	if mimeType != "text/plain" && mimeType != "image/png" {
		return errors.New("unsupported MIME type for writing to clipboard")
	}

	// Wayland
	_, err := exec.LookPath("wl-copy")
	if err == nil {
		args := []string{"--type", mimeType}
		cmd := exec.Command("wl-copy", args...)
		cmd.Stdin = bytes.NewReader(data)
		err := cmd.Run()
		if err != nil {
			return errors.New("wl-copy failed: " + err.Error())
		}
		return nil
	}

	// X11 (only supports text)
	if mimeType == "text/plain" {
		_, err := exec.LookPath("xclip")
		if err == nil {
			cmd := exec.Command("xclip", "-selection", "clipboard")
			cmd.Stdin = bytes.NewReader(data)
			err := cmd.Run()
			if err != nil {
				return errors.New("xclip failed: " + err.Error())
			}
			return nil
		}
	}

	return errors.New("no clipboard tool found, tried wl-copy and xclip")
}
