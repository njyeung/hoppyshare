package ble

import "sync"

var MAX_CHUNK_SIZE = 120

var (
	mu          sync.Mutex
	started     bool
	lastMessage []byte
	callback    func()
)

// BLE CHUNK FORMAT (6 byte header) + 120
//
// msgID [2]byte
// seq uint32 (last bit indicates LAST chunk)
// data []byte
//

func Start(clientID, deviceID string) error {
	mu.Lock()
	defer mu.Unlock()

	if started {
		return nil
	}

	if err := startBLE(clientID, deviceID); err != nil {
		return err
	}

	started = true
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
	return nil
}

func Publish(payload []byte) error {
	mu.Lock()
	on := started
	mu.Unlock()

	if !on {
		return nil
	}
	return publishBLE(payload)
}

func SetOnMessageCallback(cb func()) {
	mu.Lock()
	defer mu.Unlock()
	callback = cb
}

func GetLatestMessage() []byte {
	mu.Lock()
	defer mu.Unlock()
	return lastMessage
}

// exported to cgo layer
func onMessage(deviceID string, payload []byte) {
	mu.Lock()
	lastMessage = payload
	cb := callback
	mu.Unlock()

	if cb != nil {
		cb()
	}
}

func ClearMsg() {
	mu.Lock()
	defer mu.Unlock()

	lastMessage = nil
}

// func PublishChunked(payload []byte) error {
// 	// assign a uuid for the message
// 	msgID := make([]byte, 8)

// 	_, err := rand.Read(msgID)
// 	if err != nil {
// 		return err
// 	}
// 	totalChunks := (len(payload) + MAX_CHUNK_SIZE - 1) / MAX_CHUNK_SIZE

// 	// each chunk of the payload gets an increasing id
// 	for i := 0; i < totalChunks; i++ {
// 		start := i * MAX_CHUNK_SIZE
// 		end := start + MAX_CHUNK_SIZE
// 		if end > len(payload) {
// 			end = len(payload)
// 		}

// 		chunk := payload[start:end]

// 		buff := &bytes.Buffer{}

// 		// 	msgID 	[8]byte
// 		buff.Write(msgID)

// 		// 	seq 	uint16
// 		binary.Write(buff, binary.BigEndian, uint16(i))

// 		buff.Write(chunk)
// 	}
// 	publishBLE(payload)

// 	return nil
// }
