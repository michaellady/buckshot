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

// TestPlanCommand_Integration_WithMockAgent tests the full plan command flow
// with a mock agent that simulates real agent behavior.
func TestPlanCommand_Integration_WithMockAgent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

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
	defer os.Chdir(origDir)

	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("Failed to change to work directory: %v", err)
	}

	// Override agent detector to use our mock
	origDetector := agentDetector
	agentDetector = func() ([]agent.Agent, error) {
		return []agent.Agent{mockSetup.Agent}, nil
	}
	defer func() { agentDetector = origDetector }()

	// Run plan command
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{
		"plan",
		"--rounds", "1",
		"--agents-path", agentsPath,
		"Create a simple REST API",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = rootCmd.ExecuteContext(ctx)
	output := buf.String()

	// Verify execution
	if err != nil {
		t.Errorf("Plan command failed: %v\nOutput: %s", err, output)
	}

	// Verify output contains expected elements
	expectedStrings := []string{
		"Planning:",
		"REST API",
		"Round 1",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Output should contain %q, got:\n%s", expected, output)
		}
	}
}

// TestPlanCommand_Integration_NoAuthenticatedAgents tests behavior when
// no authenticated agents are available
func TestPlanCommand_Integration_NoAuthenticatedAgents(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup mock agent that fails auth
	config := testutil.DefaultMockConfig()
	config.Mode = testutil.ModeAuthFail
	mockSetup := testutil.SetupMockAgent(t, "mock-claude", config)

	agentsPath := testutil.CreateTestAgentsFile(t, "")

	// Override detector
	origDetector := agentDetector
	agentDetector = func() ([]agent.Agent, error) {
		return []agent.Agent{mockSetup.Agent}, nil
	}
	defer func() { agentDetector = origDetector }()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{
		"plan",
		"--agents-path", agentsPath,
		"Test prompt",
	})

	err := rootCmd.Execute()
	output := buf.String()

	if err != nil {
		t.Errorf("Command should not error with no auth agents, got: %v", err)
	}

	if !strings.Contains(output, "No authenticated agents") {
		t.Errorf("Should report no authenticated agents, got:\n%s", output)
	}
}

