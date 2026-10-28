// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

package clip

import "os/exec"

func get() (string, error) {
	out, err := run(nil, "pbpaste")
	return string(out), err
}

func set(text string) error {
	_, err := run([]byte(text), "pbcopy")
	return err
}

func hide(*exec.Cmd) {}
