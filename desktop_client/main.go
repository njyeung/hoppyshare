package main

import (
	"desktop_client/animate"
	"desktop_client/ble"
	"desktop_client/clipboard"
	"desktop_client/config"
	"desktop_client/connectivity"
	"desktop_client/mqttclient"
	"desktop_client/notification"
	"desktop_client/playsound"
	"desktop_client/settings"
	"desktop_client/startup"
	"desktop_client/systrayhelpers"
	"desktop_client/wakewatcher"
	_ "embed"
	"fmt"
	"log"
	"mime"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/getlantern/systray"
	"github.com/sqweek/dialog"
)

var clientID string
var err error

//go:embed assets/notification.wav
var notificationSound []byte

var (
	loading   bool
	loadingMu sync.RWMutex

	messageAvailable    bool
	lastMessageSource   MessageFrom
	messageMu           sync.RWMutex
	notificationTimer   *time.Timer
	notificationTimerMu sync.Mutex

	errorActive bool
	errorTimer  *time.Timer
	errorMu     sync.Mutex

	mDownloadRecent  *systray.MenuItem
	mCopyToClipboard *systray.MenuItem

	networkUp bool = true
	networkMu sync.Mutex
	bleState  bool = false
)

var bleOps = make(chan func(), 1)

func updateIconState() {
	errorMu.Lock()
	hasError := errorActive
	errorMu.Unlock()

	// Error state takes precedence over all other states
	if hasError {
		animate.SetState(animate.StateError)
		return
	}

	loadingMu.RLock()
	isLoading := loading
	loadingMu.RUnlock()

	messageMu.RLock()
	hasMessage := messageAvailable
	messageMu.RUnlock()

	if isLoading {
		animate.SetState(animate.StateLoading)
	} else if hasMessage {
		animate.SetState(animate.StateNotification)
	} else {
		animate.SetState(animate.StateIdle)
	}
}

func showErrorState() {
	errorMu.Lock()
	// Stop any existing error timer
	if errorTimer != nil {
		errorTimer.Stop()
	}

	errorActive = true
	errorMu.Unlock()

	updateIconState()

	// Set timer to clear error state after 3 seconds
	errorMu.Lock()
	errorTimer = time.AfterFunc(3*time.Second, func() {
		errorMu.Lock()
		errorActive = false
		errorMu.Unlock()
		updateIconState() // Return to normal state
	})
	errorMu.Unlock()
}

func main() {
	// Delete original executable if requested flags are passed
	handleOriginalDeletion()

	if os.Getenv("DEV_MODE") == "1" {
		notification.Notification("DEVMODE = 1, loading dev certs and keys")
		config.LoadDevFiles()
	} else {
		err = startup.Initial()
		if err != nil {
			notification.Notification("Fatal: Could not run startup script")
			log.Fatalf("Could not run startup script: %v", err)
		}
	}

	go func() {
		for op := range bleOps {
			op()
		}
	}()

	// Start systray icon thread
	animate.Initialize()
	animate.Start()

	connectivity.Start()

	clientID, err = mqttclient.Connect()
	if err != nil {
		log.Printf("Full error with mqtt: %v", err)
	}

	mqttclient.SetOnMessageCallback(func() {
		HandleNewNotification(MQTT)
	})

	ble.SetOnMessageCallback(func() {
		HandleNewNotification(BLE)
	})

	wakewatcher.SetCallback(func() {
		mqttclient.Disconnect()
		clientID, err = mqttclient.Connect()
		if err != nil {
			notification.Notification("Error: Reconnect failed after wake")
			log.Printf("Reconnect failed after wake: %v", err)
		}
	})

	systray.Run(onReady, onExit)
}

