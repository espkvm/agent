// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

//go:build !linux && !darwin && !windows

package clip

import "os/exec"

func get() (string, error) { return "", ErrNoDesktop }
func set(string) error     { return ErrNoDesktop }
func hide(*exec.Cmd)       {}
