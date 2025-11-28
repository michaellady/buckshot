package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	feedbackAgent string
)

var feedbackCmd = &cobra.Command{
	Use:   "feedback",
	Short: "Run single-agent feedback mode (comment-only)",
	Long: `Run a single agent in feedback mode to review and comment on beads.

In feedback mode, agents can only add comments to existing beads - they cannot
create new beads or modify descriptions. This provides a safe way to gather
feedback from different AI agents.

Example:
  buckshot feedback --agent claude --agents-path /path/to/AGENTS.md`,
	RunE: runFeedback,
}

func runFeedback(cmd *cobra.Command, args []string) error {
	// TODO: Implement feedback mode
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Feedback mode not yet implemented")
	return nil
}

func init() {
	feedbackCmd.Flags().StringVar(&feedbackAgent, "agent", "", "Agent to run in feedback mode (required)")
	feedbackCmd.Flags().StringVarP(&agentsPath, "agents-path", "a", "", "Path to AGENTS.md file")
	_ = feedbackCmd.MarkFlagRequired("agent")
}
