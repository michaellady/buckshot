// Package convergence detects when multi-agent rounds have stabilized.
package convergence

import (
	"regexp"
	"strings"

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
// Skipped and failed agents are ignored - only successful agents count.
func (d *defaultDetector) IsConverged(result orchestrator.RoundResult) bool {
	// If TotalChanges > 0, definitely not converged
	if result.TotalChanges > 0 {
		return false
	}

	// Check each agent result
	for _, ar := range result.AgentResults {
		// Skip skipped agents
		if ar.Skipped {
			continue
		}
		// Skip failed agents
		if ar.Error != nil {
			continue
		}
		// If any successful agent made changes, not converged
		if len(ar.BeadsChanged) > 0 {
			return false
		}
	}

	// All successful agents made no changes (or no agents ran)
	return true
}

// CheckConvergence analyzes a round and returns true if threshold met.
func (d *defaultDetector) CheckConvergence(result orchestrator.RoundResult) bool {
	if d.IsConverged(result) {
		d.consecutiveNoChange++
	} else {
		d.consecutiveNoChange = 0
	}

	return d.consecutiveNoChange >= d.threshold
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

// noChangePatterns matches phrases indicating no changes were made
var noChangePatterns = regexp.MustCompile(`(?i)(no\s+changes|nothing\s+to\s+do|all\s+tasks\s+(are\s+)?done|everything\s+is\s+complete|complete)`)

// ParseNoChangeSignal checks if agent output indicates no changes were made.
// Looks for phrases like "no changes", "complete", "nothing to do", etc.
func ParseNoChangeSignal(output string) bool {
	if output == "" {
		return false
	}
	lower := strings.ToLower(output)
	return noChangePatterns.MatchString(lower)
}
