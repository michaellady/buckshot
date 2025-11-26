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

Each agent is checked for installation and authentication status.`,
	RunE: runAgents,
}

func runAgents(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "Detecting available agents...\n\n")

	detector := agent.NewDetector()
	agents, err := detector.DetectAll()
	if err != nil {
		return fmt.Errorf("failed to detect agents: %w", err)
	}

	if len(agents) == 0 {
		fmt.Fprintf(out, "No agents found.\n")
		fmt.Fprintf(out, "\nSupported agents:\n")
		for name := range agent.KnownAgents() {
			fmt.Fprintf(out, "  - %s\n", name)
		}
		return nil
	}

	fmt.Fprintf(out, "Found %d agent(s):\n\n", len(agents))

	for _, a := range agents {
		status := "✗ not authenticated"
		if a.Authenticated {
			status = "✓ ready"
		}

		fmt.Fprintf(out, "  %s\n", a.Name)
		fmt.Fprintf(out, "    Path: %s\n", a.Path)
		fmt.Fprintf(out, "    Version: %s\n", a.Version)
		fmt.Fprintf(out, "    Status: %s\n", status)
		fmt.Fprintf(out, "\n")
	}

	return nil
}
