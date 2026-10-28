// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

package sysinfo

import (
	"os/exec"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

func osName() string {
	v, err := exec.Command("sw_vers", "-productVersion").Output()
	if err != nil {
		return "macOS"
	}
	return "macOS " + strings.TrimSpace(string(v))
}

func uptime() time.Duration {
	tv, err := unix.SysctlTimeval("kern.boottime")
	if err != nil {
		return 0
	}
	return time.Since(time.Unix(tv.Unix()))
}
