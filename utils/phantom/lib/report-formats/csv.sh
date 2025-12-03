#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# CSV Report Format
# Exports report data as CSV for spreadsheet analysis
#############################################################################

# Render report as CSV
render_report() {
    local report_type="$1"
    local json_data="$2"
    local output_file="$3"

    local csv_output=""

    case "$report_type" in
        summary)
            csv_output=$(render_summary_csv "$json_data")
            ;;
        security)
            csv_output=$(render_security_csv "$json_data")
            ;;
        compliance)
            csv_output=$(render_compliance_csv "$json_data")
            ;;
        licenses)
            csv_output=$(render_licenses_csv "$json_data")
            ;;
        supply-chain)
            csv_output=$(render_supply_chain_csv "$json_data")
            ;;
        dora)
            csv_output=$(render_dora_csv "$json_data")
            ;;
        sbom)
            csv_output=$(render_sbom_csv "$json_data")
            ;;
        full)
            csv_output=$(render_full_csv "$json_data")
            ;;
        *)
            csv_output=$(render_generic_csv "$json_data")
            ;;
    esac

    if [[ -n "$output_file" ]]; then
        echo "$csv_output" > "$output_file"
        echo "CSV report saved to: $output_file"
    else
        echo "$csv_output"
    fi
}

# Escape CSV field (handle commas, quotes, newlines)
csv_escape() {
    local value="$1"
    # Replace double quotes with two double quotes
    value="${value//\"/\"\"}"
    # If value contains comma, quote, or newline, wrap in quotes
    if [[ "$value" == *","* ]] || [[ "$value" == *"\""* ]] || [[ "$value" == *$'\n'* ]]; then
        echo "\"$value\""
    else
        echo "$value"
    fi
}

# Render summary report as CSV
render_summary_csv() {
    local json_data="$1"

    # Header row
    echo "Category,Metric,Value,Status"

    # Project info
    local project_id=$(echo "$json_data" | jq -r '.project.id // "unknown"')
    local scan_id=$(echo "$json_data" | jq -r '.project.scan_id // "unknown"')
    local profile=$(echo "$json_data" | jq -r '.project.profile // "unknown"')
    local completed=$(echo "$json_data" | jq -r '.project.completed_at // ""')

    echo "Project,ID,$(csv_escape "$project_id"),"
    echo "Project,Scan ID,$(csv_escape "$scan_id"),"
    echo "Project,Profile,$(csv_escape "$profile"),"
    echo "Project,Completed,$(csv_escape "$completed"),"

    # Risk
    local risk=$(echo "$json_data" | jq -r '.risk.level // "unknown"')
    echo "Risk,Level,$(csv_escape "$risk"),$(get_risk_status "$risk")"

    # Scores
    local overall=$(echo "$json_data" | jq -r '.scores.overall // "N/A"')
    local security=$(echo "$json_data" | jq -r '.scores.security // "N/A"')
    local compliance=$(echo "$json_data" | jq -r '.scores.compliance // "N/A"')
    local supply_chain=$(echo "$json_data" | jq -r '.scores.supply_chain // "N/A"')

    echo "Score,Overall,$overall,$(get_score_status "$overall")"
    echo "Score,Security,$security,$(get_score_status "$security")"
    echo "Score,Compliance,$compliance,$(get_score_status "$compliance")"
    echo "Score,Supply Chain,$supply_chain,$(get_score_status "$supply_chain")"

    # Vulnerabilities
    local critical=$(echo "$json_data" | jq -r '.security.vulnerabilities.critical // 0')
    local high=$(echo "$json_data" | jq -r '.security.vulnerabilities.high // 0')
    local medium=$(echo "$json_data" | jq -r '.security.vulnerabilities.medium // 0')
    local low=$(echo "$json_data" | jq -r '.security.vulnerabilities.low // 0')

    echo "Vulnerabilities,Critical,$critical,$(get_vuln_status "critical" "$critical")"
    echo "Vulnerabilities,High,$high,$(get_vuln_status "high" "$high")"
    echo "Vulnerabilities,Medium,$medium,$(get_vuln_status "medium" "$medium")"
    echo "Vulnerabilities,Low,$low,OK"

    # Secrets
    local secrets=$(echo "$json_data" | jq -r '.security.secrets.total // 0')
    echo "Secrets,Total,$secrets,$(get_secrets_status "$secrets")"

    # Dependencies
    local total_deps=$(echo "$json_data" | jq -r '.dependencies.total // 0')
    local direct_deps=$(echo "$json_data" | jq -r '.dependencies.direct // 0')

    echo "Dependencies,Total,$total_deps,"
    echo "Dependencies,Direct,$direct_deps,"
}

