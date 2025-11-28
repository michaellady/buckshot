package agent

import (
	"strings"
	"testing"
)

// TestAmpParserImplementsInterface ensures AmpParser implements OutputParser
func TestAmpParserImplementsInterface(t *testing.T) {
	var _ OutputParser = (*AmpParser)(nil)
}

// TestAmpParserExtractsAssistantMessage tests extraction from assistant message events
func TestAmpParserExtractsAssistantMessage(t *testing.T) {
	parser := &AmpParser{}

	// Amp uses Claude Code-compatible stream-json format
	input := `{"type":"system","subtype":"init","cwd":"/tmp/test","session_id":"T-abc123","tools":["Bash","Read"]}
{"type":"user","message":{"role":"user","content":[{"type":"text","text":"Hello"}]},"session_id":"T-abc123"}
{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Hello! How can I help you today?"}]},"session_id":"T-abc123"}
{"type":"result","subtype":"success","result":"Hello! How can I help you today?","session_id":"T-abc123"}`

	result := parser.Parse(input)

	if !strings.Contains(result, "Hello! How can I help you today?") {
		t.Errorf("Parse() did not extract assistant message, got: %s", result)
	}
}

// TestAmpParserExtractsResultText tests extraction from result events
func TestAmpParserExtractsResultText(t *testing.T) {
	parser := &AmpParser{}

	input := `{"type":"result","subtype":"success","is_error":false,"result":"Task completed successfully.","session_id":"T-abc123"}`

	result := parser.Parse(input)

	if !strings.Contains(result, "Task completed successfully.") {
		t.Errorf("Parse() did not extract result text, got: %s", result)
	}
}

// TestAmpParserHandlesErrorResult tests handling of error results
func TestAmpParserHandlesErrorResult(t *testing.T) {
	parser := &AmpParser{}

	// Amp error format
	input := `{"type":"result","subtype":"error_during_execution","is_error":true,"error":"Insufficient credit balance.","session_id":"T-abc123"}`

	result := parser.Parse(input)

	// Should extract the error message
	if !strings.Contains(result, "Insufficient credit balance.") {
		t.Errorf("Parse() did not extract error message, got: %s", result)
	}
}

// TestAmpParserIgnoresSystemAndUserEvents tests that system/user events are filtered
func TestAmpParserIgnoresSystemAndUserEvents(t *testing.T) {
	parser := &AmpParser{}

	input := `{"type":"system","subtype":"init","cwd":"/tmp","tools":["Bash"]}
{"type":"user","message":{"role":"user","content":[{"type":"text","text":"User prompt"}]}}
{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Assistant response"}]}}`

	result := parser.Parse(input)

	if strings.Contains(result, "User prompt") {
		t.Errorf("Parse() should filter user events, got: %s", result)
	}
	if strings.Contains(result, "cwd") || strings.Contains(result, "tools") {
		t.Errorf("Parse() should filter system events, got: %s", result)
	}
	if !strings.Contains(result, "Assistant response") {
		t.Errorf("Parse() did not extract assistant message, got: %s", result)
	}
}

// TestAmpParserHandlesEmptyInput tests graceful handling of empty input
func TestAmpParserHandlesEmptyInput(t *testing.T) {
	parser := &AmpParser{}

	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"only whitespace", "   \n\t\n   "},
		{"only system event", `{"type":"system","subtype":"init","cwd":"/tmp"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			result := parser.Parse(tt.input)
			_ = result
		})
	}
}

// TestAmpParserHandlesMultipleContentBlocks tests extraction of multiple text blocks
func TestAmpParserHandlesMultipleContentBlocks(t *testing.T) {
	parser := &AmpParser{}

	input := `{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Part one."},{"type":"text","text":"Part two."}]}}`

	result := parser.Parse(input)

	if !strings.Contains(result, "Part one.") {
		t.Errorf("Parse() did not extract first text block, got: %s", result)
	}
	if !strings.Contains(result, "Part two.") {
		t.Errorf("Parse() did not extract second text block, got: %s", result)
	}
}
