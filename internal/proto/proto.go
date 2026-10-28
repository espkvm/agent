// SPDX-FileCopyrightText: 2026 ESP-KVM contributors
// SPDX-License-Identifier: Apache-2.0

// Package proto is the line protocol between ESP-KVM and the agent: one JSON
// object per line, both ways. The device asks, the agent answers. The only
// thing the agent sends on its own is the hello, when the port opens.
// PROTOCOL.md has the whole of it.
package proto

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
)

// Version is the protocol version in the hello. A change that an older device
// would misread bumps it.
const Version = 1

// MaxLine is the longest line read. Longer ones are dropped whole.
const MaxLine = 1 << 20

// Request is what the device sends. Params depend on the op.
type Request struct {
	ID     int64           `json:"id"`
	Op     string          `json:"op"`
	Params json.RawMessage `json:"params,omitempty"`
}

// Response answers one request, by its id.
type Response struct {
	ID    int64  `json:"id"`
	OK    bool   `json:"ok"`
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
	Code  string `json:"code,omitempty"`
}

// Hello is sent when the port opens and when the device asks for it.
type Hello struct {
	Ev    string   `json:"ev"`
	Proto int      `json:"proto"`
	Agent string   `json:"agent"`
	OS    string   `json:"os"`
	Ops   []string `json:"ops"`
}

// Error carries a code the device can act on, beside the text for a person.
type Error struct {
	Code string
	Msg  string
}

func (e *Error) Error() string { return e.Msg }

// Errors with codes.
var (
	ErrUnknownOp  = &Error{"unknown_op", "this agent does not know that op"}
	ErrNotAllowed = &Error{"not_allowed", "switched off on this machine"}
	ErrBadParams  = &Error{"bad_params", "bad params"}
	ErrLocked     = &Error{"locked", "the screen is locked"}
)

// Handler runs one op. What it returns goes out as the response's data.
type Handler func(params json.RawMessage) (any, error)

// Server answers requests on one port.
type Server struct {
	AgentVersion string
	OS           string
	ops          map[string]Handler
	names        []string
}

// Handle adds an op.
func (s *Server) Handle(op string, h Handler) {
	if s.ops == nil {
		s.ops = map[string]Handler{}
	}
	s.ops[op] = h
	s.names = append(s.names, op)
}

func (s *Server) hello() Hello {
	return Hello{Ev: "hello", Proto: Version, Agent: s.AgentVersion, OS: s.OS, Ops: s.names}
}

// Serve reads requests from rw and answers them until rw fails. A line that is
// not a JSON object is skipped: whatever was on the port before the agent
// opened it must not stop it.
func (s *Server) Serve(rw io.ReadWriter) error {
	var mu sync.Mutex
	send := func(v any) error {
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		mu.Lock()
		defer mu.Unlock()
		_, err = rw.Write(append(b, '\n'))
		return err
	}
	// The newline first ends whatever half line the device may hold.
	if _, err := rw.Write([]byte("\n")); err != nil {
		return err
	}
	if err := send(s.hello()); err != nil {
		return err
	}
	r := bufio.NewReaderSize(rw, 64*1024)
	for {
		line, err := readLine(r)
		if err != nil {
			return err
		}
		line = bytes.TrimSpace(line)
		if len(line) == 0 || line[0] != '{' {
			continue
		}
		var req Request
		if json.Unmarshal(line, &req) != nil || req.Op == "" {
			continue
		}
		if err := send(s.run(req)); err != nil {
			return err
		}
	}
}

func (s *Server) run(req Request) any {
	if req.Op == "hello" {
		return Response{ID: req.ID, OK: true, Data: s.hello()}
	}
	h, ok := s.ops[req.Op]
	if !ok {
		return fail(req.ID, ErrUnknownOp)
	}
	data, err := safeCall(h, req.Params)
	if err != nil {
		return fail(req.ID, err)
	}
	return Response{ID: req.ID, OK: true, Data: data}
}

func safeCall(h Handler, p json.RawMessage) (data any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("agent error: %v", r)
		}
	}()
	return h(p)
}

func fail(id int64, err error) Response {
	var e *Error
	if errors.As(err, &e) {
		return Response{ID: id, Error: e.Msg, Code: e.Code}
	}
	return Response{ID: id, Error: err.Error(), Code: "failed"}
}

// readLine returns one line without its end. A line over MaxLine is read to
// its end and dropped.
func readLine(r *bufio.Reader) ([]byte, error) {
	var buf []byte
	tooLong := false
	for {
		part, isPrefix, err := r.ReadLine()
		if err != nil {
			return nil, err
		}
		if !tooLong {
			buf = append(buf, part...)
			if len(buf) > MaxLine {
				tooLong, buf = true, nil
			}
		}
		if !isPrefix {
			if tooLong {
				return nil, nil
			}
			return buf, nil
		}
	}
}
