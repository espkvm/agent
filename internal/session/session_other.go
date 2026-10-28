// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

//go:build !linux && !darwin && !windows

package session

import "errors"

func locked() (bool, error) { return false, errors.New("not supported on this OS") }
