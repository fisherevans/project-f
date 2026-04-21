#!/usr/bin/env bash
# Serve the web/ directory on localhost:8080 with correct WASM MIME type.
set -euo pipefail
cd "$(dirname "$0")"
PORT="${1:-8090}"
echo "Serving http://localhost:${PORT}"
python3 -c "
import http.server, socketserver, sys
port = int(sys.argv[1])
class H(http.server.SimpleHTTPRequestHandler):
    def end_headers(self):
        self.send_header('Cache-Control', 'no-store')
        super().end_headers()
H.extensions_map['.wasm'] = 'application/wasm'
with socketserver.TCPServer(('', port), H) as s:
    s.serve_forever()
" "$PORT"
