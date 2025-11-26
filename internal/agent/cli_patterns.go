// Package agent provides detection and management of AI coding agents.
//
// # CLI Invocation Patterns (Research: buckshot-b0y)
//
// ## Claude Code (`claude`)
//
// Binary: claude (typically /opt/homebrew/bin/claude or ~/.claude/local/claude)
// Version check: claude --version
// Non-interactive mode: claude -p "prompt"
// JSON output: --output-format json | stream-json
// Skip permissions: --dangerously-skip-permissions
// System prompt: --system-prompt "prompt" or --append-system-prompt "prompt"
// Continue session: --continue or --resume [sessionId]
//
// Example:
//
//	claude -p "Create a hello world function" \
//	  --output-format json \
//	  --dangerously-skip-permissions
//
// ## Codex (`codex`)
//
// Binary: codex (typically /opt/homebrew/bin/codex)
// Version check: codex --version
// Non-interactive mode: codex exec "prompt"
// JSON output: --json (outputs JSONL events)
// Skip approvals: --dangerously-bypass-approvals-and-sandbox
// Working directory: -C, --cd <DIR>
// Output last message: -o, --output-last-message <FILE>
//
// Example:
//
//	codex exec "Create a hello world function" \
//	  --json \
//	  --dangerously-bypass-approvals-and-sandbox
//
// ## Cursor Agent (`cursor-agent`)
//
// Binary: cursor-agent (typically ~/.local/bin/cursor-agent)
// Version check: cursor-agent --version
// Non-interactive mode: cursor-agent -p "prompt"
// JSON output: --output-format json | stream-json
// Skip approvals: -f, --force
// Workspace: --workspace <path>
// Resume session: --resume [chatId]
//
// Example:
//
//	cursor-agent -p "Create a hello world function" \
//	  --output-format json \
//	  --force
//
// ## Context Usage Tracking
//
// None of the CLIs directly expose context usage percentage.
// Strategies for tracking:
// 1. Parse output for context-related messages
// 2. Track token counts from JSON responses (if available)
// 3. Monitor for "context limit" error messages
// 4. Use session continuation to gauge remaining capacity
//
// ## Authentication Checks
//
// - Claude: Runs successfully with --version; auth issues surface on first prompt
// - Codex: Has `codex login` command; check exit code of simple command
// - Cursor: Has `cursor-agent status` or `cursor-agent whoami` command
package agent

// CLIPattern defines the invocation pattern for an AI agent CLI.
type CLIPattern struct {
	// Binary is the executable name
	Binary string

	// VersionArgs are the arguments to check version/installation
	VersionArgs []string

	// AuthCheckCmd is the command to verify authentication (optional)
	AuthCheckCmd []string

	// NonInteractiveArgs are base args for non-interactive mode
	NonInteractiveArgs []string

	// JSONOutputArgs are args to enable JSON output
	JSONOutputArgs []string

	// SkipApprovalsArgs are args to skip permission prompts
	SkipApprovalsArgs []string

	// SystemPromptArg is the flag for setting system prompt (if supported)
	SystemPromptArg string

	// WorkspaceDirArg is the flag for setting working directory
	WorkspaceDirArg string

	// ResumeSessionArg is the flag for resuming a session
	ResumeSessionArg string
}

// KnownAgents returns CLI patterns for all supported agents.
func KnownAgents() map[string]CLIPattern {
	return map[string]CLIPattern{
		"claude": {
			Binary:             "claude",
			VersionArgs:        []string{"--version"},
			AuthCheckCmd:       []string{"--version"}, // Auth checked on first real command
			NonInteractiveArgs: []string{"-p"},
			JSONOutputArgs:     []string{"--output-format", "stream-json"},
			SkipApprovalsArgs:  []string{"--dangerously-skip-permissions"},
			SystemPromptArg:    "--append-system-prompt",
			WorkspaceDirArg:    "", // Uses current directory
			ResumeSessionArg:   "--resume",
		},
		"codex": {
			Binary:             "codex",
			VersionArgs:        []string{"--version"},
			AuthCheckCmd:       []string{"--version"},
			NonInteractiveArgs: []string{"exec"},
			JSONOutputArgs:     []string{"--json"},
			SkipApprovalsArgs:  []string{"--dangerously-bypass-approvals-and-sandbox"},
			SystemPromptArg:    "", // Not directly supported
			WorkspaceDirArg:    "--cd",
			ResumeSessionArg:   "", // exec resume subcommand
		},
		"cursor-agent": {
			Binary:             "cursor-agent",
			VersionArgs:        []string{"--version"},
			AuthCheckCmd:       []string{"status"},
			NonInteractiveArgs: []string{"-p"},
			JSONOutputArgs:     []string{"--output-format", "stream-json"},
			SkipApprovalsArgs:  []string{"--force"},
			SystemPromptArg:    "", // Not directly supported
			WorkspaceDirArg:    "--workspace",
			ResumeSessionArg:   "--resume",
		},
	}
}
