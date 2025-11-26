package context

import (
	"os/exec"
	"strings"
	"testing"
)

// skipIfNoBeads skips the test if bd is not available or has no beads
func skipIfNoBeads(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("bd"); err != nil {
		t.Skip("Skipping test: bd command not available")
	}
	// Check if there are any beads
	out, err := exec.Command("bd", "list").Output()
	if err != nil || len(strings.TrimSpace(string(out))) == 0 {
		t.Skip("Skipping test: no beads in database")
	}
}

func TestBuild_CreatesContextWithPromptAndAgentsPath(t *testing.T) {
	builder := NewBuilder()

	prompt := "Review authentication logic"
	agentsPath := "/path/to/AGENTS.md"
	round := 1
	isFirstTurn := true

	ctx, err := builder.Build(prompt, agentsPath, round, isFirstTurn)
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	if ctx.Prompt != prompt {
		t.Errorf("Expected prompt %q, got %q", prompt, ctx.Prompt)
	}

	if ctx.AgentsPath != agentsPath {
		t.Errorf("Expected agentsPath %q, got %q", agentsPath, ctx.AgentsPath)
	}

	if ctx.Round != round {
		t.Errorf("Expected round %d, got %d", round, ctx.Round)
	}

	if ctx.IsFirstTurn != isFirstTurn {
		t.Errorf("Expected isFirstTurn %v, got %v", isFirstTurn, ctx.IsFirstTurn)
	}
}

func TestBuild_IncludesBeadsListOutput(t *testing.T) {
	skipIfNoBeads(t)

	builder := NewBuilder()

	ctx, err := builder.Build("test prompt", "/agents.md", 1, true)
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	if ctx.BeadsState == "" {
		t.Error("Expected BeadsState to be populated, got empty string")
	}

	// Should contain output that looks like bd list format
	// Expected format: "BEAD-ID [P#] [type] status - Title"
	if !strings.Contains(ctx.BeadsState, "[P") {
		t.Error("BeadsState should contain priority markers like [P1]")
	}
}

func TestBuild_IncludesBeadsShowDetailsForEachBead(t *testing.T) {
	skipIfNoBeads(t)

	builder := NewBuilder()

	ctx, err := builder.Build("test prompt", "/agents.md", 1, true)
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// BeadsState should include detailed 'bd show' output for each bead
	// Expected sections: Status, Priority, Type, Created, Updated, Description
	beadsState := ctx.BeadsState

	testCases := []string{
		"Status:",
		"Priority:",
		"Type:",
		"Created:",
		"Description:",
	}

	for _, expected := range testCases {
		if !strings.Contains(beadsState, expected) {
			t.Errorf("BeadsState should contain %q but doesn't", expected)
		}
	}
}

func TestFormat_ProducesLLMReadableOutput(t *testing.T) {
	builder := NewBuilder()

	ctx := PlanningContext{
		Prompt:      "Fix the bug in auth",
		AgentsPath:  "/path/to/AGENTS.md",
		BeadsState:  "test-123 [P1] [bug] open - Auth fails",
		Round:       1,
		IsFirstTurn: true,
	}

	output := builder.Format(ctx)

	if output == "" {
		t.Fatal("Format() returned empty string")
	}

	// Should include clear sections for LLM
	expectedSections := []string{
		"Prompt:",
		"AGENTS.md:",
		"Current Beads:",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Formatted output should contain section %q", section)
		}
	}

	// Should include the actual prompt text
	if !strings.Contains(output, ctx.Prompt) {
		t.Error("Formatted output should contain the original prompt")
	}

	// Should include the agents path
	if !strings.Contains(output, ctx.AgentsPath) {
		t.Error("Formatted output should contain the AGENTS.md path")
	}

	// Should include beads state
	if !strings.Contains(output, ctx.BeadsState) {
		t.Error("Formatted output should contain the beads state")
	}
}

func TestFormat_IncludesInstructionsForModifyingBeads(t *testing.T) {
	builder := NewBuilder()

	ctx := PlanningContext{
		Prompt:      "Review code",
		AgentsPath:  "/agents.md",
		BeadsState:  "test-456 [P1] [task] open - Test task",
		Round:       1,
		IsFirstTurn: true,
	}

	output := builder.Format(ctx)

	// Should include instructions about how to modify beads
	instructionKeywords := []string{
		"bd update",
		"bd create",
		"bd close",
	}

	foundAny := false
	for _, keyword := range instructionKeywords {
		if strings.Contains(output, keyword) {
			foundAny = true
			break
		}
	}

	if !foundAny {
		t.Error("Formatted output should include instructions for modifying beads (bd update, bd create, bd close)")
	}
}

