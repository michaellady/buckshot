package orchestrator

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/michaellady/buckshot/internal/agent"
	buckctx "github.com/michaellady/buckshot/internal/context"
	"github.com/michaellady/buckshot/internal/session"
)

// TestRoundOrchestratorInterface ensures the interface is properly defined
func TestRoundOrchestratorInterface(t *testing.T) {
	var _ RoundOrchestrator = (*defaultOrchestrator)(nil)
}

// TestRunRound_ExecutesAgentsSequentially tests that RunRound executes each agent in order
func TestRunRound_ExecutesAgentsSequentially(t *testing.T) {
	orch := NewRoundOrchestrator()
	orch.SetSessionManager(session.NewManager())
	orch.SetContextBuilder(buckctx.NewBuilder())

	agents := []agent.Agent{
		{Name: "claude", Authenticated: true},
		{Name: "codex", Authenticated: true},
		{Name: "cursor", Authenticated: true},
	}

	planCtx := buckctx.PlanningContext{
		Prompt:     "Test prompt",
		AgentsPath: "/path/to/AGENTS.md",
		Round:      1,
	}

	ctx := context.Background()
	result, err := orch.RunRound(ctx, agents, planCtx)
	if err != nil {
		t.Fatalf("RunRound() error = %v", err)
	}

	// Should have results for all agents
	if len(result.AgentResults) != len(agents) {
		t.Fatalf("RunRound() returned %d results, want %d", len(result.AgentResults), len(agents))
	}

	// Each result should correspond to the agent in order
	for i, agentResult := range result.AgentResults {
		if agentResult.Agent.Name != agents[i].Name {
			t.Errorf("AgentResult[%d].Agent.Name = %q, want %q", i, agentResult.Agent.Name, agents[i].Name)
		}
	}
}

// TestRunRound_RefreshesBeadsStateBetweenAgents tests that beads state is refreshed after each agent
func TestRunRound_RefreshesBeadsStateBetweenAgents(t *testing.T) {
	orch := NewRoundOrchestrator()
	mockBuilder := &mockContextBuilder{beadsStates: []string{
		"initial state",
		"after agent 1",
		"after agent 2",
	}}
	orch.SetContextBuilder(mockBuilder)
	orch.SetSessionManager(session.NewManager())

	agents := []agent.Agent{
		{Name: "claude", Authenticated: true},
		{Name: "codex", Authenticated: true},
	}

	planCtx := buckctx.PlanningContext{
		Prompt:     "Test prompt",
		AgentsPath: "/path/to/AGENTS.md",
		Round:      1,
	}

	ctx := context.Background()
	_, err := orch.RunRound(ctx, agents, planCtx)
	if err != nil {
		t.Fatalf("RunRound() error = %v", err)
	}

	// RefreshBeadsState should be called after each agent
	// Initial + after each agent = len(agents) calls
	if mockBuilder.refreshCalls < len(agents) {
		t.Errorf("RefreshBeadsState() called %d times, want at least %d", mockBuilder.refreshCalls, len(agents))
	}
}

// TestRunRound_TracksChangesPerAgent tests that changes are tracked per agent
func TestRunRound_TracksChangesPerAgent(t *testing.T) {
	orch := NewRoundOrchestrator()
	orch.SetSessionManager(session.NewManager())
	orch.SetContextBuilder(buckctx.NewBuilder())

	agents := []agent.Agent{
		{Name: "claude", Authenticated: true},
		{Name: "codex", Authenticated: true},
	}

	planCtx := buckctx.PlanningContext{
		Prompt:     "Create some beads",
		AgentsPath: "/path/to/AGENTS.md",
		Round:      1,
	}

	ctx := context.Background()
	result, err := orch.RunRound(ctx, agents, planCtx)
	if err != nil {
		t.Fatalf("RunRound() error = %v", err)
	}

	// Each agent result should track beads changed
	for i, agentResult := range result.AgentResults {
		if agentResult.BeadsChanged == nil {
			t.Errorf("AgentResult[%d].BeadsChanged is nil, want slice", i)
		}
	}

	// TotalChanges should be sum of all BeadsChanged
	totalFromResults := 0
	for _, ar := range result.AgentResults {
		totalFromResults += len(ar.BeadsChanged)
	}
	if result.TotalChanges != totalFromResults {
		t.Errorf("TotalChanges = %d, want %d (sum of all BeadsChanged)", result.TotalChanges, totalFromResults)
	}
}

