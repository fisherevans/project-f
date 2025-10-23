#!/bin/zsh

# Get the directory where this script is located
SCRIPT_DIR="${0:A:h}"

# Destination directory
DEST_DIR="/Users/fisher/Library/Preferences/Tiled/extensions"

# Create destination directory if it doesn't exist
mkdir -p "$DEST_DIR"

# Copy all .js files from the script's directory to the destination
cp "$SCRIPT_DIR"/*.js "$DEST_DIR"

echo "Copied all .js files from $SCRIPT_DIR to $DEST_DIR"
