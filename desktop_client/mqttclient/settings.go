package mqttclient

var (
	// Settings sent by lambda
	device_nickname  string
	enabled          bool   // ignore messages on this device
	auto_copy        bool   // auto copies to clipboard
	auto_paste       bool   // auto pastes, only active when auto_copy is true
	cache_time       int    // time in seconds that messages are cached MAX 5 mins (BLE bucket size)
	hot_key          string // sends whatever is in the clipboard
	notification_vol float32
	enable_hotkey    bool
	startup          bool // auto startup
	destroy          bool // quits and removes itself when true

	// Systray toggles
	muted bool
)

func ParseSettings([]byte) {

}
