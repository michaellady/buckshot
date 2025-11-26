// Package agent provides detection and management of AI coding agents.
package agent

// Agent represents a detected AI coding agent CLI tool.
type Agent struct {
	Name          string     // e.g., "claude", "codex", "cursor-agent"
	Path          string     // Full path to the binary
	Authenticated bool       // Whether the agent is authenticated
	Version       string     // Agent version if available
	Pattern       CLIPattern // CLI invocation pattern for this agent
}

// Detector finds and validates available AI agents.
type Detector interface {
	// DetectAll returns all available agents on the system.
	DetectAll() ([]Agent, error)

	// IsInstalled checks if a specific agent is installed.
	IsInstalled(name string) bool

	// IsAuthenticated checks if an agent is authenticated.
	IsAuthenticated(agent Agent) bool
}
