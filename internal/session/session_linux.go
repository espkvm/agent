// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// logind: the user's graphical session and its LockedHint, which GNOME, KDE
// and most lock screens set.
func locked() (bool, error) {
	id, err := loginctl("show-user", strconv.Itoa(os.Getuid()), "-p", "Display", "--value")
	if err != nil {
		return false, err
	}
	if id == "" {
		return false, errors.New("no graphical session for this user")
	}
	v, err := loginctl("show-session", id, "-p", "LockedHint", "--value")
	if err != nil {
		return false, err
	}
	return v == "yes", nil
}

func loginctl(args ...string) (string, error) {
	out, err := exec.Command("loginctl", args...).Output()
	if err != nil {
		return "", errors.New("loginctl: " + err.Error())
	}
	return strings.TrimSpace(string(out)), nil
}
