// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

// Package clip reads and sets the text clipboard of the desktop session the
// agent runs in. A system service has no desktop, so there it fails cleanly.
package clip

import (
	"bytes"
	"errors"
	"os/exec"
)

// MaxText is the most text taken either way.
const MaxText = 512 * 1024

// ErrNoDesktop is returned when there is no clipboard to reach.
var ErrNoDesktop = errors.New("no desktop session: run the agent in the user's session for the clipboard")

// ErrTooLong is returned for text over MaxText.
var ErrTooLong = errors.New("text is over 512 KB")

// Get returns the clipboard's text.
func Get() (string, error) {
	t, err := get()
	if err == nil && len(t) > MaxText {
		return "", ErrTooLong
	}
	return t, err
}

// Set puts text on the clipboard.
func Set(text string) error {
	if len(text) > MaxText {
		return ErrTooLong
	}
	return set(text)
}

func run(stdin []byte, name string, args ...string) ([]byte, error) {
	c := exec.Command(name, args...)
	hide(c)
	if stdin != nil {
		c.Stdin = bytes.NewReader(stdin)
	}
	var out, errb bytes.Buffer
	c.Stdout, c.Stderr = &out, &errb
	if err := c.Run(); err != nil {
		if msg := bytes.TrimSpace(errb.Bytes()); len(msg) > 0 {
			return nil, errors.New(name + ": " + string(msg))
		}
		return nil, err
	}
	return out.Bytes(), nil
}
