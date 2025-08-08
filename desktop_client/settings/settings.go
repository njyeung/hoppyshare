package settings

import (
	"desktop_client/config"
	"desktop_client/startup"
	"encoding/json"
	"log"
	"sync"
)

type Settings struct {
	Nickname   string // device nickname
	Enabled    bool   // ignore messages on this device
	AutoCopy   bool   // auto copies to clipboard
	AutoPaste  bool   // auto pastes, only active when auto_copy is true
	CacheTime  int    // time in seconds that messages are cached MAX 5 mins (BLE bucket size)
	Muted      bool   // a notification sound plays
	SendToSelf bool   // mqtt subscribes to itself
	Startup    bool   // auto startup
	Destroy    bool   // quits and removes itself when true
}

var (
	settingsMu sync.RWMutex
	settings   = Settings{
		Nickname:   "Unnamed Device",
		Enabled:    true,
		AutoCopy:   false,
		AutoPaste:  false, // TODO
		CacheTime:  30,
		Muted:      false,
		SendToSelf: true,
		Startup:    true,  // TODO
		Destroy:    false, // TODO
	}
)

type DeviceSettings struct {
	DeviceID string `json:"deviceid"`
	Settings struct {
		Nickname   *string `json:"nickname,omitempty"`
		Enabled    *bool   `json:"enabled,omitempty"`
		AutoCopy   *bool   `json:"auto_copy,omitempty"`
		AutoPaste  *bool   `json:"auto_paste,omitempty"`
		CacheTime  *int    `json:"cache_time,omitempty"`
		Muted      *bool   `json:"muted,omitempty"`
		SendToSelf *bool   `json:"send_to_self,omitempty"`
		Startup    *bool   `json:"startup,omitempty"`
		Destroy    *bool   `json:"destroy,omitempty"`
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
			if s.Muted != nil {
				settings.Muted = *s.Muted
			}
			if s.SendToSelf != nil {
				settings.SendToSelf = *s.SendToSelf
			}
			if s.Startup != nil {
				oldStartup := settings.Startup
				settings.Startup = *s.Startup
				if oldStartup != *s.Startup {
					if *s.Startup {
						if err := startup.EnableStartup(); err != nil {
							log.Printf("Failed to enable startup: %v", err)
						} else {
							log.Println("Startup enabled")
						}
					} else {
						if err := startup.DisableStartup(); err != nil {
							log.Printf("Failed to disable startup: %v", err)
						} else {
							log.Println("Startup disabled")
						}
					}
				}
			}
			if s.Destroy != nil {
				settings.Destroy = *s.Destroy
			}
		}
	}

	return nil
}
