package agent

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestDetectorInterface ensures Detector interface is properly defined
func TestDetectorInterface(t *testing.T) {
	// This test verifies the interface exists and can be implemented
	var _ Detector = (*DefaultDetector)(nil)
}

// TestDetectAll tests detection of all available agents
func TestDetectAll(t *testing.T) {
	d := NewDetector()
	agents, err := d.DetectAll()
	if err != nil {
		t.Fatalf("DetectAll() error = %v", err)
	}

	// Should return a slice (possibly empty if no agents installed)
	if agents == nil {
		t.Error("DetectAll() returned nil, want empty slice")
	}

	// Each detected agent should have required fields
	for _, agent := range agents {
		if agent.Name == "" {
			t.Error("Agent.Name is empty")
		}
		if agent.Path == "" {
			t.Error("Agent.Path is empty")
		}
	}
}

// TestIsInstalled tests detection of specific agent installation
func TestIsInstalled(t *testing.T) {
	d := NewDetector()

	tests := []struct {
		name      string
		agentName string
	}{
		{"claude", "claude"},
		{"codex", "codex"},
		{"cursor-agent", "cursor-agent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// IsInstalled should return bool without error
			installed := d.IsInstalled(tt.agentName)

			// Verify against actual system state
			_, err := exec.LookPath(tt.agentName)
			expectedInstalled := err == nil

			if installed != expectedInstalled {
				t.Errorf("IsInstalled(%q) = %v, want %v", tt.agentName, installed, expectedInstalled)
			}
		})
	}
}

// TestIsInstalledWithMockPath tests detection with a mock binary
func TestIsInstalledWithMockPath(t *testing.T) {
	// Create a temporary directory with a mock binary
	tmpDir, err := os.MkdirTemp("", "buckshot-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a mock "claude" binary
	mockBinary := filepath.Join(tmpDir, "claude")
	if err := os.WriteFile(mockBinary, []byte("#!/bin/sh\necho 'mock'"), 0755); err != nil {
		t.Fatalf("Failed to create mock binary: %v", err)
	}

	// Create detector with custom PATH
	d := NewDetectorWithPath(tmpDir)

	if !d.IsInstalled("claude") {
		t.Error("IsInstalled('claude') = false, want true (mock binary exists)")
	}

	if d.IsInstalled("nonexistent") {
		t.Error("IsInstalled('nonexistent') = true, want false")
	}
}

// TestIsAuthenticated tests authentication checking
func TestIsAuthenticated(t *testing.T) {
	d := NewDetector()

	// Create a test agent
	agent := Agent{
		Name: "claude",
		Path: "/usr/bin/claude",
	}

	// IsAuthenticated should return a boolean without panicking
	// Actual auth status depends on system state
	_ = d.IsAuthenticated(agent)
}

// TestDetectAllReturnsKnownAgents tests that only known agents are detected
func TestDetectAllReturnsKnownAgents(t *testing.T) {
	d := NewDetector()
	agents, err := d.DetectAll()
	if err != nil {
		t.Fatalf("DetectAll() error = %v", err)
	}

	knownNames := map[string]bool{
		"claude":       true,
		"codex":        true,
		"cursor-agent": true,
	}

	for _, agent := range agents {
		if !knownNames[agent.Name] {
			t.Errorf("DetectAll() returned unknown agent %q", agent.Name)
		}
	}
}

// TestAgentHasVersion tests that detected agents have version info
func TestAgentHasVersion(t *testing.T) {
	d := NewDetector()
	agents, err := d.DetectAll()
	if err != nil {
		t.Fatalf("DetectAll() error = %v", err)
	}

	for _, agent := range agents {
		// Installed agents should have version info
		if agent.Version == "" {
			t.Errorf("Agent %q has empty Version", agent.Name)
		}
	}
}

// TestDetectorWithEmptyPath tests detector behavior with no PATH
func TestDetectorWithEmptyPath(t *testing.T) {
	d := NewDetectorWithPath("")
	agents, err := d.DetectAll()
	if err != nil {
		t.Fatalf("DetectAll() error = %v", err)
	}

	// With empty PATH, no agents should be found
	if len(agents) != 0 {
		t.Errorf("DetectAll() with empty PATH returned %d agents, want 0", len(agents))
	}
}

// TestGetAgentPath tests getting the full path for an agent
func TestGetAgentPath(t *testing.T) {
	d := NewDetector()

	// For installed agents, GetAgentPath should return the full path
	if d.IsInstalled("claude") {
		path := d.GetAgentPath("claude")
		if path == "" {
			t.Error("GetAgentPath('claude') returned empty string for installed agent")
		}
		if !filepath.IsAbs(path) {
			t.Errorf("GetAgentPath('claude') = %q, want absolute path", path)
		}
	}

	// For non-installed agents, GetAgentPath should return empty string
	path := d.GetAgentPath("nonexistent-agent")
	if path != "" {
		t.Errorf("GetAgentPath('nonexistent-agent') = %q, want empty string", path)
	}
}
