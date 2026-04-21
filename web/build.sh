#!/usr/bin/env bash
# Build the WASM binary into web/main.wasm
set -euo pipefail

cd "$(dirname "$0")/.."

GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o web/main.wasm ./cmd/web

# Copy the matching wasm_exec.js for the current toolchain
WASM_EXEC="$(go env GOROOT)/lib/wasm/wasm_exec.js"
if [[ ! -f "$WASM_EXEC" ]]; then
    WASM_EXEC="$(go env GOROOT)/misc/wasm/wasm_exec.js"
fi
rm -f web/wasm_exec.js
cp "$WASM_EXEC" web/wasm_exec.js
chmod u+w web/wasm_exec.js

WASM_SIZE=$(stat -f%z web/main.wasm 2>/dev/null || stat -c%s web/main.wasm)
cat > web/build_info.js <<EOF
window.primortalBuildInfo = { wasmSize: $WASM_SIZE };
EOF

echo "Built web/main.wasm ($(du -h web/main.wasm | cut -f1))"
