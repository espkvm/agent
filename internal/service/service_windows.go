// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// The Run key of the current user: it starts at login in the user's session,
// where the clipboard is, and needs no administrator. A COM port is open to
// any user.
const runKey = `Software\Microsoft\Windows\CurrentVersion\Run`
const valueName = "ESP-KVM agent"

func exePath() (string, error) {
	dir, err := os.UserCacheDir() // %LOCALAPPDATA%
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "ESP-KVM", "espkvm-agent.exe"), nil
}

func Install(o Options) error {
	exe, err := exePath()
	if err != nil {
		return err
	}
	if err := copyExe(exe); err != nil {
		return err
	}
	k, _, err := registry.CreateKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	args := append([]string{"run", "--background"}, o.Args...)
	if err := k.SetStringValue(valueName, `"`+exe+`" `+strings.Join(args, " ")); err != nil {
		return err
	}
	c := exec.Command(exe, args...)
	if err := c.Start(); err != nil {
		return err
	}
	fmt.Println("Installed and started. It starts again at every login.")
	return nil
}

func Uninstall() error {
	exe, err := exePath()
	if err != nil {
		return err
	}
	if k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.SET_VALUE); err == nil {
		_ = k.DeleteValue(valueName)
		k.Close()
	}
	// Every running copy but this one.
	_ = exec.Command("taskkill", "/F", "/FI", "IMAGENAME eq espkvm-agent.exe",
		"/FI", fmt.Sprintf("PID ne %d", os.Getpid())).Run()
	if err := os.Remove(exe); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("%v (if it is still running, log out and run uninstall again)", err)
	}
	fmt.Println("Removed.")
	return nil
}
