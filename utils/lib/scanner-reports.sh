#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Scanner Reports Library
# Unified report generation for all scanners
#
# Usage:
#   source "$UTILS_ROOT/lib/scanner-reports.sh"
#   report_init "my-scanner" "1.0.0" "expressjs/express"
#   report_add_section "Summary" "$summary_json"
#   report_add_finding "critical" "SQL Injection" "src/db.js:47" "..."
#   report_generate "markdown"  # or "json", "terminal", "html"
#
# Supported Formats:
# - terminal: Colored terminal output (default)
# - json: Structured JSON output
# - markdown: GitHub-flavored markdown
# - html: Basic HTML report
# - csv: CSV export (findings only)
#############################################################################

# Source scanner-ux for colors and helpers
SCRIPT_DIR_REPORTS="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "$SCRIPT_DIR_REPORTS/scanner-ux.sh" 2>/dev/null || true

#############################################################################
# REPORT STATE
#############################################################################

_REPORT_SCANNER=""
_REPORT_VERSION=""
_REPORT_TARGET=""
_REPORT_TIMESTAMP=""
_REPORT_SECTIONS=()
_REPORT_FINDINGS=()
_REPORT_METADATA='{}'

#############################################################################
# INITIALIZATION
#############################################################################

# Initialize report
# Usage: report_init "scanner-name" "1.0.0" "target-name"
report_init() {
    _REPORT_SCANNER="${1:-scanner}"
    _REPORT_VERSION="${2:-1.0.0}"
    _REPORT_TARGET="${3:-unknown}"
    _REPORT_TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    _REPORT_SECTIONS=()
    _REPORT_FINDINGS=()
    _REPORT_METADATA='{}'
}

# Add metadata to report
# Usage: report_set_metadata "key" "value"
report_set_metadata() {
    local key="$1"
    local value="$2"

    _REPORT_METADATA=$(echo "$_REPORT_METADATA" | jq --arg k "$key" --arg v "$value" '. + {($k): $v}')
}

# Add JSON metadata
# Usage: report_set_metadata_json "key" '{"nested": "value"}'
report_set_metadata_json() {
    local key="$1"
    local json_value="$2"

    _REPORT_METADATA=$(echo "$_REPORT_METADATA" | jq --arg k "$key" --argjson v "$json_value" '. + {($k): $v}')
}

#############################################################################
# SECTIONS
#############################################################################

# Add a section to the report
# Usage: report_add_section "Section Title" "$json_data" ["description"]
report_add_section() {
    local title="$1"
    local data="$2"
    local description="${3:-}"

    local section=$(jq -n \
        --arg title "$title" \
        --arg desc "$description" \
        --argjson data "$data" \
        '{title: $title, description: $desc, data: $data}')

    _REPORT_SECTIONS+=("$section")
}

#############################################################################
# FINDINGS
#############################################################################

# Add a finding to the report
# Usage: report_add_finding "severity" "title" "location" "description" ["remediation"]
report_add_finding() {
    local severity="$1"
    local title="$2"
    local location="${3:-}"
    local description="${4:-}"
    local remediation="${5:-}"

    local finding=$(jq -n \
        --arg sev "$severity" \
        --arg title "$title" \
        --arg loc "$location" \
        --arg desc "$description" \
        --arg fix "$remediation" \
        '{
            severity: $sev,
            title: $title,
            location: $loc,
            description: $desc,
            remediation: $fix
        }')

    _REPORT_FINDINGS+=("$finding")
}

#############################################################################
# JSON OUTPUT
#############################################################################

