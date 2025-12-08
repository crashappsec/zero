#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Full Report Type
# Comprehensive report combining all sections
#############################################################################

# Generate full report data as JSON
generate_report_data() {
    local project_id="$1"
    local analysis_path="$2"

    # Load manifest
    local manifest=$(load_manifest "$analysis_path")
    local scan_id=$(echo "$manifest" | jq -r '.scan_id // "unknown"')
    local profile=$(echo "$manifest" | jq -r '.scan.profile // "standard"')
    local completed_at=$(echo "$manifest" | jq -r '.scan.completed_at // ""')
    local duration=$(echo "$manifest" | jq -r '.scan.duration_seconds // 0')
    local commit=$(echo "$manifest" | jq -r '.git.commit_short // ""')
    local branch=$(echo "$manifest" | jq -r '.git.branch // ""')

    # === SECURITY SECTION ===
    local vulns=$(aggregate_vulns "$analysis_path")
    local critical=$(echo "$vulns" | jq -r '.critical')
    local high=$(echo "$vulns" | jq -r '.high')
    local medium=$(echo "$vulns" | jq -r '.medium')
    local low=$(echo "$vulns" | jq -r '.low')
    local total_vulns=$(echo "$vulns" | jq -r '.total')

    local secrets=0
    local secrets_by_type='{}'
    if has_scanner_data "$analysis_path" "code-secrets"; then
        local sec_data=$(load_scanner_data "$analysis_path" "code-secrets")
        secrets=$(echo "$sec_data" | jq -r '.summary.total_findings // 0')
        secrets_by_type=$(echo "$sec_data" | jq '.summary.by_type // {}')
    fi

    local code_security_count=0
    if has_scanner_data "$analysis_path" "code-security"; then
        code_security_count=$(load_scanner_data "$analysis_path" "code-security" | jq -r '.summary.total // 0')
    fi

    local iac_count=0
    if has_scanner_data "$analysis_path" "iac-security"; then
        iac_count=$(load_scanner_data "$analysis_path" "iac-security" | jq -r '.summary.total // 0')
    fi

    # === DEPENDENCIES SECTION ===
    local deps=$(aggregate_deps "$analysis_path")
    local total_deps=$(echo "$deps" | jq -r '.total')
    local direct_deps=$(echo "$deps" | jq -r '.direct')

    local abandoned=0
    local deprecated=0
    local typosquatting=0
    if has_scanner_data "$analysis_path" "package-health"; then
        local health=$(load_scanner_data "$analysis_path" "package-health")
        abandoned=$(echo "$health" | jq -r '.summary.abandoned // 0')
        deprecated=$(echo "$health" | jq -r '.summary.deprecated // 0')
        typosquatting=$(echo "$health" | jq -r '.summary.typosquatting_suspects // 0')
    fi

    # === COMPLIANCE SECTION ===
    local license_status="unknown"
    local license_violations=0
    local copyleft_count=0
    if has_scanner_data "$analysis_path" "licenses"; then
        local lic=$(load_scanner_data "$analysis_path" "licenses")
        license_status=$(echo "$lic" | jq -r '.summary.overall_status // "unknown"')
        license_violations=$(echo "$lic" | jq -r '.summary.license_violations // 0')
        copyleft_count=$(echo "$lic" | jq -r '.summary.copyleft_count // 0')
    fi

    local has_readme=false
    local has_license_file=false
    local doc_score=0
    if has_scanner_data "$analysis_path" "documentation"; then
        local doc=$(load_scanner_data "$analysis_path" "documentation")
        has_readme=$(echo "$doc" | jq -r '.summary.has_readme // false')
        has_license_file=$(echo "$doc" | jq -r '.summary.has_license // false')
        doc_score=$(echo "$doc" | jq -r '.summary.score // 0')
    fi

    # === DEVOPS SECTION ===
    local dora_perf="N/A"
    local deploy_freq="N/A"
    local lead_time="N/A"
    if has_scanner_data "$analysis_path" "dora"; then
        local dora=$(load_scanner_data "$analysis_path" "dora")
        dora_perf=$(echo "$dora" | jq -r '.summary.overall_performance // "N/A"')
        deploy_freq=$(echo "$dora" | jq -r '.summary.deployment_frequency // "N/A"')
        lead_time=$(echo "$dora" | jq -r '.summary.lead_time // "N/A"')
    fi

    local total_commits=0
    local contributors=0
    if has_scanner_data "$analysis_path" "git"; then
        local git=$(load_scanner_data "$analysis_path" "git")
        total_commits=$(echo "$git" | jq -r '.summary.total_commits // 0')
        contributors=$(echo "$git" | jq -r '.summary.contributors // 0')
    fi

    local test_coverage=0
    if has_scanner_data "$analysis_path" "test-coverage"; then
        test_coverage=$(load_scanner_data "$analysis_path" "test-coverage" | jq -r '.summary.coverage_percentage // 0')
    fi

    # === OWNERSHIP SECTION ===
    local bus_factor=0
    local has_codeowners=false
    local ownership_coverage=0
    if has_scanner_data "$analysis_path" "code-ownership"; then
        local own=$(load_scanner_data "$analysis_path" "code-ownership")
        bus_factor=$(echo "$own" | jq -r '.summary.bus_factor // 0')
        has_codeowners=$(echo "$own" | jq -r '.summary.has_codeowners // false')
        ownership_coverage=$(echo "$own" | jq -r '.summary.coverage_percentage // 0')
    fi

    # === PROVENANCE SECTION ===
    local signed_packages=0
    local slsa_level="none"
    if has_scanner_data "$analysis_path" "package-provenance"; then
        local prov=$(load_scanner_data "$analysis_path" "package-provenance")
        signed_packages=$(echo "$prov" | jq -r '.summary.signed_packages // 0')
        slsa_level=$(echo "$prov" | jq -r '.summary.slsa_level // "none"')
    fi

    # === CALCULATE SCORES ===
    local risk_level=$(calculate_risk_level "$critical" "$high" "$medium")

    local security_score=100
    security_score=$((security_score - critical * 25))
    security_score=$((security_score - high * 10))
    security_score=$((security_score - secrets * 20))
    [[ $security_score -lt 0 ]] && security_score=0

    local compliance_score=100
    [[ $license_violations -gt 0 ]] && compliance_score=$((compliance_score - license_violations * 15))
    [[ "$license_status" != "pass" ]] && compliance_score=$((compliance_score - 20))
    [[ $compliance_score -lt 0 ]] && compliance_score=0

    local supply_chain_score=100
    supply_chain_score=$((supply_chain_score - abandoned * 5))
    supply_chain_score=$((supply_chain_score - typosquatting * 25))
    [[ $supply_chain_score -lt 0 ]] && supply_chain_score=0

    local overall_score=$(( (security_score + compliance_score + supply_chain_score) / 3 ))

    # Build comprehensive JSON
    jq -n \
        --arg project_id "$project_id" \
        --arg scan_id "$scan_id" \
        --arg profile "$profile" \
        --arg completed_at "$completed_at" \
        --argjson duration "$duration" \
        --arg commit "$commit" \
        --arg branch "$branch" \
        --arg risk_level "$risk_level" \
        --argjson overall_score "$overall_score" \
        --argjson security_score "$security_score" \
        --argjson compliance_score "$compliance_score" \
        --argjson supply_chain_score "$supply_chain_score" \
        --argjson critical "$critical" \
        --argjson high "$high" \
        --argjson medium "$medium" \
        --argjson low "$low" \
        --argjson total_vulns "$total_vulns" \
        --argjson secrets "$secrets" \
        --argjson secrets_by_type "$secrets_by_type" \
        --argjson code_security_count "$code_security_count" \
        --argjson iac_count "$iac_count" \
        --argjson total_deps "$total_deps" \
        --argjson direct_deps "$direct_deps" \
        --argjson abandoned "$abandoned" \
        --argjson deprecated "$deprecated" \
        --argjson typosquatting "$typosquatting" \
        --arg license_status "$license_status" \
        --argjson license_violations "$license_violations" \
        --argjson copyleft_count "$copyleft_count" \
        --arg has_readme "$has_readme" \
        --arg has_license_file "$has_license_file" \
        --argjson doc_score "$doc_score" \
        --arg dora_perf "$dora_perf" \
        --arg deploy_freq "$deploy_freq" \
        --arg lead_time "$lead_time" \
        --argjson total_commits "$total_commits" \
        --argjson contributors "$contributors" \
        --argjson test_coverage "$test_coverage" \
        --argjson bus_factor "$bus_factor" \
        --arg has_codeowners "$has_codeowners" \
        --argjson ownership_coverage "$ownership_coverage" \
        --argjson signed_packages "$signed_packages" \
        --arg slsa_level "$slsa_level" \
        '{
            report_type: "full",
            report_version: "1.0.0",
            generated_at: (now | todate),
            project: {
                id: $project_id,
                scan_id: $scan_id,
                profile: $profile,
                completed_at: $completed_at,
                duration_seconds: $duration,
                git: {
                    commit: $commit,
                    branch: $branch
                }
            },
            scores: {
                overall: $overall_score,
                security: $security_score,
                compliance: $compliance_score,
                supply_chain: $supply_chain_score
            },
            risk: {
                level: $risk_level
            },
            security: {
                vulnerabilities: {
                    critical: $critical,
                    high: $high,
                    medium: $medium,
                    low: $low,
                    total: $total_vulns
                },
                secrets: {
                    total: $secrets,
                    by_type: $secrets_by_type
                },
                code_security_findings: $code_security_count,
                iac_security_findings: $iac_count
            },
            dependencies: {
                total: $total_deps,
                direct: $direct_deps,
                transitive: ($total_deps - $direct_deps),
                health: {
                    abandoned: $abandoned,
                    deprecated: $deprecated,
                    typosquatting_suspects: $typosquatting
                }
            },
            compliance: {
                license_status: $license_status,
                license_violations: $license_violations,
                copyleft_packages: $copyleft_count,
                documentation: {
                    has_readme: ($has_readme == "true"),
                    has_license_file: ($has_license_file == "true"),
                    score: $doc_score
                }
            },
            devops: {
                dora: {
                    performance: $dora_perf,
                    deployment_frequency: $deploy_freq,
                    lead_time: $lead_time
                },
                git: {
                    total_commits: $total_commits,
                    contributors: $contributors
                },
                test_coverage_percentage: $test_coverage
            },
            ownership: {
                bus_factor: $bus_factor,
                has_codeowners: ($has_codeowners == "true"),
                coverage_percentage: $ownership_coverage
            },
            provenance: {
                signed_packages: $signed_packages,
                slsa_level: $slsa_level
            }
        }'
}

