// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

// Package session tells whether the screen of the session the agent runs in
// is locked. The clipboard is not read then: the person is away.
package session

// Locked reports whether the screen is locked. An error means it could not be
// told, and the caller should act as if it were.
func Locked() (bool, error) { return locked() }
