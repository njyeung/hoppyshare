package animate

import (
	"desktop_client/settings"
	"desktop_client/systrayhelpers"
	_ "embed"
	"runtime"
	"sync"
	"time"
)

//go:embed macOS/default.png
var defaultIconMacOS []byte

//go:embed macOS/notification.png
var notificationIconMacOS []byte

//go:embed macOS/loading-1.png
var loading1IconMacOS []byte

//go:embed macOS/loading-2.png
var loading2IconMacOS []byte

//go:embed macOS/loading-3.png
var loading3IconMacOS []byte

//go:embed macOS/loading-4.png
var loading4IconMacOS []byte

//go:embed macOS/error.png
var errorMacOS []byte

//go:embed windows/default.ico
var defaultIconWindows []byte

//go:embed windows/notification.ico
var notificationIconWindows []byte

//go:embed windows/loading-1.ico
var loading1IconWindows []byte

//go:embed windows/loading-2.ico
var loading2IconWindows []byte

//go:embed windows/loading-3.ico
var loading3IconWindows []byte

//go:embed windows/loading-4.ico
var loading4IconWindows []byte

//go:embed windows/error.ico
var errorWindows []byte

// Animation states
type State int

const (
	StateIdle State = iota
	StateLoading
	StateNotification
	StateError
)

// Animation system
type Animator struct {
	currentState State
	stateMu      sync.RWMutex

	// Platform-specific icons
	defaultIcon      []byte
	notificationIcon []byte
	spinnerIcons     [][]byte
	errorIcon        []byte

	// Animation control
	ticker     *time.Ticker
	frameIndex int
	stopCh     chan struct{}
	running    bool
	runningMu  sync.Mutex
}

var globalAnimator *Animator

// Initialize sets up the animation system based on the current platform
func Initialize() {
	globalAnimator = &Animator{
		currentState: StateIdle,
		stopCh:       make(chan struct{}),
		frameIndex:   0,
	}

	// Set platform-specific icons
	if runtime.GOOS == "windows" {
		globalAnimator.defaultIcon = defaultIconWindows
		globalAnimator.notificationIcon = notificationIconWindows
		globalAnimator.spinnerIcons = [][]byte{
			loading1IconWindows,
			loading2IconWindows,
			loading3IconWindows,
			loading4IconWindows,
		}
		globalAnimator.errorIcon = errorWindows
	} else {
		globalAnimator.defaultIcon = defaultIconMacOS
		globalAnimator.notificationIcon = notificationIconMacOS
		globalAnimator.spinnerIcons = [][]byte{
			loading1IconMacOS,
			loading2IconMacOS,
			loading3IconMacOS,
			loading4IconMacOS,
		}
		globalAnimator.errorIcon = errorMacOS
	}
}

// Start begins the animation loop
func Start() {
	if globalAnimator == nil {
		panic("animate: must call Initialize() before Start()")
	}

	globalAnimator.runningMu.Lock()
	if globalAnimator.running {
		globalAnimator.runningMu.Unlock()
		return // Already running
	}
	globalAnimator.running = true
	globalAnimator.runningMu.Unlock()

	globalAnimator.ticker = time.NewTicker(200 * time.Millisecond)

	go func() {
		defer globalAnimator.ticker.Stop()

		for {
			select {
			case <-globalAnimator.stopCh:
				globalAnimator.runningMu.Lock()
				globalAnimator.running = false
				globalAnimator.runningMu.Unlock()
				return

			case <-globalAnimator.ticker.C:
				globalAnimator.updateFrame()
			}
		}
	}()
}

// Stop halts the animation loop
func Stop() {
	if globalAnimator == nil || !globalAnimator.running {
		return
	}

	close(globalAnimator.stopCh)
	globalAnimator.stopCh = make(chan struct{}) // Reset for potential restart
}

// SetState changes the current animation state
func SetState(state State) {
	if globalAnimator == nil {
		return
	}

	globalAnimator.stateMu.Lock()
	globalAnimator.currentState = state
	globalAnimator.stateMu.Unlock()
}

// GetState returns the current animation state
func GetState() State {
	if globalAnimator == nil {
		return StateIdle
	}

	globalAnimator.stateMu.RLock()
	defer globalAnimator.stateMu.RUnlock()
	return globalAnimator.currentState
}

// updateFrame handles the frame-by-frame animation logic
func (a *Animator) updateFrame() {
	a.stateMu.RLock()
	currentState := a.currentState
	a.stateMu.RUnlock()

	var iconToShow []byte

	switch currentState {
	case StateLoading:
		// Loading animation
		iconToShow = a.spinnerIcons[a.frameIndex%len(a.spinnerIcons)]
		a.frameIndex++

	case StateNotification:
		// Static notification icon (could add subtle pulse later)
		iconToShow = a.notificationIcon
	case StateError:
		// Error icon when stuff breaks
		iconToShow = a.errorIcon
	case StateIdle:
		if settings.GetSettings().LightAnimations {
			iconToShow = a.defaultIcon
		} else {
			// Slow breathing (slower than loading)
			breathingIndex := (a.frameIndex / 3) % len(a.spinnerIcons)
			iconToShow = a.spinnerIcons[breathingIndex]
			a.frameIndex++
		}

	}

	if iconToShow != nil {
		systrayhelpers.SetIcon(iconToShow)
	}
}
