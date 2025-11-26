// Package presentation formats dispatch results for display.
package presentation

import (
	"encoding/json"
	"fmt"
	"strings"
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
	if len(results) == 0 {
		switch format {
		case FormatJSON:
			return "[]"
		default:
			return ""
		}
	}

	switch format {
	case FormatJSON:
		return f.formatJSON(results)
	case FormatMarkdown:
		return f.formatMarkdown(results)
	default:
		return f.formatTerminal(results)
	}
}

// SetMaxResponseLength sets the maximum response length before truncation.
func (f *formatter) SetMaxResponseLength(length int) {
	f.maxResponseLength = length
}

// formatTerminal formats results for terminal display with box-drawing characters.
func (f *formatter) formatTerminal(results []AgentResult) string {
	var sb strings.Builder

	successCount := 0
	failCount := 0

	for i, r := range results {
		if i > 0 {
			sb.WriteString("\n")
		}

		// Box top
		sb.WriteString("┌──────────────────────────────────────────────────────────────────────────────┐\n")

		// Agent name and duration
		duration := formatDuration(r.Duration)
		if r.Error != nil {
			sb.WriteString(fmt.Sprintf("│ %-40s %33s │\n", r.Agent.Name+" [ERROR]", duration))
			failCount++
		} else {
			sb.WriteString(fmt.Sprintf("│ %-40s %33s │\n", r.Agent.Name, duration))
			successCount++
		}

		// Separator
		sb.WriteString("├──────────────────────────────────────────────────────────────────────────────┤\n")

		// Content (response or error)
		if r.Error != nil {
			sb.WriteString(fmt.Sprintf("│ Error: %-68s │\n", r.Error.Error()))
		} else {
			response := r.Response.Output
			if f.maxResponseLength > 0 && len(response) > f.maxResponseLength {
				response = response[:f.maxResponseLength] + "... [truncated]"
			}

			// Wrap response in box
			lines := wrapText(response, 76)
			for _, line := range lines {
				sb.WriteString(fmt.Sprintf("│ %-76s │\n", line))
			}
		}

		// Box bottom
		sb.WriteString("└──────────────────────────────────────────────────────────────────────────────┘\n")
	}

	// Summary
	sb.WriteString(fmt.Sprintf("\nSummary: %d agents, %d succeeded, %d failed\n", len(results), successCount, failCount))

	return sb.String()
}

// formatJSON formats results as structured JSON.
func (f *formatter) formatJSON(results []AgentResult) string {
	type jsonResult struct {
		Agent    string  `json:"agent"`
		Response string  `json:"response"`
		Error    string  `json:"error,omitempty"`
		Duration string  `json:"duration"`
		DurationMs int64 `json:"duration_ms"`
	}

	jsonResults := make([]jsonResult, len(results))
	for i, r := range results {
		jr := jsonResult{
			Agent:      r.Agent.Name,
			Response:   r.Response.Output,
			Duration:   formatDuration(r.Duration),
			DurationMs: r.Duration.Milliseconds(),
		}
		if r.Error != nil {
			jr.Error = r.Error.Error()
		}
		jsonResults[i] = jr
	}

	data, err := json.MarshalIndent(jsonResults, "", "  ")
	if err != nil {
		return "[]"
	}
	return string(data)
}

// formatMarkdown formats results as markdown.
func (f *formatter) formatMarkdown(results []AgentResult) string {
	var sb strings.Builder

	sb.WriteString("# Agent Responses\n\n")

	for _, r := range results {
		sb.WriteString(fmt.Sprintf("## %s\n\n", r.Agent.Name))
		sb.WriteString(fmt.Sprintf("**Duration:** %s\n\n", formatDuration(r.Duration)))

		if r.Error != nil {
			sb.WriteString(fmt.Sprintf("**Error:** %s\n\n", r.Error.Error()))
		} else {
			sb.WriteString(r.Response.Output)
			sb.WriteString("\n\n")
		}

		sb.WriteString("---\n\n")
	}

	return sb.String()
}

// formatDuration formats a duration for display.
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

// wrapText wraps text to fit within a given width.
func wrapText(text string, width int) []string {
	if text == "" {
		return []string{""}
	}

	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		if paragraph == "" {
			lines = append(lines, "")
			continue
		}

		words := strings.Fields(paragraph)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}

		currentLine := words[0]
		for _, word := range words[1:] {
			if len(currentLine)+1+len(word) <= width {
				currentLine += " " + word
			} else {
				lines = append(lines, currentLine)
				currentLine = word
			}
		}
		lines = append(lines, currentLine)
	}

	return lines
}
