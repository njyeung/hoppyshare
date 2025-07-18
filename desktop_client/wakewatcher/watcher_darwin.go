package wakewatcher

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

__attribute__((constructor))
static void setup() {
    [[[NSWorkspace sharedWorkspace] notificationCenter]
        addObserverForName:NSWorkspaceDidWakeNotification
        object:nil
        queue:[NSOperationQueue mainQueue]
        usingBlock:^(NSNotification *note) {
            extern void handleWakeEvent();
            handleWakeEvent();
        }];
}
*/
import "C"

import (
	"sync"
	"time"
)

var (
	WakeCallback func()
	lastWakeTime time.Time
	wakeMutex    sync.Mutex
)

//export handleWakeEvent
func handleWakeEvent() {
	wakeMutex.Lock()
	now := time.Now()

	// Debounce wake cuz macOS sends 2 for some reason?
	if now.Sub(lastWakeTime) < 2*time.Second {
		wakeMutex.Unlock()
		return
	}
	lastWakeTime = now
	wakeMutex.Unlock()

	if WakeCallback != nil {
		go WakeCallback()
	}
}
