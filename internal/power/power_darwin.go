// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

package power

// System Events asks the apps to quit, like the Apple menu does, and works
// without root. A root agent could use shutdown(8), but the apps would not get
// their chance to save.
func command(action string) (string, []string) {
	verb := "shut down"
	if action == Reboot {
		verb = "restart"
	}
	return "osascript", []string{"-e", `tell application "System Events" to ` + verb}
}
