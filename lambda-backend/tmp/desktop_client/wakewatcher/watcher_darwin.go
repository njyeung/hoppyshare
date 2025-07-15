//go:build darwin

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
            NSLog(@"Received wake notification");
            extern void handleWakeEvent();
            handleWakeEvent();
        }];
}
*/
import "C"

//export handleWakeEvent
func handleWakeEvent() {
	go WakeCallback()
}

var WakeCallback func()
