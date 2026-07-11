#!/usr/bin/env bash
# Build the ez-reset Wails app for Windows (amd64) from Linux/macOS.
# Requires: Go, the Wails CLI (go install github.com/wailsapp/wails/v2/cmd/wails@latest)
# and the frontend dependencies installed (npm install in ./frontend).
set -euo pipefail

cd "$(dirname "$0")"

echo "==> Ensuring frontend is built"
( cd frontend && npm install && npm run build )

echo "==> Building Windows/amd64 binary"
wails build -platform windows/amd64

echo "==> Done: build/bin/ezreset.exe"
