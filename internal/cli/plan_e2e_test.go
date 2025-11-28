//go:build e2e

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

// TestPlanCommand_E2E_RealAgents tests the plan command against real agents.
// This test requires at least one real agent (claude, codex, cursor) to be
// installed and authenticated on the system.
func TestPlanCommand_E2E_RealAgents(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Reset flags
	resetPlanFlags()

	// Detect real agents on the system
	detector := agent.NewDetector()
	agents, err := detector.DetectAll()
	if err != nil {
		t.Fatalf("Failed to detect agents: %v", err)
	}

	// Filter to authenticated agents
	var authAgents []agent.Agent
	for _, a := range agents {
		if a.Authenticated {
			authAgents = append(authAgents, a)
		}
	}

	if len(authAgents) == 0 {
		t.Skip("No authenticated agents available - skipping e2e test")
	}

	t.Logf("Found %d authenticated agent(s): %v", len(authAgents), agentNames(authAgents))

	// Setup test environment
	agentsPath := testutil.CreateTestAgentsFile(t, e2eAgentsContent)
	workDir := testutil.CreateTestBeadsDir(t)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("Failed to change to work directory: %v", err)
	}

	// Run plan command with real agents
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{
		"plan",
		"--rounds", "1",
		"--agents-path", agentsPath,
		"--verbose",
		"Say hello and confirm you received this prompt",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	err = rootCmd.ExecuteContext(ctx)
	output := buf.String()

	t.Logf("Output:\n%s", output)

	if err != nil {
		t.Errorf("Plan command failed: %v", err)
	}

	// Verify basic output structure
	if !strings.Contains(output, "Planning:") {
		t.Error("Output should contain 'Planning:'")
	}
	if !strings.Contains(output, "Round 1") {
		t.Error("Output should contain 'Round 1'")
	}
	if !strings.Contains(output, "Planning complete") {
		t.Error("Output should contain 'Planning complete'")
	}
}

// TestPlanCommand_E2E_SingleAgent tests with a single specific agent.
func TestPlanCommand_E2E_SingleAgent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Reset flags
	resetPlanFlags()

	// Try each known agent in order of preference
	preferredAgents := []string{"claude", "codex", "cursor"}
	detector := agent.NewDetector()

	var selectedAgent *agent.Agent
	for _, name := range preferredAgents {
		if detector.IsInstalled(name) {
			agents, _ := detector.DetectAll()
			for _, a := range agents {
				if a.Name == name && a.Authenticated {
					selectedAgent = &a
					break
				}
			}
			if selectedAgent != nil {
				break
			}
		}
	}

	if selectedAgent == nil {
		t.Skip("No preferred agent (claude, codex, cursor) is authenticated")
	}

	t.Logf("Testing with agent: %s (version: %s)", selectedAgent.Name, selectedAgent.Version)

	// Setup
	agentsPath := testutil.CreateTestAgentsFile(t, e2eAgentsContent)
	workDir := testutil.CreateTestBeadsDir(t)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("Failed to change to work directory: %v", err)
	}

	// Use setAgentDetector to inject only our selected agent
	restore := setAgentDetector(func() ([]agent.Agent, error) {
		return []agent.Agent{*selectedAgent}, nil
	})
	defer restore()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{
		"plan",
		"--rounds", "1",
		"--agents-path", agentsPath,
		"--verbose",
		"Respond with exactly: BUCKSHOT_E2E_OK",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	err = rootCmd.ExecuteContext(ctx)
	output := buf.String()

	t.Logf("Output:\n%s", output)

	if err != nil {
		t.Errorf("Plan command with %s failed: %v", selectedAgent.Name, err)
	}

	// Verify the agent was used
	if !strings.Contains(output, selectedAgent.Name) {
		t.Errorf("Output should mention agent name '%s'", selectedAgent.Name)
	}
}

// TestPlanCommand_E2E_MultipleRounds tests multiple planning rounds.
func TestPlanCommand_E2E_MultipleRounds(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Reset flags
	resetPlanFlags()

	// Detect real agents
	detector := agent.NewDetector()
	agents, _ := detector.DetectAll()

	var authAgents []agent.Agent
	for _, a := range agents {
		if a.Authenticated {
			authAgents = append(authAgents, a)
		}
	}

	if len(authAgents) == 0 {
		t.Skip("No authenticated agents available")
	}

	// Setup
	agentsPath := testutil.CreateTestAgentsFile(t, e2eAgentsContent)
	workDir := testutil.CreateTestBeadsDir(t)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("Failed to change to work directory: %v", err)
	}

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{
		"plan",
		"--rounds", "2",
		"--agents-path", agentsPath,
		"--verbose",
		"Simple task: count from 1 to 3",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	err = rootCmd.ExecuteContext(ctx)
	output := buf.String()

	t.Logf("Output:\n%s", output)

	if err != nil {
		t.Errorf("Multi-round plan failed: %v", err)
	}

	// Should complete both rounds
	if !strings.Contains(output, "Round 1") {
		t.Error("Output should contain 'Round 1'")
	}
	if !strings.Contains(output, "Round 2") {
		t.Error("Output should contain 'Round 2'")
	}
	if !strings.Contains(output, "Completed 2 round") {
		t.Error("Output should indicate 2 rounds completed")
	}
}

// TestPlanCommand_E2E_AgentSelection tests --agents flag with real agents.
func TestPlanCommand_E2E_AgentSelection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Reset flags
	resetPlanFlags()

	// Detect real agents
	detector := agent.NewDetector()
	agents, _ := detector.DetectAll()

	var authAgents []agent.Agent
	for _, a := range agents {
		if a.Authenticated {
			authAgents = append(authAgents, a)
		}
	}

	if len(authAgents) < 2 {
		t.Skipf("Need at least 2 authenticated agents for selection test, got %d", len(authAgents))
	}

	// Select first agent only
	selectedName := authAgents[0].Name
	t.Logf("Selecting agent: %s (from %d available)", selectedName, len(authAgents))

	// Setup
	agentsPath := testutil.CreateTestAgentsFile(t, e2eAgentsContent)
	workDir := testutil.CreateTestBeadsDir(t)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("Failed to change to work directory: %v", err)
	}

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{
		"plan",
		"--rounds", "1",
		"--agents", selectedName,
		"--agents-path", agentsPath,
		"--verbose",
		"Acknowledge receipt",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	err = rootCmd.ExecuteContext(ctx)
	output := buf.String()

	t.Logf("Output:\n%s", output)

	if err != nil {
		t.Errorf("Agent selection failed: %v", err)
	}

	// Should only use selected agent
	if !strings.Contains(output, "Using 1 agent") {
		t.Error("Should use only 1 agent")
	}
	if !strings.Contains(output, selectedName) {
		t.Errorf("Should use selected agent '%s'", selectedName)
	}
}

// e2eAgentsContent is a minimal AGENTS.md for e2e testing
const e2eAgentsContent = `# E2E Test AGENTS.md

You are participating in an end-to-end test of the buckshot multi-agent planning system.

## Instructions

- Keep responses brief and focused
- Acknowledge that you received the prompt
- If asked to respond with a specific phrase, do so exactly
- Do not create any files or make system changes
`

// agentNames returns a slice of agent names
func agentNames(agents []agent.Agent) []string {
	names := make([]string, len(agents))
	for i, a := range agents {
		names[i] = a.Name
	}
	return names
}
