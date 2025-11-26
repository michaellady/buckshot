// Package dispatch provides parallel agent dispatch and result collection.
package dispatch

import (
	"context"
	"sort"

	"github.com/michaellady/buckshot/internal/agent"
	"github.com/michaellady/buckshot/internal/session"
)

// Result represents the outcome of dispatching to a single agent.
type Result struct {
	Agent    agent.Agent      // The agent that was dispatched to
	Response session.Response // The agent's response
	Error    error            // Error if dispatch failed
}

// Dispatcher handles parallel dispatch to multiple agents.
type Dispatcher interface {
	// Dispatch sends a prompt to multiple agents concurrently and collects results.
	// Results are returned in deterministic order (sorted by agent name).
	// Respects context timeout/cancellation.
	Dispatch(ctx context.Context, sessions []session.Session, prompt string) []Result
}

// dispatcher is the default implementation.
type dispatcher struct{}

// New creates a new Dispatcher.
func New() Dispatcher {
	return &dispatcher{}
}

// Dispatch sends a prompt to multiple agents concurrently.
// Results are always returned sorted by agent name for deterministic output.
func (d *dispatcher) Dispatch(ctx context.Context, sessions []session.Session, prompt string) []Result {
	// TODO: Implement parallel dispatch
	// This is intentionally unimplemented to make tests fail (TDD RED phase)
	results := make([]Result, len(sessions))
	for i, sess := range sessions {
		results[i] = Result{
			Agent: sess.Agent(),
		}
	}

	// Sort by agent name for deterministic output
	sort.Slice(results, func(i, j int) bool {
		return results[i].Agent.Name < results[j].Agent.Name
	})

	return results
}