// TestRunRound_HandlesAgentFailuresGracefully tests that failures don't stop the round
func TestRunRound_HandlesAgentFailuresGracefully(t *testing.T) {
	orch := NewRoundOrchestrator()
	mockMgr := &mockSessionManager{
		failForAgent: "codex", // This agent will fail
	}
	orch.SetSessionManager(mockMgr)
	orch.SetContextBuilder(buckctx.NewBuilder())

	agents := []agent.Agent{
		{Name: "claude", Authenticated: true},
		{Name: "codex", Authenticated: true}, // Will fail
		{Name: "cursor", Authenticated: true},
	}

	planCtx := buckctx.PlanningContext{
		Prompt:     "Test prompt",
		AgentsPath: "/path/to/AGENTS.md",
		Round:      1,
	}

	ctx := context.Background()
	result, err := orch.RunRound(ctx, agents, planCtx)

	// RunRound should NOT return an error even if an agent fails
	if err != nil {
		t.Fatalf("RunRound() should not return error for agent failures, got %v", err)
	}

	// Should still have results for all agents
	if len(result.AgentResults) != len(agents) {
		t.Fatalf("RunRound() returned %d results, want %d", len(result.AgentResults), len(agents))
	}

	// codex result should have error
	codexResult := result.AgentResults[1]
	if codexResult.Error == nil {
		t.Error("AgentResult for codex should have error, got nil")
	}

	// Other agents should have succeeded
	if result.AgentResults[0].Error != nil {
		t.Errorf("AgentResult for claude should not have error, got %v", result.AgentResults[0].Error)
	}
	if result.AgentResults[2].Error != nil {
		t.Errorf("AgentResult for cursor should not have error, got %v", result.AgentResults[2].Error)
	}

	// FailedCount should be 1
	if result.FailedCount != 1 {
		t.Errorf("FailedCount = %d, want 1", result.FailedCount)
	}
}

// TestRunRound_ReturnsRoundNumber tests that result contains correct round number
func TestRunRound_ReturnsRoundNumber(t *testing.T) {
	orch := NewRoundOrchestrator()
	orch.SetSessionManager(session.NewManager())
	orch.SetContextBuilder(buckctx.NewBuilder())

	agents := []agent.Agent{
		{Name: "claude", Authenticated: true},
	}

	testCases := []int{1, 2, 5, 10}

	for _, roundNum := range testCases {
		planCtx := buckctx.PlanningContext{
			Prompt:     "Test prompt",
			AgentsPath: "/path/to/AGENTS.md",
			Round:      roundNum,
		}

		ctx := context.Background()
		result, err := orch.RunRound(ctx, agents, planCtx)
		if err != nil {
			t.Fatalf("RunRound() error = %v", err)
		}

		if result.Round != roundNum {
			t.Errorf("Round %d: result.Round = %d, want %d", roundNum, result.Round, roundNum)
		}
	}
}

// TestRunRound_WithEmptyAgentList tests handling of empty agent list
func TestRunRound_WithEmptyAgentList(t *testing.T) {
	orch := NewRoundOrchestrator()
	orch.SetSessionManager(session.NewManager())
	orch.SetContextBuilder(buckctx.NewBuilder())

	agents := []agent.Agent{} // Empty

	planCtx := buckctx.PlanningContext{
		Prompt:     "Test prompt",
		AgentsPath: "/path/to/AGENTS.md",
		Round:      1,
	}

	ctx := context.Background()
	result, err := orch.RunRound(ctx, agents, planCtx)
	if err != nil {
		t.Fatalf("RunRound() with empty agents should not error, got %v", err)
	}

	if len(result.AgentResults) != 0 {
		t.Errorf("RunRound() with empty agents returned %d results, want 0", len(result.AgentResults))
	}

	if result.TotalChanges != 0 {
		t.Errorf("TotalChanges = %d, want 0 for empty round", result.TotalChanges)
	}
}

