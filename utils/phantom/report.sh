#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Phantom Report
# Generate CLI summary reports for scanned projects
#
# Usage: ./report.sh <org/repo>           # Single project report
#        ./report.sh --org <name>         # Org aggregate report
#        ./report.sh --json               # JSON output
#############################################################################

set -e

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Load Gibson library
source "$SCRIPT_DIR/lib/gibson.sh"

#############################################################################
# Configuration
#############################################################################

TARGET=""
ORG=""
JSON_OUTPUT=false

#############################################################################
# Usage
#############################################################################

usage() {
    cat << EOF
Phantom Report - Generate CLI summary reports

Usage: $0 [options] [<org/repo>]

TARGETS:
    <org/repo>          Report for a specific project
    --org <name>        Aggregate report for all projects in an org

OPTIONS:
    --json              Output in JSON format
    -h, --help          Show this help

EXAMPLES:
    $0 expressjs/express            # Single project report
    $0 --org expressjs              # Org aggregate report
    $0 expressjs/express --json     # JSON output

EOF
    exit 0
}

#############################################################################
# Argument Parsing
#############################################################################

parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -h|--help)
                usage
                ;;
            --org)
                ORG="$2"
                shift 2
                ;;
            --json)
                JSON_OUTPUT=true
                shift
                ;;
            -*)
                echo -e "${RED}Error: Unknown option $1${NC}" >&2
                exit 1
                ;;
            *)
                if [[ -z "$TARGET" ]]; then
                    TARGET="$1"
                else
                    echo -e "${RED}Error: Multiple targets specified${NC}" >&2
                    exit 1
                fi
                shift
                ;;
        esac
    done

    if [[ -z "$TARGET" ]] && [[ -z "$ORG" ]]; then
        echo -e "${RED}Error: No target specified${NC}" >&2
        echo "Usage: $0 <org/repo> or $0 --org <name>"
        exit 1
    fi
}

#############################################################################
# Report Functions
#############################################################################

# Format relative time
relative_time() {
    local timestamp="$1"
    if [[ -z "$timestamp" ]] || [[ "$timestamp" == "null" ]]; then
        echo "unknown"
        return
    fi

    local ts_epoch=$(date -j -f "%Y-%m-%dT%H:%M:%SZ" "$timestamp" +%s 2>/dev/null || date -d "$timestamp" +%s 2>/dev/null)
    local now_epoch=$(date +%s)
    local diff=$((now_epoch - ts_epoch))

    if [[ $diff -lt 60 ]]; then
        echo "just now"
    elif [[ $diff -lt 3600 ]]; then
        local mins=$((diff / 60))
        echo "$mins minute$([ $mins -ne 1 ] && echo 's') ago"
    elif [[ $diff -lt 86400 ]]; then
        local hours=$((diff / 3600))
        echo "$hours hour$([ $hours -ne 1 ] && echo 's') ago"
    elif [[ $diff -lt 604800 ]]; then
        local days=$((diff / 86400))
        echo "$days day$([ $days -ne 1 ] && echo 's') ago"
    else
        local weeks=$((diff / 604800))
        echo "$weeks week$([ $weeks -ne 1 ] && echo 's') ago"
    fi
}

