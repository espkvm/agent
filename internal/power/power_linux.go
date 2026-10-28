// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

package power

// systemd: polkit lets the user of the active session do it, root always can.
func command(action string) (string, []string) {
	if action == Reboot {
		return "systemctl", []string{"reboot"}
	}
	return "systemctl", []string{"poweroff"}
}
