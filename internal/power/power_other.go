// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

//go:build !linux && !darwin && !windows

package power

func command(action string) (string, []string) {
	if action == Reboot {
		return "shutdown", []string{"-r", "now"}
	}
	return "shutdown", []string{"-p", "now"}
}