_report_to_json() {
    # Build sections array
    local sections_json="[]"
    for section in "${_REPORT_SECTIONS[@]}"; do
        sections_json=$(echo "$sections_json" | jq --argjson s "$section" '. + [$s]')
    done

    # Build findings array
    local findings_json="[]"
    for finding in "${_REPORT_FINDINGS[@]}"; do
        findings_json=$(echo "$findings_json" | jq --argjson f "$finding" '. + [$f]')
    done

    # Count findings by severity
    local critical=$(echo "$findings_json" | jq '[.[] | select(.severity == "critical")] | length')
    local high=$(echo "$findings_json" | jq '[.[] | select(.severity == "high")] | length')
    local medium=$(echo "$findings_json" | jq '[.[] | select(.severity == "medium")] | length')
    local low=$(echo "$findings_json" | jq '[.[] | select(.severity == "low")] | length')

    jq -n \
        --arg scanner "$_REPORT_SCANNER" \
        --arg version "$_REPORT_VERSION" \
        --arg target "$_REPORT_TARGET" \
        --arg timestamp "$_REPORT_TIMESTAMP" \
        --argjson metadata "$_REPORT_METADATA" \
        --argjson sections "$sections_json" \
        --argjson findings "$findings_json" \
        --argjson critical "$critical" \
        --argjson high "$high" \
        --argjson medium "$medium" \
        --argjson low "$low" \
        '{
            report: {
                scanner: $scanner,
                version: $version,
                target: $target,
                timestamp: $timestamp,
                metadata: $metadata
            },
            summary: {
                total_findings: ($findings | length),
                by_severity: {
                    critical: $critical,
                    high: $high,
                    medium: $medium,
                    low: $low
                }
            },
            sections: $sections,
            findings: $findings
        }'
}

#############################################################################
# MARKDOWN OUTPUT
#############################################################################

_report_to_markdown() {
    local json=$(_report_to_json)

    # Get movie quote if agent personality is loaded
    local quote=""
    if [[ "${_AGENT_PERSONALITY_LOADED:-false}" == "true" ]]; then
        local context="general"
        local critical=$(echo "$json" | jq -r '.summary.by_severity.critical')
        [[ "$critical" -gt 0 ]] && context="security"
        quote=$(get_movie_quote "$context" 2>/dev/null || true)
    fi

    cat << EOF
# ${_REPORT_SCANNER} Report
EOF

    # Add movie quote if available
    if [[ -n "$quote" ]]; then
        cat << EOF

> *"${quote}"*
EOF
    fi

    cat << EOF

**Target:** \`${_REPORT_TARGET}\`
**Generated:** ${_REPORT_TIMESTAMP}
**Scanner Version:** ${_REPORT_VERSION}

---

## Summary

| Severity | Count |
|----------|-------|
| Critical | $(echo "$json" | jq -r '.summary.by_severity.critical') |
| High | $(echo "$json" | jq -r '.summary.by_severity.high') |
| Medium | $(echo "$json" | jq -r '.summary.by_severity.medium') |
| Low | $(echo "$json" | jq -r '.summary.by_severity.low') |
| **Total** | **$(echo "$json" | jq -r '.summary.total_findings')** |

EOF

    # Output sections
    for section in "${_REPORT_SECTIONS[@]}"; do
        local title=$(echo "$section" | jq -r '.title')
        local desc=$(echo "$section" | jq -r '.description')
        local data=$(echo "$section" | jq -r '.data')

        echo "## $title"
        echo ""
        [[ -n "$desc" ]] && echo "$desc" && echo ""

        # If data is an array, try to make a table
        if echo "$data" | jq -e 'type == "array"' &>/dev/null; then
            local first=$(echo "$data" | jq '.[0]')
            if echo "$first" | jq -e 'type == "object"' &>/dev/null; then
                # Get keys for header
                local keys=$(echo "$first" | jq -r 'keys[]' | head -5)
                local header="|"
                local separator="|"
                for key in $keys; do
                    header+=" $key |"
                    separator+="---|"
                done
                echo "$header"
                echo "$separator"

                # Output rows
                echo "$data" | jq -c '.[]' | while read -r row; do
                    local rowstr="|"
                    for key in $keys; do
                        local val=$(echo "$row" | jq -r --arg k "$key" '.[$k] // ""' | head -c 50)
                        rowstr+=" $val |"
                    done
                    echo "$rowstr"
                done
            else
                # Simple array
                echo "$data" | jq -r '.[]' | sed 's/^/- /'
            fi
        else
            # Object or scalar
            echo '```json'
            echo "$data" | jq '.'
            echo '```'
        fi
        echo ""
    done

    # Output findings
    if [[ ${#_REPORT_FINDINGS[@]} -gt 0 ]]; then
        echo "## Findings"
        echo ""

        for finding in "${_REPORT_FINDINGS[@]}"; do
            local severity=$(echo "$finding" | jq -r '.severity')
            local title=$(echo "$finding" | jq -r '.title')
            local location=$(echo "$finding" | jq -r '.location')
            local desc=$(echo "$finding" | jq -r '.description')
            local fix=$(echo "$finding" | jq -r '.remediation')

            local icon=""
            case "$severity" in
                critical) icon="🔴" ;;
                high)     icon="🟠" ;;
                medium)   icon="🟡" ;;
                low)      icon="🟢" ;;
            esac

            echo "### $icon $title"
            echo ""
            echo "**Severity:** ${severity^}"
            [[ -n "$location" && "$location" != "null" ]] && echo "**Location:** \`$location\`"
            echo ""
            [[ -n "$desc" && "$desc" != "null" ]] && echo "$desc" && echo ""
            [[ -n "$fix" && "$fix" != "null" ]] && echo "**Remediation:** $fix" && echo ""
        done
    fi

    cat << EOF

