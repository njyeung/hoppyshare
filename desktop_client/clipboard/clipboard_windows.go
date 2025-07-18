//go:build windows

package clipboard

import "errors"

func readClipboard() ([]byte, string, error) {
	return nil, "", errors.New("not implemented")
}
func writeClipboard(data []byte, mimeType string) error {
	return nil
}
