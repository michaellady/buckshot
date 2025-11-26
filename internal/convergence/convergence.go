// Package convergence detects when multi-agent rounds have stabilized.
package convergence

import (
	"github.com/michaellady/buckshot/internal/orchestrator"
)

// Detector determines if the multi-agent protocol has converged.
type Detector interface {
	// IsConverged returns true if the round indicates convergence.
	// A round is converged if ALL agents reported no changes.
	IsConverged(result orchestrator.RoundResult) bool

	// CheckConvergence analyzes a round result and updates internal state.
	// Returns true if convergence threshold has been met.
	CheckConvergence(result orchestrator.RoundResult) bool

	// Reset clears the convergence tracking state.
	Reset()

	// ConsecutiveNoChangeRounds returns the current count of consecutive
	// rounds where all agents reported no changes.
	ConsecutiveNoChangeRounds() int

	// SetThreshold sets the number of consecutive no-change rounds
	// required to declare convergence. Default is 1.
	SetThreshold(n int)
}

// defaultDetector is a stub implementation.
type defaultDetector struct {
	threshold          int
	consecutiveNoChange int
}

// NewDetector creates a new convergence detector.
func NewDetector() Detector {
	return &defaultDetector{
		threshold: 1, // Default: converge after 1 round of no changes
	}
}

// IsConverged returns true if the round indicates no changes from any agent.
func (d *defaultDetector) IsConverged(result orchestrator.RoundResult) bool {
	// Stub: Always returns false
	return false
}

// CheckConvergence analyzes a round and returns true if threshold met.
func (d *defaultDetector) CheckConvergence(result orchestrator.RoundResult) bool {
	// Stub: Always returns false
	return false
}

// Reset clears the convergence tracking state.
func (d *defaultDetector) Reset() {
	d.consecutiveNoChange = 0
}

// ConsecutiveNoChangeRounds returns the current count.
func (d *defaultDetector) ConsecutiveNoChangeRounds() int {
	return d.consecutiveNoChange
}

// SetThreshold sets the convergence threshold.
func (d *defaultDetector) SetThreshold(n int) {
	if n < 1 {
		n = 1
	}
	d.threshold = n
}

// ParseNoChangeSignal checks if agent output indicates no changes were made.
// Looks for phrases like "no changes", "complete", "nothing to do", etc.
func ParseNoChangeSignal(output string) bool {
	// Stub: Always returns false
	return false
}
