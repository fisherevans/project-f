#!/bin/bash
# Primortal - macOS Launch Helper
# This script removes the quarantine attribute to bypass Gatekeeper warnings

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GAME_BINARY="$SCRIPT_DIR/primortal"

echo "Primortal - macOS Launch Helper"
echo "================================"
echo ""

# Check if binary exists
if [ ! -f "$GAME_BINARY" ]; then
    echo "❌ Error: primortal binary not found at: $GAME_BINARY"
    exit 1
fi

# Remove quarantine attribute
echo "Removing quarantine attribute from game files..."
xattr -dr com.apple.quarantine "$SCRIPT_DIR"

if [ $? -eq 0 ]; then
    echo "✅ Quarantine removed successfully"
    echo ""
    echo "Launching Primortal..."
    echo ""
    "$GAME_BINARY"
else
    echo "⚠️  Failed to remove quarantine. You may need to run:"
    echo "   Right-click primortal → Open"
    echo ""
    read -p "Press Enter to try launching anyway..."
    "$GAME_BINARY"
fi
