package settings

import (
	"desktop_client/config"
	"encoding/json"
	"log"
	"sync"
)

type Settings struct {
	Nickname        string  // device nickname
	Enabled         bool    // ignore messages on this device
	AutoCopy        bool    // auto copies to clipboard
	AutoPaste       bool    // auto pastes, only active when auto_copy is true
	CacheTime       int     // time in seconds that messages are cached MAX 5 mins (BLE bucket size)
	Hotkey          string  // sends whatever is in the clipboard
	EnableHotkey    bool    // enables hotkey
	NotificationVol float32 // volume of notification sound
	SendToSelf      bool    // mqtt subscribes to itself
	BleAlwaysOff    bool    // BLE will never turn on
	Startup         bool    // auto startup
	Destroy         bool    // quits and removes itself when true
}

var (
	settingsMu sync.RWMutex
	settings   = Settings{
		Nickname:        "Unnamed Device",
		Enabled:         true,
		AutoCopy:        false,
		AutoPaste:       false, // TODO
		CacheTime:       30,
		Hotkey:          "",    // TODO
		EnableHotkey:    false, // TODO
		NotificationVol: 1,     // TODO implemented for linux
		SendToSelf:      true,
		BleAlwaysOff:    false, // TODO
		Startup:         true,  // TODO
		Destroy:         false, // TODO
	}
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
		BleAlwaysOff    *bool    `json:"ble_always_off,omitempty"`
		Startup         *bool    `json:"startup,omitempty"`
		Destroy         *bool    `json:"destroy,omitempty"`
	} `json:"settings"`
}

func GetSettings() Settings {
	settingsMu.RLock()
	defer settingsMu.RUnlock()

	return settings
}
func ParseSettings(data []byte) error {
	var allSettings []DeviceSettings

	err := json.Unmarshal(data, &allSettings)
	if err != nil {
		return err
	}

	settingsMu.Lock()
	defer settingsMu.Unlock()

	for _, d := range allSettings {
		if d.DeviceID == config.DeviceID {
			s := d.Settings
			if s.Nickname != nil {
				settings.Nickname = *s.Nickname
			}
			if s.Enabled != nil {
				settings.Enabled = *s.Enabled
			}
			if s.AutoCopy != nil {
				settings.AutoCopy = *s.AutoCopy
			}
			if s.AutoPaste != nil {
				settings.AutoPaste = *s.AutoPaste
			}
			if s.CacheTime != nil {
				settings.CacheTime = *s.CacheTime
			}
			if s.Hotkey != nil {
				settings.Hotkey = *s.Hotkey
			}
			if s.EnableHotkey != nil {
				settings.EnableHotkey = *s.EnableHotkey
			}
			if s.NotificationVol != nil {
				settings.NotificationVol = *s.NotificationVol
			}
			if s.SendToSelf != nil {
				settings.SendToSelf = *s.SendToSelf
			}
			if s.Startup != nil {
				settings.Startup = *s.Startup
			}
			if s.Destroy != nil {
				settings.Destroy = *s.Destroy
			}
		}
	}

	log.Println(settings)
	return nil
}
