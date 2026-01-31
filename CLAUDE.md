# CLAUDE.md

This file provides guidance for AI agents working with the salesforce-cli codebase.

## Project Overview

salesforce-cli is a command-line interface for Salesforce written in Go. It uses OAuth 2.0 for authentication with secure token storage in the system keychain.

**Binary name:** `sfdc`
**Module:** `github.com/open-cli-collective/salesforce-cli`

## Quick Commands

```bash
# Build
make build

# Run tests
make test

# Run tests with coverage
make test-cover

# Lint
make lint

# Format code
make fmt

# All checks (format, lint, test)
make verify

# Install locally
make install

# Clean build artifacts
make clean
```

## Architecture

```
salesforce-cli/
├── cmd/sfdc/main.go              # Entry point
├── api/                          # Public Go library (Salesforce REST API client)
│   ├── client.go                 # Client struct, HTTP helpers
│   ├── types.go                  # Data types (SObject, QueryResult, etc.)
│   └── errors.go                 # Error types and parsing
├── internal/
│   ├── cmd/                      # Cobra commands
│   │   ├── root/                 # Root command, Options struct, global flags
│   │   ├── initcmd/              # sfdc init (OAuth setup wizard)
│   │   ├── configcmd/            # sfdc config {show,test,clear}
│   │   └── completion/           # Shell completion
│   ├── auth/                     # OAuth 2.0 implementation
│   │   ├── auth.go               # GetHTTPClient, token exchange
│   │   └── config.go             # Credential paths
│   ├── config/                   # Configuration file management
│   ├── keychain/                 # Secure token storage
│   │   ├── keychain.go           # Cross-platform interface
│   │   ├── keychain_darwin.go    # macOS Keychain
│   │   ├── keychain_linux.go     # Linux secret-tool
│   │   ├── keychain_windows.go   # Windows fallback
│   │   └── token_source.go       # PersistentTokenSource wrapper
│   ├── errors/                   # Shared error types
│   ├── version/                  # Build-time version injection
│   └── view/                     # Output formatting (table/json/plain)
├── .github/workflows/
│   ├── ci.yml                    # Build, test, lint on PR/push
│   ├── auto-release.yml          # Create tags on main push
│   └── release.yml               # Build and release binaries
├── packaging/
│   ├── chocolatey/               # Windows Chocolatey package
│   └── winget/                   # Windows Winget manifests
├── snap/                         # Snap package (name: ocli-sfdc)
├── Makefile                      # Build, test, lint targets
├── .goreleaser.yml               # Cross-platform builds
└── .golangci.yml                 # Linter config
```

## Key Patterns

### OAuth 2.0 Authentication

Uses OAuth 2.0 Web Server Flow with automatic token refresh:
- Tokens stored in platform keychain (macOS Keychain, Linux secret-tool)
- Falls back to `~/.config/salesforce-cli/token.json` with 0600 permissions
- `PersistentTokenSource` automatically persists refreshed tokens

### Options Struct Pattern

Commands use an Options struct for dependency injection:

```go
type Options struct {
    Output  string      // table, json, plain
    NoColor bool
    Stdin   io.Reader   // Injectable for testing
    Stdout  io.Writer
    Stderr  io.Writer
}
```

### Register Pattern

Each command package exports a Register function:

```go
func Register(rootCmd *cobra.Command, opts *root.Options) {
    cmd := &cobra.Command{
        Use:   "config",
        Short: "Manage configuration",
    }
    cmd.AddCommand(newShowCmd(opts))
    rootCmd.AddCommand(cmd)
}
```

### View Pattern

Use the View struct for formatted output:

```go
v := view.New(opts.Output, opts.NoColor)
v.Table(headers, rows)  // Table output
v.JSON(data)            // JSON output
```

## Testing

- Unit tests in `*_test.go` files alongside source
- Use `testify/assert` for assertions
- Table-driven tests with `t.Run()`
- Use `httptest.NewServer()` to mock API responses
- Use `t.TempDir()` for file operations

Run tests: `make test`
Coverage report: `make test-cover && open coverage.html`

## Environment Variables

Variables are checked in precedence order (first match wins):

| Setting | Precedence |
|---------|------------|
| Instance URL | `SFDC_INSTANCE_URL` → `SALESFORCE_INSTANCE_URL` → config |
| Client ID | `SFDC_CLIENT_ID` → `SALESFORCE_CLIENT_ID` → config |
| Access Token | `SFDC_ACCESS_TOKEN` (direct, bypasses OAuth) |

## Commit Conventions

Use conventional commits:

```
type(scope): description

feat(query): add SOQL query command
fix(auth): handle token refresh errors
docs(readme): add installation instructions
```

| Prefix | Purpose | Triggers Release? |
|--------|---------|-------------------|
| `feat:` | New features | Yes |
| `fix:` | Bug fixes | Yes |
| `docs:` | Documentation only | No |
| `test:` | Adding/updating tests | No |
| `refactor:` | Code changes that don't fix bugs or add features | No |
| `chore:` | Maintenance tasks | No |
| `ci:` | CI/CD changes | No |

## CI & Release Workflow

Releases are automated with a dual-gate system:

**Gate 1 - Path filter:** Only triggers when Go code changes (`**.go`, `go.mod`, `go.sum`)
**Gate 2 - Commit prefix:** Only `feat:` and `fix:` commits create releases

This means:
- `feat: add command` + Go files changed → release
- `fix: handle edge case` + Go files changed → release
- `docs:`, `ci:`, `test:`, `refactor:` → no release

## Dependencies

Key dependencies:
- `github.com/spf13/cobra` - CLI framework
- `github.com/fatih/color` - Colored terminal output
- `golang.org/x/oauth2` - OAuth 2.0 client
- `github.com/stretchr/testify` - Testing assertions

## Salesforce API

Base URL pattern: `https://{instance}.my.salesforce.com/services/data/vXX.0/`

Key endpoints:
- `/services/data/` - API versions
- `/services/data/vXX.0/query?q=SELECT...` - SOQL queries
- `/services/data/vXX.0/sobjects/` - Object operations
- `/services/data/vXX.0/sobjects/{Object}/describe` - Object metadata
