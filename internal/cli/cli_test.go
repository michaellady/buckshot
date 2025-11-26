package cli

import (
	"bytes"
	"strings"
	"testing"
)

// TestRootCommand tests the root command exists and has expected structure
func TestRootCommand(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd is nil")
	}

	if rootCmd.Use != "buckshot" {
		t.Errorf("rootCmd.Use = %q, want %q", rootCmd.Use, "buckshot")
	}

	// Should have subcommands
	subCommands := rootCmd.Commands()
	if len(subCommands) < 2 {
		t.Errorf("rootCmd has %d subcommands, want at least 2 (plan, agents)", len(subCommands))
	}
}

// TestPlanCommand_Exists tests the plan command exists
func TestPlanCommand_Exists(t *testing.T) {
	if planCmd == nil {
		t.Fatal("planCmd is nil")
	}

	if planCmd.Use != "plan [prompt]" {
		t.Errorf("planCmd.Use = %q, want %q", planCmd.Use, "plan [prompt]")
	}
}

// TestPlanCommand_RequiresPrompt tests that plan command requires a prompt argument
func TestPlanCommand_RequiresPrompt(t *testing.T) {
	// Reset command state
	rootCmd.SetArgs([]string{"plan"})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	if err == nil {
		t.Error("plan command should require prompt argument, got no error")
	}

	// Error should mention the required argument
	errStr := err.Error()
	if !strings.Contains(errStr, "accepts 1 arg") {
		t.Errorf("Error message should mention required args, got: %s", errStr)
	}
}

// TestPlanCommand_AcceptsPrompt tests that plan command accepts a prompt
func TestPlanCommand_AcceptsPrompt(t *testing.T) {
	rootCmd.SetArgs([]string{"plan", "Create a new feature"})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("plan command with prompt should not error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Create a new feature") {
		t.Errorf("Output should contain prompt, got: %s", output)
	}
}

// TestPlanCommand_RoundsFlag tests the --rounds flag
func TestPlanCommand_RoundsFlag(t *testing.T) {
	// Check default value
	flag := planCmd.Flags().Lookup("rounds")
	if flag == nil {
		t.Fatal("--rounds flag not found")
	}

	if flag.DefValue != "3" {
		t.Errorf("--rounds default = %q, want %q", flag.DefValue, "3")
	}

	// Check shorthand
	if flag.Shorthand != "r" {
		t.Errorf("--rounds shorthand = %q, want %q", flag.Shorthand, "r")
	}
}

// TestPlanCommand_RoundsFlagCustomValue tests setting custom rounds
func TestPlanCommand_RoundsFlagCustomValue(t *testing.T) {
	// Reset to default
	rounds = 3

	rootCmd.SetArgs([]string{"plan", "--rounds", "5", "Test prompt"})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("plan command with --rounds should not error, got: %v", err)
	}

	if rounds != 5 {
		t.Errorf("rounds = %d, want 5", rounds)
	}
}

// TestPlanCommand_AgentsPathFlag tests the --agents-path flag
func TestPlanCommand_AgentsPathFlag(t *testing.T) {
	flag := planCmd.Flags().Lookup("agents-path")
	if flag == nil {
		t.Fatal("--agents-path flag not found")
	}

	if flag.Shorthand != "a" {
		t.Errorf("--agents-path shorthand = %q, want %q", flag.Shorthand, "a")
	}
}

// TestPlanCommand_AgentsPathFlagCustomValue tests setting custom agents path
func TestPlanCommand_AgentsPathFlagCustomValue(t *testing.T) {
	// Reset
	agentsPath = ""

	rootCmd.SetArgs([]string{"plan", "--agents-path", "/custom/AGENTS.md", "Test prompt"})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("plan command with --agents-path should not error, got: %v", err)
	}

	if agentsPath != "/custom/AGENTS.md" {
		t.Errorf("agentsPath = %q, want %q", agentsPath, "/custom/AGENTS.md")
	}
}

// TestPlanCommand_AgentsFlag tests the --agents flag
func TestPlanCommand_AgentsFlag(t *testing.T) {
	flag := planCmd.Flags().Lookup("agents")
	if flag == nil {
		t.Fatal("--agents flag not found")
	}
}

// TestPlanCommand_AgentsFlagMultipleValues tests selecting multiple agents
func TestPlanCommand_AgentsFlagMultipleValues(t *testing.T) {
	// Reset
	selectedAgents = nil

	rootCmd.SetArgs([]string{"plan", "--agents", "claude,codex", "Test prompt"})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("plan command with --agents should not error, got: %v", err)
	}

	if len(selectedAgents) != 2 {
		t.Errorf("selectedAgents has %d items, want 2", len(selectedAgents))
	}

	if selectedAgents[0] != "claude" || selectedAgents[1] != "codex" {
		t.Errorf("selectedAgents = %v, want [claude, codex]", selectedAgents)
	}
}

