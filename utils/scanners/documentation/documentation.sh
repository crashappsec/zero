#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Documentation - Data Collector
# Analyzes documentation coverage: README, API docs, comments, ADRs
# Outputs JSON for agent analysis - NO AI calls
#
# Usage: ./documentation-data.sh [options] <target>
# Output: JSON with documentation inventory and coverage metrics
#############################################################################

set -e

# Colors for terminal output (stderr only)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCANNERS_ROOT="$(dirname "$SCRIPT_DIR")"
UTILS_ROOT="$(dirname "$SCANNERS_ROOT")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"

# Default options
OUTPUT_FILE=""
LOCAL_PATH=""
REPO=""
ORG=""
TEMP_DIR=""
CLEANUP=true
TARGET=""

usage() {
    cat << EOF
Documentation - Data Collector (JSON output for agent analysis)

Usage: $0 [OPTIONS] <target>

TARGET:
    Git repository URL      Clone and analyze repository
    Local directory path    Analyze local repository

OPTIONS:
    --local-path PATH       Use pre-cloned repository (skips cloning)
    --repo OWNER/REPO       GitHub repository (looks in zero cache)
    --org ORG               GitHub org (uses first repo found in zero cache)
    -o, --output FILE       Write JSON to file (default: stdout)
    -k, --keep-clone        Keep cloned repository
    -h, --help              Show this help

OUTPUT:
    JSON object with:
    - summary: documentation score, file counts
    - standard_files: README, LICENSE, CONTRIBUTING, etc.
    - api_docs: OpenAPI specs, JSDoc coverage
    - documentation_files: docs/, ADRs, runbooks
    - missing: recommended files not found

EXAMPLES:
    $0 https://github.com/expressjs/express
    $0 --local-path ~/.zero/projects/foo/repo
    $0 -o documentation.json /path/to/project

EOF
    exit 0
}

# Clone repository
clone_repository() {
    local repo_url="$1"
    TEMP_DIR=$(mktemp -d)

    echo -e "${BLUE}Cloning repository...${NC}" >&2
    if git clone --depth 1 "$repo_url" "$TEMP_DIR" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Cloned${NC}" >&2
        return 0
    else
        echo '{"error": "Failed to clone repository"}'
        exit 1
    fi
}

