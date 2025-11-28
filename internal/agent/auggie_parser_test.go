package agent

import (
	"strings"
	"testing"
)

// TestAuggieParserImplementsInterface ensures AuggieParser implements OutputParser
func TestAuggieParserImplementsInterface(t *testing.T) {
	var _ OutputParser = (*AuggieParser)(nil)
}

// TestAuggieParserExtractsResultText tests extraction from result events
func TestAuggieParserExtractsResultText(t *testing.T) {
	parser := &AuggieParser{}

	// Auggie JSON format - simple result object
	input := `{"type":"result","result":"\nHello there, nice to meet you!\n","is_error":false,"subtype":"success","session_id":"abc123","num_turns":0}`

	result := parser.Parse(input)

	if !strings.Contains(result, "Hello there, nice to meet you!") {
		t.Errorf("Parse() did not extract result text, got: %s", result)
	}
}

// TestAuggieParserTrimsWhitespace tests that leading/trailing whitespace is trimmed
func TestAuggieParserTrimsWhitespace(t *testing.T) {
	parser := &AuggieParser{}

	input := `{"type":"result","result":"\n\n  The answer is here.  \n\n","is_error":false}`

	result := parser.Parse(input)

	// Should trim the excessive whitespace
	if strings.HasPrefix(result, "\n\n") || strings.HasSuffix(result, "\n\n") {
		t.Errorf("Parse() did not trim whitespace properly, got: %q", result)
	}
	if !strings.Contains(result, "The answer is here.") {
		t.Errorf("Parse() did not preserve content, got: %s", result)
	}
}

// TestAuggieParserHandlesErrorResult tests handling of error results
func TestAuggieParserHandlesErrorResult(t *testing.T) {
	parser := &AuggieParser{}

	input := `{"type":"result","result":"","is_error":true,"subtype":"error","error":"Something went wrong"}`

	result := parser.Parse(input)

	// Should extract the error message
	if !strings.Contains(result, "Something went wrong") {
		t.Errorf("Parse() did not extract error message, got: %s", result)
	}
}

// TestAuggieParserHandlesEmptyInput tests graceful handling of empty input
func TestAuggieParserHandlesEmptyInput(t *testing.T) {
	parser := &AuggieParser{}

	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"only whitespace", "   \n\t\n   "},
		{"empty result", `{"type":"result","result":"","is_error":false}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			result := parser.Parse(tt.input)
			_ = result
		})
	}
}

// TestAuggieParserHandlesMalformedJSON tests graceful handling of invalid JSON
func TestAuggieParserHandlesMalformedJSON(t *testing.T) {
	parser := &AuggieParser{}

	tests := []struct {
		name  string
		input string
	}{
		{"truncated json", `{"type":"result","result":`},
		{"not json at all", "This is plain text output"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic, should return original input
			result := parser.Parse(tt.input)
			if result != tt.input {
				t.Errorf("Parse() should return original on malformed JSON, got: %s", result)
			}
		})
	}
}
