package agent

import (
	"strings"
	"testing"
)

// TestClaudeParserImplementsInterface ensures ClaudeParser implements OutputParser
func TestClaudeParserImplementsInterface(t *testing.T) {
	var _ OutputParser = (*ClaudeParser)(nil)
}

// TestClaudeParserExtractsAssistantMessage tests extraction from assistant message events
func TestClaudeParserExtractsAssistantMessage(t *testing.T) {
	parser := &ClaudeParser{}

	// Claude stream-json format with assistant message
	input := `{"type":"system","subtype":"init","cwd":"/tmp/test","session_id":"abc123"}
{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Hello there, nice to meet!"}]},"session_id":"abc123"}
{"type":"result","subtype":"success","result":"Hello there, nice to meet!","session_id":"abc123"}`

	result := parser.Parse(input)

	if !strings.Contains(result, "Hello there, nice to meet!") {
		t.Errorf("Parse() did not extract assistant message, got: %s", result)
	}
}

// TestClaudeParserExtractsResultText tests extraction from result events
func TestClaudeParserExtractsResultText(t *testing.T) {
	parser := &ClaudeParser{}

	input := `{"type":"result","subtype":"success","is_error":false,"result":"The answer is 42.","session_id":"abc123"}`

	result := parser.Parse(input)

	if !strings.Contains(result, "The answer is 42.") {
		t.Errorf("Parse() did not extract result text, got: %s", result)
	}
}

// TestClaudeParserHandlesMultipleContentBlocks tests extraction of multiple text blocks
func TestClaudeParserHandlesMultipleContentBlocks(t *testing.T) {
	parser := &ClaudeParser{}

	input := `{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"First part."},{"type":"text","text":"Second part."}]},"session_id":"abc123"}`

	result := parser.Parse(input)

	if !strings.Contains(result, "First part.") {
		t.Errorf("Parse() did not extract first text block, got: %s", result)
	}
	if !strings.Contains(result, "Second part.") {
		t.Errorf("Parse() did not extract second text block, got: %s", result)
	}
}

// TestClaudeParserIgnoresToolUseEvents tests that tool use events are filtered
func TestClaudeParserIgnoresToolUseEvents(t *testing.T) {
	parser := &ClaudeParser{}

	input := `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool_123","name":"Bash","input":{"command":"ls"}}]},"session_id":"abc123"}
{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Here's the result."}]},"session_id":"abc123"}`

	result := parser.Parse(input)

	// Should not contain raw tool use JSON
	if strings.Contains(result, "tool_use") {
		t.Errorf("Parse() should filter tool_use events, got: %s", result)
	}
	if !strings.Contains(result, "Here's the result.") {
		t.Errorf("Parse() did not extract text after tool use, got: %s", result)
	}
}

// TestClaudeParserHandlesEmptyInput tests graceful handling of empty input
func TestClaudeParserHandlesEmptyInput(t *testing.T) {
	parser := &ClaudeParser{}

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

// TestClaudeParserPreservesOrder tests that output maintains chronological order
func TestClaudeParserPreservesOrder(t *testing.T) {
	parser := &ClaudeParser{}

	input := `{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"First"}]},"session_id":"abc"}
{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Second"}]},"session_id":"abc"}
{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Third"}]},"session_id":"abc"}`

	result := parser.Parse(input)

	firstIdx := strings.Index(result, "First")
	secondIdx := strings.Index(result, "Second")
	thirdIdx := strings.Index(result, "Third")

	if firstIdx == -1 || secondIdx == -1 || thirdIdx == -1 {
		t.Fatalf("Parse() missing content, got: %s", result)
	}

	if firstIdx >= secondIdx || secondIdx >= thirdIdx {
		t.Errorf("Parse() did not preserve order: First@%d, Second@%d, Third@%d", firstIdx, secondIdx, thirdIdx)
	}
}