# Cleanup
cleanup() {
    if [[ "$CLEANUP" == true ]] && [[ -n "$TEMP_DIR" ]] && [[ -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
}
trap cleanup EXIT

# Detect if target is a Git URL
is_git_url() {
    [[ "$1" =~ ^(https?|git)://.*\.git$ ]] || [[ "$1" =~ ^git@.*:.*\.git$ ]] || [[ "$1" =~ github\.com|gitlab\.com|bitbucket\.org ]]
}

# Get file info
get_file_info() {
    local file="$1"
    local repo_dir="$2"

    if [[ ! -f "$file" ]]; then
        echo "null"
        return
    fi

    local rel_path="${file#$repo_dir/}"
    local size=$(wc -c < "$file" 2>/dev/null | tr -d ' ')
    local lines=$(wc -l < "$file" 2>/dev/null | tr -d ' ')

    # Get last modified date from git
    local last_modified=""
    if [[ -d "$repo_dir/.git" ]]; then
        last_modified=$(cd "$repo_dir" && git log -1 --format="%aI" -- "$rel_path" 2>/dev/null || echo "")
    fi

    jq -n \
        --arg path "$rel_path" \
        --argjson size "$size" \
        --argjson lines "$lines" \
        --arg modified "$last_modified" \
        '{
            "path": $path,
            "size_bytes": $size,
            "lines": $lines,
            "last_modified": (if $modified != "" then $modified else null end)
        }'
}

# Analyze README sections
analyze_readme() {
    local file="$1"
    local repo_dir="$2"

    if [[ ! -f "$file" ]]; then
        echo '{"exists": false}'
        return
    fi

    local content=$(cat "$file")
    local sections="[]"

    # Common README sections to look for
    local section_patterns=(
        "Installation|Install|Setup|Getting Started"
        "Usage|How to Use|Quick Start"
        "API|Reference|Documentation"
        "Configuration|Config|Options"
        "Contributing|Contribution|Contributors"
        "License|Licence"
        "Testing|Tests|Test"
        "Examples|Demo"
        "Requirements|Prerequisites|Dependencies"
        "Changelog|History|Release Notes"
        "FAQ|Questions"
        "Troubleshooting|Common Issues"
        "Security|Vulnerability"
        "Support|Help|Contact"
    )

    for pattern in "${section_patterns[@]}"; do
        local section_name=$(echo "$pattern" | cut -d'|' -f1)
        if echo "$content" | grep -qiE "^#+[[:space:]]*($pattern)"; then
            sections=$(echo "$sections" | jq --arg name "$section_name" '. + [$name]')
        fi
    done

    # Check for badges
    local has_badges=false
    if echo "$content" | grep -qE "\!\[.*\]\(.*badge.*\)|\!\[.*\]\(.*shields\.io"; then
        has_badges=true
    fi

    # Check for code examples
    local has_code_examples=false
    if echo "$content" | grep -qE '```'; then
        has_code_examples=true
    fi

    local file_info=$(get_file_info "$file" "$repo_dir")

    echo "$file_info" | jq \
        --argjson sections "$sections" \
        --argjson has_badges "$has_badges" \
        --argjson has_code_examples "$has_code_examples" \
        '. + {
            "exists": true,
            "sections": $sections,
            "has_badges": $has_badges,
            "has_code_examples": $has_code_examples
        }'
}

# Find standard documentation files
find_standard_files() {
    local repo_dir="$1"
    local files="{}"

    # README variants
    local readme=$(find "$repo_dir" -maxdepth 1 -iname "readme*" -type f 2>/dev/null | head -1)
    if [[ -n "$readme" ]]; then
        files=$(echo "$files" | jq --argjson info "$(analyze_readme "$readme" "$repo_dir")" '.readme = $info')
    else
        files=$(echo "$files" | jq '.readme = {"exists": false}')
    fi

    # LICENSE
    local license=$(find "$repo_dir" -maxdepth 1 -iname "license*" -o -iname "licence*" -type f 2>/dev/null | head -1)
    if [[ -n "$license" ]]; then
        local license_type=""
        local content=$(head -20 "$license")
        if echo "$content" | grep -qi "MIT"; then license_type="MIT"
        elif echo "$content" | grep -qi "Apache"; then license_type="Apache-2.0"
        elif echo "$content" | grep -qi "GPL"; then license_type="GPL"
        elif echo "$content" | grep -qi "BSD"; then license_type="BSD"
        elif echo "$content" | grep -qi "ISC"; then license_type="ISC"
        fi

        local info=$(get_file_info "$license" "$repo_dir")
        files=$(echo "$files" | jq --argjson info "$info" --arg type "$license_type" \
            '.license = ($info + {"exists": true, "type": (if $type != "" then $type else null end)})')
    else
        files=$(echo "$files" | jq '.license = {"exists": false}')
    fi

    # CONTRIBUTING
    local contributing=$(find "$repo_dir" -maxdepth 1 -iname "contributing*" -type f 2>/dev/null | head -1)
    if [[ -n "$contributing" ]]; then
        local info=$(get_file_info "$contributing" "$repo_dir")
        files=$(echo "$files" | jq --argjson info "$info" '.contributing = ($info + {"exists": true})')
    else
        files=$(echo "$files" | jq '.contributing = {"exists": false}')
    fi

    # CHANGELOG
    local changelog=$(find "$repo_dir" -maxdepth 1 -iname "changelog*" -o -iname "history*" -o -iname "changes*" -type f 2>/dev/null | head -1)
    if [[ -n "$changelog" ]]; then
        local info=$(get_file_info "$changelog" "$repo_dir")
        files=$(echo "$files" | jq --argjson info "$info" '.changelog = ($info + {"exists": true})')
    else
        files=$(echo "$files" | jq '.changelog = {"exists": false}')
    fi

    # CODE_OF_CONDUCT
    local coc=$(find "$repo_dir" -maxdepth 1 -iname "code_of_conduct*" -o -iname "code-of-conduct*" -type f 2>/dev/null | head -1)
    if [[ -n "$coc" ]]; then
        local info=$(get_file_info "$coc" "$repo_dir")
        files=$(echo "$files" | jq --argjson info "$info" '.code_of_conduct = ($info + {"exists": true})')
    else
        files=$(echo "$files" | jq '.code_of_conduct = {"exists": false}')
    fi

    # SECURITY
    local security=$(find "$repo_dir" -maxdepth 1 -iname "security*" -type f 2>/dev/null | head -1)
    if [[ -n "$security" ]]; then
        local info=$(get_file_info "$security" "$repo_dir")
        files=$(echo "$files" | jq --argjson info "$info" '.security = ($info + {"exists": true})')
    else
        files=$(echo "$files" | jq '.security = {"exists": false}')
    fi

    # CODEOWNERS
    local codeowners=$(find "$repo_dir" -name "CODEOWNERS" -type f 2>/dev/null | head -1)
    if [[ -n "$codeowners" ]]; then
        local info=$(get_file_info "$codeowners" "$repo_dir")
        files=$(echo "$files" | jq --argjson info "$info" '.codeowners = ($info + {"exists": true})')
    else
        files=$(echo "$files" | jq '.codeowners = {"exists": false}')
    fi

    echo "$files"
}

# Find API documentation
find_api_docs() {
    local repo_dir="$1"
    local api_docs="{}"

    # OpenAPI/Swagger specs
    local openapi_files=$(find "$repo_dir" -type f \( \
        -name "openapi.yaml" -o -name "openapi.yml" -o -name "openapi.json" \
        -o -name "swagger.yaml" -o -name "swagger.yml" -o -name "swagger.json" \
        -o -name "api.yaml" -o -name "api.yml" -o -name "api.json" \
    \) ! -path "*node_modules*" ! -path "*vendor*" 2>/dev/null)

    local openapi_list="[]"
    while IFS= read -r file; do
        [[ -z "$file" ]] && continue
        local rel_path="${file#$repo_dir/}"
        openapi_list=$(echo "$openapi_list" | jq --arg path "$rel_path" '. + [$path]')
    done <<< "$openapi_files"
    api_docs=$(echo "$api_docs" | jq --argjson list "$openapi_list" '.openapi_specs = $list')

    # GraphQL schemas
    local graphql_files=$(find "$repo_dir" -type f \( \
        -name "*.graphql" -o -name "*.gql" -o -name "schema.graphql" \
    \) ! -path "*node_modules*" ! -path "*vendor*" 2>/dev/null | head -20)

    local graphql_list="[]"
    while IFS= read -r file; do
        [[ -z "$file" ]] && continue
        local rel_path="${file#$repo_dir/}"
        graphql_list=$(echo "$graphql_list" | jq --arg path "$rel_path" '. + [$path]')
    done <<< "$graphql_files"
    api_docs=$(echo "$api_docs" | jq --argjson list "$graphql_list" '.graphql_schemas = $list')

    # Check for JSDoc/TSDoc
    local jsdoc_config=$(find "$repo_dir" -maxdepth 2 -name "jsdoc.json" -o -name ".jsdoc.json" -o -name "jsdoc.conf.json" 2>/dev/null | head -1)
    if [[ -n "$jsdoc_config" ]]; then
        api_docs=$(echo "$api_docs" | jq '.jsdoc_config = true')
    else
        api_docs=$(echo "$api_docs" | jq '.jsdoc_config = false')
    fi

    # Check for TypeDoc
    local typedoc_config=$(find "$repo_dir" -maxdepth 2 -name "typedoc.json" -o -name "typedoc.js" 2>/dev/null | head -1)
    if [[ -n "$typedoc_config" ]]; then
        api_docs=$(echo "$api_docs" | jq '.typedoc_config = true')
    else
        api_docs=$(echo "$api_docs" | jq '.typedoc_config = false')
    fi

    # Check for Sphinx (Python)
    local sphinx_config=$(find "$repo_dir" -name "conf.py" -path "*/docs/*" 2>/dev/null | head -1)
    if [[ -n "$sphinx_config" ]]; then
        api_docs=$(echo "$api_docs" | jq '.sphinx_config = true')
    else
        api_docs=$(echo "$api_docs" | jq '.sphinx_config = false')
    fi

    echo "$api_docs"
}

# Find documentation directories and files
find_doc_directories() {
    local repo_dir="$1"
    local doc_dirs="[]"
    local doc_files="[]"
    local adrs="[]"

    # Common documentation directories
    local dirs_to_check=("docs" "doc" "documentation" "wiki" "guides" "tutorials")

    for dir_name in "${dirs_to_check[@]}"; do
        local dir_path="$repo_dir/$dir_name"
        if [[ -d "$dir_path" ]]; then
            local file_count=$(find "$dir_path" -type f \( -name "*.md" -o -name "*.rst" -o -name "*.txt" -o -name "*.html" \) 2>/dev/null | wc -l | tr -d ' ')
            doc_dirs=$(echo "$doc_dirs" | jq \
                --arg name "$dir_name" \
                --argjson count "$file_count" \
                '. + [{"name": $name, "file_count": $count}]')
        fi
    done

    # Find ADRs (Architecture Decision Records)
    local adr_dirs=$(find "$repo_dir" -type d -iname "adr*" -o -iname "decisions" 2>/dev/null)
    while IFS= read -r adr_dir; do
        [[ -z "$adr_dir" ]] && continue
        local adr_files=$(find "$adr_dir" -type f -name "*.md" 2>/dev/null)
        while IFS= read -r adr_file; do
            [[ -z "$adr_file" ]] && continue
            local rel_path="${adr_file#$repo_dir/}"
            adrs=$(echo "$adrs" | jq --arg path "$rel_path" '. + [$path]')
        done <<< "$adr_files"
    done <<< "$adr_dirs"

    # Find runbooks
    local runbooks="[]"
    local runbook_files=$(find "$repo_dir" -type f -iname "*runbook*" -o -iname "*playbook*" 2>/dev/null | head -20)
    while IFS= read -r file; do
        [[ -z "$file" ]] && continue
        local rel_path="${file#$repo_dir/}"
        runbooks=$(echo "$runbooks" | jq --arg path "$rel_path" '. + [$path]')
    done <<< "$runbook_files"

    # Count total markdown files
    local total_md=$(find "$repo_dir" -type f -name "*.md" ! -path "*node_modules*" ! -path "*vendor*" 2>/dev/null | wc -l | tr -d ' ')

    jq -n \
        --argjson directories "$doc_dirs" \
        --argjson adrs "$adrs" \
        --argjson runbooks "$runbooks" \
        --argjson total_md "$total_md" \
        '{
            "directories": $directories,
            "adrs": $adrs,
            "runbooks": $runbooks,
            "total_markdown_files": $total_md
        }'
}

# Calculate comment ratio (simplified)
calculate_comment_ratio() {
    local repo_dir="$1"

    # Use cloc if available
    if command -v cloc &> /dev/null; then
        local cloc_output=$(cloc --json "$repo_dir" 2>/dev/null || echo '{}')
        if [[ "$cloc_output" != "{}" ]]; then
            local code_lines=$(echo "$cloc_output" | jq -r '.SUM.code // 0')
            local comment_lines=$(echo "$cloc_output" | jq -r '.SUM.comment // 0')

            if [[ "$code_lines" -gt 0 ]]; then
                local ratio=$(echo "scale=3; $comment_lines / $code_lines" | bc 2>/dev/null || echo "0")
                jq -n \
                    --argjson code "$code_lines" \
                    --argjson comments "$comment_lines" \
                    --arg ratio "$ratio" \
                    '{
                        "code_lines": $code,
                        "comment_lines": $comments,
                        "ratio": ($ratio | tonumber)
                    }'
                return
            fi
        fi
    fi

    echo '{"code_lines": 0, "comment_lines": 0, "ratio": 0, "note": "cloc not installed for accurate stats"}'
}

