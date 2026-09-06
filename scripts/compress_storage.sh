#!/bin/sh
set -eu

LOG_DIR="/app/data/logs"
ARCHIVE_DIR="/app/data/logs/archive"
mkdir -p "$ARCHIVE_DIR"

echo "--> [1/2] Compacting log files older than 30 days with Zstandard Level 10..."
if [ -d "$LOG_DIR" ]; then
    find "$LOG_DIR" -maxdepth 1 -type f -name "*.log" -mtime +30 -exec sh -c '
        for file do
            if command -v zstd >/dev/null 2>&1; then
                zstd -10 --rm "$file" -o "'"$ARCHIVE_DIR"'/$(basename "$file").zst"
            else
                gzip -9 "$file" && mv "${file}.gz" "'"$ARCHIVE_DIR"'/"
            fi
        done
    ' sh {} +
fi

echo "--> [2/2] Pruning PostgreSQL spend records older than 90 days..."
if [ -n "${DATABASE_URL:-}" ] && command -v psql >/dev/null 2>&1; then
    psql "$DATABASE_URL" -c "DELETE FROM \"LiteLLM_SpendLogs\" WHERE \"startTime\" < NOW() - INTERVAL '90 DAYS';"
else
    echo "Notice: DATABASE_URL or psql client not detected in local environment; skipping direct database prune."
fi

echo "Storage hygiene routine complete."