func onReady() {
	systrayhelpers.SetTitle("")
	systrayhelpers.SetTooltip("Disconnected")
	mSendClipboard := systray.AddMenuItem("Send Clipboard", "Send clipboard contents")
	mSendFile := systray.AddMenuItem("Send File", "Send an file")
	systray.AddSeparator()
	mDownloadRecent = systray.AddMenuItem("Download", "Download the most recent file")
	mCopyToClipboard = systray.AddMenuItem("Copy to Clipboard", "Download the most recent file")
	systray.AddSeparator()
	mBLE := systray.AddMenuItemCheckbox("BLE", "Use BLE", !networkUp)
	systray.AddSeparator()
	mRestart := systray.AddMenuItem("Restart", "Restart the app")
	mQuit := systray.AddMenuItem("Quit", "Quit the app")

	mDownloadRecent.Disable()
	mCopyToClipboard.Disable()

	connectivity.OnChange(func(up bool) {
		select {
		case bleOps <- func() {
			networkMu.Lock()
			defer networkMu.Unlock()

			networkUp = up

			if settings.GetSettings().AutoBLE {
				if up {
					ble.Stop()
					bleState = false
					mBLE.Uncheck()
				} else {
					ble.Start(clientID, config.DeviceID)
					bleState = true
					mBLE.Check()
				}
			}
		}:
		default:
		}
	})

	go func() {
		for {
			<-mSendClipboard.ClickedCh
			go PublishClipboard()
		}
	}()
	go func() {
		for {
			<-mSendFile.ClickedCh
			go PublishFile()
		}
	}()
	go func() {
		for {
			<-mDownloadRecent.ClickedCh
			go DownloadRecent()
		}
	}()
	go func() {
		for {
			<-mCopyToClipboard.ClickedCh
			go CopyRecentToClipboard()
		}
	}()
	go func() {
		for {
			<-mRestart.ClickedCh
			var err error
			clientID, err = mqttclient.Connect()
			if err != nil {
				notification.Notification("Error: Could not restart MQTT client")
				log.Printf("Reconnect error: %v", err)
			}
		}
	}()
	go func() {
		for {
			<-mQuit.ClickedCh
			systray.Quit()
			return
		}
	}()
	go func() {
		for {
			<-mBLE.ClickedCh

			select {
			case bleOps <- func() {
				networkMu.Lock()
				defer networkMu.Unlock()

				if mBLE.Checked() {
					ble.Stop()
					bleState = false
					mBLE.Uncheck()
				} else {
					ble.Start(clientID, config.DeviceID)
					bleState = true
					mBLE.Check()
				}
			}:
			default:
				// If bleOps channel is full, skip this click
			}
		}
	}()

}

type MessageFrom int

const (
	MQTT MessageFrom = iota
	BLE
)

func HandleNewNotification(source MessageFrom) {
	messageMu.Lock()
	messageAvailable = true
	lastMessageSource = source
	messageMu.Unlock()

	mDownloadRecent.Enable()
	mCopyToClipboard.Enable()

	notificationTimerMu.Lock()
	if notificationTimer != nil {
		notificationTimer.Stop()
	}

	cacheTime := settings.GetSettings().CacheTime

	notificationTimer = time.AfterFunc(time.Duration(cacheTime)*time.Second, func() {
		messageMu.Lock()
		messageAvailable = false
		messageMu.Unlock()

		mDownloadRecent.Disable()
		mCopyToClipboard.Disable()

		mqttclient.ClearMsg()
		ble.ClearMsg()

		messageMu.Lock()
		messageAvailable = false
		messageMu.Unlock()
		updateIconState()
	})
	notificationTimerMu.Unlock()

	playsound.Play(notificationSound)

	if settings.GetSettings().AutoCopy {
		CopyRecentToClipboard()
	}
}

func onExit() {
	// Cleanup
}

var defaultExtensions = map[string]string{
	"text/plain":               ".txt",
	"application/octet-stream": ".bin",
}

func ExtensionFromMime(mimeType string) string {
	if ext, ok := defaultExtensions[mimeType]; ok {
		return ext
	}

	exts, err := mime.ExtensionsByType(mimeType)
	if err != nil || len(exts) == 0 {
		return ".bin"
	}
	return exts[0]
}

func PublishClipboard() {
	loadingMu.Lock()
	loading = true
	loadingMu.Unlock()
	updateIconState()

	content, mimeType, err := clipboard.Read()

	if len([]byte(content)) > 80*1024*1024 {
		notifyErr := notification.Notification("Clipboard too large (>80MB). Operation cancelled.")
		if notifyErr != nil {
			log.Println("Notification error:", notifyErr)
		}

		loadingMu.Lock()
		loading = false
		loadingMu.Unlock()
		showErrorState()
		return
	}

	if err != nil {
		notification.Notification("Could not read clipboard contents: " + err.Error())

		loadingMu.Lock()
		loading = false
		loadingMu.Unlock()
		updateIconState()
		return
	}

	ext := ExtensionFromMime(mimeType)

	topic := fmt.Sprintf("users/%s/notes", clientID)
	filename := fmt.Sprintf("clipboard%s", ext)

	if bleState {
		if len([]byte(content)) > 3*1024*1024 {
			notifyErr := notification.Notification("Clipboard too large (>3MB). Operation cancelled.")
			if notifyErr != nil {
				log.Println("Notification error:", notifyErr)
			}

			loadingMu.Lock()
			loading = false
			loadingMu.Unlock()
			showErrorState()
			return
		}

		ble.Publish([]byte(content), mimeType, filename)
	} else {
		err = mqttclient.Publish(topic, []byte(content), mimeType, filename)
		if err != nil {
			log.Println(err)
		}
	}

	loadingMu.Lock()
	loading = false
	loadingMu.Unlock()
	updateIconState()
}

