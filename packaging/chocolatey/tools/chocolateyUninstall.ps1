$ErrorActionPreference = 'Stop'

$toolsDir = Split-Path -Parent $MyInvocation.MyCommand.Definition

Write-Host "Uninstalling sfdc..."

# Remove extracted files
Remove-Item "$toolsDir\sfdc.exe" -Force -ErrorAction SilentlyContinue
Remove-Item "$toolsDir\LICENSE" -Force -ErrorAction SilentlyContinue
Remove-Item "$toolsDir\README.md" -Force -ErrorAction SilentlyContinue

# Remove .ignore files created during install
Remove-Item "$toolsDir\*.ignore" -Force -ErrorAction SilentlyContinue

Write-Host "sfdc has been uninstalled."
