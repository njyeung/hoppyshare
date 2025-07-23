//go:build linux

package ble

import (
	"fmt"
	"hash/fnv"
	"log"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
)

type linuxBLE struct {
	conn               *dbus.Conn
	clientID           string
	deviceID           string
	serviceUUID        string
	characteristicUUID string

	// Peripheral components
	advertisement  dbus.BusObject
	service        dbus.BusObject
	characteristic dbus.BusObject
	isAdvertising  bool

	// Central components
	adapter               dbus.BusObject
	discoveredDevices     map[string]dbus.BusObject
	connectedDevices      map[string]dbus.BusObject
	remoteCharacteristics map[string]dbus.BusObject

	mu      sync.Mutex
	started bool
}

var linuxBLEInstance *linuxBLE

func startBLE(clientID, deviceID string) error {
	if linuxBLEInstance != nil && linuxBLEInstance.started {
		return nil
	}

	instance, err := newLinuxBLE(clientID, deviceID)
	if err != nil {
		return fmt.Errorf("failed to create Linux BLE: %v", err)
	}

	err = instance.start()
	if err != nil {
		return fmt.Errorf("failed to start Linux BLE: %v", err)
	}

	linuxBLEInstance = instance
	log.Printf("Linux BLE started")
	return nil
}

func stopBLE() error {
	if linuxBLEInstance == nil {
		return nil
	}

	err := linuxBLEInstance.stop()
	linuxBLEInstance = nil
	log.Printf("Linux BLE stopped")
	return err
}

func publishBLE(payload []byte) error {
	if linuxBLEInstance == nil {
		return fmt.Errorf("Linux BLE not started")
	}
	return linuxBLEInstance.sendData(payload)
}

func newLinuxBLE(clientID, deviceID string) (*linuxBLE, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to system bus: %v", err)
	}

	// Generate service UUID from clientID hash (like BLEBridge.m)
	h := fnv.New32a()
	h.Write([]byte(clientID))
	shortHash := uint16(h.Sum32() & 0xFFFF)
	serviceUUID := fmt.Sprintf("0000%04x-0000-1000-8000-00805f9b34fb", shortHash)

	return &linuxBLE{
		conn:                  conn,
		clientID:              clientID,
		deviceID:              deviceID,
		serviceUUID:           serviceUUID,
		characteristicUUID:    "0000fff1-0000-1000-8000-00805f9b34fb",
		discoveredDevices:     make(map[string]dbus.BusObject),
		connectedDevices:      make(map[string]dbus.BusObject),
		remoteCharacteristics: make(map[string]dbus.BusObject),
	}, nil
}

func (l *linuxBLE) start() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.started {
		return nil
	}

	// Get BlueZ adapter
	l.adapter = l.conn.Object("org.bluez", dbus.ObjectPath("/org/bluez/hci0"))

	// Start as peripheral (advertiser)
	err := l.startPeripheral()
	if err != nil {
		log.Printf("Failed to start peripheral: %v", err)
	}

	// Start as central (scanner)
	err = l.startCentral()
	if err != nil {
		log.Printf("Failed to start central: %v", err)
	}

	l.started = true
	return nil
}

func (l *linuxBLE) stop() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.started {
		return nil
	}

	// Stop advertising
	if l.isAdvertising {
		l.stopAdvertising()
	}

	// Stop scanning
	l.adapter.Call("org.bluez.Adapter1.StopDiscovery", 0)

	// Disconnect from all devices
	for _, device := range l.connectedDevices {
		device.Call("org.bluez.Device1.Disconnect", 0)
	}

	l.discoveredDevices = make(map[string]dbus.BusObject)
	l.connectedDevices = make(map[string]dbus.BusObject)
	l.remoteCharacteristics = make(map[string]dbus.BusObject)

	l.started = false
	return nil
}

