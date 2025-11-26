package session

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/michaellady/buckshot/internal/agent"
)

// mockOutputParser is a test parser that transforms output
type mockOutputParser struct {
	prefix string
}

func (p *mockOutputParser) Parse(output string) string {
	return p.prefix + output
}

// setupMockCodexWithJSONOutput creates a mock codex that outputs JSON streaming format
func setupMockCodexWithJSONOutput(t *testing.T) string {
	t.Helper()

	// Mock codex that outputs JSON streaming format
	mockScript := `#!/bin/bash
# Mock codex for parser integration testing
# Output JSON streaming format like real codex

echo '{"type":"item","item":{"type":"message","role":"assistant","content":[{"type":"text","text":"I will help you with that."}]}}'
echo '{"type":"item","item":{"type":"function_call_output","output":"file1.txt\nfile2.txt"}}'
echo "Context: 5% used (10000/200000 tokens)"

# Read commands from stdin and respond
while IFS= read -r line; do
    if [[ -n "$line" ]]; then
        echo '{"type":"item","item":{"type":"message","role":"assistant","content":[{"type":"text","text":"Processing your request."}]}}'
        echo '{"type":"aggregated_output","output":"Command completed successfully"}'
        echo "Context: 15% used (30000/200000 tokens)"
    fi
done
`

	tmpDir := t.TempDir()
	mockPath := filepath.Join(tmpDir, "mock-codex")

	if err := os.WriteFile(mockPath, []byte(mockScript), 0755); err != nil {
		t.Fatalf("Failed to create mock codex: %v", err)
	}

	return mockPath
}

// TestSessionSendUsesAgentParser tests that Send() uses the agent's Parser
func TestSessionSendUsesAgentParser(t *testing.T) {
	// Create a mock agent with a parser that adds a prefix
	mockPath := setupMockCodexWithJSONOutput(t)
	testAgent := agent.Agent{
		Name:          "codex",
		Path:          mockPath,
		Authenticated: true,
		Version:       "1.0.0",
		Pattern:       agent.KnownAgents()["codex"],
		Parser:        &mockOutputParser{prefix: "[PARSED] "},
	}

	mgr := NewManager()
	sess, err := mgr.CreateSession(testAgent)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	defer sess.Close()

	ctx := context.Background()
	agentsPath := newTestAgentsFile(t)
	if err := sess.Start(ctx, agentsPath); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	resp, err := sess.Send(ctx, "test prompt")
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	// The output should have been transformed by the parser
	if !strings.HasPrefix(resp.Output, "[PARSED] ") {
		t.Errorf("Send() output was not parsed, got: %q, want prefix '[PARSED] '", resp.Output)
	}
}

// TestSessionSendWithCodexParserExtractsText tests that CodexParser extracts clean text
func TestSessionSendWithCodexParserExtractsText(t *testing.T) {
	mockPath := setupMockCodexWithJSONOutput(t)
	testAgent := agent.Agent{
		Name:          "codex",
		Path:          mockPath,
		Authenticated: true,
		Version:       "1.0.0",
		Pattern:       agent.KnownAgents()["codex"],
		Parser:        &agent.CodexParser{},
	}

	mgr := NewManager()
	sess, err := mgr.CreateSession(testAgent)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	defer sess.Close()

	ctx := context.Background()
	agentsPath := newTestAgentsFile(t)
	if err := sess.Start(ctx, agentsPath); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	resp, err := sess.Send(ctx, "test prompt")
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	// The output should contain extracted text, not raw JSON
	if strings.Contains(resp.Output, `{"type":"item"`) {
		t.Errorf("Send() output contains raw JSON, want parsed text. Got: %q", resp.Output)
	}

	// Should contain extracted text content (from either initial output or response)
	hasExpectedContent := strings.Contains(resp.Output, "Processing your request") ||
		strings.Contains(resp.Output, "I will help you with that") ||
		strings.Contains(resp.Output, "Command completed successfully")
	if !hasExpectedContent {
		t.Errorf("Send() output missing extracted text content, got: %q", resp.Output)
	}
}

// TestSessionSendWithoutParserReturnsRawOutput tests that nil parser returns raw output
func TestSessionSendWithoutParserReturnsRawOutput(t *testing.T) {
	mockPath := setupMockCodexWithJSONOutput(t)
	testAgent := agent.Agent{
		Name:          "codex",
		Path:          mockPath,
		Authenticated: true,
		Version:       "1.0.0",
		Pattern:       agent.KnownAgents()["codex"],
		Parser:        nil, // No parser
	}

	mgr := NewManager()
	sess, err := mgr.CreateSession(testAgent)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	defer sess.Close()

	ctx := context.Background()
	agentsPath := newTestAgentsFile(t)
	if err := sess.Start(ctx, agentsPath); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	resp, err := sess.Send(ctx, "test prompt")
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	// Without a parser, the raw JSON output should be preserved
	// (This test passes currently since parser integration isn't implemented yet)
	if resp.Output == "" {
		t.Error("Send() output is empty, want raw output")
	}
}

// TestSessionSendWithNoopParserReturnsUnchanged tests NoopParser returns input unchanged
func TestSessionSendWithNoopParserReturnsUnchanged(t *testing.T) {
	mockPath := setupMockCodexWithJSONOutput(t)
	testAgent := agent.Agent{
		Name:          "codex",
		Path:          mockPath,
		Authenticated: true,
		Version:       "1.0.0",
		Pattern:       agent.KnownAgents()["codex"],
		Parser:        &agent.NoopParser{},
	}

	mgr := NewManager()
	sess, err := mgr.CreateSession(testAgent)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	defer sess.Close()

	ctx := context.Background()
	agentsPath := newTestAgentsFile(t)
	if err := sess.Start(ctx, agentsPath); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	resp, err := sess.Send(ctx, "test prompt")
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	// NoopParser should return the output unchanged (including raw JSON)
	if resp.Output == "" {
		t.Error("Send() output is empty, want raw output unchanged")
	}

	// With NoopParser, the raw format should be preserved
	// (either contains JSON or the raw text depending on mock)
}

// TestParsedOutputAppearsInResponse tests that parsed output is in Response.Output
func TestParsedOutputAppearsInResponse(t *testing.T) {
	mockPath := setupMockCodexWithJSONOutput(t)

	// Create a parser that transforms the output in a verifiable way
	transformingParser := &mockOutputParser{prefix: "TRANSFORMED:"}

	testAgent := agent.Agent{
		Name:          "codex",
		Path:          mockPath,
		Authenticated: true,
		Version:       "1.0.0",
		Pattern:       agent.KnownAgents()["codex"],
		Parser:        transformingParser,
	}

	mgr := NewManager()
	sess, err := mgr.CreateSession(testAgent)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	defer sess.Close()

	ctx := context.Background()
	agentsPath := newTestAgentsFile(t)
	if err := sess.Start(ctx, agentsPath); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	resp, err := sess.Send(ctx, "test")
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	// The Response.Output field should contain the transformed output
	if !strings.HasPrefix(resp.Output, "TRANSFORMED:") {
		t.Errorf("Response.Output does not contain transformed output, got: %q", resp.Output)
	}
}
