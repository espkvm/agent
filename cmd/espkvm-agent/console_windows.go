// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

package main

import "golang.org/x/sys/windows"

var freeConsole = windows.NewLazySystemDLL("kernel32.dll").NewProc("FreeConsole")

// Started at login from the Run key, a console program gets a window of its
// own. Letting go of the console closes it.
func detachConsole() { _, _, _ = freeConsole.Call() }
