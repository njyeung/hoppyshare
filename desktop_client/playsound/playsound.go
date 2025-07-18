package playsound

// Play plays the notification sound (platform-specific implementation).
func Play(notificationSound []byte) {
	play(notificationSound)
}
