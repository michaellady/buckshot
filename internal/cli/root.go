package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "buckshot",
	Short: "Multi-agent planning protocol for AI coding assistants",
	Long: `Buckshot orchestrates multiple AI coding agents (Claude Code, Codex, Cursor)
to collaboratively plan and refine development tasks using beads (bd) for issue tracking.

Each planning round, all available agents analyze the current plan and suggest
improvements until the team converges on a complete solution.`,
}

func Execute(version string) error {
	rootCmd.Version = version
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(planCmd)
	rootCmd.AddCommand(agentsCmd)
}
