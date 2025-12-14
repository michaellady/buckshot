package session

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/michaellady/buckshot/internal/agent"
)

// TestSession_SetOutputStream tests that Session accepts an io.Writer for streaming output.
func TestSession_SetOutputStream(t *testing.T) {
	mgr := NewManager()
	sess, err := mgr.CreateSession(newTestAgentWithMock(t))
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	defer func() { _ = sess.Close() }()

	// Session should accept a stream writer
	var buf bytes.Buffer

	// This will fail until SetOutputStream is added to Session interface
	streamSetter, ok := sess.(interface{ SetOutputStream(*bytes.Buffer) })
	if !ok {
		t.Fatal("Session does not implement SetOutputStream - RED phase expected")
	}
	streamSetter.SetOutputStream(&buf)
}

// TestSession_StreamsOutputInRealTime tests that output is written to stream as each line arrives.
func TestSession_StreamsOutputInRealTime(t *testing.T) {
	// Create a mock agent that outputs multiple lines with delays
	mockScript := createDelayedOutputMock(t)
	testAgent := agent.Agent{
		Name:          "test-stream",
		Path:          mockScript,
		Authenticated: true,
		Version:       "1.0.0",
		Pattern:       agent.KnownAgents()["claude"],
	}

	mgr := NewManager()
	sess, err := mgr.CreateSession(testAgent)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	defer func() { _ = sess.Close() }()

	// Set up output stream
	var buf bytes.Buffer
	streamSetter, ok := sess.(interface{ SetOutputStream(*bytes.Buffer) })
	if !ok {
		t.Fatal("Session does not implement SetOutputStream - RED phase expected")
	}
	streamSetter.SetOutputStream(&buf)

	ctx := context.Background()
	agentsPath := newTestAgentsFile(t)
	if err := sess.Start(ctx, agentsPath); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Track when lines arrive
	lineArrivalTimes := make([]time.Time, 0)
	go func() {
		for {
			if strings.Count(buf.String(), "\n") > len(lineArrivalTimes) {
				lineArrivalTimes = append(lineArrivalTimes, time.Now())
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Send a prompt that triggers multi-line output
	_, err = sess.Send(ctx, "output test")
	if err != nil {
		t.Errorf("Send() error = %v", err)
	}

	// Output should have been streamed (multiple lines in buffer)
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) < 2 {
		t.Errorf("Expected multiple lines in stream buffer, got %d", len(lines))
	}

	// Lines should have arrived at different times (real-time, not buffered)
	if len(lineArrivalTimes) >= 2 {
		gap := lineArrivalTimes[1].Sub(lineArrivalTimes[0])
		if gap < 10*time.Millisecond {
			t.Logf("Warning: lines arrived very quickly (%v) - may be buffered", gap)
		}
	}
}

// TestSession_StreamNilWriterNoOp tests that nil writer doesn't panic.
func TestSession_StreamNilWriterNoOp(t *testing.T) {
	mgr := NewManager()
	sess, err := mgr.CreateSession(newTestAgentWithMock(t))
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	defer func() { _ = sess.Close() }()

	// Setting nil writer should not panic
	streamSetter, ok := sess.(interface{ SetOutputStream(*bytes.Buffer) })
	if !ok {
		t.Fatal("Session does not implement SetOutputStream - RED phase expected")
	}

	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SetOutputStream(nil) panicked: %v", r)
		}
	}()
	streamSetter.SetOutputStream(nil)

	// Should still be able to use session normally
	ctx := context.Background()
	agentsPath := newTestAgentsFile(t)
	if err := sess.Start(ctx, agentsPath); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
}

// createDelayedOutputMock creates a mock script that outputs lines with delays.
func createDelayedOutputMock(t *testing.T) string {
	t.Helper()

	// Create a temporary script that outputs multiple lines with delays
	script := `#!/bin/bash
echo "Line 1"
sleep 0.1
echo "Line 2"
sleep 0.1
echo "Line 3"
echo "Context: 5% used"
`
	return createMockScript(t, script)
}

// createMockScript creates a temporary executable script.
func createMockScript(t *testing.T, content string) string {
	t.Helper()

	tmpDir := t.TempDir()
	scriptPath := tmpDir + "/mock-agent"

	if err := writeExecutableScript(scriptPath, content); err != nil {
		t.Fatalf("Failed to create mock script: %v", err)
	}

	return scriptPath
}

// writeExecutableScript writes a script and makes it executable.
func writeExecutableScript(path, content string) error {
	// This helper will be implemented in the GREEN phase
	// For now, this test file establishes the RED phase contract
	return nil
}
