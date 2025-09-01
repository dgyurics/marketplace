#!/bin/bash

# =============================================================================
# Production Data Backup Script
# =============================================================================

# Get backup destination
if [ -z "$1" ]; then
    echo "Enter backup destination (e.g., /media/usb-drive, /backup):"
    read -r BACKUP_DEST
else
    BACKUP_DEST="$1"
fi

# Validate backup destination
if [ ! -d "$BACKUP_DEST" ]; then
    print_error "Backup destination '$BACKUP_DEST' does not exist!"
    exit 1
fi

if [ ! -w "$BACKUP_DEST" ]; then
    print_error "Cannot write to backup destination '$BACKUP_DEST'!"
    exit 1
fi

# Create timestamped backup directory
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_DIR="$BACKUP_DEST/docker-volumes-backup_$TIMESTAMP"
mkdir -p "$BACKUP_DIR"

print_info "Starting backup to: $BACKUP_DIR"

# Your marketplace volumes
VOLUMES=("marketplace_postgres-data" "marketplace_images-data" "marketplace_ssl-certs")

# Backup each volume
for volume in "${VOLUMES[@]}"; do
    print_info "Backing up volume: $volume"
    
    # Check if volume exists
    if ! docker volume inspect "$volume" >/dev/null 2>&1; then
        print_warn "Volume '$volume' not found, skipping..."
        continue
    fi
    
    # Create backup using docker run with busybox
    docker run --rm \
        -v "$volume":/source:ro \
        -v "$BACKUP_DIR":/backup \
        busybox \
        tar czf "/backup/${volume}.tar.gz" -C /source .
    
    print_info "✓ Volume '$volume' backed up successfully"
done