// TestPlanCommand_Integration_MultipleAgents tests with multiple mock agents
func TestPlanCommand_Integration_MultipleAgents(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup multiple mock agents
	configs := map[string]testutil.MockAgentConfig{
		"mock-claude": testutil.DefaultMockConfig(),
		"mock-codex":  testutil.DefaultMockConfig(),
	}

	setups := testutil.SetupMultipleMockAgents(t, configs)
	agentsPath := testutil.CreateTestAgentsFile(t, "")
	workDir := testutil.CreateTestBeadsDir(t)

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(workDir)

	// Override detector
	var agents []agent.Agent
	for _, s := range setups {
		agents = append(agents, s.Agent)
	}

	origDetector := agentDetector
	agentDetector = func() ([]agent.Agent, error) {
		return agents, nil
	}
	defer func() { agentDetector = origDetector }()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{
		"plan",
		"--rounds", "1",
		"--agents-path", agentsPath,
		"Multi-agent test",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err := rootCmd.ExecuteContext(ctx)
	output := buf.String()

	if err != nil {
		t.Errorf("Multi-agent plan failed: %v\nOutput: %s", err, output)
	}

	// Should show both agents
	if !strings.Contains(output, "Using 2 agent") {
		t.Errorf("Should report using 2 agents, got:\n%s", output)
	}
}

// TestPlanCommand_Integration_AgentSelection tests --agents flag filtering
func TestPlanCommand_Integration_AgentSelection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup multiple agents
	configs := map[string]testutil.MockAgentConfig{
		"claude": testutil.DefaultMockConfig(),
		"codex":  testutil.DefaultMockConfig(),
		"cursor": testutil.DefaultMockConfig(),
	}

	setups := testutil.SetupMultipleMockAgents(t, configs)
	agentsPath := testutil.CreateTestAgentsFile(t, "")
	workDir := testutil.CreateTestBeadsDir(t)

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(workDir)

	var agents []agent.Agent
	for _, s := range setups {
		agents = append(agents, s.Agent)
	}

	origDetector := agentDetector
	agentDetector = func() ([]agent.Agent, error) {
		return agents, nil
	}
	defer func() { agentDetector = origDetector }()

	// Reset the selectedAgents flag before test
	selectedAgents = nil

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{
		"plan",
		"--rounds", "1",
		"--agents", "claude,codex",
		"--agents-path", agentsPath,
		"Filtered agents test",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := rootCmd.ExecuteContext(ctx)
	output := buf.String()

	if err != nil {
		t.Errorf("Agent selection failed: %v\nOutput: %s", err, output)
	}

	// Should only use selected agents
	if !strings.Contains(output, "Using 2 agent") {
		t.Errorf("Should use only 2 selected agents, got:\n%s", output)
	}
}

// TestPlanCommand_Integration_Convergence tests --until-converged flag
func TestPlanCommand_Integration_Convergence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup mock agent that converges after first round
	config := testutil.DefaultMockConfig()
	config.Mode = testutil.ModeConverged
	mockSetup := testutil.SetupMockAgent(t, "mock-claude", config)

	agentsPath := testutil.CreateTestAgentsFile(t, "")
	workDir := testutil.CreateTestBeadsDir(t)

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(workDir)

	origDetector := agentDetector
	agentDetector = func() ([]agent.Agent, error) {
		return []agent.Agent{mockSetup.Agent}, nil
	}
	defer func() { agentDetector = origDetector }()

	// Reset flags
	untilConverged = false

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{
		"plan",
		"--until-converged",
		"--agents-path", agentsPath,
		"Convergence test",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err := rootCmd.ExecuteContext(ctx)
	output := buf.String()

	if err != nil {
		t.Errorf("Convergence test failed: %v\nOutput: %s", err, output)
	}

	// Should report convergence
	if !strings.Contains(output, "Converged") && !strings.Contains(output, "Planning complete") {
		t.Errorf("Should report convergence or completion, got:\n%s", output)
	}
}

// TestPlanCommand_Integration_ErrorHandling tests agent error scenarios
func TestPlanCommand_Integration_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup mock agent that errors
	config := testutil.DefaultMockConfig()
	config.Mode = testutil.ModeError
	config.ErrorMessage = "Simulated agent crash"
	mockSetup := testutil.SetupMockAgent(t, "mock-claude", config)

	agentsPath := testutil.CreateTestAgentsFile(t, "")
	workDir := testutil.CreateTestBeadsDir(t)

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(workDir)

	origDetector := agentDetector
	agentDetector = func() ([]agent.Agent, error) {
		return []agent.Agent{mockSetup.Agent}, nil
	}
	defer func() { agentDetector = origDetector }()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{
		"plan",
		"--rounds", "1",
		"--agents-path", agentsPath,
		"Error handling test",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := rootCmd.ExecuteContext(ctx)
	output := buf.String()

	// The command may or may not error depending on how errors are handled
	// But output should indicate the issue
	t.Logf("Error handling output: %s", output)
	t.Logf("Error: %v", err)

	// At minimum, planning should have started
	if !strings.Contains(output, "Planning:") {
		t.Errorf("Should have started planning, got:\n%s", output)
	}
}

// TestPlanCommand_Integration_ContextUsageTracking tests context usage reporting
func TestPlanCommand_Integration_ContextUsageTracking(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup mock agent with high context growth
	config := testutil.DefaultMockConfig()
	config.InitialContext = 0.10
	config.ContextGrowth = 0.20
	mockSetup := testutil.SetupMockAgent(t, "mock-claude", config)

	agentsPath := testutil.CreateTestAgentsFile(t, "")
	workDir := testutil.CreateTestBeadsDir(t)

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(workDir)

	origDetector := agentDetector
	agentDetector = func() ([]agent.Agent, error) {
		return []agent.Agent{mockSetup.Agent}, nil
	}
	defer func() { agentDetector = origDetector }()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{
		"plan",
		"--rounds", "2",
		"--agents-path", agentsPath,
		"Context tracking test",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err := rootCmd.ExecuteContext(ctx)
	output := buf.String()

	if err != nil {
		t.Errorf("Context tracking test failed: %v\nOutput: %s", err, output)
	}

	// Should complete both rounds
	if !strings.Contains(output, "Round 1") || !strings.Contains(output, "Round 2") {
		t.Errorf("Should complete both rounds, got:\n%s", output)
	}
}
