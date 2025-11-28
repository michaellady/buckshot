package session

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/michaellady/buckshot/internal/agent"
	"github.com/michaellady/buckshot/internal/testutil"
)

// TestRunOneShot_ReturnsOutput tests that RunOneShot captures agent output.
func TestRunOneShot_ReturnsOutput(t *testing.T) {
	// Setup mock agent that outputs a known response
	mockSetup := testutil.SetupMockAgent(t, "mock-oneshot", testutil.DefaultMockConfig())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := RunOneShot(ctx, mockSetup.Agent, "Say hello")
	if err != nil {
		t.Fatalf("RunOneShot failed: %v", err)
	}

	// Should return non-empty output
	if result.Output == "" {
		t.Error("RunOneShot should return non-empty output")
	}
}

// TestRunOneShot_WaitsForProcessExit tests that RunOneShot waits for the process to complete.
func TestRunOneShot_WaitsForProcessExit(t *testing.T) {
	mockSetup := testutil.SetupMockAgent(t, "mock-oneshot", testutil.DefaultMockConfig())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	start := time.Now()
	result, err := RunOneShot(ctx, mockSetup.Agent, "Test prompt")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("RunOneShot failed: %v", err)
	}

	// Should have waited for some time (mock agent takes at least a moment)
	if elapsed < 100*time.Millisecond {
		t.Logf("Warning: RunOneShot returned very quickly (%v), may not have waited for process", elapsed)
	}

	// Exit code should be 0 for successful execution
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}
}

// TestRunOneShot_RespectsContextCancellation tests timeout behavior.
func TestRunOneShot_RespectsContextCancellation(t *testing.T) {
	mockSetup := testutil.SetupMockAgent(t, "mock-oneshot", testutil.DefaultMockConfig())

	// Very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, err := RunOneShot(ctx, mockSetup.Agent, "Test prompt")

	// Should return an error due to context cancellation
	if err == nil {
		t.Error("RunOneShot should return error when context is cancelled")
	}
}

// TestRunOneShot_BuildsCorrectCommand tests that the command is built correctly.
func TestRunOneShot_BuildsCorrectCommand(t *testing.T) {
	// Create a mock agent with known pattern
	ag := agent.Agent{
		Name:          "test-agent",
		Path:          "/bin/echo", // Use echo for testing
		Authenticated: true,
		Pattern: agent.CLIPattern{
			NonInteractiveArgs: []string{"-n"},
			JSONOutputArgs:     []string{},
			SkipApprovalsArgs:  []string{},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := RunOneShot(ctx, ag, "test-prompt")
	if err != nil {
		t.Fatalf("RunOneShot failed: %v", err)
	}

	// Echo should output the prompt (with -n flag, no newline)
	if !strings.Contains(result.Output, "test-prompt") {
		t.Errorf("Output should contain prompt, got: %q", result.Output)
	}
}

// TestRunOneShot_CapturesStderr tests that stderr is also captured.
func TestRunOneShot_CapturesStderr(t *testing.T) {
	// Use a shell command that writes to stderr
	ag := agent.Agent{
		Name:          "test-stderr",
		Path:          "/bin/sh",
		Authenticated: true,
		Pattern: agent.CLIPattern{
			NonInteractiveArgs: []string{"-c"},
			JSONOutputArgs:     []string{},
			SkipApprovalsArgs:  []string{},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Command that writes to stderr
	result, err := RunOneShot(ctx, ag, "echo 'stderr message' >&2")
	if err != nil {
		t.Fatalf("RunOneShot failed: %v", err)
	}

	// Should capture stderr output
	if !strings.Contains(result.Output, "stderr message") {
		t.Errorf("Output should contain stderr message, got: %q", result.Output)
	}
}

// TestRunOneShot_HandlesNonZeroExitCode tests handling of failed commands.
func TestRunOneShot_HandlesNonZeroExitCode(t *testing.T) {
	ag := agent.Agent{
		Name:          "test-fail",
		Path:          "/bin/sh",
		Authenticated: true,
		Pattern: agent.CLIPattern{
			NonInteractiveArgs: []string{"-c"},
			JSONOutputArgs:     []string{},
			SkipApprovalsArgs:  []string{},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := RunOneShot(ctx, ag, "exit 42")

	// Should capture the exit code
	if result.ExitCode != 42 {
		t.Errorf("Expected exit code 42, got %d", result.ExitCode)
	}

	// Error should indicate non-zero exit
	if err == nil {
		t.Error("RunOneShot should return error for non-zero exit code")
	}
}

// TestRunOneShot_AppliesParser tests that output parser is applied.
func TestRunOneShot_AppliesParser(t *testing.T) {
	// Create agent with a parser that transforms output
	ag := agent.Agent{
		Name:          "test-parser",
		Path:          "/bin/echo",
		Authenticated: true,
		Pattern: agent.CLIPattern{
			NonInteractiveArgs: []string{},
			JSONOutputArgs:     []string{},
			SkipApprovalsArgs:  []string{},
		},
		Parser: &testParser{prefix: "PARSED: "},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := RunOneShot(ctx, ag, "raw output")
	if err != nil {
		t.Fatalf("RunOneShot failed: %v", err)
	}

	// Output should be transformed by parser
	if !strings.HasPrefix(result.Output, "PARSED: ") {
		t.Errorf("Output should be transformed by parser, got: %q", result.Output)
	}
}

// testParser is a simple parser for testing.
type testParser struct {
	prefix string
}

func (p *testParser) Parse(output string) string {
	return p.prefix + output
}