# Get missing recommended files
get_missing_files() {
    local standard_files="$1"
    local missing="[]"

    # Check each recommended file
    local recommended=("readme" "license" "contributing" "changelog" "security" "code_of_conduct")

    for file_type in "${recommended[@]}"; do
        local exists=$(echo "$standard_files" | jq -r ".$file_type.exists // false")
        if [[ "$exists" == "false" ]]; then
            missing=$(echo "$missing" | jq --arg file "$file_type" '. + [$file]')
        fi
    done

    echo "$missing"
}

# Calculate documentation score
calculate_doc_score() {
    local standard_files="$1"
    local api_docs="$2"
    local doc_dirs="$3"
    local comment_ratio="$4"

    local score=0
    local max_score=100

    # README (20 points)
    local readme_exists=$(echo "$standard_files" | jq -r '.readme.exists // false')
    if [[ "$readme_exists" == "true" ]]; then
        score=$((score + 10))
        local readme_sections=$(echo "$standard_files" | jq -r '.readme.sections | length // 0')
        if [[ "$readme_sections" -ge 3 ]]; then
            score=$((score + 5))
        fi
        local has_code=$(echo "$standard_files" | jq -r '.readme.has_code_examples // false')
        if [[ "$has_code" == "true" ]]; then
            score=$((score + 5))
        fi
    fi

    # LICENSE (10 points)
    local license_exists=$(echo "$standard_files" | jq -r '.license.exists // false')
    if [[ "$license_exists" == "true" ]]; then
        score=$((score + 10))
    fi

    # CONTRIBUTING (10 points)
    local contributing_exists=$(echo "$standard_files" | jq -r '.contributing.exists // false')
    if [[ "$contributing_exists" == "true" ]]; then
        score=$((score + 10))
    fi

    # CHANGELOG (10 points)
    local changelog_exists=$(echo "$standard_files" | jq -r '.changelog.exists // false')
    if [[ "$changelog_exists" == "true" ]]; then
        score=$((score + 10))
    fi

    # SECURITY (10 points)
    local security_exists=$(echo "$standard_files" | jq -r '.security.exists // false')
    if [[ "$security_exists" == "true" ]]; then
        score=$((score + 10))
    fi

    # API Documentation (15 points)
    local openapi_count=$(echo "$api_docs" | jq -r '.openapi_specs | length // 0')
    local has_jsdoc=$(echo "$api_docs" | jq -r '.jsdoc_config // false')
    local has_typedoc=$(echo "$api_docs" | jq -r '.typedoc_config // false')
    local has_sphinx=$(echo "$api_docs" | jq -r '.sphinx_config // false')

    if [[ "$openapi_count" -gt 0 ]]; then
        score=$((score + 10))
    fi
    if [[ "$has_jsdoc" == "true" ]] || [[ "$has_typedoc" == "true" ]] || [[ "$has_sphinx" == "true" ]]; then
        score=$((score + 5))
    fi

    # Documentation directory (10 points)
    local doc_dir_count=$(echo "$doc_dirs" | jq -r '.directories | length // 0')
    if [[ "$doc_dir_count" -gt 0 ]]; then
        score=$((score + 10))
    fi

    # ADRs (5 points)
    local adr_count=$(echo "$doc_dirs" | jq -r '.adrs | length // 0')
    if [[ "$adr_count" -gt 0 ]]; then
        score=$((score + 5))
    fi

    # Comment ratio (10 points)
    local ratio=$(echo "$comment_ratio" | jq -r '.ratio // 0')
    if (( $(echo "$ratio >= 0.1" | bc -l 2>/dev/null || echo 0) )); then
        score=$((score + 10))
    elif (( $(echo "$ratio >= 0.05" | bc -l 2>/dev/null || echo 0) )); then
        score=$((score + 5))
    fi

    echo "$score"
}

