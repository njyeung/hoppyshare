//go:build darwin
// +build darwin

package ble

/*
#cgo CFLAGS: -I${SRCDIR} -x objective-c
#cgo LDFLAGS: -framework Foundation -framework CoreBluetooth
#import "BLEBridge.h"
*/
import "C"
import "unsafe"

func startBLE(clientID, deviceID string) error {
	cClient := C.CString(clientID)
	defer C.free(unsafe.Pointer(cClient))
	cDev := C.CString(deviceID)
	defer C.free(unsafe.Pointer(cDev))

	C.BLEBridgeStart(cClient, cDev)
	return nil
}

func stopBLE() error {
	C.BLEBridgeStop()
	return nil
}

func publishBLE(payload []byte) error {
	if len(payload) == 0 {
		return nil
	}
	cData := C.CBytes(payload)
	defer C.free(cData)

	C.BLEBridgeSend(cData, C.int(len(payload)))
	return nil
}

//export GoOnBLEMessage
func GoOnBLEMessage(deviceID *C.char, data *C.char, length C.int) {
	if onMessage != nil {
		id := C.GoString(deviceID)
		body := C.GoBytes(unsafe.Pointer(data), length)
		onMessage(id, body)
	}
}