func PublishFile() {
	loadingMu.Lock()
	loading = true
	loadingMu.Unlock()
	updateIconState()

	filePath, err := dialog.File().Title("Select a File").Load()
	if err != nil {
		log.Printf("File selection failed: %v", err)
		loadingMu.Lock()
		loading = false
		loadingMu.Unlock()
		updateIconState()
		return
	}

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		notifyErr := notification.Notification("Could not read file: " + err.Error())
		if notifyErr != nil {
			log.Printf("Failed to read file")
			log.Println("Notification error:", notifyErr)
		}

		loadingMu.Lock()
		loading = false
		loadingMu.Unlock()
		updateIconState()
		return
	}

	// file size > 80MB
	if len(fileBytes) > 80*1024*1024 {
		notifyErr := notification.Notification("File is too large (>80MB). Operation cancelled.")
		if notifyErr != nil {
			log.Println("Notification error:", notifyErr)
		}

		loadingMu.Lock()
		loading = false
		loadingMu.Unlock()
		showErrorState()
		return
	}

	fileName := filepath.Base(filePath)
	ext := filepath.Ext(fileName)

	// crude MIME type guess
	mimeType := mime.TypeByExtension(ext)

	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	topic := fmt.Sprintf("users/%s/notes", clientID)

	if bleState {
		if len(fileBytes) > 3*1024*1024 {
			notifyErr := notification.Notification("File too large for BLE (>3MB). Operation cancelled")
			if notifyErr != nil {
				log.Println("Notification error:", notifyErr)
			}

			loadingMu.Lock()
			loading = false
			loadingMu.Unlock()
			showErrorState()
			return
		}

		err := ble.Publish(fileBytes, mimeType, fileName)
		if err != nil {
			log.Println(err)
		}
	} else {
		err = mqttclient.Publish(topic, fileBytes, mimeType, fileName)
		if err != nil {
			loadingMu.Lock()
			loading = false
			loadingMu.Unlock()
			showErrorState()
			return
		}
	}

	loadingMu.Lock()
	loading = false
	loadingMu.Unlock()
	updateIconState()
}

func DownloadRecent() {

	var (
		fname string
		data  []byte
		ok    bool
	)

	messageMu.RLock()
	source := lastMessageSource
	messageMu.RUnlock()

	switch source {
	case MQTT:
		fname, _, data, ok = mqttclient.GetLastMessage()
	case BLE:
		fname, _, data, ok = ble.GetLastMessage()
	}

	if !ok {
		log.Println("No recent message to download")
		return
	}

	// guess ext by name
	ft := filepath.Ext(fname)
	// clipboard data has no fname (shouldnt be possible tbh)
	if ft == "" {
		ft = ".txt"
	}

	defaultName := fname
	if defaultName == "" {
		defaultName = "clipboard" + ft
	}

	// configure save-as dialog
	savePath, err := dialog.File().Title("Save Recent As").Filter(ft).SetStartFile(defaultName).Save()
	if err != nil {
		log.Printf("Save dialog canceled or failed: %v", err)
		return
	}

	// if user didn’t type an extension, tack on our ft (with leading dot)
	if filepath.Ext(savePath) == "" {
		savePath += ft
	}

	// Write out file
	if err := os.WriteFile(savePath, data, 0644); err != nil {
		log.Printf("Failed to write file: %v", err)
	} else {
		log.Printf("Saved recent note to %s", savePath)
	}
}

func CopyRecentToClipboard() {
	var (
		ctype string
		data  []byte
		ok    bool
	)

	messageMu.RLock()
	source := lastMessageSource
	messageMu.RUnlock()

	switch source {
	case MQTT:
		_, ctype, data, ok = mqttclient.GetLastMessage()
	case BLE:
		_, ctype, data, ok = ble.GetLastMessage()
	}

	if !ok {
		log.Println("No recent message to copy")
		return
	}

	if err := clipboard.Write(data, ctype); err != nil {
		log.Println("Couldn't copy to clipboard")
	}
}

// handleOriginalDeletion checks for --delete-original flag and deletes the specified file
func handleOriginalDeletion() {
	for i, arg := range os.Args {
		if arg == "--delete-original" && i+1 < len(os.Args) {
			originalPath := os.Args[i+1]

			// Small delay to ensure original process has fully exited
			time.Sleep(100 * time.Millisecond)

			// Attempt to delete the original executable
			err := os.Remove(originalPath)
			if err != nil {
				log.Printf("Could not delete original executable at %s: %v", originalPath, err)
			} else {
				log.Printf("Successfully deleted original executable at %s", originalPath)
			}
			break
		}
	}
}