func (l *linuxBLE) startPeripheral() error {
	// Register GATT application
	err := l.registerGATTApplication()
	if err != nil {
		return fmt.Errorf("failed to register GATT application: %v", err)
	}

	// Start advertising
	err = l.startAdvertising()
	if err != nil {
		return fmt.Errorf("failed to start advertising: %v", err)
	}

	log.Printf("Linux peripheral started")
	return nil
}

func (l *linuxBLE) registerGATTApplication() error {
	// This would involve implementing BlueZ GATT D-Bus interface
	// For brevity, this is a simplified version
	// Real implementation needs to export D-Bus objects for service/characteristic

	appPath := dbus.ObjectPath("/com/example/ble/app")
	servicePath := dbus.ObjectPath("/com/example/ble/app/service0")
	charPath := dbus.ObjectPath("/com/example/ble/app/service0/char0")

	// Export service object
	l.conn.Export(&gattService{
		uuid: l.serviceUUID,
	}, servicePath, "org.bluez.GattService1")

	// Export characteristic object
	l.conn.Export(&gattCharacteristic{
		uuid:     l.characteristicUUID,
		service:  servicePath,
		instance: l,
	}, charPath, "org.bluez.GattCharacteristic1")

	// Export application object
	l.conn.Export(&gattApplication{
		services: []dbus.ObjectPath{servicePath},
	}, appPath, "org.freedesktop.DBus.ObjectManager")

	// Register application with BlueZ
	gattManager := l.conn.Object("org.bluez", dbus.ObjectPath("/org/bluez/hci0"))
	call := gattManager.Call("org.bluez.GattManager1.RegisterApplication", 0, appPath, map[string]dbus.Variant{})

	return call.Err
}

func (l *linuxBLE) startAdvertising() error {
	advPath := dbus.ObjectPath("/com/example/ble/adv")

	// Export advertisement object
	l.conn.Export(&advertisementObject{
		serviceUUIDs: []string{l.serviceUUID},
		localName:    l.deviceID,
	}, advPath, "org.bluez.LEAdvertisement1")

	// Register advertisement with BlueZ
	advManager := l.conn.Object("org.bluez", dbus.ObjectPath("/org/bluez/hci0"))
	call := advManager.Call("org.bluez.LEAdvertisingManager1.RegisterAdvertisement", 0, advPath, map[string]dbus.Variant{})

	if call.Err != nil {
		return call.Err
	}

	l.isAdvertising = true
	log.Printf("Linux advertising started as %s", l.deviceID)
	return nil
}

func (l *linuxBLE) stopAdvertising() error {
	if !l.isAdvertising {
		return nil
	}

	advPath := dbus.ObjectPath("/com/example/ble/adv")
	advManager := l.conn.Object("org.bluez", dbus.ObjectPath("/org/bluez/hci0"))
	call := advManager.Call("org.bluez.LEAdvertisingManager1.UnregisterAdvertisement", 0, advPath)

	l.isAdvertising = false
	return call.Err
}

func (l *linuxBLE) startCentral() error {
	// Set discovery filter for our service UUID
	filter := map[string]dbus.Variant{
		"UUIDs": dbus.MakeVariant([]string{l.serviceUUID}),
	}

	call := l.adapter.Call("org.bluez.Adapter1.SetDiscoveryFilter", 0, filter)
	if call.Err != nil {
		log.Printf("Failed to set discovery filter: %v", call.Err)
	}

	// Start discovery
	call = l.adapter.Call("org.bluez.Adapter1.StartDiscovery", 0)
	if call.Err != nil {
		return fmt.Errorf("failed to start discovery: %v", call.Err)
	}

	// Listen for device discoveries
	go l.listenForDevices()

	log.Printf("Linux central started")
	return nil
}

func (l *linuxBLE) listenForDevices() {
	// Set up D-Bus signal matching for device discoveries
	l.conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		"type='signal',interface='org.freedesktop.DBus.Properties',member='PropertiesChanged',arg0='org.bluez.Device1'")

	c := make(chan *dbus.Signal, 10)
	l.conn.Signal(c)

	for sig := range c {
		if sig.Name == "org.freedesktop.DBus.Properties.PropertiesChanged" {
			go l.handleDeviceSignal(sig)
		}
	}
}

