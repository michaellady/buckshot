package session

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/michaellady/buckshot/internal/agent"
)

// setupMockClaude creates a mock claude binary for testing.
// This is called once before tests run.
func setupMockClaude(t *testing.T) string {
	t.Helper()

	// Create a temporary mock claude script
	mockScript := `#!/bin/bash
# Mock claude for session testing
echo "Mock Claude started"
echo "Context: 1% used (2000/200000 tokens)"

# Read commands from stdin and respond
while IFS= read -r line; do
    if [[ -n "$line" ]]; then
        echo "Mock response to: $line"
        # Simulate increasing context usage
        echo "Context: 10% used (20000/200000 tokens)"
    fi
done
`

	tmpDir := t.TempDir()
	mockPath := filepath.Join(tmpDir, "mock-claude")

	if err := os.WriteFile(mockPath, []byte(mockScript), 0755); err != nil {
		t.Fatalf("Failed to create mock claude: %v", err)
	}

	return mockPath
}

// newTestAgent creates an agent.Agent for testing with the correct path.
// It detects the actual claude binary location on the system.
func newTestAgent() agent.Agent {
	return agent.Agent{
		Name:          "claude",
		Path:          mockClaudePath(),
		Authenticated: true,
		Version:       "1.0.0",
		Pattern:       agent.KnownAgents()["claude"],
	}
}

// newUnauthenticatedTestAgent creates an unauthenticated agent for testing.
func newUnauthenticatedTestAgent() agent.Agent {
	a := newTestAgent()
	a.Authenticated = false
	return a
}

// mockClaudePath returns the path to use for mock claude in tests.
func mockClaudePath() string {
	if path := os.Getenv("TEST_MOCK_CLAUDE"); path != "" {
		return path
	}

	// First try exec.LookPath to find claude in PATH
	if path, err := exec.LookPath("claude"); err == nil {
		return path
	}

	// Check if claude exists at common locations
	for _, path := range []string{
		"/opt/homebrew/bin/claude",
		"/usr/local/bin/claude",
		"/usr/bin/claude",
	} {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "/usr/bin/claude" // Fallback to test default
}
