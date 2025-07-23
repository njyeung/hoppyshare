//go:build darwin
// +build darwin

package connectivity

/*
#cgo LDFLAGS: -framework SystemConfiguration -framework CoreFoundation
#include <stdlib.h>
#include <stdbool.h>
#include <SystemConfiguration/SystemConfiguration.h>
#include <netinet/in.h>

// forward the Go callback
extern void goNotifyNetworkChange(bool up);

// reachability reference
static SCNetworkReachabilityRef reachRef = NULL;

static void reachabilityCallback(
    SCNetworkReachabilityRef target,
    SCNetworkReachabilityFlags flags,
    void *info
) {
    bool up = (flags & kSCNetworkFlagsReachable) != 0 && (flags & kSCNetworkFlagsConnectionRequired) == 0;
    goNotifyNetworkChange(up);
}

static void StartWatcher() {
    if (reachRef != NULL) return;

    struct sockaddr_in addr;
    bzero(&addr, sizeof(addr));
    addr.sin_len = sizeof(addr);
    addr.sin_family = AF_INET;

    reachRef = SCNetworkReachabilityCreateWithAddress(NULL, (const struct sockaddr*)&addr);
    if (reachRef == NULL) return;

    SCNetworkReachabilityContext ctx = { 0, NULL, NULL, NULL, NULL };
    if (SCNetworkReachabilitySetCallback(reachRef, reachabilityCallback, &ctx)) {
        SCNetworkReachabilityScheduleWithRunLoop(reachRef, CFRunLoopGetMain(), kCFRunLoopDefaultMode);
    }
}
*/
import "C"

func StartWatcher() error {
	C.StartWatcher()
	return nil
}

//export goNotifyNetworkChange
func goNotifyNetworkChange(up C.bool) {
	networkChanged(bool(up))
}
