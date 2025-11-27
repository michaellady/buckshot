// Package testutil provides test utilities for buckshot integration tests.
package testutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/michaellady/buckshot/internal/agent"
)

// MockAgentMode defines the behavior mode for the mock agent
type MockAgentMode string

const (
	// ModeDefault is normal operation with simulated responses
	ModeDefault MockAgentMode = "default"
	// ModeError causes the agent to return an error after some messages
	ModeError MockAgentMode = "error"
	// ModeTimeout causes the agent to hang (for timeout testing)
	ModeTimeout MockAgentMode = "timeout"
	// ModeAuthFail causes authentication to fail
	ModeAuthFail MockAgentMode = "auth_fail"
	// ModeConverged causes the agent to report no changes needed
	ModeConverged MockAgentMode = "converged"
)

// MockAgentConfig configures mock agent behavior
type MockAgentConfig struct {
	Mode           MockAgentMode
	InitialContext float64 // 0.0-1.0
	ContextGrowth  float64 // Growth per message
	ResponseDelay  int     // Milliseconds
	ErrorMessage   string
}

// DefaultMockConfig returns a default mock agent configuration
func DefaultMockConfig() MockAgentConfig {
	return MockAgentConfig{
		Mode:           ModeDefault,
		InitialContext: 0.01,
		ContextGrowth:  0.05,
		ResponseDelay:  0,
		ErrorMessage:   "Mock error occurred",
	}
}

// MockAgentSetup contains the setup for a mock agent in tests
type MockAgentSetup struct {
	BinaryPath string   // Path to the mock agent binary
	Agent      agent.Agent
	Cleanup    func()
}

// BuildMockAgent builds the mock agent binary and returns its path.
// The binary is built in a temporary directory and cleaned up when the test completes.
func BuildMockAgent(t *testing.T) string {
	t.Helper()

	// Find the mock agent source
	mockagentSrc := findMockAgentSource(t)

	// Create a temporary directory for the binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "mock-agent")

	// Build the mock agent
	cmd := exec.Command("go", "build", "-o", binaryPath, mockagentSrc)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build mock agent: %v", err)
	}

	return binaryPath
}

// SetupMockAgent creates a complete mock agent setup for testing
func SetupMockAgent(t *testing.T, name string, config MockAgentConfig) *MockAgentSetup {
	t.Helper()

	binaryPath := BuildMockAgent(t)

	// Create a wrapper script that passes the config flags
	wrapperPath := createAgentWrapper(t, binaryPath, name, config)

	setup := &MockAgentSetup{
		BinaryPath: wrapperPath,
		Agent: agent.Agent{
			Name:          name,
			Path:          wrapperPath,
			Authenticated: config.Mode != ModeAuthFail,
			Version:       "1.0.0-mock",
			Pattern:       createMockPattern(name),
		},
		Cleanup: func() {
			// Cleanup is handled by t.TempDir()
		},
	}

	return setup
}

// SetupMultipleMockAgents creates multiple mock agents for multi-agent testing
func SetupMultipleMockAgents(t *testing.T, configs map[string]MockAgentConfig) []*MockAgentSetup {
	t.Helper()

	var setups []*MockAgentSetup
	for name, config := range configs {
		setup := SetupMockAgent(t, name, config)
		setups = append(setups, setup)
	}
	return setups
}

// CreateTestAgentsFile creates a temporary AGENTS.md file for testing
func CreateTestAgentsFile(t *testing.T, content string) string {
	t.Helper()

	if content == "" {
		content = DefaultAgentsContent
	}

	tmpDir := t.TempDir()
	agentsPath := filepath.Join(tmpDir, "AGENTS.md")

	if err := os.WriteFile(agentsPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test AGENTS.md: %v", err)
	}

	return agentsPath
}

// DefaultAgentsContent is a minimal AGENTS.md for testing
const DefaultAgentsContent = `# Test AGENTS.md

This is a test AGENTS.md file for buckshot testing.

## Instructions

- Analyze the planning prompt
- Create beads using 'bd create' for tasks
- Use 'bd list' to see current state
- Report changes made
`

// CreateTestBeadsDir creates a temporary .beads directory for testing
func CreateTestBeadsDir(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	beadsDir := filepath.Join(tmpDir, ".beads")

	if err := os.MkdirAll(beadsDir, 0755); err != nil {
		t.Fatalf("Failed to create .beads directory: %v", err)
	}

	// Create empty issues.jsonl
	issuesPath := filepath.Join(beadsDir, "issues.jsonl")
	if err := os.WriteFile(issuesPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create issues.jsonl: %v", err)
	}

	return tmpDir
}

// findMockAgentSource finds the path to the mock agent source code
func findMockAgentSource(t *testing.T) string {
	t.Helper()

	// Try relative paths from common test locations
	paths := []string{
		"testdata/mockagent",
		"../testdata/mockagent",
		"../../testdata/mockagent",
		"../../../testdata/mockagent",
	}

	for _, p := range paths {
		if _, err := os.Stat(filepath.Join(p, "main.go")); err == nil {
			return p
		}
	}

	// Try from GOPATH/module root
	if gomod := findGoMod(t); gomod != "" {
		mockPath := filepath.Join(filepath.Dir(gomod), "testdata", "mockagent")
		if _, err := os.Stat(filepath.Join(mockPath, "main.go")); err == nil {
			return mockPath
		}
	}

	t.Fatal("Could not find mock agent source code")
	return ""
}

// findGoMod finds the go.mod file by walking up the directory tree
func findGoMod(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		gomod := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(gomod); err == nil {
			return gomod
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// createAgentWrapper creates a shell wrapper that invokes the mock agent with config
func createAgentWrapper(t *testing.T, binaryPath, name string, config MockAgentConfig) string {
	t.Helper()

	tmpDir := t.TempDir()
	wrapperPath := filepath.Join(tmpDir, name)

	// Create wrapper script
	script := `#!/bin/bash
exec "` + binaryPath + `" \
    -mode="` + string(config.Mode) + `" \
    -initial-context=` + formatFloat(config.InitialContext) + ` \
    -context-growth=` + formatFloat(config.ContextGrowth) + ` \
    -delay=` + formatInt(config.ResponseDelay) + ` \
    -error-msg="` + config.ErrorMessage + `" \
    "$@"
`

	if err := os.WriteFile(wrapperPath, []byte(script), 0755); err != nil {
		t.Fatalf("Failed to create agent wrapper: %v", err)
	}

	return wrapperPath
}

// createMockPattern creates a CLI pattern for the mock agent
func createMockPattern(name string) agent.CLIPattern {
	return agent.CLIPattern{
		Binary:             name,
		VersionArgs:        []string{"--version"},
		AuthCheckCmd:       []string{"auth"},
		NonInteractiveArgs: []string{"-p"},
		JSONOutputArgs:     []string{},
		SkipApprovalsArgs:  []string{},
	}
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%.4f", f)
}

func formatInt(i int) string {
	return fmt.Sprintf("%d", i)
}