# Render security report as CSV
render_security_csv() {
    local json_data="$1"

    echo "Category,Severity,Count,Scanner,Details"

    # Vulnerabilities by severity
    local critical=$(echo "$json_data" | jq -r '.vulnerabilities.critical // 0')
    local high=$(echo "$json_data" | jq -r '.vulnerabilities.high // 0')
    local medium=$(echo "$json_data" | jq -r '.vulnerabilities.medium // 0')
    local low=$(echo "$json_data" | jq -r '.vulnerabilities.low // 0')

    echo "Vulnerability,Critical,$critical,package-vulns,"
    echo "Vulnerability,High,$high,package-vulns,"
    echo "Vulnerability,Medium,$medium,package-vulns,"
    echo "Vulnerability,Low,$low,package-vulns,"

    # Secrets
    local secrets=$(echo "$json_data" | jq -r '.secrets.total // 0')
    echo "Secret,All,$secrets,code-secrets,"

    # Secret types breakdown
    echo "$json_data" | jq -r '.secrets.by_type // {} | to_entries[] | "Secret,\(.key),\(.value),code-secrets,"' 2>/dev/null

    # Code security findings
    local code_sec=$(echo "$json_data" | jq -r '.code_security.total // 0')
    echo "Code Security,All,$code_sec,code-security,"

    # IaC findings
    local iac=$(echo "$json_data" | jq -r '.iac_security.total // 0')
    echo "IaC Security,All,$iac,iac-security,"

    # Individual vulnerabilities if available
    echo ""
    echo "# Vulnerability Details"
    echo "Package,Version,CVE,Severity,Fixed Version"
    echo "$json_data" | jq -r '.vulnerability_details // [] | .[] | "\(.package // "unknown"),\(.version // "unknown"),\(.cve // "N/A"),\(.severity // "unknown"),\(.fixed_version // "N/A")"' 2>/dev/null
}

# Render compliance report as CSV
render_compliance_csv() {
    local json_data="$1"

    echo "Category,Item,Status,Details"

    # License status
    local lic_status=$(echo "$json_data" | jq -r '.licenses.overall_status // "unknown"')
    local violations=$(echo "$json_data" | jq -r '.licenses.violations // 0')
    local copyleft=$(echo "$json_data" | jq -r '.licenses.copyleft_count // 0')

    echo "License,Overall Status,$lic_status,"
    echo "License,Violations,$violations,$([ "$violations" -gt 0 ] && echo "FAIL" || echo "PASS")"
    echo "License,Copyleft Packages,$copyleft,"

    # Documentation
    local has_readme=$(echo "$json_data" | jq -r '.documentation.has_readme // false')
    local has_license=$(echo "$json_data" | jq -r '.documentation.has_license // false')
    local has_contrib=$(echo "$json_data" | jq -r '.documentation.has_contributing // false')
    local has_coc=$(echo "$json_data" | jq -r '.documentation.has_code_of_conduct // false')
    local doc_score=$(echo "$json_data" | jq -r '.documentation.score // 0')

    echo "Documentation,README,$([ "$has_readme" == "true" ] && echo "Present" || echo "Missing"),$([ "$has_readme" == "true" ] && echo "PASS" || echo "WARN")"
    echo "Documentation,LICENSE,$([ "$has_license" == "true" ] && echo "Present" || echo "Missing"),$([ "$has_license" == "true" ] && echo "PASS" || echo "WARN")"
    echo "Documentation,CONTRIBUTING,$([ "$has_contrib" == "true" ] && echo "Present" || echo "Missing"),"
    echo "Documentation,CODE_OF_CONDUCT,$([ "$has_coc" == "true" ] && echo "Present" || echo "Missing"),"
    echo "Documentation,Score,$doc_score,"

    # Ownership
    local has_codeowners=$(echo "$json_data" | jq -r '.ownership.has_codeowners // false')
    local bus_factor=$(echo "$json_data" | jq -r '.ownership.bus_factor // 0')

    echo "Ownership,CODEOWNERS,$([ "$has_codeowners" == "true" ] && echo "Present" || echo "Missing"),"
    echo "Ownership,Bus Factor,$bus_factor,$([ "$bus_factor" -lt 2 ] && echo "RISK" || echo "OK")"

    # License breakdown
    echo ""
    echo "# License Breakdown"
    echo "License Type,Package Count"
    echo "$json_data" | jq -r '.licenses.by_type // {} | to_entries[] | "\(.key),\(.value)"' 2>/dev/null
}

