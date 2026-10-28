// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

package clip

import (
	"os/exec"
	"strings"
	"syscall"
)

// PowerShell, so there is no cgo and no clipboard code of our own. UTF-8 both
// ways, set explicitly: the console's code page would mangle anything else.
const psGet = "[Console]::OutputEncoding=[Text.Encoding]::UTF8; Get-Clipboard -Raw"
const psSet = "[Console]::InputEncoding=[Text.Encoding]::UTF8; Set-Clipboard -Value ([Console]::In.ReadToEnd())"

func get() (string, error) {
	out, err := run(nil, "powershell", "-NoProfile", "-NonInteractive", "-Command", psGet)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(out), "\r\n"), nil
}

func set(text string) error {
	_, err := run([]byte(text), "powershell", "-NoProfile", "-NonInteractive", "-Command", psSet)
	return err
}

func hide(c *exec.Cmd) {
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000} // CREATE_NO_WINDOW
}
