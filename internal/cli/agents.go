package cli

import (
	"fmt"

	"github.com/michaellady/buckshot/internal/agent"
	"github.com/spf13/cobra"
)

var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "List available AI coding agents",
	Long: `List all detected AI coding agents and their status.

Buckshot looks for the following CLI tools:
  - claude (Claude Code)
  - codex (OpenAI Codex CLI)
  - cursor-agent (Cursor Agent)
  - auggie (Augment Code)
  - gemini (Google Gemini CLI)
  - amp (Amp CLI)

Each agent is checked for installation and authentication status.`,
	RunE: runAgents,
}

func runAgents(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	_, _ = fmt.Fprintf(out, "Detecting available agents...\n\n")

	detector := agent.NewDetector()
	agents, err := detector.DetectAll()
	if err != nil {
		return fmt.Errorf("failed to detect agents: %w", err)
	}

	if len(agents) == 0 {
		_, _ = fmt.Fprintf(out, "No agents found.\n")
		_, _ = fmt.Fprintf(out, "\nSupported agents:\n")
		for name := range agent.KnownAgents() {
			_, _ = fmt.Fprintf(out, "  - %s\n", name)
		}
		return nil
	}

	_, _ = fmt.Fprintf(out, "Found %d agent(s):\n\n", len(agents))

	for _, a := range agents {
		status := "✗ not authenticated"
		if a.Authenticated {
			status = "✓ ready"
		}

		_, _ = fmt.Fprintf(out, "  %s\n", a.Name)
		_, _ = fmt.Fprintf(out, "    Path: %s\n", a.Path)
		_, _ = fmt.Fprintf(out, "    Version: %s\n", a.Version)
		_, _ = fmt.Fprintf(out, "    Status: %s\n", status)
		_, _ = fmt.Fprintf(out, "\n")
	}

	return nil
}
