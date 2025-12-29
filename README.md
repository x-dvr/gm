# gm

A simple Go version manager that can help you install and manage multiple versions of Go on your system.

## Features

- Install multiple Go versions side-by-side
- Switch between installed versions

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/x-dvr/gm/master/install.sh | bash
```

## Usage

### Install Go

Install the latest version of Go:

```bash
gm install latest
```

Install a specific version:

```bash
gm install 1.22.0
# or
gm install go1.22.0
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

Configure your shell to use the current Go version:

```bash
eval $(gm env)
```

Installation script adds this call to your shell profile (`.bashrc`, `.zshenv`, etc.) to automatically set up the environment on new shell sessions.

## Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `gm install <version>` | `gm i <version>` | Install a specific Go version |
| `gm use <version>` | - | Set a version as current |
| `gm list` | `gm ls` | List all installed versions |
| `gm env` | - | Output shell commands to set environment variables |
| `gm upgrade` | `gm up` | Upgrade gm to the latest version |