---

*Generated by ${_REPORT_SCANNER} v${_REPORT_VERSION} • crashoverride.com*
EOF

    # Add signoff if agent personality is loaded
    if [[ "${_AGENT_PERSONALITY_LOADED:-false}" == "true" ]]; then
        echo ""
        echo "*Hack the planet.*"
    fi
}

#############################################################################
# TERMINAL OUTPUT
#############################################################################

_report_to_terminal() {
    local json=$(_report_to_json)

    # Header
    echo -e "${SCANNER_BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${SCANNER_NC}"
    echo -e "${SCANNER_BOLD}${_REPORT_SCANNER} Report${SCANNER_NC}"
    echo -e "${SCANNER_BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${SCANNER_NC}"
    echo ""
    echo -e "${SCANNER_DIM}Target:${SCANNER_NC}    ${_REPORT_TARGET}"
    echo -e "${SCANNER_DIM}Generated:${SCANNER_NC} ${_REPORT_TIMESTAMP}"
    echo ""

    # Summary
    local critical=$(echo "$json" | jq -r '.summary.by_severity.critical')
    local high=$(echo "$json" | jq -r '.summary.by_severity.high')
    local medium=$(echo "$json" | jq -r '.summary.by_severity.medium')
    local low=$(echo "$json" | jq -r '.summary.by_severity.low')
    local total=$(echo "$json" | jq -r '.summary.total_findings')

    echo -e "${SCANNER_BOLD}Summary${SCANNER_NC}"
    [[ $critical -gt 0 ]] && echo -e "  ${SCANNER_RED}●${SCANNER_NC} Critical: $critical"
    [[ $high -gt 0 ]]     && echo -e "  ${SCANNER_RED}○${SCANNER_NC} High: $high"
    [[ $medium -gt 0 ]]   && echo -e "  ${SCANNER_YELLOW}●${SCANNER_NC} Medium: $medium"
    [[ $low -gt 0 ]]      && echo -e "  ${SCANNER_GREEN}○${SCANNER_NC} Low: $low"
    echo -e "  ${SCANNER_DIM}Total: $total${SCANNER_NC}"
    echo ""

    # Sections
    for section in "${_REPORT_SECTIONS[@]}"; do
        local title=$(echo "$section" | jq -r '.title')
        echo -e "${SCANNER_BOLD}${title}${SCANNER_NC}"
        echo -e "${SCANNER_DIM}────────────────────────────────────────${SCANNER_NC}"

        local data=$(echo "$section" | jq -r '.data')

        # Simple display for terminal
        if echo "$data" | jq -e 'type == "array"' &>/dev/null; then
            echo "$data" | jq -r '.[] | if type == "object" then to_entries | map("\(.key): \(.value)") | join(", ") else . end' | head -10 | while read -r line; do
                echo "  • $line"
            done
            local count=$(echo "$data" | jq 'length')
            [[ $count -gt 10 ]] && echo -e "  ${SCANNER_DIM}... and $((count - 10)) more${SCANNER_NC}"
        elif echo "$data" | jq -e 'type == "object"' &>/dev/null; then
            echo "$data" | jq -r 'to_entries[] | "  \(.key): \(.value)"' | head -10
        else
            echo "  $data"
        fi
        echo ""
    done

    # Findings
    if [[ ${#_REPORT_FINDINGS[@]} -gt 0 ]]; then
        echo -e "${SCANNER_BOLD}Findings${SCANNER_NC}"
        echo -e "${SCANNER_DIM}────────────────────────────────────────${SCANNER_NC}"

        for finding in "${_REPORT_FINDINGS[@]}"; do
            local severity=$(echo "$finding" | jq -r '.severity')
            local title=$(echo "$finding" | jq -r '.title')
            local location=$(echo "$finding" | jq -r '.location')

            local icon=""
            case "$severity" in
                critical) icon="${SCANNER_RED}●${SCANNER_NC}" ;;
                high)     icon="${SCANNER_RED}○${SCANNER_NC}" ;;
                medium)   icon="${SCANNER_YELLOW}●${SCANNER_NC}" ;;
                low)      icon="${SCANNER_GREEN}○${SCANNER_NC}" ;;
            esac

            echo -e "  ${icon} ${title}"
            [[ -n "$location" && "$location" != "null" ]] && echo -e "    ${SCANNER_DIM}${location}${SCANNER_NC}"
        done
        echo ""
    fi

    echo -e "${SCANNER_DIM}Generated by ${_REPORT_SCANNER} v${_REPORT_VERSION}${SCANNER_NC}"
}