# Render licenses report as CSV
render_licenses_csv() {
    local json_data="$1"

    echo "# License Report"
    echo "Category,License,Package Count,Status"

    # Repository license
    local repo_license=$(echo "$json_data" | jq -r '.repository_license.license // "Not Found"')
    local repo_file=$(echo "$json_data" | jq -r '.repository_license.file // ""')
    echo "Repository,$repo_license,1,$repo_file"

    echo ""
    echo "# Dependency Licenses"
    echo "License,Package Count,Category"

    # All licenses
    echo "$json_data" | jq -r '
        .dependency_licenses.by_license // {} | to_entries | sort_by(-.value.count)[] |
        .key as $license |
        .value.count as $count |
        (if ($license | test("GPL|AGPL"; "i")) then "DENIED"
         elif ($license | test("LGPL|MPL|EPL"; "i")) then "REVIEW"
         else "ALLOWED" end) as $category |
        "\($license),\($count),\($category)"
    ' 2>/dev/null

    echo ""
    echo "# Denied License Packages"
    echo "License,Package,Version"
    echo "$json_data" | jq -r '
        .dependency_licenses.denied // [] | .[] |
        .license as $lic | .packages[]? | "\($lic),\(.name),\(.version)"
    ' 2>/dev/null

    echo ""
    echo "# Review Required Packages"
    echo "License,Package,Version"
    echo "$json_data" | jq -r '
        .dependency_licenses.review_required // [] | .[] |
        .license as $lic | .packages[]? | "\($lic),\(.name),\(.version)"
    ' 2>/dev/null
}

# Render supply chain report as CSV
render_supply_chain_csv() {
    local json_data="$1"

    echo "Category,Metric,Value,Status"

    # Overview
    local score=$(echo "$json_data" | jq -r '.supply_chain_score // 0')
    local risk=$(echo "$json_data" | jq -r '.risk.level // "unknown"')

    echo "Overview,Supply Chain Score,$score,$(get_score_status "$score")"
    echo "Overview,Risk Level,$risk,$(get_risk_status "$risk")"

    # Dependencies
    local total=$(echo "$json_data" | jq -r '.dependencies.total // 0')
    local direct=$(echo "$json_data" | jq -r '.dependencies.direct // 0')
    local transitive=$(echo "$json_data" | jq -r '.dependencies.transitive // 0')

    echo "Dependencies,Total,$total,"
    echo "Dependencies,Direct,$direct,"
    echo "Dependencies,Transitive,$transitive,"

    # Health issues
    local abandoned=$(echo "$json_data" | jq -r '.health.abandoned // 0')
    local deprecated=$(echo "$json_data" | jq -r '.health.deprecated // 0')
    local outdated=$(echo "$json_data" | jq -r '.health.outdated // 0')
    local typosquat=$(echo "$json_data" | jq -r '.health.typosquatting_suspects // 0')

    echo "Health,Abandoned,$abandoned,$([ "$abandoned" -gt 0 ] && echo "WARN" || echo "OK")"
    echo "Health,Deprecated,$deprecated,$([ "$deprecated" -gt 0 ] && echo "WARN" || echo "OK")"
    echo "Health,Outdated,$outdated,"
    echo "Health,Typosquatting Suspects,$typosquat,$([ "$typosquat" -gt 0 ] && echo "CRITICAL" || echo "OK")"

    # Provenance
    local signed=$(echo "$json_data" | jq -r '.provenance.signed_packages // 0')
    local slsa=$(echo "$json_data" | jq -r '.provenance.slsa_level // "none"')
    local signed_pct=$(echo "$json_data" | jq -r '.provenance.signed_percentage // 0')

    echo "Provenance,Signed Packages,$signed,"
    echo "Provenance,Signed Percentage,${signed_pct}%,"
    echo "Provenance,SLSA Level,$slsa,"

    # Package list
    echo ""
    echo "# Dependencies"
    echo "Name,Version,Ecosystem,Direct,License"
    echo "$json_data" | jq -r '.dependencies.packages // [] | .[] | "\(.name // "unknown"),\(.version // "unknown"),\(.ecosystem // "unknown"),\(.direct // false),\(.license // "unknown")"' 2>/dev/null
}

