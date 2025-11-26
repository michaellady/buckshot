package dispatch

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/michaellady/buckshot/internal/agent"
	"github.com/michaellady/buckshot/internal/session"
)

// mockSession implements session.Session for testing.
type mockSession struct {
	agent       agent.Agent
	sendFunc    func(ctx context.Context, prompt string) (session.Response, error)
	sendDelay   time.Duration
	contextUsed float64
	alive       bool
	started     bool
	closed      bool
	mu          sync.Mutex
}

func newMockSession(name string) *mockSession {
	return &mockSession{
		agent: agent.Agent{
			Name:          name,
			Authenticated: true,
		},
		alive: true,
		sendFunc: func(ctx context.Context, prompt string) (session.Response, error) {
			return session.Response{Output: "response from " + name}, nil
		},
	}
}

func (m *mockSession) Start(ctx context.Context, agentsPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.started = true
	return nil
}

func (m *mockSession) Send(ctx context.Context, prompt string) (session.Response, error) {
	if m.sendDelay > 0 {
		select {
		case <-time.After(m.sendDelay):
		case <-ctx.Done():
			return session.Response{}, ctx.Err()
		}
	}
	if m.sendFunc != nil {
		return m.sendFunc(ctx, prompt)
	}
	return session.Response{}, nil
}

func (m *mockSession) ContextUsage() float64 {
	return m.contextUsed
}

func (m *mockSession) IsAlive() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.alive
}

func (m *mockSession) Agent() agent.Agent {
	return m.agent
}

func (m *mockSession) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	m.alive = false
	return nil
}

// TestDispatchConcurrent verifies that dispatch sends to all agents concurrently.
func TestDispatchConcurrent(t *testing.T) {
	var concurrentCalls int32
	var maxConcurrent int32
	var mu sync.Mutex

	sessions := make([]session.Session, 3)
	for i := 0; i < 3; i++ {
		name := []string{"alice", "bob", "charlie"}[i]
		mock := newMockSession(name)
		mock.sendDelay = 50 * time.Millisecond
		mock.sendFunc = func(ctx context.Context, prompt string) (session.Response, error) {
			// Track concurrent execution
			current := atomic.AddInt32(&concurrentCalls, 1)
			mu.Lock()
			if current > maxConcurrent {
				maxConcurrent = current
			}
			mu.Unlock()

			// Simulate work
			select {
			case <-time.After(50 * time.Millisecond):
			case <-ctx.Done():
				atomic.AddInt32(&concurrentCalls, -1)
				return session.Response{}, ctx.Err()
			}

			atomic.AddInt32(&concurrentCalls, -1)
			return session.Response{Output: "done"}, nil
		}
		sessions[i] = mock
	}

	d := New()
	ctx := context.Background()

	start := time.Now()
	results := d.Dispatch(ctx, sessions, "test prompt")
	elapsed := time.Since(start)

	// With 3 agents taking 50ms each:
	// - Sequential: ~150ms
	// - Parallel: ~50ms
	if elapsed > 100*time.Millisecond {
		t.Errorf("Dispatch took %v, expected < 100ms for parallel execution", elapsed)
	}

	// Verify we had concurrent execution
	if maxConcurrent < 2 {
		t.Errorf("Max concurrent calls was %d, expected at least 2 for parallel execution", maxConcurrent)
	}

	// Verify all results returned
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
}

// TestDispatchCollectsAllResults verifies all agent responses are collected.
func TestDispatchCollectsAllResults(t *testing.T) {
	sessions := []session.Session{
		newMockSession("agent1"),
		newMockSession("agent2"),
		newMockSession("agent3"),
	}

	// Set unique responses
	for i, s := range sessions {
		mock := s.(*mockSession)
		name := mock.agent.Name
		mock.sendFunc = func(ctx context.Context, prompt string) (session.Response, error) {
			return session.Response{
				Output:       "response from " + name,
				ContextUsage: float64(i) * 0.1,
			}, nil
		}
	}

	d := New()
	results := d.Dispatch(context.Background(), sessions, "test")

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// Verify each result has a response
	for _, r := range results {
		if r.Response.Output == "" {
			t.Errorf("Agent %s has empty response", r.Agent.Name)
		}
	}
}

// TestDispatchPartialFailure verifies handling when some agents fail.
func TestDispatchPartialFailure(t *testing.T) {
	errAgent := errors.New("agent failed")

	sessions := []session.Session{
		newMockSession("failing"),
		newMockSession("succeeding"),
	}

	// First agent fails
	sessions[0].(*mockSession).sendFunc = func(ctx context.Context, prompt string) (session.Response, error) {
		return session.Response{}, errAgent
	}

	// Second agent succeeds
	sessions[1].(*mockSession).sendFunc = func(ctx context.Context, prompt string) (session.Response, error) {
		return session.Response{Output: "success"}, nil
	}

	d := New()
	results := d.Dispatch(context.Background(), sessions, "test")

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Find results by name (order is deterministic)
	var failingResult, succeedingResult Result
	for _, r := range results {
		switch r.Agent.Name {
		case "failing":
			failingResult = r
		case "succeeding":
			succeedingResult = r
		}
	}

	// Verify failure is captured
	if failingResult.Error == nil {
		t.Error("Expected error for failing agent")
	}
	if !errors.Is(failingResult.Error, errAgent) {
		t.Errorf("Expected errAgent, got %v", failingResult.Error)
	}

	// Verify success is captured
	if succeedingResult.Error != nil {
		t.Errorf("Expected no error for succeeding agent, got %v", succeedingResult.Error)
	}
	if succeedingResult.Response.Output != "success" {
		t.Errorf("Expected 'success' output, got %q", succeedingResult.Response.Output)
	}
}

