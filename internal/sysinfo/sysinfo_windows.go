// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

package sysinfo

import (
	"fmt"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func osName() string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return "Windows"
	}
	defer k.Close()
	name, _, _ := k.GetStringValue("ProductName")
	build, _, _ := k.GetStringValue("CurrentBuild")
	// Windows 11 still says "Windows 10" in ProductName; the build tells them apart.
	var n int
	if _, err := fmt.Sscan(build, &n); err == nil && n >= 22000 && len(name) > 10 && name[:10] == "Windows 10" {
		name = "Windows 11" + name[10:]
	}
	if build != "" {
		name += " (build " + build + ")"
	}
	return name
}

func uptime() time.Duration {
	return time.Duration(windows.DurationSinceBoot())
}
