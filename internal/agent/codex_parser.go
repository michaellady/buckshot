package agent

import (
	"encoding/json"
	"strings"
)

// CodexParser parses Codex JSON streaming output into clean text.
type CodexParser struct{}

// Parse transforms Codex JSONL output into readable text.
func (p *CodexParser) Parse(output string) string {
	if output == "" || strings.TrimSpace(output) == "" {
		return output
	}

	var result strings.Builder
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try to parse as JSON
		if !strings.HasPrefix(line, "{") {
			continue // Skip non-JSON lines
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
func (p *CodexParser) extractFromLine(line string) string {
	var event map[string]interface{}
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		return ""
	}

	eventType, _ := event["type"].(string)

	switch eventType {
	case "item":
		return p.extractFromItem(event)
	case "aggregated_output":
		if output, ok := event["output"].(string); ok {
			return output
		}
	}

	return ""
}

// extractFromItem extracts content from an item event.
func (p *CodexParser) extractFromItem(event map[string]interface{}) string {
	item, ok := event["item"].(map[string]interface{})
	if !ok {
		return ""
	}

	itemType, _ := item["type"].(string)

	switch itemType {
	case "message":
		return p.extractFromMessage(item)
	case "function_call_output":
		if output, ok := item["output"].(string); ok {
			return output
		}
	}

	return ""
}

// extractFromMessage extracts text from a message item.
func (p *CodexParser) extractFromMessage(item map[string]interface{}) string {
	content, ok := item["content"].([]interface{})
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

		switch blockType {
		case "text":
			if text, ok := contentBlock["text"].(string); ok && text != "" {
				parts = append(parts, text)
			}
		case "thinking":
			if thinking, ok := contentBlock["thinking"].(string); ok && thinking != "" {
				parts = append(parts, thinking)
			}
		}
	}

	return strings.Join(parts, "\n")
}
