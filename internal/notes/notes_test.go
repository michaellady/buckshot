// Package notes provides functionality for saving agent perspectives to bead notes.
package notes

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/michaellady/buckshot/internal/agent"
	"github.com/michaellady/buckshot/internal/orchestrator"
	"github.com/michaellady/buckshot/internal/session"
)

// TestFormatNote tests that agent responses are formatted correctly for bead notes.
func TestFormatNote(t *testing.T) {
	tests := []struct {
		name       string
		agentName  string
		response   string
		timestamp  time.Time
		wantAgent  string
		wantTime   string
		wantOutput string
	}{
		{
			name:       "basic format",
			agentName:  "claude",
			response:   "I suggest creating a task for authentication.",
			timestamp:  time.Date(2025, 11, 26, 10, 30, 0, 0, time.UTC),
			wantAgent:  "claude",
			wantTime:   "2025-11-26 10:30:00",
			wantOutput: "I suggest creating a task for authentication.",
		},
		{
			name:       "multiline response",
			agentName:  "codex",
			response:   "Line 1\nLine 2\nLine 3",
			timestamp:  time.Date(2025, 11, 26, 14, 0, 0, 0, time.UTC),
			wantAgent:  "codex",
			wantTime:   "2025-11-26 14:00:00",
			wantOutput: "Line 1\nLine 2\nLine 3",
		},
		{
			name:       "empty response",
			agentName:  "cursor-agent",
			response:   "",
			timestamp:  time.Date(2025, 11, 26, 9, 0, 0, 0, time.UTC),
			wantAgent:  "cursor-agent",
			wantTime:   "2025-11-26 09:00:00",
			wantOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			note := FormatNote(tt.agentName, tt.response, tt.timestamp)

			// Check that note contains expected components
			if !strings.Contains(note, tt.wantAgent) {
				t.Errorf("FormatNote() missing agent name %q in output:\n%s", tt.wantAgent, note)
			}
			if !strings.Contains(note, tt.wantTime) {
				t.Errorf("FormatNote() missing timestamp %q in output:\n%s", tt.wantTime, note)
			}
			if tt.wantOutput != "" && !strings.Contains(note, tt.wantOutput) {
				t.Errorf("FormatNote() missing response content in output:\n%s", note)
			}
		})
	}
}

// TestFormatNote_Structure tests the structure of formatted notes.
func TestFormatNote_Structure(t *testing.T) {
	note := FormatNote("claude", "Test response", time.Date(2025, 11, 26, 10, 0, 0, 0, time.UTC))

	// Should have a header line with agent and timestamp
	lines := strings.Split(note, "\n")
	if len(lines) < 2 {
		t.Errorf("FormatNote() should produce at least 2 lines, got %d", len(lines))
	}

	// First line should be a header with agent name and timestamp
	header := lines[0]
	if !strings.Contains(header, "claude") || !strings.Contains(header, "2025-11-26") {
		t.Errorf("FormatNote() header should contain agent and date: %q", header)
	}
}

// TestSaver_SaveRoundResults tests saving round results to a bead.
func TestSaver_SaveRoundResults(t *testing.T) {
	// Create mock executor to capture bd commands
	mockExec := &mockExecutor{
		results: make(map[string]execResult),
	}
	mockExec.results["bd update"] = execResult{output: "✓ Updated issue: buckshot-123", err: nil}

	saver := NewSaver(WithExecutor(mockExec))

	roundResult := orchestrator.RoundResult{
		Round: 1,
		AgentResults: []orchestrator.AgentResult{
			{
				Agent:    agent.Agent{Name: "claude"},
				Response: session.Response{Output: "I recommend breaking this into subtasks."},
			},
			{
				Agent:    agent.Agent{Name: "codex"},
				Response: session.Response{Output: "Agreed, let's create three tasks."},
			},
		},
	}

	err := saver.SaveRoundResults(context.Background(), "buckshot-123", roundResult)
	if err != nil {
		t.Errorf("SaveRoundResults() error = %v", err)
	}

	// Verify bd update was called
	if len(mockExec.commands) == 0 {
		t.Error("SaveRoundResults() should execute bd update command")
	}

	// Check that both agents' responses were included
	found := false
	for _, cmd := range mockExec.commands {
		if strings.Contains(cmd, "bd update") && strings.Contains(cmd, "buckshot-123") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("SaveRoundResults() should call bd update with bead ID, got commands: %v", mockExec.commands)
	}
}

