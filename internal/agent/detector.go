package agent

// DefaultDetector is the default implementation of Detector.
// TODO: Implement in buckshot-vk6
type DefaultDetector struct {
	searchPath string
}

// NewDetector creates a new detector using the system PATH.
func NewDetector() *DefaultDetector {
	// TODO: Implement in buckshot-vk6
	return &DefaultDetector{}
}

// NewDetectorWithPath creates a detector with a custom search path.
func NewDetectorWithPath(path string) *DefaultDetector {
	// TODO: Implement in buckshot-vk6
	return &DefaultDetector{searchPath: path}
}

// DetectAll returns all available agents on the system.
func (d *DefaultDetector) DetectAll() ([]Agent, error) {
	// TODO: Implement in buckshot-vk6
	// This stub returns empty slice - tests will fail
	return []Agent{}, nil
}

// IsInstalled checks if a specific agent is installed.
func (d *DefaultDetector) IsInstalled(name string) bool {
	// TODO: Implement in buckshot-vk6
	// This stub returns false - tests will fail
	return false
}

// IsAuthenticated checks if an agent is authenticated.
func (d *DefaultDetector) IsAuthenticated(agent Agent) bool {
	// TODO: Implement in buckshot-vk6
	// This stub returns false - tests will fail
	return false
}

// GetAgentPath returns the full path for an agent binary.
func (d *DefaultDetector) GetAgentPath(name string) string {
	// TODO: Implement in buckshot-vk6
	// This stub returns empty string - tests will fail
	return ""
}
