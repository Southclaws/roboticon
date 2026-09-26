#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
mkdir -p web/dist
GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o web/dist/roboticon.wasm ./cmd/wasm
goroot=$(go env GOROOT)
support="$goroot/lib/wasm/wasm_exec.js"
if [ ! -f "$support" ]; then support="$goroot/misc/wasm/wasm_exec.js"; fi
cp "$support" web/dist/wasm_exec.js
cp web/index.html web/style.css web/app.js web/worker.js web/social.png web/dist/
touch web/dist/.nojekyll
