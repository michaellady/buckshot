package convergence

import (
	"context"
	"testing"

	"github.com/michaellady/buckshot/internal/agent"
	"github.com/michaellady/buckshot/internal/orchestrator"
	"github.com/michaellady/buckshot/internal/session"
)

// TestDetectorInterface ensures the interface is properly defined
func TestDetectorInterface(t *testing.T) {
	var _ Detector = (*defaultDetector)(nil)
}

// TestIsConverged_AllAgentsNoChanges tests convergence when all agents report no changes
func TestIsConverged_AllAgentsNoChanges(t *testing.T) {
	detector := NewDetector()

	result := orchestrator.RoundResult{
		Round: 1,
		AgentResults: []orchestrator.AgentResult{
			{
				Agent:        agent.Agent{Name: "claude"},
				Response:     session.Response{Output: "No changes needed"},
				BeadsChanged: []string{}, // Empty = no changes
			},
			{
				Agent:        agent.Agent{Name: "codex"},
				Response:     session.Response{Output: "Everything looks good"},
				BeadsChanged: []string{}, // Empty = no changes
			},
		},
		TotalChanges: 0,
	}

	if !detector.IsConverged(result) {
		t.Error("IsConverged() = false, want true when all agents have no changes")
	}
}

// TestIsConverged_AnyAgentMadeChanges tests that any change means not converged
func TestIsConverged_AnyAgentMadeChanges(t *testing.T) {
	detector := NewDetector()

	result := orchestrator.RoundResult{
		Round: 1,
		AgentResults: []orchestrator.AgentResult{
			{
				Agent:        agent.Agent{Name: "claude"},
				Response:     session.Response{Output: "No changes needed"},
				BeadsChanged: []string{}, // No changes
			},
			{
				Agent:        agent.Agent{Name: "codex"},
				Response:     session.Response{Output: "Created new bead"},
				BeadsChanged: []string{"buckshot-abc"}, // Made a change!
			},
		},
		TotalChanges: 1,
	}

	if detector.IsConverged(result) {
		t.Error("IsConverged() = true, want false when any agent made changes")
	}
}

// TestIsConverged_UsesTotalChanges tests that TotalChanges > 0 means not converged
func TestIsConverged_UsesTotalChanges(t *testing.T) {
	detector := NewDetector()

	result := orchestrator.RoundResult{
		Round:        1,
		AgentResults: []orchestrator.AgentResult{},
		TotalChanges: 5, // Changes detected even if AgentResults is empty
	}

	if detector.IsConverged(result) {
		t.Error("IsConverged() = true, want false when TotalChanges > 0")
	}
}

// TestIsConverged_EmptyRound tests that empty round is considered converged
func TestIsConverged_EmptyRound(t *testing.T) {
	detector := NewDetector()

	result := orchestrator.RoundResult{
		Round:        1,
		AgentResults: []orchestrator.AgentResult{},
		TotalChanges: 0,
	}

	if !detector.IsConverged(result) {
		t.Error("IsConverged() = false, want true for empty round with no changes")
	}
}

// TestCheckConvergence_TracksConsecutiveRounds tests multi-round tracking
func TestCheckConvergence_TracksConsecutiveRounds(t *testing.T) {
	detector := NewDetector()
	detector.SetThreshold(3) // Require 3 consecutive no-change rounds

	noChangeResult := orchestrator.RoundResult{
		Round:        1,
		TotalChanges: 0,
		AgentResults: []orchestrator.AgentResult{
			{Agent: agent.Agent{Name: "claude"}, BeadsChanged: []string{}},
		},
	}

	// First no-change round
	if detector.CheckConvergence(noChangeResult) {
		t.Error("CheckConvergence() round 1 = true, want false (need 3 rounds)")
	}
	if detector.ConsecutiveNoChangeRounds() != 1 {
		t.Errorf("ConsecutiveNoChangeRounds() = %d, want 1", detector.ConsecutiveNoChangeRounds())
	}

	// Second no-change round
	noChangeResult.Round = 2
	if detector.CheckConvergence(noChangeResult) {
		t.Error("CheckConvergence() round 2 = true, want false (need 3 rounds)")
	}
	if detector.ConsecutiveNoChangeRounds() != 2 {
		t.Errorf("ConsecutiveNoChangeRounds() = %d, want 2", detector.ConsecutiveNoChangeRounds())
	}

	// Third no-change round - should now be converged
	noChangeResult.Round = 3
	if !detector.CheckConvergence(noChangeResult) {
		t.Error("CheckConvergence() round 3 = false, want true (hit threshold)")
	}
	if detector.ConsecutiveNoChangeRounds() != 3 {
		t.Errorf("ConsecutiveNoChangeRounds() = %d, want 3", detector.ConsecutiveNoChangeRounds())
	}
}

