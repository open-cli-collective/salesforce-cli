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
   - Runs `render.ps1` to write literal AMD64/ARM64 release URLs and checksums into `chocolateyInstall.ps1`
   - Updates the nuspec version
   - Packs and pushes to Chocolatey

2. The URLs and checksums in `chocolateyInstall.ps1` start as placeholders and are replaced before packing.

## Manual Publishing

If you need to publish manually:

```powershell
cd packaging/chocolatey
$version = '0.1.0'
$amd64Hash = (Get-Content checksums.txt | Select-String 'windows_amd64.zip').Line.Split()[0]
$arm64Hash = (Get-Content checksums.txt | Select-String 'windows_arm64.zip').Line.Split()[0]
pwsh ./render.ps1 -Version $version -Amd64Checksum $amd64Hash -Arm64Checksum $arm64Hash

choco pack
choco push "salesforce-cli.$version.nupkg" --source https://push.chocolatey.org/ --api-key YOUR_API_KEY
```

## Testing Locally

```powershell
cd packaging/chocolatey
$version = '0.1.0'
$amd64Hash = (Get-Content checksums.txt | Select-String 'windows_amd64.zip').Line.Split()[0]
$arm64Hash = (Get-Content checksums.txt | Select-String 'windows_arm64.zip').Line.Split()[0]
pwsh ./render.ps1 -Version $version -Amd64Checksum $amd64Hash -Arm64Checksum $arm64Hash
choco pack
choco install salesforce-cli -s . --pre
```
