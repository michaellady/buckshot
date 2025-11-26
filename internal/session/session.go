// Package session provides persistent agent session management.
package session

import (
	"context"

	"github.com/michaellady/buckshot/internal/agent"
)

// Response represents an agent's response to a prompt.
type Response struct {
	Output       string  // The agent's output
	ContextUsage float64 // Context usage as 0.0-1.0
	Error        error   // Any error that occurred
}

// Session represents a persistent connection to an AI agent.
type Session interface {
	// Start initializes the session with the path to AGENTS.md.
	Start(ctx context.Context, agentsPath string) error

	// Send sends a prompt to the agent and returns the response.
	Send(ctx context.Context, prompt string) (Response, error)

	// ContextUsage returns the current context usage (0.0 to 1.0).
	ContextUsage() float64

	// IsAlive returns whether the session is still active.
	IsAlive() bool

	// Agent returns the underlying agent for this session.
	Agent() agent.Agent

	// Close terminates the session.
	Close() error
}

// Manager handles creation and lifecycle of agent sessions.
type Manager interface {
	// CreateSession creates a new session for the given agent.
	CreateSession(agent agent.Agent) (Session, error)

	// ShouldRespawn returns true if session context > threshold.
	ShouldRespawn(session Session, threshold float64) bool
}
