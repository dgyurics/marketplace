#!/bin/bash

# Local volume restore script
# - Restores exactly one Docker volume from a selected backup directory

set -euo pipefail

print_info() { echo "[INFO] $1"; }
print_error() { echo "[ERROR] $1"; }

if [ "$#" -ne 2 ]; then
    print_error "Usage: $0 <docker-volume-name> <backup-directory>"
    exit 1
fi

VOLUME="$1"
BACKUP_DIR="$2"
ARCHIVE_PATH="$BACKUP_DIR/${VOLUME}.tar.gz"

if [ -z "$VOLUME" ]; then
    print_error "Volume name cannot be empty."
    exit 1
fi

if [ ! -d "$BACKUP_DIR" ]; then
    print_error "Backup directory not found: '$BACKUP_DIR'"
    exit 1
fi

if [ ! -f "$ARCHIVE_PATH" ]; then
    print_error "Backup archive not found: '$ARCHIVE_PATH'"
    exit 1
fi

# Ensure volume exists before restore
docker volume create "$VOLUME" >/dev/null

print_info "Restoring volume '$VOLUME' from: $ARCHIVE_PATH"
docker run --rm \
    -v "$VOLUME":/target \
    -v "$BACKUP_DIR":/backup:ro \
    busybox \
    sh -c "find /target -mindepth 1 -delete && tar xzf '/backup/${VOLUME}.tar.gz' -C /target" || {
    print_error "Failed to restore volume: $VOLUME"
    exit 1
}

print_info "Restore complete."