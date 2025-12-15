package session

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/michaellady/buckshot/internal/agent"
)

// OneShotResult represents the result of a one-shot agent execution.
type OneShotResult struct {
	Output   string // Combined stdout/stderr output
	ExitCode int    // Process exit code
	Error    error  // Any error during execution
}

// RunOneShot executes an agent in one-shot mode and waits for completion.
// This is used for agents that run a single prompt and exit (auggie --print,
// amp --execute, gemini positional, codex exec).
//
// Unlike interactive sessions, one-shot execution:
// - Builds command with prompt as argument
// - Runs synchronously until process exits
// - Captures all output
// - Returns when process completes
func RunOneShot(ctx context.Context, ag agent.Agent, prompt string) (OneShotResult, error) {
	// Build command arguments
	args := buildOneShotArgs(ag.Pattern, prompt)

	// Create command with context for cancellation
	cmd := exec.CommandContext(ctx, ag.Path, args...)

	// Capture stdout and stderr together
	var outputBuf bytes.Buffer
	cmd.Stdout = &outputBuf
	cmd.Stderr = &outputBuf

	// Run command and wait for completion
	err := cmd.Run()

	// Get output
	output := outputBuf.String()

	// Apply parser if available
	if ag.Parser != nil {
		output = ag.Parser.Parse(output)
	}

	// Get exit code
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			// Other error (e.g., context cancelled, command not found)
			return OneShotResult{
				Output:   output,
				ExitCode: -1,
				Error:    err,
			}, err
		}
	}

	// Return result
	result := OneShotResult{
		Output:   output,
		ExitCode: exitCode,
		Error:    nil,
	}

	// If exit code is non-zero, set error
	if exitCode != 0 {
		result.Error = fmt.Errorf("agent exited with code %d", exitCode)
		return result, result.Error
	}

	return result, nil
}

// RunOneShotStreaming executes an agent in one-shot mode with real-time output streaming.
// Output is written to the provided io.Writer as it arrives, while still returning
// the complete result. If stream is nil, behaves identically to RunOneShot.
func RunOneShotStreaming(ctx context.Context, ag agent.Agent, prompt string, stream io.Writer) (OneShotResult, error) {
	// If no stream, delegate to RunOneShot
	if stream == nil {
		return RunOneShot(ctx, ag, prompt)
	}

	// Build command arguments
	args := buildOneShotArgs(ag.Pattern, prompt)

	// Create command with context for cancellation
	cmd := exec.CommandContext(ctx, ag.Path, args...)

	// Get stdout pipe for streaming
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return OneShotResult{
			ExitCode: -1,
			Error:    fmt.Errorf("failed to create stdout pipe: %w", err),
		}, err
	}

	// Capture stderr separately
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	// Start the command
	if err := cmd.Start(); err != nil {
		return OneShotResult{
			ExitCode: -1,
			Error:    fmt.Errorf("failed to start agent: %w", err),
		}, err
	}

	// Read stdout line by line, streaming and buffering
	var outputBuf bytes.Buffer
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		// Write to both stream and buffer
		_, _ = fmt.Fprintln(stream, line)
		outputBuf.WriteString(line)
		outputBuf.WriteString("\n")
	}

	// Wait for command to complete
	err = cmd.Wait()

	// Combine stdout and stderr
	output := outputBuf.String()
	if stderrBuf.Len() > 0 {
		output += stderrBuf.String()
		// Also stream stderr
		_, _ = fmt.Fprint(stream, stderrBuf.String())
	}

	// Apply parser if available
	if ag.Parser != nil {
		output = ag.Parser.Parse(output)
	}

	// Get exit code
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return OneShotResult{
				Output:   output,
				ExitCode: -1,
				Error:    err,
			}, err
		}
	}

	// Return result
	result := OneShotResult{
		Output:   output,
		ExitCode: exitCode,
		Error:    nil,
	}

	// If exit code is non-zero, set error
	if exitCode != 0 {
		result.Error = fmt.Errorf("agent exited with code %d", exitCode)
		return result, result.Error
	}

	return result, nil
}

// buildOneShotArgs builds command arguments for one-shot execution.
func buildOneShotArgs(pattern agent.CLIPattern, prompt string) []string {
	var args []string

	// Add non-interactive mode args
	args = append(args, pattern.NonInteractiveArgs...)

	// Add the prompt
	args = append(args, prompt)

	// Add JSON output args if available
	if len(pattern.JSONOutputArgs) > 0 {
		args = append(args, pattern.JSONOutputArgs...)
	}

	// Add skip approvals args if available
	if len(pattern.SkipApprovalsArgs) > 0 {
		args = append(args, pattern.SkipApprovalsArgs...)
	}

	return args
}
