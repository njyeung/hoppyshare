//go:build linux

package ble

import (
	"fmt"
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
	gattManager       dbus.BusObject
	advManager        dbus.BusObject
	advertisement     dbus.BusObject
	application       dbus.BusObject
	service           dbus.BusObject
	characteristic    dbus.BusObject
	isAdvertising     bool
	subscribedClients map[string]bool

	// Central components
	adapter               dbus.BusObject
	discoveredDevices     map[string]dbus.BusObject
	connectedDevices      map[string]dbus.BusObject
	remoteCharacteristics map[string]dbus.BusObject
	isScanning            bool

	mu      sync.RWMutex
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
	log.Printf("Linux BLE started with service UUID: %s", instance.serviceUUID)
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

// Use same hash algorithm as BLEBridge.m and BLEBridge_windows.cpp
func simpleHash(str string) uint32 {
	var hash uint32 = 0
	for _, c := range []byte(str) {
		hash = hash*31 + uint32(c)
	}
	return hash
}

func newLinuxBLE(clientID, deviceID string) (*linuxBLE, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to system bus: %v", err)
	}

	// Generate service UUID from clientID hash (matching BLEBridge.m exactly)
	hash := simpleHash(clientID)
	shortHash := uint16(hash & 0xFFFF)
	serviceUUID := fmt.Sprintf("0000%04x-0000-1000-8000-00805f9b34fb", shortHash)

	log.Printf("Generated service UUID hash: %d, shortHash: %d", hash, shortHash)

	return &linuxBLE{
		conn:                  conn,
		clientID:              clientID,
		deviceID:              deviceID,
		serviceUUID:           serviceUUID,
		characteristicUUID:    "0000fff1-0000-1000-8000-00805f9b34fb",
		discoveredDevices:     make(map[string]dbus.BusObject),
		connectedDevices:      make(map[string]dbus.BusObject),
		remoteCharacteristics: make(map[string]dbus.BusObject),
		subscribedClients:     make(map[string]bool),
	}, nil
}

func (l *linuxBLE) start() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.started {
		return nil
	}

	// Get BlueZ objects
	l.adapter = l.conn.Object("org.bluez", dbus.ObjectPath("/org/bluez/hci0"))
	l.gattManager = l.conn.Object("org.bluez", dbus.ObjectPath("/org/bluez/hci0"))
	l.advManager = l.conn.Object("org.bluez", dbus.ObjectPath("/org/bluez/hci0"))

	log.Printf("Starting BLE with %s / %s", l.serviceUUID, l.deviceID)

	// Start as peripheral (advertiser) - matches macOS/Windows behavior
	err := l.startPeripheral()
	if err != nil {
		log.Printf("Failed to start peripheral: %v", err)
		// Continue anyway
	}

	// Start as central (scanner) - matches macOS/Windows behavior
	err = l.startCentral()
	if err != nil {
		log.Printf("Failed to start central: %v", err)
		// Continue anyway
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

	log.Printf("Stopping BLE")

	// Stop advertising
	if l.isAdvertising {
		l.stopAdvertising()
	}

	// Stop scanning
	if l.isScanning {
		l.adapter.Call("org.bluez.Adapter1.StopDiscovery", 0)
		l.isScanning = false
	}

	// Disconnect from all devices
	for _, device := range l.connectedDevices {
		device.Call("org.bluez.Device1.Disconnect", 0)
	}

	// Clean up maps
	l.discoveredDevices = make(map[string]dbus.BusObject)
	l.connectedDevices = make(map[string]dbus.BusObject)
	l.remoteCharacteristics = make(map[string]dbus.BusObject)
	l.subscribedClients = make(map[string]bool)

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

	log.Printf("Peripheral started")
	return nil
}

