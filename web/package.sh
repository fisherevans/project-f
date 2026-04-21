#!/usr/bin/env bash
# Package web/ contents into a zip for itch.io upload.
set -euo pipefail

cd "$(dirname "$0")/.."

./web/build.sh

OUT=web/primortal-web.zip
rm -f "$OUT"

(cd web && zip -9 "../$OUT" index.html wasm_exec.js main.wasm build_info.js)

echo "Packaged $OUT ($(du -h "$OUT" | cut -f1))"
