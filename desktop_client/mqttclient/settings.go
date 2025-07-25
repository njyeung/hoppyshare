package mqttclient

import (
	"desktop_client/config"
	"encoding/json"
)

var (
	// Settings sent by lambda
	nickname         string  = "Unnamed Device" // device nickname
	enabled          bool    = true             // ignore messages on this device
	auto_copy        bool    = false            // auto copies to clipboard
	auto_paste       bool    = false            // auto pastes, only active when auto_copy is true
	cache_time       int     = 30               // time in seconds that messages are cached MAX 5 mins (BLE bucket size)
	hotkey           string  = ""               // sends whatever is in the clipboard
	enable_hotkey    bool    = false            // enables hotkey
	notification_vol float32 = 1                // volume of notification sound
	send_to_self     bool    = true             // mqtt subscribes to itself
	startup          bool    = true             // auto startup
	destroy          bool    = false            // quits and removes itself when true
)

type DeviceSettings struct {
	DeviceID string `json:"deviceid"`
	Settings struct {
		Nickname        *string  `json:"nickname,omitempty"`
		Enabled         *bool    `json:"enabled,omitempty"`
		AutoCopy        *bool    `json:"auto_copy,omitempty"`
		AutoPaste       *bool    `json:"auto_paste,omitempty"`
		CacheTime       *int     `json:"cache_time,omitempty"`
		Hotkey          *string  `json:"hotkey,omitempty"`
		EnableHotkey    *bool    `json:"enable_hotkey,omitempty"`
		NotificationVol *float32 `json:"notification_vol,omitempty"`
		SendToSelf      *bool    `json:"send_to_self,omitempty"`
		Startup         *bool    `json:"startup,omitempty"`
		Destroy         *bool    `json:"destroy,omitempty"`
	} `json:"settings"`
}

func ParseSettings(data []byte) error {
	var allSettings []DeviceSettings
	err := json.Unmarshal(data, &allSettings)
	if err != nil {
		return err
	}

	for _, d := range allSettings {
		if d.DeviceID == config.DeviceID {
			s := d.Settings
			if s.Nickname != nil {
				nickname = *s.Nickname
			}
			if s.Enabled != nil {
				enabled = *s.Enabled
			}
			if s.AutoCopy != nil {
				auto_copy = *s.AutoCopy
			}
			if s.AutoPaste != nil {
				auto_paste = *s.AutoPaste
			}
			if s.CacheTime != nil {
				cache_time = *s.CacheTime
			}
			if s.Hotkey != nil {
				hotkey = *s.Hotkey
			}
			if s.EnableHotkey != nil {
				enable_hotkey = *s.EnableHotkey
			}
			if s.NotificationVol != nil {
				notification_vol = *s.NotificationVol
			}
			if s.SendToSelf != nil {
				send_to_self = *s.SendToSelf
			}
			if s.Startup != nil {
				startup = *s.Startup
			}
			if s.Destroy != nil {
				destroy = *s.Destroy
			}
		}
	}

	return nil
}
