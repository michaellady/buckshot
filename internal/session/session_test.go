package session

import (
	"context"
	"testing"

	"github.com/michaellady/buckshot/internal/agent"
)

// TestSessionInterface ensures Session interface is properly defined
func TestSessionInterface(t *testing.T) {
	// This test verifies the interface exists and can be implemented
	var _ Session = (*DefaultSession)(nil)
}

// TestManagerInterface ensures Manager interface is properly defined
func TestManagerInterface(t *testing.T) {
	// This test verifies the interface exists and can be implemented
	var _ Manager = (*DefaultManager)(nil)
}

// TestSessionStart tests that Start initializes session with AGENTS.md path
func TestSessionStart(t *testing.T) {
	// Create a mock agent
	mockAgent := agent.Agent{
		Name:          "claude",
		Path:          "/usr/bin/claude",
		Authenticated: true,
		Version:       "1.0.0",
		Pattern:       agent.KnownAgents()["claude"],
	}

	mgr := NewManager()
	sess, err := mgr.CreateSession(mockAgent)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	defer sess.Close()

	// Start should initialize the session with AGENTS.md path
	ctx := context.Background()
	agentsPath := "/Users/mikelady/dev/AGENTS/AGENTS.md"
	err = sess.Start(ctx, agentsPath)
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}

	// After Start, session should be alive
	if !sess.IsAlive() {
		t.Error("IsAlive() = false after Start(), want true")
	}
}

// TestSessionStartWithInvalidPath tests Start with non-existent AGENTS.md
func TestSessionStartWithInvalidPath(t *testing.T) {
	mockAgent := agent.Agent{
		Name:          "claude",
		Path:          "/usr/bin/claude",
		Authenticated: true,
		Version:       "1.0.0",
		Pattern:       agent.KnownAgents()["claude"],
	}

	mgr := NewManager()
	sess, err := mgr.CreateSession(mockAgent)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	defer sess.Close()

	ctx := context.Background()
	err = sess.Start(ctx, "/nonexistent/AGENTS.md")
	if err == nil {
		t.Error("Start() with invalid path should return error, got nil")
	}
}

// TestSessionSend tests sending a prompt and receiving a response
func TestSessionSend(t *testing.T) {
	mockAgent := agent.Agent{
		Name:          "claude",
		Path:          "/usr/bin/claude",
		Authenticated: true,
		Version:       "1.0.0",
		Pattern:       agent.KnownAgents()["claude"],
	}

	mgr := NewManager()
	sess, err := mgr.CreateSession(mockAgent)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	defer sess.Close()

	ctx := context.Background()
	agentsPath := "/Users/mikelady/dev/AGENTS/AGENTS.md"
	if err := sess.Start(ctx, agentsPath); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Send a simple prompt
	resp, err := sess.Send(ctx, "echo 'hello'")
	if err != nil {
		t.Errorf("Send() error = %v", err)
	}

	// Response should have output
	if resp.Output == "" {
		t.Error("Send() response.Output is empty, want non-empty")
	}

	// Response should have valid context usage
	if resp.ContextUsage < 0.0 || resp.ContextUsage > 1.0 {
		t.Errorf("Send() response.ContextUsage = %f, want 0.0-1.0", resp.ContextUsage)
	}

	// Response should not have an error
	if resp.Error != nil {
		t.Errorf("Send() response.Error = %v, want nil", resp.Error)
	}
}

// TestSessionSendWithoutStart tests that Send fails if Start not called
func TestSessionSendWithoutStart(t *testing.T) {
	mockAgent := agent.Agent{
		Name:          "claude",
		Path:          "/usr/bin/claude",
		Authenticated: true,
		Version:       "1.0.0",
		Pattern:       agent.KnownAgents()["claude"],
	}

	mgr := NewManager()
	sess, err := mgr.CreateSession(mockAgent)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	defer sess.Close()

	ctx := context.Background()
	_, err = sess.Send(ctx, "echo 'hello'")
	if err == nil {
		t.Error("Send() without Start() should return error, got nil")
	}
}

