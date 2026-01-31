# salesforce-cli (sfdc)

A command-line interface for Salesforce.

## Installation

### macOS (Homebrew)

```bash
brew tap open-cli-collective/tap
brew install --cask salesforce-cli
```

### Windows (Chocolatey)

```powershell
choco install salesforce-cli
```

### Windows (Winget)

```powershell
winget install OpenCLICollective.salesforce-cli
```

### Linux (deb/rpm)

Download the appropriate package from the [releases page](https://github.com/open-cli-collective/salesforce-cli/releases).

### From Source

```bash
go install github.com/open-cli-collective/salesforce-cli/cmd/sfdc@latest
```

## Quick Start

1. Set up authentication:
   ```bash
   sfdc init
   ```

2. Query data:
   ```bash
   sfdc query "SELECT Id, Name FROM Account LIMIT 5"
   ```

## Configuration

The CLI stores configuration in `~/.config/salesforce-cli/config.json` and OAuth tokens in your system keychain (macOS Keychain, Linux secret-tool) or a secure file fallback.

### Environment Variables

| Variable | Description |
|----------|-------------|
| `SFDC_INSTANCE_URL` | Salesforce instance URL |
| `SFDC_CLIENT_ID` | Connected App consumer key |
| `SFDC_ACCESS_TOKEN` | Direct access token (bypasses OAuth) |

## Commands

- `sfdc init` - Set up OAuth authentication
- `sfdc config show` - Display current configuration
- `sfdc config test` - Verify API connectivity
- `sfdc config clear` - Remove stored credentials
- `sfdc completion` - Generate shell completion scripts

## License

MIT License - see [LICENSE](LICENSE) for details.
