# Buckshot

Multi-agent planning protocol for AI coding assistants.

Buckshot orchestrates multiple AI coding agents (Claude Code, Codex, Cursor) to collaboratively plan and refine development tasks using [beads](https://github.com/steveyegge/beads) (`bd`) for issue tracking.

## How It Works

1. You provide a planning prompt (e.g., "Build user authentication with JWT")
2. Each available AI agent takes a turn analyzing and refining the plan
3. Agents create, modify, and reorganize beads using `bd` commands
4. The process continues for N rounds or until agents converge (no more changes)
5. You end up with a well-reviewed, multi-perspective plan in beads

## Installation

```bash
go install github.com/michaellady/buckshot/cmd/buckshot@latest
```

Or build from source:

```bash
git clone https://github.com/michaellady/buckshot.git
cd buckshot
make install
```

## Prerequisites

- [beads](https://github.com/steveyegge/beads) (`bd`) - Issue tracking
- At least one AI coding agent CLI:
  - `claude` - [Claude Code](https://claude.ai/code)
  - `codex` - [OpenAI Codex CLI](https://github.com/openai/codex)
  - `cursor-agent` - [Cursor Agent](https://cursor.sh)

## Usage

### List Available Agents

```bash
buckshot agents
```

### Run Planning Protocol

```bash
# Basic usage (3 rounds by default)
buckshot plan "Build user authentication with JWT"

# Specify AGENTS.md for agents to follow
buckshot plan "Build auth system" --agents-path ./AGENTS.md

# Run more rounds
buckshot plan "Complex feature" --rounds 5

# Run until all agents agree the plan is complete
buckshot plan "Design API" --until-converged

# Use specific agents only
buckshot plan "Quick task" --agents claude,codex
```

## Architecture

```
                    buckshot plan "prompt"
                            │
                            ▼
                    ┌───────────────┐
                    │  Orchestrator │
                    └───────────────┘
                            │
         ┌──────────────────┼──────────────────┐
         ▼                  ▼                  ▼
   ┌──────────┐      ┌──────────┐      ┌──────────┐
   │  Claude  │      │  Codex   │      │  Cursor  │
   │ Session  │      │ Session  │      │ Session  │
   └──────────┘      └──────────┘      └──────────┘
         │                  │                  │
         └──────────────────┼──────────────────┘
                            ▼
                    ┌───────────────┐
                    │  beads (bd)   │
                    └───────────────┘
```

## Development

```bash
# Run tests
make test

# Run linter
make lint

# Generate coverage report
make coverage

# Build binary
make build
```

## License

MIT
