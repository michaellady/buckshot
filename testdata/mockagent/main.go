// Package main implements a mock AI agent for testing buckshot.
//
// This mock agent simulates the behavior of real AI coding assistants
// (Claude Code, Codex, etc.) for end-to-end testing purposes.
//
// Usage modes:
//   - Default: Echo responses with simulated context usage
//   - Error: Simulate various error conditions
//   - Timeout: Simulate slow or hanging responses
//   - Conversation: Full conversation mode for integration tests
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// Config controls mock agent behavior
type Config struct {
	Mode           string  // "default", "error", "timeout", "auth_fail", "converged"
	InitialContext float64 // Starting context usage (0.0-1.0)
	ContextGrowth  float64 // Context growth per message
	ResponseDelay  int     // Delay in milliseconds before responding
	ErrorMessage   string  // Custom error message for error mode
	Version        string  // Version string to report
}

var config Config

func main() {
	// Handle --version flag BEFORE parsing (matches real agent behavior)
	// This must be done before flag.Parse() since --version is a boolean-style flag
	for _, arg := range os.Args[1:] {
		if arg == "--version" || arg == "-v" {
			fmt.Println("mock-agent 1.0.0-mock")
			os.Exit(0)
		}
	}

	// Parse flags
	var prompt string
	flag.StringVar(&config.Mode, "mode", "default", "Mock behavior mode")
	flag.Float64Var(&config.InitialContext, "initial-context", 0.01, "Initial context usage (0.0-1.0)")
	flag.Float64Var(&config.ContextGrowth, "context-growth", 0.05, "Context growth per message")
	flag.IntVar(&config.ResponseDelay, "delay", 0, "Response delay in milliseconds")
	flag.StringVar(&config.ErrorMessage, "error-msg", "Mock error occurred", "Error message for error mode")
	flag.StringVar(&config.Version, "mock-version", "1.0.0-mock", "Version string for mock responses")
	flag.StringVar(&prompt, "p", "", "Prompt to process (non-interactive mode)")
	flag.Parse()

	args := flag.Args()

	// Handle auth check (matches real agent behavior)
	// Real agents exit 0 when authenticated, non-0 when not
	for _, arg := range args {
		if arg == "auth" || arg == "--auth-check" {
			if config.Mode == "auth_fail" {
				fmt.Fprintln(os.Stderr, "Not authenticated")
				os.Exit(1)
			}
			fmt.Println("Authenticated")
			os.Exit(0)
		}
	}

	// If we have a prompt, run in one-shot mode
	if prompt != "" {
		handlePrompt(prompt)
		return
	}

	// Otherwise run in conversation mode (reading from stdin)
	runConversationMode()
}

func handlePrompt(prompt string) {
	if config.ResponseDelay > 0 {
		time.Sleep(time.Duration(config.ResponseDelay) * time.Millisecond)
	}

	switch config.Mode {
	case "error":
		fmt.Fprintf(os.Stderr, "Error: %s\n", config.ErrorMessage)
		os.Exit(1)
	case "timeout":
		// Simulate a very long operation
		time.Sleep(10 * time.Minute)
	case "converged":
		// Respond indicating no changes needed
		fmt.Println("I've analyzed the current beads and the plan looks complete. No changes needed.")
		printContextUsage(config.InitialContext)
	default:
		// Normal response with simulated work
		response := generateResponse(prompt)
		fmt.Println(response)
		printContextUsage(config.InitialContext)
	}
}

func runConversationMode() {
	scanner := bufio.NewScanner(os.Stdin)
	contextUsage := config.InitialContext
	messageCount := 0

	// Print initial context
	printContextUsage(contextUsage)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		messageCount++

		if config.ResponseDelay > 0 {
			time.Sleep(time.Duration(config.ResponseDelay) * time.Millisecond)
		}

		switch config.Mode {
		case "error":
			if messageCount >= 2 {
				fmt.Fprintf(os.Stderr, "Error: %s\n", config.ErrorMessage)
				os.Exit(1)
			}
			fmt.Println(generateResponse(line))
		case "converged":
			// After first message, indicate no changes
			if messageCount > 1 {
				fmt.Println("No further changes needed. The plan is complete.")
			} else {
				fmt.Println(generateResponse(line))
			}
		default:
			fmt.Println(generateResponse(line))
		}

		// Update context usage
		contextUsage += config.ContextGrowth
		if contextUsage > 1.0 {
			contextUsage = 1.0
		}
		printContextUsage(contextUsage)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
}

func generateResponse(prompt string) string {
	// Generate a response that looks like what a real agent might produce
	lowerPrompt := strings.ToLower(prompt)

	if strings.Contains(lowerPrompt, "plan") || strings.Contains(lowerPrompt, "beads") {
		return fmt.Sprintf(`I'll analyze the planning task and work with the beads system.

Looking at the prompt: "%s"

I'll execute the following steps:
1. Check current beads with 'bd list'
2. Analyze the requirements
3. Create/update beads as needed

Let me run: bd create "Implement core functionality" -t task -p 1 -d "Based on planning analysis"

The bead has been created. I've also identified some follow-up tasks that should be tracked.`, truncate(prompt, 50))
	}

	if strings.Contains(lowerPrompt, "test") {
		return fmt.Sprintf(`I'll help with testing.

Running: go test -v ./...

All tests passed. Coverage looks good.

Response to: %s`, truncate(prompt, 50))
	}

	return fmt.Sprintf(`Analyzed the request: "%s"

I've completed the requested task. The work has been documented and any relevant beads have been updated.`, truncate(prompt, 50))
}

func printContextUsage(usage float64) {
	// Match the format that real agents use
	usedTokens := int(usage * 200000)
	fmt.Printf("\nContext: %.0f%% used (%d/200000 tokens)\n", usage*100, usedTokens)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// JSONResponse represents a structured response format (for Codex-style agents)
type JSONResponse struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	Message string `json:"message,omitempty"`
}

// OutputJSONMode outputs responses in JSON format (for testing Codex parser)
func outputJSONMode(prompt string) {
	responses := []JSONResponse{
		{Type: "message", Message: "Analyzing request..."},
		{Type: "message", Message: generateResponse(prompt)},
	}

	for _, r := range responses {
		data, _ := json.Marshal(r)
		fmt.Println(string(data))
	}
}
