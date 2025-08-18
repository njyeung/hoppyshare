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
	"syscall"

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
	var scriptText string
	var pid = os.Getpid()
	var debugLogPath = filepath.Join(os.TempDir(), "hoppyshare_uninstall.log")
	switch runtime.GOOS {
	case "windows":
		binaryPath = filepath.Join(os.Getenv("LOCALAPPDATA"), "HoppyShare", "HoppyShare.exe")
		scriptPath = filepath.Join(os.Getenv("TEMP"), "uninstall_hoppyshare.bat")
		scriptText = fmt.Sprintf(`@echo off
		setlocal enabledelayedexpansion
		set PID="%d"
		set BIN="%s"
		set DIR="%s"
		set LOG="%s"

		echo [%%DATE%% %%TIME%%] Starting uninstall PID=%%PID%% BIN=%%BIN%% > %%LOG%%

		:waitloop
		for /f "tokens=1" %%a in ('powershell -NoProfile -Command "try { Get-Process -Id %%%%%%%%PID%%%%%%%% | Out-Null; Write-Output RUNNING } catch { Write-Output GONE }"') do set STATE=%%a
		if /I "%%STATE%%"=="RUNNING" (
		ping -n 2 127.0.0.1 >nul
		goto waitloop
		)

		echo [%%DATE%% %%TIME%%] Parent gone; deleting file... >> %%LOG%%
		del /f /q %%BIN%% >> %%LOG%% 2>&1

		echo [%%DATE%% %%TIME%%] Removing dir (best-effort)... >> %%LOG%%
		rmdir /s /q %%DIR%% >> %%LOG%% 2>&1

		echo [%%DATE%% %%TIME%%] Done. Self-delete. >> %%LOG%%
		del "%%~f0"
		`, pid, binaryPath, filepath.Dir(binaryPath), debugLogPath)
	case "darwin":
		binaryPath = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "HoppyShare", "HoppyShare")
		fallthrough
	case "linux":
		if runtime.GOOS == "linux" {
			binaryPath = filepath.Join(os.Getenv("HOME"), ".local", "bin", "hoppyshare")
		}
		scriptPath = filepath.Join(os.TempDir(), "uninstall_hoppyshare.sh")

		// Use kill -0 to probe PID existence, then remove; log for debugging
		scriptText = fmt.Sprintf(`#!/usr/bin/env bash
		PID=%d
		BIN="%s"
		DIR="%s"
		LOG="%s"

		printf '[%%s] Starting uninstall helper. PID=%d BIN=%q\n' "$(date)" "$PID" "$BIN" > "$LOG"

		# Wait for parent PID to exit
		while kill -0 "$PID" 2>/dev/null; do
		sleep 0.2
		done

		printf '[%%s] Parent gone; deleting file...\n' "$(date)" >> "$LOG"
		rm -f "$BIN" >> "$LOG" 2>&1

		printf '[%%s] Removing dir (best-effort)...\n' "$(date)" >> "$LOG"
		rmdir "$DIR" >> "$LOG" 2>&1 || true

		printf '[%%s] Done. Self-delete.\n' "$(date)" >> "$LOG"
		rm -f "$0"
		`, pid, binaryPath, filepath.Dir(binaryPath), debugLogPath)

	default:
		return errors.New("unsupported OS")
	}

	// Check if we can find binary at expected path
	if _, err := os.Stat(binaryPath); err != nil {
		if os.IsNotExist(err) {
			notification.Notification("Could not find binary in expected location")
			return nil
		}
		notification.Notification("Stat failed for binary path")
		return nil
	}

	// Create the cleanup script
	if err := os.WriteFile(scriptPath, []byte(scriptText), 0755); err != nil {
		return fmt.Errorf("failed to create cleanup script: %v", err)
	}

	// Execute the cleanup script in the background
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command(
			"powershell",
			"-NoProfile",
			"-ExecutionPolicy", "Bypass",
			"-WindowStyle", "Hidden",
			"-File", scriptPath,
		)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		_ = cmd.Start()
	default:
		cmd = exec.Command("sh", "-c", scriptPath+" &")
	}

	if err := cmd.Start(); err != nil {
		_ = os.Remove(scriptPath)
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
