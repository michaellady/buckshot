package agent

import (
	"encoding/json"
	"strings"
)

// GeminiParser parses Gemini CLI stream-json output format.
type GeminiParser struct{}

// Parse transforms Gemini stream-json output into readable text.
func (p *GeminiParser) Parse(output string) string {
	if output == "" || strings.TrimSpace(output) == "" {
		return output
	}

	var result strings.Builder
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "{") {
			continue
		}

		extracted := p.extractFromLine(line)
		if extracted != "" {
			// For Gemini deltas, concatenate without newlines
			result.WriteString(extracted)
		}
	}

	if result.Len() == 0 {
		return output
	}

	return result.String()
}

// extractFromLine extracts readable content from a single JSON line.
func (p *GeminiParser) extractFromLine(line string) string {
	var event map[string]interface{}
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		return ""
	}

	eventType, _ := event["type"].(string)

	if eventType == "message" {
		role, _ := event["role"].(string)
		if role == "assistant" {
			if content, ok := event["content"].(string); ok {
				return content
			}
		}
	}

	return ""
}
