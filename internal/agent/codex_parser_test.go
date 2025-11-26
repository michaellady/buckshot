package agent

import (
	"strings"
	"testing"
)

// TestCodexParserImplementsInterface ensures CodexParser implements OutputParser
func TestCodexParserImplementsInterface(t *testing.T) {
	var _ OutputParser = (*CodexParser)(nil)
}

// TestCodexParserExtractsReasoningText tests extraction of reasoning from item.text
func TestCodexParserExtractsReasoningText(t *testing.T) {
	parser := &CodexParser{}

	// Codex outputs JSONL with reasoning in item events
	input := `{"type":"item","item":{"type":"message","role":"assistant","content":[{"type":"text","text":"Let me analyze this problem."}]}}
{"type":"item","item":{"type":"message","role":"assistant","content":[{"type":"text","text":"The solution is to use a loop."}]}}`

	result := parser.Parse(input)

	if !strings.Contains(result, "Let me analyze this problem.") {
		t.Errorf("Parse() did not extract first reasoning text, got: %s", result)
	}
	if !strings.Contains(result, "The solution is to use a loop.") {
		t.Errorf("Parse() did not extract second reasoning text, got: %s", result)
	}
}

// TestCodexParserExtractsCommandOutput tests extraction of command execution results
func TestCodexParserExtractsCommandOutput(t *testing.T) {
	parser := &CodexParser{}

	// Command output appears in function_call_output events
	input := `{"type":"item","item":{"type":"function_call_output","output":"total 24\ndrwxr-xr-x  5 user  staff  160 Nov 26 09:00 .\ndrwxr-xr-x  3 user  staff   96 Nov 26 08:00 .."}}`

	result := parser.Parse(input)

	if !strings.Contains(result, "total 24") {
		t.Errorf("Parse() did not extract command output, got: %s", result)
	}
}

// TestCodexParserExtractsAggregatedOutput tests extraction from aggregated_output events
func TestCodexParserExtractsAggregatedOutput(t *testing.T) {
	parser := &CodexParser{}

	input := `{"type":"aggregated_output","output":"Build successful\nAll tests passed"}`

	result := parser.Parse(input)

	if !strings.Contains(result, "Build successful") {
		t.Errorf("Parse() did not extract aggregated_output, got: %s", result)
	}
	if !strings.Contains(result, "All tests passed") {
		t.Errorf("Parse() did not extract full aggregated_output, got: %s", result)
	}
}

// TestCodexParserHandlesMalformedJSON tests graceful handling of invalid JSON
func TestCodexParserHandlesMalformedJSON(t *testing.T) {
	parser := &CodexParser{}

	tests := []struct {
		name  string
		input string
	}{
		{"truncated json", `{"type":"item","item":`},
		{"not json at all", "This is plain text output"},
		{"partial line", `{"type":"item"` + "\n" + `{"type":"complete"}`},
		{"empty input", ""},
		{"only whitespace", "   \n\t\n   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			result := parser.Parse(tt.input)

			// Should return something (either extracted content or original)
			// The key is it doesn't crash
			_ = result
		})
	}
}

// TestCodexParserHandlesMixedContent tests parsing of mixed JSON and plain text
func TestCodexParserHandlesMixedContent(t *testing.T) {
	parser := &CodexParser{}

	// Real output often has mixed content
	input := `Starting execution...
{"type":"item","item":{"type":"message","role":"assistant","content":[{"type":"text","text":"I'll help with that."}]}}
Some status message
{"type":"aggregated_output","output":"Done!"}`

	result := parser.Parse(input)

	// Should extract the JSON content
	if !strings.Contains(result, "I'll help with that.") {
		t.Errorf("Parse() did not extract message from mixed content, got: %s", result)
	}
	if !strings.Contains(result, "Done!") {
		t.Errorf("Parse() did not extract aggregated_output from mixed content, got: %s", result)
	}
}

// TestCodexParserPreservesOrder tests that output maintains chronological order
func TestCodexParserPreservesOrder(t *testing.T) {
	parser := &CodexParser{}

	input := `{"type":"item","item":{"type":"message","role":"assistant","content":[{"type":"text","text":"First"}]}}
{"type":"item","item":{"type":"message","role":"assistant","content":[{"type":"text","text":"Second"}]}}
{"type":"item","item":{"type":"message","role":"assistant","content":[{"type":"text","text":"Third"}]}}`

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

// TestCodexParserHandlesThinkingBlocks tests extraction of thinking/reasoning blocks
func TestCodexParserHandlesThinkingBlocks(t *testing.T) {
	parser := &CodexParser{}

	// Some responses include thinking blocks
	input := `{"type":"item","item":{"type":"message","role":"assistant","content":[{"type":"thinking","thinking":"Let me think about this..."},{"type":"text","text":"Here's my answer."}]}}`

	result := parser.Parse(input)

	// Should include both thinking and text
	if !strings.Contains(result, "Here's my answer.") {
		t.Errorf("Parse() did not extract text content, got: %s", result)
	}
}

// TestCodexParserHandlesToolCalls tests extraction when tool calls are present
func TestCodexParserHandlesToolCalls(t *testing.T) {
	parser := &CodexParser{}

	input := `{"type":"item","item":{"type":"function_call","name":"shell","arguments":"{\"command\":\"ls -la\"}"}}
{"type":"item","item":{"type":"function_call_output","output":"file1.txt\nfile2.txt"}}`

	result := parser.Parse(input)

	// Should include command output
	if !strings.Contains(result, "file1.txt") {
		t.Errorf("Parse() did not extract tool output, got: %s", result)
	}
}
