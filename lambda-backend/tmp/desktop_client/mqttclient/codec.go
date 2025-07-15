package mqttclient

import (
	"bytes"
	"errors"
	"io"
	"log"
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

[1 byte] MIME type length (N)
[N bytes] MIME type string (ex: "text/plain", "image/png")

[1 byte] Filename length (M)
[M bytes] Filename string (ex: "note.png", can be nothing if we're sending clipboard)

[32 bytes] Device ID (device UUID string)

[... bytes] Payload (the raw file/text/image data)

*/

func encodeMessage(mimeType, filename string, deviceId [32]byte, payload []byte) ([]byte, error) {
	buf := new(bytes.Buffer)

	if len(mimeType) > 255 {
		log.Println("MIME type too long")
		return nil, errors.New("mime type too long")
	}

	buf.WriteByte(byte(len(mimeType)))
	buf.WriteString(mimeType)

	if len(filename) > 255 {
		return nil, errors.New("filename too long")
	}
	buf.WriteByte(byte(len(filename)))
	buf.WriteString(filename)

	buf.Write(deviceId[:])
	buf.Write(payload)

	return buf.Bytes(), nil
}

func decodeMessage(data []byte) (*DecodedPayload, error) {
	buf := bytes.NewReader(data)

	// MIME Type
	typeLen, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	typeBytes := make([]byte, typeLen)
	if _, err := buf.Read(typeBytes); err != nil {
		return nil, err
	}

	// Filename
	nameLen, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	nameBytes := make([]byte, nameLen)
	if _, err := buf.Read(nameBytes); err != nil {
		return nil, err
	}

	// Device ID
	var devID [32]byte
	if _, err := buf.Read(devID[:]); err != nil {
		return nil, err
	}

	// Remaining payload
	payload, err := io.ReadAll(buf)
	if err != nil {
		return nil, err
	}

	return &DecodedPayload{
		Type:     string(typeBytes),
		Filename: string(nameBytes),
		DeviceID: devID,
		Payload:  payload,
	}, nil
}
