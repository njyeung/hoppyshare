//go:build windows

package wakewatcher


//export handleWakeEvent
func handleWakeEvent() {
	// no-op
}

var WakeCallback func()
