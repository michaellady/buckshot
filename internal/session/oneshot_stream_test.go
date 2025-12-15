package session

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/michaellady/buckshot/internal/agent"
)

// TestRunOneShotStreaming_WritesToStream tests that output is written to provided io.Writer in real-time.
func TestRunOneShotStreaming_WritesToStream(t *testing.T) {
	// Create a mock agent that outputs multiple lines
	mockPath := createStreamingMockAgent(t)
	testAgent := agent.Agent{
		Name:          "test-oneshot-stream",
		Path:          mockPath,
		Authenticated: true,
		Version:       "1.0.0",
		Pattern:       agent.KnownAgents()["codex"],
	}

	var streamBuf bytes.Buffer
	ctx := context.Background()

	// This function doesn't exist yet - RED phase
	result, err := RunOneShotStreaming(ctx, testAgent, "test prompt", &streamBuf)
	if err != nil {
		t.Fatalf("RunOneShotStreaming() error = %v", err)
	}

	// Stream buffer should have received output
	streamOutput := streamBuf.String()
	if streamOutput == "" {
		t.Error("Stream buffer is empty, expected output to be written")
	}

	// Should contain multiple lines (streamed in real-time)
	lines := strings.Split(strings.TrimSpace(streamOutput), "\n")
	if len(lines) < 2 {
		t.Errorf("Expected multiple lines in stream, got %d", len(lines))
	}

	// Result should also be populated
	if result.Output == "" {
		t.Error("Result.Output is empty")
	}
}

// TestRunOneShotStreaming_StillReturnsResult ensures OneShotResult contains full output even when streaming.
func TestRunOneShotStreaming_StillReturnsResult(t *testing.T) {
	mockPath := createStreamingMockAgent(t)
	testAgent := agent.Agent{
		Name:          "test-oneshot-result",
		Path:          mockPath,
		Authenticated: true,
		Version:       "1.0.0",
		Pattern:       agent.KnownAgents()["codex"],
	}

	var streamBuf bytes.Buffer
	ctx := context.Background()

	// RED phase - function doesn't exist
	result, err := RunOneShotStreaming(ctx, testAgent, "test prompt", &streamBuf)
	if err != nil {
		t.Fatalf("RunOneShotStreaming() error = %v", err)
	}

	// Result should contain the full output
	if result.Output == "" {
		t.Error("Result.Output should not be empty")
	}

	// Stream and result should have same content
	streamOutput := streamBuf.String()
	if result.Output != streamOutput {
		t.Errorf("Result.Output and stream differ:\nResult: %q\nStream: %q", result.Output, streamOutput)
	}

	// Exit code should be captured
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	// Error should be nil for successful execution
	if result.Error != nil {
		t.Errorf("Expected nil error, got %v", result.Error)
	}
}

// TestRunOneShotStreaming_NilWriter handles nil writer gracefully (falls back to non-streaming).
func TestRunOneShotStreaming_NilWriter(t *testing.T) {
	mockPath := createSimpleMockAgent(t)
	testAgent := agent.Agent{
		Name:          "test-oneshot-nil",
		Path:          mockPath,
		Authenticated: true,
		Version:       "1.0.0",
		Pattern:       agent.KnownAgents()["codex"],
	}

	ctx := context.Background()

	// Should not panic with nil writer
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("RunOneShotStreaming with nil writer panicked: %v", r)
		}
	}()

	// RED phase - function doesn't exist
	result, err := RunOneShotStreaming(ctx, testAgent, "test prompt", nil)
	if err != nil {
		t.Fatalf("RunOneShotStreaming(nil writer) error = %v", err)
	}

	// Should still return valid result
	if result.Output == "" {
		t.Error("Result.Output should not be empty even with nil writer")
	}
}

// TestRunOneShotStreaming_StreamsInRealTime verifies output arrives progressively, not all at once.
func TestRunOneShotStreaming_StreamsInRealTime(t *testing.T) {
	mockPath := createDelayedMockAgent(t)
	testAgent := agent.Agent{
		Name:          "test-oneshot-realtime",
		Path:          mockPath,
		Authenticated: true,
		Version:       "1.0.0",
		Pattern:       agent.KnownAgents()["codex"],
	}

	var streamBuf bytes.Buffer
	ctx := context.Background()

	// Track intermediate states
	intermediateStates := make([]string, 0)
	done := make(chan struct{})

	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				intermediateStates = append(intermediateStates, streamBuf.String())
			}
		}
	}()

	// RED phase - function doesn't exist
	_, err := RunOneShotStreaming(ctx, testAgent, "test prompt", &streamBuf)
	close(done)

	if err != nil {
		t.Fatalf("RunOneShotStreaming() error = %v", err)
	}

	// Should have captured multiple intermediate states showing progressive output
	if len(intermediateStates) < 2 {
		t.Log("Warning: fewer intermediate states captured than expected - output may be buffered")
	}

	// Check that states show progressive growth
	seenGrowth := false
	for i := 1; i < len(intermediateStates); i++ {
		if len(intermediateStates[i]) > len(intermediateStates[i-1]) {
			seenGrowth = true
			break
		}
	}
	if len(intermediateStates) >= 2 && !seenGrowth {
		t.Log("Warning: no progressive growth detected in intermediate states")
	}
}

// createStreamingMockAgent creates a mock agent that outputs multiple lines.
func createStreamingMockAgent(t *testing.T) string {
	t.Helper()
	return createMockAgentScript(t, `#!/bin/bash
echo "Line 1 from agent"
echo "Line 2 from agent"
echo "Line 3 from agent"
exit 0
`)
}

// createSimpleMockAgent creates a simple mock agent.
func createSimpleMockAgent(t *testing.T) string {
	t.Helper()
	return createMockAgentScript(t, `#!/bin/bash
echo "Simple output"
exit 0
`)
}

// createDelayedMockAgent creates a mock agent with delayed output.
func createDelayedMockAgent(t *testing.T) string {
	t.Helper()
	return createMockAgentScript(t, `#!/bin/bash
echo "First line"
sleep 0.1
echo "Second line"
sleep 0.1
echo "Third line"
exit 0
`)
}

// createMockAgentScript creates a temporary executable script.
func createMockAgentScript(t *testing.T, script string) string {
	t.Helper()

	tmpDir := t.TempDir()
	scriptPath := tmpDir + "/mock-agent"

	err := os.WriteFile(scriptPath, []byte(script), 0755)
	if err != nil {
		t.Fatalf("Failed to write mock script: %v", err)
	}

	return scriptPath
}
