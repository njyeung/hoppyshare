package uninstall

import (
	"desktop_client/notification"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/emersion/go-autostart"
	"github.com/zalando/go-keyring"
)

const keyringService = "HoppyShare"

func RunUninstall() error {
	notification.Notification("Destroy flag is set. Uninstalling...")

	if err := unregisterStartup(); err != nil {
		log.Printf("Warning: Failed to unregister startup: %v", err)
		notification.Notification("Warning: Failed to unregister startup")
	}

	if err := removeStartupBinary(); err != nil {
		log.Printf("Warning: Failed to remove installed binary: %v", err)
		notification.Notification("Warning: Failed to remove installed binary")
	}

	if err := clearKeychain(); err != nil {
		log.Printf("Warning: Failed to clear keychain: %v", err)
		notification.Notification("Warning: Failed to clear keychain")
	}

	log.Println("Uninstall complete. Exiting.")
	notification.Notification("Uninstall complete. Exiting")

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
	var binaryPath string
	var scriptPath string
	var scriptContent string

	switch runtime.GOOS {
	case "windows":
		binaryPath = filepath.Join(os.Getenv("LOCALAPPDATA"), "HoppyShare", "HoppyShare.exe")
		scriptPath = filepath.Join(os.Getenv("TEMP"), "uninstall_hoppyshare.bat")
		scriptContent = fmt.Sprintf(`@echo off
			timeout /t 2 /nobreak > nul
			del /f /q "%s" 2>nul
			rmdir "%s" 2>nul
			del /f /q "%%~f0" 2>nul`, binaryPath, filepath.Dir(binaryPath))
	case "darwin", "linux":
		if runtime.GOOS == "darwin" {
			binaryPath = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "HoppyShare", "HoppyShare")
		} else {
			binaryPath = filepath.Join(os.Getenv("HOME"), ".local", "bin", "hoppyshare")
		}
		scriptPath = filepath.Join(os.TempDir(), "uninstall_hoppyshare.sh")
		scriptContent = fmt.Sprintf(`#!/bin/bash
			sleep 2
			rm -f "%s"
			rmdir "%s" 2>/dev/null
			rm -f "$0"`, binaryPath, filepath.Dir(binaryPath))
	default:
		return errors.New("unsupported OS")
	}

	// Check if binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return nil
	}

	// Create the cleanup script
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		return fmt.Errorf("failed to create cleanup script: %v", err)
	}

	// Execute the cleanup script in the background
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", "start", "/B", scriptPath)
	} else {
		cmd = exec.Command("sh", "-c", scriptPath+" &")
	}

	if err := cmd.Start(); err != nil {
		os.Remove(scriptPath) // Clean up script if we can't run it
		return fmt.Errorf("failed to start cleanup script: %v", err)
	}

	return nil
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
