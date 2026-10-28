// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

// Package sysinfo says what the machine is: its name, OS, uptime and addresses.
package sysinfo

import (
	"net"
	"os"
	"runtime"
	"time"
)

// Iface is one network interface that is up and has an address.
type Iface struct {
	Name string   `json:"name"`
	MAC  string   `json:"mac,omitempty"`
	IPs  []string `json:"ips"`
}

// Info is the answer to the info op.
type Info struct {
	Hostname string  `json:"hostname"`
	OS       string  `json:"os"`
	OSName   string  `json:"osName,omitempty"`
	Arch     string  `json:"arch"`
	UptimeS  int64   `json:"uptimeS,omitempty"`
	User     string  `json:"user,omitempty"`
	Ifaces   []Iface `json:"ifaces"`
}

// Get collects Info. Parts it cannot find are left empty.
func Get() Info {
	h, _ := os.Hostname()
	in := Info{Hostname: h, OS: runtime.GOOS, Arch: runtime.GOARCH, OSName: osName(), User: userName()}
	if up := uptime(); up > 0 {
		in.UptimeS = int64(up / time.Second)
	}
	in.Ifaces = ifaces()
	return in
}

func ifaces() []Iface {
	list, err := net.Interfaces()
	if err != nil {
		return nil
	}
	out := []Iface{}
	for _, i := range list {
		if i.Flags&net.FlagUp == 0 || i.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := i.Addrs()
		var ips []string
		for _, a := range addrs {
			if n, ok := a.(*net.IPNet); ok && !n.IP.IsLinkLocalUnicast() {
				ips = append(ips, n.IP.String())
			}
		}
		if len(ips) == 0 {
			continue
		}
		out = append(out, Iface{Name: i.Name, MAC: i.HardwareAddr.String(), IPs: ips})
	}
	return out
}

func userName() string {
	for _, k := range []string{"USER", "USERNAME"} {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}
