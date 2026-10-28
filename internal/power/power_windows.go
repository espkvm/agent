// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

package power

// A short delay rather than 0, so the agent's answer goes out first.
func command(action string) (string, []string) {
	flag := "/s"
	if action == Reboot {
		flag = "/r"
	}
	return "shutdown", []string{flag, "/t", "3", "/c", "ESP-KVM agent"}
}
