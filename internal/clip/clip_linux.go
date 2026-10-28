// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

package clip

import (
	"os"
	"os/exec"
)

// Wayland first, then X11 with whichever tool is installed.
func tool(set bool) (string, []string, error) {
	have := func(n string) bool { _, err := exec.LookPath(n); return err == nil }
	if os.Getenv("WAYLAND_DISPLAY") != "" && have("wl-paste") {
		if set {
			return "wl-copy", nil, nil
		}
		return "wl-paste", []string{"--no-newline"}, nil
	}
	if os.Getenv("DISPLAY") != "" {
		switch {
		case have("xclip") && set:
			return "xclip", []string{"-selection", "clipboard", "-in"}, nil
		case have("xclip"):
			return "xclip", []string{"-selection", "clipboard", "-out"}, nil
		case have("xsel") && set:
			return "xsel", []string{"--clipboard", "--input"}, nil
		case have("xsel"):
			return "xsel", []string{"--clipboard", "--output"}, nil
		}
		return "", nil, errNoTool
	}
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return "", nil, errNoTool
	}
	return "", nil, ErrNoDesktop
}

var errNoTool = &toolError{}

type toolError struct{}

func (*toolError) Error() string {
	return "no clipboard tool: install wl-clipboard (Wayland) or xclip (X11)"
}

func get() (string, error) {
	name, args, err := tool(false)
	if err != nil {
		return "", err
	}
	out, err := run(nil, name, args...)
	return string(out), err
}

func set(text string) error {
	name, args, err := tool(true)
	if err != nil {
		return err
	}
	_, err = run([]byte(text), name, args...)
	return err
}

func hide(*exec.Cmd) {}
