package agent

import (
	"encoding/json"
	"strings"
)

// AuggieParser parses Auggie JSON output format.
type AuggieParser struct{}

// Parse transforms Auggie JSON output into readable text.
func (p *AuggieParser) Parse(output string) string {
	if output == "" || strings.TrimSpace(output) == "" {
		return output
	}

	// Auggie outputs a single JSON object (not JSONL)
	output = strings.TrimSpace(output)
	if !strings.HasPrefix(output, "{") {
		return output
	}

	var event map[string]interface{}
	if err := json.Unmarshal([]byte(output), &event); err != nil {
		return output
	}

	eventType, _ := event["type"].(string)

	if eventType == "result" {
		// Check for error first
		if isError, _ := event["is_error"].(bool); isError {
			if errMsg, ok := event["error"].(string); ok && errMsg != "" {
				return errMsg
			}
		}

		// Extract result text
		if result, ok := event["result"].(string); ok {
			return strings.TrimSpace(result)
		}
	}

	return output
}