# Generate single project report
project_report() {
    local project_id="$1"
    local analysis_path=$(gibson_project_analysis_path "$project_id")
    local manifest="$analysis_path/manifest.json"

    if [[ ! -f "$manifest" ]]; then
        echo -e "${RED}Error: No analysis found for '$project_id'${NC}" >&2
        exit 1
    fi

    if [[ "$JSON_OUTPUT" == "true" ]]; then
        project_report_json "$project_id"
        return
    fi

    # Read manifest data
    local scan_id=$(jq -r '.scan_id // "legacy"' "$manifest" 2>/dev/null)
    local schema_version=$(jq -r '.schema_version // "1.0.0"' "$manifest" 2>/dev/null)
    local profile=$(jq -r '.scan.profile // .mode // "standard"' "$manifest" 2>/dev/null)
    local completed_at=$(jq -r '.scan.completed_at // .completed_at // "unknown"' "$manifest" 2>/dev/null)
    local duration=$(jq -r '.scan.duration_seconds // "N/A"' "$manifest" 2>/dev/null)

    # Git info
    local commit_short=$(jq -r '.git.commit_short // "unknown"' "$manifest" 2>/dev/null)
    local branch=$(jq -r '.git.branch // "unknown"' "$manifest" 2>/dev/null)

    # Summary
    local risk_level=$(jq -r '.summary.risk_level // "unknown"' "$manifest" 2>/dev/null)

    # Print header
    echo
    echo -e "${BOLD}Report: $project_id${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    printf "  %-14s %s\n" "Project:" "$project_id"
    printf "  %-14s %s (%s)\n" "Analyzed:" "$(echo "$completed_at" | cut -d'T' -f1,2 | tr 'T' ' ' | cut -d':' -f1,2)" "$(relative_time "$completed_at")"
    printf "  %-14s %s\n" "Scan ID:" "$scan_id"
    printf "  %-14s %s\n" "Profile:" "$profile"
    if [[ "$commit_short" != "unknown" ]] && [[ "$commit_short" != "null" ]]; then
        printf "  %-14s %s (%s)\n" "Commit:" "$commit_short" "$branch"
    fi
    if [[ "$duration" != "N/A" ]] && [[ "$duration" != "null" ]]; then
        printf "  %-14s %ss\n" "Duration:" "$duration"
    fi

    echo
    echo -e "${BOLD}SECURITY SUMMARY${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    # Risk level with color
    local risk_upper=$(echo "$risk_level" | tr '[:lower:]' '[:upper:]')
    local risk_color="$GREEN"
    case "$risk_level" in
        critical) risk_color="$RED" ;;
        high) risk_color="$RED" ;;
        medium) risk_color="$YELLOW" ;;
    esac
    printf "  %-14s ${risk_color}%s${NC}\n" "Risk Level:" "$risk_upper"
    echo

    # Vulnerabilities
    if [[ -f "$analysis_path/package-vulns.json" ]]; then
        local c=$(jq -r '.summary.critical // 0' "$analysis_path/package-vulns.json" 2>/dev/null)
        local h=$(jq -r '.summary.high // 0' "$analysis_path/package-vulns.json" 2>/dev/null)
        local m=$(jq -r '.summary.medium // 0' "$analysis_path/package-vulns.json" 2>/dev/null)
        local l=$(jq -r '.summary.low // 0' "$analysis_path/package-vulns.json" 2>/dev/null)
        printf "  Vulnerabilities:\n"
        printf "    %-12s " "Critical:"
        [[ "$c" -gt 0 ]] && echo -e "${RED}$c${NC}" || echo "0"
        printf "    %-12s " "High:"
        [[ "$h" -gt 0 ]] && echo -e "${YELLOW}$h${NC}" || echo "0"
        printf "    %-12s %s\n" "Medium:" "$m"
        printf "    %-12s %s\n" "Low:" "$l"
    fi

    # Licenses
    if [[ -f "$analysis_path/licenses.json" ]]; then
        local lic_status=$(jq -r '.summary.overall_status // "unknown"' "$analysis_path/licenses.json" 2>/dev/null)
        local violations=$(jq -r '.summary.license_violations // 0' "$analysis_path/licenses.json" 2>/dev/null)
        printf "\n  %-14s " "Licenses:"
        if [[ "$violations" -gt 0 ]]; then
            echo -e "${RED}$violations violations${NC}"
        elif [[ "$lic_status" == "pass" ]]; then
            echo -e "${GREEN}Compliant${NC}"
        else
            echo "$lic_status"
        fi
    fi

    # Secrets
    if [[ -f "$analysis_path/code-secrets.json" ]]; then
        local secrets=$(jq -r '.summary.total_findings // 0' "$analysis_path/code-secrets.json" 2>/dev/null)
        printf "  %-14s " "Secrets:"
        if [[ "$secrets" -gt 0 ]]; then
            echo -e "${RED}$secrets exposed${NC}"
        else
            echo -e "${GREEN}0 exposed${NC}"
        fi
    fi

    echo
    echo -e "${BOLD}DEPENDENCIES${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    if [[ -f "$analysis_path/package-sbom.json" ]]; then
        local total=$(jq -r '.total_dependencies // .summary.total // 0' "$analysis_path/package-sbom.json" 2>/dev/null)
        local direct=$(jq -r '.direct_dependencies // .summary.direct // 0' "$analysis_path/package-sbom.json" 2>/dev/null)
        printf "  %-14s %s packages\n" "Total:" "$total"
        printf "  %-14s %s\n" "Direct:" "$direct"
    fi

    if [[ -f "$analysis_path/package-health.json" ]]; then
        local abandoned=$(jq -r '.summary.abandoned // 0' "$analysis_path/package-health.json" 2>/dev/null)
        if [[ "$abandoned" -gt 0 ]]; then
            printf "  %-14s " "Abandoned:"
            echo -e "${YELLOW}$abandoned${NC}"
        fi
    fi

    # DORA metrics if available
    if [[ -f "$analysis_path/dora.json" ]]; then
        local perf=$(jq -r '.summary.overall_performance // "N/A"' "$analysis_path/dora.json" 2>/dev/null)
        if [[ "$perf" != "N/A" ]] && [[ "$perf" != "null" ]]; then
            echo
            echo -e "${BOLD}DORA METRICS${NC}"
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
            local perf_color="$NC"
            [[ "$perf" == "ELITE" ]] && perf_color="$GREEN"
            [[ "$perf" == "HIGH" ]] && perf_color="$GREEN"
            [[ "$perf" == "LOW" ]] && perf_color="$RED"
            printf "  %-14s ${perf_color}%s${NC}\n" "Performance:" "$perf"
        fi
    fi

    # Ownership if available
    if [[ -f "$analysis_path/code-ownership.json" ]]; then
        local contributors=$(jq -r '.summary.active_contributors // 0' "$analysis_path/code-ownership.json" 2>/dev/null)
        echo
        echo -e "${BOLD}OWNERSHIP${NC}"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        printf "  %-14s %s\n" "Contributors:" "$contributors"
    fi

    echo
}

