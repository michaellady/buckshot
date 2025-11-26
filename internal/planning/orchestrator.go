// Package planning provides the multi-agent planning protocol orchestration.
package planning

import (
	"context"

	"github.com/michaellady/buckshot/internal/session"
)

// AgentResult captures what happened when an agent took its turn.
type AgentResult struct {
	AgentName    string   // Name of the agent
	Changes      []string // Summary of changes made
	NoChanges    bool     // True if agent reported no changes needed
	Error        error    // Any error that occurred
	ContextUsage float64  // Context usage after this turn
}

// RoundResult captures the results of one planning round.
type RoundResult struct {
	RoundNumber  int           // Which round this was
	AgentResults []AgentResult // Results from each agent
	Converged    bool          // True if all agents reported no changes
}

// Orchestrator manages the multi-agent planning protocol.
type Orchestrator interface {
	// RunRound executes one planning round with all agents.
	RunRound(ctx context.Context, sessions []session.Session, prompt string, agentsPath string) (RoundResult, error)

	// RunProtocol executes the full planning protocol.
	RunProtocol(ctx context.Context, config Config) ([]RoundResult, error)
}

// Config holds configuration for a planning session.
type Config struct {
	Prompt         string            // The user's planning prompt
	AgentsPath     string            // Path to AGENTS.md
	MaxRounds      int               // Maximum number of rounds
	UntilConverged bool              // Run until convergence instead of fixed rounds
	Sessions       []session.Session // Active agent sessions
}
