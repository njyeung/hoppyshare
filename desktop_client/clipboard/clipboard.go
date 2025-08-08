package clipboard

// Reads content from the clipboard and returns the given MIME type (e.g., "text/plain", "image/png").
func Read() ([]byte, string, error) {
	return readClipboard()
}

// Puts content onto the clipboard with the given MIME type (e.g., "text/plain", "image/png").
func Write(data []byte, mimeType string) error {
	return writeClipboard(data, mimeType)
}
