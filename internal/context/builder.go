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

// defaultBuilder is a stub implementation of Builder.
type defaultBuilder struct{}

// NewBuilder creates a new Builder instance.
func NewBuilder() Builder {
	return &defaultBuilder{}
}

// Build creates a planning context (stub implementation).
func (b *defaultBuilder) Build(prompt string, agentsPath string, round int, isFirstTurn bool) (PlanningContext, error) {
	// TODO: Implement actual context building
	return PlanningContext{
		Prompt:      prompt,
		AgentsPath:  agentsPath,
		BeadsState:  "", // Stub: should gather bd list + bd show output
		Round:       round,
		IsFirstTurn: isFirstTurn,
	}, nil
}

// Format converts a PlanningContext to a prompt string (stub implementation).
func (b *defaultBuilder) Format(ctx PlanningContext) string {
	// TODO: Implement actual formatting with proper sections and instructions
	return "" // Stub: returns empty string
}

// RefreshBeadsState updates the beads state (stub implementation).
func (b *defaultBuilder) RefreshBeadsState(ctx *PlanningContext) error {
	// TODO: Implement actual refresh logic
	// Stub: does nothing
	return nil
}
