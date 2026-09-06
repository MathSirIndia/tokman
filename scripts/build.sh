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

UNAME_S="$(uname -s 2>/dev/null || echo "Unknown")"

# Only apply Linux-specific X11 library symlink fixes when running on Linux
if [ "$UNAME_S" = "Linux" ]; then
    mkdir -p "$HOME/.local/lib"
    if [ -f "/usr/lib/x86_64-linux-gnu/libXxf86vm.so.1" ] && [ ! -f "$HOME/.local/lib/libXxf86vm.so" ]; then
        ln -sf /usr/lib/x86_64-linux-gnu/libXxf86vm.so.1 "$HOME/.local/lib/libXxf86vm.so"
    fi
    export LIBRARY_PATH="$HOME/.local/lib:$LIBRARY_PATH"
    export CGO_LDFLAGS="-L$HOME/.local/lib $CGO_LDFLAGS"
fi

BIN_NAME="tokman"
if [ "$OS" = "Windows_NT" ] || [ "${UNAME_S:0:5}" = "MINGW" ] || [ "${UNAME_S:0:4}" = "MSYS" ]; then
    BIN_NAME="tokman.exe"
fi

mkdir -p build/bin

echo "--> Compiling standalone native Go binary on $UNAME_S (Gateway + Embedded SQLite + Fyne GUI)..."
go build -trimpath -ldflags="-s -w" -o "build/bin/$BIN_NAME" .

echo "--> Build complete. Binary located at: build/bin/$BIN_NAME"
ls -lh "build/bin/$BIN_NAME"

