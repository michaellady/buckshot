// Package orchestrator coordinates multi-agent round execution.
package orchestrator

import (
	"context"

	"github.com/michaellady/buckshot/internal/agent"
	buckctx "github.com/michaellady/buckshot/internal/context"
	"github.com/michaellady/buckshot/internal/session"
)

// AgentResult represents the outcome of a single agent's turn.
type AgentResult struct {
	Agent         agent.Agent       // The agent that ran
	Response      session.Response  // The agent's response
	BeadsChanged  []string          // IDs of beads created/modified
	Error         error             // Error if agent failed
	Skipped       bool              // True if agent was skipped (e.g., due to previous failure)
}

// RoundResult represents the outcome of a complete round.
type RoundResult struct {
	Round         int            // Round number (1-indexed)
	AgentResults  []AgentResult  // Results from each agent
	TotalChanges  int            // Total beads created/modified
	FailedCount   int            // Number of agents that failed
	SkippedCount  int            // Number of agents that were skipped
}

// RoundOrchestrator coordinates executing multiple agents in a round.
type RoundOrchestrator interface {
	// RunRound executes each agent in sequence with the given context.
	// Each agent sees the beads state AFTER previous agents in the round.
	RunRound(ctx context.Context, agents []agent.Agent, planCtx buckctx.PlanningContext) (RoundResult, error)

	// SetSessionManager sets the session manager for creating agent sessions.
	SetSessionManager(mgr session.Manager)

	// SetContextBuilder sets the context builder for refreshing beads state.
	SetContextBuilder(builder buckctx.Builder)
}

// defaultOrchestrator is a stub implementation.
type defaultOrchestrator struct {
	sessionMgr     session.Manager
	contextBuilder buckctx.Builder
}

// NewRoundOrchestrator creates a new round orchestrator.
func NewRoundOrchestrator() RoundOrchestrator {
	return &defaultOrchestrator{}
}

// RunRound executes agents in sequence (stub implementation).
func (o *defaultOrchestrator) RunRound(ctx context.Context, agents []agent.Agent, planCtx buckctx.PlanningContext) (RoundResult, error) {
	// Stub: Returns empty result
	return RoundResult{Round: planCtx.Round}, nil
}

// SetSessionManager sets the session manager.
func (o *defaultOrchestrator) SetSessionManager(mgr session.Manager) {
	o.sessionMgr = mgr
}

// SetContextBuilder sets the context builder.
func (o *defaultOrchestrator) SetContextBuilder(builder buckctx.Builder) {
	o.contextBuilder = builder
}
