#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Markdown Format Output
# GitHub-flavored markdown output for reports
#############################################################################

# Format report output to markdown
# Usage: format_report_output <json_data> <target_id>
format_report_output() {
    local json_data="$1"
    local target_id="$2"

    local report_type=$(echo "$json_data" | jq -r '.report_type')

    case "$report_type" in
        summary)
            format_summary_markdown "$json_data" "$target_id"
            ;;
        security)
            format_security_markdown "$json_data" "$target_id"
            ;;
        *)
            format_summary_markdown "$json_data" "$target_id"
            ;;
    esac
}

# Get badge color for shields.io
get_badge_color() {
    local risk="$1"
    case "$risk" in
        critical) echo "red" ;;
        high) echo "orange" ;;
        medium) echo "yellow" ;;
        low) echo "green" ;;
        none) echo "brightgreen" ;;
        *) echo "lightgrey" ;;
    esac
}

# Generate shields.io badge URL
# Usage: badge <label> <value> <color>
badge() {
    local label="$1"
    local value="$2"
    local color="$3"

    # URL encode spaces
    label="${label// /%20}"
    value="${value// /%20}"

    echo "![${label}](https://img.shields.io/badge/${label}-${value}-${color})"
}

# Format summary report for markdown
format_summary_markdown() {
    local json="$1"
    local target_id="$2"

    # Check if this is an org report
    local is_org=$(echo "$json" | jq -r 'has("organization")')

    if [[ "$is_org" == "true" ]]; then
        format_org_summary_markdown "$json"
    else
        format_project_summary_markdown "$json"
    fi
}