# Generate org aggregate full data
generate_org_report_data() {
    local org="$1"
    local projects="$2"

    local project_count=$(echo "$projects" | wc -w | tr -d ' ')
    local total_critical=0
    local total_high=0
    local total_secrets=0
    local total_deps=0
    local total_violations=0

    for repo in $projects; do
        local project_id="$org/$repo"
        local analysis_path="$ZERO_PROJECTS_DIR/$project_id/analysis"

        if [[ -d "$analysis_path" ]]; then
            local vulns=$(aggregate_vulns "$analysis_path")
            total_critical=$((total_critical + $(echo "$vulns" | jq -r '.critical')))
            total_high=$((total_high + $(echo "$vulns" | jq -r '.high')))

            local deps=$(aggregate_deps "$analysis_path")
            total_deps=$((total_deps + $(echo "$deps" | jq -r '.total')))

            if has_scanner_data "$analysis_path" "code-secrets"; then
                total_secrets=$((total_secrets + $(load_scanner_data "$analysis_path" "code-secrets" | jq -r '.summary.total_findings // 0')))
            fi

            if has_scanner_data "$analysis_path" "licenses"; then
                total_violations=$((total_violations + $(load_scanner_data "$analysis_path" "licenses" | jq -r '.summary.license_violations // 0')))
            fi
        fi
    done

    local risk_level=$(calculate_risk_level "$total_critical" "$total_high" "0")

    jq -n \
        --arg org "$org" \
        --argjson project_count "$project_count" \
        --arg risk_level "$risk_level" \
        --argjson total_critical "$total_critical" \
        --argjson total_high "$total_high" \
        --argjson total_secrets "$total_secrets" \
        --argjson total_deps "$total_deps" \
        --argjson total_violations "$total_violations" \
        '{
            report_type: "full",
            report_version: "1.0.0",
            generated_at: (now | todate),
            organization: $org,
            projects: {
                count: $project_count
            },
            risk: {
                level: $risk_level
            },
            security: {
                critical_vulnerabilities: $total_critical,
                high_vulnerabilities: $total_high,
                exposed_secrets: $total_secrets
            },
            dependencies: {
                total: $total_deps
            },
            compliance: {
                license_violations: $total_violations
            }
        }'
}

export -f generate_report_data
export -f generate_org_report_data
