package main

import (
	"bytes"
	"desktop_client/mqttclient"
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

	"github.com/atotto/clipboard"
	"github.com/getlantern/systray"
	"github.com/sqweek/dialog"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
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

func main() {
	clientID, err = mqttclient.Connect()
	if err != nil {
		log.Printf("Full error with mqtt: %v", err)
	}

	mqttclient.SetOnMessageCallback(HandleNewNotification)
	wakewatcher.SetCallback(func() {
		log.Println("[WAKE] Detected wake from sleep, restarting MQTT client...")
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

	systray.SetIcon(iconDefault)
	systray.SetTitle("")
	systray.SetTooltip("Disconnected")
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
			select {
			case <-mSendClipboard.ClickedCh:
				log.Println("Sending clipboard contents")
				go PublishClipboard()
			case <-mSendFile.ClickedCh:
				log.Println("Sending file...")
				go PublishFile()
			case <-mDownloadRecent.ClickedCh:
				log.Println("Downloading recent")
				go DownloadRecent()
			case <-mCopyToClipboard.ClickedCh:
				log.Println("Copying to clipboard")
				go CopyRecentToClipboard()
			case <-mRestart.ClickedCh:
				mqttclient.Disconnect()
				clientID, err = mqttclient.Connect()
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()

	// icon logic
	go func() {
		spinnerIdx := 0

		for {
			time.Sleep(200 * time.Millisecond)

			loadingMu.RLock()
			isLoading := loading
			loadingMu.RUnlock()

			messageMu.RLock()
			hasMessage := messageAvailable
			messageMu.RUnlock()

			if isLoading {
				systray.SetIcon(spinnerIcons[spinnerIdx%len(spinnerIcons)])
				spinnerIdx++
				continue
			} else if hasMessage {
				systray.SetIcon(iconNotify)
			} else {
				systray.SetIcon(iconDefault)
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

		go mqttclient.ClearMsg()

	})
	notificationTimerMu.Unlock()

	PlaySound()
}

func PlaySound() {
	streamer, format, err := wav.Decode(bytes.NewReader(notificationSound))
	if err != nil {
		log.Fatal("Failed to decode notification sound:", err)
	}
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		streamer.Close()
	})))
}

func onExit() {
	// Cleanup here if needed
}

func PublishClipboard() {
	loadingMu.Lock()
	loading = true
	loadingMu.Unlock()

	content, err := clipboard.ReadAll()

	if err != nil {
		log.Printf("Clipboard read error: %v", err)
		loadingMu.Lock()
		loading = false
		loadingMu.Unlock()
		return
	}

	topic := fmt.Sprintf("users/%s/notes", clientID)
	err = mqttclient.Publish(topic, []byte(content), "text/plain", "")
	if err != nil {
		log.Println(err)
	}
	loadingMu.Lock()
	loading = false
	loadingMu.Unlock()
}

func PublishFile() {
	loadingMu.Lock()
	loading = true
	loadingMu.Unlock()

	filePath, err := dialog.File().Title("Select a File").Load()
	if err != nil {
		log.Printf("File selection failed: %v", err)
		loadingMu.Lock()
		loading = false
		loadingMu.Unlock()
		return
	}

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Failed to read file: %v", err)
		loadingMu.Lock()
		loading = false
		loadingMu.Unlock()
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
}

func DownloadRecent() {
	fname, _, data, ok := mqttclient.GetLastMessage()

	if !ok {
		log.Println("No recent message to download")
		return
	}

	// guess file type (with dot)
	ft := filepath.Ext(fname)
	// clipboard data has no fname
	if ft == "" {
		ft = ".txt"
	}

	// build sensible default name
	defaultName := fname
	if defaultName == "" {
		defaultName = "copy_text" + ft
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

	// 4) write out
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

	if ctype == "text/plain" {
		text := string(data)
		if err := clipboard.WriteAll(text); err != nil {
			log.Printf("Clipboard write failed: %v", err)
		} else {
			log.Println("Copied recent text note to clipboard")
		}
	} else {
		log.Printf("Cannot copy non-text (%s) to clipboard", ctype)
	}
}
