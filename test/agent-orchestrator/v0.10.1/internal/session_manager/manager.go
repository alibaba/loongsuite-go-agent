// Package sessionmanager mirrors the public surface of the agent-orchestrator
// daemon's internal/session_manager package so the instrumentation rule can be
// exercised against a representative shape. The real daemon owns much richer
// types; only the fields the hook reads need to line up.
package sessionmanager

import (
	"context"
	"errors"
	"sync/atomic"
)

// ID types are distinct string types so they can't be swapped at a call site.
type (
	SessionID string
	ProjectID string
	IssueID   string
)

// AgentHarness names a CLI coding agent adapter (claude-code, codex, ...).
type AgentHarness string

// SessionKind distinguishes worker sessions from orchestrator sessions.
type SessionKind string

// Session kinds.
const (
	KindWorker       SessionKind = "worker"
	KindOrchestrator SessionKind = "orchestrator"
)

// SpawnConfig is the request to start a new session. It carries the same fields
// the daemon's ports.SpawnConfig exposes; the hook reads them via reflection.
type SpawnConfig struct {
	ProjectID ProjectID
	IssueID   IssueID
	Kind      SessionKind
	Harness   AgentHarness
	Branch    string
	Prompt    string
}

// SessionRecord is the persistence shape returned by Spawn.
type SessionRecord struct {
	ID          SessionID
	ProjectID   ProjectID
	IssueID     IssueID
	Kind        SessionKind
	Harness     AgentHarness
	DisplayName string
}

// Manager drives session command operations. The real Manager wires runtime,
// agent, workspace, storage, messenger, and lifecycle dependencies; this stub
// keeps only the bookkeeping needed for Spawn/Send to behave deterministically.
type Manager struct {
	counter atomic.Int64
}

// New returns a stub Manager.
func New() *Manager { return &Manager{} }

// Spawn creates a session record. The real implementation launches a runtime
// and provisions a workspace; here we just synthesize an ID so callers can
// exercise the success path.
func (m *Manager) Spawn(ctx context.Context, cfg SpawnConfig) (SessionRecord, error) {
	_ = ctx
	if cfg.Harness == "" {
		return SessionRecord{}, errors.New("spawn: harness required")
	}
	n := m.counter.Add(1)
	return SessionRecord{
		ID:        SessionID("sess-" + string(cfg.Harness) + "-" + itoa(int(n))),
		ProjectID: cfg.ProjectID,
		IssueID:   cfg.IssueID,
		Kind:      cfg.Kind,
		Harness:   cfg.Harness,
	}, nil
}

// Send delivers a message to a running session.
func (m *Manager) Send(ctx context.Context, id SessionID, message string) error {
	_ = ctx
	_ = m
	if id == "" {
		return errors.New("send: empty session id")
	}
	_ = message
	return nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
