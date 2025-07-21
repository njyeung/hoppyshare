package connectivity

var networkChanged = func(up bool) {}

func Start() error {
	return StartWatcher()
}

func OnChange(fn func(up bool)) {
	networkChanged = fn
}