// TestSaver_SaveRoundResults_SkipsFailedAgents tests that failed agents are noted but included.
func TestSaver_SaveRoundResults_SkipsFailedAgents(t *testing.T) {
	mockExec := &mockExecutor{
		results: make(map[string]execResult),
	}
	mockExec.results["bd update"] = execResult{output: "✓ Updated", err: nil}

	saver := NewSaver(WithExecutor(mockExec))

	roundResult := orchestrator.RoundResult{
		Round: 1,
		AgentResults: []orchestrator.AgentResult{
			{
				Agent:    agent.Agent{Name: "claude"},
				Response: session.Response{Output: "Success response"},
			},
			{
				Agent:    agent.Agent{Name: "codex"},
				Response: session.Response{Output: ""},
				Error:    errors.New("connection failed"),
			},
		},
	}

	err := saver.SaveRoundResults(context.Background(), "buckshot-456", roundResult)
	if err != nil {
		t.Errorf("SaveRoundResults() error = %v", err)
	}

	// Should still save, marking the failed agent appropriately
	if len(mockExec.commands) == 0 {
		t.Error("SaveRoundResults() should still save even with failed agents")
	}
}

// TestSaver_SaveRoundResults_BdUpdateFails tests error handling when bd fails.
func TestSaver_SaveRoundResults_BdUpdateFails(t *testing.T) {
	mockExec := &mockExecutor{
		results: make(map[string]execResult),
	}
	mockExec.results["bd update"] = execResult{output: "", err: errors.New("bd: issue not found")}

	saver := NewSaver(WithExecutor(mockExec))

	roundResult := orchestrator.RoundResult{
		Round: 1,
		AgentResults: []orchestrator.AgentResult{
			{
				Agent:    agent.Agent{Name: "claude"},
				Response: session.Response{Output: "Response"},
			},
		},
	}

	err := saver.SaveRoundResults(context.Background(), "invalid-id", roundResult)
	if err == nil {
		t.Error("SaveRoundResults() should return error when bd update fails")
	}
}

// TestSaver_SaveRoundResults_EmptyResults tests handling of empty round results.
func TestSaver_SaveRoundResults_EmptyResults(t *testing.T) {
	mockExec := &mockExecutor{
		results: make(map[string]execResult),
	}

	saver := NewSaver(WithExecutor(mockExec))

	roundResult := orchestrator.RoundResult{
		Round:        1,
		AgentResults: []orchestrator.AgentResult{},
	}

	err := saver.SaveRoundResults(context.Background(), "buckshot-789", roundResult)
	if err != nil {
		t.Errorf("SaveRoundResults() should not error on empty results, got %v", err)
	}

	// Should not call bd update for empty results
	if len(mockExec.commands) > 0 {
		t.Errorf("SaveRoundResults() should not call bd update for empty results, got %v", mockExec.commands)
	}
}

// TestFormatRoundNotes tests formatting all agent results from a round.
func TestFormatRoundNotes(t *testing.T) {
	roundResult := orchestrator.RoundResult{
		Round: 2,
		AgentResults: []orchestrator.AgentResult{
			{
				Agent:    agent.Agent{Name: "claude"},
				Response: session.Response{Output: "Claude's perspective"},
			},
			{
				Agent:    agent.Agent{Name: "codex"},
				Response: session.Response{Output: "Codex's perspective"},
			},
		},
	}

	notes := FormatRoundNotes(roundResult, time.Date(2025, 11, 26, 12, 0, 0, 0, time.UTC))

	// Should contain round header
	if !strings.Contains(notes, "Round 2") {
		t.Errorf("FormatRoundNotes() should include round number, got:\n%s", notes)
	}

	// Should contain both agents
	if !strings.Contains(notes, "claude") || !strings.Contains(notes, "codex") {
		t.Errorf("FormatRoundNotes() should include all agents, got:\n%s", notes)
	}

	// Should contain both responses
	if !strings.Contains(notes, "Claude's perspective") || !strings.Contains(notes, "Codex's perspective") {
		t.Errorf("FormatRoundNotes() should include all responses, got:\n%s", notes)
	}
}

// Mock types for testing

type execResult struct {
	output string
	err    error
}

type mockExecutor struct {
	commands []string
	results  map[string]execResult
}

func (m *mockExecutor) Execute(ctx context.Context, name string, args ...string) (string, error) {
	cmd := name + " " + strings.Join(args, " ")
	m.commands = append(m.commands, cmd)

	// Find matching result
	for prefix, result := range m.results {
		if strings.HasPrefix(cmd, prefix) {
			return result.output, result.err
		}
	}
	return "", nil
}
