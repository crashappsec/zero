#!/bin/bash
# RAG Pattern Statistics
# Counts patterns in the rag/ directory

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RAG_DIR="${SCRIPT_DIR}/../rag"

echo "=== Zero RAG Pattern Statistics ==="
echo ""

# Count markdown files
total_patterns=$(find "$RAG_DIR" -name "*.md" -type f | wc -l | tr -d ' ')
echo "Total RAG patterns: $total_patterns"

# Count by category
echo ""
echo "By category:"
for dir in "$RAG_DIR"/*/; do
    if [ -d "$dir" ]; then
        name=$(basename "$dir")
        count=$(find "$dir" -name "*.md" -type f | wc -l | tr -d ' ')
        printf "  %-30s %d\n" "$name" "$count"
    fi
done

echo ""
echo "Last updated: $(date '+%Y-%m-%d %H:%M:%S')"