# Render DORA report as CSV
render_dora_csv() {
    local json_data="$1"

    echo "Metric,Level,Value,Unit,Benchmark"

    # Overall
    local overall=$(echo "$json_data" | jq -r '.dora.overall_performance // "N/A"')
    echo "Overall Performance,$overall,,,"

    # Four key metrics
    local df_level=$(echo "$json_data" | jq -r '.dora.metrics.deployment_frequency.level // "N/A"')
    local df_value=$(echo "$json_data" | jq -r '.dora.metrics.deployment_frequency.value // 0')
    local df_unit=$(echo "$json_data" | jq -r '.dora.metrics.deployment_frequency.unit // "per_day"')

    local lt_level=$(echo "$json_data" | jq -r '.dora.metrics.lead_time.level // "N/A"')
    local lt_hours=$(echo "$json_data" | jq -r '.dora.metrics.lead_time.hours // 0')

    local cfr_level=$(echo "$json_data" | jq -r '.dora.metrics.change_failure_rate.level // "N/A"')
    local cfr_pct=$(echo "$json_data" | jq -r '.dora.metrics.change_failure_rate.percentage // 0')

    local mttr_level=$(echo "$json_data" | jq -r '.dora.metrics.mttr.level // "N/A"')
    local mttr_hours=$(echo "$json_data" | jq -r '.dora.metrics.mttr.hours // 0')

    echo "Deployment Frequency,$df_level,$df_value,$df_unit,Elite: multiple/day"
    echo "Lead Time for Changes,$lt_level,$lt_hours,hours,Elite: <1 hour"
    echo "Change Failure Rate,$cfr_level,$cfr_pct,%,Elite: <5%"
    echo "Mean Time to Recovery,$mttr_level,$mttr_hours,hours,Elite: <1 hour"

    # Git activity
    echo ""
    echo "# Git Activity"
    echo "Metric,Value"
    local commits=$(echo "$json_data" | jq -r '.git_activity.total_commits // 0')
    local contributors=$(echo "$json_data" | jq -r '.git_activity.contributors // 0')
    local avg_commits=$(echo "$json_data" | jq -r '.git_activity.avg_commits_per_day // 0')

    echo "Total Commits,$commits"
    echo "Contributors,$contributors"
    echo "Avg Commits/Day,$avg_commits"

    # Recommendations
    echo ""
    echo "# Recommendations"
    echo "Recommendation"
    echo "$json_data" | jq -r '.recommendations // [] | .[] | "\"\(.)\"" ' 2>/dev/null
}

# Render SBOM report as CSV
render_sbom_csv() {
    local json_data="$1"

    # Summary section
    echo "# SBOM Summary"
    echo "Metric,Value"

    local total=$(echo "$json_data" | jq -r '.summary.total_components // 0')
    local direct=$(echo "$json_data" | jq -r '.summary.direct_dependencies // 0')
    local transitive=$(echo "$json_data" | jq -r '.summary.transitive_dependencies // 0')
    local signed=$(echo "$json_data" | jq -r '.summary.signed_components // 0')

    echo "Total Components,$total"
    echo "Direct Dependencies,$direct"
    echo "Transitive Dependencies,$transitive"
    echo "Signed Components,$signed"

    # Ecosystem breakdown
    echo ""
    echo "# By Ecosystem"
    echo "Ecosystem,Count"
    echo "$json_data" | jq -r '.summary.ecosystems // {} | to_entries[] | "\(.key),\(.value)"' 2>/dev/null

    # Full component list (this is the main SBOM data)
    echo ""
    echo "# Components"
    echo "Name,Version,Ecosystem,Direct,License"
    echo "$json_data" | jq -r '.components // [] | .[] | "\(.name // "unknown"),\(.version // "unknown"),\(.ecosystem // "unknown"),\(.direct // false),\(.license // "unknown")"' 2>/dev/null
}

