package playsound

import "desktop_client/settings"

// Play plays the notification sound (platform-specific implementation).
func Play(notificationSound []byte) {
	play(notificationSound, settings.GetSettings().NotificationVol)
}
