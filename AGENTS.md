# Copilot Instructions for gencmd

## Project Overview

gencmd is a CLI tool that generates bash commands from natural language descriptions using LLMs (OpenAI, Google Gemini, Anthropic, Ollama, VertexAI). It operates interactively like `fzf`, filtering both command history and AI-generated suggestions.

## Architecture

### Core Components
- **`main.go`**: Entry point that loads XDG-based environment from `~/.config/gencmd/.env`
- **`cmd/`**: Cobra CLI commands (`root.go`, `init.go`, `generate.go`, `config.go`, etc.)
- **`config/`**: Configuration management with XDG directory support and multi-provider LLM setup
- **`ctrl/`**: Business logic controller managing history and LLM interactions
- **`ui/`**: Bubble Tea TUI with fuzzy filtering, command selection, and key bindings

### Data Flow
1. User invokes via key binding (Ctrl+G) or direct command
2. `ctrl.Controller` loads history from `~/.local/share/gencmd/history.jsonl`
3. UI shows filtered history + generates new commands via LLM APIs
4. Selected commands are pasted to terminal (not executed automatically)

## Key Patterns

### XDG Directory Convention
All data uses XDG base directories:
```go
// Configuration: ~/.config/gencmd/
configPath, _ := xdg.ConfigFile("gencmd/config.yaml")
// Data: ~/.local/share/gencmd/
historyPath, _ := xdg.DataFile("gencmd/history.jsonl")
```

### History Management
- JSONL format with atomic writes using temp files + rename
- Deduplication logic keeps most recent entries first
- Separate `rejected.jsonl` for deleted commands
- Thread-safe operations through file locking patterns

### Multi-Provider LLM Support
Provider detection via environment variables with Firebase Genkit v1 integration:
```go
// Auto-detects from: GEMINI_API_KEY, OPENAI_API_KEY, ANTHROPIC_API_KEY, OLLAMA_HOST, GOOGLE_GENAI_USE_VERTEXAI
cfg := config.DefaultFromEnv()
```

Current supported providers and default models:
- **GoogleAI**: `gemini-2.0-flash-lite` (free tier: 200 requests/day)
- **VertexAI**: `gemini-2.0-flash-lite` (supports Claude models via Model Garden)
- **OpenAI**: `gpt-4o-mini` (supports custom base URLs)
- **Anthropic**: `claude-3-5-haiku-latest`
- **Ollama**: `gemma-3` (local inference)

### UI State Management
Bubble Tea model with clear state transitions:
- `statePrompting` → `stateGenerating` → `stateSelecting` → `stateSelected`
- Fuzzy filtering combines history and generated commands
- Key bindings: Ctrl+G (invoke), Ctrl+D (delete history), Ctrl+H (toggle help)

## Development Workflows

### Building
```bash
make build          # Basic build
make schema         # Generate JSON schema for config
go build .          # Direct Go build
```

### Testing
```bash
go test -v ./...                    # Run all tests
go test -coverprofile=coverage.txt  # With coverage
```

### Configuration Schema
Auto-generated from Go structs via `hack/jsonschema/main.go`:
```bash
go run ./hack/jsonschema/main.go > config-schema.json
```

## Important Implementation Details

### Configuration Loading Priority
1. `~/.config/gencmd/config.yaml` (highest)
2. Environment variables
3. Embedded defaults

### Error Handling Convention
- Controllers return typed errors with context wrapping
- UI gracefully handles `ui.ErrUserCancel` without showing error
- Configuration loading continues with defaults on missing files

### Key Files to Modify
- **Adding LLM providers**: `config/config.go` + `ctrl/model.go` (uses Firebase Genkit plugins)
- **UI changes**: `ui/model.go` (state machine) + `ui/keymap.go` (bindings)
- **New commands**: Add to `cmd/` + register in `init()` functions
- **History format changes**: Update `ctrl.HistoryEntry` + migration logic
- **Model configuration**: Each provider uses specific Genkit plugins with proper initialization

### Testing Patterns
- Use `t.TempDir()` for file-based tests
- Controller tests verify JSONL integrity and atomic operations
- Configuration tests use `xdg.Reload()` after setting `XDG_CONFIG_HOME`

## Common Gotchas
- Always use XDG paths, never hardcode `~/.config/`
- History operations must be atomic (temp file + rename pattern)
- UI model updates require proper Bubble Tea command patterns
- Provider environment variables need exact naming (see `config.DefaultFromEnv()`)
- Firebase Genkit v1 requires proper plugin initialization for each provider
- VertexAI can use different model types (Gemini or Claude via Model Garden)
- Ollama models require explicit model definition with capabilities configuration
