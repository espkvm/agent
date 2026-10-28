// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

// Package port finds ESP-KVM's serial port and opens it.
package port

import (
	"errors"
	"io"
	"strings"
	"time"

	"go.bug.st/serial"
)

// The USB IDs ESP-KVM uses when its serial port is on.
const (
	VID = "303A"
	PID = "4013"
)

// ErrNotFound is returned when no ESP-KVM port is attached.
var ErrNotFound = errors.New("no ESP-KVM serial port found: in ESP-KVM's settings set \"Serial port over USB\" to the agent")

// Find returns the name of ESP-KVM's port, like /dev/ttyACM0 or COM5.
func Find() (string, error) {
	ports, err := list()
	if err != nil {
		return "", err
	}
	for _, p := range ports {
		if strings.EqualFold(p.vid, VID) && strings.EqualFold(p.pid, PID) {
			return p.name, nil
		}
	}
	return "", ErrNotFound
}

type found struct{ name, vid, pid string }

// Open opens the port raw. The baud rate means nothing on a USB port, but
// some drivers want one.
func Open(name string) (io.ReadWriteCloser, error) {
	p, err := serial.Open(name, &serial.Mode{BaudRate: 115200})
	if err != nil {
		return nil, err
	}
	// DTR tells the device somebody is listening.
	_ = p.SetDTR(true)
	return p, nil
}

// Wait polls Find until a port turns up or stop is closed.
func Wait(stop <-chan struct{}, every time.Duration) (string, bool) {
	for {
		if name, err := Find(); err == nil {
			return name, true
		}
		select {
		case <-stop:
			return "", false
		case <-time.After(every):
		}
	}
}
