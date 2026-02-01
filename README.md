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

### Linux (Snap)

```bash
sudo snap install ocli-sfdc
```

### Linux (deb/rpm)

Download the appropriate package from the [releases page](https://github.com/open-cli-collective/salesforce-cli/releases).

### From Source

```bash
go install github.com/open-cli-collective/salesforce-cli/cmd/sfdc@latest
```

## Quick Start

```bash
# Set up OAuth authentication
sfdc init

# Query data
sfdc query "SELECT Id, Name FROM Account LIMIT 5"

# Search across objects
sfdc search "Acme"

# Get a record
sfdc record get Account 001xx000003DGbYAAW

# Check org limits
sfdc limits
```

## Configuration

The CLI stores configuration in `~/.config/salesforce-cli/config.json` and OAuth tokens in your system keychain (macOS Keychain, Linux secret-tool) or a secure file fallback.

### Environment Variables

| Variable | Description |
|----------|-------------|
| `SFDC_INSTANCE_URL` | Salesforce instance URL |
| `SFDC_CLIENT_ID` | Connected App consumer key |
| `SFDC_ACCESS_TOKEN` | Direct access token (bypasses OAuth) |

### Commands

```bash
sfdc config show   # Display current configuration
sfdc config test   # Verify API connectivity
sfdc config clear  # Remove stored credentials
```

## Global Flags

All commands support these flags:

| Flag | Description |
|------|-------------|
| `-o, --output` | Output format: `table`, `json`, `plain` (default: `table`) |
| `--no-color` | Disable colored output |
| `-v, --verbose` | Enable verbose output |
| `--api-version` | Salesforce API version (default: `v62.0`) |

## Commands

### Query & Search

#### SOQL Query

```bash
# Basic query
sfdc query "SELECT Id, Name FROM Account LIMIT 10"

# Include deleted/archived records
sfdc query "SELECT Id, Name FROM Account" --all

# Fetch all pages (large datasets)
sfdc query "SELECT Id, Name FROM Contact" --no-limit

# JSON output
sfdc query "SELECT Id, Name, Phone FROM Contact" -o json
```

#### SOSL Search

```bash
# Simple search
sfdc search "Acme"

# Limit to specific objects
sfdc search "John Smith" --in Account,Contact

# Specify return fields
sfdc search "test" --returning "Account(Id,Name),Contact(Id,FirstName,LastName)"

# Full SOSL syntax
sfdc search "FIND {Acme} IN ALL FIELDS RETURNING Account(Id,Name)"
```

### Records

```bash
# Get a record
sfdc record get Account 001xx000003DGbYAAW
sfdc record get Contact 003xx000001abcd --fields Name,Email,Phone

# Create a record
sfdc record create Account --set Name="Acme Corp"
sfdc record create Contact --set FirstName=John --set LastName=Doe --set Email=john@example.com

# Update a record
sfdc record update Account 001xx000003DGbYAAW --set Name="New Name"
sfdc record update Contact 003xx000001abcd --set Phone="555-1234" --set Email=new@example.com

# Delete a record
sfdc record delete Account 001xx000003DGbYAAW --confirm
```

### Objects

```bash
# List all objects
sfdc object list
sfdc object list --custom-only

# Describe object metadata
sfdc object describe Account

# List fields
sfdc object fields Account
sfdc object fields Account --required-only
```

### Org Limits

```bash
# Show all limits
sfdc limits

# Show specific limit
sfdc limits --show DailyApiRequests
```

### Bulk API 2.0

For large data operations (thousands or millions of records).

#### Import

```bash
# Insert records
sfdc bulk import Account --file accounts.csv --operation insert

# Update records (requires Id column)
sfdc bulk import Account --file accounts.csv --operation update

# Upsert with external ID
sfdc bulk import Contact --file contacts.csv --operation upsert --external-id Email

# Delete records (requires Id column)
sfdc bulk import Account --file delete-ids.csv --operation delete

# Wait for completion
sfdc bulk import Account --file accounts.csv --operation insert --wait
```

#### Export

```bash
# Export to stdout
sfdc bulk export "SELECT Id, Name, Industry FROM Account"

# Export to file
sfdc bulk export "SELECT Id, Name FROM Account" --output accounts.csv
sfdc bulk export "SELECT * FROM Contact" --output contacts.csv
```

#### Job Management

```bash
# List recent jobs
sfdc bulk job list

# Check job status
sfdc bulk job status 750xx000000001

# Get successful results
sfdc bulk job results 750xx000000001
sfdc bulk job results 750xx000000001 --output results.csv

# Get failed records
sfdc bulk job errors 750xx000000001
sfdc bulk job errors 750xx000000001 --output errors.csv

# Abort a job
sfdc bulk job abort 750xx000000001
```

### Apex (Tooling API)

#### List & Get Source

```bash
# List Apex classes
sfdc apex list
sfdc apex list --triggers

# Get source code
sfdc apex get MyController
sfdc apex get MyController --output MyController.cls
sfdc apex get MyTrigger --trigger
```

#### Execute Anonymous Apex

```bash
# From argument
sfdc apex execute "System.debug('Hello');"

# From file
sfdc apex execute --file script.apex

# From stdin
echo "System.debug(UserInfo.getUserName());" | sfdc apex execute -
```

#### Run Tests

```bash
# Run all tests in a class
sfdc apex test --class MyControllerTest

# Run specific method
sfdc apex test --class MyControllerTest --method testCreate

# Wait for completion
sfdc apex test --class MyTest --wait
```

### Debug Logs

```bash
# List recent logs
sfdc log list
sfdc log list --limit 20
sfdc log list --user 005xxx

# Get log content
sfdc log get 07L1x000000ABCD
sfdc log get 07L1x000000ABCD --output debug.log

# Stream new logs (Ctrl+C to stop)
sfdc log tail
sfdc log tail --user 005xxx
sfdc log tail --interval 5
```

### Code Coverage

```bash
# Show all coverage
sfdc coverage

# Coverage for specific class
sfdc coverage --class MyController

# Fail if below threshold
sfdc coverage --min 75
```

### Metadata API

Basic metadata operations. For complex workflows, use the official Salesforce CLI (sf).

```bash
# List available metadata types
sfdc metadata types

# List components of a type
sfdc metadata list --type ApexClass
sfdc metadata list --type ApexTrigger

# Retrieve components
sfdc metadata retrieve --type ApexClass --output ./src
sfdc metadata retrieve --type ApexClass --name MyController --output ./src

# Deploy from directory
sfdc metadata deploy --source ./src
sfdc metadata deploy --source ./src --check-only
sfdc metadata deploy --source ./src --test-level RunLocalTests
sfdc metadata deploy --source ./src --wait
```

### Shell Completion

```bash
# Bash
source <(sfdc completion bash)

# Zsh
sfdc completion zsh > "${fpath[1]}/_sfdc"

# Fish
sfdc completion fish | source

# PowerShell
sfdc completion powershell | Out-String | Invoke-Expression
```

## License

MIT License - see [LICENSE](LICENSE) for details.
