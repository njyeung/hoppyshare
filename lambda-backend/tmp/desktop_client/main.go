package main

import (
	"log"
	"desktop_client/mqttclient"
	"github.com/getlantern/systray"
    _ "embed"
)

//go:embed mqttclient/lambda_output/device_id.txt
var deviceIdRaw []byte

func main() {
    deviceID := string(deviceIdRaw)
	_ = deviceID
    go mqttclient.RunMQTT()

	systray.Run(onReady, onExit)
}

func onReady() {
	// systray.SetIcon(myIcon)
	systray.SetTitle("SnapNotes")
	systray.SetTooltip("SnapNotes Agent")

	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")

	go func() {
		<-mQuit.ClickedCh
		log.Println("Exiting...")
		systray.Quit()
	}()
}

func onExit() {
	// Cleanup here if needed
}