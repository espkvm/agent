// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

//go:build !darwin || cgo

package port

import "go.bug.st/serial/enumerator"

func list() ([]found, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, err
	}
	var out []found
	for _, p := range ports {
		if p.IsUSB {
			out = append(out, found{p.Name, p.VID, p.PID})
		}
	}
	return out, nil
}
