package main

import (
	"desktop_client/clipboard"
	"desktop_client/config"
	"desktop_client/mqttclient"
	"desktop_client/playsound"
	"desktop_client/systrayhelpers"
	"desktop_client/wakewatcher"
	_ "embed"
	"fmt"
	"log"
	"mime"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/getlantern/systray"
	"github.com/sqweek/dialog"
)

var clientID string
var err error

//go:embed assets/macOS/default.png
var defaultIconMacOS []byte

//go:embed assets/macOS/notification.png
var notificationIconMacOS []byte

//go:embed assets/macOS/loading1.png
var loading1IconMacOS []byte

//go:embed assets/macOS/loading2.png
var loading2IconMacOS []byte

//go:embed assets/macOS/loading3.png
var loading3IconMacOS []byte

//go:embed assets/macOS/loading4.png
var loading4IconMacOS []byte

//go:embed assets/windows/default.ico
var defaultIconWindows []byte

//go:embed assets/windows/notification.ico
var notificationIconWindows []byte

//go:embed assets/windows/loading1.ico
var loading1IconWindows []byte

//go:embed assets/windows/loading2.ico
var loading2IconWindows []byte

//go:embed assets/windows/loading3.ico
var loading3IconWindows []byte

//go:embed assets/windows/loading4.ico
var loading4IconWindows []byte

//go:embed assets/notification.wav
var notificationSound []byte

const MESSAGE_CACHE_DURATION = 120

var (
	loading   bool
	loadingMu sync.RWMutex

	messageAvailable    bool
	messageMu           sync.RWMutex
	notificationTimer   *time.Timer
	notificationTimerMu sync.Mutex

	mDownloadRecent  *systray.MenuItem
	mCopyToClipboard *systray.MenuItem
)

var iconUpdateCh = make(chan struct{}, 1)

func requestIconUpdate() {
	select {
	case iconUpdateCh <- struct{}{}:
	default:

	}
}

func main() {
	if err := config.LoadEmbeddedConfig(); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	clientID, err = mqttclient.Connect()
	if err != nil {
		log.Printf("Full error with mqtt: %v", err)
	}

	mqttclient.SetOnMessageCallback(HandleNewNotification)

	wakewatcher.SetCallback(func() {
		mqttclient.Disconnect()
		clientID, err = mqttclient.Connect()
		if err != nil {
			log.Printf("Reconnect failed after wake: %v", err)
		}
	})

	systray.Run(onReady, onExit)
}

func onReady() {
	var (
		iconDefault  []byte
		iconNotify   []byte
		spinnerIcons [][]byte
	)
	if runtime.GOOS == "windows" {
		iconDefault = defaultIconWindows
		iconNotify = notificationIconWindows
		spinnerIcons = [][]byte{loading1IconWindows, loading2IconWindows, loading3IconWindows, loading4IconWindows}
	} else {
		iconDefault = defaultIconMacOS
		iconNotify = notificationIconMacOS
		spinnerIcons = [][]byte{loading1IconMacOS, loading2IconMacOS, loading3IconMacOS, loading4IconMacOS}
	}

	systrayhelpers.SetIcon(iconDefault)
	systrayhelpers.SetTitle("")
	systrayhelpers.SetTooltip("Disconnected")
	mSendClipboard := systray.AddMenuItem("Send Clipboard", "Send clipboard contents")
	mSendFile := systray.AddMenuItem("Send File", "Send an file")
	systray.AddSeparator()
	mDownloadRecent = systray.AddMenuItem("Download", "Download the most recent file")
	mCopyToClipboard = systray.AddMenuItem("Copy to Clipboard", "Download the most recent file")
	mDownloadRecent.Disable()
	mCopyToClipboard.Disable()
	systray.AddSeparator()
	mRestart := systray.AddMenuItem("Restart", "Restart the app")
	mQuit := systray.AddMenuItem("Quit", "Quit the app")

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

	// icon logic
	go func() {
		spinnerIdx := 0
		var ticker *time.Ticker

		for {
			select {
			case <-iconUpdateCh:
				loadingMu.RLock()
				isLoading := loading
				loadingMu.RUnlock()

				messageMu.RLock()
				hasMessage := messageAvailable
				messageMu.RUnlock()

				if isLoading {
					if ticker == nil {
						ticker = time.NewTicker(200 * time.Millisecond)
					}
				} else {
					if ticker != nil {
						ticker.Stop()
						ticker = nil
					}

					if hasMessage {
						systrayhelpers.SetIcon(iconNotify)
					} else {
						systrayhelpers.SetIcon(iconDefault)
					}
				}

			case <-func() <-chan time.Time {
				if ticker != nil {
					return ticker.C
				}
				return make(chan time.Time)
			}():
				systrayhelpers.SetIcon(spinnerIcons[spinnerIdx%len(spinnerIcons)])
				spinnerIdx++
			}
		}
	}()
}

func HandleNewNotification() {

	messageMu.Lock()
	messageAvailable = true
	messageMu.Unlock()

	mDownloadRecent.Enable()
	mCopyToClipboard.Enable()

	notificationTimerMu.Lock()
	if notificationTimer != nil {
		notificationTimer.Stop()
	}

	notificationTimer = time.AfterFunc(MESSAGE_CACHE_DURATION*time.Second, func() {
		messageMu.Lock()
		messageAvailable = false
		messageMu.Unlock()

		mDownloadRecent.Disable()
		mCopyToClipboard.Disable()

		mqttclient.ClearMsg()

		messageMu.Lock()
		messageAvailable = false
		messageMu.Unlock()
		requestIconUpdate()

	})
	notificationTimerMu.Unlock()

	playsound.Play(notificationSound)
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
	requestIconUpdate()

	content, mimeType, err := clipboard.Read()

	if err != nil {
		log.Printf("Clipboard read error: %v", err)
		loadingMu.Lock()
		loading = false
		loadingMu.Unlock()
		requestIconUpdate()
		return
	}

	ext := ExtensionFromMime(mimeType)

	topic := fmt.Sprintf("users/%s/notes", clientID)
	filename := fmt.Sprintf("clipboard%s", ext)
	log.Println(filename)
	err = mqttclient.Publish(topic, []byte(content), mimeType, filename)
	if err != nil {
		log.Println(err)
	}

	loadingMu.Lock()
	loading = false
	loadingMu.Unlock()
	requestIconUpdate()
}

func PublishFile() {
	loadingMu.Lock()
	loading = true
	loadingMu.Unlock()
	requestIconUpdate()

	filePath, err := dialog.File().Title("Select a File").Load()
	if err != nil {
		log.Printf("File selection failed: %v", err)
		loadingMu.Lock()
		loading = false
		loadingMu.Unlock()
		requestIconUpdate()
		return
	}

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Failed to read file: %v", err)
		loadingMu.Lock()
		loading = false
		loadingMu.Unlock()
		requestIconUpdate()
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

	err = mqttclient.Publish(topic, fileBytes, mimeType, fileName)
	if err != nil {
		log.Println(err)
	}

	loadingMu.Lock()
	loading = false
	loadingMu.Unlock()
	requestIconUpdate()
}

func DownloadRecent() {
	fname, _, data, ok := mqttclient.GetLastMessage()

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
	_, ctype, data, ok := mqttclient.GetLastMessage()

	if !ok {
		log.Println("No recent message to copy")
		return
	}

	err := clipboard.Write(data, ctype)

	if err != nil {
		log.Println("Couldn't copy to clipboard")
	}
}
