// Package context provides planning context building for agents.
package context

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// PlanningContext represents the context sent to an agent.
type PlanningContext struct {
	Prompt       string // The user's original prompt
	AgentsPath   string // Path to AGENTS.md for agent to read
	BeadsState   string // Current state of beads (bd list + bd show)
	Round        int    // Current round number
	IsFirstTurn  bool   // Whether this is the first agent in the protocol
	FeedbackMode bool   // Whether agent is in comment-only feedback mode
	AgentName    string // Name of the agent (used as comment author in feedback mode)
}

// Builder constructs planning contexts for agents.
type Builder interface {
	// Build creates a planning context for an agent.
	Build(prompt string, agentsPath string, round int, isFirstTurn bool) (PlanningContext, error)

	// Format converts a PlanningContext to a prompt string.
	Format(ctx PlanningContext) string

	// FormatFeedback converts a PlanningContext to a feedback-only prompt string.
	// In feedback mode, agents can only add comments to beads, not modify them.
	FormatFeedback(ctx PlanningContext) string

	// RefreshBeadsState updates the beads state in the context.
	RefreshBeadsState(ctx *PlanningContext) error
}

// defaultBuilder is the default implementation of Builder.
type defaultBuilder struct{}

// NewBuilder creates a new Builder instance.
func NewBuilder() Builder {
	return &defaultBuilder{}
}

// Build creates a planning context.
func (b *defaultBuilder) Build(prompt string, agentsPath string, round int, isFirstTurn bool) (PlanningContext, error) {
	ctx := PlanningContext{
		Prompt:      prompt,
		AgentsPath:  agentsPath,
		Round:       round,
		IsFirstTurn: isFirstTurn,
	}

	// Gather current beads state
	if err := b.RefreshBeadsState(&ctx); err != nil {
		return ctx, err
	}

	return ctx, nil
}

// Format converts a PlanningContext to a prompt string.
func (b *defaultBuilder) Format(ctx PlanningContext) string {
	var buf bytes.Buffer

	// First turn includes guidance to read AGENTS.md
	if ctx.IsFirstTurn {
		fmt.Fprintf(&buf, "please read and apply %s\n\n", ctx.AgentsPath)
	}

	// Show round number for subsequent rounds
	if ctx.Round > 1 {
		fmt.Fprintf(&buf, "## Round %d\n\n", ctx.Round)
	}

	// User's prompt
	fmt.Fprintf(&buf, "Prompt: %s\n\n", ctx.Prompt)

	// AGENTS.md path
	fmt.Fprintf(&buf, "AGENTS.md: %s\n\n", ctx.AgentsPath)

	// Current beads state
	fmt.Fprintf(&buf, "Current Beads:\n%s\n\n", ctx.BeadsState)

	// Instructions for modifying beads
	fmt.Fprintln(&buf, "Instructions:")
	fmt.Fprintln(&buf, "- Use `bd create` to create new beads")
	fmt.Fprintln(&buf, "- Use `bd update` to modify existing beads")
	fmt.Fprintln(&buf, "- Use `bd close` to close completed beads")
	fmt.Fprintln(&buf, "- Report changes made and whether plan seems complete")

	return buf.String()
}

// FormatFeedback converts a PlanningContext to a feedback-only prompt string.
// In feedback mode, agents can only add comments to beads, not modify them.
func (b *defaultBuilder) FormatFeedback(ctx PlanningContext) string {
	var buf bytes.Buffer

	// First turn includes guidance to read AGENTS.md
	if ctx.IsFirstTurn {
		fmt.Fprintf(&buf, "please read and apply %s\n\n", ctx.AgentsPath)
	}

	// Main feedback instruction
	fmt.Fprintln(&buf, "## Feedback Mode (Comment-Only)")
	fmt.Fprintln(&buf, "")
	fmt.Fprintln(&buf, "Please ultrathink to read and analyze the repository and the beads task descriptions and comments.")
	fmt.Fprintln(&buf, "Leave comments with your CLI name as the author in any issues that require your input that is")
	fmt.Fprintln(&buf, "substantially different or better from the content that is already there.")
	fmt.Fprintln(&buf, "")
	fmt.Fprintln(&buf, "**IMPORTANT: Do not edit the description or anything else related to the beads besides adding your comments.**")
	fmt.Fprintln(&buf, "")

	// Agent identification
	fmt.Fprintf(&buf, "Your agent name: %s\n\n", ctx.AgentName)

	// AGENTS.md path
	fmt.Fprintf(&buf, "AGENTS.md: %s\n\n", ctx.AgentsPath)

	// Current beads state
	fmt.Fprintf(&buf, "Current Beads:\n%s\n\n", ctx.BeadsState)

	// Instructions for commenting only
	fmt.Fprintln(&buf, "Instructions:")
	fmt.Fprintf(&buf, "- Use `bd comment <issue-id> \"<comment>\" --author %s` to add comments\n", ctx.AgentName)
	fmt.Fprintln(&buf, "- Only comment on issues where you have substantive input that is different or better")
	fmt.Fprintln(&buf, "- Do not use `bd update` or `bd create` - this is comment-only mode")
	fmt.Fprintln(&buf, "- Read existing comments before adding yours to avoid redundancy")

	return buf.String()
}

// RefreshBeadsState updates the beads state in the context.
func (b *defaultBuilder) RefreshBeadsState(ctx *PlanningContext) error {
	var buf bytes.Buffer

	// Get bd list output
	listCmd := exec.Command("bd", "list")
	listOut, err := listCmd.Output()
	if err != nil {
		// If bd is not available or fails, use empty state
		ctx.BeadsState = "(No beads found or bd command unavailable)"
		return nil
	}

	fmt.Fprintf(&buf, "=== Beads List ===\n%s\n", string(listOut))

	// Parse bd list to get issue IDs
	issueIDs := parseIssueIDs(string(listOut))

	// Get detailed info for each bead
	if len(issueIDs) > 0 {
		fmt.Fprintf(&buf, "\n=== Bead Details ===\n")
		for _, id := range issueIDs {
			showCmd := exec.Command("bd", "show", id)
			showOut, err := showCmd.Output()
			if err != nil {
				continue
			}
			fmt.Fprintf(&buf, "\n%s\n", string(showOut))
		}
	}

	ctx.BeadsState = buf.String()
	return nil
}

// parseIssueIDs extracts issue IDs from bd list output.
// Format: "ISSUE-ID [P#] [type] status - Title"
func parseIssueIDs(listOutput string) []string {
	var ids []string
	lines := strings.Split(listOutput, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Extract first field (issue ID) before space
		parts := strings.Fields(line)
		if len(parts) > 0 {
			id := parts[0]
			// Basic validation: should contain a hyphen
			if strings.Contains(id, "-") {
				ids = append(ids, id)
			}
		}
	}

	return ids
}
