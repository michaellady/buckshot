package agent

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DefaultDetector is the default implementation of Detector.
type DefaultDetector struct {
	searchPath string
}

// NewDetector creates a new detector using the system PATH.
func NewDetector() *DefaultDetector {
	return &DefaultDetector{
		searchPath: os.Getenv("PATH"),
	}
}

// NewDetectorWithPath creates a detector with a custom search path.
func NewDetectorWithPath(path string) *DefaultDetector {
	return &DefaultDetector{searchPath: path}
}

// DetectAll returns all available agents on the system.
func (d *DefaultDetector) DetectAll() ([]Agent, error) {
	agents := []Agent{}
	knownAgents := KnownAgents()

	for name, pattern := range knownAgents {
		if d.IsInstalled(name) {
			agent := Agent{
				Name:    name,
				Path:    d.GetAgentPath(name),
				Pattern: pattern,
				Parser:  GetParserForAgent(name),
			}

			// Get version
			agent.Version = d.getVersion(agent)

			// Check authentication
			agent.Authenticated = d.IsAuthenticated(agent)

			agents = append(agents, agent)
		}
	}

	return agents, nil
}

// GetParserForAgent returns the appropriate output parser for a given agent.
func GetParserForAgent(name string) OutputParser {
	switch name {
	case "claude":
		return &ClaudeParser{}
	case "codex":
		return &CodexParser{}
	case "cursor-agent":
		return &CursorParser{}
	case "auggie":
		return &AuggieParser{}
	case "gemini":
		return &GeminiParser{}
	case "amp":
		return &AmpParser{}
	default:
		return &NoopParser{}
	}
}

// IsInstalled checks if a specific agent is installed.
func (d *DefaultDetector) IsInstalled(name string) bool {
	return d.GetAgentPath(name) != ""
}

// IsAuthenticated checks if an agent is authenticated.
func (d *DefaultDetector) IsAuthenticated(agent Agent) bool {
	if agent.Path == "" {
		return false
	}

	// For most agents, if they're installed and version works, assume authenticated
	// Real auth check would require running a command that hits the API
	pattern, ok := KnownAgents()[agent.Name]
	if !ok {
		return false
	}

	// Try running the auth check command
	var cmd *exec.Cmd
	if len(pattern.AuthCheckCmd) > 0 {
		cmd = exec.Command(agent.Path, pattern.AuthCheckCmd...)
	} else {
		// Fall back to version check
		cmd = exec.Command(agent.Path, pattern.VersionArgs...)
	}

	err := cmd.Run()
	return err == nil
}

// GetAgentPath returns the full path for an agent binary.
func (d *DefaultDetector) GetAgentPath(name string) string {
	if d.searchPath == "" {
		return ""
	}

	// Check each directory in the search path
	for _, dir := range filepath.SplitList(d.searchPath) {
		path := filepath.Join(dir, name)
		if info, err := os.Stat(path); err == nil {
			// Check if it's executable
			if info.Mode()&0111 != 0 {
				return path
			}
		}
	}

	return ""
}

// getVersion retrieves the version string for an agent.
func (d *DefaultDetector) getVersion(agent Agent) string {
	if agent.Path == "" {
		return ""
	}

	pattern, ok := KnownAgents()[agent.Name]
	if !ok || len(pattern.VersionArgs) == 0 {
		return ""
	}

	cmd := exec.Command(agent.Path, pattern.VersionArgs...)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Return first line of output, trimmed
	version := strings.TrimSpace(string(output))
	if idx := strings.Index(version, "\n"); idx != -1 {
		version = version[:idx]
	}

	return version
}