func (l *linuxBLE) registerGATTApplication() error {
	appPath := dbus.ObjectPath("/com/desktopClient/ble/app")
	servicePath := dbus.ObjectPath("/com/desktopClient/ble/app/service0")
	charPath := dbus.ObjectPath("/com/desktopClient/ble/app/service0/char0")

	// Create and export application
	app := &gattApplication{
		services: []dbus.ObjectPath{servicePath},
		path:     appPath,
		conn:     l.conn,
	}
	err := l.conn.Export(app, appPath, "org.freedesktop.DBus.ObjectManager")
	if err != nil {
		return fmt.Errorf("failed to export application: %v", err)
	}

	// Create and export service
	service := &gattService{
		uuid: l.serviceUUID,
		path: servicePath,
		conn: l.conn,
	}
	err = l.conn.ExportMethodTable(map[string]interface{}{
		"GetAll":           service.GetAll,
		"Get":              service.Get,
		"Set":              service.Set,
		"GetAllProperties": service.GetAllProperties,
	}, servicePath, "org.freedesktop.DBus.Properties")
	if err != nil {
		return fmt.Errorf("failed to export service properties: %v", err)
	}

	// Create and export characteristic
	char := &gattCharacteristic{
		uuid:     l.characteristicUUID,
		service:  servicePath,
		path:     charPath,
		instance: l,
		conn:     l.conn,
	}

	// Export characteristic properties interface
	err = l.conn.ExportMethodTable(map[string]interface{}{
		"GetAll":           char.GetAll,
		"Get":              char.Get,
		"Set":              char.Set,
		"GetAllProperties": char.GetAllProperties,
	}, charPath, "org.freedesktop.DBus.Properties")
	if err != nil {
		return fmt.Errorf("failed to export characteristic properties: %v", err)
	}

	// Export characteristic methods interface
	err = l.conn.ExportMethodTable(map[string]interface{}{
		"ReadValue":   char.ReadValue,
		"WriteValue":  char.WriteValue,
		"StartNotify": char.StartNotify,
		"StopNotify":  char.StopNotify,
	}, charPath, "org.bluez.GattCharacteristic1")
	if err != nil {
		return fmt.Errorf("failed to export characteristic methods: %v", err)
	}

	// Store references
	l.application = l.conn.Object("org.bluez", appPath)
	l.service = l.conn.Object("org.bluez", servicePath)
	l.characteristic = l.conn.Object("org.bluez", charPath)

	// Register application with BlueZ
	call := l.gattManager.Call("org.bluez.GattManager1.RegisterApplication", 0, appPath, map[string]dbus.Variant{})
	if call.Err != nil {
		return fmt.Errorf("failed to register GATT application: %v", call.Err)
	}

	log.Printf("GATT application registered successfully")
	return nil
}

func (l *linuxBLE) startAdvertising() error {
	advPath := dbus.ObjectPath("/com/desktopClient/ble/adv")

	// Create advertisement object
	adv := &advertisementObject{
		serviceUUIDs: []string{l.serviceUUID},
		localName:    l.deviceID,
		path:         advPath,
		conn:         l.conn,
	}

	// Export advertisement object with properties interface
	err := l.conn.ExportMethodTable(map[string]interface{}{
		"GetAll":           adv.GetAll,
		"Get":              adv.Get,
		"Set":              adv.Set,
		"GetAllProperties": adv.GetAllProperties,
	}, advPath, "org.freedesktop.DBus.Properties")
	if err != nil {
		return fmt.Errorf("failed to export advertisement properties: %v", err)
	}

	// Export advertisement methods interface
	err = l.conn.ExportMethodTable(map[string]interface{}{
		"Release": adv.Release,
	}, advPath, "org.bluez.LEAdvertisement1")
	if err != nil {
		return fmt.Errorf("failed to export advertisement methods: %v", err)
	}

	// Store reference
	l.advertisement = l.conn.Object("org.bluez", advPath)

	// Register advertisement with BlueZ
	call := l.advManager.Call("org.bluez.LEAdvertisingManager1.RegisterAdvertisement", 0, advPath, map[string]dbus.Variant{})
	if call.Err != nil {
		return fmt.Errorf("failed to register advertisement: %v", call.Err)
	}

	l.isAdvertising = true
	log.Printf("Advertising as %s", l.deviceID)
	return nil
}