// TestCheckConvergence_ResetsOnChange tests that a change resets the counter
func TestCheckConvergence_ResetsOnChange(t *testing.T) {
	detector := NewDetector()
	detector.SetThreshold(3)

	noChangeResult := orchestrator.RoundResult{
		Round:        1,
		TotalChanges: 0,
		AgentResults: []orchestrator.AgentResult{
			{Agent: agent.Agent{Name: "claude"}, BeadsChanged: []string{}},
		},
	}

	changeResult := orchestrator.RoundResult{
		Round:        2,
		TotalChanges: 1,
		AgentResults: []orchestrator.AgentResult{
			{Agent: agent.Agent{Name: "claude"}, BeadsChanged: []string{"buckshot-xyz"}},
		},
	}

	// Two no-change rounds
	detector.CheckConvergence(noChangeResult)
	noChangeResult.Round = 2
	detector.CheckConvergence(noChangeResult)
	if detector.ConsecutiveNoChangeRounds() != 2 {
		t.Errorf("ConsecutiveNoChangeRounds() = %d, want 2", detector.ConsecutiveNoChangeRounds())
	}

	// A round with changes should reset counter
	changeResult.Round = 3
	detector.CheckConvergence(changeResult)
	if detector.ConsecutiveNoChangeRounds() != 0 {
		t.Errorf("ConsecutiveNoChangeRounds() after change = %d, want 0", detector.ConsecutiveNoChangeRounds())
	}
}

// TestSetThreshold_MinimumOne tests that threshold can't be less than 1
func TestSetThreshold_MinimumOne(t *testing.T) {
	detector := NewDetector().(*defaultDetector)

	detector.SetThreshold(0)
	if detector.threshold != 1 {
		t.Errorf("SetThreshold(0) resulted in threshold = %d, want 1", detector.threshold)
	}

	detector.SetThreshold(-5)
	if detector.threshold != 1 {
		t.Errorf("SetThreshold(-5) resulted in threshold = %d, want 1", detector.threshold)
	}
}

// TestReset_ClearsState tests that Reset clears the counter
func TestReset_ClearsState(t *testing.T) {
	detector := NewDetector()

	noChangeResult := orchestrator.RoundResult{
		Round:        1,
		TotalChanges: 0,
		AgentResults: []orchestrator.AgentResult{},
	}

	detector.CheckConvergence(noChangeResult)
	detector.CheckConvergence(noChangeResult)

	if detector.ConsecutiveNoChangeRounds() != 2 {
		t.Fatalf("ConsecutiveNoChangeRounds() = %d, want 2", detector.ConsecutiveNoChangeRounds())
	}

	detector.Reset()

	if detector.ConsecutiveNoChangeRounds() != 0 {
		t.Errorf("ConsecutiveNoChangeRounds() after Reset = %d, want 0", detector.ConsecutiveNoChangeRounds())
	}
}

// TestParseNoChangeSignal_DetectsNoChanges tests parsing agent output
func TestParseNoChangeSignal_DetectsNoChanges(t *testing.T) {
	testCases := []struct {
		output string
		want   bool
	}{
		{"No changes needed", true},
		{"no changes were made", true},
		{"Everything is complete", true},
		{"Nothing to do", true},
		{"all tasks are done", true},
		{"I've made some updates", false},
		{"Created buckshot-abc", false},
		{"Fixed the bug", false},
		{"", false},
	}

	for _, tc := range testCases {
		got := ParseNoChangeSignal(tc.output)
		if got != tc.want {
			t.Errorf("ParseNoChangeSignal(%q) = %v, want %v", tc.output, got, tc.want)
		}
	}
}

// TestIsConverged_SkippedAgentsDontCount tests that skipped agents are ignored
func TestIsConverged_SkippedAgentsDontCount(t *testing.T) {
	detector := NewDetector()

	result := orchestrator.RoundResult{
		Round: 1,
		AgentResults: []orchestrator.AgentResult{
			{
				Agent:        agent.Agent{Name: "claude"},
				BeadsChanged: []string{},
				Skipped:      false,
			},
			{
				Agent:        agent.Agent{Name: "codex"},
				BeadsChanged: []string{},
				Skipped:      true, // Skipped agent
			},
		},
		TotalChanges: 0,
		SkippedCount: 1,
	}

	// Should be converged - the only active agent made no changes
	if !detector.IsConverged(result) {
		t.Error("IsConverged() = false, want true (skipped agents shouldn't block convergence)")
	}
}

// TestIsConverged_FailedAgentsDontBlockConvergence tests failed agent handling
func TestIsConverged_FailedAgentsDontBlockConvergence(t *testing.T) {
	detector := NewDetector()

	result := orchestrator.RoundResult{
		Round: 1,
		AgentResults: []orchestrator.AgentResult{
			{
				Agent:        agent.Agent{Name: "claude"},
				BeadsChanged: []string{},
				Error:        nil,
			},
			{
				Agent:        agent.Agent{Name: "codex"},
				BeadsChanged: []string{},
				Error:        context.DeadlineExceeded, // Failed
			},
		},
		TotalChanges: 0,
		FailedCount:  1,
	}

	// Should be converged - the successful agent made no changes
	// Failed agents don't contribute to convergence decision
	if !detector.IsConverged(result) {
		t.Error("IsConverged() = false, want true (failed agents shouldn't block convergence)")
	}
}

