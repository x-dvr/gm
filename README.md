# gm - Go version manager

A simple Go version manager that can help you install and manage multiple versions of Go on your system.

## Features

- Install multiple Go versions side-by-side
- Switch between installed versions

## Supported Platforms

Legend:

✅ - Supported and tested by maintainer

ℹ️ - Expected to work, but not actively tested

| OS | Architecture | Shell | Status |
|----|--------------|-------|--------|
| Linux | amd64 | zsh | ✅ |
| Linux | amd64 | bash, fish | ℹ️ |
| Linux | arm64 | zsh, bash, fish | ℹ️ |
| macOS | amd64 | zsh, bash, fish | ℹ️ |
| macOS | arm64 | zsh | ℹ️ |
| macOS | arm64 | bash | ℹ️ |
| macOS | arm64 | fish | ℹ️ |
| Windows | amd64 | PowerShell* | ✅ |
| Windows | arm64 | PowerShell* | ℹ️ |

\* Due to name conflicts this tool must be executed as `gm.exe` in Powershell.

## Installation

### Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/x-dvr/gm/master/install.sh | bash
```

### Windows

```powershell
irm https://raw.githubusercontent.com/x-dvr/gm/master/install.ps1 | iex
```

## Usage

### Install Go

Install the latest version of Go:

```bash
gm install latest
# or
gm i
```

Install a specific version:

```bash
gm install 1.22.0
# or
gm i go1.22.0
```

### Switch Go Version

Set a specific version as current:

```bash
gm use 1.22.0
# or
gm use latest
```

### List Installed Versions

View all installed Go versions:

```bash
gm list
# or
gm ls
```

The current version will be marked with a check mark.

### Upgrade gm

Update gm to the latest version:

```bash
gm upgrade
# or
gm up
```

### Set Environment Variables

Configure your shell to use the current Go version (set environment variables):

**Linux / macOS:**
```bash
eval $(gm env)
```

**Windows:**
```powershell
gm env
```

Installation script automatically adds this command to your shell profile (`.bashrc`, `.zshenv`, etc.) on unix-like systems to set up the environment on new shell sessions. On Windows this command is executed once in installation script to setup user-scoped environment variables.

## Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `gm install <version>` | `gm i <version>` | Install a specific Go version |
| `gm use <version>` | - | Set a version as current |
| `gm list` | `gm ls` | List all installed versions |
| `gm env` | - | Output shell commands to set environment variables |
| `gm upgrade` | `gm up` | Upgrade gm to the latest version |
