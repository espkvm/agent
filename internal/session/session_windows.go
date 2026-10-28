// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"syscall"
	"unsafe"
)

var (
	user32      = syscall.NewLazyDLL("user32.dll")
	openInput   = user32.NewProc("OpenInputDesktop")
	closeDesk   = user32.NewProc("CloseDesktop")
	getObjectIn = user32.NewProc("GetUserObjectInformationW")
)

// The desktop taking input is "Default" while the user works. The lock screen
// (and UAC) is a different desktop this process cannot even open.
func locked() (bool, error) {
	const readObjects, uoiName = 0x0001, 2
	h, _, _ := openInput.Call(0, 0, readObjects)
	if h == 0 {
		return true, nil
	}
	defer closeDesk.Call(h)
	var name [64]uint16
	var n uint32
	r, _, err := getObjectIn.Call(h, uoiName, uintptr(unsafe.Pointer(&name[0])), unsafe.Sizeof(name), uintptr(unsafe.Pointer(&n)))
	if r == 0 {
		return false, err
	}
	return syscall.UTF16ToString(name[:]) != "Default", nil
}
