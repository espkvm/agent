// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

// Package service installs the agent so it starts by itself, the way each OS
// does that, and takes it out again.
package service

import (
	"io"
	"os"
	"path/filepath"
)

// Options for install.
type Options struct {
	// Desktop: run in the user's session, with the clipboard. Linux only; on
	// Windows and macOS the agent always runs in the user's session.
	Desktop bool
	// Power: the device may shut the machine down. A Linux system service
	// then needs the OS to let its user do that.
	Power bool
	// Args passed to "run", like --allow-power.
	Args []string
}

// Name is how the agent shows up in service lists.
const Name = "espkvm-agent"

// copyExe puts this executable at dst, unless it already runs from there.
func copyExe(dst string) error {
	src, err := os.Executable()
	if err != nil {
		return err
	}
	if s, err := filepath.EvalSymlinks(src); err == nil {
		src = s
	}
	if src == dst {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	// A new file then a rename, so a running copy is not written over.
	tmp := dst + ".new"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}
