package config

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const marker = "\n--APPEND_MARKER--\n"
const maxTailSize = 64 * 1024

// Exported variables
var (
	DeviceID string
	CertPem  []byte
	KeyPem   []byte
	CAPem    []byte
	GroupKey []byte
)

type embeddedData struct {
	DeviceID string `json:"device_id"`
	Cert     string `json:"cert"`
	Key      string `json:"key"`
	CACert   string `json:"ca_cert"`
	GroupKey string `json:"group_key"` // hex encoded
}

func LoadEmbeddedConfig() error {
	if os.Getenv("DEV_MODE") == "1" {
		return loadDevFiles()
	}

	exePath, err := os.Executable()

	if err != nil {
		return fmt.Errorf("cannot find executable path: %w", err)
	}
	exePath, _ = filepath.EvalSymlinks(exePath)

	f, err := os.Open(exePath)
	if err != nil {
		return fmt.Errorf("cannot open executable: %w", err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return fmt.Errorf("cannot stat executable: %w", err)
	}

	size := fi.Size()
	start := int64(0)
	if size > maxTailSize {
		start = size - maxTailSize
	}
	f.Seek(start, 0)

	buf := make([]byte, maxTailSize)
	n, _ := f.Read(buf)
	buf = buf[:n]

	idx := bytes.LastIndex(buf, []byte(marker))
	if idx == -1 {
		return fmt.Errorf("marker not found in binary")
	}

	var raw embeddedData
	if err := json.Unmarshal(buf[idx+len(marker):], &raw); err != nil {
		return fmt.Errorf("cannot parse embedded JSON: %w", err)
	}

	DeviceID = string(raw.DeviceID)
	CertPem = []byte(raw.Cert)
	KeyPem = []byte(raw.Key)
	CAPem = []byte(raw.CACert)

	GroupKey, err = hex.DecodeString(raw.GroupKey)
	if err != nil {
		return fmt.Errorf("cannot decode group key: %w", err)
	}

	return nil
}

func loadDevFiles() error {
	read := func(path string) ([]byte, error) {
		return os.ReadFile(filepath.Clean(path))
	}

	var err error

	CertPem, err = read("./config/certs/cert.pem")
	if err != nil {
		return fmt.Errorf("dev mode: failed to read cert.pem: %w", err)
	}

	KeyPem, err = read("./config/certs/key.pem")
	if err != nil {
		return fmt.Errorf("dev mode: failed to read key.pem: %w", err)
	}

	CAPem, err = read("./config/certs/ca.crt")
	if err != nil {
		return fmt.Errorf("dev mode: failed to read ca.crt: %w", err)
	}

	GroupKey, err = read("./config/certs/group_key.enc")
	if err != nil {
		return fmt.Errorf("dev mode: invalid group key hex: %w", err)
	}

	idBytes, err := read("./config/certs/device_id.txt")
	if err != nil {
		return fmt.Errorf("dev mode: failed to read device_id.txt: %w", err)
	}
	DeviceID = strings.TrimSpace(string(idBytes))

	if len(DeviceID) == 0 {
		return fmt.Errorf("dev mode: device ID is empty")
	}

	fmt.Println("[config] Loaded config in DEV_MODE")

	return nil
}
