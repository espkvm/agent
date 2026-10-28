// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

//go:build !linux && !darwin && !windows

package sysinfo

import "time"

func osName() string        { return "" }
func uptime() time.Duration { return 0 }