func (l *linuxBLE) stopAdvertising() error {
	if !l.isAdvertising {
		return nil
	}

	advPath := dbus.ObjectPath("/com/desktopClient/ble/adv")
	call := l.advManager.Call("org.bluez.LEAdvertisingManager1.UnregisterAdvertisement", 0, advPath)

	l.isAdvertising = false
	log.Printf("Advertising stopped")
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

	l.isScanning = true
	log.Printf("Scanning for %s", l.serviceUUID)

	// Listen for device discoveries
	go l.listenForDevices()

	return nil
}

func (l *linuxBLE) listenForDevices() {
	// Set up D-Bus signal matching for device discoveries
	matchRule := "type='signal',interface='org.freedesktop.DBus.Properties',member='PropertiesChanged',arg0='org.bluez.Device1'"
	err := l.conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, matchRule).Err
	if err != nil {
		log.Printf("Failed to add signal match: %v", err)
		return
	}

	c := make(chan *dbus.Signal, 10)
	l.conn.Signal(c)

	go func() {
		for sig := range c {
			if sig.Name == "org.freedesktop.DBus.Properties.PropertiesChanged" &&
				len(sig.Body) >= 1 && sig.Body[0] == "org.bluez.Device1" {
				go l.handleDeviceSignal(sig)
			}
		}
	}()
}

func (l *linuxBLE) handleDeviceSignal(sig *dbus.Signal) {
	if len(sig.Body) < 2 {
		return
	}

	devicePath := string(sig.Path)
	if !strings.Contains(devicePath, "/org/bluez/hci0/dev_") {
		return
	}

	// Check if this device was already processed
	l.mu.RLock()
	if _, exists := l.discoveredDevices[devicePath]; exists {
		l.mu.RUnlock()
		return
	}
	l.mu.RUnlock()

	// Check properties changed
	changed, ok := sig.Body[1].(map[string]dbus.Variant)
	if !ok {
		return
	}

	// Look for ServicesResolved or UUIDs properties
	if servicesResolved, exists := changed["ServicesResolved"]; exists {
		if resolved, ok := servicesResolved.Value().(bool); ok && resolved {
			go l.checkAndConnectDevice(sig.Path, devicePath)
		}
	} else if _, exists := changed["UUIDs"]; exists {
		go l.checkAndConnectDevice(sig.Path, devicePath)
	}
}

func (l *linuxBLE) checkAndConnectDevice(path dbus.ObjectPath, devicePath string) {
	device := l.conn.Object("org.bluez", path)

	// Get device properties
	var name, uuids dbus.Variant
	device.Call("org.freedesktop.DBus.Properties.Get", 0, "org.bluez.Device1", "Name").Store(&name)
	device.Call("org.freedesktop.DBus.Properties.Get", 0, "org.bluez.Device1", "UUIDs").Store(&uuids)

	deviceName, _ := name.Value().(string)
	deviceUUIDs, _ := uuids.Value().([]string)

	// Skip self (matches BLEBridge.m behavior)
	if deviceName == l.deviceID {
		log.Printf("Ignoring self-advertisement")
		return
	}

	// Check if device advertises our service UUID
	hasOurService := false
	for _, uuid := range deviceUUIDs {
		if strings.EqualFold(uuid, l.serviceUUID) {
			hasOurService = true
			break
		}
	}

	if !hasOurService {
		return
	}

	log.Printf("Discovered device: %s with our service", deviceName)
	l.mu.Lock()
	l.discoveredDevices[devicePath] = device
	l.mu.Unlock()

	go l.connectToDevice(device, devicePath)
}

