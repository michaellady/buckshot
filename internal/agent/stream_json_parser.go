package agent

import (
	"encoding/json"
	"strings"
)

// StreamJSONParser parses Claude Code-compatible stream-json output format.
// Used by: Claude, Cursor-agent, Amp (all use compatible formats).
type StreamJSONParser struct{}

// Parse transforms stream-json output into readable text.
func (p *StreamJSONParser) Parse(output string) string {
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
			if result.Len() > 0 {
				result.WriteString("\n")
			}
			result.WriteString(extracted)
		}
	}

	if result.Len() == 0 {
		return output
	}

	return result.String()
}

// extractFromLine extracts readable content from a single JSON line.
func (p *StreamJSONParser) extractFromLine(line string) string {
	var event map[string]interface{}
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		return ""
	}

	eventType, _ := event["type"].(string)

	switch eventType {
	case "assistant":
		return p.extractFromAssistant(event)
	case "result":
		return p.extractFromResult(event)
	}

	return ""
}

// extractFromAssistant extracts content from an assistant message event.
func (p *StreamJSONParser) extractFromAssistant(event map[string]interface{}) string {
	message, ok := event["message"].(map[string]interface{})
	if !ok {
		return ""
	}

	content, ok := message["content"].([]interface{})
	if !ok {
		return ""
	}

	var parts []string
	for _, c := range content {
		contentBlock, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		blockType, _ := contentBlock["type"].(string)
		if blockType == "text" {
			if text, ok := contentBlock["text"].(string); ok && text != "" {
				parts = append(parts, text)
			}
		}
	}

	return strings.Join(parts, "\n")
}

// extractFromResult extracts content from a result event.
func (p *StreamJSONParser) extractFromResult(event map[string]interface{}) string {
	// Check for error first
	if isError, _ := event["is_error"].(bool); isError {
		if errMsg, ok := event["error"].(string); ok && errMsg != "" {
			return errMsg
		}
	}

	// Extract result text
	if result, ok := event["result"].(string); ok && result != "" {
		return result
	}

	return ""
}

// ClaudeParser parses Claude Code stream-json output.
type ClaudeParser struct {
	StreamJSONParser
}

// CursorParser parses Cursor-agent stream-json output (Claude-compatible).
type CursorParser struct {
	StreamJSONParser
}

// AmpParser parses Amp stream-json output (Claude-compatible).
type AmpParser struct {
	StreamJSONParser
}
