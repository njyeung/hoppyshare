package systrayhelpers

func SetIcon(icon []byte) {
	setTrayIcon(icon)
}

func SetTitle(title string) {
	setTrayTitle(title)
}

func SetTooltip(tooltip string) {
	setTrayTooltip(tooltip)
}