// TestRunRound_SkipsUnauthenticatedAgents tests that unauthenticated agents are skipped
func TestRunRound_SkipsUnauthenticatedAgents(t *testing.T) {
	orch := NewRoundOrchestrator()
	orch.SetSessionManager(session.NewManager())
	orch.SetContextBuilder(buckctx.NewBuilder())

	agents := []agent.Agent{
		{Name: "claude", Authenticated: true},
		{Name: "codex", Authenticated: false}, // Not authenticated
		{Name: "cursor", Authenticated: true},
	}

	planCtx := buckctx.PlanningContext{
		Prompt:     "Test prompt",
		AgentsPath: "/path/to/AGENTS.md",
		Round:      1,
	}

	ctx := context.Background()
	result, err := orch.RunRound(ctx, agents, planCtx)
	if err != nil {
		t.Fatalf("RunRound() error = %v", err)
	}

	// Should still have results for all agents
	if len(result.AgentResults) != len(agents) {
		t.Fatalf("RunRound() returned %d results, want %d", len(result.AgentResults), len(agents))
	}

	// codex should be skipped
	codexResult := result.AgentResults[1]
	if !codexResult.Skipped {
		t.Error("AgentResult for unauthenticated codex should be Skipped=true")
	}

	// SkippedCount should be 1
	if result.SkippedCount != 1 {
		t.Errorf("SkippedCount = %d, want 1", result.SkippedCount)
	}
}

// TestRunRound_PropagatesOutputStreamToSession tests that SetOutputStream propagates to sessions
func TestRunRound_PropagatesOutputStreamToSession(t *testing.T) {
	orch := NewRoundOrchestrator()

	// Use a mock session manager that tracks stream setting
	mockMgr := &mockStreamingSessionManager{}
	orch.SetSessionManager(mockMgr)
	orch.SetContextBuilder(buckctx.NewBuilder())

	// Set output stream on orchestrator
	var buf bytes.Buffer
	orch.SetOutputStream(&buf)

	agents := []agent.Agent{
		{Name: "claude", Authenticated: true},
	}

	planCtx := buckctx.PlanningContext{
		Prompt:     "Test prompt",
		AgentsPath: "/path/to/AGENTS.md",
		Round:      1,
	}

	ctx := context.Background()
	_, err := orch.RunRound(ctx, agents, planCtx)
	if err != nil {
		t.Fatalf("RunRound() error = %v", err)
	}

	// Verify that the session received the output stream
	if mockMgr.lastSession == nil {
		t.Fatal("No session was created")
	}

	streamingSession, ok := mockMgr.lastSession.(*mockStreamingSession)
	if !ok {
		t.Fatal("Session is not a mockStreamingSession")
	}

	if streamingSession.outputStream == nil {
		t.Error("Session did not receive output stream - SetOutputStream was not called")
	}
}

// TestRunRound_OutputStreamIsNilByDefault tests that no stream is set when not configured
func TestRunRound_OutputStreamIsNilByDefault(t *testing.T) {
	orch := NewRoundOrchestrator()

	mockMgr := &mockStreamingSessionManager{}
	orch.SetSessionManager(mockMgr)
	orch.SetContextBuilder(buckctx.NewBuilder())

	// Don't set output stream

	agents := []agent.Agent{
		{Name: "claude", Authenticated: true},
	}

	planCtx := buckctx.PlanningContext{
		Prompt:     "Test prompt",
		AgentsPath: "/path/to/AGENTS.md",
		Round:      1,
	}

	ctx := context.Background()
	_, _ = orch.RunRound(ctx, agents, planCtx)

	// Session should not have output stream set
	if mockMgr.lastSession != nil {
		streamingSession := mockMgr.lastSession.(*mockStreamingSession)
		if streamingSession.outputStream != nil {
			t.Error("Session should not have output stream when orchestrator stream is nil")
		}
	}
}

// TestSetOutputStream_AcceptsWriter tests that SetOutputStream accepts an io.Writer
func TestSetOutputStream_AcceptsWriter(t *testing.T) {
	orch := NewRoundOrchestrator()

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SetOutputStream panicked: %v", r)
		}
	}()

	var buf bytes.Buffer
	orch.SetOutputStream(&buf)
}

