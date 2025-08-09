package ble

import (
	"bytes"
	"crypto/rand"
	"desktop_client/config"
	"desktop_client/mqttclient"
	"desktop_client/notification"
	"desktop_client/settings"
	"encoding/binary"
	"log"
	"net"
	"sync"
	"time"
)

var MAX_CHUNK_SIZE = 500

var (
	mu          sync.Mutex
	started     bool
	lastMessage []byte
	callback    func()
)

// BLE CHUNK FORMAT (4 byte header) + 500
//
// msgID [2]byte
// seq uint16 (last bit indicates LAST chunk)
// data []byte
//

type chunkBuffer struct {
	chunks   map[uint32][]byte
	received map[uint32]bool
	total    uint32
}

var (
	assembleMu sync.Mutex
	buffers    = make(map[string]*chunkBuffer)
)

func Start(clientID, deviceID string) error {

	if !settings.GetSettings().Enabled {
		return nil
	}

	mu.Lock()
	defer mu.Unlock()

	if started {
		return nil
	}

	if err := startBLE(clientID, deviceID); err != nil {
		return err
	}

	started = true
	notification.Notification("BLE bridge started")

	return nil
}

func Stop() error {
	mu.Lock()
	defer mu.Unlock()

	if !started {
		return nil
	}

	if err := stopBLE(); err != nil {
		return err
	}

	started = false
	notification.Notification("BLE bridge stopped")

	return nil
}

func Publish(content []byte, mimeType string, filename string) error {
	mu.Lock()
	on := started
	mu.Unlock()

	if !on {
		return nil
	}

	payload, err := mqttclient.EncodeMessage(mimeType, filename, config.DeviceID, content)
	if err != nil {
		return err
	}

	return PublishChunked(payload)
}

func SetOnMessageCallback(cb func()) {
	mu.Lock()
	defer mu.Unlock()
	callback = cb
}

func GetLastMessage() (filename string, contentType string, data []byte, ok bool) {
	mu.Lock()
	defer mu.Unlock()

	decoded, err := mqttclient.DecodeMessage(lastMessage)
	if err != nil {
		return "", "", nil, false
	}

	return decoded.Filename, decoded.Type, decoded.Payload, true
}

// exported to cgo layer
func onMessage(deviceID string, payload []byte) {
	handleChunk(payload)
}

func ClearMsg() {
	mu.Lock()
	lastMessage = nil
	mu.Unlock()

	assembleMu.Lock()
	defer assembleMu.Unlock()
	for k := range buffers {
		delete(buffers, k)
	}
}

func PublishChunked(payload []byte) error {
	// assign a uuid for the message
	msgID := make([]byte, 2)

	_, err := rand.Read(msgID)
	if err != nil {
		return err
	}
	totalChunks := (len(payload) + MAX_CHUNK_SIZE - 1) / MAX_CHUNK_SIZE

	// each chunk of the payload gets an increasing id
	for i := 0; i < totalChunks; i++ {
		log.Printf("SENDING CHUNK")
		start := i * MAX_CHUNK_SIZE
		end := start + MAX_CHUNK_SIZE
		if end > len(payload) {
			end = len(payload)
		}

		chunk := payload[start:end]

		buff := &bytes.Buffer{}

		// 	msgID 	[2]byte
		buff.Write(msgID)

		// 	seq 	uint16 (least significant bit is LAST flag)
		seq := uint16(i) << 1
		if i == totalChunks-1 {
			seq |= 1
		} else {
			seq &^= 1
		}
		binary.Write(buff, binary.BigEndian, seq)

		// 	data 	[]byte
		buff.Write(chunk)

		publishBLE(buff.Bytes())
		time.Sleep(20 * time.Millisecond)
	}

	return nil
}

func handleChunk(chunk []byte) {
	log.Printf("RECEIVED CHUNK")
	if len(chunk) < 4 {
		return
	}

	msgId := chunk[0:2]
	seqRaw := binary.BigEndian.Uint16(chunk[2:4])
	seqIndex := uint32(seqRaw >> 1)
	isLast := (seqRaw & 1) == 1
	chunkData := chunk[4:]
	msgKey := string(msgId)

	assembleMu.Lock()
	buf, exists := buffers[msgKey]
	if !exists {
		buf = &chunkBuffer{
			chunks:   make(map[uint32][]byte),
			received: make(map[uint32]bool),
		}

		buffers[msgKey] = buf
	}

	if isLast {
		buf.total = seqIndex + 1
		log.Printf("IS LAST")
	}

	buf.chunks[seqIndex] = chunkData
	buf.received[seqIndex] = true

	if buf.total > 0 {
		complete := true
		for i := uint32(0); i < buf.total; i++ {
			if _, ok := buf.chunks[i]; !ok {
				complete = false
				break
			}
		}
		if complete {
			full := &bytes.Buffer{}
			for i := uint32(0); i < buf.total; i++ {
				full.Write(buf.chunks[i])
			}
			delete(buffers, msgKey)
			assembleMu.Unlock()

			mu.Lock()
			lastMessage = full.Bytes()
			cb := callback
			mu.Unlock()

			if cb != nil {
				cb()
			}
			return
		}
	}
	assembleMu.Unlock()
}

// Generates Message ID
// 1 byte 		MAC addr hash
// 1 byte		5 minute time bucket hash
func GenerateMsgID() []byte {

	msgID := make([]byte, 2)

	mac, err := getMACAddr()
	if err != nil {
		// Could not get MAC addr, falback to random byte
		msgID[0] = byte(time.Now().UnixNano() % 256)
	} else {
		var hash byte

		for _, b := range mac {
			hash ^= b
		}

		msgID[0] = hash
	}

	now := time.Now().Unix()
	fiveMinBucket := byte((now / 300))
	msgID[1] = fiveMinBucket

	return msgID
}

func getMACAddr() (net.HardwareAddr, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback == 0 && len(iface.HardwareAddr) > 0 {
			return iface.HardwareAddr, nil
		}
	}

	return nil, err
}
