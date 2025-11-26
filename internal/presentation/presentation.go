// Package presentation formats dispatch results for display.
package presentation

import (
	"time"

	"github.com/michaellady/buckshot/internal/dispatch"
)

// OutputFormat specifies the output format for results.
type OutputFormat int

const (
	// FormatTerminal uses colored, bordered sections for terminal display.
	FormatTerminal OutputFormat = iota
	// FormatJSON outputs structured JSON for piping.
	FormatJSON
	// FormatMarkdown outputs markdown for saving.
	FormatMarkdown
)

// AgentResult extends dispatch.Result with presentation metadata.
type AgentResult struct {
	dispatch.Result
	Duration time.Duration // How long the agent took to respond
}

// Formatter handles formatting of dispatch results.
type Formatter interface {
	// Format formats results in the specified output format.
	Format(results []AgentResult, format OutputFormat) string

	// SetMaxResponseLength sets the maximum response length before truncation.
	SetMaxResponseLength(length int)
}

// formatter is the default implementation.
type formatter struct {
	maxResponseLength int
}

// New creates a new Formatter.
func New() Formatter {
	return &formatter{
		maxResponseLength: 1000, // Default max length
	}
}

// Format formats results in the specified output format.
func (f *formatter) Format(results []AgentResult, format OutputFormat) string {
	// TODO: Implement formatting
	// This is intentionally unimplemented to make tests fail (TDD RED phase)
	return ""
}

// SetMaxResponseLength sets the maximum response length before truncation.
func (f *formatter) SetMaxResponseLength(length int) {
	f.maxResponseLength = length
}
