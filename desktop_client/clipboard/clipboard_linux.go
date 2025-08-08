//go:build linux

package clipboard

import (
	"bytes"
	"errors"
	"os/exec"
)

func readClipboard() ([]byte, string, error) {
	var out bytes.Buffer

	// Wayland - check available types and pick the best one
	_, err := exec.LookPath("wl-paste")
	if err == nil {
		// Get list of available types
		var typesOut bytes.Buffer
		typesCmd := exec.Command("wl-paste", "--list-types")
		typesCmd.Stdout = &typesOut
		if typesCmd.Run() == nil {
			availableTypes := typesOut.Bytes()

			// Check for image formats first
			for _, mimeType := range []string{"image/png", "image/jpeg", "image/gif"} {
				if bytes.Contains(availableTypes, []byte(mimeType)) {
					cmd := exec.Command("wl-paste", "--type", mimeType)
					cmd.Stdout = &out
					if cmd.Run() == nil && out.Len() > 0 {
						return out.Bytes(), mimeType, nil
					}
					out.Reset()
				}
			}

			// Check for text if no images found
			if bytes.Contains(availableTypes, []byte("text/plain")) {
				cmd := exec.Command("wl-paste", "--type", "text/plain", "--no-newline")
				cmd.Stdout = &out
				if cmd.Run() == nil {
					return out.Bytes(), "text/plain", nil
				}
			}
		}

		// Fallback to default (usually text)
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
	if mimeType != "text/plain" && mimeType != "image/png" && mimeType != "image/jpeg" && mimeType != "image/gif" {
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