# Main analysis
analyze_target() {
    local repo_dir="$1"

    echo -e "${BLUE}Finding standard documentation files...${NC}" >&2
    local standard_files=$(find_standard_files "$repo_dir")
    echo -e "${GREEN}✓ Scanned standard files${NC}" >&2

    echo -e "${BLUE}Finding API documentation...${NC}" >&2
    local api_docs=$(find_api_docs "$repo_dir")
    local openapi_count=$(echo "$api_docs" | jq -r '.openapi_specs | length // 0')
    echo -e "${GREEN}✓ Found $openapi_count OpenAPI specs${NC}" >&2

    echo -e "${BLUE}Scanning documentation directories...${NC}" >&2
    local doc_dirs=$(find_doc_directories "$repo_dir")
    local dir_count=$(echo "$doc_dirs" | jq -r '.directories | length // 0')
    local adr_count=$(echo "$doc_dirs" | jq -r '.adrs | length // 0')
    echo -e "${GREEN}✓ Found $dir_count doc directories, $adr_count ADRs${NC}" >&2

    echo -e "${BLUE}Calculating comment ratio...${NC}" >&2
    local comment_ratio=$(calculate_comment_ratio "$repo_dir")
    local ratio=$(echo "$comment_ratio" | jq -r '.ratio // 0')
    echo -e "${GREEN}✓ Comment ratio: $ratio${NC}" >&2

    # Calculate score
    local doc_score=$(calculate_doc_score "$standard_files" "$api_docs" "$doc_dirs" "$comment_ratio")

    # Determine level
    local doc_level="poor"
    if [[ "$doc_score" -ge 80 ]]; then
        doc_level="excellent"
    elif [[ "$doc_score" -ge 60 ]]; then
        doc_level="good"
    elif [[ "$doc_score" -ge 40 ]]; then
        doc_level="fair"
    fi

    echo -e "${CYAN}Documentation Score: $doc_score/100 ($doc_level)${NC}" >&2

    # Get missing files
    local missing=$(get_missing_files "$standard_files")

    # Build final output
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    jq -n \
        --arg ts "$timestamp" \
        --arg tgt "$TARGET" \
        --arg ver "1.0.0" \
        --argjson score "$doc_score" \
        --arg level "$doc_level" \
        --argjson standard "$standard_files" \
        --argjson api "$api_docs" \
        --argjson dirs "$doc_dirs" \
        --argjson comments "$comment_ratio" \
        --argjson missing "$missing" \
        '{
            analyzer: "documentation",
            version: $ver,
            timestamp: $ts,
            target: $tgt,
            summary: {
                documentation_score: $score,
                documentation_level: $level,
                readme_exists: $standard.readme.exists,
                license_exists: $standard.license.exists,
                api_docs_present: (($api.openapi_specs | length) > 0 or $api.jsdoc_config or $api.typedoc_config),
                comment_ratio: $comments.ratio,
                missing_count: ($missing | length)
            },
            standard_files: $standard,
            api_documentation: $api,
            documentation_directories: $dirs,
            comment_analysis: $comments,
            missing_recommended: $missing
        }'
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help) usage ;;
        --local-path)
            LOCAL_PATH="$2"
            shift 2
            ;;
        --repo)
            REPO="$2"
            shift 2
            ;;
        --org)
            ORG="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        -k|--keep-clone)
            CLEANUP=false
            shift
            ;;
        -*)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
        *)
            TARGET="$1"
            shift
            ;;
    esac
