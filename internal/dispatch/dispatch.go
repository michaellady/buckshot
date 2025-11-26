// Package dispatch provides parallel agent dispatch and result collection.
package dispatch

import (
	"context"
	"sort"
	"sync"

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
	if len(sessions) == 0 {
		return []Result{}
	}

	// Channel to collect results from goroutines
	resultCh := make(chan Result, len(sessions))

	// WaitGroup to track completion of all goroutines
	var wg sync.WaitGroup

	// Fan-out: spawn a goroutine for each session
	for _, sess := range sessions {
		wg.Add(1)
		go func(s session.Session) {
			defer wg.Done()

			result := Result{
				Agent: s.Agent(),
			}

			// Send prompt and capture response/error
			resp, err := s.Send(ctx, prompt)
			result.Response = resp
			result.Error = err

			resultCh <- result
		}(sess)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Fan-in: collect all results
	results := make([]Result, 0, len(sessions))
	for result := range resultCh {
		results = append(results, result)
	}

	// Sort by agent name for deterministic output
	sort.Slice(results, func(i, j int) bool {
		return results[i].Agent.Name < results[j].Agent.Name
	})

	return results
}