// TestSetOutputStream_AcceptsNil tests that SetOutputStream accepts nil
func TestSetOutputStream_AcceptsNil(t *testing.T) {
	orch := NewRoundOrchestrator()

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SetOutputStream(nil) panicked: %v", r)
		}
	}()

	orch.SetOutputStream(nil)
}

// Mock implementations for testing

// mockStreamingSessionManager tracks sessions that support streaming
type mockStreamingSessionManager struct {
	lastSession session.Session
}

func (m *mockStreamingSessionManager) CreateSession(a agent.Agent) (session.Session, error) {
	sess := &mockStreamingSession{agent: a}
	m.lastSession = sess
	return sess, nil
}

func (m *mockStreamingSessionManager) ShouldRespawn(s session.Session, threshold float64) bool {
	return false
}

// mockStreamingSession implements Session with SetOutputStream support
type mockStreamingSession struct {
	agent        agent.Agent
	started      bool
	outputStream io.Writer
}

func (s *mockStreamingSession) Start(ctx context.Context, agentsPath string) error {
	s.started = true
	return nil
}

func (s *mockStreamingSession) Send(ctx context.Context, prompt string) (session.Response, error) {
	// Write to output stream if set
	if s.outputStream != nil {
		_, _ = s.outputStream.Write([]byte("Mock streaming output\n"))
	}
	return session.Response{
		Output:       "Mock response",
		ContextUsage: 0.1,
	}, nil
}

func (s *mockStreamingSession) ContextUsage() float64 {
	return 0.1
}

func (s *mockStreamingSession) IsAlive() bool {
	return s.started
}

func (s *mockStreamingSession) Agent() agent.Agent {
	return s.agent
}

func (s *mockStreamingSession) Close() error {
	s.started = false
	return nil
}

func (s *mockStreamingSession) SetOutputStream(w io.Writer) {
	s.outputStream = w
}

type mockContextBuilder struct {
	beadsStates  []string
	refreshCalls int
	currentIdx   int
}

func (m *mockContextBuilder) Build(prompt string, agentsPath string, round int, isFirstTurn bool) (buckctx.PlanningContext, error) {
	return buckctx.PlanningContext{
		Prompt:      prompt,
		AgentsPath:  agentsPath,
		Round:       round,
		IsFirstTurn: isFirstTurn,
	}, nil
}

func (m *mockContextBuilder) Format(ctx buckctx.PlanningContext) string {
	return ctx.Prompt
}

func (m *mockContextBuilder) FormatFeedback(ctx buckctx.PlanningContext) string {
	return ctx.Prompt
}

func (m *mockContextBuilder) RefreshBeadsState(ctx *buckctx.PlanningContext) error {
	m.refreshCalls++
	if m.currentIdx < len(m.beadsStates) {
		ctx.BeadsState = m.beadsStates[m.currentIdx]
		m.currentIdx++
	}
	return nil
}

type mockSessionManager struct {
	failForAgent string
}

func (m *mockSessionManager) CreateSession(a agent.Agent) (session.Session, error) {
	return &mockSession{agent: a, shouldFail: a.Name == m.failForAgent}, nil
}

func (m *mockSessionManager) ShouldRespawn(s session.Session, threshold float64) bool {
	return false
}

type mockSession struct {
	agent      agent.Agent
	shouldFail bool
	started    bool
}

func (s *mockSession) Start(ctx context.Context, agentsPath string) error {
	s.started = true
	return nil
}

func (s *mockSession) Send(ctx context.Context, prompt string) (session.Response, error) {
	if s.shouldFail {
		return session.Response{Error: context.DeadlineExceeded}, context.DeadlineExceeded
	}
	return session.Response{
		Output:       "Mock response",
		ContextUsage: 0.1,
	}, nil
}

func (s *mockSession) ContextUsage() float64 {
	return 0.1
}

func (s *mockSession) IsAlive() bool {
	return s.started
}

func (s *mockSession) Agent() agent.Agent {
	return s.agent
}

func (s *mockSession) Close() error {
	s.started = false
	return nil
}
