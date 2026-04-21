#!/usr/bin/env bash
# Build the WASM bundle then serve it on all interfaces so phones on the
# same WiFi can reach it. Default port 8080; override with first argument.
set -euo pipefail

cd "$(dirname "$0")/.."

echo "Building WASM..."
./web/build.sh

PORT="${1:-8082}"

echo ""
echo "Listening on:"
echo "  http://localhost:${PORT}  (this machine)"
# Print every non-loopback IPv4 so the user knows which address to type on their phone.
ifconfig 2>/dev/null | awk '/inet / && !/127\.0\.0\.1/ { print "  http://" $2 ":'${PORT}'  (wifi)" }' \
  || ip -4 addr 2>/dev/null | awk '/inet/ && !/127\.0\.0\.1/ { split($2,a,"/"); print "  http://" a[1] ":'${PORT}'  (wifi)" }'
echo ""

python3 -c "
import http.server, socketserver, sys, os
port = int(sys.argv[1])
os.chdir(sys.argv[2])
class H(http.server.SimpleHTTPRequestHandler):
    def end_headers(self):
        self.send_header('Cache-Control', 'no-store')
        super().end_headers()
    def log_message(self, fmt, *args):
        pass  # suppress per-request noise
H.extensions_map['.wasm'] = 'application/wasm'
socketserver.TCPServer.allow_reuse_address = True
with socketserver.TCPServer(('', port), H) as s:
    s.serve_forever()
" "$PORT" "web"
