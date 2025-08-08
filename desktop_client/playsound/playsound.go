package playsound

import "desktop_client/settings"

// Play plays the notification sound (platform-specific implementation).
func Play(notificationSound []byte) {
	if !settings.GetSettings().Muted {
		play(notificationSound)
	}
}
