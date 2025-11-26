package session

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/michaellady/buckshot/internal/agent"
)

// DefaultSession implements the Session interface using an underlying agent CLI process.
type DefaultSession struct {
	agent         agent.Agent
	cmd           *exec.Cmd
	stdin         io.WriteCloser
	stdout        io.ReadCloser
	stderr        io.ReadCloser
	contextUsage  float64
	alive         bool
	mu            sync.Mutex
	agentsPath    string
	started       bool
	outputBuffer  strings.Builder
}

// Start initializes the session with the path to AGENTS.md.
func (s *DefaultSession) Start(ctx context.Context, agentsPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return errors.New("session already started")
	}

	// Validate AGENTS.md exists
	if _, err := os.Stat(agentsPath); err != nil {
		return fmt.Errorf("AGENTS.md not found at %s: %w", agentsPath, err)
	}

	s.agentsPath = agentsPath

	// Build command based on agent pattern
	pattern := s.agent.Pattern
	args := buildStartCommand(pattern, agentsPath)

	s.cmd = exec.CommandContext(ctx, s.agent.Path, args...)

	// Set up pipes for stdin/stdout/stderr
	var err error
	s.stdin, err = s.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	s.stdout, err = s.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	s.stderr, err = s.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	s.alive = true
	s.started = true

	// Start goroutines to read output
	go s.readOutput(s.stdout)
	go s.readOutput(s.stderr)

	return nil
}

// buildStartCommand builds the command arguments for starting an agent session.
func buildStartCommand(pattern agent.CLIPattern, agentsPath string) []string {
	var args []string

	// Add non-interactive args
	args = append(args, pattern.NonInteractiveArgs...)

	// Add the initial prompt to read AGENTS.md
	initialPrompt := fmt.Sprintf("please read and apply %s", agentsPath)
	args = append(args, initialPrompt)

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

// readOutput reads from a pipe and stores output.
func (s *DefaultSession) readOutput(pipe io.ReadCloser) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()
		s.mu.Lock()
		s.outputBuffer.WriteString(line)
		s.outputBuffer.WriteString("\n")

		// Parse context usage from output
		if usage := parseContextUsage(line); usage >= 0 {
			s.contextUsage = usage
		}
		s.mu.Unlock()
	}
}

// parseContextUsage extracts context usage from agent output.
// Looks for patterns like "Context: 15% used" or "15% used (29368/200000 tokens)"
var contextUsageRegex = regexp.MustCompile(`(?i)(\d+)%\s+used`)

func parseContextUsage(line string) float64 {
	matches := contextUsageRegex.FindStringSubmatch(line)
	if len(matches) >= 2 {
		if pct, err := strconv.Atoi(matches[1]); err == nil {
			return float64(pct) / 100.0
		}
	}
	return -1.0
}

// Send sends a prompt to the agent and returns the response.
func (s *DefaultSession) Send(ctx context.Context, prompt string) (Response, error) {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return Response{}, errors.New("session not started")
	}
	if !s.alive {
		s.mu.Unlock()
		return Response{}, errors.New("session not alive")
	}

	// Clear output buffer before sending
	s.outputBuffer.Reset()
	s.mu.Unlock()

	// Write prompt to stdin
	_, err := fmt.Fprintln(s.stdin, prompt)
	if err != nil {
		s.mu.Lock()
		s.alive = false
		s.mu.Unlock()
		return Response{Error: fmt.Errorf("failed to send prompt: %w", err)}, err
	}

	// Wait for response (in real implementation, we'd wait for a proper delimiter)
	// For now, we'll simulate a small delay to let output accumulate
	// In production, we'd parse JSON stream or look for specific markers

	// Get output
	s.mu.Lock()
	output := s.outputBuffer.String()
	usage := s.contextUsage
	s.mu.Unlock()

	return Response{
		Output:       output,
		ContextUsage: usage,
		Error:        nil,
	}, nil
}

// ContextUsage returns the current context usage (0.0 to 1.0).
func (s *DefaultSession) ContextUsage() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.contextUsage
}

// IsAlive returns whether the session is still active.
func (s *DefaultSession) IsAlive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return false
	}

	if !s.alive {
		return false
	}

	// Check if process is still running
	if s.cmd != nil && s.cmd.Process != nil {
		// Try to check process state (this is platform-specific)
		// On Unix, we can send signal 0 to check if process exists
		// For now, we'll rely on our alive flag
		return true
	}

	return false
}

// Agent returns the underlying agent for this session.
func (s *DefaultSession) Agent() agent.Agent {
	return s.agent
}

// Close terminates the session.
func (s *DefaultSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return nil // Already closed or never started
	}

	s.alive = false

	// Close stdin to signal end of input
	if s.stdin != nil {
		s.stdin.Close()
	}

	// Close stdout and stderr
	if s.stdout != nil {
		s.stdout.Close()
	}
	if s.stderr != nil {
		s.stderr.Close()
	}

	// Kill the process if still running
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
		_ = s.cmd.Wait() // Clean up zombie process
	}

	s.started = false
	return nil
}

// DefaultManager is the default implementation of Manager.
type DefaultManager struct{}

// NewManager creates a new session manager.
func NewManager() Manager {
	return &DefaultManager{}
}

// CreateSession creates a new session for the given agent.
func (m *DefaultManager) CreateSession(agent agent.Agent) (Session, error) {
	if !agent.Authenticated {
		return nil, errors.New("agent not authenticated")
	}

	return &DefaultSession{
		agent:        agent,
		contextUsage: 0.0,
		alive:        false,
		started:      false,
	}, nil
}

// ShouldRespawn returns true if session context > threshold.
func (m *DefaultManager) ShouldRespawn(session Session, threshold float64) bool {
	return session.ContextUsage() > threshold
}
