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
        code-ownership)
            format_code_ownership_markdown "$json_data" "$target_id"
            ;;
        ai-adoption)
            format_ai_adoption_markdown "$json_data" "$target_id"
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

# Format code-ownership report for markdown (3-tier view)
format_code_ownership_markdown() {
    local json="$1"
    local target_id="$2"

    local project_id=$(echo "$json" | jq -r '.project.id // "Unknown"')
    local profile=$(echo "$json" | jq -r '.project.profile // "standard"')
    local completed_at=$(echo "$json" | jq -r '.project.completed_at // ""')

    local tier1=$(echo "$json" | jq -r '.tiers.basic // false')
    local tier2=$(echo "$json" | jq -r '.tiers.analysis // false')
    local tier3=$(echo "$json" | jq -r '.tiers.ai_insights // false')

    cat << EOF
# Code Ownership Report

**Project:** ${project_id}
**Scanned:** ${completed_at}
**Profile:** ${profile}

---

## Tier 1: Basic (CODEOWNERS Detection)

EOF

    if [[ "$tier1" == "true" ]]; then
        local codeowners_exists=$(echo "$json" | jq -r '.tier1_basic.codeowners.exists // false')
        if [[ "$codeowners_exists" == "true" ]]; then
            local codeowners_path=$(echo "$json" | jq -r '.tier1_basic.codeowners.path // ""')
            local codeowners_valid=$(echo "$json" | jq -r '.tier1_basic.codeowners.valid // "unknown"')
            local codeowners_patterns=$(echo "$json" | jq -r '.tier1_basic.codeowners.total_patterns // 0')
            local codeowners_owners=$(echo "$json" | jq -r '.tier1_basic.codeowners.unique_owners // 0')
            echo "✅ **CODEOWNERS File:** Present (\`${codeowners_path}\`)"
            echo "- **Syntax:** ${codeowners_valid}"
            echo "- **Patterns:** ${codeowners_patterns}"
            echo "- **Unique Owners:** ${codeowners_owners}"
        else
            echo "⚠️ **CODEOWNERS File:** Not Found"
        fi
    else
        echo "*No CODEOWNERS data available*"
    fi

    cat << EOF

---

## Tier 2: Analysis (Bus Factor & Concentration)

EOF

    if [[ "$tier2" == "true" ]]; then
        local bus_factor=$(echo "$json" | jq -r '.tier2_analysis.bus_factor.value // 0')
        local risk_level=$(echo "$json" | jq -r '.tier2_analysis.bus_factor.risk_level // "unknown"')
        local risk_desc=$(echo "$json" | jq -r '.tier2_analysis.bus_factor.risk_description // ""')
        local gini=$(echo "$json" | jq -r '.tier2_analysis.concentration.gini_coefficient // 0')
        local top1_pct=$(echo "$json" | jq -r '.tier2_analysis.concentration.top_contributor_percentage // 0')
        local top3_pct=$(echo "$json" | jq -r '.tier2_analysis.concentration.top_3_contributors_percentage // 0')
        local total_commits=$(echo "$json" | jq -r '.tier2_analysis.summary.total_commits // 0')
        local active_contributors=$(echo "$json" | jq -r '.tier2_analysis.summary.active_contributors // 0')

        local risk_emoji="🟢"
        [[ "$risk_level" == "medium" ]] && risk_emoji="🟡"
        [[ "$risk_level" == "high" ]] && risk_emoji="🟠"
        [[ "$risk_level" == "critical" ]] && risk_emoji="🔴"
        local risk_level_upper=$(echo "$risk_level" | tr '[:lower:]' '[:upper:]')

        echo "### Bus Factor: ${bus_factor} ${risk_emoji} (${risk_level_upper})"
        echo ""
        echo "> ${risk_desc}"
        echo ""
        echo "### Concentration Metrics"
        echo ""
        echo "| Metric | Value |"
        echo "|--------|-------|"
        echo "| Gini Coefficient | ${gini} |"
        echo "| Top Contributor | ${top1_pct}% |"
        echo "| Top 3 Contributors | ${top3_pct}% |"
        echo ""
        echo "### Activity"
        echo ""
        echo "- **Total Commits:** ${total_commits}"
        echo "- **Active Contributors:** ${active_contributors}"
        echo ""
        echo "### Top Contributors"
        echo ""
        echo "| Contributor | Commits | Ownership |"
        echo "|-------------|---------|-----------|"
        echo "$json" | jq -r '.tier2_analysis.contributors[:5][] | "| \(.name) | \(.commits) | \(.ownership_percentage)% |"' 2>/dev/null
    else
        echo "*No bus factor analysis available - run bus-factor scanner*"
    fi

    cat << EOF

---

## Tier 3: AI Insights (Claude Analysis)

EOF

    if [[ "$tier3" == "true" ]]; then
        local ai_model=$(echo "$json" | jq -r '.tier3_ai_insights.model // ""')
        if [[ -n "$ai_model" ]]; then
            echo "*Analyzed by: ${ai_model}*"
            echo ""
        fi

        echo "### Key Insights"
        echo ""
        echo "$json" | jq -r '.tier3_ai_insights.insights[]? | "- \(.)"' 2>/dev/null
        echo ""
        echo "### Recommendations"
        echo ""
        echo "$json" | jq -r '.tier3_ai_insights.recommendations[]? | "- \(.)"' 2>/dev/null

        # Risk areas
        local risk_count=$(echo "$json" | jq -r '.tier3_ai_insights.risk_areas | length // 0')
        if [[ "$risk_count" -gt 0 ]]; then
            echo ""
            echo "### Risk Areas"
            echo ""
            echo "| Area | Owner | Risk | Reason |"
            echo "|------|-------|------|--------|"
            echo "$json" | jq -r '.tier3_ai_insights.risk_areas[]? | "| \(.area) | \(.owner) | \(.risk | ascii_upcase) | \(.reason) |"' 2>/dev/null
        fi

        # Action items
        local action_count=$(echo "$json" | jq -r '.tier3_ai_insights.action_items | length // 0')
        if [[ "$action_count" -gt 0 ]]; then
            echo ""
            echo "### Action Items"
            echo ""
            echo "| Priority | Action | Owner |"
            echo "|----------|--------|-------|"
            echo "$json" | jq -r '.tier3_ai_insights.action_items[]? | "| \(.priority | ascii_upcase) | \(.action) | \(.owner) |"' 2>/dev/null
        fi

        # Suggested CODEOWNERS
        local codeowners_count=$(echo "$json" | jq -r '.tier3_ai_insights.suggested_codeowners | length // 0')
        if [[ "$codeowners_count" -gt 0 ]]; then
            echo ""
            echo "### Suggested CODEOWNERS"
            echo ""
            echo "\`\`\`"
            echo "$json" | jq -r '.tier3_ai_insights.suggested_codeowners[]?' 2>/dev/null
            echo "\`\`\`"
        fi
    else
        echo "*AI analysis not available*"
        echo ""
        echo "Use \`--deep\` profile or run with Claude for insights like:"
        echo "- Critical risk areas and succession planning"
        echo "- Knowledge transfer recommendations"
        echo "- Auto-generated optimal CODEOWNERS file"
    fi

    cat << EOF

---

*Generated by Phantom Report v${REPORT_VERSION:-1.0.0}*
*$(date '+%Y-%m-%d %H:%M:%S %Z')*
EOF
}