# Format project summary for markdown
format_project_summary_markdown() {
    local json="$1"

    # Extract data
    local project_id=$(echo "$json" | jq -r '.project.id')
    local profile=$(echo "$json" | jq -r '.project.profile')
    local completed_at=$(echo "$json" | jq -r '.project.completed_at')
    local commit=$(echo "$json" | jq -r '.project.git.commit')
    local branch=$(echo "$json" | jq -r '.project.git.branch')

    local risk_level=$(echo "$json" | jq -r '.risk.level')
    local critical=$(echo "$json" | jq -r '.risk.vulnerabilities.critical')
    local high=$(echo "$json" | jq -r '.risk.vulnerabilities.high')
    local medium=$(echo "$json" | jq -r '.risk.vulnerabilities.medium')
    local low=$(echo "$json" | jq -r '.risk.vulnerabilities.low')
    local total_vulns=$(echo "$json" | jq -r '.risk.vulnerabilities.total')

    local total_deps=$(echo "$json" | jq -r '.dependencies.total')
    local direct_deps=$(echo "$json" | jq -r '.dependencies.direct')
    local abandoned=$(echo "$json" | jq -r '.dependencies.abandoned')

    local secrets=$(echo "$json" | jq -r '.secrets.exposed')
    local license_status=$(echo "$json" | jq -r '.licenses.status')
    local license_violations=$(echo "$json" | jq -r '.licenses.violations')
    local dora_perf=$(echo "$json" | jq -r '.dora.performance')

    local risk_color=$(get_badge_color "$risk_level")
    local risk_upper=$(echo "$risk_level" | tr '[:lower:]' '[:upper:]')

    # Output markdown
    cat << EOF
# Phantom Summary Report

**Repository:** \`$project_id\`

$(badge "Risk" "$risk_upper" "$risk_color") $(badge "Vulnerabilities" "$total_vulns" "$(if [[ $total_vulns -gt 0 ]]; then echo "orange"; else echo "green"; fi)") $(badge "Dependencies" "$total_deps" "blue") $(badge "Secrets" "$secrets" "$(if [[ $secrets -gt 0 ]]; then echo "red"; else echo "green"; fi)")

---

## Scan Information

| Field | Value |
|-------|-------|
| **Project** | \`$project_id\` |
| **Scanned** | $(format_timestamp "$completed_at") |
| **Profile** | $profile |
EOF

    if [[ -n "$commit" ]] && [[ "$commit" != "null" ]] && [[ "$commit" != "" ]]; then
        echo "| **Commit** | \`$commit\` ($branch) |"
    fi

    cat << EOF

---

## Risk Assessment

### Overall Risk: $(get_risk_emoji "$risk_level") **$risk_upper**

EOF

    # Vulnerability table
    cat << EOF
### Vulnerabilities

| Severity | Count | Status |
|----------|-------|--------|
EOF

    if [[ "$critical" -gt 0 ]]; then
        echo "| 🔴 Critical | **$critical** | Immediate action required |"
    else
        echo "| 🔴 Critical | 0 | ✅ |"
    fi

    if [[ "$high" -gt 0 ]]; then
        echo "| 🟠 High | **$high** | Address this sprint |"
    else
        echo "| 🟠 High | 0 | ✅ |"
    fi

    echo "| 🟡 Medium | $medium | Plan remediation |"
    echo "| 🟢 Low | $low | Monitor |"

    cat << EOF

---

## Dependencies

| Metric | Value |
|--------|-------|
| **Total Packages** | $total_deps |
| **Direct Dependencies** | $direct_deps |
EOF

    if [[ "$abandoned" -gt 0 ]]; then
        echo "| **Abandoned Packages** | ⚠️ $abandoned |"
    fi

    cat << EOF

---

## Security

| Check | Status |
|-------|--------|
EOF

    if [[ "$secrets" -gt 0 ]]; then
        echo "| **Exposed Secrets** | 🔴 $secrets found |"
    else
        echo "| **Exposed Secrets** | ✅ None detected |"
    fi

    if [[ "$license_violations" -gt 0 ]]; then
        echo "| **License Compliance** | ⚠️ $license_violations violations |"
    elif [[ "$license_status" == "pass" ]]; then
        echo "| **License Compliance** | ✅ Compliant |"
    else
        echo "| **License Compliance** | $license_status |"
    fi

    # DORA metrics if available
    if [[ "$dora_perf" != "N/A" ]] && [[ "$dora_perf" != "null" ]]; then
        cat << EOF

---

## DevOps Performance (DORA)

| Metric | Value |
|--------|-------|
EOF
        case "$dora_perf" in
            ELITE) echo "| **Performance Level** | 🚀 Elite |" ;;
            HIGH) echo "| **Performance Level** | 📈 High |" ;;
            MEDIUM) echo "| **Performance Level** | 📊 Medium |" ;;
            LOW) echo "| **Performance Level** | 📉 Low |" ;;
            *) echo "| **Performance Level** | $dora_perf |" ;;
        esac
    fi

    # Top issues
    local top_issues=$(echo "$json" | jq -r '.top_issues[]' 2>/dev/null)
    if [[ -n "$top_issues" ]]; then
        cat << EOF

---

## Top Issues

EOF
        local issue_num=1
        echo "$top_issues" | while read -r issue; do
            echo "$issue_num. $issue"
            ((issue_num++))
        done
    fi

    cat << EOF

---

<details>
<summary>📋 Raw Data</summary>

\`\`\`json
$(echo "$json" | jq '.')
\`\`\`

</details>

---

*Generated by [Phantom](https://github.com/crashoverride/phantom) Report v${REPORT_VERSION}*
*$(date '+%Y-%m-%d %H:%M:%S %Z')*
EOF
}

# Format org summary for markdown
format_org_summary_markdown() {
    local json="$1"

    local org=$(echo "$json" | jq -r '.organization')
    local project_count=$(echo "$json" | jq -r '.projects.count')
    local risk_level=$(echo "$json" | jq -r '.risk.level')
    local total_vulns=$(echo "$json" | jq -r '.risk.vulnerabilities.total')
    local critical=$(echo "$json" | jq -r '.risk.vulnerabilities.critical')
    local high=$(echo "$json" | jq -r '.risk.vulnerabilities.high')
    local total_deps=$(echo "$json" | jq -r '.dependencies.total')

    local risk_color=$(get_badge_color "$risk_level")
    local risk_upper=$(echo "$risk_level" | tr '[:lower:]' '[:upper:]')

    cat << EOF
# Organization Summary Report

**Organization:** \`$org\`

$(badge "Risk" "$risk_upper" "$risk_color") $(badge "Projects" "$project_count" "blue") $(badge "Vulnerabilities" "$total_vulns" "$(if [[ $total_vulns -gt 0 ]]; then echo "orange"; else echo "green"; fi)")

---

## Overview

| Metric | Value |
|--------|-------|
| **Projects Scanned** | $project_count |
| **Overall Risk** | $(get_risk_emoji "$risk_level") $risk_upper |
| **Total Vulnerabilities** | $total_vulns |
| **Total Dependencies** | $total_deps |

---

## Vulnerability Summary

| Severity | Count |
|----------|-------|
| 🔴 Critical | $critical |
| 🟠 High | $high |

EOF

    # At-risk repos
    local at_risk=$(echo "$json" | jq -r '.projects.at_risk[]' 2>/dev/null)
    if [[ -n "$at_risk" ]]; then
        cat << EOF
---

## At-Risk Repositories

The following repositories have critical or high severity vulnerabilities:

EOF
        echo "$at_risk" | while read -r repo; do
            echo "- \`$repo\`"
        done
    fi

    cat << EOF

---

*Generated by [Phantom](https://github.com/crashoverride/phantom) Report v${REPORT_VERSION}*
*$(date '+%Y-%m-%d %H:%M:%S %Z')*
EOF
}

# Format security report for markdown (placeholder)
format_security_markdown() {
    local json="$1"
    local target_id="$2"

    # For now, fall back to summary
    format_summary_markdown "$json" "$target_id"
}

export -f format_report_output
export -f format_summary_markdown
export -f format_project_summary_markdown
export -f format_org_summary_markdown
export -f format_security_markdown
export -f get_badge_color
export -f badge
