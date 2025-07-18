package mqttclient

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"desktop_client/config"
	"encoding/pem"
	"errors"
	"io"
)

type DecodedPayload struct {
	Type     string
	Filename string
	DeviceID [32]byte
	Payload  []byte
}

func encodeMessage(mimeType, filename string, deviceID string, payload []byte) ([]byte, error) {
	groupKey, err := decryptGroupKey(config.GroupKey, config.KeyPem)
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

	hashedID := hashDeviceID(deviceID)
	buf.Write(hashedID[:])

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

	headerLen := 1 + len(typeBytes) + 1 + len(nameBytes) + 32
	header := data[:headerLen]

	groupKey, err := decryptGroupKey(config.GroupKey, config.KeyPem)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, 12)
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

	return rsa.DecryptOAEP(sha256.New(), rand.Reader, rsaKey, enc, nil)
}

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

func hashDeviceID(id string) [32]byte {
	return sha256.Sum256([]byte(id))
}