// TestSessionContextUsage tests that ContextUsage returns 0.0-1.0
func TestSessionContextUsage(t *testing.T) {
	mockAgent := agent.Agent{
		Name:          "claude",
		Path:          "/usr/bin/claude",
		Authenticated: true,
		Version:       "1.0.0",
		Pattern:       agent.KnownAgents()["claude"],
	}

	mgr := NewManager()
	sess, err := mgr.CreateSession(mockAgent)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	defer sess.Close()

	// Before Start, context usage should be 0
	usage := sess.ContextUsage()
	if usage != 0.0 {
		t.Errorf("ContextUsage() before Start = %f, want 0.0", usage)
	}

	ctx := context.Background()
	agentsPath := "/Users/mikelady/dev/AGENTS/AGENTS.md"
	if err := sess.Start(ctx, agentsPath); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// After Start, context usage should still be valid
	usage = sess.ContextUsage()
	if usage < 0.0 || usage > 1.0 {
		t.Errorf("ContextUsage() after Start = %f, want 0.0-1.0", usage)
	}

	// After Send, context usage should increase
	if _, err := sess.Send(ctx, "echo 'test'"); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	newUsage := sess.ContextUsage()
	if newUsage < 0.0 || newUsage > 1.0 {
		t.Errorf("ContextUsage() after Send = %f, want 0.0-1.0", newUsage)
	}

	if newUsage <= usage {
		t.Errorf("ContextUsage() after Send = %f, want > %f", newUsage, usage)
	}
}

// TestSessionIsAlive tests session lifecycle checks
func TestSessionIsAlive(t *testing.T) {
	mockAgent := agent.Agent{
		Name:          "claude",
		Path:          "/usr/bin/claude",
		Authenticated: true,
		Version:       "1.0.0",
		Pattern:       agent.KnownAgents()["claude"],
	}

	mgr := NewManager()
	sess, err := mgr.CreateSession(mockAgent)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	// Before Start, session should not be alive
	if sess.IsAlive() {
		t.Error("IsAlive() before Start = true, want false")
	}

	ctx := context.Background()
	agentsPath := "/Users/mikelady/dev/AGENTS/AGENTS.md"
	if err := sess.Start(ctx, agentsPath); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// After Start, session should be alive
	if !sess.IsAlive() {
		t.Error("IsAlive() after Start = false, want true")
	}

	// After Close, session should not be alive
	if err := sess.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	if sess.IsAlive() {
		t.Error("IsAlive() after Close = true, want false")
	}
}

// TestSessionClose tests clean termination
func TestSessionClose(t *testing.T) {
	mockAgent := agent.Agent{
		Name:          "claude",
		Path:          "/usr/bin/claude",
		Authenticated: true,
		Version:       "1.0.0",
		Pattern:       agent.KnownAgents()["claude"],
	}

	mgr := NewManager()
	sess, err := mgr.CreateSession(mockAgent)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	ctx := context.Background()
	agentsPath := "/Users/mikelady/dev/AGENTS/AGENTS.md"
	if err := sess.Start(ctx, agentsPath); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Close should succeed
	err = sess.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}

	// Multiple Close calls should be safe (idempotent)
	err = sess.Close()
	if err != nil {
		t.Errorf("Second Close() error = %v, want nil (idempotent)", err)
	}
}

// TestSessionAgent tests getting the underlying agent
func TestSessionAgent(t *testing.T) {
	mockAgent := agent.Agent{
		Name:          "claude",
		Path:          "/usr/bin/claude",
		Authenticated: true,
		Version:       "1.0.0",
		Pattern:       agent.KnownAgents()["claude"],
	}

	mgr := NewManager()
	sess, err := mgr.CreateSession(mockAgent)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	defer sess.Close()

	// Agent() should return the same agent
	returnedAgent := sess.Agent()
	if returnedAgent.Name != mockAgent.Name {
		t.Errorf("Agent().Name = %q, want %q", returnedAgent.Name, mockAgent.Name)
	}
	if returnedAgent.Path != mockAgent.Path {
		t.Errorf("Agent().Path = %q, want %q", returnedAgent.Path, mockAgent.Path)
	}
}

// TestManagerCreateSession tests session creation
func TestManagerCreateSession(t *testing.T) {
	mockAgent := agent.Agent{
		Name:          "claude",
		Path:          "/usr/bin/claude",
		Authenticated: true,
		Version:       "1.0.0",
		Pattern:       agent.KnownAgents()["claude"],
	}

	mgr := NewManager()
	sess, err := mgr.CreateSession(mockAgent)
	if err != nil {
		t.Errorf("CreateSession() error = %v", err)
	}
	if sess == nil {
		t.Error("CreateSession() returned nil session")
	}
	defer sess.Close()
}

