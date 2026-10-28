// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

// Package power shuts the machine down or restarts it the way the OS wants it
// done, so programs get to close and disks are flushed.
package power

import (
	"errors"
	"os/exec"
)

// Actions the device may ask for.
const (
	Shutdown = "shutdown"
	Reboot   = "reboot"
)

// ErrAction is returned for an action this package does not know.
var ErrAction = errors.New("action must be shutdown or reboot")

// Do starts the action. It returns once the OS has taken the request, before
// the machine goes down, so the answer still reaches the device.
func Do(action string) error {
	if action != Shutdown && action != Reboot {
		return ErrAction
	}
	name, args := command(action)
	c := exec.Command(name, args...)
	out, err := c.CombinedOutput()
	if err != nil && len(out) > 0 {
		return errors.New(name + ": " + string(out))
	}
	return err
}
