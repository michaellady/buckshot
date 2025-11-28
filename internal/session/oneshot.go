package session

import (
	"context"

	"github.com/michaellady/buckshot/internal/agent"
)

// OneShotResult represents the result of a one-shot agent execution.
type OneShotResult struct {
	Output   string // Combined stdout/stderr output
	ExitCode int    // Process exit code
	Error    error  // Any error during execution
}

// RunOneShot executes an agent in one-shot mode and waits for completion.
// This is used for agents that run a single prompt and exit (auggie --print,
// amp --execute, gemini positional, codex exec).
//
// Unlike interactive sessions, one-shot execution:
// - Builds command with prompt as argument
// - Runs synchronously until process exits
// - Captures all output
// - Returns when process completes
func RunOneShot(ctx context.Context, ag agent.Agent, prompt string) (OneShotResult, error) {
	// TODO: Implement one-shot execution
	return OneShotResult{}, nil
}
