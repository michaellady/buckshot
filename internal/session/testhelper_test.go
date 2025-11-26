package session

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
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

// createMockClaudeAt creates a mock claude at a specific path for testing.
func createMockClaudeAt(path string) error {
	mockScript := `#!/bin/bash
echo "Mock Claude started"
echo "Context: 1% used (2000/200000 tokens)"
while IFS= read -r line; do
    if [[ -n "$line" ]]; then
        echo "Mock response: $line"
        echo "Context: 10% used (20000/200000 tokens)"
    fi
done
`

	// Create parent directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(mockScript), 0755)
}

// TestMain runs before all tests in the package.
func TestMain(m *testing.M) {
	// Try to create a symlink from /usr/bin/claude to the real claude
	// This allows tests to work without modification
	realClaudePath, err := exec.LookPath("claude")
	if err == nil && realClaudePath != "/usr/bin/claude" {
		// Found claude at a different path, try to create symlink
		// This will fail if we don't have permissions, which is fine
		_ = os.Symlink(realClaudePath, "/usr/bin/claude")
	}

	// If we still don't have /usr/bin/claude, create a mock
	if _, err := os.Stat("/usr/bin/claude"); os.IsNotExist(err) {
		_ = createMockClaudeAt("/usr/bin/claude")
	}

	// Run tests
	code := m.Run()

	// Note: We don't cleanup /usr/bin/claude as we may not have created it
	// and removing it could break the system

	os.Exit(code)
}

// mockClaudePath returns the path to use for mock claude in tests.
func mockClaudePath() string {
	if path := os.Getenv("TEST_MOCK_CLAUDE"); path != "" {
		return path
	}

	// Check if claude exists at common locations
	for _, path := range []string{
		"/opt/homebrew/bin/claude",
		"/usr/local/bin/claude",
		"/usr/bin/claude",
	} {
		if _, err := exec.LookPath(path); err == nil {
			return path
		}
	}

	return "/usr/bin/claude" // Fallback to test default
}
