#!/bin/bash
set -e

# Configuration
REPO="x-dvr/gm"
INSTALL_DIR="${HOME}/.gm/bin"
BINARY_NAME="gm"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

EXT="tar.gz"

echo "Platform detected: $OS/$ARCH"

# Get latest release
echo "📦 Fetching latest release..."
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO/releases/latest")
VERSION=$(echo "$LATEST_RELEASE" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
    echo "❌ Failed to fetch latest version"
    exit 1
fi

echo "📌 Latest version: $VERSION"

# Download
ARCHIVE_NAME="${BINARY_NAME}_${OS}.${ARCH}.${EXT}"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/$ARCHIVE_NAME"

echo "⬇️  Downloading from: $DOWNLOAD_URL"

TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

curl -L -o "$TMP_DIR/$ARCHIVE_NAME" "$DOWNLOAD_URL"

# Extract
echo "📂 Extracting archive..."
ls -la "$TMP_DIR"
cd "$TMP_DIR"
tar -xzf "$ARCHIVE_NAME"

# Install
mkdir -p "$INSTALL_DIR"
echo "📥 Installing to $INSTALL_DIR/$BINARY_NAME..."
mv "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

# Update PATH if needed
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo ""
    echo "⚙️  Adding $INSTALL_DIR to PATH..."

    # Detect shell and update appropriate profile
    SHELL_PROFILE=""
    if [ -n "$ZSH_VERSION" ]; then
        SHELL_PROFILE="$HOME/.zshenv"
        # Create .zshenv if it doesn't exist
        if [ ! -f "$SHELL_PROFILE" ]; then
            touch "$SHELL_PROFILE"
        fi
    elif [ -n "$BASH_VERSION" ]; then
        if [ -f "$HOME/.bashrc" ]; then
            SHELL_PROFILE="$HOME/.bashrc"
        elif [ -f "$HOME/.bash_profile" ]; then
            SHELL_PROFILE="$HOME/.bash_profile"
        fi
    fi

    if [ -n "$SHELL_PROFILE" ]; then
        # Check if already added
        if grep -q "# Added by $BINARY_NAME installer" "$SHELL_PROFILE" 2>/dev/null; then
            echo "ℹ️  PATH already configured in $SHELL_PROFILE"
        else
            echo "" >> "$SHELL_PROFILE"
            echo "# Added by $BINARY_NAME installer" >> "$SHELL_PROFILE"
            echo "export PATH=\"\$HOME/.gm/bin:\$PATH\"" >> "$SHELL_PROFILE"
            echo "✅ Updated $SHELL_PROFILE"
            echo "   Run: source $SHELL_PROFILE"
        fi
    else
        echo "⚠️  Could not detect shell profile"
        echo "   Please add manually: export PATH=\"\$HOME/.gm/bin:\$PATH\""
    fi
fi

echo ""
echo "✅ Successfully installed $BINARY_NAME $VERSION"
echo ""
echo "Run '$BINARY_NAME --version' to verify installation"