func (l *linuxBLE) handleDeviceSignal(sig *dbus.Signal) {
	if len(sig.Body) < 2 {
		return
	}

	devicePath := string(sig.Path)
	if !strings.Contains(devicePath, "/org/bluez/hci0/dev_") {
		return
	}

	// Check if this is a device we want to connect to
	device := l.conn.Object("org.bluez", sig.Path)

	var name dbus.Variant
	device.Call("org.freedesktop.DBus.Properties.Get", 0, "org.bluez.Device1", "Name").Store(&name)

	deviceName, ok := name.Value().(string)
	if !ok || deviceName == l.deviceID {
		return // Ignore self or unnamed devices
	}

	if deviceName != "DesktopClient" && !strings.HasPrefix(deviceName, "DC-") {
		return // Not our target device
	}

	log.Printf("Discovered device: %s", deviceName)
	go l.connectToDevice(device, devicePath)
}

func (l *linuxBLE) connectToDevice(device dbus.BusObject, devicePath string) {
	// Connect to device
	call := device.Call("org.bluez.Device1.Connect", 0)
	if call.Err != nil {
		log.Printf("Failed to connect to device: %v", call.Err)
		return
	}

	l.mu.Lock()
	l.connectedDevices[devicePath] = device
	l.mu.Unlock()

	// Find our service and characteristic
	go l.discoverServices(device, devicePath)
}

func (l *linuxBLE) discoverServices(device dbus.BusObject, devicePath string) {
	// Get all objects to find services
	objManager := l.conn.Object("org.bluez", dbus.ObjectPath("/"))
	var objects map[dbus.ObjectPath]map[string]map[string]dbus.Variant

	call := objManager.Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0)
	if call.Err != nil {
		log.Printf("Failed to get managed objects: %v", call.Err)
		return
	}
	call.Store(&objects)

	// Find our service
	for path, interfaces := range objects {
		if !strings.HasPrefix(string(path), devicePath) {
			continue
		}

		if serviceProps, exists := interfaces["org.bluez.GattService1"]; exists {
			if uuid, ok := serviceProps["UUID"].Value().(string); ok && uuid == l.serviceUUID {
				// Found our service, now find the characteristic
				l.findCharacteristic(path, objects, devicePath)
				return
			}
		}
	}
}

func (l *linuxBLE) findCharacteristic(servicePath dbus.ObjectPath, objects map[dbus.ObjectPath]map[string]map[string]dbus.Variant, devicePath string) {
	for path, interfaces := range objects {
		if !strings.HasPrefix(string(path), string(servicePath)) {
			continue
		}

		if charProps, exists := interfaces["org.bluez.GattCharacteristic1"]; exists {
			if uuid, ok := charProps["UUID"].Value().(string); ok && uuid == l.characteristicUUID {
				// Found our characteristic
				char := l.conn.Object("org.bluez", path)

				l.mu.Lock()
				l.remoteCharacteristics[devicePath] = char
				l.mu.Unlock()

				// Subscribe to notifications
				call := char.Call("org.bluez.GattCharacteristic1.StartNotify", 0)
				if call.Err != nil {
					log.Printf("Failed to start notifications: %v", call.Err)
				} else {
					log.Printf("Subscribed to notifications from %s", devicePath)
				}

				// Listen for value changes
				go l.listenForValueChanges(char, devicePath)
				return
			}
		}
	}
}