done

# Main execution
scan_path=""

if [[ -n "$LOCAL_PATH" ]]; then
    [[ ! -d "$LOCAL_PATH" ]] && { echo '{"error": "Local path does not exist"}'; exit 1; }
    scan_path="$LOCAL_PATH"
    TARGET="$LOCAL_PATH"
elif [[ -n "$REPO" ]]; then
    # Look in zero cache
    REPO_ORG=$(echo "$REPO" | cut -d'/' -f1)
    REPO_NAME=$(echo "$REPO" | cut -d'/' -f2)
    ZERO_CACHE_PATH="$HOME/.zero/projects/$REPO_ORG/$REPO_NAME/repo"
    GIBSON_PATH="$HOME/.gibson/projects/${REPO_ORG}-${REPO_NAME}/repo"

    if [[ -d "$ZERO_CACHE_PATH" ]]; then
        scan_path="$ZERO_CACHE_PATH"
        TARGET="$REPO"
    elif [[ -d "$GIBSON_PATH" ]]; then
        scan_path="$GIBSON_PATH"
        TARGET="$REPO"
    else
        echo '{"error": "Repository not found in cache. Clone it first or use --local-path"}'
        exit 1
    fi
elif [[ -n "$ORG" ]]; then
    # Scan ALL repos in the org
    ORG_PATH="$HOME/.zero/projects/$ORG"
    if [[ -d "$ORG_PATH" ]]; then
        # Collect repos with and without cloned code
        REPOS_TO_SCAN=()
        REPOS_NOT_CLONED=()
        for repo_dir in "$ORG_PATH"/*/; do
            repo_name=$(basename "$repo_dir")
            if [[ -d "$repo_dir/repo" ]]; then
                REPOS_TO_SCAN+=("$repo_name")
            else
                REPOS_NOT_CLONED+=("$repo_name")
            fi
        done

        # Check if there are uncloned repos and prompt user
        if [[ ${#REPOS_NOT_CLONED[@]} -gt 0 ]]; then
            echo -e "${YELLOW}Found ${#REPOS_NOT_CLONED[@]} repositories without cloned code:${NC}" >&2
            for repo in "${REPOS_NOT_CLONED[@]}"; do
                echo -e "  - $repo" >&2
            done
            echo "" >&2

            read -p "Would you like to hydrate these repos for analysis? [y/N] " -n 1 -r >&2
            echo "" >&2

            if [[ $REPLY =~ ^[Yy]$ ]]; then
                echo -e "${BLUE}Hydrating ${#REPOS_NOT_CLONED[@]} repositories...${NC}" >&2
                for repo in "${REPOS_NOT_CLONED[@]}"; do
                    echo -e "${CYAN}Cloning $ORG/$repo...${NC}" >&2
                    "$REPO_ROOT/utils/zero/hydrate.sh" --repo "$ORG/$repo" --quick >&2 2>&1 || true
                    if [[ -d "$ORG_PATH/$repo/repo" ]]; then
                        REPOS_TO_SCAN+=("$repo")
                        echo -e "${GREEN}✓ $repo ready${NC}" >&2
                    else
                        echo -e "${RED}✗ Failed to clone $repo${NC}" >&2
                    fi
                done
                echo "" >&2
            else
                echo -e "${CYAN}Continuing with ${#REPOS_TO_SCAN[@]} already-cloned repositories...${NC}" >&2
            fi
        fi

        if [[ ${#REPOS_TO_SCAN[@]} -eq 0 ]]; then
            echo '{"error": "No repositories with cloned code found in org cache. Hydrate repos first."}'
            exit 1
        fi

        # Analyze each repo and aggregate results
        echo -e "${BLUE}Scanning ${#REPOS_TO_SCAN[@]} repositories in $ORG...${NC}" >&2

        all_results="[]"
        repo_count=0
        total_repos=${#REPOS_TO_SCAN[@]}

        for repo_name in "${REPOS_TO_SCAN[@]}"; do
            ((repo_count++))
            scan_path="$ORG_PATH/$repo_name/repo"
            TARGET="$ORG/$repo_name"

            echo -e "\n${CYAN}[$repo_count/$total_repos] Analyzing: $TARGET${NC}" >&2

            repo_json=$(analyze_target "$scan_path")
            repo_json=$(echo "$repo_json" | jq --arg repo "$TARGET" '. + {repository: $repo}')

            all_results=$(echo "$all_results" | jq --argjson repo "$repo_json" '. + [$repo]')
        done

        # Build aggregated output
        timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

        final_json=$(jq -n \
            --arg ts "$timestamp" \
            --arg org "$ORG" \
            --arg ver "1.0.0" \
            --argjson repo_count "$total_repos" \
            --argjson repositories "$all_results" \
            '{
                analyzer: "documentation",
                version: $ver,
                timestamp: $ts,
                organization: $org,
                summary: {
                    repositories_scanned: $repo_count
                },
                repositories: $repositories
            }')

        echo -e "\n${CYAN}=== Organization Summary ===${NC}" >&2
        echo -e "${CYAN}Repos analyzed: $total_repos${NC}" >&2

        # Output
        if [[ -n "$OUTPUT_FILE" ]]; then
            echo "$final_json" > "$OUTPUT_FILE"
            echo -e "${GREEN}✓ Results written to $OUTPUT_FILE${NC}" >&2
        else
            echo "$final_json"
        fi
        exit 0
    else
        echo '{"error": "Org not found in cache. Hydrate repos first."}'
        exit 1
    fi
elif [[ -n "$TARGET" ]]; then
    if is_git_url "$TARGET"; then
        clone_repository "$TARGET"
        scan_path="$TEMP_DIR"
    elif [[ -d "$TARGET" ]]; then
        scan_path="$TARGET"
    else
        echo '{"error": "Invalid target - must be URL or directory"}'
        exit 1
    fi
else
    echo '{"error": "No target specified"}'
    exit 1
fi

echo -e "${BLUE}Analyzing documentation: $TARGET${NC}" >&2

final_json=$(analyze_target "$scan_path")

# Output
if [[ -n "$OUTPUT_FILE" ]]; then
    echo "$final_json" > "$OUTPUT_FILE"
    echo -e "${GREEN}✓ Results written to $OUTPUT_FILE${NC}" >&2
else
    echo "$final_json"
fi
