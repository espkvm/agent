// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

package proto

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
)

// pipePort joins the test, playing the device, to a Server.
type pipePort struct {
	io.Reader
	io.Writer
}

func start(t *testing.T, s *Server) (send func(string), next func() map[string]any) {
	t.Helper()
	devR, agentW := io.Pipe()
	agentR, devW := io.Pipe()
	go func() { _ = s.Serve(pipePort{agentR, agentW}) }()
	t.Cleanup(func() { devW.Close(); devR.Close() })
	lines := bufio.NewScanner(devR)
	next = func() map[string]any {
		for lines.Scan() {
			if strings.TrimSpace(lines.Text()) == "" {
				continue
			}
			var m map[string]any
			if err := json.Unmarshal(lines.Bytes(), &m); err != nil {
				t.Fatalf("agent sent %q: %v", lines.Text(), err)
			}
			return m
		}
		t.Fatal("agent closed")
		return nil
	}
	send = func(s string) {
		if _, err := io.WriteString(devW, s); err != nil {
			t.Fatal(err)
		}
	}
	return send, next
}

func TestHelloRequestsAndErrors(t *testing.T) {
	s := &Server{AgentVersion: "test", OS: "linux"}
	s.Handle("echo", func(p json.RawMessage) (any, error) {
		var v struct{ Text string }
		if json.Unmarshal(p, &v) != nil {
			return nil, ErrBadParams
		}
		return map[string]string{"text": v.Text}, nil
	})
	s.Handle("boom", func(json.RawMessage) (any, error) { panic("x") })
	s.Handle("off", func(json.RawMessage) (any, error) { return nil, ErrNotAllowed })
	s.Handle("plain", func(json.RawMessage) (any, error) { return nil, errors.New("nope") })
	send, next := start(t, s)

	h := next()
	if h["ev"] != "hello" || h["proto"].(float64) != Version || h["os"] != "linux" {
		t.Fatalf("hello: %v", h)
	}

	// Noise before the first request: a login prompt, a half line, junk JSON.
	send("login: \r\n{\"id\":\nnot json {\n{\"op\":\"\"}\n")
	send(`{"id":1,"op":"echo","params":{"text":"hi"}}` + "\n")
	if r := next(); r["id"].(float64) != 1 || r["ok"] != true || r["data"].(map[string]any)["text"] != "hi" {
		t.Fatalf("echo: %v", r)
	}
	cases := []struct {
		req, code string
	}{
		{`{"id":2,"op":"nothing"}`, "unknown_op"},
		{`{"id":3,"op":"boom"}`, "failed"},
		{`{"id":4,"op":"off"}`, "not_allowed"},
		{`{"id":5,"op":"plain"}`, "failed"},
	}
	for _, c := range cases {
		send(c.req + "\r\n")
		if r := next(); r["ok"] == true || r["code"] != c.code {
			t.Fatalf("%s: %v", c.req, r)
		}
	}
	send(`{"id":6,"op":"hello"}` + "\n")
	if r := next(); r["data"].(map[string]any)["ev"] != "hello" {
		t.Fatalf("hello op: %v", r)
	}
}

func TestLongLineDropped(t *testing.T) {
	s := &Server{}
	s.Handle("ping", func(json.RawMessage) (any, error) { return "pong", nil })
	send, next := start(t, s)
	next() // hello
	send("{" + strings.Repeat("a", MaxLine+10) + "\n")
	send(`{"id":7,"op":"ping"}` + "\n")
	if r := next(); r["id"].(float64) != 7 || r["data"] != "pong" {
		t.Fatalf("after a long line: %v", r)
	}
}
