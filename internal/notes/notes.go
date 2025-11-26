// Package notes provides functionality for saving agent perspectives to bead notes.
package notes

import (
	"context"
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
	s := &saver{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// SaveRoundResults saves all agent results from a round to a bead's notes.
func (s *saver) SaveRoundResults(ctx context.Context, beadID string, result orchestrator.RoundResult) error {
	// TODO: implement
	return nil
}

// FormatNote formats a single agent's response as a note entry.
func FormatNote(agentName, response string, timestamp time.Time) string {
	// TODO: implement
	return ""
}

// FormatRoundNotes formats all agent results from a round as notes.
func FormatRoundNotes(result orchestrator.RoundResult, timestamp time.Time) string {
	// TODO: implement
	return ""
}
