#!/bin/bash
# Copyright (c) 2024 Gibson Powers Contributors
# 
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Copyright Header Validation Script
# Checks that all source files have proper copyright headers
# Usage: ./check-copyright.sh [--fix]
#############################################################################

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

FIX_MODE=false
if [[ "$1" == "--fix" ]]; then
    FIX_MODE=true
fi

# Copyright header templates
MD_HEADER='<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->
'

SH_HEADER='# Copyright (c) 2024 Gibson Powers Contributors
# 
# SPDX-License-Identifier: GPL-3.0
'

# Files to exclude
EXCLUDE_PATTERNS=(
    ".git/"
    "node_modules/"
    "*.skill"  # Binary files
    "LICENSE"  # License file itself
    ".DS_Store"
)

# Build exclude arguments for find
EXCLUDE_ARGS=()
for pattern in "${EXCLUDE_PATTERNS[@]}"; do
    EXCLUDE_ARGS+=(-not -path "*/$pattern" -not -path "*$pattern")
done

MISSING_COUNT=0
FIXED_COUNT=0

echo "Checking copyright headers..."
echo ""

# Function to check and optionally fix copyright header
check_file() {
    local file="$1"
    local header_pattern="$2"
    local fix_header="$3"

    if ! head -n 10 "$file" | grep -q "Copyright (c) 2024 Gibson Powers Contributors"; then
        if [[ "$FIX_MODE" == true ]]; then
            echo -e "${YELLOW}Fixing: $file${NC}"

            # Handle shebang for shell scripts
            if [[ "$file" == *.sh ]]; then
                if head -n 1 "$file" | grep -q "^#!/"; then
                    shebang=$(head -n 1 "$file")
                    tail -n +2 "$file" > "$file.tmp"
                    echo "$shebang" > "$file.new"
                    echo "$fix_header" >> "$file.new"
                    cat "$file.tmp" >> "$file.new"
                    mv "$file.new" "$file"
                    rm "$file.tmp"
                else
                    echo "$fix_header$(cat "$file")" > "$file"
                fi
            else
                echo "$fix_header$(cat "$file")" > "$file"
            fi
            ((FIXED_COUNT++))
        else
            echo -e "${RED}Missing copyright header: $file${NC}"
            ((MISSING_COUNT++))
        fi
    fi
}

# Check markdown files
while IFS= read -r -d '' file; do
    check_file "$file" "$MD_HEADER" "$MD_HEADER"
done < <(find . -type f -name "*.md" "${EXCLUDE_ARGS[@]}" -print0)

# Check shell scripts
while IFS= read -r -d '' file; do
    check_file "$file" "$SH_HEADER" "$SH_HEADER"
done < <(find . -type f -name "*.sh" "${EXCLUDE_ARGS[@]}" -print0)

echo ""
if [[ "$FIX_MODE" == true ]]; then
    echo -e "${GREEN}Fixed $FIXED_COUNT file(s)${NC}"
    exit 0
else
    if [[ $MISSING_COUNT -eq 0 ]]; then
        echo -e "${GREEN}All files have copyright headers!${NC}"
        exit 0
    else
        echo -e "${RED}Found $MISSING_COUNT file(s) missing copyright headers${NC}"
        echo -e "${YELLOW}Run with --fix to automatically add headers${NC}"
        exit 1
    fi
fi
