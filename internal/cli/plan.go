package cli

import (
	"fmt"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt := args[0]
		fmt.Printf("Planning: %s\n", prompt)
		fmt.Printf("Rounds: %d, Agents path: %s\n", rounds, agentsPath)
		// TODO: Implement planning protocol
		return nil
	},
}

func init() {
	planCmd.Flags().IntVarP(&rounds, "rounds", "r", 3, "Number of planning rounds")
	planCmd.Flags().StringVarP(&agentsPath, "agents-path", "a", "", "Path to AGENTS.md file")
	planCmd.Flags().StringSliceVar(&selectedAgents, "agents", nil, "Specific agents to use (default: all available)")
	planCmd.Flags().BoolVar(&untilConverged, "until-converged", false, "Run until all agents report no changes")
}
