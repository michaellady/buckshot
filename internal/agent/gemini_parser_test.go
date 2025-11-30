package agent

import (
	"strings"
	"testing"
)

// TestGeminiParserImplementsInterface ensures GeminiParser implements OutputParser
func TestGeminiParserImplementsInterface(t *testing.T) {
	var _ OutputParser = (*GeminiParser)(nil)
}

// TestGeminiParserExtractsAssistantDeltas tests extraction from delta message events
func TestGeminiParserExtractsAssistantDeltas(t *testing.T) {
	parser := &GeminiParser{}

	// Gemini stream-json format with delta messages
	input := `{"type":"init","timestamp":"2025-11-28T16:00:05.332Z","session_id":"abc123","model":"auto"}
{"type":"message","timestamp":"2025-11-28T16:00:05.333Z","role":"user","content":"Say hello"}
{"type":"message","timestamp":"2025-11-28T16:00:08.466Z","role":"assistant","content":"Hello!","delta":true}
{"type":"message","timestamp":"2025-11-28T16:00:08.466Z","role":"assistant","content":" How can I help?","delta":true}
{"type":"result","timestamp":"2025-11-28T16:00:08.478Z","status":"success"}`

	result := parser.Parse(input)

	if !strings.Contains(result, "Hello!") {
		t.Errorf("Parse() did not extract first delta, got: %s", result)
	}
	if !strings.Contains(result, "How can I help?") {
		t.Errorf("Parse() did not extract second delta, got: %s", result)
	}
}

// TestGeminiParserConcatenatesDeltas tests that delta messages are properly joined
func TestGeminiParserConcatenatesDeltas(t *testing.T) {
	parser := &GeminiParser{}

	input := `{"type":"message","role":"assistant","content":"Hello","delta":true}
{"type":"message","role":"assistant","content":" world","delta":true}
{"type":"message","role":"assistant","content":"!","delta":true}`

	result := parser.Parse(input)

	// Deltas should be concatenated without extra newlines
	if !strings.Contains(result, "Hello world!") {
		t.Errorf("Parse() did not concatenate deltas properly, got: %s", result)
	}
}

// TestGeminiParserIgnoresUserMessages tests that user messages are filtered
func TestGeminiParserIgnoresUserMessages(t *testing.T) {
	parser := &GeminiParser{}

	input := `{"type":"message","role":"user","content":"What is 2+2?"}
{"type":"message","role":"assistant","content":"The answer is 4.","delta":true}`

	result := parser.Parse(input)

	if strings.Contains(result, "What is 2+2?") {
		t.Errorf("Parse() should filter user messages, got: %s", result)
	}
	if !strings.Contains(result, "The answer is 4.") {
		t.Errorf("Parse() did not extract assistant message, got: %s", result)
	}
}

// TestGeminiParserIgnoresInitAndResult tests that init/result events don't pollute output
func TestGeminiParserIgnoresInitAndResult(t *testing.T) {
	parser := &GeminiParser{}

	input := `{"type":"init","session_id":"abc123","model":"auto"}
{"type":"message","role":"assistant","content":"My response.","delta":true}
{"type":"result","status":"success","stats":{"total_tokens":100}}`

	result := parser.Parse(input)

	if strings.Contains(result, "init") || strings.Contains(result, "session_id") {
		t.Errorf("Parse() should filter init events, got: %s", result)
	}
	if strings.Contains(result, "total_tokens") || strings.Contains(result, "stats") {
		t.Errorf("Parse() should filter result stats, got: %s", result)
	}
	if !strings.Contains(result, "My response.") {
		t.Errorf("Parse() did not extract assistant message, got: %s", result)
	}
}

// TestGeminiParserHandlesEmptyInput tests graceful handling of empty input
func TestGeminiParserHandlesEmptyInput(t *testing.T) {
	parser := &GeminiParser{}

	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"only whitespace", "   \n\t\n   "},
		{"only init event", `{"type":"init","model":"auto"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			result := parser.Parse(tt.input)
			_ = result
		})
	}
}

// TestGeminiParserHandlesNonDeltaAssistant tests handling of non-delta assistant messages
func TestGeminiParserHandlesNonDeltaAssistant(t *testing.T) {
	parser := &GeminiParser{}

	// Some responses might not have delta:true
	input := `{"type":"message","role":"assistant","content":"Complete response here."}`

	result := parser.Parse(input)

	if !strings.Contains(result, "Complete response here.") {
		t.Errorf("Parse() did not extract non-delta assistant message, got: %s", result)
	}
}
