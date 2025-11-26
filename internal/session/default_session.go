package session

import (
	"context"
	"errors"

	"github.com/michaellady/buckshot/internal/agent"
)

// DefaultSession is a stub implementation of Session interface.
// This is the RED phase of TDD - tests should fail.
type DefaultSession struct {
	agent agent.Agent
}

// Start initializes the session with the path to AGENTS.md.
func (s *DefaultSession) Start(ctx context.Context, agentsPath string) error {
	// Stub: Always fails
	return errors.New("not implemented")
}

// Send sends a prompt to the agent and returns the response.
func (s *DefaultSession) Send(ctx context.Context, prompt string) (Response, error) {
	// Stub: Returns empty response with error
	return Response{}, errors.New("not implemented")
}

// ContextUsage returns the current context usage (0.0 to 1.0).
func (s *DefaultSession) ContextUsage() float64 {
	// Stub: Always returns 0
	return 0.0
}

// IsAlive returns whether the session is still active.
func (s *DefaultSession) IsAlive() bool {
	// Stub: Always returns false
	return false
}

// Agent returns the underlying agent for this session.
func (s *DefaultSession) Agent() agent.Agent {
	return s.agent
}

// Close terminates the session.
func (s *DefaultSession) Close() error {
	// Stub: Always succeeds (no-op)
	return nil
}

// DefaultManager is a stub implementation of Manager interface.
type DefaultManager struct{}

// NewManager creates a new session manager.
func NewManager() Manager {
	return &DefaultManager{}
}

// CreateSession creates a new session for the given agent.
func (m *DefaultManager) CreateSession(agent agent.Agent) (Session, error) {
	// Stub: Always fails for unauthenticated agents
	if !agent.Authenticated {
		return nil, errors.New("agent not authenticated")
	}
	// Returns a stub session that will fail tests
	return &DefaultSession{agent: agent}, nil
}

// ShouldRespawn returns true if session context > threshold.
func (m *DefaultManager) ShouldRespawn(session Session, threshold float64) bool {
	// Stub: Simple comparison
	return session.ContextUsage() > threshold
}
