package cli

import (
	"fmt"

	"github.com/michaellady/buckshot/internal/agent"
	"github.com/michaellady/buckshot/internal/convergence"
	buckctx "github.com/michaellady/buckshot/internal/context"
	"github.com/michaellady/buckshot/internal/orchestrator"
	"github.com/michaellady/buckshot/internal/session"
	"github.com/spf13/cobra"
)

var (
	rounds         int
	agentsPath     string
	selectedAgents []string
	untilConverged bool
)

var planCmd = &cobra.Command{
	Use:   "plan [prompt]",
	Short: "Run multi-agent planning protocol",
	Long: `Run the multi-agent planning protocol on a given prompt.

Each round, all available agents take turns analyzing the prompt and current
beads state, creating/modifying/reorganizing the plan. Agents persist across
rounds if their context usage stays below 50%.

The protocol continues for the specified number of rounds or until all agents
report no further changes (convergence).`,
	Args: cobra.ExactArgs(1),
	RunE: runPlan,
}

func runPlan(cmd *cobra.Command, args []string) error {
	prompt := args[0]
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "Planning: %s\n", prompt)
	fmt.Fprintf(out, "Rounds: %d, Agents path: %s\n", rounds, agentsPath)

	// Detect available agents
	detector := agent.NewDetector()
	agents, err := detector.DetectAll()
	if err != nil {
		return fmt.Errorf("failed to detect agents: %w", err)
	}

	// Filter to selected agents if specified
	if len(selectedAgents) > 0 {
		agents = filterAgents(agents, selectedAgents)
	}

	// Filter to authenticated agents only
	var authAgents []agent.Agent
	for _, a := range agents {
		if a.Authenticated {
			authAgents = append(authAgents, a)
		}
	}

	if len(authAgents) == 0 {
		fmt.Fprintf(out, "No authenticated agents available\n")
		return nil
	}

	fmt.Fprintf(out, "Using %d agent(s): ", len(authAgents))
	for i, a := range authAgents {
		if i > 0 {
			fmt.Fprintf(out, ", ")
		}
		fmt.Fprintf(out, "%s", a.Name)
	}
	fmt.Fprintf(out, "\n")

	// Set up orchestrator
	orch := orchestrator.NewRoundOrchestrator()
	orch.SetSessionManager(session.NewManager())
	orch.SetContextBuilder(buckctx.NewBuilder())

	// Set up convergence detector
	convDetector := convergence.NewDetector()

	// Build initial planning context
	builder := buckctx.NewBuilder()
	planCtx, err := builder.Build(prompt, agentsPath, 1, true)
	if err != nil {
		return fmt.Errorf("failed to build planning context: %w", err)
	}

	// Run rounds
	maxRounds := rounds
	if untilConverged {
		maxRounds = 100 // Safety limit
	}

	for round := 1; round <= maxRounds; round++ {
		fmt.Fprintf(out, "\n=== Round %d ===\n", round)

		planCtx.Round = round
		planCtx.IsFirstTurn = (round == 1)

		result, err := orch.RunRound(cmd.Context(), authAgents, planCtx)
		if err != nil {
			return fmt.Errorf("round %d failed: %w", round, err)
		}

		// Report results
		fmt.Fprintf(out, "Changes: %d, Failed: %d, Skipped: %d\n",
			result.TotalChanges, result.FailedCount, result.SkippedCount)

		// Check convergence
		if untilConverged && convDetector.CheckConvergence(result) {
			fmt.Fprintf(out, "\nConverged after %d round(s)\n", round)
			break
		}

		if !untilConverged && round >= rounds {
			fmt.Fprintf(out, "\nCompleted %d round(s)\n", rounds)
			break
		}
	}

	fmt.Fprintf(out, "\nPlanning complete.\n")
	return nil
}

// filterAgents returns only agents whose names are in the selected list
func filterAgents(agents []agent.Agent, selected []string) []agent.Agent {
	selectedSet := make(map[string]bool)
	for _, name := range selected {
		selectedSet[name] = true
	}

	var filtered []agent.Agent
	for _, a := range agents {
		if selectedSet[a.Name] {
			filtered = append(filtered, a)
		}
	}
	return filtered
}

func init() {
	planCmd.Flags().IntVarP(&rounds, "rounds", "r", 3, "Number of planning rounds")
	planCmd.Flags().StringVarP(&agentsPath, "agents-path", "a", "", "Path to AGENTS.md file")
	planCmd.Flags().StringSliceVar(&selectedAgents, "agents", nil, "Specific agents to use (default: all available)")
	planCmd.Flags().BoolVar(&untilConverged, "until-converged", false, "Run until all agents report no changes")
}
