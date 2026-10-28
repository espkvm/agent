// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"bytes"
	"errors"
	"os/exec"
)

// The console user's entry in the I/O registry carries CGSSessionScreenIsLocked
// while the screen is locked.
func locked() (bool, error) {
	out, err := exec.Command("ioreg", "-n", "Root", "-d1", "-a").Output()
	if err != nil {
		return false, errors.New("ioreg: " + err.Error())
	}
	_, after, found := bytes.Cut(out, []byte("<key>CGSSessionScreenIsLocked</key>"))
	return found && bytes.HasPrefix(bytes.TrimSpace(after), []byte("<true/>")), nil
}
