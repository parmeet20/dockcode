#!/usr/bin/env bash
set -euo pipefail

REPO="parmeet20/dockcode"
BINARY="dockcode"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case $ARCH in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported arch: $ARCH"; exit 1 ;;
esac

VERSION="${VERSION:-$(curl -sf https://api.github.com/repos/${REPO}/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')}"

EXT="tar.gz"
[ "$OS" = "windows" ] && EXT="zip"

URL="https://github.com/${REPO}/releases/download/${VERSION}/dockercode_${VERSION}_${OS}_${ARCH}.${EXT}"

echo "Downloading DockCode ${VERSION} for ${OS}/${ARCH}..."
TMP=$(mktemp -d)
curl -sL "$URL" -o "$TMP/archive.$EXT"

if [ "$EXT" = "zip" ]; then
  unzip -q "$TMP/archive.$EXT" -d "$TMP"
else
  tar -xzf "$TMP/archive.$EXT" -C "$TMP"
fi

install -m 755 "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
rm -rf "$TMP"

echo "✓ DockCode installed to $INSTALL_DIR/$BINARY"
echo "  Run: dockcode"
