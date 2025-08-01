package uninstall

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/emersion/go-autostart"
	"github.com/zalando/go-keyring"
)

const keyringService = "HoppyShare"

func RunUninstall() error {
	log.Println("Destroy flag is set. Uninstalling...")

	if err := unregisterStartup(); err != nil {
		log.Printf("Warning: Failed to unregister startup: %v", err)
	}

	if err := removeStartupBinary(); err != nil {
		log.Printf("Warning: Failed to remove installed binary: %v", err)
	}

	if err := clearKeychain(); err != nil {
		log.Printf("Warning: Failed to clear keychain: %v", err)
	}

	log.Println("Uninstall complete. Exiting.")
	os.Exit(0)
	return nil
}

func unregisterStartup() error {
	app := &autostart.App{
		Name: "HoppyShare",
	}

	return app.Disable()
}

func removeStartupBinary() error {
	var path string

	switch runtime.GOOS {
	case "windows":
		path = filepath.Join(os.Getenv("LOCALAPPDATA"), "HoppyShare", "HoppyShare.exe")
	case "darwin":
		path = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "HoppyShare", "HoppyShare")
	case "linux":
		path = filepath.Join(os.Getenv("HOME"), ".local", "bin", "hoppyshare")
	default:
		return errors.New("unsupported OS")
	}

	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}

	return err
}

func clearKeychain() error {
	var errs []error

	var keyringItems = []string{
		"CA", "Cert", "Key", "GroupKey", "DeviceID",
	}

	for _, item := range keyringItems {
		if err := keyring.Delete(keyringService, item); err != nil && err != keyring.ErrNotFound {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.New("one or more keychain items failed to delete")
	}

	return nil
}
