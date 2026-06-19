# FormaTeX CLI installer for Windows
# Usage: iwr https://raw.githubusercontent.com/formatexio/cli/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$Repo = "formatexio/cli"
$BinName = "formatex"
$Asset = "formatex-windows-amd64.exe"
$ReleaseUrl = "https://github.com/$Repo/releases/latest/download/$Asset"

# Default install location: %LOCALAPPDATA%\formatex\
$InstallDir = Join-Path $env:LOCALAPPDATA "formatex"
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

$Dest = Join-Path $InstallDir "$BinName.exe"

Write-Host "Downloading FormaTeX CLI..."
Invoke-WebRequest -Uri $ReleaseUrl -OutFile $Dest -UseBasicParsing

# Add to user PATH if not already there
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$UserPath;$InstallDir", "User")
    Write-Host "Added $InstallDir to your PATH."
    Write-Host "Restart your terminal for the PATH change to take effect."
}

Write-Host "FormaTeX CLI installed to $Dest"
Write-Host "Run 'formatex login' to get started."
