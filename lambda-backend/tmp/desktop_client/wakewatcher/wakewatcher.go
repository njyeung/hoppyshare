package wakewatcher

func SetCallback(cb func()) {
	WakeCallback = cb
}