# Generate JSON project report
project_report_json() {
    local project_id="$1"
    local analysis_path=$(gibson_project_analysis_path "$project_id")
    local manifest="$analysis_path/manifest.json"

    # Build comprehensive JSON report
    jq -n \
        --arg project_id "$project_id" \
        --slurpfile manifest "$manifest" \
        --slurpfile vulns "$analysis_path/package-vulns.json" \
        --slurpfile sbom "$analysis_path/package-sbom.json" \
        --slurpfile licenses "$analysis_path/licenses.json" \
        '{
            project_id: $project_id,
            scan_id: $manifest[0].scan_id,
            git: $manifest[0].git,
            scan: $manifest[0].scan,
            summary: $manifest[0].summary,
            vulnerabilities: ($vulns[0].summary // {}),
            dependencies: ($sbom[0].summary // {}),
            licenses: ($licenses[0].summary // {})
        }' 2>/dev/null || cat "$manifest"
}

# Generate org aggregate report
org_report() {
    local org="$1"
    local projects=$(gibson_list_org_projects "$org")

    if [[ -z "$projects" ]]; then
        echo -e "${RED}Error: No projects found for org '$org'${NC}" >&2
        exit 1
    fi

    if [[ "$JSON_OUTPUT" == "true" ]]; then
        org_report_json "$org"
        return
    fi

    # Get org index if available
    local org_index="$GIBSON_PROJECTS_DIR/$org/_index.json"

    echo
    echo -e "${BOLD}Org Report: $org${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    local project_count=$(echo "$projects" | wc -w | tr -d ' ')
    printf "  %-14s %s\n" "Projects:" "$project_count"

    if [[ -f "$org_index" ]]; then
        local updated=$(jq -r '.updated_at // "unknown"' "$org_index" 2>/dev/null)
        printf "  %-14s %s\n" "Updated:" "$(relative_time "$updated")"

        echo
        echo -e "${BOLD}AGGREGATE SECURITY${NC}"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

        local total_vulns=$(jq -r '.aggregate.total_vulnerabilities // 0' "$org_index" 2>/dev/null)
        local critical=$(jq -r '.aggregate.critical // 0' "$org_index" 2>/dev/null)
        local high=$(jq -r '.aggregate.high // 0' "$org_index" 2>/dev/null)
        local total_deps=$(jq -r '.aggregate.total_dependencies // 0' "$org_index" 2>/dev/null)

        printf "  %-14s %s\n" "Total Vulns:" "$total_vulns"
        printf "    %-12s " "Critical:"
        [[ "$critical" -gt 0 ]] && echo -e "${RED}$critical${NC}" || echo "0"
        printf "    %-12s " "High:"
        [[ "$high" -gt 0 ]] && echo -e "${YELLOW}$high${NC}" || echo "0"
        printf "  %-14s %s\n" "Dependencies:" "$total_deps"

        local repos_at_risk=$(jq -r '.aggregate.repos_at_risk // [] | join(", ")' "$org_index" 2>/dev/null)
        if [[ -n "$repos_at_risk" ]] && [[ "$repos_at_risk" != "" ]]; then
            printf "\n  ${YELLOW}At Risk:${NC} %s\n" "$repos_at_risk"
        fi
    fi

    echo
    echo -e "${BOLD}PROJECTS${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    for repo in $projects; do
        local project_id="$org/$repo"
        local manifest="$GIBSON_PROJECTS_DIR/$project_id/analysis/manifest.json"
        if [[ -f "$manifest" ]]; then
            local risk=$(jq -r '.summary.risk_level // "unknown"' "$manifest" 2>/dev/null)
            local risk_color="$GREEN"
            case "$risk" in
                critical) risk_color="$RED" ;;
                high) risk_color="$RED" ;;
                medium) risk_color="$YELLOW" ;;
            esac
            local vulns=$(jq -r '.summary.total_vulnerabilities // 0' "$manifest" 2>/dev/null)
            printf "  %-30s ${risk_color}%-8s${NC} %s vulns\n" "$repo" "$risk" "$vulns"
        else
            printf "  %-30s ${DIM}no analysis${NC}\n" "$repo"
        fi
    done

    echo
}

# Generate JSON org report
org_report_json() {
    local org="$1"
    local org_index="$GIBSON_PROJECTS_DIR/$org/_index.json"

    if [[ -f "$org_index" ]]; then
        cat "$org_index"
    else
        # Build report from project manifests
        gibson_get_org_index "$org" 2>/dev/null || echo '{"error": "No org index available"}'
    fi
}

#############################################################################
# Main
#############################################################################

main() {
    parse_args "$@"

    # Ensure Gibson is initialized
    gibson_ensure_initialized

    if [[ -n "$ORG" ]]; then
        org_report "$ORG"
    else
        local project_id=$(gibson_project_id "$TARGET")
        project_report "$project_id"
    fi
}

main "$@"
