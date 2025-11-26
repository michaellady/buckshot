// Package context provides planning context building for agents.
package context

// PlanningContext represents the context sent to an agent.
type PlanningContext struct {
	Prompt      string // The user's original prompt
	AgentsPath  string // Path to AGENTS.md for agent to read
	BeadsState  string // Current state of beads (bd list + bd show)
	Round       int    // Current round number
	IsFirstTurn bool   // Whether this is the first agent in the protocol
}

// Builder constructs planning contexts for agents.
type Builder interface {
	// Build creates a planning context for an agent.
	Build(prompt string, agentsPath string, round int, isFirstTurn bool) (PlanningContext, error)

	// Format converts a PlanningContext to a prompt string.
	Format(ctx PlanningContext) string

	// RefreshBeadsState updates the beads state in the context.
	RefreshBeadsState(ctx *PlanningContext) error
}
