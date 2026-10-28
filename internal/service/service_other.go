// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

//go:build !linux && !darwin && !windows

package service

import "errors"

func Install(Options) error {
	return errors.New("install is not supported here: start \"espkvm-agent run\" yourself")
}
func Uninstall() error { return errors.New("not supported here") }
