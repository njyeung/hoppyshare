//go:build !linux

package systrayhelpers

import "github.com/getlantern/systray"

func setTrayIcon(icon []byte) {
	systray.SetIcon(icon)
}

func setTrayTitle(title string) {
	systray.SetTitle(title)
}

func setTrayTooltip(tooltip string) {
	systray.SetTooltip(tooltip)
}