// TestPlanCommand_UntilConvergedFlag tests the --until-converged flag
func TestPlanCommand_UntilConvergedFlag(t *testing.T) {
	flag := planCmd.Flags().Lookup("until-converged")
	if flag == nil {
		t.Fatal("--until-converged flag not found")
	}

	if flag.DefValue != "false" {
		t.Errorf("--until-converged default = %q, want %q", flag.DefValue, "false")
	}
}

// TestPlanCommand_UntilConvergedFlagSet tests enabling until-converged
func TestPlanCommand_UntilConvergedFlagSet(t *testing.T) {
	// Reset
	untilConverged = false

	rootCmd.SetArgs([]string{"plan", "--until-converged", "Test prompt"})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("plan command with --until-converged should not error, got: %v", err)
	}

	if !untilConverged {
		t.Error("untilConverged = false, want true")
	}
}

// TestAgentsCommand_Exists tests the agents command exists
func TestAgentsCommand_Exists(t *testing.T) {
	if agentsCmd == nil {
		t.Fatal("agentsCmd is nil")
	}

	if agentsCmd.Use != "agents" {
		t.Errorf("agentsCmd.Use = %q, want %q", agentsCmd.Use, "agents")
	}
}

// TestAgentsCommand_ListsAgents tests that agents command runs and lists agents
func TestAgentsCommand_ListsAgents(t *testing.T) {
	rootCmd.SetArgs([]string{"agents"})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("agents command should not error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "agent") {
		t.Errorf("Output should mention agents, got: %s", output)
	}
}

// TestAgentsCommand_ShowsStatus tests that agents command shows agent status
func TestAgentsCommand_ShowsStatus(t *testing.T) {
	rootCmd.SetArgs([]string{"agents"})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("agents command should not error, got: %v", err)
	}

	// In implementation, output should show status for each agent
	// For now, just verify it runs
	output := buf.String()
	if output == "" {
		t.Error("agents command should produce output")
	}
}

// TestVersion tests the --version flag
func TestVersion(t *testing.T) {
	rootCmd.Version = "1.0.0"
	rootCmd.SetArgs([]string{"--version"})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("--version should not error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "1.0.0") {
		t.Errorf("Output should contain version, got: %s", output)
	}
}

// TestExecute tests the Execute function
func TestExecute(t *testing.T) {
	// Execute with no args should show help (not error)
	rootCmd.SetArgs([]string{})

	err := Execute("test-version")
	// No args should show help but not error
	if err != nil {
		t.Errorf("Execute() with no args should not error, got: %v", err)
	}
}

// TestPlanCommand_OutputFormat tests the expected output format
func TestPlanCommand_OutputFormat(t *testing.T) {
	rootCmd.SetArgs([]string{"plan", "Build a REST API"})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("plan command should not error, got: %v", err)
	}

	output := buf.String()

	// Should show the prompt
	if !strings.Contains(output, "Build a REST API") {
		t.Errorf("Output should contain prompt, got: %s", output)
	}

	// Should show rounds configuration
	if !strings.Contains(output, "Rounds") {
		t.Errorf("Output should show rounds, got: %s", output)
	}
}

// TestPlanCommand_ShorthandFlags tests shorthand flag usage
func TestPlanCommand_ShorthandFlags(t *testing.T) {
	// Reset
	rounds = 3
	agentsPath = ""

	rootCmd.SetArgs([]string{"plan", "-r", "7", "-a", "/path/to/agents.md", "Test"})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("shorthand flags should work, got: %v", err)
	}

	if rounds != 7 {
		t.Errorf("rounds = %d, want 7", rounds)
	}
	if agentsPath != "/path/to/agents.md" {
		t.Errorf("agentsPath = %q, want %q", agentsPath, "/path/to/agents.md")
	}
}

// TestPlanCommand_SaveFlag tests the --save flag
func TestPlanCommand_SaveFlag(t *testing.T) {
	flag := planCmd.Flags().Lookup("save")
	if flag == nil {
		t.Fatal("--save flag not found")
	}

	if flag.DefValue != "" {
		t.Errorf("--save default = %q, want empty string", flag.DefValue)
	}
}

// TestPlanCommand_SaveFlagCustomValue tests setting a bead ID to save to
func TestPlanCommand_SaveFlagCustomValue(t *testing.T) {
	// Reset
	saveToBead = ""

	rootCmd.SetArgs([]string{"plan", "--save", "buckshot-123", "Test prompt"})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("plan command with --save should not error, got: %v", err)
	}

	// Verify the flag value was set correctly
	if saveToBead != "buckshot-123" {
		t.Errorf("saveToBead = %q, want %q", saveToBead, "buckshot-123")
	}

	// Note: The "Saving perspectives to:" message is only printed when
	// authenticated agents are available. In CI/test environments without
	// agents, the command exits early with "No authenticated agents available".
}
