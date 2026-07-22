#!/usr/bin/env bash

set -euo pipefail

REPO="parmeet20/dockcode"
BINARY="dockcode"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"

case "$OS" in
    linux)
        EXT="tar.gz"
        ;;
    darwin)
        EXT="tar.gz"
        ;;
    *)
        echo "❌ Unsupported operating system: $OS"
        exit 1
        ;;
esac

# Detect architecture
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "❌ Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

echo "🔍 Fetching latest DockCode release..."

VERSION="${VERSION:-$(
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' \
    | sed -E 's/.*"([^"]+)".*/\1/'
)}"

if [ -z "$VERSION" ]; then
    echo "❌ Failed to determine latest release."
    exit 1
fi

ARCHIVE="dockcode_${OS}_${ARCH}.${EXT}"

URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

echo "📦 Downloading ${ARCHIVE}..."

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

if ! curl -fLsS "$URL" -o "$TMP_DIR/archive.${EXT}"; then
    echo ""
    echo "❌ Failed to download release asset."
    echo "Expected:"
    echo "$URL"
    exit 1
fi

echo "📂 Extracting archive..."

tar -xzf "$TMP_DIR/archive.${EXT}" -C "$TMP_DIR"

if [ ! -f "$TMP_DIR/$BINARY" ]; then
    echo "❌ Binary '$BINARY' not found inside archive."
    exit 1
fi

if [ -w "$INSTALL_DIR" ]; then
    SUDO=""
else
    SUDO="sudo"
fi

echo "🚀 Installing to ${INSTALL_DIR}..."

$SUDO install -m 755 "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"

echo ""
echo "✅ DockCode installed successfully!"
echo ""
echo "Version : $VERSION"
echo "Location: ${INSTALL_DIR}/${BINARY}"
echo ""
echo "Run:"
echo "  dockcode --help"