#############################################################################
# CSV OUTPUT (Findings only)
#############################################################################

_report_to_csv() {
    echo "severity,title,location,description,remediation"

    for finding in "${_REPORT_FINDINGS[@]}"; do
        local severity=$(echo "$finding" | jq -r '.severity' | sed 's/"/""/g')
        local title=$(echo "$finding" | jq -r '.title' | sed 's/"/""/g')
        local location=$(echo "$finding" | jq -r '.location // ""' | sed 's/"/""/g')
        local desc=$(echo "$finding" | jq -r '.description // ""' | sed 's/"/""/g')
        local fix=$(echo "$finding" | jq -r '.remediation // ""' | sed 's/"/""/g')

        echo "\"$severity\",\"$title\",\"$location\",\"$desc\",\"$fix\""
    done
}

#############################################################################
# HTML OUTPUT
#############################################################################

_report_to_html() {
    local json=$(_report_to_json)

    # Get movie quote if agent personality is loaded
    local quote=""
    if [[ "${_AGENT_PERSONALITY_LOADED:-false}" == "true" ]]; then
        local context="general"
        local critical=$(echo "$json" | jq -r '.summary.by_severity.critical')
        [[ "$critical" -gt 0 ]] && context="security"
        quote=$(get_movie_quote "$context" 2>/dev/null || true)
    fi

    cat << EOF
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>${_REPORT_SCANNER} Report - ${_REPORT_TARGET}</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; background: #f5f5f5; }
        .container { background: white; border-radius: 8px; padding: 30px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { color: #333; border-bottom: 2px solid #0066cc; padding-bottom: 10px; }
        h2 { color: #444; margin-top: 30px; }
        .meta { color: #666; font-size: 14px; margin-bottom: 20px; }
        .summary { display: flex; gap: 20px; margin: 20px 0; }
        .summary-card { background: #f8f9fa; border-radius: 6px; padding: 15px 25px; text-align: center; }
        .summary-card.critical { border-left: 4px solid #dc3545; }
        .summary-card.high { border-left: 4px solid #fd7e14; }
        .summary-card.medium { border-left: 4px solid #ffc107; }
        .summary-card.low { border-left: 4px solid #28a745; }
        .summary-card .count { font-size: 24px; font-weight: bold; }
        .summary-card .label { font-size: 12px; color: #666; text-transform: uppercase; }
        .finding { border: 1px solid #ddd; border-radius: 6px; padding: 15px; margin: 10px 0; }
        .finding.critical { border-left: 4px solid #dc3545; }
        .finding.high { border-left: 4px solid #fd7e14; }
        .finding.medium { border-left: 4px solid #ffc107; }
        .finding.low { border-left: 4px solid #28a745; }
        .finding h3 { margin: 0 0 10px 0; font-size: 16px; }
        .finding .location { font-family: monospace; background: #f5f5f5; padding: 2px 6px; border-radius: 3px; font-size: 13px; }
        .badge { display: inline-block; padding: 2px 8px; border-radius: 3px; font-size: 12px; font-weight: bold; color: white; }
        .badge.critical { background: #dc3545; }
        .badge.high { background: #fd7e14; }
        .badge.medium { background: #ffc107; color: #333; }
        .badge.low { background: #28a745; }
        table { width: 100%; border-collapse: collapse; margin: 15px 0; }
        th, td { padding: 10px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background: #f8f9fa; font-weight: 600; }
        footer { margin-top: 30px; padding-top: 15px; border-top: 1px solid #ddd; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>${_REPORT_SCANNER} Report</h1>
EOF

    # Add quote if available
    if [[ -n "$quote" ]]; then
        cat << EOF
        <blockquote style="font-style: italic; color: #666; border-left: 3px solid #0066cc; padding-left: 15px; margin: 20px 0;">
            "${quote}"
        </blockquote>
EOF
    fi

    cat << EOF
        <div class="meta">
            <strong>Target:</strong> ${_REPORT_TARGET}<br>
            <strong>Generated:</strong> ${_REPORT_TIMESTAMP}<br>
            <strong>Scanner Version:</strong> ${_REPORT_VERSION}
        </div>

        <h2>Summary</h2>
        <div class="summary">
            <div class="summary-card critical">
                <div class="count">$(echo "$json" | jq -r '.summary.by_severity.critical')</div>
                <div class="label">Critical</div>
            </div>
            <div class="summary-card high">
                <div class="count">$(echo "$json" | jq -r '.summary.by_severity.high')</div>
                <div class="label">High</div>
            </div>
            <div class="summary-card medium">
                <div class="count">$(echo "$json" | jq -r '.summary.by_severity.medium')</div>
                <div class="label">Medium</div>
            </div>
            <div class="summary-card low">
                <div class="count">$(echo "$json" | jq -r '.summary.by_severity.low')</div>
                <div class="label">Low</div>
            </div>
        </div>
EOF

    # Findings section
    if [[ ${#_REPORT_FINDINGS[@]} -gt 0 ]]; then
        echo "        <h2>Findings</h2>"

        for finding in "${_REPORT_FINDINGS[@]}"; do
            local severity=$(echo "$finding" | jq -r '.severity')
            local title=$(echo "$finding" | jq -r '.title')
            local location=$(echo "$finding" | jq -r '.location')
            local desc=$(echo "$finding" | jq -r '.description')
            local fix=$(echo "$finding" | jq -r '.remediation')

            cat << EOF
        <div class="finding $severity">
            <h3><span class="badge $severity">${severity^^}</span> $title</h3>
EOF
            [[ -n "$location" && "$location" != "null" ]] && echo "            <p><span class=\"location\">$location</span></p>"
            [[ -n "$desc" && "$desc" != "null" ]] && echo "            <p>$desc</p>"
            [[ -n "$fix" && "$fix" != "null" ]] && echo "            <p><strong>Remediation:</strong> $fix</p>"
            echo "        </div>"
        done
    fi

    cat << EOF

        <footer>
            Generated by ${_REPORT_SCANNER} v${_REPORT_VERSION} •
            <a href="https://crashoverride.com" style="color: #0066cc;">crashoverride.com</a>
            <br><em>Hack the planet.</em>
        </footer>
    </div>
</body>
</html>
EOF
}

#############################################################################
# MAIN GENERATION FUNCTION
#############################################################################

# Generate report in specified format
# Usage: report_generate "markdown" | "json" | "terminal" | "html" | "csv"
report_generate() {
    local format="${1:-terminal}"

    case "$format" in
        json)
            _report_to_json
            ;;
        markdown|md)
            _report_to_markdown
            ;;
        terminal|term)
            _report_to_terminal
            ;;
        html)
            _report_to_html
            ;;
        csv)
            _report_to_csv
            ;;
        *)
            echo "Unknown format: $format" >&2
            echo "Supported: json, markdown, terminal, html, csv" >&2
            return 1
            ;;
    esac
}

# Save report to file
# Usage: report_save "output.md" "markdown"
report_save() {
    local output_file="$1"
    local format="${2:-}"

    # Auto-detect format from extension if not specified
    if [[ -z "$format" ]]; then
        case "${output_file##*.}" in
            json) format="json" ;;
            md)   format="markdown" ;;
            html) format="html" ;;
            csv)  format="csv" ;;
            *)    format="terminal" ;;
        esac
    fi

    report_generate "$format" > "$output_file"
    scanner_success "Report saved to $output_file"
}

#############################################################################
# EXPORT FUNCTIONS
#############################################################################

export -f report_init
export -f report_set_metadata
export -f report_set_metadata_json
export -f report_add_section
export -f report_add_finding
export -f report_generate
export -f report_save
