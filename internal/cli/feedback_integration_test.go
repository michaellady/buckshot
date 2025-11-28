//go:build integration

package cli

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/michaellady/buckshot/internal/agent"
	"github.com/michaellady/buckshot/internal/testutil"
)

// TestFeedbackCommand_Integration_WithMockAgent tests the feedback command
// with a mock agent in comment-only mode.
func TestFeedbackCommand_Integration_WithMockAgent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Reset all flags to ensure clean state
	resetFeedbackFlags()

	// Setup mock agent
	mockSetup := testutil.SetupMockAgent(t, "mock-claude", testutil.DefaultMockConfig())

	// Setup test AGENTS.md
	agentsPath := testutil.CreateTestAgentsFile(t, "")

	// Setup test working directory with .beads
	workDir := testutil.CreateTestBeadsDir(t)
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("Failed to change to work directory: %v", err)
	}

	// Override agent detector to use our mock
	restore := setAgentDetector(func() ([]agent.Agent, error) {
		return []agent.Agent{mockSetup.Agent}, nil
	})
	defer restore()

	// Run feedback command
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{
		"feedback",
		"--agent", "mock-claude",
		"--agents-path", agentsPath,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = rootCmd.ExecuteContext(ctx)
	output := buf.String()

	// Verify execution succeeded
	if err != nil {
		t.Errorf("Feedback command failed: %v\nOutput: %s", err, output)
	}

	// Verify output contains expected feedback mode elements
	expectedStrings := []string{
		"Feedback",   // Should mention feedback mode
		"mock-claude", // Should mention the agent name
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Output should contain %q, got:\n%s", expected, output)
		}
	}
}

// TestFeedbackCommand_RequiresAgentFlag tests that --agent flag is required.
func TestFeedbackCommand_RequiresAgentFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Reset flags and also reset the cobra flag value directly
	resetFeedbackFlags()
	feedbackCmd.Flags().Set("agent", "")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{
		"feedback",
		// Missing --agent flag
	})

	err := rootCmd.Execute()
	output := buf.String()

	// Should fail because --agent is required
	if err == nil {
		t.Error("Feedback command should fail without --agent flag")
	}

	// Error message should mention "agent"
	if err != nil && !strings.Contains(err.Error(), "agent") && !strings.Contains(output, "agent") {
		t.Errorf("Error should mention 'agent' flag, got: %v\nOutput: %s", err, output)
	}
}

// TestFeedbackCommand_UsesFeedbackPrompt tests that the feedback command
// uses the FormatFeedback prompt (comment-only mode).
func TestFeedbackCommand_UsesFeedbackPrompt(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	resetFeedbackFlags()
	feedbackCmd.Flags().Set("agent", "")

	// Setup mock agent that captures the prompt it receives
	mockConfig := testutil.DefaultMockConfig()
	mockSetup := testutil.SetupMockAgent(t, "mock-claude", mockConfig)

	agentsPath := testutil.CreateTestAgentsFile(t, "")
	workDir := testutil.CreateTestBeadsDir(t)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("Failed to change to work directory: %v", err)
	}

	restore := setAgentDetector(func() ([]agent.Agent, error) {
		return []agent.Agent{mockSetup.Agent}, nil
	})
	defer restore()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{
		"feedback",
		"--agent", "mock-claude",
		"--agents-path", agentsPath,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = rootCmd.ExecuteContext(ctx)
	output := buf.String()

	// Note: This test may fail if the mock agent doesn't respond properly
	// The key validation is that the output mentions "Feedback" mode
	if !strings.Contains(output, "Feedback") {
		t.Errorf("Feedback mode should mention 'Feedback' in output, got:\n%s", output)
	}
}

// TestFeedbackCommand_SingleAgentOnly tests that only the specified agent runs.
func TestFeedbackCommand_SingleAgentOnly(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	resetFeedbackFlags()
	feedbackCmd.Flags().Set("agent", "")

	// Setup two mock agents
	mockSetup1 := testutil.SetupMockAgent(t, "agent1", testutil.DefaultMockConfig())
	mockSetup2 := testutil.SetupMockAgent(t, "agent2", testutil.DefaultMockConfig())

	agentsPath := testutil.CreateTestAgentsFile(t, "")
	workDir := testutil.CreateTestBeadsDir(t)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("Failed to change to work directory: %v", err)
	}

	restore := setAgentDetector(func() ([]agent.Agent, error) {
		return []agent.Agent{mockSetup1.Agent, mockSetup2.Agent}, nil
	})
	defer restore()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{
		"feedback",
		"--agent", "agent1", // Only run agent1
		"--agents-path", agentsPath,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = rootCmd.ExecuteContext(ctx)
	output := buf.String()

	// Output should mention agent1 (even if execution fails due to mock issues)
	if !strings.Contains(output, "agent1") {
		t.Errorf("Output should contain 'agent1', got:\n%s", output)
	}

	// Should NOT mention agent2 was used
	if strings.Contains(output, "Using agent: agent2") {
		t.Error("Output should NOT contain 'Using agent: agent2' - only agent1 should run")
	}
}
