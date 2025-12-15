package cli

import (
	"testing"
)

// TestPlanCmd_StreamFlag_Accepted verifies the --stream flag is accepted by plan command.
func TestPlanCmd_StreamFlag_Accepted(t *testing.T) {
	// Check that --stream flag exists
	flag := planCmd.Flags().Lookup("stream")
	if flag == nil {
		t.Fatal("--stream flag not found")
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

// TestPlanCmd_StreamFlag_ParsesCorrectly verifies the --stream flag sets the streamOutput variable.
func TestPlanCmd_StreamFlag_ParsesCorrectly(t *testing.T) {
	// Save and restore original value
	originalValue := streamOutput
	defer func() { streamOutput = originalValue }()

	// Reset
	streamOutput = false

	// Parse just the --stream flag
	if err := planCmd.Flags().Set("stream", "true"); err != nil {
		t.Fatalf("Failed to set --stream flag: %v", err)
	}

	// Verify flag value
	if !streamOutput {
		t.Error("streamOutput = false after setting --stream, want true")
	}

	// Reset for next test
	_ = planCmd.Flags().Set("stream", "false")
}

// TestFeedbackCmd_StreamFlag_Accepted verifies the --stream flag works for feedback command.
func TestFeedbackCmd_StreamFlag_Accepted(t *testing.T) {
	// Check that --stream flag exists on feedback command
	flag := feedbackCmd.Flags().Lookup("stream")
	if flag == nil {
		t.Fatal("feedback --stream flag not found")
	}

	// Check shorthand
	if flag.Shorthand != "s" {
		t.Errorf("feedback --stream shorthand = %q, want %q", flag.Shorthand, "s")
	}

	// Check default value
	if flag.DefValue != "false" {
		t.Errorf("feedback --stream default = %q, want %q", flag.DefValue, "false")
	}
}

// TestFeedbackCmd_StreamFlag_ParsesCorrectly verifies the --stream flag sets feedbackStreamOutput.
func TestFeedbackCmd_StreamFlag_ParsesCorrectly(t *testing.T) {
	// Save and restore original value
	originalValue := feedbackStreamOutput
	defer func() { feedbackStreamOutput = originalValue }()

	// Reset
	feedbackStreamOutput = false

	// Parse just the --stream flag
	if err := feedbackCmd.Flags().Set("stream", "true"); err != nil {
		t.Fatalf("Failed to set --stream flag: %v", err)
	}

	// Verify flag value
	if !feedbackStreamOutput {
		t.Error("feedbackStreamOutput = false after setting --stream, want true")
	}

	// Reset for next test
	_ = feedbackCmd.Flags().Set("stream", "false")
}

// TestStreamAndVerboseFlags verifies --stream and --verbose flags can be set together.
func TestStreamAndVerboseFlags(t *testing.T) {
	// Save and restore original values
	originalStream := streamOutput
	originalVerbose := verbose
	defer func() {
		streamOutput = originalStream
		verbose = originalVerbose
	}()

	// Reset
	streamOutput = false
	verbose = false

	// Set both flags
	if err := planCmd.Flags().Set("stream", "true"); err != nil {
		t.Fatalf("Failed to set --stream flag: %v", err)
	}
	if err := planCmd.Flags().Set("verbose", "true"); err != nil {
		t.Fatalf("Failed to set --verbose flag: %v", err)
	}

	// Verify both flags
	if !streamOutput {
		t.Error("streamOutput = false, want true")
	}
	if !verbose {
		t.Error("verbose = false, want true")
	}

	// Reset for next test
	_ = planCmd.Flags().Set("stream", "false")
	_ = planCmd.Flags().Set("verbose", "false")
}
