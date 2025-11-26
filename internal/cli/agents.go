package cli

import (
	"fmt"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Detecting available agents...")
		// TODO: Implement agent detection
		return nil
	},
}
