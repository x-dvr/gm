# GM (Go version manager) installation script for Windows
# Usage: irm https://raw.githubusercontent.com/x-dvr/gm/master/install.ps1 | iex

$ErrorActionPreference = "Stop"

# Configuration
$REPO = "x-dvr/gm"
$INSTALL_DIR = Join-Path $env:USERPROFILE ".gm\bin"
$BINARY_NAME = "gm.exe"
$EXT = "zip"

# Detect architecture
$ARCH = $env:PROCESSOR_ARCHITECTURE
switch ($ARCH) {
    "AMD64" { $ARCH = "amd64" }
    "ARM64" { $ARCH = "arm64" }
    default {
        Write-Host "❌ Unsupported architecture: $ARCH" -ForegroundColor Red
        exit 1
    }
}

$OS = "windows"
Write-Host "Platform detected: $OS/$ARCH" -ForegroundColor Cyan

# Get latest release
Write-Host "📦 Fetching latest release..." -ForegroundColor Cyan
try {
    $LATEST_RELEASE = Invoke-RestMethod -Uri "https://api.github.com/repos/$REPO/releases/latest"
    $VERSION = $LATEST_RELEASE.tag_name
} catch {
    Write-Host "❌ Failed to fetch latest version: $_" -ForegroundColor Red
    exit 1
}

if (-not $VERSION) {
    Write-Host "❌ Failed to parse version from release" -ForegroundColor Red
    exit 1
}

Write-Host "📌 Latest version: $VERSION" -ForegroundColor Green

# Download
$ARCHIVE_NAME = "${BINARY_NAME}_${OS}.${ARCH}.${EXT}"
$DOWNLOAD_URL = "https://github.com/$REPO/releases/download/$VERSION/$ARCHIVE_NAME"

Write-Host "⬇️  Downloading from: $DOWNLOAD_URL" -ForegroundColor Cyan

$TMP_DIR = Join-Path $env:TEMP "gm-install-$(Get-Random)"
New-Item -ItemType Directory -Path $TMP_DIR -Force | Out-Null
$ARCHIVE_PATH = Join-Path $TMP_DIR $ARCHIVE_NAME

try {
    Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile $ARCHIVE_PATH -UseBasicParsing
} catch {
    Write-Host "❌ Failed to download: $_" -ForegroundColor Red
    Remove-Item -Path $TMP_DIR -Recurse -Force -ErrorAction SilentlyContinue
    exit 1
}

# Extract
Write-Host "📂 Extracting archive..." -ForegroundColor Cyan
try {
    Expand-Archive -Path $ARCHIVE_PATH -DestinationPath $TMP_DIR -Force
} catch {
    Write-Host "❌ Failed to extract archive: $_" -ForegroundColor Red
    Remove-Item -Path $TMP_DIR -Recurse -Force -ErrorAction SilentlyContinue
    exit 1
}

# Install
Write-Host "📥 Installing to $INSTALL_DIR\$BINARY_NAME..." -ForegroundColor Cyan
if (-not (Test-Path $INSTALL_DIR)) {
    New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
}

$SOURCE_BINARY = Join-Path $TMP_DIR $BINARY_NAME
$DEST_BINARY = Join-Path $INSTALL_DIR $BINARY_NAME

# Stop if file is in use and retry
$retries = 3
for ($i = 0; $i -lt $retries; $i++) {
    try {
        Copy-Item -Path $SOURCE_BINARY -Destination $DEST_BINARY -Force
        break
    } catch {
        if ($i -eq ($retries - 1)) {
            Write-Host "❌ Failed to copy binary: $_" -ForegroundColor Red
            Remove-Item -Path $TMP_DIR -Recurse -Force -ErrorAction SilentlyContinue
            exit 1
        }
        Start-Sleep -Seconds 1
    }
}

# Cleanup
Remove-Item -Path $TMP_DIR -Recurse -Force -ErrorAction SilentlyContinue

# Update PATH if needed
$CurrentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($CurrentPath -notlike "*$INSTALL_DIR*") {
    Write-Host ""
    Write-Host "⚙️  Adding $INSTALL_DIR to PATH..." -ForegroundColor Cyan

    $NewPath = "$INSTALL_DIR;$CurrentPath"
    [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")

    # Update current session PATH
    $env:Path = "$INSTALL_DIR;$env:Path"

    Write-Host "✅ Updated PATH environment variable" -ForegroundColor Green
} else {
    Write-Host "ℹ️  PATH already contains $INSTALL_DIR" -ForegroundColor Yellow
}

# Run gm env to set up Go environment variables
Write-Host ""
Write-Host "⚙️  Configuring Go environment variables..." -ForegroundColor Cyan
try {
    & "$DEST_BINARY" env
} catch {
    Write-Host "⚠️  Could not configure environment variables: $_" -ForegroundColor Yellow
    Write-Host "   You can run 'gm env' manually later" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "✅ Successfully installed $BINARY_NAME $VERSION" -ForegroundColor Green
Write-Host ""
Write-Host "Run 'gm --version' to verify installation" -ForegroundColor Cyan
Write-Host "Note: You may need to restart your terminal for PATH changes to take effect" -ForegroundColor Yellow
