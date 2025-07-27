package config

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/zalando/go-keyring"
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

const keyringService = "HoppyShare"

type preDecrypt struct {
	DeviceID      string `json:"device_id"`
	EncryptedBlob string `json:"encrypted_blob"`
}

func LoadKeysFromKeychain() error {
	items := map[string]*[]byte{
		"CA":       &CAPem,
		"Cert":     &CertPem,
		"Key":      &KeyPem,
		"GroupKey": &GroupKey,
	}

	for key, dst := range items {
		enc, err := keyring.Get(keyringService, key)
		if err != nil {
			return fmt.Errorf("keychain: could get Get %q: %w", key, err)
		}
		data, err := base64.StdEncoding.DecodeString(enc)
		if err != nil {
			return fmt.Errorf("keychain: could not Base64-decode %q: %w", key, err)
		}
		*dst = data
	}

	enc, err := keyring.Get(keyringService, "DeviceID")
	if err != nil {
		return fmt.Errorf("keychain: could not get DeviceID: %w", err)
	}
	idBytes, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return fmt.Errorf("keychain: could not Base64-decode DeviceID: %w", err)
	}

	DeviceID = string(idBytes)

	return nil
}

type decryptedConfig struct {
	Cert     string `json:"cert"`
	Key      string `json:"key"`
	CACert   string `json:"ca_cert"`
	GroupKey string `json:"group_key"` // hex encoded
}

func LoadEmbeddedConfig() error {
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
	_, err = f.Seek(start, 0)
	if err != nil {
		return fmt.Errorf("seek failed: %w", err)
	}

	buf := make([]byte, maxTailSize)
	n, _ := f.Read(buf)
	buf = buf[:n]

	idx := bytes.LastIndex(buf, []byte(marker))
	if idx == -1 {
		return fmt.Errorf("marker not found in binary")
	}

	var embedded preDecrypt
	if err := json.Unmarshal(buf[idx+len(marker):], &embedded); err != nil {
		return fmt.Errorf("cannot parse embedded JSON: %w", err)
	}

	DeviceID = embedded.DeviceID

	encKeyBase64, err := FetchEncryptionKey(embedded, "https://en43r23fua.execute-api.us-east-2.amazonaws.com")

	log.Println("device id")
	print(embedded.DeviceID)
	log.Println("encrypted blob")
	print(embedded.EncryptedBlob)
	log.Println("encKeyBase64")
	log.Println(encKeyBase64)
	if err != nil {
		return fmt.Errorf("failed to fetch decryption key: %w", err)
	}
	encKey, err := base64.StdEncoding.DecodeString(encKeyBase64)
	if err != nil {
		return fmt.Errorf("invalid base64 decryption key: %w", err)
	}

	cipherData, err := base64.StdEncoding.DecodeString(embedded.EncryptedBlob)
	if err != nil {
		return fmt.Errorf("invalid base64 encrypted blob: %w", err)
	}
	if len(cipherData) < 12 {
		return fmt.Errorf("encrypted blob too short")
	}
	nonce := cipherData[:12]
	ciphertext := cipherData[12:]

	block, err := aes.NewCipher(encKey)
	if err != nil {
		return fmt.Errorf("failed to init cipher: %w", err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to init AES-GCM: %w", err)
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, []byte(DeviceID))
	if err != nil {
		return fmt.Errorf("failed to decrypt blob: %w", err)
	}

	var raw decryptedConfig
	if err := json.Unmarshal(plaintext, &raw); err != nil {
		return fmt.Errorf("failed to parse decrypted JSON: %w", err)
	}

	CertPem = []byte(raw.Cert)
	KeyPem = []byte(raw.Key)
	CAPem = []byte(raw.CACert)
	GroupKey, err = hex.DecodeString(raw.GroupKey)
	if err != nil {
		return fmt.Errorf("invalid hex group key: %w", err)
	}

	return nil
}

func LoadDevFiles() error {
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

type decryptResponse struct {
	EncryptionKey string `json:"encryption_key"`
}

func FetchEncryptionKey(cfg preDecrypt, apiBase string) (string, error) {
	reqBody := map[string]string{
		"encrypted_blob": cfg.EncryptedBlob,
	}
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	url := fmt.Sprintf("%s/prod/api/decrypt/%s", apiBase, cfg.DeviceID)

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Handle non-200 responses
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	var fullResp decryptResponse
	if err := json.Unmarshal(body, &fullResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}

	return fullResp.EncryptionKey, nil
}
