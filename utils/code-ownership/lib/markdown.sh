#!/bin/bash
# Copyright (c) 2024 Gibson Powers Contributors
# 
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Markdown Report Generator
# Generate human-readable markdown reports from analysis data
#
# Key Features:
# - Formatted tables (ownership, metrics, SPOFs)
# - Badge generation (health scores, trends)
# - ASCII charts and visualizations
# - Emoji indicators
# - Collapsible sections
# - GitHub-flavored markdown
#############################################################################

# Generate health badge
generate_health_badge() {
    local score="$1"

    if (( $(echo "$score >= 85" | bc -l) )); then
        echo "![Health](https://img.shields.io/badge/Health-Excellent-brightgreen)"
    elif (( $(echo "$score >= 70" | bc -l) )); then
        echo "![Health](https://img.shields.io/badge/Health-Good-green)"
    elif (( $(echo "$score >= 50" | bc -l) )); then
        echo "![Health](https://img.shields.io/badge/Health-Fair-yellow)"
    else
        echo "![Health](https://img.shields.io/badge/Health-Poor-red)"
    fi
}

# Generate trend indicator
generate_trend_indicator() {
    local value="$1"

    if (( $(echo "$value > 0" | bc -l) )); then
        echo "📈 Improving (+$(printf "%.1f" "$value"))"
    elif (( $(echo "$value < 0" | bc -l) )); then
        echo "📉 Declining ($(printf "%.1f" "$value"))"
    else
        echo "➡️ Stable"
    fi
}

# Generate risk emoji
generate_risk_emoji() {
    local risk="$1"

    case "$risk" in
        Critical) echo "🔴" ;;
        High) echo "🟠" ;;
        Medium) echo "🟡" ;;
        Low) echo "🟢" ;;
        *) echo "⚪" ;;
    esac
}

