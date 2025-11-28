package agent

import (
	"strings"
	"testing"
)

// TestCursorParserImplementsInterface ensures CursorParser implements OutputParser
func TestCursorParserImplementsInterface(t *testing.T) {
	var _ OutputParser = (*CursorParser)(nil)
}

// TestCursorParserExtractsAssistantMessage tests extraction from assistant message events
func TestCursorParserExtractsAssistantMessage(t *testing.T) {
	parser := &CursorParser{}

	// Cursor stream-json format (Claude-compatible)
	input := `{"type":"system","subtype":"init","apiKeySource":"login","cwd":"/tmp/test","session_id":"abc123"}
{"type":"user","message":{"role":"user","content":[{"type":"text","text":"Hello"}]},"session_id":"abc123"}
{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Hello! How are you doing?"}]},"session_id":"abc123"}
{"type":"result","subtype":"success","result":"Hello! How are you doing?","session_id":"abc123"}`

	result := parser.Parse(input)

	if !strings.Contains(result, "Hello! How are you doing?") {
		t.Errorf("Parse() did not extract assistant message, got: %s", result)
	}
}

// TestCursorParserIgnoresThinkingDeltas tests that thinking delta events are filtered
func TestCursorParserIgnoresThinkingDeltas(t *testing.T) {
	parser := &CursorParser{}

	// Cursor outputs many empty thinking delta events
	input := `{"type":"thinking","subtype":"delta","text":"","session_id":"abc"}
{"type":"thinking","subtype":"delta","text":"","session_id":"abc"}
{"type":"thinking","subtype":"completed","session_id":"abc"}
{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Here's my answer."}]},"session_id":"abc"}`

	result := parser.Parse(input)

	// Should only contain the actual answer
	if !strings.Contains(result, "Here's my answer.") {
		t.Errorf("Parse() did not extract assistant message, got: %s", result)
	}
	// Should not have empty lines from thinking deltas
	if strings.Contains(result, "thinking") {
		t.Errorf("Parse() should filter thinking events, got: %s", result)
	}
}

// TestCursorParserExtractsResultText tests extraction from result events
func TestCursorParserExtractsResultText(t *testing.T) {
	parser := &CursorParser{}

	input := `{"type":"result","subtype":"success","duration_ms":4891,"is_error":false,"result":"The final answer.","session_id":"abc123"}`

	result := parser.Parse(input)

	if !strings.Contains(result, "The final answer.") {
		t.Errorf("Parse() did not extract result text, got: %s", result)
	}
}

// TestCursorParserHandlesEmptyInput tests graceful handling of empty input
func TestCursorParserHandlesEmptyInput(t *testing.T) {
	parser := &CursorParser{}

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
