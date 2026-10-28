// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

// espkvm-agent runs on the machine ESP-KVM is plugged into and answers the
// device over the USB serial port: what the machine is, its clipboard, and,
// when allowed here, reading the clipboard and a clean shutdown.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/espkvm/agent/internal/clip"
	"github.com/espkvm/agent/internal/port"
	"github.com/espkvm/agent/internal/power"
	"github.com/espkvm/agent/internal/proto"
	"github.com/espkvm/agent/internal/service"
	"github.com/espkvm/agent/internal/session"
	"github.com/espkvm/agent/internal/sysinfo"
)

// version is set at build time: -ldflags "-X main.version=1.2.3".
var version = "dev"

const usage = `espkvm-agent %s - the ESP-KVM agent for the machine it is plugged into

  espkvm-agent run [--port NAME] [ALLOW...]      answer the device (foreground)
  espkvm-agent install [--desktop] [ALLOW...]    start it at every boot or login
  espkvm-agent uninstall                         take it out again
  espkvm-agent find                              which port is ESP-KVM's
  espkvm-agent info                              what it would tell the device
  espkvm-agent version

By default the device may read what the machine is and put text on the
clipboard. ALLOW adds more:
  --allow-clipboard-read  let the device read the clipboard, while the screen
                          is not locked
  --allow-power           let the device shut this machine down or restart it
--desktop (Linux) runs it in the user's desktop session, with the clipboard,
instead of as a system service.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Printf(usage, version)
		os.Exit(2)
	}
	cmd, args := os.Args[1], os.Args[2:]
	var err error
	switch cmd {
	case "run":
		err = runCmd(args)
	case "install":
		err = installCmd(args)
	case "uninstall":
		err = service.Uninstall()
	case "find":
		var name string
		if name, err = port.Find(); err == nil {
			fmt.Println(name)
		}
	case "info":
		err = printJSON(sysinfo.Get())
	case "version", "--version", "-v":
		fmt.Println(version)
	case "help", "--help", "-h":
		fmt.Printf(usage, version)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", cmd)
		fmt.Printf(usage, version)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "espkvm-agent:", err)
		os.Exit(1)
	}
}

func printJSON(v any) error {
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	return e.Encode(v)
}

// What the machine allows the device beyond the defaults.
type allow struct {
	clipRead, power bool
}

func allowFlags(fs *flag.FlagSet) *allow {
	a := &allow{}
	fs.BoolVar(&a.clipRead, "allow-clipboard-read", false, "let the device read the clipboard while the screen is not locked")
	fs.BoolVar(&a.power, "allow-power", false, "let the device shut down or restart this machine")
	return a
}

func installCmd(args []string) error {
	fs := flag.NewFlagSet("install", flag.ExitOnError)
	desktop := fs.Bool("desktop", false, "Linux: run in the desktop session, with the clipboard")
	a := allowFlags(fs)
	_ = fs.Parse(args)
	var run []string
	if a.clipRead {
		run = append(run, "--allow-clipboard-read")
	}
	if a.power {
		run = append(run, "--allow-power")
	}
	return service.Install(service.Options{Desktop: *desktop, Power: a.power, Args: run})
}

func runCmd(args []string) error {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	name := fs.String("port", "", "the serial port; found by its USB IDs when empty")
	a := allowFlags(fs)
	background := fs.Bool("background", false, "Windows: drop the console window")
	_ = fs.Parse(args)
	if *background {
		detachConsole()
	}
	log.SetFlags(log.LstdFlags)

	srv := newServer(*a)
	stop := make(chan struct{})
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() { <-sig; close(stop); os.Exit(0) }()

	log.Printf("espkvm-agent %s on %s, clipboard read %s, power %s", version, runtime.GOOS, onOff(a.clipRead), onOff(a.power))
	announced := false
	for {
		p := *name
		if p == "" {
			if !announced {
				log.Print("waiting for ESP-KVM's serial port")
				announced = true
			}
			var ok bool
			if p, ok = port.Wait(stop, 2*time.Second); !ok {
				return nil
			}
		}
		rw, err := port.Open(p)
		if err != nil {
			log.Printf("%s: %v", p, err)
			time.Sleep(3 * time.Second)
			continue
		}
		log.Printf("connected on %s", p)
		err = srv.Serve(rw)
		rw.Close()
		if err != nil && err != io.EOF {
			log.Printf("%s: %v", p, err)
		}
		log.Printf("disconnected from %s", p)
		announced = false
		time.Sleep(time.Second)
	}
}

func onOff(b bool) string {
	if b {
		return "allowed"
	}
	return "not allowed"
}

func newServer(a allow) *proto.Server {
	s := &proto.Server{AgentVersion: version, OS: runtime.GOOS}
	s.Handle("ping", func(json.RawMessage) (any, error) { return "pong", nil })
	s.Handle("info", func(json.RawMessage) (any, error) { return sysinfo.Get(), nil })
	s.Handle("clipboard.get", func(json.RawMessage) (any, error) {
		if !a.clipRead {
			return nil, proto.ErrNotAllowed
		}
		t, err := clip.Get()
		if err != nil {
			return nil, clipErr(err)
		}
		// Not while the person is away: a password copied a minute ago may sit
		// there. Checked after the read, so "no desktop" still comes out as such.
		if l, err := session.Locked(); err != nil {
			return nil, fmt.Errorf("cannot tell whether the screen is locked: %w", err)
		} else if l {
			return nil, proto.ErrLocked
		}
		return map[string]string{"text": t}, nil
	})
	s.Handle("clipboard.set", func(p json.RawMessage) (any, error) {
		var v struct {
			Text *string `json:"text"`
		}
		if json.Unmarshal(p, &v) != nil || v.Text == nil {
			return nil, proto.ErrBadParams
		}
		return nil, clipErr(clip.Set(*v.Text))
	})
	s.Handle("power", func(p json.RawMessage) (any, error) {
		if !a.power {
			return nil, proto.ErrNotAllowed
		}
		var v struct {
			Action string `json:"action"`
		}
		if json.Unmarshal(p, &v) != nil || (v.Action != power.Shutdown && v.Action != power.Reboot) {
			return nil, proto.ErrBadParams
		}
		// After the answer has gone out: the machine may be gone a moment later.
		go func() {
			time.Sleep(500 * time.Millisecond)
			log.Printf("power: %s, asked by the device", v.Action)
			if err := power.Do(v.Action); err != nil {
				log.Printf("power: %v", err)
			}
		}()
		return map[string]string{"action": v.Action}, nil
	})
	return s
}

// The device tells "no desktop here" apart from a clipboard that failed.
func clipErr(err error) error {
	if errors.Is(err, clip.ErrNoDesktop) {
		return &proto.Error{Code: "no_desktop", Msg: err.Error()}
	}
	return err
}
