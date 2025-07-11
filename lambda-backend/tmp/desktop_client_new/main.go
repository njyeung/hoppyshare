package main

import (
	"desktop_client/mqttclient"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/atotto/clipboard"
	"github.com/getlantern/systray"
	"github.com/sqweek/dialog"
)

var clientID string
var err error

func main() {
	fmt.Printf("SnapNotes PID: %d\n", os.Getpid())

	clientID, err = mqttclient.Connect()
	if err != nil {
		log.Printf("Full error with mqtt: %v", err)
	}
	systray.Run(onReady, onExit)
}

func onReady() {
	// systray.SetIcon(myIcon)
	systray.SetTitle("SnapNotes")
	systray.SetTooltip("SnapNotes Agent")
	mSendClipboard := systray.AddMenuItem("Send Clipboard", "Send clipboard contents")
	mSendFile := systray.AddMenuItem("Send File", "Send an file")
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")

	go func() {
		for {
			select {
			case <-mSendClipboard.ClickedCh:
				log.Println("Sending clipboard contents")
				go PublishClipboard()
			case <-mSendFile.ClickedCh:
				log.Println("Sending file...")
				go PublishFile()
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	// Cleanup here if needed
}

func PublishClipboard() {
	content, err := clipboard.ReadAll()
	if err != nil {
		log.Printf("Clipboard read error: %v", err)
		return
	}

	topic := fmt.Sprintf("users/%s/notes", clientID)
	mqttclient.Publish(topic, []byte(content), "text/plain", "")
}

func PublishFile() {
	filePath, err := dialog.File().Title("Select a File").Load()
	if err != nil {
		log.Printf("File selection failed: %v", err)
		return
	}

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Failed to read file: %v", err)
		return
	}

	fileName := filepath.Base(filePath)
	ext := filepath.Ext(fileName)

	// crude MIME type guess
	mime := map[string]string{
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
		".pdf":  "application/pdf",
	}[ext]

	if mime == "" {
		mime = "application/octet-stream"
	}

	topic := fmt.Sprintf("users/%s/notes", clientID)
	mqttclient.Publish(topic, fileBytes, mime, fileName)
}