// TestManagerCreateSessionWithUnauthenticatedAgent tests that creation fails for unauthenticated agent
func TestManagerCreateSessionWithUnauthenticatedAgent(t *testing.T) {
	mockAgent := agent.Agent{
		Name:          "claude",
		Path:          "/usr/bin/claude",
		Authenticated: false, // Not authenticated
		Version:       "1.0.0",
		Pattern:       agent.KnownAgents()["claude"],
	}

	mgr := NewManager()
	_, err := mgr.CreateSession(mockAgent)
	if err == nil {
		t.Error("CreateSession() with unauthenticated agent should return error, got nil")
	}
}

// TestManagerShouldRespawn tests context threshold checking
func TestManagerShouldRespawn(t *testing.T) {
	mockAgent := agent.Agent{
		Name:          "claude",
		Path:          "/usr/bin/claude",
		Authenticated: true,
		Version:       "1.0.0",
		Pattern:       agent.KnownAgents()["claude"],
	}

	mgr := NewManager()
	sess, err := mgr.CreateSession(mockAgent)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	defer sess.Close()

	// With 0% context usage, should not respawn at 50% threshold
	if mgr.ShouldRespawn(sess, 0.5) {
		t.Error("ShouldRespawn() at 0% usage with 50% threshold = true, want false")
	}

	// Simulate high context usage by sending prompts
	ctx := context.Background()
	agentsPath := "/Users/mikelady/dev/AGENTS/AGENTS.md"
	if err := sess.Start(ctx, agentsPath); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Send multiple prompts to increase context usage
	for i := 0; i < 10; i++ {
		_, _ = sess.Send(ctx, "echo 'test'")
	}

	// After many sends, context usage should be higher
	usage := sess.ContextUsage()
	if usage >= 0.5 {
		// If usage is above 50%, ShouldRespawn(0.5) should return true
		if !mgr.ShouldRespawn(sess, 0.5) {
			t.Errorf("ShouldRespawn() at %.1f%% usage with 50%% threshold = false, want true", usage*100)
		}
	}
}

// TestSessionPersistence tests that sessions persist if context < 50%
func TestSessionPersistence(t *testing.T) {
	mockAgent := agent.Agent{
		Name:          "claude",
		Path:          "/usr/bin/claude",
		Authenticated: true,
		Version:       "1.0.0",
		Pattern:       agent.KnownAgents()["claude"],
	}

	mgr := NewManager()
	sess, err := mgr.CreateSession(mockAgent)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	defer sess.Close()

	ctx := context.Background()
	agentsPath := "/Users/mikelady/dev/AGENTS/AGENTS.md"
	if err := sess.Start(ctx, agentsPath); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Send a small prompt - context should be < 50%
	_, err = sess.Send(ctx, "echo 'hello'")
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	usage := sess.ContextUsage()
	if usage >= 0.5 {
		t.Skip("Skipping test - single prompt exceeded 50% context")
	}

	// Session should still be alive after prompt with low context
	if !sess.IsAlive() {
		t.Error("IsAlive() = false after low-context prompt, want true (session should persist)")
	}

	// Should not need respawn at 50% threshold
	if mgr.ShouldRespawn(sess, 0.5) {
		t.Errorf("ShouldRespawn() at %.1f%% usage with 50%% threshold = true, want false", usage*100)
	}
}

// TestSessionMultipleSends tests sending multiple prompts
func TestSessionMultipleSends(t *testing.T) {
	mockAgent := agent.Agent{
		Name:          "claude",
		Path:          "/usr/bin/claude",
		Authenticated: true,
		Version:       "1.0.0",
		Pattern:       agent.KnownAgents()["claude"],
	}

	mgr := NewManager()
	sess, err := mgr.CreateSession(mockAgent)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	defer sess.Close()

	ctx := context.Background()
	agentsPath := "/Users/mikelady/dev/AGENTS/AGENTS.md"
	if err := sess.Start(ctx, agentsPath); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Send multiple prompts
	prompts := []string{
		"echo 'first'",
		"echo 'second'",
		"echo 'third'",
	}

	var lastUsage float64
	for i, prompt := range prompts {
		resp, err := sess.Send(ctx, prompt)
		if err != nil {
			t.Errorf("Send() prompt %d error = %v", i, err)
		}

		if resp.Output == "" {
			t.Errorf("Send() prompt %d output is empty", i)
		}

		// Context usage should increase with each send
		usage := sess.ContextUsage()
		if i > 0 && usage <= lastUsage {
			t.Errorf("Send() prompt %d usage = %f, want > %f", i, usage, lastUsage)
		}
		lastUsage = usage
	}

	// Session should still be alive after multiple prompts
	if !sess.IsAlive() {
		t.Error("IsAlive() = false after multiple prompts, want true")
	}
}
