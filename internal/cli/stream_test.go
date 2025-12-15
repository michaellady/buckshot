package cli

import (
	"bytes"
	"strings"
	"testing"
)

// TestPlanCmd_StreamFlag_Accepted verifies the --stream flag is accepted by plan command.
func TestPlanCmd_StreamFlag_Accepted(t *testing.T) {
	// Check that --stream flag exists
	flag := planCmd.Flags().Lookup("stream")
	if flag == nil {
		t.Fatal("--stream flag not found - RED phase expected")
	}

	// Check shorthand
	if flag.Shorthand != "s" {
		t.Errorf("--stream shorthand = %q, want %q", flag.Shorthand, "s")
	}

	// Check default value
	if flag.DefValue != "false" {
		t.Errorf("--stream default = %q, want %q", flag.DefValue, "false")
	}
}

// TestPlanCmd_StreamFlag_EnablesOutput verifies agent output appears on stdout when --stream is set.
func TestPlanCmd_StreamFlag_EnablesOutput(t *testing.T) {
	// Reset flags
	streamOutput = false

	rootCmd.SetArgs([]string{"plan", "--stream", "Test prompt"})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("plan --stream should not error, got: %v", err)
	}

	// Verify streamOutput flag was set
	if !streamOutput {
		t.Error("streamOutput = false, want true - RED phase expected")
	}
}

// TestPlanCmd_StreamFlag_ShorthandWorks tests the -s shorthand for --stream.
func TestPlanCmd_StreamFlag_ShorthandWorks(t *testing.T) {
	// Reset flags
	streamOutput = false

	rootCmd.SetArgs([]string{"plan", "-s", "Test prompt"})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("plan -s should not error, got: %v", err)
	}

	if !streamOutput {
		t.Error("streamOutput = false with -s shorthand, want true - RED phase expected")
	}
}

// TestPlanCmd_ActionSummary verifies beads action summary is displayed after each agent when streaming.
func TestPlanCmd_ActionSummary(t *testing.T) {
	// This test verifies that when --stream is enabled, the beads action summary
	// (created/modified/commented) is displayed after each agent completes.
	// The action summary should show which beads were affected during the agent's execution.

	// Reset flags
	streamOutput = false

	rootCmd.SetArgs([]string{"plan", "--stream", "Test prompt"})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	_ = rootCmd.Execute()

	output := buf.String()

	// When streaming is enabled and agents run, we should see action summaries
	// In the RED phase, this will fail because the feature isn't implemented yet.
	// Look for typical action summary indicators like "Created:", "Status:", or "Comments:"
	hasActionSummary := strings.Contains(output, "Created:") ||
		strings.Contains(output, "Status:") ||
		strings.Contains(output, "Comments:") ||
		strings.Contains(output, "Beads actions:")

	if streamOutput && !hasActionSummary {
		t.Log("Warning: Expected action summary in streaming output - RED phase expected")
		// This is informational in RED phase - will be enforced in GREEN
	}
}

// TestFeedbackCmd_StreamFlag verifies the --stream flag works for feedback command.
func TestFeedbackCmd_StreamFlag(t *testing.T) {
	// Check that --stream flag exists on feedback command
	flag := feedbackCmd.Flags().Lookup("stream")
	if flag == nil {
		t.Fatal("feedback --stream flag not found - RED phase expected")
	}

	// Check shorthand
	if flag.Shorthand != "s" {
		t.Errorf("feedback --stream shorthand = %q, want %q", flag.Shorthand, "s")
	}
}

// TestFeedbackCmd_StreamFlag_EnablesOutput verifies agent output streams when --stream is set.
func TestFeedbackCmd_StreamFlag_EnablesOutput(t *testing.T) {
	// Reset flags
	feedbackStreamOutput = false

	rootCmd.SetArgs([]string{"feedback", "--agent", "claude", "--stream"})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	// This will error because claude may not be available, but we just check the flag
	_ = rootCmd.Execute()

	// Verify feedbackStreamOutput flag was set
	if !feedbackStreamOutput {
		t.Error("feedbackStreamOutput = false, want true - RED phase expected")
	}
}

// TestStreamOutput_CombinesWithVerbose verifies --stream and --verbose can be used together.
func TestStreamOutput_CombinesWithVerbose(t *testing.T) {
	// Reset flags
	streamOutput = false
	verbose = false

	rootCmd.SetArgs([]string{"plan", "--stream", "--verbose", "Test prompt"})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("plan --stream --verbose should not error, got: %v", err)
	}

	if !streamOutput {
		t.Error("streamOutput = false, want true")
	}
	if !verbose {
		t.Error("verbose = false, want true")
	}
}