# Render full report as CSV (multi-sheet style with sections)
render_full_csv() {
    local json_data="$1"

    # Overview section
    echo "# Overview"
    echo "Metric,Value,Status"

    local project_id=$(echo "$json_data" | jq -r '.project.id // "unknown"')
    local risk=$(echo "$json_data" | jq -r '.risk.level // "unknown"')
    local overall=$(echo "$json_data" | jq -r '.scores.overall // 0')
    local security=$(echo "$json_data" | jq -r '.scores.security // 0')
    local compliance=$(echo "$json_data" | jq -r '.scores.compliance // 0')
    local supply_chain=$(echo "$json_data" | jq -r '.scores.supply_chain // 0')

    echo "Project,$project_id,"
    echo "Risk Level,$risk,$(get_risk_status "$risk")"
    echo "Overall Score,$overall,$(get_score_status "$overall")"
    echo "Security Score,$security,$(get_score_status "$security")"
    echo "Compliance Score,$compliance,$(get_score_status "$compliance")"
    echo "Supply Chain Score,$supply_chain,$(get_score_status "$supply_chain")"

    # Security section
    echo ""
    echo "# Security"
    echo "Category,Severity,Count"

    local critical=$(echo "$json_data" | jq -r '.security.vulnerabilities.critical // 0')
    local high=$(echo "$json_data" | jq -r '.security.vulnerabilities.high // 0')
    local medium=$(echo "$json_data" | jq -r '.security.vulnerabilities.medium // 0')
    local low=$(echo "$json_data" | jq -r '.security.vulnerabilities.low // 0')
    local secrets=$(echo "$json_data" | jq -r '.security.secrets.total // 0')
    local code_sec=$(echo "$json_data" | jq -r '.security.code_security_findings // 0')
    local iac=$(echo "$json_data" | jq -r '.security.iac_security_findings // 0')

    echo "Vulnerabilities,Critical,$critical"
    echo "Vulnerabilities,High,$high"
    echo "Vulnerabilities,Medium,$medium"
    echo "Vulnerabilities,Low,$low"
    echo "Secrets,Total,$secrets"
    echo "Code Security,Total,$code_sec"
    echo "IaC Security,Total,$iac"

    # Dependencies section
    echo ""
    echo "# Dependencies"
    echo "Metric,Value"

    local total_deps=$(echo "$json_data" | jq -r '.dependencies.total // 0')
    local direct_deps=$(echo "$json_data" | jq -r '.dependencies.direct // 0')
    local abandoned=$(echo "$json_data" | jq -r '.dependencies.health.abandoned // 0')
    local deprecated=$(echo "$json_data" | jq -r '.dependencies.health.deprecated // 0')
    local typosquat=$(echo "$json_data" | jq -r '.dependencies.health.typosquatting_suspects // 0')

    echo "Total,$total_deps"
    echo "Direct,$direct_deps"
    echo "Abandoned,$abandoned"
    echo "Deprecated,$deprecated"
    echo "Typosquatting Suspects,$typosquat"

    # Compliance section
    echo ""
    echo "# Compliance"
    echo "Metric,Value"

    local lic_status=$(echo "$json_data" | jq -r '.compliance.license_status // "unknown"')
    local violations=$(echo "$json_data" | jq -r '.compliance.license_violations // 0')
    local copyleft=$(echo "$json_data" | jq -r '.compliance.copyleft_packages // 0')
    local has_readme=$(echo "$json_data" | jq -r '.compliance.documentation.has_readme // false')
    local has_license=$(echo "$json_data" | jq -r '.compliance.documentation.has_license_file // false')

    echo "License Status,$lic_status"
    echo "License Violations,$violations"
    echo "Copyleft Packages,$copyleft"
    echo "Has README,$has_readme"
    echo "Has LICENSE,$has_license"

    # DevOps section
    echo ""
    echo "# DevOps"
    echo "Metric,Value"

    local dora_perf=$(echo "$json_data" | jq -r '.devops.dora.performance // "N/A"')
    local deploy_freq=$(echo "$json_data" | jq -r '.devops.dora.deployment_frequency // "N/A"')
    local commits=$(echo "$json_data" | jq -r '.devops.git.total_commits // 0')
    local contributors=$(echo "$json_data" | jq -r '.devops.git.contributors // 0')
    local test_cov=$(echo "$json_data" | jq -r '.devops.test_coverage_percentage // 0')

    echo "DORA Performance,$dora_perf"
    echo "Deployment Frequency,$deploy_freq"
    echo "Total Commits,$commits"
    echo "Contributors,$contributors"
    echo "Test Coverage,${test_cov}%"

    # Ownership section
    echo ""
    echo "# Ownership"
    echo "Metric,Value"

    local bus_factor=$(echo "$json_data" | jq -r '.ownership.bus_factor // 0')
    local has_codeowners=$(echo "$json_data" | jq -r '.ownership.has_codeowners // false')
    local own_coverage=$(echo "$json_data" | jq -r '.ownership.coverage_percentage // 0')

    echo "Bus Factor,$bus_factor"
    echo "Has CODEOWNERS,$has_codeowners"
    echo "Ownership Coverage,${own_coverage}%"

    # Provenance section
    echo ""
    echo "# Provenance"
    echo "Metric,Value"

    local signed=$(echo "$json_data" | jq -r '.provenance.signed_packages // 0')
    local slsa=$(echo "$json_data" | jq -r '.provenance.slsa_level // "none"')

    echo "Signed Packages,$signed"
    echo "SLSA Level,$slsa"
}

