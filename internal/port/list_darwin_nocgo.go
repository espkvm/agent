// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

//go:build darwin && !cgo

package port

import (
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Without cgo there is no IOKit, so ask ioreg which callout device belongs to
// a USB device with our IDs. ioreg prints the tree with the IDs above the
// serial node that carries the device name.
var (
	reVendor  = regexp.MustCompile(`"idVendor" = (\d+)`)
	reProduct = regexp.MustCompile(`"idProduct" = (\d+)`)
	reCallout = regexp.MustCompile(`"IOCalloutDevice" = "([^"]+)"`)
)

func list() ([]found, error) {
	out, err := exec.Command("ioreg", "-r", "-c", "IOUSBHostDevice", "-l").Output()
	if err != nil {
		return fallback()
	}
	var res []found
	var vid, pid string
	for _, line := range strings.Split(string(out), "\n") {
		if m := reVendor.FindStringSubmatch(line); m != nil {
			vid = hex4(m[1])
		}
		if m := reProduct.FindStringSubmatch(line); m != nil {
			pid = hex4(m[1])
		}
		if m := reCallout.FindStringSubmatch(line); m != nil {
			res = append(res, found{m[1], vid, pid})
		}
	}
	return res, nil
}

// fallback: every usbmodem port, taken as ours. With several, --port picks one.
func fallback() ([]found, error) {
	names, _ := filepath.Glob("/dev/cu.usbmodem*")
	var res []found
	for _, n := range names {
		res = append(res, found{n, VID, PID})
	}
	return res, nil
}

func hex4(dec string) string {
	var n int
	for _, c := range dec {
		n = n*10 + int(c-'0')
	}
	const digits = "0123456789ABCDEF"
	return string([]byte{digits[n>>12&15], digits[n>>8&15], digits[n>>4&15], digits[n&15]})
}
