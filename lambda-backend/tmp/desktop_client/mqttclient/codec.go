package mqttclient

import (
	"bytes"
	"errors"
	"io"
	_ "embed"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
    "crypto/aes"
    "crypto/cipher"
    "hash"
)

// Type is 4 bytes
type DecodedPayload struct {
	Type     string
	Filename string
	DeviceID [32]byte
	Payload  []byte
}

/*
Each message consists of the following byte layout:

[1 byte]  MIME type length (N)
[N bytes] MIME type string (e.g., "text/plain", "image/png")

[1 byte]  Filename length (M)
[M bytes] Filename string (e.g., "note.png"; can be empty if clipboard)

[32 bytes] Device ID (Device UUID as 32-byte fixed string)

[12 bytes] Nonce (AES-GCM nonce)

[... bytes] Ciphertext (AES-GCM encrypted payload)
- Payload = raw file/text/image data
- Authenticated with header as AAD (type, filename, device ID)
- Includes 16-byte GCM tag at the end of ciphertext (standard behavior)

Decryption:
- The header ([type][filename][device ID]) is passed as AAD to AES-GCM
- The payload must be decrypted using the shared group_key
*/

//go:embed lambda_output/group_key.enc
var groupKeyEnc []byte

//go:embed lambda_output/key.pem
var keyPEM []byte

func encryptAESGCM(key, plaintext, aad []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, aad)
	return nonce, ciphertext, nil
}

func decryptAESGCM(key, nonce, ciphertext, aad []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, nonce, ciphertext, aad)
}

func decryptGroupKey(enc []byte, privKeyPEM []byte) ([]byte, error) {
	block, _ := pem.Decode(privKeyPEM)
	if block == nil {
		return nil, errors.New("failed to parse PEM block")
	}
	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		privKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	}

	rsaKey, ok := privKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}

	return rsa.DecryptOAEP(
		sha256Hash(),
		rand.Reader,
		rsaKey,
		enc,
		nil,
	)
}

func sha256Hash() hash.Hash {
	return crypto.SHA256.New()
}

func encodeMessage(mimeType, filename string, deviceId [32]byte, payload []byte) ([]byte, error) {
	groupKey, err := decryptGroupKey(groupKeyEnc, keyPEM)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)

	if len(mimeType) > 255 || len(filename) > 255 {
		return nil, errors.New("mime or filename too long")
	}

	buf.WriteByte(byte(len(mimeType)))
	buf.WriteString(mimeType)

	buf.WriteByte(byte(len(filename)))
	buf.WriteString(filename)

	buf.Write(deviceId[:])

	header := buf.Bytes()

	nonce, ciphertext, err := encryptAESGCM(groupKey, payload, header)
	if err != nil {
		return nil, err
	}

	buf.Write(nonce)
	buf.Write(ciphertext)

	return buf.Bytes(), nil
}

func decodeMessage(data []byte) (*DecodedPayload, error) {
	buf := bytes.NewReader(data)

	typeLen, _ := buf.ReadByte()
	typeBytes := make([]byte, typeLen)
	buf.Read(typeBytes)

	nameLen, _ := buf.ReadByte()
	nameBytes := make([]byte, nameLen)
	buf.Read(nameBytes)

	var devID [32]byte
	buf.Read(devID[:])

	// Construct header again for AAD
	headerLen := 1 + len(typeBytes) + 1 + len(nameBytes) + 32
	header := data[:headerLen]

	groupKey, err := decryptGroupKey(groupKeyEnc, keyPEM)
	if err != nil {
		return nil, err
	}

	nonceSize := 12
	nonce := make([]byte, nonceSize)
	buf.Read(nonce)

	ciphertext, _ := io.ReadAll(buf)
	plaintext, err := decryptAESGCM(groupKey, nonce, ciphertext, header)
	if err != nil {
		return nil, err
	}

	return &DecodedPayload{
		Type:     string(typeBytes),
		Filename: string(nameBytes),
		DeviceID: devID,
		Payload:  plaintext,
	}, nil
}
