package agent

// OutputParser transforms raw agent output into clean text.
type OutputParser interface {
	// Parse transforms the raw output from an agent CLI into clean text.
	Parse(output string) string
}

// NoopParser is an OutputParser that returns input unchanged.
// Use this as the default parser or for agents that don't need parsing.
type NoopParser struct{}

// Parse returns the input unchanged.
func (p *NoopParser) Parse(output string) string {
	return output
}