# Generate markdown header
generate_markdown_header() {
    local repo_name="$1"
    local analysis_date="$2"

    cat << EOF
# Code Ownership Report

**Repository:** \`$repo_name\`
**Generated:** $analysis_date
**Analyser Version:** 2.5.0

---

EOF
}

# Generate executive summary section
generate_executive_summary() {
    local json_data="$1"

    local health_score=$(echo "$json_data" | jq -r '.ownership_health.health_score')
    local health_grade=$(echo "$json_data" | jq -r '.ownership_health.health_grade')
    local coverage=$(echo "$json_data" | jq -r '.ownership_health.coverage_percentage')
    local bus_factor=$(echo "$json_data" | jq -r '.ownership_health.bus_factor')
    local gini=$(echo "$json_data" | jq -r '.ownership_health.gini_coefficient')

    local health_badge=$(generate_health_badge "$health_score")

    cat << EOF
## 📊 Executive Summary

$health_badge

| Metric | Value | Status |
|--------|-------|--------|
| **Overall Health** | ${health_score}/100 ($health_grade) | $(if (( $(echo "$health_score >= 70" | bc -l) )); then echo "✅ Good"; else echo "⚠️ Needs Attention"; fi) |
| **Coverage** | ${coverage}% | $(if (( $(echo "$coverage >= 80" | bc -l) )); then echo "✅ Excellent"; elif (( $(echo "$coverage >= 60" | bc -l) )); then echo "⚠️ Fair"; else echo "❌ Poor"; fi) |
| **Bus Factor** | $bus_factor | $(if [[ $bus_factor -ge 3 ]]; then echo "✅ Healthy"; elif [[ $bus_factor -eq 2 ]]; then echo "⚠️ Moderate Risk"; else echo "❌ High Risk"; fi) |
| **Distribution (Gini)** | $gini | $(if (( $(echo "$gini <= 0.5" | bc -l) )); then echo "✅ Well Distributed"; else echo "⚠️ Concentrated"; fi) |

EOF
}

# Generate repository metrics table
generate_repository_metrics() {
    local json_data="$1"

    local total_files=$(echo "$json_data" | jq -r '.repository_metrics.total_files')
    local total_commits=$(echo "$json_data" | jq -r '.repository_metrics.total_commits')
    local active_contributors=$(echo "$json_data" | jq -r '.repository_metrics.active_contributors')
    local analysis_days=$(echo "$json_data" | jq -r '.metadata.time_period_days')

    cat << EOF
## 📈 Repository Metrics

| Metric | Value |
|--------|-------|
| Total Files | $total_files |
| Total Commits (${analysis_days}d) | $total_commits |
| Active Contributors | $active_contributors |
| Analysis Method | $(echo "$json_data" | jq -r '.metadata.analysis_method') |

EOF
}

# Generate top contributors table
generate_contributors_table() {
    local json_data="$1"
    local limit="${2:-10}"

    cat << EOF
## 👥 Top Contributors

| Rank | Contributor | Files Owned |
|------|-------------|-------------|
EOF

    echo "$json_data" | jq -r --arg limit "$limit" '
        .contributors
        | sort_by(-.files_owned)
        | limit($limit | tonumber; .[])
        | "\(.email)|\(.files_owned)"
    ' | awk -F'|' '{
        printf "| %d | `%s` | %d |\n", NR, $1, $2
    }'

    echo ""
}

# Generate SPOF table
generate_spof_table() {
    local json_data="$1"
    local limit="${2:-10}"

    local spof_count=$(echo "$json_data" | jq -r '.single_points_of_failure | length')

    if [[ $spof_count -eq 0 ]]; then
        cat << EOF
## 🎯 Single Points of Failure

✅ **No critical single points of failure detected!**

EOF
        return
    fi

    cat << EOF
## 🎯 Single Points of Failure

⚠️ **$spof_count files identified as potential risks**

| File | Risk | Score | Contributors |
|------|------|-------|--------------|
EOF

    echo "$json_data" | jq -r --arg limit "$limit" '
        .single_points_of_failure
        | sort_by(-.score)
        | limit($limit | tonumber; .[])
        | "\(.file)|\(.risk)|\(.score)|\(.contributors)"
    ' | while IFS='|' read -r file risk score contributors; do
        local emoji=$(generate_risk_emoji "$risk")
        printf "| \`%s\` | %s %s | %d/6 | %d |\n" "$file" "$emoji" "$risk" "$score" "$contributors"
    done

    echo ""
}

# Generate recommendations section
generate_recommendations() {
    local json_data="$1"

    local needs_attention=$(echo "$json_data" | jq -r '.recommendations.needs_attention')
    local bus_factor=$(echo "$json_data" | jq -r '.ownership_health.bus_factor')
    local coverage=$(echo "$json_data" | jq -r '.ownership_health.coverage_percentage')
    local critical_spofs=$(echo "$json_data" | jq -r '.recommendations.critical_spofs')
    local high_spofs=$(echo "$json_data" | jq -r '.recommendations.high_spofs')

    cat << EOF
## 💡 Recommendations

**Overall Status:** $needs_attention

EOF

    if [[ $bus_factor -lt 2 ]]; then
        cat << EOF
### 🔴 Critical: Low Bus Factor

Your bus factor is $bus_factor, meaning the project would stall if $bus_factor person(s) left.

**Actions:**
- Pair programming sessions to share knowledge
- Cross-training initiatives
- Documentation of critical systems
- Code review requirements

EOF
    fi

    if (( $(echo "$coverage < 70" | bc -l) )); then
        cat << EOF
### ⚠️ Warning: Low Coverage

Only ${coverage}% of files have clear ownership.

**Actions:**
- Update CODEOWNERS file
- Assign owners to uncovered files
- Regular ownership reviews

EOF
    fi

    if [[ $critical_spofs -gt 0 ]] || [[ $high_spofs -gt 0 ]]; then
        cat << EOF
### ⚠️ High-Risk Files Detected

- **Critical Risk:** $critical_spofs files
- **High Risk:** $high_spofs files

**Actions:**
- Identify backup owners
- Create knowledge transfer plans
- Increase test coverage
- Add documentation

EOF
    fi

    if [[ $bus_factor -ge 3 ]] && (( $(echo "$coverage >= 80" | bc -l) )) && [[ $critical_spofs -eq 0 ]]; then
        cat << EOF
### ✅ Excellent Ownership Health

Your repository has:
- Strong bus factor ($bus_factor)
- Good coverage (${coverage}%)
- No critical risks

**Keep it up!**

EOF
    fi
}

# Generate distribution visualization
generate_distribution_chart() {
    local json_data="$1"

    local gini=$(echo "$json_data" | jq -r '.ownership_health.gini_coefficient')

    cat << EOF
## 📊 Ownership Distribution

**Gini Coefficient:** $gini
$(if (( $(echo "$gini <= 0.3" | bc -l) )); then echo "✅ Excellent - Very even distribution"; elif (( $(echo "$gini <= 0.5" | bc -l) )); then echo "⚠️ Good - Moderate concentration"; else echo "❌ Poor - High concentration"; fi)

\`\`\`
Perfect Equality (0.0) ◄──────────────► Perfect Inequality (1.0)
                       $(printf "%-${gini}s" "" | tr ' ' '─')▼
                       Current: $gini
\`\`\`

EOF

    # Show top-N concentration
    local top1=$(echo "$json_data" | jq -r '
        .contributors
        | sort_by(-.files_owned)
        | .[0].files_owned
    ')
    local total=$(echo "$json_data" | jq -r '.repository_metrics.total_files')
    local top1_pct=$(echo "scale=1; ($top1 / $total) * 100" | bc -l)

    cat << EOF
**Top Contributor Concentration:**
- Top 1 contributor owns ${top1_pct}% of files
$(if (( $(echo "$top1_pct > 50" | bc -l) )); then echo "  ⚠️ High concentration"; fi)

EOF
}

# Generate full markdown report
generate_markdown_report() {
    local json_data="$1"

    local repo_name=$(echo "$json_data" | jq -r '.metadata.repository')
    local analysis_date=$(echo "$json_data" | jq -r '.metadata.analysis_date')

    # Generate sections
    generate_markdown_header "$repo_name" "$analysis_date"
    generate_executive_summary "$json_data"
    generate_repository_metrics "$json_data"
    generate_contributors_table "$json_data" 10
    generate_distribution_chart "$json_data"
    generate_spof_table "$json_data" 15
    generate_recommendations "$json_data"

    # Footer
    cat << EOF
---

<details>
<summary>📋 Detailed Metrics</summary>

\`\`\`json
$(echo "$json_data" | jq '.')
\`\`\`

</details>

---

*Generated by Code Ownership Analyser v2.5*
*Analysis Method: $(echo "$json_data" | jq -r '.metadata.analysis_method')*

EOF
}

# Generate comparison markdown (two snapshots)
generate_comparison_report() {
    local snapshot1="$1"
    local snapshot2="$2"

    local date1=$(echo "$snapshot1" | jq -r '.metadata.analysis_date')
    local date2=$(echo "$snapshot2" | jq -r '.metadata.analysis_date')

    local health1=$(echo "$snapshot1" | jq -r '.ownership_health.health_score')
    local health2=$(echo "$snapshot2" | jq -r '.ownership_health.health_score')
    local health_delta=$(echo "scale=1; $health2 - $health1" | bc -l)

    local coverage1=$(echo "$snapshot1" | jq -r '.ownership_health.coverage_percentage')
    local coverage2=$(echo "$snapshot2" | jq -r '.ownership_health.coverage_percentage')
    local coverage_delta=$(echo "scale=1; $coverage2 - $coverage1" | bc -l)

    local bus_factor1=$(echo "$snapshot1" | jq -r '.ownership_health.bus_factor')
    local bus_factor2=$(echo "$snapshot2" | jq -r '.ownership_health.bus_factor')
    local bus_factor_delta=$(echo "$bus_factor2 - $bus_factor1" | bc)

    cat << EOF
# Ownership Comparison Report

**Comparing:** \`$date1\` → \`$date2\`

## 📊 Change Summary

| Metric | Before | After | Change |
|--------|--------|-------|--------|
| **Health Score** | ${health1} | ${health2} | $(generate_trend_indicator "$health_delta") |
| **Coverage** | ${coverage1}% | ${coverage2}% | $(generate_trend_indicator "$coverage_delta") |
| **Bus Factor** | ${bus_factor1} | ${bus_factor2} | $(if [[ $bus_factor_delta -gt 0 ]]; then echo "✅ +$bus_factor_delta"; elif [[ $bus_factor_delta -lt 0 ]]; then echo "⚠️ $bus_factor_delta"; else echo "→ No change"; fi) |

EOF
}

# Export functions
export -f generate_health_badge
export -f generate_trend_indicator
export -f generate_risk_emoji
export -f generate_markdown_header
export -f generate_executive_summary
export -f generate_repository_metrics
export -f generate_contributors_table
export -f generate_spof_table
export -f generate_recommendations
export -f generate_distribution_chart
export -f generate_markdown_report
export -f generate_comparison_report