# Format AI adoption report for markdown
format_ai_adoption_markdown() {
    local json="$1"
    local target_id="$2"

    local project_id=$(echo "$json" | jq -r '.project.id // "Unknown"')
    local profile=$(echo "$json" | jq -r '.project.profile // "standard"')
    local completed_at=$(echo "$json" | jq -r '.project.completed_at // ""')
    local commit=$(echo "$json" | jq -r '.project.git.commit // ""')
    local branch=$(echo "$json" | jq -r '.project.git.branch // ""')

    local has_ai=$(echo "$json" | jq -r '.summary.has_ai_adoption // false')
    local ai_tech_count=$(echo "$json" | jq -r '.summary.ai_technologies_count // 0')
    local ai_cat_count=$(echo "$json" | jq -r '.summary.ai_categories_count // 0')
    local total_contributors=$(echo "$json" | jq -r '.summary.total_contributors // 0')
    local bus_factor=$(echo "$json" | jq -r '.summary.bus_factor // 0')
    local bus_factor_risk=$(echo "$json" | jq -r '.summary.bus_factor_risk // "unknown"')

    local ai_badge_color="green"
    [[ "$has_ai" == "true" ]] && ai_badge_color="blue"

    cat << EOF
# AI Adoption Report

**Project:** \`${project_id}\`
**Scanned:** ${completed_at}
**Profile:** ${profile}
EOF

    if [[ -n "$commit" ]] && [[ "$commit" != "null" ]]; then
        echo "**Commit:** \`${commit}\` (${branch})"
    fi

    cat << EOF

$(badge "AI%20Adoption" "$(if [[ "$has_ai" == "true" ]]; then echo "Detected"; else echo "None"; fi)" "$ai_badge_color") $(badge "Technologies" "$ai_tech_count" "blue") $(badge "Contributors" "$total_contributors" "lightgrey")

---

## Summary

| Metric | Value |
|--------|-------|
| **AI Technologies** | $ai_tech_count |
| **AI Categories** | $ai_cat_count |
| **Contributors** | $total_contributors |
| **Bus Factor** | $bus_factor ($bus_factor_risk) |

EOF

    if [[ "$has_ai" == "true" ]]; then
        cat << EOF
---

## AI Technologies by Category

EOF

        echo "$json" | jq -r '.categories[] | "### \(.label) (\(.count))\n\n\(.technologies | map("- \(.)") | join("\n"))\n"' 2>/dev/null

        cat << EOF

---

## AI Technology Details

| Technology | Category | Confidence | Detection Method |
|------------|----------|------------|------------------|
EOF

        echo "$json" | jq -r '.ai_technologies[] | "| \(.name) | \(.category) | \(.confidence)% | \(.detection_methods | join(", ")) |"' 2>/dev/null

        cat << EOF

### Evidence

EOF

        echo "$json" | jq -r '.ai_technologies[] | "**\(.name):** \(.evidence | join("; "))"' 2>/dev/null
    else
        cat << EOF
---

## No AI Technologies Detected

This repository does not appear to use any AI/ML technologies based on:
- Package dependencies (SBOM analysis)
- Configuration files
- Environment variables

EOF
    fi

    local contrib_count=$(echo "$json" | jq -r '.contributors | length // 0')
    if [[ "$contrib_count" -gt 0 ]]; then
        cat << EOF

---

## Top Contributors

*Potential AI adopters based on overall contribution activity*

| Rank | Contributor | Commits | Lines Added |
|------|-------------|---------|-------------|
EOF

        local rank=1
        echo "$json" | jq -r '.contributors[:10][] | "\(.name)|\(.commits)|\(.lines_added)"' 2>/dev/null | while IFS='|' read -r name commits lines; do
            [[ -z "$name" ]] && continue
            echo "| $rank | $name | $commits | $lines |"
            ((rank++))
        done
    fi

    cat << EOF

---

## Governance Considerations

EOF

    if [[ "$has_ai" == "true" ]]; then
        cat << EOF
> ⚠️ **AI technologies detected.** Consider the following:

- [ ] Review AI vendor agreements and data policies
- [ ] Ensure compliance with organizational AI guidelines
- [ ] Track AI-related costs and API usage
- [ ] Assess security implications of AI integrations

EOF
    else
        echo "✅ No AI governance concerns identified."
    fi

    cat << EOF

---

*Phase 1 MVP: Shows AI technologies + contributors separately.*
*Phase 2 will add file-level correlation (who uses which AI where).*

---

*Generated by Phantom Report v${REPORT_VERSION:-1.0.0}*
*$(date '+%Y-%m-%d %H:%M:%S %Z')*
EOF
}

export -f format_report_output
export -f format_summary_markdown
export -f format_project_summary_markdown
export -f format_org_summary_markdown
export -f format_security_markdown
export -f format_code_ownership_markdown
export -f format_ai_adoption_markdown
export -f get_badge_color
export -f badge