# Render generic CSV (flatten JSON to key-value pairs)
render_generic_csv() {
    local json_data="$1"

    echo "Key,Value"
    # Flatten JSON to dotted paths
    echo "$json_data" | jq -r '
        paths(scalars) as $p |
        [($p | map(tostring) | join(".")), getpath($p)] |
        @csv
    ' 2>/dev/null
}

# Helper functions for status
get_risk_status() {
    local risk="$1"
    case "$risk" in
        critical) echo "CRITICAL" ;;
        high) echo "HIGH" ;;
        medium) echo "MEDIUM" ;;
        low) echo "OK" ;;
        *) echo "" ;;
    esac
}

get_score_status() {
    local score="$1"
    if [[ "$score" == "N/A" ]]; then
        echo ""
    elif [[ "$score" -ge 80 ]]; then
        echo "GOOD"
    elif [[ "$score" -ge 60 ]]; then
        echo "FAIR"
    elif [[ "$score" -ge 40 ]]; then
        echo "POOR"
    else
        echo "CRITICAL"
    fi
}

get_vuln_status() {
    local severity="$1"
    local count="$2"

    if [[ "$count" -eq 0 ]]; then
        echo "OK"
    else
        case "$severity" in
            critical) echo "CRITICAL" ;;
            high) echo "HIGH" ;;
            medium) echo "MEDIUM" ;;
            *) echo "OK" ;;
        esac
    fi
}

get_secrets_status() {
    local count="$1"
    if [[ "$count" -eq 0 ]]; then
        echo "OK"
    else
        echo "CRITICAL"
    fi
}

# Bridge function to match format expected by report.sh
format_report_output() {
    local json_data="$1"
    local target_id="$2"
    local report_type=$(echo "$json_data" | jq -r '.report_type // "summary"')
    render_report "$report_type" "$json_data" ""
}

export -f format_report_output
export -f render_report
export -f csv_escape
export -f render_summary_csv
export -f render_security_csv
export -f render_compliance_csv
export -f render_licenses_csv
export -f render_supply_chain_csv
export -f render_dora_csv
export -f render_sbom_csv
export -f render_full_csv
export -f render_generic_csv
export -f get_risk_status
export -f get_score_status
export -f get_vuln_status
export -f get_secrets_status
