// Package orchestrator coordinates multi-agent round execution.
package orchestrator

import (
	"context"
	"os/exec"

	"github.com/michaellady/buckshot/internal/agent"
	buckctx "github.com/michaellady/buckshot/internal/context"
	"github.com/michaellady/buckshot/internal/session"
)

// newOSCmd wraps exec.Command for shell execution
func newOSCmd(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

// ProgressReporter receives progress updates during round execution.
type ProgressReporter interface {
	// OnAgentStart is called when an agent begins its turn.
	OnAgentStart(round, agentIndex, totalAgents int, agent agent.Agent)
	// OnAgentComplete is called when an agent finishes its turn.
	OnAgentComplete(round, agentIndex, totalAgents int, result AgentResult, beadsDiff string)
}

// AgentResult represents the outcome of a single agent's turn.
type AgentResult struct {
	Agent         agent.Agent       // The agent that ran
	Response      session.Response  // The agent's response
	BeadsChanged  []string          // IDs of beads created/modified
	Error         error             // Error if agent failed
	Skipped       bool              // True if agent was skipped (e.g., due to previous failure)
}

// RoundResult represents the outcome of a complete round.
type RoundResult struct {
	Round         int            // Round number (1-indexed)
	AgentResults  []AgentResult  // Results from each agent
	TotalChanges  int            // Total beads created/modified
	FailedCount   int            // Number of agents that failed
	SkippedCount  int            // Number of agents that were skipped
}

// RoundOrchestrator coordinates executing multiple agents in a round.
type RoundOrchestrator interface {
	// RunRound executes each agent in sequence with the given context.
	// Each agent sees the beads state AFTER previous agents in the round.
	RunRound(ctx context.Context, agents []agent.Agent, planCtx buckctx.PlanningContext) (RoundResult, error)

	// SetSessionManager sets the session manager for creating agent sessions.
	SetSessionManager(mgr session.Manager)

	// SetContextBuilder sets the context builder for refreshing beads state.
	SetContextBuilder(builder buckctx.Builder)

	// SetProgressReporter sets the progress reporter for verbose output.
	SetProgressReporter(reporter ProgressReporter)
}

// defaultOrchestrator is the default implementation.
type defaultOrchestrator struct {
	sessionMgr       session.Manager
	contextBuilder   buckctx.Builder
	progressReporter ProgressReporter
}

// NewRoundOrchestrator creates a new round orchestrator.
func NewRoundOrchestrator() RoundOrchestrator {
	return &defaultOrchestrator{}
}

// RunRound executes agents in sequence.
// Each agent sees the beads state AFTER previous agents in the round.
func (o *defaultOrchestrator) RunRound(ctx context.Context, agents []agent.Agent, planCtx buckctx.PlanningContext) (RoundResult, error) {
	result := RoundResult{
		Round:        planCtx.Round,
		AgentResults: make([]AgentResult, 0, len(agents)),
	}

	// Process each agent in sequence
	for i, ag := range agents {
		agentResult := AgentResult{
			Agent:        ag,
			BeadsChanged: []string{},
		}

		// Skip unauthenticated agents
		if !ag.Authenticated {
			agentResult.Skipped = true
			result.SkippedCount++
			result.AgentResults = append(result.AgentResults, agentResult)
			if o.progressReporter != nil {
				o.progressReporter.OnAgentComplete(planCtx.Round, i+1, len(agents), agentResult, "")
			}
			continue
		}

		// Report agent start
		if o.progressReporter != nil {
			o.progressReporter.OnAgentStart(planCtx.Round, i+1, len(agents), ag)
		}

		// Capture beads state before this agent
		beadsBefore := captureBeadsState()

		// Refresh beads state before each agent (except first which already has it)
		if i > 0 && o.contextBuilder != nil {
			_ = o.contextBuilder.RefreshBeadsState(&planCtx)
		}

		// Create session for this agent
		if o.sessionMgr == nil {
			agentResult.Error = context.Canceled
			result.FailedCount++
			result.AgentResults = append(result.AgentResults, agentResult)
			if o.progressReporter != nil {
				o.progressReporter.OnAgentComplete(planCtx.Round, i+1, len(agents), agentResult, "")
			}
			continue
		}

		sess, err := o.sessionMgr.CreateSession(ag)
		if err != nil {
			agentResult.Error = err
			result.FailedCount++
			result.AgentResults = append(result.AgentResults, agentResult)
			if o.progressReporter != nil {
				o.progressReporter.OnAgentComplete(planCtx.Round, i+1, len(agents), agentResult, "")
			}
			continue
		}
		defer func() { _ = sess.Close() }()

		// Start the session
		if err := sess.Start(ctx, planCtx.AgentsPath); err != nil {
			agentResult.Error = err
			result.FailedCount++
			result.AgentResults = append(result.AgentResults, agentResult)
			if o.progressReporter != nil {
				o.progressReporter.OnAgentComplete(planCtx.Round, i+1, len(agents), agentResult, "")
			}
			continue
		}

		// Format and send the prompt
		prompt := planCtx.Prompt
		if o.contextBuilder != nil {
			prompt = o.contextBuilder.Format(planCtx)
		}

		resp, err := sess.Send(ctx, prompt)
		if err != nil {
			agentResult.Error = err
			agentResult.Response = resp
			result.FailedCount++
			result.AgentResults = append(result.AgentResults, agentResult)
			if o.progressReporter != nil {
				beadsAfter := captureBeadsState()
				diff := diffBeadsState(beadsBefore, beadsAfter)
				o.progressReporter.OnAgentComplete(planCtx.Round, i+1, len(agents), agentResult, diff)
			}
			continue
		}

		agentResult.Response = resp

		// Parse response for bead changes (simplified: look for bead IDs in output)
		agentResult.BeadsChanged = parseBeadChanges(resp.Output)
		result.TotalChanges += len(agentResult.BeadsChanged)

		result.AgentResults = append(result.AgentResults, agentResult)

		// Report agent complete with beads diff
		if o.progressReporter != nil {
			beadsAfter := captureBeadsState()
			diff := diffBeadsState(beadsBefore, beadsAfter)
			o.progressReporter.OnAgentComplete(planCtx.Round, i+1, len(agents), agentResult, diff)
		}
	}

	// Refresh beads state after all agents for next round
	if o.contextBuilder != nil && len(agents) > 0 {
		_ = o.contextBuilder.RefreshBeadsState(&planCtx)
	}

	return result, nil
}

// parseBeadChanges extracts bead IDs from agent output.
// Looks for patterns like "buckshot-xxx" or "Created: buckshot-xxx"
func parseBeadChanges(output string) []string {
	// Simple implementation - can be enhanced with regex
	// For now, returns empty slice (changes tracked by beads diff)
	return []string{}
}

// SetSessionManager sets the session manager.
func (o *defaultOrchestrator) SetSessionManager(mgr session.Manager) {
	o.sessionMgr = mgr
}

// SetContextBuilder sets the context builder.
func (o *defaultOrchestrator) SetContextBuilder(builder buckctx.Builder) {
	o.contextBuilder = builder
}

// SetProgressReporter sets the progress reporter.
func (o *defaultOrchestrator) SetProgressReporter(reporter ProgressReporter) {
	o.progressReporter = reporter
}

// captureBeadsState captures the current beads state by running `bd list --json`.
func captureBeadsState() string {
	out, err := runBdCommand("list", "--json")
	if err != nil {
		return ""
	}
	return out
}

// diffBeadsState computes a human-readable diff between two beads states.
func diffBeadsState(before, after string) string {
	if before == after {
		return "(no changes)"
	}
	if before == "" && after == "" {
		return "(no beads)"
	}
	if before == "" {
		return "(beads initialized)\n" + after
	}
	if after == "" {
		return "(beads cleared)"
	}
	// For now, just show a simple diff indicator
	// A more sophisticated diff could parse JSON and compare fields
	return computeSimpleDiff(before, after)
}

// runBdCommand executes a bd command and returns its output.
func runBdCommand(args ...string) (string, error) {
	// Import os/exec inline to avoid adding to package imports
	// This is a simple helper that shells out to bd
	cmd := execCommand("bd", args...)
	out, err := cmd.Output()
	return string(out), err
}

// execCommand is a variable for testing - allows mocking exec.Command
var execCommand = defaultExecCommand

func defaultExecCommand(name string, args ...string) cmdRunner {
	return &realCmd{name: name, args: args}
}

type cmdRunner interface {
	Output() ([]byte, error)
}

type realCmd struct {
	name string
	args []string
}

func (c *realCmd) Output() ([]byte, error) {
	cmd := newOSCmd(c.name, c.args...)
	return cmd.Output()
}

// computeSimpleDiff computes a simple line-by-line diff
func computeSimpleDiff(before, after string) string {
	// Simple approach: show what changed
	beforeLines := splitLines(before)
	afterLines := splitLines(after)

	var diff string
	beforeSet := make(map[string]bool)
	for _, line := range beforeLines {
		beforeSet[line] = true
	}

	afterSet := make(map[string]bool)
	for _, line := range afterLines {
		afterSet[line] = true
	}

	// Find removed lines
	for _, line := range beforeLines {
		if !afterSet[line] && line != "" {
			diff += "- " + line + "\n"
		}
	}

	// Find added lines
	for _, line := range afterLines {
		if !beforeSet[line] && line != "" {
			diff += "+ " + line + "\n"
		}
	}

	if diff == "" {
		return "(whitespace changes only)"
	}
	return diff
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