// TestDispatchRespectsTimeout verifies context timeout is respected.
func TestDispatchRespectsTimeout(t *testing.T) {
	sessions := []session.Session{
		newMockSession("slow"),
		newMockSession("fast"),
	}

	// Slow agent takes 500ms
	sessions[0].(*mockSession).sendFunc = func(ctx context.Context, prompt string) (session.Response, error) {
		select {
		case <-time.After(500 * time.Millisecond):
			return session.Response{Output: "slow done"}, nil
		case <-ctx.Done():
			return session.Response{}, ctx.Err()
		}
	}

	// Fast agent completes immediately
	sessions[1].(*mockSession).sendFunc = func(ctx context.Context, prompt string) (session.Response, error) {
		return session.Response{Output: "fast done"}, nil
	}

	d := New()

	// Set 100ms timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	results := d.Dispatch(ctx, sessions, "test")
	elapsed := time.Since(start)

	// Should complete near timeout, not wait for slow agent
	if elapsed > 200*time.Millisecond {
		t.Errorf("Dispatch took %v, expected to respect ~100ms timeout", elapsed)
	}

	// Find results
	var slowResult, fastResult Result
	for _, r := range results {
		switch r.Agent.Name {
		case "slow":
			slowResult = r
		case "fast":
			fastResult = r
		}
	}

	// Slow agent should have timeout error
	if slowResult.Error == nil {
		t.Error("Expected timeout error for slow agent")
	}
	if !errors.Is(slowResult.Error, context.DeadlineExceeded) {
		t.Errorf("Expected DeadlineExceeded, got %v", slowResult.Error)
	}

	// Fast agent should succeed
	if fastResult.Error != nil {
		t.Errorf("Expected no error for fast agent, got %v", fastResult.Error)
	}
}

// TestDispatchDeterministicOrder verifies results are sorted by agent name.
func TestDispatchDeterministicOrder(t *testing.T) {
	// Create sessions in non-alphabetical order
	sessions := []session.Session{
		newMockSession("zebra"),
		newMockSession("alpha"),
		newMockSession("mango"),
	}

	d := New()
	results := d.Dispatch(context.Background(), sessions, "test")

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// Verify alphabetical order
	expectedOrder := []string{"alpha", "mango", "zebra"}
	for i, r := range results {
		if r.Agent.Name != expectedOrder[i] {
			t.Errorf("Result %d: expected agent %s, got %s", i, expectedOrder[i], r.Agent.Name)
		}
	}
}

// TestDispatchEmptySessions verifies handling of empty session list.
func TestDispatchEmptySessions(t *testing.T) {
	d := New()
	results := d.Dispatch(context.Background(), nil, "test")

	if results == nil {
		t.Error("Expected non-nil results slice")
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

// TestDispatchContextCancellation verifies in-flight requests are cancelled.
func TestDispatchContextCancellation(t *testing.T) {
	cancelledCount := int32(0)

	sessions := make([]session.Session, 3)
	for i := 0; i < 3; i++ {
		mock := newMockSession([]string{"a", "b", "c"}[i])
		mock.sendFunc = func(ctx context.Context, prompt string) (session.Response, error) {
			select {
			case <-time.After(500 * time.Millisecond):
				return session.Response{Output: "done"}, nil
			case <-ctx.Done():
				atomic.AddInt32(&cancelledCount, 1)
				return session.Response{}, ctx.Err()
			}
		}
		sessions[i] = mock
	}

	d := New()

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel after 50ms
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	results := d.Dispatch(ctx, sessions, "test")
	elapsed := time.Since(start)

	// Should complete quickly after cancellation
	if elapsed > 150*time.Millisecond {
		t.Errorf("Dispatch took %v, expected quick return after cancellation", elapsed)
	}

	// All agents should have been cancelled
	for _, r := range results {
		if r.Error == nil {
			t.Errorf("Agent %s should have cancellation error", r.Agent.Name)
		}
		if !errors.Is(r.Error, context.Canceled) {
			t.Errorf("Agent %s: expected Canceled, got %v", r.Agent.Name, r.Error)
		}
	}
}

// TestDispatchSingleAgent verifies dispatch works with just one agent.
func TestDispatchSingleAgent(t *testing.T) {
	mock := newMockSession("solo")
	mock.sendFunc = func(ctx context.Context, prompt string) (session.Response, error) {
		return session.Response{Output: "solo response"}, nil
	}

	d := New()
	results := d.Dispatch(context.Background(), []session.Session{mock}, "test")

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].Agent.Name != "solo" {
		t.Errorf("Expected agent 'solo', got %s", results[0].Agent.Name)
	}

	if results[0].Response.Output != "solo response" {
		t.Errorf("Expected 'solo response', got %q", results[0].Response.Output)
	}

	if results[0].Error != nil {
		t.Errorf("Expected no error, got %v", results[0].Error)
	}
}
