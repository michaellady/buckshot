package presentation

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/michaellady/buckshot/internal/agent"
	"github.com/michaellady/buckshot/internal/dispatch"
	"github.com/michaellady/buckshot/internal/session"
)

func makeResult(name string, output string, err error, duration time.Duration) AgentResult {
	return AgentResult{
		Result: dispatch.Result{
			Agent: agent.Agent{
				Name:          name,
				Authenticated: true,
			},
			Response: session.Response{
				Output: output,
			},
			Error: err,
		},
		Duration: duration,
	}
}

// TestFormatTerminalMultipleAgents verifies terminal formatting with multiple agents.
func TestFormatTerminalMultipleAgents(t *testing.T) {
	results := []AgentResult{
		makeResult("claude", "This is Claude's response about the task.", nil, 2*time.Second),
		makeResult("codex", "Codex here with my analysis.", nil, 1500*time.Millisecond),
	}

	f := New()
	output := f.Format(results, FormatTerminal)

	// Should contain both agent names
	if !strings.Contains(output, "claude") {
		t.Error("Output should contain 'claude'")
	}
	if !strings.Contains(output, "codex") {
		t.Error("Output should contain 'codex'")
	}

	// Should contain both responses
	if !strings.Contains(output, "Claude's response") {
		t.Error("Output should contain Claude's response")
	}
	if !strings.Contains(output, "Codex here") {
		t.Error("Output should contain Codex's response")
	}

	// Should not be empty
	if output == "" {
		t.Error("Output should not be empty")
	}
}

// TestFormatTerminalShowsDuration verifies that duration is displayed.
func TestFormatTerminalShowsDuration(t *testing.T) {
	results := []AgentResult{
		makeResult("claude", "Response here.", nil, 2500*time.Millisecond),
	}

	f := New()
	output := f.Format(results, FormatTerminal)

	// Should show duration (2.5s or 2500ms or similar)
	if !strings.Contains(output, "2.5") && !strings.Contains(output, "2500") {
		t.Errorf("Output should contain duration, got: %s", output)
	}
}

// TestFormatTerminalClearlySeparated verifies agents are visually separated.
func TestFormatTerminalClearlySeparated(t *testing.T) {
	results := []AgentResult{
		makeResult("agent1", "First response.", nil, time.Second),
		makeResult("agent2", "Second response.", nil, time.Second),
	}

	f := New()
	output := f.Format(results, FormatTerminal)

	// Should have some visual separator (box chars, dashes, newlines, etc.)
	// At minimum, should have multiple lines
	lines := strings.Split(output, "\n")
	if len(lines) < 4 {
		t.Errorf("Output should have multiple lines for separation, got %d lines", len(lines))
	}

	// Responses should appear in order or be clearly labeled
	idx1 := strings.Index(output, "agent1")
	idx2 := strings.Index(output, "agent2")
	if idx1 == -1 || idx2 == -1 {
		t.Error("Both agent names should appear in output")
	}
}

// TestFormatTerminalTruncatesLongResponses verifies long responses are truncated.
func TestFormatTerminalTruncatesLongResponses(t *testing.T) {
	longResponse := strings.Repeat("This is a very long response. ", 100) // ~3000 chars
	results := []AgentResult{
		makeResult("claude", longResponse, nil, time.Second),
	}

	f := New()
	f.SetMaxResponseLength(500)
	output := f.Format(results, FormatTerminal)

	// Output should be truncated (allow overhead for box formatting and line wrapping)
	if len(output) > 1500 {
		t.Errorf("Output should be truncated, got length %d", len(output))
	}

	// Should indicate truncation
	if !strings.Contains(output, "...") && !strings.Contains(output, "truncated") {
		t.Error("Output should indicate truncation")
	}
}

// TestFormatTerminalShowsErrors verifies errors are displayed distinctly.
func TestFormatTerminalShowsErrors(t *testing.T) {
	testErr := errors.New("connection timeout")
	results := []AgentResult{
		makeResult("claude", "", testErr, time.Second),
		makeResult("codex", "Success response.", nil, time.Second),
	}

	f := New()
	output := f.Format(results, FormatTerminal)

	// Should indicate error for claude
	if !strings.Contains(output, "error") && !strings.Contains(output, "Error") && !strings.Contains(output, "ERROR") && !strings.Contains(output, "failed") && !strings.Contains(output, "Failed") {
		t.Error("Output should indicate error status")
	}

	// Should show the error message
	if !strings.Contains(output, "timeout") {
		t.Error("Output should contain error message")
	}

	// Should still show successful response
	if !strings.Contains(output, "Success response") {
		t.Error("Output should contain successful response")
	}
}

