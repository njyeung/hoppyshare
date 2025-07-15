//go:build linux

package wakewatcher

//export handleWakeEvent
func handleWakeEvent() {
	// no-op
}

var WakeCallback func()