func (l *linuxBLE) listenForValueChanges(char dbus.BusObject, devicePath string) {
	// Listen for characteristic value changes
	l.conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		fmt.Sprintf("type='signal',interface='org.freedesktop.DBus.Properties',member='PropertiesChanged',path='%s'", char.Path()))

	c := make(chan *dbus.Signal, 10)
	l.conn.Signal(c)

	for sig := range c {
		if sig.Path == char.Path() && sig.Name == "org.freedesktop.DBus.Properties.PropertiesChanged" {
			if len(sig.Body) >= 2 {
				if changed, ok := sig.Body[1].(map[string]dbus.Variant); ok {
					if value, exists := changed["Value"]; exists {
						if data, ok := value.Value().([]byte); ok {
							log.Printf("Received %d bytes from %s", len(data), devicePath)
							onMessage(devicePath, data)
						}
					}
				}
			}
		}
	}
}

func (l *linuxBLE) sendData(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	// Send to all connected remote characteristics
	l.mu.Lock()
	chars := make(map[string]dbus.BusObject)
	for k, v := range l.remoteCharacteristics {
		chars[k] = v
	}
	l.mu.Unlock()

	if len(chars) == 0 {
		return fmt.Errorf("no connected devices")
	}

	for devicePath, char := range chars {
		call := char.Call("org.bluez.GattCharacteristic1.WriteValue", 0, data, map[string]dbus.Variant{})
		if call.Err != nil {
			log.Printf("Failed to write to %s: %v", devicePath, call.Err)
		} else {
			log.Printf("Sent %d bytes to %s", len(data), devicePath)
		}
	}

	return nil
}

// D-Bus exported objects for GATT server
type gattApplication struct {
	services []dbus.ObjectPath
}

func (g *gattApplication) GetManagedObjects() (map[dbus.ObjectPath]map[string]map[string]dbus.Variant, *dbus.Error) {
	// Return service objects
	return make(map[dbus.ObjectPath]map[string]map[string]dbus.Variant), nil
}

type gattService struct {
	uuid string
}

func (g *gattService) GetAll(interface_name string) (map[string]dbus.Variant, *dbus.Error) {
	if interface_name != "org.bluez.GattService1" {
		return nil, dbus.NewError("org.freedesktop.DBus.Error.UnknownInterface", nil)
	}

	return map[string]dbus.Variant{
		"UUID":    dbus.MakeVariant(g.uuid),
		"Primary": dbus.MakeVariant(true),
	}, nil
}

type gattCharacteristic struct {
	uuid     string
	service  dbus.ObjectPath
	instance *linuxBLE
}

func (g *gattCharacteristic) GetAll(interface_name string) (map[string]dbus.Variant, *dbus.Error) {
	if interface_name != "org.bluez.GattCharacteristic1" {
		return nil, dbus.NewError("org.freedesktop.DBus.Error.UnknownInterface", nil)
	}

	return map[string]dbus.Variant{
		"UUID":    dbus.MakeVariant(g.uuid),
		"Service": dbus.MakeVariant(g.service),
		"Flags":   dbus.MakeVariant([]string{"read", "write", "write-without-response", "notify"}),
	}, nil
}

func (g *gattCharacteristic) ReadValue(options map[string]dbus.Variant) ([]byte, *dbus.Error) {
	return []byte{}, nil
}

func (g *gattCharacteristic) WriteValue(value []byte, options map[string]dbus.Variant) *dbus.Error {
	if len(value) > 0 {
		log.Printf("Received %d bytes via GATT write", len(value))
		onMessage("local", value)
	}
	return nil
}

type advertisementObject struct {
	serviceUUIDs []string
	localName    string
}

func (a *advertisementObject) GetAll(interface_name string) (map[string]dbus.Variant, *dbus.Error) {
	if interface_name != "org.bluez.LEAdvertisement1" {
		return nil, dbus.NewError("org.freedesktop.DBus.Error.UnknownInterface", nil)
	}

	return map[string]dbus.Variant{
		"Type":         dbus.MakeVariant("peripheral"),
		"ServiceUUIDs": dbus.MakeVariant(a.serviceUUIDs),
		"LocalName":    dbus.MakeVariant(a.localName),
	}, nil
}

func (a *advertisementObject) Release() *dbus.Error {
	return nil
}
