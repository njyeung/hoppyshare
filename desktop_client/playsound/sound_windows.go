//go:build windows
// +build windows

package playsound

import (
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	winmm          = syscall.NewLazyDLL("winmm.dll")
	procPlaySoundW = winmm.NewProc("PlaySoundW")
)

const (
	SND_FILENAME = 0x00020000
	SND_ASYNC    = 0x00000001
)

func play(sound []byte) {
	tmpPath := filepath.Join(os.TempDir(), "notification.wav")
	os.WriteFile(tmpPath, sound, 0644)

	pathPtr, _ := syscall.UTF16PtrFromString(tmpPath)

	procPlaySoundW.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		0,
		SND_FILENAME|SND_ASYNC,
	)
}
