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

### Set Environment Variables

Configure your shell to use the current Go version:

```bash
eval $(gm env)
```

Add this to your shell profile (`.bashrc`, `.zshrc`, etc.) to automatically set up the environment on new shell sessions.

## Quick Start

Get started with the latest Go version in two commands:

```bash
gm install latest
gm use latest
```

Don't forget to configure your shell:

```bash
eval $(gm env)
```

## Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `gm install <version>` | `gm i <version>` | Install a specific Go version |
| `gm use <version>` | - | Set a version as current |
| `gm list` | `gm ls` | List all installed versions |
| `gm env` | - | Output shell commands to set environment variables |
