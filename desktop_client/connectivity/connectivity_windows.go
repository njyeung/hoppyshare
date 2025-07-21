//go:build windows
// +build windows

package connectivity

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modiphlpapi = windows.NewLazySystemDLL("iphlpapi.dll")
	procNotify  = modiphlpapi.NewProc("NotifyIpInterfaceChange")
)

type _MibNotificationSpinLock uint64

func StartWatcher() error {
	// AF_UNSPEC = 0, watch IPv4 & IPv6, upcall into goOnChange
	r, _, e := procNotify.Call(
		uintptr(0),
		uintptr(unsafe.Pointer(windows.NewCallback(goOnChange))),
		uintptr(0),
		uintptr(0),
		uintptr(0),
	)
	if r != 0 {
		return error(e)
	}
	return nil
}

//export goOnChange
func goOnChange(context, row, changeType uintptr) uintptr {
	// changeType tells you if connectivity to Internet is gained/lost
	networkChanged(changeType == windows.MibParameterNotification)
	return 0
}
