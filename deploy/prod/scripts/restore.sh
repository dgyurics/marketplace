#!/bin/bash

# =============================================================================
# Production Data Restore Script
# =============================================================================

set -euo pipefail  # Exit on errors, undefined variables, and pipe failures

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
print_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Get backup directory
if [ -z "${1:-}" ]; then
    echo "Enter path to backup directory:"
    read -r BACKUP_DIR
else
    BACKUP_DIR="$1"
fi

# Validate backup directory
if [ ! -d "$BACKUP_DIR" ]; then
    print_error "Backup directory '$BACKUP_DIR' does not exist!"
    exit 1
fi

# Check if backup files exist
if ! ls "$BACKUP_DIR"/*.tar.gz >/dev/null 2>&1; then
    print_error "No backup files (*.tar.gz) found in '$BACKUP_DIR'!"
    exit 1
fi

print_info "Found backup files in: $BACKUP_DIR"
print_info "Backup files:"
ls -la "$BACKUP_DIR"/*.tar.gz

echo
print_warn "WARNING: This will OVERWRITE existing Docker volume data!"
print_warn "Make sure your containers are stopped before proceeding."
echo
echo -n "Do you want to continue? (y/N): "
read -r confirm

if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
    print_info "Restore cancelled."
    exit 0
fi

echo
print_info "Starting restore process..."

# Counter for restored volumes
restored_count=0

# Restore each backup file
for backup_file in "$BACKUP_DIR"/*.tar.gz; do
    volume_name=$(basename "$backup_file" .tar.gz)
    print_info "Restoring volume: $volume_name"
    
    # Create volume if it doesn't exist (suppress output but not errors)
    if docker volume create "$volume_name" >/dev/null 2>&1; then
        print_info "Created new volume: $volume_name"
    else
        print_info "Using existing volume: $volume_name"
    fi
    
    # Restore volume data
    docker run --rm \
        -v "$volume_name":/target \
        -v "$BACKUP_DIR":/backup:ro \
        busybox \
        sh -c "cd /target && rm -rf ./* && tar xzf /backup/$(basename "$backup_file") -C ."
    
    print_info "✓ Volume '$volume_name' restored successfully"
    ((restored_count++))
done

echo
print_info "Restore completed successfully!"
print_info "Restored $restored_count volumes"
print_info "You can now start your containers."

# Show volume info
echo
print_info "Volume information:"
for backup_file in "$BACKUP_DIR"/*.tar.gz; do
    volume_name=$(basename "$backup_file" .tar.gz)
    if docker volume inspect "$volume_name" >/dev/null 2>&1; then
        mount_point=$(docker volume inspect "$volume_name" --format '{{.Mountpoint}}' 2>/dev/null || echo "N/A")
        echo "  - $volume_name: $mount_point"
    fi
done