func (l *linuxBLE) connectToDevice(device dbus.BusObject, devicePath string) {
	// Connect to device
	call := device.Call("org.bluez.Device1.Connect", 0)
	if call.Err != nil {
		log.Printf("Failed to connect: %v", call.Err)
		l.mu.Lock()
		delete(l.discoveredDevices, devicePath)
		l.mu.Unlock()
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
			if uuid, ok := serviceProps["UUID"].Value().(string); ok &&
				strings.EqualFold(uuid, l.serviceUUID) {
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
			if uuid, ok := charProps["UUID"].Value().(string); ok &&
				strings.EqualFold(uuid, l.characteristicUUID) {
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
	matchRule := fmt.Sprintf("type='signal',interface='org.freedesktop.DBus.Properties',member='PropertiesChanged',path='%s'", char.Path())
	l.conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, matchRule)

	c := make(chan *dbus.Signal, 10)
	l.conn.Signal(c)

	for sig := range c {
		if sig.Path == char.Path() && sig.Name == "org.freedesktop.DBus.Properties.PropertiesChanged" {
			if len(sig.Body) >= 2 {
				if changed, ok := sig.Body[1].(map[string]dbus.Variant); ok {
					if value, exists := changed["Value"]; exists {
						if data, ok := value.Value().([]byte); ok {
							log.Printf("Received %d bytes from remote device", len(data))
							// Extract device name from path for callback
							deviceName := strings.Replace(devicePath, "/org/bluez/hci0/dev_", "", 1)
							deviceName = strings.Replace(deviceName, "_", ":", -1)
							onMessage(deviceName, data)
						}
					}
				}
			}
		}
	}
}

func (l *linuxBLE) sendData(data []byte) error {
	if len(data) == 0 {
		log.Printf("No data to send")
		return nil
	}

	sentToAny := false

	// Send to remote characteristics
	l.mu.RLock()
	remoteChars := make(map[string]dbus.BusObject)
	for k, v := range l.remoteCharacteristics {
		remoteChars[k] = v
	}
	l.mu.RUnlock()

	for devicePath, char := range remoteChars {
		call := char.Call("org.bluez.GattCharacteristic1.WriteValue", 0, data, map[string]dbus.Variant{})
		if call.Err != nil {
			log.Printf("Failed to write to %s: %v", devicePath, call.Err)
		} else {
			log.Printf("Sent %d bytes to remote device", len(data))
			sentToAny = true
		}
	}

	// Send to local subscribers (like BLEBridge.m does with updateValue)
	l.mu.RLock()
	hasSubscribers := len(l.subscribedClients) > 0
	l.mu.RUnlock()

	if hasSubscribers && l.characteristic != nil {
		// Notify all subscribed clients by sending a PropertiesChanged signal
		// This is equivalent to updateValue in BLEBridge.m
		go l.notifySubscribedClients(data)
		log.Printf("Sent %d bytes to local subscribers", len(data))
		sentToAny = true
	}

	if !sentToAny {
		if len(remoteChars) == 0 && !hasSubscribers {
			log.Printf("No subscribers or remote devices")
		}
		return fmt.Errorf("no recipients for data")
	}

	log.Printf("Sending %d bytes", len(data))
	return nil
}

func (l *linuxBLE) notifySubscribedClients(data []byte) {
	// Send PropertiesChanged signal for our characteristic
	// This mimics the updateValue:forCharacteristic:onSubscribedCentrals behavior
	if l.characteristic != nil {
		err := l.conn.Emit(
			l.characteristic.Path(),
			"org.freedesktop.DBus.Properties.PropertiesChanged",
			"org.bluez.GattCharacteristic1",
			map[string]dbus.Variant{
				"Value": dbus.MakeVariant(data),
			},
			[]string{},
		)
		if err != nil {
			log.Printf("Failed to emit PropertiesChanged signal: %v", err)
		}
	}
}

// GATT Application object
type gattApplication struct {
	services []dbus.ObjectPath
	path     dbus.ObjectPath
	conn     *dbus.Conn
}

// GATT Service object
type gattService struct {
	uuid string
	path dbus.ObjectPath
	conn *dbus.Conn
}

// GATT Characteristic object
type gattCharacteristic struct {
	uuid     string
	service  dbus.ObjectPath
	path     dbus.ObjectPath
	instance *linuxBLE
	conn     *dbus.Conn
	value    []byte
	mu       sync.RWMutex
}

func (g *gattApplication) GetManagedObjects() (map[dbus.ObjectPath]map[string]map[string]dbus.Variant, *dbus.Error) {
	objects := make(map[dbus.ObjectPath]map[string]map[string]dbus.Variant)

	// Add service objects
	for _, servicePath := range g.services {
		objects[servicePath] = map[string]map[string]dbus.Variant{
			"org.bluez.GattService1": {
				"UUID":    dbus.MakeVariant(""), // Will be filled by service object
				"Primary": dbus.MakeVariant(true),
			},
		}

		// Add characteristic for this service
		charPath := dbus.ObjectPath(string(servicePath) + "/char0")
		objects[charPath] = map[string]map[string]dbus.Variant{
			"org.bluez.GattCharacteristic1": {
				"UUID":    dbus.MakeVariant(""), // Will be filled by characteristic object
				"Service": dbus.MakeVariant(servicePath),
				"Flags":   dbus.MakeVariant([]string{"read", "write", "write-without-response", "notify"}),
			},
		}
	}

	return objects, nil
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

func (g *gattService) Get(interface_name, property_name string) (dbus.Variant, *dbus.Error) {
	props, err := g.GetAll(interface_name)
	if err != nil {
		return dbus.Variant{}, err
	}
	if val, ok := props[property_name]; ok {
		return val, nil
	}
	return dbus.Variant{}, dbus.NewError("org.freedesktop.DBus.Error.UnknownProperty", nil)
}

func (g *gattService) Set(interface_name, property_name string, value dbus.Variant) *dbus.Error {
	return dbus.NewError("org.freedesktop.DBus.Error.PropertyReadOnly", nil)
}

func (g *gattService) GetAllProperties(interface_name string) (map[string]dbus.Variant, *dbus.Error) {
	return g.GetAll(interface_name)
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

func (g *gattCharacteristic) Get(interface_name, property_name string) (dbus.Variant, *dbus.Error) {
	props, err := g.GetAll(interface_name)
	if err != nil {
		return dbus.Variant{}, err
	}
	if val, ok := props[property_name]; ok {
		return val, nil
	}
	return dbus.Variant{}, dbus.NewError("org.freedesktop.DBus.Error.UnknownProperty", nil)
}

func (g *gattCharacteristic) Set(interface_name, property_name string, value dbus.Variant) *dbus.Error {
	return dbus.NewError("org.freedesktop.DBus.Error.PropertyReadOnly", nil)
}

func (g *gattCharacteristic) GetAllProperties(interface_name string) (map[string]dbus.Variant, *dbus.Error) {
	return g.GetAll(interface_name)
}

func (g *gattCharacteristic) ReadValue(options map[string]dbus.Variant) ([]byte, *dbus.Error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.value, nil
}

func (g *gattCharacteristic) WriteValue(value []byte, options map[string]dbus.Variant) *dbus.Error {
	if len(value) > 0 {
		log.Printf("Received %d bytes via GATT write", len(value))
		// Call the Go callback
		onMessage("local", value)
	}
	return nil
}

func (g *gattCharacteristic) StartNotify() *dbus.Error {
	log.Printf("Client subscribed")
	// Add to subscribed clients
	g.instance.mu.Lock()
	g.instance.subscribedClients["client"] = true
	g.instance.mu.Unlock()
	return nil
}

func (g *gattCharacteristic) StopNotify() *dbus.Error {
	log.Printf("Client unsubscribed")
	// Remove from subscribed clients
	g.instance.mu.Lock()
	delete(g.instance.subscribedClients, "client")
	g.instance.mu.Unlock()
	return nil
}

// Advertisement object
type advertisementObject struct {
	serviceUUIDs []string
	localName    string
	path         dbus.ObjectPath
	conn         *dbus.Conn
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

func (a *advertisementObject) Get(interface_name, property_name string) (dbus.Variant, *dbus.Error) {
	props, err := a.GetAll(interface_name)
	if err != nil {
		return dbus.Variant{}, err
	}
	if val, ok := props[property_name]; ok {
		return val, nil
	}
	return dbus.Variant{}, dbus.NewError("org.freedesktop.DBus.Error.UnknownProperty", nil)
}

func (a *advertisementObject) Set(interface_name, property_name string, value dbus.Variant) *dbus.Error {
	return dbus.NewError("org.freedesktop.DBus.Error.PropertyReadOnly", nil)
}

func (a *advertisementObject) GetAllProperties(interface_name string) (map[string]dbus.Variant, *dbus.Error) {
	return a.GetAll(interface_name)
}

func (a *advertisementObject) Release() *dbus.Error {
	log.Printf("Advertisement released")
	return nil
}