func TestFormat_FirstTurnIncludesAgentGuidance(t *testing.T) {
	builder := NewBuilder()

	ctx := PlanningContext{
		Prompt:      "Review beads",
		AgentsPath:  "/agents.md",
		BeadsState:  "test-789 [P1] [task] open - Task",
		Round:       1,
		IsFirstTurn: true,
	}

	output := builder.Format(ctx)

	// First turn should include guidance for the agent
	guidanceKeywords := []string{
		"Read AGENTS.md",
		"read and apply",
	}

	foundAny := false
	for _, keyword := range guidanceKeywords {
		if strings.Contains(output, keyword) {
			foundAny = true
			break
		}
	}

	if !foundAny {
		t.Error("First turn should include guidance to read AGENTS.md")
	}
}

func TestFormat_SubsequentRoundsIndicateRoundNumber(t *testing.T) {
	builder := NewBuilder()

	ctx := PlanningContext{
		Prompt:      "Continue review",
		AgentsPath:  "/agents.md",
		BeadsState:  "test-abc [P1] [task] open - Task",
		Round:       3,
		IsFirstTurn: false,
	}

	output := builder.Format(ctx)

	// Should indicate round number
	if !strings.Contains(output, "Round") && !strings.Contains(output, "round") {
		t.Error("Output should indicate the round number for subsequent rounds")
	}
}

func TestRefreshBeadsState_UpdatesBeadsState(t *testing.T) {
	builder := NewBuilder()

	ctx := PlanningContext{
		Prompt:      "Test",
		AgentsPath:  "/agents.md",
		BeadsState:  "old state",
		Round:       1,
		IsFirstTurn: true,
	}

	err := builder.RefreshBeadsState(&ctx)
	if err != nil {
		t.Fatalf("RefreshBeadsState() failed: %v", err)
	}

	if ctx.BeadsState == "old state" {
		t.Error("RefreshBeadsState() should update BeadsState, but it remained unchanged")
	}

	if ctx.BeadsState == "" {
		t.Error("RefreshBeadsState() should populate BeadsState, but got empty string")
	}
}

func TestBuild_HandlesMissingBeadsGracefully(t *testing.T) {
	builder := NewBuilder()

	// This test ensures that if 'bd list' returns empty or fails,
	// the builder doesn't crash
	ctx, err := builder.Build("prompt", "/agents.md", 1, true)
	if err != nil {
		t.Fatalf("Build() should handle missing beads gracefully, got error: %v", err)
	}

	// Should still return a valid context, even if BeadsState is minimal
	if ctx.Prompt != "prompt" {
		t.Error("Context should still have prompt set even if beads are missing")
	}
}

func TestFormat_ClearSectionSeparation(t *testing.T) {
	builder := NewBuilder()

	ctx := PlanningContext{
		Prompt:      "Test prompt",
		AgentsPath:  "/agents.md",
		BeadsState:  "bead-123 [P1] [task] open - Test",
		Round:       1,
		IsFirstTurn: true,
	}

	output := builder.Format(ctx)

	// Check that sections are clearly separated (e.g., with blank lines or headers)
	lines := strings.Split(output, "\n")

	if len(lines) < 5 {
		t.Error("Formatted output should have multiple lines with clear structure")
	}

	// Should have some empty lines for separation
	hasEmptyLines := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			hasEmptyLines = true
			break
		}
	}

	if !hasEmptyLines {
		t.Error("Formatted output should use blank lines to separate sections for readability")
	}
}

func TestBuild_IncludesBeadDependencies(t *testing.T) {
	skipIfNoBeads(t)

	builder := NewBuilder()

	ctx, err := builder.Build("test", "/agents.md", 1, true)
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// BeadsState should include dependency information from 'bd show'
	// Expected to contain "Depends on" or "Blocks" sections
	beadsState := ctx.BeadsState

	// Check if at least one bead has dependency information
	// (since we know from our sample data that some beads have dependencies)
	hasDependencyInfo := strings.Contains(beadsState, "Depends on") ||
		strings.Contains(beadsState, "Blocks") ||
		strings.Contains(beadsState, "→") || // Arrow symbols used in bd show
		strings.Contains(beadsState, "←")

	if !hasDependencyInfo {
		t.Error("BeadsState should include dependency information from bd show output")
	}
}
