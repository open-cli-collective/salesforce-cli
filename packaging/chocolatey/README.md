# Chocolatey Package

This directory contains the Chocolatey package configuration for salesforce-cli.

## Package ID

`salesforce-cli`

## Installation

```powershell
choco install salesforce-cli
```

## How Releases Work

1. When a new version is released, the release workflow:
   - Downloads the Windows binaries
   - Extracts SHA256 checksums from `checksums.txt`
   - Updates `chocolateyInstall.ps1` with the checksums
   - Updates the nuspec version
   - Packs and pushes to Chocolatey

2. The checksums in `chocolateyInstall.ps1` are placeholders that get replaced during CI.

## Manual Publishing

If you need to publish manually:

```powershell
# Update version in nuspec
# Update checksums in chocolateyInstall.ps1

cd packaging/chocolatey
choco pack
choco push salesforce-cli.<version>.nupkg --source https://push.chocolatey.org/ --api-key YOUR_API_KEY
```

## Testing Locally

```powershell
cd packaging/chocolatey
choco pack
choco install salesforce-cli -s . --pre
```
