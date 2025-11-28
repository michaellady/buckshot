package cli

import (
	"sync"

	"github.com/michaellady/buckshot/internal/agent"
)

// agentDetectorMu protects agentDetector from concurrent access in tests
//
//nolint:unused // Used by integration tests (//go:build integration)
var agentDetectorMu sync.Mutex

// setAgentDetector safely sets the agent detector function for testing.
// It returns a cleanup function that restores the original detector and releases the mutex.
// Usage:
//
//	restore := setAgentDetector(func() ([]agent.Agent, error) {
//	    return []agent.Agent{mockAgent}, nil
//	})
//	defer restore()
//
//nolint:unused // Used by integration tests (//go:build integration)
func setAgentDetector(fn func() ([]agent.Agent, error)) func() {
	agentDetectorMu.Lock()
	orig := agentDetector
	agentDetector = fn
	return func() {
		agentDetector = orig
		agentDetectorMu.Unlock()
	}
}

// resetPlanFlags resets all plan command flags to their default values.
// This MUST be called at the start of each integration test to ensure clean state.
//
//nolint:unused // Used by integration tests (//go:build integration)
func resetPlanFlags() {
	selectedAgents = nil
	untilConverged = false
	rounds = 3
	agentsPath = ""
	saveToBead = ""
	verbose = false
}

// resetFeedbackFlags resets all feedback command flags to their default values.
// This MUST be called at the start of each integration test to ensure clean state.
//
//nolint:unused // Used by integration tests (//go:build integration)
func resetFeedbackFlags() {
	feedbackAgent = ""
	agentsPath = ""
}
