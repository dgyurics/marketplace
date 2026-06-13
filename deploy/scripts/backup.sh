#!/bin/bash

# Local volume backup script
# - Backs up exactly one Docker volume passed as an argument

set -euo pipefail

print_info() { echo "[INFO] $1"; }
print_error() { echo "[ERROR] $1"; }

if [ "$#" -ne 1 ]; then
    print_error "Usage: $0 <docker-volume-name>"
    exit 1
fi

VOLUME="$1"

BACKUP_DEST="$PWD"

if [ ! -w "$BACKUP_DEST" ]; then
    print_error "Current directory is not writable: '$BACKUP_DEST'."
    exit 1
fi

if [ -z "$VOLUME" ]; then
    print_error "Volume name cannot be empty."
    exit 1
fi

if ! docker volume inspect "$VOLUME" >/dev/null 2>&1; then
    print_error "Volume '$VOLUME' not found."
    exit 1
fi

TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_DIR="$BACKUP_DEST/marketplace-local-backup_$TIMESTAMP"
mkdir -p "$BACKUP_DIR"

print_info "Starting backup to: $BACKUP_DIR"
print_info "Backing up volume: $VOLUME"
docker run --rm \
    -v "$VOLUME":/source:ro \
    -v "$BACKUP_DIR":/backup \
    busybox \
    tar czf "/backup/${VOLUME}.tar.gz" -C /source --exclude='lost+found' . || {
    print_error "Failed to backup volume: $VOLUME"
    exit 1
}

print_info "Backup complete: $BACKUP_DIR"