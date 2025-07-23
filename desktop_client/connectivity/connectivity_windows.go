//go:build windows
// +build windows

package connectivity

import (
	"log"
	"net"
	"time"
)

func StartWatcher() error {
	// Use a simple polling approach instead of Windows API callbacks to avoid stupid CGO
	go pollNetworkState()
	return nil
}

func pollNetworkState() {
	lastState := isNetworkUp()
	networkChanged(lastState)

	for {
		time.Sleep(5 * time.Second)
		currentState := isNetworkUp()
		if currentState != lastState {
			networkChanged(currentState)
			lastState = currentState
		}
	}
}

func isNetworkUp() bool {
	// Check if we can resolve a DNS name and have network interfaces up
	_, err := net.LookupHost("google.com")
	if err != nil {
		// Also check if we have any non-loopback interfaces up
		interfaces, err := net.Interfaces()
		if err != nil {
			return false
		}

		for _, iface := range interfaces {
			if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
				// We have an interface up, but DNS failed, so partial connectivity
				log.Printf("Interface %s is up but DNS resolution failed", iface.Name)
				return false
			}
		}
		return false
	}
	return true
}
