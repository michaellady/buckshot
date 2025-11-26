// Package notes provides functionality for saving agent perspectives to bead notes.
package notes

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/michaellady/buckshot/internal/orchestrator"
)

// Executor runs shell commands.
type Executor interface {
	Execute(ctx context.Context, name string, args ...string) (string, error)
}

// Saver saves agent perspectives to bead notes.
type Saver interface {
	// SaveRoundResults saves all agent results from a round to a bead's notes.
	SaveRoundResults(ctx context.Context, beadID string, result orchestrator.RoundResult) error
}

// Option configures a Saver.
type Option func(*saver)

// WithExecutor sets a custom executor for running bd commands.
func WithExecutor(exec Executor) Option {
	return func(s *saver) {
		s.executor = exec
	}
}

// saver is the default implementation.
type saver struct {
	executor Executor
}

// NewSaver creates a new Saver.
func NewSaver(opts ...Option) Saver {
	s := &saver{
		executor: &defaultExecutor{},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// SaveRoundResults saves all agent results from a round to a bead's notes.
func (s *saver) SaveRoundResults(ctx context.Context, beadID string, result orchestrator.RoundResult) error {
	// Skip if no agent results
	if len(result.AgentResults) == 0 {
		return nil
	}

	// Format all results as notes
	notes := FormatRoundNotes(result, time.Now())

	// Execute bd update --notes
	_, err := s.executor.Execute(ctx, "bd", "update", beadID, "--notes", notes)
	if err != nil {
		return fmt.Errorf("failed to save notes to bead %s: %w", beadID, err)
	}

	return nil
}

// FormatNote formats a single agent's response as a note entry.
func FormatNote(agentName, response string, timestamp time.Time) string {
	timeStr := timestamp.Format("2006-01-02 15:04:05")
	header := fmt.Sprintf("### %s @ %s", agentName, timeStr)

	if response == "" {
		return header + "\n(no response)"
	}

	return header + "\n" + response
}

// FormatRoundNotes formats all agent results from a round as notes.
func FormatRoundNotes(result orchestrator.RoundResult, timestamp time.Time) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## Round %d\n\n", result.Round))

	for i, agentResult := range result.AgentResults {
		if i > 0 {
			sb.WriteString("\n---\n\n")
		}

		response := agentResult.Response.Output
		if agentResult.Error != nil {
			response = fmt.Sprintf("[ERROR: %s]", agentResult.Error.Error())
		}

		note := FormatNote(agentResult.Agent.Name, response, timestamp)
		sb.WriteString(note)
		sb.WriteString("\n")
	}

	return sb.String()
}

// defaultExecutor executes commands using os/exec.
type defaultExecutor struct{}

func (e *defaultExecutor) Execute(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
