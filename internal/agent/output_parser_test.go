package agent

import (
	"testing"
)

// TestOutputParserInterface ensures OutputParser interface is properly defined
func TestOutputParserInterface(t *testing.T) {
	// This test verifies the interface exists and can be implemented
	var _ OutputParser = (*NoopParser)(nil)
}

// TestNoopParserReturnsInputUnchanged tests that NoopParser returns input unchanged
func TestNoopParserReturnsInputUnchanged(t *testing.T) {
	parser := &NoopParser{}

	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"simple text", "Hello, world!"},
		{"multiline text", "line1\nline2\nline3"},
		{"text with special chars", "foo\tbar\nbaz \"quoted\""},
		{"json-like output", `{"type":"message","content":"test"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)
			if result != tt.input {
				t.Errorf("NoopParser.Parse(%q) = %q, want %q", tt.input, result, tt.input)
			}
		})
	}
}

// mockParser is a test implementation of OutputParser
type mockParser struct {
	prefix string
}

func (m *mockParser) Parse(output string) string {
	return m.prefix + output
}

// TestMockImplementationWorks verifies that mock implementations satisfy the interface
func TestMockImplementationWorks(t *testing.T) {
	var parser OutputParser = &mockParser{prefix: "[parsed] "}

	input := "test output"
	result := parser.Parse(input)
	expected := "[parsed] test output"

	if result != expected {
		t.Errorf("mockParser.Parse(%q) = %q, want %q", input, result, expected)
	}
}

// TestAgentHasParserField verifies that Agent struct has a Parser field
func TestAgentHasParserField(t *testing.T) {
	agent := Agent{
		Name:   "test",
		Parser: &NoopParser{},
	}

	if agent.Parser == nil {
		t.Error("Agent.Parser is nil, want non-nil OutputParser")
	}

	// The Parser should be usable
	result := agent.Parser.Parse("test")
	if result != "test" {
		t.Errorf("Agent.Parser.Parse('test') = %q, want 'test'", result)
	}
}
