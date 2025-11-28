package cli

import (
	"fmt"

	"github.com/michaellady/buckshot/internal/agent"
	buckctx "github.com/michaellady/buckshot/internal/context"
	"github.com/michaellady/buckshot/internal/session"
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
	out := cmd.OutOrStdout()

	_, _ = fmt.Fprintf(out, "Feedback mode: %s\n", feedbackAgent)

	// Detect available agents
	agents, err := agentDetector()
	if err != nil {
		return fmt.Errorf("failed to detect agents: %w", err)
	}

	// Find the requested agent
	var targetAgent *agent.Agent
	for i, a := range agents {
		if a.Name == feedbackAgent {
			targetAgent = &agents[i]
			break
		}
	}

	if targetAgent == nil {
		return fmt.Errorf("agent %q not found", feedbackAgent)
	}

	if !targetAgent.Authenticated {
		return fmt.Errorf("agent %q is not authenticated", feedbackAgent)
	}

	_, _ = fmt.Fprintf(out, "Using agent: %s\n", targetAgent.Name)

	// Build feedback context
	builder := buckctx.NewBuilder()
	planCtx, err := builder.Build("", agentsPath, 1, true)
	if err != nil {
		return fmt.Errorf("failed to build context: %w", err)
	}

	// Set feedback mode fields
	planCtx.FeedbackMode = true
	planCtx.AgentName = targetAgent.Name

	// Format the feedback prompt
	prompt := builder.FormatFeedback(planCtx)

	_, _ = fmt.Fprintf(out, "Running %s in one-shot mode...\n", targetAgent.Name)

	// Use RunOneShot for one-shot execution (waits for process exit)
	result, err := session.RunOneShot(cmd.Context(), *targetAgent, prompt)
	if err != nil {
		// Still show output even if there was an error
		if result.Output != "" {
			_, _ = fmt.Fprintf(out, "\n=== %s Response ===\n", targetAgent.Name)
			_, _ = fmt.Fprintln(out, result.Output)
		}
		return fmt.Errorf("agent %s failed (exit code %d): %w", targetAgent.Name, result.ExitCode, err)
	}

	_, _ = fmt.Fprintf(out, "\n=== %s Response ===\n", targetAgent.Name)
	_, _ = fmt.Fprintln(out, result.Output)

	_, _ = fmt.Fprintf(out, "\nFeedback complete.\n")
	return nil
}

func init() {
	feedbackCmd.Flags().StringVar(&feedbackAgent, "agent", "", "Agent to run in feedback mode (required)")
	feedbackCmd.Flags().StringVarP(&agentsPath, "agents-path", "a", "", "Path to AGENTS.md file")
	_ = feedbackCmd.MarkFlagRequired("agent")
}
