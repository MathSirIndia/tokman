#!/usr/bin/env bash
set -eo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

echo "================================================================="
echo "   TOKMAN AI GATEWAY MESH: PRODUCTION NATIVE BINARY COMPILER     "
echo "================================================================="

cd "$ROOT_DIR"

# Auto-detect Go in development directory
if ! command -v go >/dev/null 2>&1; then
    if [ -d "$HOME/development/go/bin" ]; then
        export PATH="$HOME/development/go/bin:$HOME/go/bin:$PATH"
    fi
fi

# Ensure user-level native library symlinks are linked
mkdir -p "$HOME/.local/lib"
if [ -f "/usr/lib/x86_64-linux-gnu/libXxf86vm.so.1" ] && [ ! -f "$HOME/.local/lib/libXxf86vm.so" ]; then
    ln -sf /usr/lib/x86_64-linux-gnu/libXxf86vm.so.1 "$HOME/.local/lib/libXxf86vm.so"
fi

export LIBRARY_PATH="$HOME/.local/lib:$LIBRARY_PATH"
export CGO_LDFLAGS="-L$HOME/.local/lib $CGO_LDFLAGS"

mkdir -p build/bin

echo "--> Compiling standalone native Go binary (Gateway + Embedded SQLite + Fyne GUI)..."
go build -trimpath -ldflags="-s -w" -o build/bin/tokman .

echo "--> Build complete. Binary located at: build/bin/tokman"
ls -lh build/bin/tokman
