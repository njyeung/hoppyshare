package startup

import (
	"desktop_client/config"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/emersion/go-autostart"
	"github.com/zalando/go-keyring"
)

const keyringService = "HoppyShare"

func Initial() error {

	// AFTERWARDS
	err := config.LoadKeysFromKeychain()
	// If we found our keys, then it's not the first time launching
	if err == nil {
		return nil
	}

	// FIRST TIME LAUNCH
	if err := config.LoadEmbeddedConfig(); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	err = storeKeysInKeychain()
	if err != nil {
		return err
	}

	// Get the original executable path
	oldExe, err := os.Executable()
	if err != nil {
		return err
	}
	oldExe, err = filepath.Abs(oldExe)
	if err != nil {
		return err
	}

	// Move copied executable to startup directory and register it for startup
	path, err := moveToStartupDir()
	if err != nil {
		return err
	}
	err = registerStartup(path)
	if err != nil {
		return err
	}

	// Pass the original executable path to the new process so it can delete it
	args := os.Args[1:]
	if oldExe != path {
		args = append([]string{"--delete-original", oldExe}, args...)
	}
	cmd := exec.Command(path, args...)
	err = cmd.Start()
	if err != nil {
		return err
	}

	os.Exit(0)
	return nil
}

func storeKeysInKeychain() error {
	items := map[string][]byte{
		"CA":       config.CAPem,
		"Cert":     config.CertPem,
		"Key":      config.KeyPem,
		"GroupKey": config.GroupKey,
		"DeviceID": []byte(config.DeviceID),
	}

	for name, data := range items {
		enc := base64.StdEncoding.EncodeToString(data)

		err := keyring.Set(keyringService, name, enc)
		if err != nil {
			return errors.New("failed to store in keychain")
		}

	}

	return nil
}

func moveToStartupDir() (string, error) {
	oldExe, err := os.Executable()
	if err != nil {
		return "", err
	}

	oldExe, err = filepath.Abs(oldExe)
	if err != nil {
		return "", err
	}

	var (
		appDir  string
		newExec string
	)

	switch runtime.GOOS {
	// C:%LOCALAPPDATA%\HoppyShare\HoppyShare.exe
	case "windows":
		appDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "HoppyShare")
		newExec = filepath.Join(appDir, "HoppyShare.exe")
	// /Library/Application Support/HoppyShare/HoppyShare
	case "darwin":
		appDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "HoppyShare")
		newExec = filepath.Join(appDir, "HoppyShare")
	// /.local/bin
	case "linux":
		appDir = filepath.Join(os.Getenv("HOME"), ".local", "bin")
		newExec = filepath.Join(appDir, "hoppyshare")
	default:
		return "", errors.New("could not detect a supported OS: " + runtime.GOOS)
	}

	err = os.MkdirAll(appDir, 0o755)
	if err != nil {
		return "", err
	}

	// smart ahhh user moved it into the right spot
	if oldExe == newExec {
		return newExec, nil
	}

	in, err := os.Open(oldExe)
	if err != nil {
		return "", err
	}

	defer in.Close()

	out, err := os.Create(newExec)
	if err != nil {
		return "", err
	}

	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return "", err
	}

	err = out.Chmod(0o755)
	if err != nil {
		return "", err
	}

	return newExec, nil
}

func registerStartup(path string) error {
	app := &autostart.App{
		Name:        "HoppyShare",
		DisplayName: "HoppyShare",
		Exec:        []string{path},
	}

	return app.Enable()
}

// EnableStartup enables startup for the current executable location
func EnableStartup() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, err = filepath.Abs(exe)
	if err != nil {
		return err
	}
	
	return registerStartup(exe)
}

// DisableStartup disables startup for HoppyShare
func DisableStartup() error {
	app := &autostart.App{
		Name:        "HoppyShare",
		DisplayName: "HoppyShare",
	}

	return app.Disable()
}
