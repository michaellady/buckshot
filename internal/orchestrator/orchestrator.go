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

// defaultOrchestrator is the default implementation.
type defaultOrchestrator struct {
	sessionMgr     session.Manager
	contextBuilder buckctx.Builder
}

// NewRoundOrchestrator creates a new round orchestrator.
func NewRoundOrchestrator() RoundOrchestrator {
	return &defaultOrchestrator{}
}

// RunRound executes agents in sequence.
// Each agent sees the beads state AFTER previous agents in the round.
func (o *defaultOrchestrator) RunRound(ctx context.Context, agents []agent.Agent, planCtx buckctx.PlanningContext) (RoundResult, error) {
	result := RoundResult{
		Round:        planCtx.Round,
		AgentResults: make([]AgentResult, 0, len(agents)),
	}

	// Process each agent in sequence
	for i, ag := range agents {
		agentResult := AgentResult{
			Agent:        ag,
			BeadsChanged: []string{},
		}

		// Skip unauthenticated agents
		if !ag.Authenticated {
			agentResult.Skipped = true
			result.SkippedCount++
			result.AgentResults = append(result.AgentResults, agentResult)
			continue
		}

		// Refresh beads state before each agent (except first which already has it)
		if i > 0 && o.contextBuilder != nil {
			_ = o.contextBuilder.RefreshBeadsState(&planCtx)
		}

		// Create session for this agent
		if o.sessionMgr == nil {
			agentResult.Error = context.Canceled
			result.FailedCount++
			result.AgentResults = append(result.AgentResults, agentResult)
			continue
		}

		sess, err := o.sessionMgr.CreateSession(ag)
		if err != nil {
			agentResult.Error = err
			result.FailedCount++
			result.AgentResults = append(result.AgentResults, agentResult)
			continue
		}
		defer sess.Close()

		// Start the session
		if err := sess.Start(ctx, planCtx.AgentsPath); err != nil {
			agentResult.Error = err
			result.FailedCount++
			result.AgentResults = append(result.AgentResults, agentResult)
			continue
		}

		// Format and send the prompt
		prompt := planCtx.Prompt
		if o.contextBuilder != nil {
			prompt = o.contextBuilder.Format(planCtx)
		}

		resp, err := sess.Send(ctx, prompt)
		if err != nil {
			agentResult.Error = err
			agentResult.Response = resp
			result.FailedCount++
			result.AgentResults = append(result.AgentResults, agentResult)
			continue
		}

		agentResult.Response = resp

		// Parse response for bead changes (simplified: look for bead IDs in output)
		agentResult.BeadsChanged = parseBeadChanges(resp.Output)
		result.TotalChanges += len(agentResult.BeadsChanged)

		result.AgentResults = append(result.AgentResults, agentResult)
	}

	// Refresh beads state after all agents for next round
	if o.contextBuilder != nil && len(agents) > 0 {
		_ = o.contextBuilder.RefreshBeadsState(&planCtx)
	}

	return result, nil
}

// parseBeadChanges extracts bead IDs from agent output.
// Looks for patterns like "buckshot-xxx" or "Created: buckshot-xxx"
func parseBeadChanges(output string) []string {
	// Simple implementation - can be enhanced with regex
	// For now, returns empty slice (changes tracked by beads diff)
	return []string{}
}

// SetSessionManager sets the session manager.
func (o *defaultOrchestrator) SetSessionManager(mgr session.Manager) {
	o.sessionMgr = mgr
}

// SetContextBuilder sets the context builder.
func (o *defaultOrchestrator) SetContextBuilder(builder buckctx.Builder) {
	o.contextBuilder = builder
}
