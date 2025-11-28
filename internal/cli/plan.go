package cli

import (
	"fmt"
	"io"
	"time"

	"github.com/michaellady/buckshot/internal/agent"
	"github.com/michaellady/buckshot/internal/convergence"
	buckctx "github.com/michaellady/buckshot/internal/context"
	"github.com/michaellady/buckshot/internal/notes"
	"github.com/michaellady/buckshot/internal/orchestrator"
	"github.com/michaellady/buckshot/internal/session"
	"github.com/spf13/cobra"
)

var (
	rounds         int
	agentsPath     string
	selectedAgents []string
	untilConverged bool
	saveToBead     string
	verbose        bool
)

// terminalProgressReporter implements orchestrator.ProgressReporter for terminal output.
type terminalProgressReporter struct {
	out       io.Writer
	startTime time.Time
}

func newTerminalProgressReporter(out io.Writer) *terminalProgressReporter {
	return &terminalProgressReporter{out: out}
}

func (r *terminalProgressReporter) OnAgentStart(round, agentIndex, totalAgents int, ag agent.Agent) {
	r.startTime = time.Now()
	_, _ = fmt.Fprintf(r.out, "\n  [Round %d] Agent %d/%d: %s - STARTED\n", round, agentIndex, totalAgents, ag.Name)
}

func (r *terminalProgressReporter) OnAgentComplete(round, agentIndex, totalAgents int, result orchestrator.AgentResult, beadsDiff string) {
	elapsed := time.Since(r.startTime)
	status := "COMPLETED"
	if result.Error != nil {
		status = fmt.Sprintf("FAILED: %v", result.Error)
	} else if result.Skipped {
		status = "SKIPPED"
	}
	_, _ = fmt.Fprintf(r.out, "  [Round %d] Agent %d/%d: %s - %s (%.1fs)\n", round, agentIndex, totalAgents, result.Agent.Name, status, elapsed.Seconds())
	if beadsDiff != "" && beadsDiff != "(no changes)" && !result.Skipped {
		_, _ = fmt.Fprintf(r.out, "  Beads diff:\n")
		// Indent the diff output
		for _, line := range splitDiffLines(beadsDiff) {
			if line != "" {
				_, _ = fmt.Fprintf(r.out, "    %s\n", line)
			}
		}
	}
}

func splitDiffLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// agentDetector is the function used to detect agents.
// It can be overridden in tests to inject mock agents.
var agentDetector = defaultAgentDetector

// defaultAgentDetector returns agents using the standard detector
func defaultAgentDetector() ([]agent.Agent, error) {
	detector := agent.NewDetector()
	return detector.DetectAll()
}

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

	_, _ = fmt.Fprintf(out, "Planning: %s\n", prompt)
	_, _ = fmt.Fprintf(out, "Rounds: %d, Agents path: %s\n", rounds, agentsPath)

	// Detect available agents (uses agentDetector which can be overridden in tests)
	agents, err := agentDetector()
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
		_, _ = fmt.Fprintf(out, "No authenticated agents available\n")
		return nil
	}

	_, _ = fmt.Fprintf(out, "Using %d agent(s): ", len(authAgents))
	for i, a := range authAgents {
		if i > 0 {
			_, _ = fmt.Fprintf(out, ", ")
		}
		_, _ = fmt.Fprintf(out, "%s", a.Name)
	}
	_, _ = fmt.Fprintf(out, "\n")

	// Set up orchestrator
	orch := orchestrator.NewRoundOrchestrator()
	orch.SetSessionManager(session.NewManager())
	orch.SetContextBuilder(buckctx.NewBuilder())

	// Set up progress reporter if verbose mode is enabled
	if verbose {
		orch.SetProgressReporter(newTerminalProgressReporter(out))
	}

	// Set up convergence detector
	convDetector := convergence.NewDetector()

	// Set up notes saver if --save flag is set
	var noteSaver notes.Saver
	if saveToBead != "" {
		noteSaver = notes.NewSaver()
		_, _ = fmt.Fprintf(out, "Saving perspectives to: %s\n", saveToBead)
	}

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
		_, _ = fmt.Fprintf(out, "\n=== Round %d ===\n", round)

		planCtx.Round = round
		planCtx.IsFirstTurn = (round == 1)

		result, err := orch.RunRound(cmd.Context(), authAgents, planCtx)
		if err != nil {
			return fmt.Errorf("round %d failed: %w", round, err)
		}

		// Report results
		_, _ = fmt.Fprintf(out, "Changes: %d, Failed: %d, Skipped: %d\n",
			result.TotalChanges, result.FailedCount, result.SkippedCount)

		// Save perspectives to bead if --save flag is set
		if noteSaver != nil {
			if err := noteSaver.SaveRoundResults(cmd.Context(), saveToBead, result); err != nil {
				_, _ = fmt.Fprintf(out, "Warning: failed to save perspectives: %v\n", err)
			} else {
				_, _ = fmt.Fprintf(out, "Saved round %d perspectives to %s\n", round, saveToBead)
			}
		}

		// Check convergence
		if untilConverged && convDetector.CheckConvergence(result) {
			_, _ = fmt.Fprintf(out, "\nConverged after %d round(s)\n", round)
			break
		}

		if !untilConverged && round >= rounds {
			_, _ = fmt.Fprintf(out, "\nCompleted %d round(s)\n", rounds)
			break
		}
	}

	_, _ = fmt.Fprintf(out, "\nPlanning complete.\n")
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
	planCmd.Flags().StringVar(&saveToBead, "save", "", "Save agent perspectives to specified bead ID")
	planCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed progress with agent timing and beads diff")
}
