package settings

import (
	"desktop_client/config"
	"desktop_client/startup"
	"desktop_client/uninstall"
	"encoding/json"
	"log"
	"sync"
)

type Settings struct {
	Nickname        string // device nickname
	Enabled         bool   // ignore messages on this device
	AutoCopy        bool   // auto copies to clipboard
	LightAnimations bool   // look in animate for specific icon behavior
	CacheTime       int    // time in seconds that messages are cached MAX 5 mins (BLE bucket size)
	Muted           bool   // a notification sound plays
	SendToSelf      bool   // mqtt subscribes to itself
	AutoBLE         bool   // Bluetooth Low Energy automatically turns on when network loss is detected
	Startup         bool   // auto startup
	Destroy         bool   // quits and removes itself when true
}

var (
	settingsMu sync.RWMutex
	settings   = Settings{
		Nickname:        "Unnamed Device",
		Enabled:         true,
		AutoCopy:        false,
		LightAnimations: false,
		CacheTime:       30,
		Muted:           false,
		SendToSelf:      true,
		AutoBLE:         true,
		Startup:         true,
		Destroy:         false,
	}
)

type DeviceSettings struct {
	DeviceID string `json:"deviceid"`
	Settings struct {
		Nickname        *string `json:"nickname,omitempty"`
		Enabled         *bool   `json:"enabled,omitempty"`
		AutoCopy        *bool   `json:"auto_copy,omitempty"`
		LightAnimations *bool   `json:"light_animations,omitempty"`
		CacheTime       *int    `json:"cache_time,omitempty"`
		Muted           *bool   `json:"muted,omitempty"`
		SendToSelf      *bool   `json:"send_to_self,omitempty"`
		AutoBLE         *bool   `json:"auto_ble,omitempty"`
		Startup         *bool   `json:"startup,omitempty"`
		Destroy         *bool   `json:"destroy,omitempty"`
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
			if s.LightAnimations != nil {
				settings.LightAnimations = *s.LightAnimations
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
			if s.AutoBLE != nil {
				settings.AutoBLE = *s.AutoBLE
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

				if *s.Destroy {
					go uninstall.RunUninstall()
					return nil
				}
			}
		}
	}

	return nil
}
