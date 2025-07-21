//go:build linux
// +build linux

package connectivity

import (
	"net"

	"github.com/vishvananda/netlink"
)

func StartWatcher() error {
	updates := make(chan netlink.LinkUpdate)
	done := make(chan struct{})
	if err := netlink.LinkSubscribe(updates, done); err != nil {
		return err
	}

	go func() {
		for update := range updates {
			attrs := update.Link.Attrs()
			if attrs == nil {
				continue
			}
			if attrs.Name == "lo" {
				continue
			}
			// determine new state
			up := attrs.Flags&net.FlagUp != 0
			networkChanged(up)
		}
	}()
	return nil
}