// TestFormatJSONStructured verifies JSON output is properly structured.
func TestFormatJSONStructured(t *testing.T) {
	results := []AgentResult{
		makeResult("claude", "Claude's analysis.", nil, 2*time.Second),
		makeResult("codex", "Codex's input.", nil, time.Second),
	}

	f := New()
	output := f.Format(results, FormatJSON)

	// Should be valid JSON
	var parsed interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Errorf("Output should be valid JSON: %v", err)
	}

	// Should contain agent data
	if !strings.Contains(output, "claude") {
		t.Error("JSON should contain 'claude'")
	}
	if !strings.Contains(output, "codex") {
		t.Error("JSON should contain 'codex'")
	}
}

// TestFormatJSONIncludesAllFields verifies JSON includes required fields.
func TestFormatJSONIncludesAllFields(t *testing.T) {
	testErr := errors.New("test error")
	results := []AgentResult{
		makeResult("claude", "Response text.", testErr, 1500*time.Millisecond),
	}

	f := New()
	output := f.Format(results, FormatJSON)

	// Parse JSON to check structure
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		// Try as array
		var arr []map[string]interface{}
		if err2 := json.Unmarshal([]byte(output), &arr); err2 != nil {
			t.Fatalf("Output should be valid JSON object or array: %v", err)
		}
		if len(arr) > 0 {
			data = arr[0]
		}
	}

	// Check for results array if wrapped
	if results, ok := data["results"].([]interface{}); ok {
		if len(results) > 0 {
			data = results[0].(map[string]interface{})
		}
	}

	// Should have agent name, response, error, duration fields
	requiredFields := []string{"agent", "response", "duration"}
	for _, field := range requiredFields {
		found := false
		for key := range data {
			if strings.Contains(strings.ToLower(key), strings.ToLower(field)) {
				found = true
				break
			}
		}
		// Check nested structures too
		if !found {
			if !strings.Contains(strings.ToLower(output), strings.ToLower(field)) {
				t.Errorf("JSON should contain field related to '%s'", field)
			}
		}
	}
}

// TestFormatEmptyResults verifies handling of empty results.
func TestFormatEmptyResults(t *testing.T) {
	f := New()

	// Terminal format
	termOutput := f.Format(nil, FormatTerminal)
	if termOutput == "" {
		// Empty is acceptable, but shouldn't panic
	}

	// JSON format should still be valid
	jsonOutput := f.Format([]AgentResult{}, FormatJSON)
	if jsonOutput != "" {
		var parsed interface{}
		if err := json.Unmarshal([]byte(jsonOutput), &parsed); err != nil {
			t.Errorf("Empty JSON output should still be valid JSON: %v", err)
		}
	}
}

// TestFormatMarkdown verifies markdown output format.
func TestFormatMarkdown(t *testing.T) {
	results := []AgentResult{
		makeResult("claude", "This is **important**.", nil, time.Second),
	}

	f := New()
	output := f.Format(results, FormatMarkdown)

	// Should contain markdown headers
	if !strings.Contains(output, "#") {
		t.Error("Markdown output should contain headers")
	}

	// Should contain agent name
	if !strings.Contains(output, "claude") {
		t.Error("Markdown should contain agent name")
	}
}

// TestFormatTerminalSummary verifies summary line is included.
func TestFormatTerminalSummary(t *testing.T) {
	testErr := errors.New("failed")
	results := []AgentResult{
		makeResult("agent1", "Success.", nil, time.Second),
		makeResult("agent2", "", testErr, time.Second),
		makeResult("agent3", "Also success.", nil, time.Second),
	}

	f := New()
	output := f.Format(results, FormatTerminal)

	// Should have summary with counts
	// "3 agents" or "2 succeeded" or "1 failed" etc.
	hasAgentCount := strings.Contains(output, "3") || strings.Contains(output, "agents")
	hasFailCount := strings.Contains(output, "1") && (strings.Contains(output, "fail") || strings.Contains(output, "error"))

	if !hasAgentCount && !hasFailCount {
		t.Error("Output should include summary with agent/failure counts")
	}
}

// TestFormatSingleAgent verifies single agent formatting.
func TestFormatSingleAgent(t *testing.T) {
	results := []AgentResult{
		makeResult("solo", "Only agent response.", nil, 500*time.Millisecond),
	}

	f := New()
	output := f.Format(results, FormatTerminal)

	if !strings.Contains(output, "solo") {
		t.Error("Output should contain agent name")
	}
	if !strings.Contains(output, "Only agent response") {
		t.Error("Output should contain response")
	}
}
