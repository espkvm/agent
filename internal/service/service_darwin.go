// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"fmt"
	"html"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const label = "io.espkvm.agent"

// A LaunchAgent, not a daemon: it runs in the user's session, where the
// clipboard is, and opening the port needs no root on macOS.
const plist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key><string>%s</string>
	<key>ProgramArguments</key>
	<array>%s</array>
	<key>RunAtLoad</key><true/>
	<key>KeepAlive</key><true/>
</dict>
</plist>
`

func paths() (exe, pl string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	return filepath.Join(home, "Library", "Application Support", "ESP-KVM", "espkvm-agent"),
		filepath.Join(home, "Library", "LaunchAgents", label+".plist"), nil
}

func Install(o Options) error {
	exe, pl, err := paths()
	if err != nil {
		return err
	}
	if err := copyExe(exe); err != nil {
		return err
	}
	var b strings.Builder
	for _, a := range append([]string{exe, "run"}, o.Args...) {
		b.WriteString("<string>" + html.EscapeString(a) + "</string>")
	}
	if err := os.MkdirAll(filepath.Dir(pl), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(pl, []byte(fmt.Sprintf(plist, label, b.String())), 0o644); err != nil {
		return err
	}
	dom := fmt.Sprintf("gui/%d", os.Getuid())
	_ = exec.Command("launchctl", "bootout", dom+"/"+label).Run()
	if out, err := exec.Command("launchctl", "bootstrap", dom, pl).CombinedOutput(); err != nil {
		return fmt.Errorf("launchctl: %s", strings.TrimSpace(string(out)))
	}
	fmt.Println("Installed and started. It starts again at every login.")
	return nil
}

func Uninstall() error {
	exe, pl, err := paths()
	if err != nil {
		return err
	}
	_ = exec.Command("launchctl", "bootout", fmt.Sprintf("gui/%d/%s", os.Getuid(), label)).Run()
	for _, p := range []string{pl, exe} {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	fmt.Println("Removed.")
	return nil
}
