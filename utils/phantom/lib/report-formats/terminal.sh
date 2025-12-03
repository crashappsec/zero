#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Terminal Format Output
# Colored terminal output for reports
#############################################################################

# Box drawing characters
BOX_TL="╔"
BOX_TR="╗"
BOX_BL="╚"
BOX_BR="╝"
BOX_H="═"
BOX_V="║"
BOX_LINE="━"

# Format report output to terminal
# Usage: format_report_output <json_data> <target_id>
format_report_output() {
    local json_data="$1"
    local target_id="$2"

    local report_type=$(echo "$json_data" | jq -r '.report_type')

    case "$report_type" in
        summary)
            format_summary_terminal "$json_data" "$target_id"
            ;;
        security)
            format_security_terminal "$json_data" "$target_id"
            ;;
        licenses)
            format_licenses_terminal "$json_data" "$target_id"
            ;;
        sbom)
            format_sbom_terminal "$json_data" "$target_id"
            ;;
        compliance)
            format_compliance_terminal "$json_data" "$target_id"
            ;;
        supply-chain)
            format_supply_chain_terminal "$json_data" "$target_id"
            ;;
        dora)
            format_dora_terminal "$json_data" "$target_id"
            ;;
        full)
            format_full_terminal "$json_data" "$target_id"
            ;;
        *)
            format_summary_terminal "$json_data" "$target_id"
            ;;
    esac
}

# Format summary report for terminal
format_summary_terminal() {
    local json="$1"
    local target_id="$2"

    # Check if this is an org report
    local is_org=$(echo "$json" | jq -r 'has("organization")')

    if [[ "$is_org" == "true" ]]; then
        format_org_summary_terminal "$json"
    else
        format_project_summary_terminal "$json"
    fi
}

# Format project summary for terminal
format_project_summary_terminal() {
    local json="$1"

    # Extract data
    local project_id=$(echo "$json" | jq -r '.project.id')
    local scan_id=$(echo "$json" | jq -r '.project.scan_id')
    local profile=$(echo "$json" | jq -r '.project.profile')
    local completed_at=$(echo "$json" | jq -r '.project.completed_at')
    local duration=$(echo "$json" | jq -r '.project.duration_seconds')
    local commit=$(echo "$json" | jq -r '.project.git.commit')
    local branch=$(echo "$json" | jq -r '.project.git.branch')

    # Vulnerabilities (moved out of risk object)
    local critical=$(echo "$json" | jq -r '.vulnerabilities.critical // .risk.vulnerabilities.critical // 0')
    local high=$(echo "$json" | jq -r '.vulnerabilities.high // .risk.vulnerabilities.high // 0')
    local medium=$(echo "$json" | jq -r '.vulnerabilities.medium // .risk.vulnerabilities.medium // 0')
    local low=$(echo "$json" | jq -r '.vulnerabilities.low // .risk.vulnerabilities.low // 0')

    local total_deps=$(echo "$json" | jq -r '.dependencies.total')
    local direct_deps=$(echo "$json" | jq -r '.dependencies.direct')
    local abandoned=$(echo "$json" | jq -r '.dependencies.abandoned')

    local secrets=$(echo "$json" | jq -r '.secrets.exposed')
    local license_status=$(echo "$json" | jq -r '.licenses.status')
    local license_violations=$(echo "$json" | jq -r '.licenses.violations')
    local dora_perf=$(echo "$json" | jq -r '.dora.performance')

    # Repository info
    local repo_size=$(echo "$json" | jq -r '.repository.size // "unknown"')
    local repo_files=$(echo "$json" | jq -r '.repository.files // 0')
    local languages=$(echo "$json" | jq -r '.repository.languages // [] | join(", ")')

    # Print header box
    echo
    printf "${BOLD}${BOX_TL}"
    printf '%*s' 66 '' | tr ' ' "$BOX_H"
    printf "${BOX_TR}${NC}\n"
    printf "${BOLD}${BOX_V}  PHANTOM SUMMARY REPORT%*s${BOX_V}${NC}\n" 42 ''
    printf "${BOLD}${BOX_BL}"
    printf '%*s' 66 '' | tr ' ' "$BOX_H"
    printf "${BOX_BR}${NC}\n"
    echo

    # Project info
    printf "  ${BOLD}Project:${NC}     %s\n" "$project_id"
    printf "  ${BOLD}Scanned:${NC}     %s ${DIM}(%s)${NC}\n" "$(format_timestamp "$completed_at")" "$(relative_time "$completed_at")"
    printf "  ${BOLD}Profile:${NC}     %s\n" "$profile"
    if [[ -n "$commit" ]] && [[ "$commit" != "null" ]] && [[ "$commit" != "" ]]; then
        printf "  ${BOLD}Commit:${NC}      %s ${DIM}(%s)${NC}\n" "$commit" "$branch"
    fi

    echo
    hr "$BOX_LINE" 68
    echo

    # Repository section
    printf "  ${BOLD}REPOSITORY${NC}\n"
    printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

    printf "  %-16s │  %s\n" "Size" "$repo_size"
    printf "  %-16s │  %s\n" "Files" "$(format_number "$repo_files")"
    if [[ -n "$languages" ]] && [[ "$languages" != "" ]]; then
        printf "  %-16s │  %s\n" "Languages" "$languages"
    fi

    echo
    hr "$BOX_LINE" 68
    echo

    # Packages section with breakdown
    printf "  ${BOLD}PACKAGES${NC}\n"
    printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

    # Show package breakdown by ecosystem
    local packages_count=$(echo "$json" | jq -r '.dependencies.packages | length // 0')
    if [[ "$packages_count" -gt 0 ]]; then
        echo "$json" | jq -r '.dependencies.packages[] | "\(.ecosystem)|\(.count)|\(.sources | join(", "))"' 2>/dev/null | while IFS='|' read -r ecosystem count sources; do
            # Map ecosystem to display name
            local eco_display
            case "$ecosystem" in
                npm) eco_display="NPM" ;;
                python) eco_display="Python" ;;
                github-action) eco_display="GitHub Actions" ;;
                github-action-workflow) eco_display="GH Workflows" ;;
                binary) eco_display="Binary" ;;
                go) eco_display="Go" ;;
                rust) eco_display="Rust" ;;
                java) eco_display="Java" ;;
                maven) eco_display="Maven" ;;
                gradle) eco_display="Gradle" ;;
                nuget) eco_display="NuGet" ;;
                composer) eco_display="Composer" ;;
                gem) eco_display="Ruby Gems" ;;
                *) eco_display="$(echo "${ecosystem:0:1}" | tr '[:lower:]' '[:upper:]')${ecosystem:1}" ;;
            esac
            # Clean up source paths (remove leading / only)
            local clean_sources=$(echo "$sources" | sed 's|^/||' | sed 's|, /|, |g')
            printf "  %-16s │  %s packages ${DIM}from %s${NC}\n" "$eco_display" "$count" "$clean_sources"
        done
    else
        printf "  %-16s │  %s total" "Dependencies" "$total_deps"
        if [[ "$direct_deps" != "0" ]]; then
            printf " ${DIM}(%s direct)${NC}" "$direct_deps"
        fi
        echo
    fi

    echo
    hr "$BOX_LINE" 68
    echo

    # Security section
    printf "  ${BOLD}SECURITY${NC}\n"
    printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

    # Vulnerabilities
    printf "  %-16s │  " "Vulnerabilities"
    if [[ "$critical" -gt 0 ]]; then
        printf "${RED}%s critical${NC}  " "$critical"
    fi
    if [[ "$high" -gt 0 ]]; then
        printf "${YELLOW}%s high${NC}  " "$high"
    fi
    if [[ "$medium" -gt 0 ]]; then
        printf "%s medium  " "$medium"
    fi
    if [[ "$low" -gt 0 ]]; then
        printf "${DIM}%s low${NC}" "$low"
    fi
    if [[ "$critical" -eq 0 ]] && [[ "$high" -eq 0 ]] && [[ "$medium" -eq 0 ]] && [[ "$low" -eq 0 ]]; then
        printf "${GREEN}None${NC}"
    fi
    echo

    # Secrets
    printf "  %-16s │  " "Secrets"
    if [[ "$secrets" -gt 0 ]]; then
        printf "${RED}%s exposed${NC}\n" "$secrets"
    else
        printf "${GREEN}0 exposed${NC}\n"
    fi

    # Licenses
    printf "  %-16s │  " "Licenses"
    if [[ "$license_violations" -gt 0 ]]; then
        printf "${YELLOW}%s violations${NC}\n" "$license_violations"
    elif [[ "$license_status" == "pass" ]]; then
        printf "${GREEN}Compliant${NC}\n"
    else
        printf "%s\n" "$license_status"
    fi

    # Abandoned packages
    if [[ "$abandoned" -gt 0 ]]; then
        printf "  %-16s │  ${YELLOW}%s packages${NC}\n" "Abandoned" "$abandoned"
    fi

    # DORA
    if [[ "$dora_perf" != "N/A" ]] && [[ "$dora_perf" != "null" ]]; then
        echo
        hr "$BOX_LINE" 68
        echo
        printf "  ${BOLD}DORA METRICS${NC}\n"
        printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"
        printf "  %-16s │  " "Performance"
        case "$dora_perf" in
            ELITE) printf "${GREEN}%s${NC}\n" "$dora_perf" ;;
            HIGH) printf "${GREEN}%s${NC}\n" "$dora_perf" ;;
            MEDIUM) printf "${YELLOW}%s${NC}\n" "$dora_perf" ;;
            LOW) printf "${RED}%s${NC}\n" "$dora_perf" ;;
            *) printf "%s\n" "$dora_perf" ;;
        esac
    fi

    # Top issues section
    local top_issues=$(echo "$json" | jq -r '.top_issues[]' 2>/dev/null)
    if [[ -n "$top_issues" ]]; then
        echo
        hr "$BOX_LINE" 68
        echo
        printf "  ${BOLD}TOP ISSUES${NC}\n"
        printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

        local issue_num=1
        echo "$top_issues" | while read -r issue; do
            printf "  %d. %s\n" "$issue_num" "$issue"
            ((issue_num++))
        done
    fi

    # Scan sources section
    local sources_count=$(echo "$json" | jq -r '.scan_sources | length // 0')
    if [[ "$sources_count" -gt 0 ]]; then
        echo
        hr "$BOX_LINE" 68
        echo
        printf "  ${BOLD}DATA SOURCES${NC} ${DIM}(%s scanners)${NC}\n" "$sources_count"
        printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"
        local scanner_list=$(echo "$json" | jq -r '[.scan_sources[] | .scanner] | join(", ")' 2>/dev/null)
        printf "  ${DIM}%s${NC}\n" "$scanner_list"
    fi

    echo
    hr "$BOX_LINE" 68
    echo -e "  ${DIM}Generated by Phantom Report v${REPORT_VERSION}${NC}"
    echo
}

# Format org summary for terminal
format_org_summary_terminal() {
    local json="$1"

    local org=$(echo "$json" | jq -r '.organization')
    local project_count=$(echo "$json" | jq -r '.projects.count')
    local total_vulns=$(echo "$json" | jq -r '.risk.vulnerabilities.total')
    local critical=$(echo "$json" | jq -r '.risk.vulnerabilities.critical')
    local high=$(echo "$json" | jq -r '.risk.vulnerabilities.high')
    local total_deps=$(echo "$json" | jq -r '.dependencies.total')
    local at_risk=$(echo "$json" | jq -r '.projects.at_risk | join(", ")')

    # Header
    echo
    printf "${BOLD}${BOX_TL}"
    printf '%*s' 66 '' | tr ' ' "$BOX_H"
    printf "${BOX_TR}${NC}\n"
    printf "${BOLD}${BOX_V}  ORGANIZATION SUMMARY: %-42s${BOX_V}${NC}\n" "$org"
    printf "${BOLD}${BOX_BL}"
    printf '%*s' 66 '' | tr ' ' "$BOX_H"
    printf "${BOX_BR}${NC}\n"
    echo

    printf "  ${BOLD}Projects:${NC}    %s\n" "$project_count"

    echo
    hr "$BOX_LINE" 68
    echo

    # Aggregate metrics
    printf "  ${BOLD}AGGREGATE METRICS${NC}\n"
    printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

    printf "  %-16s │  %s\n" "Dependencies" "$(format_number "$total_deps")"

    # Vulnerabilities
    printf "  %-16s │  " "Vulnerabilities"
    if [[ "$critical" -gt 0 ]]; then
        printf "${RED}%s critical${NC}  " "$critical"
    fi
    if [[ "$high" -gt 0 ]]; then
        printf "${YELLOW}%s high${NC}  " "$high"
    fi
    local other=$((total_vulns - critical - high))
    if [[ "$other" -gt 0 ]]; then
        printf "${DIM}%s other${NC}" "$other"
    fi
    if [[ "$total_vulns" -eq 0 ]]; then
        printf "${GREEN}None${NC}"
    fi
    echo

    # At-risk repos
    if [[ -n "$at_risk" ]] && [[ "$at_risk" != "" ]]; then
        echo
        printf "  ${BOLD}REPOS WITH ISSUES${NC}\n"
        printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"
        printf "  ${YELLOW}%s${NC}\n" "$at_risk"
    fi

    echo
    hr "$BOX_LINE" 68
    echo -e "  ${DIM}Generated by Phantom Report v${REPORT_VERSION}${NC}"
    echo
}

# Format security report for terminal
format_security_terminal() {
    local json="$1"
    local target_id="$2"

    # Extract project data
    local project_id=$(echo "$json" | jq -r '.project.id')
    local profile=$(echo "$json" | jq -r '.project.profile')
    local completed_at=$(echo "$json" | jq -r '.project.completed_at')

    # Vulnerability summary
    local critical=$(echo "$json" | jq -r '.vulnerabilities.summary.critical // 0')
    local high=$(echo "$json" | jq -r '.vulnerabilities.summary.high // 0')
    local medium=$(echo "$json" | jq -r '.vulnerabilities.summary.medium // 0')
    local low=$(echo "$json" | jq -r '.vulnerabilities.summary.low // 0')
    local total_vulns=$(echo "$json" | jq -r '.vulnerabilities.summary.total // 0')

    # Secrets summary
    local secrets_count=$(echo "$json" | jq -r '.secrets.summary.total // 0')

    # Code security summary
    local code_sec_total=$(echo "$json" | jq -r '.code_security.summary.total // 0')

    # IaC security summary
    local iac_total=$(echo "$json" | jq -r '.iac_security.summary.total // 0')

    # Print header box
    echo
    printf "${BOLD}${BOX_TL}"
    printf '%*s' 66 '' | tr ' ' "$BOX_H"
    printf "${BOX_TR}${NC}\n"
    printf "${BOLD}${BOX_V}  PHANTOM SECURITY REPORT%*s${BOX_V}${NC}\n" 41 ''
    printf "${BOLD}${BOX_BL}"
    printf '%*s' 66 '' | tr ' ' "$BOX_H"
    printf "${BOX_BR}${NC}\n"
    echo

    # Project info
    printf "  ${BOLD}Project:${NC}     %s\n" "$project_id"
    printf "  ${BOLD}Scanned:${NC}     %s ${DIM}(%s)${NC}\n" "$(format_timestamp "$completed_at")" "$(relative_time "$completed_at")"
    printf "  ${BOLD}Profile:${NC}     %s\n" "$profile"

    echo
    hr "$BOX_LINE" 68
    echo

    # Security overview
    printf "  ${BOLD}SECURITY OVERVIEW${NC}\n"
    printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

    printf "  %-18s │  " "Vulnerabilities"
    if [[ "$total_vulns" -eq 0 ]]; then
        printf "${GREEN}None${NC}\n"
    else
        printf "%s total" "$total_vulns"
        if [[ "$critical" -gt 0 ]]; then printf " (${RED}%s critical${NC})" "$critical"; fi
        if [[ "$high" -gt 0 ]]; then printf " (${YELLOW}%s high${NC})" "$high"; fi
        echo
    fi

    printf "  %-18s │  " "Secrets"
    if [[ "$secrets_count" -eq 0 ]]; then
        printf "${GREEN}0 exposed${NC}\n"
    else
        printf "${RED}%s exposed${NC}\n" "$secrets_count"
    fi

    printf "  %-18s │  %s findings\n" "Code Security" "$code_sec_total"
    printf "  %-18s │  %s findings\n" "IaC Security" "$iac_total"

    # Vulnerability details
    if [[ "$total_vulns" -gt 0 ]]; then
        echo
        hr "$BOX_LINE" 68
        echo
        printf "  ${BOLD}VULNERABILITIES${NC} ${DIM}(%s total)${NC}\n" "$total_vulns"
        printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

        # Show breakdown by severity
        if [[ "$critical" -gt 0 ]]; then
            printf "  ${RED}Critical:${NC} %s\n" "$critical"
        fi
        if [[ "$high" -gt 0 ]]; then
            printf "  ${YELLOW}High:${NC}     %s\n" "$high"
        fi
        if [[ "$medium" -gt 0 ]]; then
            printf "  Medium:   %s\n" "$medium"
        fi
        if [[ "$low" -gt 0 ]]; then
            printf "  ${DIM}Low:      %s${NC}\n" "$low"
        fi

        echo
        printf "  ${BOLD}Details:${NC}\n"
        printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

        # Show vulnerability details
        local vuln_count=$(echo "$json" | jq -r '.vulnerabilities.details | length')
        if [[ "$vuln_count" -gt 0 ]]; then
            echo "$json" | jq -r '.vulnerabilities.details[] | @json' 2>/dev/null | while read -r vuln_json; do
                local severity=$(echo "$vuln_json" | jq -r '.severity // "unknown"')
                local pkg=$(echo "$vuln_json" | jq -r '.package // "unknown"')
                local version=$(echo "$vuln_json" | jq -r '.version // "unknown"')
                local id=$(echo "$vuln_json" | jq -r '.id // "unknown"')
                local summary=$(echo "$vuln_json" | jq -r '.summary // "No description"')
                local ecosystem=$(echo "$vuln_json" | jq -r '.ecosystem // "unknown"')
                local aliases=$(echo "$vuln_json" | jq -r '.aliases // [] | join(", ")' 2>/dev/null)
                local fix_available=$(echo "$vuln_json" | jq -r '.fix_available // "unknown"')
                local fixed_version=$(echo "$vuln_json" | jq -r '.fixed_version // null')
                local osv_url=$(echo "$vuln_json" | jq -r '.osv_url // ("https://osv.dev/vulnerability/" + .id)')
                local references=$(echo "$vuln_json" | jq -r '.references // []')

                # Color by severity
                local sev_color=""
                case "$severity" in
                    critical) sev_color="${RED}" ;;
                    high) sev_color="${YELLOW}" ;;
                    medium) sev_color="" ;;
                    low) sev_color="${DIM}" ;;
                esac

                # Print vulnerability details
                printf "\n  ${sev_color}[%s]${NC} %s" "$severity" "$id"
                if [[ -n "$aliases" ]] && [[ "$aliases" != "" ]]; then
                    printf " ${DIM}(%s)${NC}" "$aliases"
                fi
                printf "\n"
                printf "  Package: %s@%s (%s)\n" "$pkg" "$version" "$ecosystem"
                printf "  %s\n" "$summary"

                # Show fix info if available
                if [[ "$fix_available" == "yes" ]] && [[ "$fixed_version" != "null" ]] && [[ -n "$fixed_version" ]]; then
                    printf "  ${GREEN}Fix: Upgrade to %s${NC}\n" "$fixed_version"
                elif [[ "$fix_available" == "no" ]]; then
                    printf "  ${YELLOW}No fix available${NC}\n"
                fi

                # Show key references
                local advisory_url=$(echo "$references" | jq -r '.[] | select(.type == "ADVISORY") | .url' 2>/dev/null | head -1)
                if [[ -n "$advisory_url" ]] && [[ "$advisory_url" != "" ]]; then
                    printf "  ${CYAN}Advisory: %s${NC}\n" "$advisory_url"
                fi
                printf "  ${CYAN}Details: %s${NC}\n" "$osv_url"
            done
        fi
    fi

    # Secrets details
    if [[ "$secrets_count" -gt 0 ]]; then
        echo
        hr "$BOX_LINE" 68
        echo
        printf "  ${BOLD}EXPOSED SECRETS${NC} ${DIM}(%s total)${NC}\n" "$secrets_count"
        printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

        # Show by type
        local types=$(echo "$json" | jq -r '.secrets.summary.by_type | to_entries[] | "\(.key): \(.value)"' 2>/dev/null)
        if [[ -n "$types" ]]; then
            echo "$types" | while read -r type_info; do
                printf "  ${RED}%s${NC}\n" "$type_info"
            done
        fi

        # Show secret details
        local secrets_detail_count=$(echo "$json" | jq -r '.secrets.details | length')
        if [[ "$secrets_detail_count" -gt 0 ]]; then
            echo
            echo "$json" | jq -r '.secrets.details[] | "\(.type // "unknown")|\(.file // "unknown")|\(.line // 0)"' 2>/dev/null | while IFS='|' read -r type file line; do
                printf "  ${RED}%s${NC} in %s:%s\n" "$type" "$file" "$line"
            done
        fi
    fi

    echo
    hr "$BOX_LINE" 68
    echo -e "  ${DIM}Generated by Phantom Report v${REPORT_VERSION}${NC}"
    echo
}

# Format license report for terminal
format_licenses_terminal() {
    local json="$1"
    local target_id="$2"

    # Extract project data
    local project_id=$(echo "$json" | jq -r '.project.id')
    local profile=$(echo "$json" | jq -r '.project.profile')
    local completed_at=$(echo "$json" | jq -r '.project.completed_at')
    local overall_status=$(echo "$json" | jq -r '.overall_status // "unknown"')

    # Repository license
    local repo_license=$(echo "$json" | jq -r '.repository_license.license // "Not Found"')
    local repo_license_file=$(echo "$json" | jq -r '.repository_license.file // null')

    # Summary
    local proj_violations=$(echo "$json" | jq -r '.summary.project_license_violations // 0')
    local dep_violations=$(echo "$json" | jq -r '.summary.dependency_license_violations // 0')
    local total_deps=$(echo "$json" | jq -r '.summary.total_dependencies_scanned // 0')
    local denied_count=$(echo "$json" | jq -r '.summary.denied_license_packages // 0')
    local review_count=$(echo "$json" | jq -r '.summary.review_required_packages // 0')

    # Print header box
    echo
    printf "${BOLD}${BOX_TL}"
    printf '%*s' 66 '' | tr ' ' "$BOX_H"
    printf "${BOX_TR}${NC}\n"
    printf "${BOLD}${BOX_V}  PHANTOM LICENSE REPORT%*s${BOX_V}${NC}\n" 42 ''
    printf "${BOLD}${BOX_BL}"
    printf '%*s' 66 '' | tr ' ' "$BOX_H"
    printf "${BOX_BR}${NC}\n"
    echo

    # Project info
    printf "  ${BOLD}Project:${NC}     %s\n" "$project_id"
    printf "  ${BOLD}Scanned:${NC}     %s ${DIM}(%s)${NC}\n" "$(format_timestamp "$completed_at")" "$(relative_time "$completed_at")"
    printf "  ${BOLD}Profile:${NC}     %s\n" "$profile"

    # Overall status
    printf "  ${BOLD}Status:${NC}      "
    case "$overall_status" in
        pass) printf "${GREEN}PASS${NC}\n" ;;
        fail) printf "${RED}FAIL${NC}\n" ;;
        warning) printf "${YELLOW}WARNING${NC}\n" ;;
        *) printf "%s\n" "$overall_status" ;;
    esac

    echo
    hr "$BOX_LINE" 68
    echo

    # Repository license section
    printf "  ${BOLD}REPOSITORY LICENSE${NC}\n"
    printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

    printf "  %-16s │  " "License"
    if [[ "$repo_license" == "Not Found" ]] || [[ "$repo_license" == "null" ]] || [[ -z "$repo_license" ]]; then
        printf "${YELLOW}Not Found${NC}\n"
    else
        printf "%s\n" "$repo_license"
    fi

    if [[ -n "$repo_license_file" ]] && [[ "$repo_license_file" != "null" ]]; then
        printf "  %-16s │  %s\n" "Source" "$repo_license_file"
    fi

    echo
    hr "$BOX_LINE" 68
    echo

    # License overview
    printf "  ${BOLD}LICENSE OVERVIEW${NC}\n"
    printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

    printf "  %-20s │  %s\n" "Dependencies" "$total_deps"
    printf "  %-20s │  " "Denied Licenses"
    if [[ "$denied_count" -gt 0 ]]; then
        printf "${RED}%s packages${NC}\n" "$denied_count"
    else
        printf "${GREEN}None${NC}\n"
    fi
    printf "  %-20s │  " "Needs Review"
    if [[ "$review_count" -gt 0 ]]; then
        printf "${YELLOW}%s packages${NC}\n" "$review_count"
    else
        printf "${GREEN}None${NC}\n"
    fi

    # Denied licenses (GPL, AGPL)
    local denied_list=$(echo "$json" | jq -r '.dependency_licenses.denied // []')
    local denied_len=$(echo "$denied_list" | jq -r 'length')
    if [[ "$denied_len" -gt 0 ]]; then
        echo
        hr "$BOX_LINE" 68
        echo
        printf "  ${BOLD}${RED}DENIED LICENSES${NC}${NC} ${DIM}(GPL, AGPL - require removal/replacement)${NC}\n"
        printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

        echo "$denied_list" | jq -r '.[] | @json' 2>/dev/null | while read -r item; do
            local license=$(echo "$item" | jq -r '.license')
            local count=$(echo "$item" | jq -r '.count')
            local packages=$(echo "$item" | jq -r '[.packages[]? | "\(.name)@\(.version)"] | join(", ")' 2>/dev/null)

            printf "\n  ${RED}%s${NC} ${DIM}(%s packages)${NC}\n" "$license" "$count"
            # Wrap long package lists
            if [[ ${#packages} -gt 60 ]]; then
                echo "$item" | jq -r '.packages[]? | "  • \(.name)@\(.version)"' 2>/dev/null
            else
                printf "  %s\n" "$packages"
            fi
        done
    fi

    # Review required licenses (LGPL, MPL, EPL)
    local review_list=$(echo "$json" | jq -r '.dependency_licenses.review_required // []')
    local review_len=$(echo "$review_list" | jq -r 'length')
    if [[ "$review_len" -gt 0 ]]; then
        echo
        hr "$BOX_LINE" 68
        echo
        printf "  ${BOLD}${YELLOW}REVIEW REQUIRED${NC}${NC} ${DIM}(LGPL, MPL, EPL - may have restrictions)${NC}\n"
        printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

        echo "$review_list" | jq -r '.[] | @json' 2>/dev/null | while read -r item; do
            local license=$(echo "$item" | jq -r '.license')
            local count=$(echo "$item" | jq -r '.count')
            local packages=$(echo "$item" | jq -r '[.packages[]? | "\(.name)@\(.version)"] | join(", ")' 2>/dev/null)

            printf "\n  ${YELLOW}%s${NC} ${DIM}(%s packages)${NC}\n" "$license" "$count"
            if [[ ${#packages} -gt 60 ]]; then
                echo "$item" | jq -r '.packages[]? | "  • \(.name)@\(.version)"' 2>/dev/null
            else
                printf "  %s\n" "$packages"
            fi
        done
    fi

    # All licenses breakdown
    local all_licenses=$(echo "$json" | jq -r '.dependency_licenses.by_license // {}')
    local license_count=$(echo "$all_licenses" | jq -r 'keys | length')
    if [[ "$license_count" -gt 0 ]]; then
        echo
        hr "$BOX_LINE" 68
        echo
        printf "  ${BOLD}ALL DEPENDENCY LICENSES${NC} ${DIM}(%s license types)${NC}\n" "$license_count"
        printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

        # Sort by count descending and show top 15
        echo "$all_licenses" | jq -r 'to_entries | sort_by(-.value.count) | .[:15][] | "\(.key)|\(.value.count)"' 2>/dev/null | while IFS='|' read -r license count; do
            [[ -z "$license" ]] && continue

            # Color based on license type
            local lic_color=""
            if echo "$license" | grep -qiE "GPL|AGPL"; then
                lic_color="${RED}"
            elif echo "$license" | grep -qiE "LGPL|MPL|EPL"; then
                lic_color="${YELLOW}"
            else
                lic_color="${GREEN}"
            fi

            printf "  ${lic_color}%-20s${NC} │  %s packages\n" "$license" "$count"
        done

        if [[ "$license_count" -gt 15 ]]; then
            printf "  ${DIM}... and %s more license types${NC}\n" "$((license_count - 15))"
        fi
    fi

    # Content policy section (if any issues)
    local profanity_count=$(echo "$json" | jq -r '.content_policy.profanity_issues // 0')
    local inclusive_count=$(echo "$json" | jq -r '.content_policy.inclusive_language_issues // 0')
    if [[ "$profanity_count" -gt 0 ]] || [[ "$inclusive_count" -gt 0 ]]; then
        echo
        hr "$BOX_LINE" 68
        echo
        printf "  ${BOLD}CONTENT POLICY${NC}\n"
        printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

        if [[ "$profanity_count" -gt 0 ]]; then
            printf "  %-20s │  ${YELLOW}%s issues${NC}\n" "Profanity" "$profanity_count"
        fi
        if [[ "$inclusive_count" -gt 0 ]]; then
            printf "  %-20s │  ${YELLOW}%s issues${NC}\n" "Non-inclusive terms" "$inclusive_count"
        fi
    fi

    # Policy reference
    echo
    hr "$BOX_LINE" 68
    echo
    printf "  ${BOLD}LICENSE POLICY${NC}\n"
    printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

    local allowed=$(echo "$json" | jq -r '.policy.allowed_licenses // [] | join(", ")' 2>/dev/null)
    local denied=$(echo "$json" | jq -r '.policy.denied_licenses // [] | join(", ")' 2>/dev/null)
    local review=$(echo "$json" | jq -r '.policy.review_required // [] | join(", ")' 2>/dev/null)

    printf "  ${GREEN}Allowed:${NC}  %s\n" "$allowed"
    printf "  ${RED}Denied:${NC}   %s\n" "$denied"
    printf "  ${YELLOW}Review:${NC}   %s\n" "$review"

    echo
    hr "$BOX_LINE" 68
    echo -e "  ${DIM}Generated by Phantom Report v${REPORT_VERSION}${NC}"
    echo
}

# Format SBOM report for terminal
format_sbom_terminal() {
    local json="$1"
    local target_id="$2"

    local project_id=$(echo "$json" | jq -r '.project.id // .organization // "Unknown"')
    local profile=$(echo "$json" | jq -r '.project.profile // "standard"')
    local completed_at=$(echo "$json" | jq -r '.project.completed_at // ""')
    local commit=$(echo "$json" | jq -r '.project.git.commit // ""')
    local branch=$(echo "$json" | jq -r '.project.git.branch // ""')

    local total=$(echo "$json" | jq -r '.summary.total_components // 0')
    local direct=$(echo "$json" | jq -r '.summary.direct_dependencies // 0')
    local transitive=$(echo "$json" | jq -r '.summary.transitive_dependencies // 0')
    local signed=$(echo "$json" | jq -r '.summary.signed_components // 0')
    local has_cyclonedx=$(echo "$json" | jq -r '.formats_available.cyclonedx // false')

    echo
    printf "${BOLD}${BOX_TL}"
    printf '%*s' 66 '' | tr ' ' "$BOX_H"
    printf "${BOX_TR}${NC}\n"
    printf "${BOLD}${BOX_V}  PHANTOM SBOM REPORT%*s${BOX_V}${NC}\n" 45 ''
    printf "${BOLD}${BOX_BL}"
    printf '%*s' 66 '' | tr ' ' "$BOX_H"
    printf "${BOX_BR}${NC}\n"
    echo

    printf "  ${BOLD}Project:${NC}     %s\n" "$project_id"
    printf "  ${BOLD}Scanned:${NC}     %s ${DIM}(%s)${NC}\n" "$(format_timestamp "$completed_at")" "$(relative_time "$completed_at")"
    printf "  ${BOLD}Profile:${NC}     %s\n" "$profile"
    if [[ -n "$commit" ]] && [[ "$commit" != "null" ]]; then
        printf "  ${BOLD}Commit:${NC}      %s ${DIM}(%s)${NC}\n" "$commit" "$branch"
    fi

    echo
    hr "$BOX_LINE" 68
    echo

    printf "  ${BOLD}SUMMARY${NC}\n"
    printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

    printf "  %-20s │  %s\n" "Total Components" "$(format_number "$total")"
    printf "  %-20s │  %s\n" "Direct Dependencies" "$(format_number "$direct")"
    printf "  %-20s │  %s\n" "Transitive" "$(format_number "$transitive")"
    if [[ "$signed" -gt 0 ]]; then
        printf "  %-20s │  %s\n" "Signed Components" "$(format_number "$signed")"
    fi
    printf "  %-20s │  %s\n" "CycloneDX Available" "$([ "$has_cyclonedx" == "true" ] && echo "Yes" || echo "No")"

    # Ecosystem breakdown
    local ecosystems=$(echo "$json" | jq -r '.summary.ecosystems // {}')
    local eco_count=$(echo "$ecosystems" | jq -r 'keys | length')
    if [[ "$eco_count" -gt 0 ]]; then
        echo
        hr "$BOX_LINE" 68
        echo
        printf "  ${BOLD}ECOSYSTEMS${NC}\n"
        printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

        echo "$ecosystems" | jq -r 'to_entries | sort_by(-.value) | .[] | "\(.key)|\(.value)"' 2>/dev/null | while IFS='|' read -r eco count; do
            [[ -z "$eco" ]] && continue
            printf "  %-20s │  %s packages\n" "$eco" "$count"
        done
    fi

    # Component list (first 20)
    local comp_count=$(echo "$json" | jq -r '.components | length // 0')
    if [[ "$comp_count" -gt 0 ]]; then
        echo
        hr "$BOX_LINE" 68
        echo
        printf "  ${BOLD}COMPONENTS${NC} ${DIM}(showing first 20 of %s)${NC}\n" "$comp_count"
        printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

        echo "$json" | jq -r '.components[0:20][] | "\(.name)|\(.version)|\(.ecosystem)"' 2>/dev/null | while IFS='|' read -r name version eco; do
            printf "  %-30s │  %-12s │  %s\n" "$name" "$version" "$eco"
        done
    fi

    echo
    hr "$BOX_LINE" 68
    echo -e "  ${DIM}Generated by Phantom Report v${REPORT_VERSION}${NC}"
    echo
}

# Format compliance report for terminal
format_compliance_terminal() {
    local json="$1"
    local target_id="$2"

    local project_id=$(echo "$json" | jq -r '.project.id // .organization // "Unknown"')
    local profile=$(echo "$json" | jq -r '.project.profile // "standard"')
    local completed_at=$(echo "$json" | jq -r '.project.completed_at // ""')
    local overall_status=$(echo "$json" | jq -r '.overall_status // "unknown"')
    local compliance_score=$(echo "$json" | jq -r '.compliance_score // 0')

    local license_status=$(echo "$json" | jq -r '.licenses.status // "unknown"')
    local license_violations=$(echo "$json" | jq -r '.licenses.violations // 0')
    local copyleft=$(echo "$json" | jq -r '.licenses.copyleft_packages // 0')
    local unknown_lic=$(echo "$json" | jq -r '.licenses.unknown_licenses // 0')

    local total_deps=$(echo "$json" | jq -r '.sbom.total_dependencies // 0')
    local has_readme=$(echo "$json" | jq -r '.documentation.has_readme // false')
    local has_license=$(echo "$json" | jq -r '.documentation.has_license_file // false')
    local has_codeowners=$(echo "$json" | jq -r '.ownership.has_codeowners // false')

    echo
    printf "${BOLD}${BOX_TL}"
    printf '%*s' 66 '' | tr ' ' "$BOX_H"
    printf "${BOX_TR}${NC}\n"
    printf "${BOLD}${BOX_V}  PHANTOM COMPLIANCE REPORT%*s${BOX_V}${NC}\n" 39 ''
    printf "${BOLD}${BOX_BL}"
    printf '%*s' 66 '' | tr ' ' "$BOX_H"
    printf "${BOX_BR}${NC}\n"
    echo

    printf "  ${BOLD}Project:${NC}     %s\n" "$project_id"
    printf "  ${BOLD}Scanned:${NC}     %s ${DIM}(%s)${NC}\n" "$(format_timestamp "$completed_at")" "$(relative_time "$completed_at")"
    printf "  ${BOLD}Status:${NC}      "
    case "$overall_status" in
        PASS) printf "${GREEN}%s${NC}\n" "$overall_status" ;;
        FAIL) printf "${RED}%s${NC}\n" "$overall_status" ;;
        WARN) printf "${YELLOW}%s${NC}\n" "$overall_status" ;;
        *) printf "%s\n" "$overall_status" ;;
    esac

    echo
    hr "$BOX_LINE" 68
    echo

    printf "  ${BOLD}LICENSE COMPLIANCE${NC}\n"
    printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

    printf "  %-20s │  " "Status"
    case "$license_status" in
        pass) printf "${GREEN}Pass${NC}\n" ;;
        fail) printf "${RED}Fail${NC}\n" ;;
        *) printf "%s\n" "$license_status" ;;
    esac
    printf "  %-20s │  " "Violations"
    if [[ "$license_violations" -gt 0 ]]; then
        printf "${RED}%s${NC}\n" "$license_violations"
    else
        printf "${GREEN}0${NC}\n"
    fi
    printf "  %-20s │  %s\n" "Copyleft Packages" "$copyleft"
    printf "  %-20s │  %s\n" "Unknown Licenses" "$unknown_lic"

    echo
    hr "$BOX_LINE" 68
    echo

    printf "  ${BOLD}DOCUMENTATION${NC}\n"
    printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

    printf "  %-20s │  " "README"
    if [[ "$has_readme" == "true" ]]; then printf "${GREEN}Present${NC}\n"; else printf "${YELLOW}Missing${NC}\n"; fi
    printf "  %-20s │  " "LICENSE File"
    if [[ "$has_license" == "true" ]]; then printf "${GREEN}Present${NC}\n"; else printf "${RED}Missing${NC}\n"; fi
    printf "  %-20s │  " "CODEOWNERS"
    if [[ "$has_codeowners" == "true" ]]; then printf "${GREEN}Present${NC}\n"; else printf "${DIM}Not present${NC}\n"; fi

    echo
    hr "$BOX_LINE" 68
    echo

    printf "  ${BOLD}SBOM${NC}\n"
    printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"
    printf "  %-20s │  %s\n" "Total Dependencies" "$(format_number "$total_deps")"

    echo
    hr "$BOX_LINE" 68
    echo -e "  ${DIM}Generated by Phantom Report v${REPORT_VERSION}${NC}"
    echo
}

# Format supply-chain report for terminal
format_supply_chain_terminal() {
    local json="$1"
    local target_id="$2"

    local project_id=$(echo "$json" | jq -r '.project.id // .organization // "Unknown"')
    local profile=$(echo "$json" | jq -r '.project.profile // "standard"')
    local completed_at=$(echo "$json" | jq -r '.project.completed_at // ""')

    local total=$(echo "$json" | jq -r '.dependencies.total // 0')
    local direct=$(echo "$json" | jq -r '.dependencies.direct // 0')
    local abandoned=$(echo "$json" | jq -r '.health.abandoned // 0')
    local deprecated=$(echo "$json" | jq -r '.health.deprecated // 0')
    local outdated=$(echo "$json" | jq -r '.health.outdated // 0')
    local typosquat=$(echo "$json" | jq -r '.health.typosquatting_suspects // 0')

    echo
    printf "${BOLD}${BOX_TL}"
    printf '%*s' 66 '' | tr ' ' "$BOX_H"
    printf "${BOX_TR}${NC}\n"
    printf "${BOLD}${BOX_V}  PHANTOM SUPPLY CHAIN REPORT%*s${BOX_V}${NC}\n" 37 ''
    printf "${BOLD}${BOX_BL}"
    printf '%*s' 66 '' | tr ' ' "$BOX_H"
    printf "${BOX_BR}${NC}\n"
    echo

    printf "  ${BOLD}Project:${NC}     %s\n" "$project_id"
    printf "  ${BOLD}Scanned:${NC}     %s ${DIM}(%s)${NC}\n" "$(format_timestamp "$completed_at")" "$(relative_time "$completed_at")"
    printf "  ${BOLD}Profile:${NC}     %s\n" "$profile"

    echo
    hr "$BOX_LINE" 68
    echo

    printf "  ${BOLD}DEPENDENCIES${NC}\n"
    printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

    printf "  %-20s │  %s\n" "Total" "$(format_number "$total")"
    printf "  %-20s │  %s\n" "Direct" "$(format_number "$direct")"
    printf "  %-20s │  %s\n" "Transitive" "$(format_number "$((total - direct))")"

    echo
    hr "$BOX_LINE" 68
    echo

    printf "  ${BOLD}HEALTH ISSUES${NC}\n"
    printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

    printf "  %-20s │  " "Abandoned"
    if [[ "$abandoned" -gt 0 ]]; then printf "${YELLOW}%s${NC}\n" "$abandoned"; else printf "${GREEN}0${NC}\n"; fi
    printf "  %-20s │  " "Deprecated"
    if [[ "$deprecated" -gt 0 ]]; then printf "${YELLOW}%s${NC}\n" "$deprecated"; else printf "${GREEN}0${NC}\n"; fi
    printf "  %-20s │  %s\n" "Outdated" "$outdated"
    printf "  %-20s │  " "Typosquatting"
    if [[ "$typosquat" -gt 0 ]]; then printf "${RED}%s${NC}\n" "$typosquat"; else printf "${GREEN}0${NC}\n"; fi

    # Provenance
    local signed=$(echo "$json" | jq -r '.provenance.signed_packages // 0')
    local slsa=$(echo "$json" | jq -r '.provenance.slsa_level // "none"')

    echo
    hr "$BOX_LINE" 68
    echo

    printf "  ${BOLD}PROVENANCE${NC}\n"
    printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"
    printf "  %-20s │  %s\n" "Signed Packages" "$signed"
    printf "  %-20s │  %s\n" "SLSA Level" "$slsa"

    echo
    hr "$BOX_LINE" 68
    echo -e "  ${DIM}Generated by Phantom Report v${REPORT_VERSION}${NC}"
    echo
}

# Format DORA report for terminal
format_dora_terminal() {
    local json="$1"
    local target_id="$2"

    local project_id=$(echo "$json" | jq -r '.project.id // .organization // "Unknown"')
    local profile=$(echo "$json" | jq -r '.project.profile // "standard"')
    local completed_at=$(echo "$json" | jq -r '.project.completed_at // ""')

    local overall_perf=$(echo "$json" | jq -r '.dora.overall_performance // "N/A"')
    local description=$(echo "$json" | jq -r '.dora.description // ""')

    local deploy_freq=$(echo "$json" | jq -r '.dora.deployment_frequency.value // "N/A"')
    local deploy_level=$(echo "$json" | jq -r '.dora.deployment_frequency.level // "N/A"')
    local lead_time=$(echo "$json" | jq -r '.dora.lead_time.value // "N/A"')
    local lead_level=$(echo "$json" | jq -r '.dora.lead_time.level // "N/A"')
    local change_fail=$(echo "$json" | jq -r '.dora.change_failure_rate.value // "N/A"')
    local change_level=$(echo "$json" | jq -r '.dora.change_failure_rate.level // "N/A"')
    local mttr=$(echo "$json" | jq -r '.dora.mttr.value // "N/A"')
    local mttr_level=$(echo "$json" | jq -r '.dora.mttr.level // "N/A"')

    echo
    printf "${BOLD}${BOX_TL}"
    printf '%*s' 66 '' | tr ' ' "$BOX_H"
    printf "${BOX_TR}${NC}\n"
    printf "${BOLD}${BOX_V}  PHANTOM DORA METRICS REPORT%*s${BOX_V}${NC}\n" 37 ''
    printf "${BOLD}${BOX_BL}"
    printf '%*s' 66 '' | tr ' ' "$BOX_H"
    printf "${BOX_BR}${NC}\n"
    echo

    printf "  ${BOLD}Project:${NC}     %s\n" "$project_id"
    printf "  ${BOLD}Scanned:${NC}     %s ${DIM}(%s)${NC}\n" "$(format_timestamp "$completed_at")" "$(relative_time "$completed_at")"

    echo
    hr "$BOX_LINE" 68
    echo

    printf "  ${BOLD}OVERALL PERFORMANCE${NC}\n"
    printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

    printf "  %-20s │  " "Level"
    case "$overall_perf" in
        ELITE) printf "${GREEN}%s${NC}\n" "$overall_perf" ;;
        HIGH) printf "${GREEN}%s${NC}\n" "$overall_perf" ;;
        MEDIUM) printf "${YELLOW}%s${NC}\n" "$overall_perf" ;;
        LOW) printf "${RED}%s${NC}\n" "$overall_perf" ;;
        *) printf "%s\n" "$overall_perf" ;;
    esac

    if [[ -n "$description" ]] && [[ "$description" != "null" ]]; then
        printf "\n  ${DIM}%s${NC}\n" "$description"
    fi

    echo
    hr "$BOX_LINE" 68
    echo

    printf "  ${BOLD}METRICS${NC}\n"
    printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

    printf "  %-24s │  %-20s │  %s\n" "Deployment Frequency" "$deploy_freq" "$deploy_level"
    printf "  %-24s │  %-20s │  %s\n" "Lead Time for Changes" "$lead_time" "$lead_level"
    printf "  %-24s │  %-20s │  %s\n" "Change Failure Rate" "$change_fail" "$change_level"
    printf "  %-24s │  %-20s │  %s\n" "Mean Time to Recovery" "$mttr" "$mttr_level"

    echo
    hr "$BOX_LINE" 68
    echo -e "  ${DIM}Generated by Phantom Report v${REPORT_VERSION}${NC}"
    echo
}

# Format full report for terminal
format_full_terminal() {
    local json="$1"
    local target_id="$2"

    local project_id=$(echo "$json" | jq -r '.project.id // .organization // "Unknown"')
    local profile=$(echo "$json" | jq -r '.project.profile // "standard"')
    local completed_at=$(echo "$json" | jq -r '.project.completed_at // ""')

    # Security
    local critical=$(echo "$json" | jq -r '.security.vulnerabilities.critical // 0')
    local high=$(echo "$json" | jq -r '.security.vulnerabilities.high // 0')
    local medium=$(echo "$json" | jq -r '.security.vulnerabilities.medium // 0')
    local secrets=$(echo "$json" | jq -r '.security.secrets.total // 0')

    # Dependencies
    local total_deps=$(echo "$json" | jq -r '.dependencies.total // 0')
    local abandoned=$(echo "$json" | jq -r '.dependencies.health.abandoned // 0')

    # Compliance
    local license_status=$(echo "$json" | jq -r '.compliance.license_status // "unknown"')

    echo
    printf "${BOLD}${BOX_TL}"
    printf '%*s' 66 '' | tr ' ' "$BOX_H"
    printf "${BOX_TR}${NC}\n"
    printf "${BOLD}${BOX_V}  PHANTOM FULL ANALYSIS REPORT%*s${BOX_V}${NC}\n" 36 ''
    printf "${BOLD}${BOX_BL}"
    printf '%*s' 66 '' | tr ' ' "$BOX_H"
    printf "${BOX_BR}${NC}\n"
    echo

    printf "  ${BOLD}Project:${NC}     %s\n" "$project_id"
    printf "  ${BOLD}Scanned:${NC}     %s ${DIM}(%s)${NC}\n" "$(format_timestamp "$completed_at")" "$(relative_time "$completed_at")"
    printf "  ${BOLD}Profile:${NC}     %s\n" "$profile"

    echo
    hr "$BOX_LINE" 68
    echo

    printf "  ${BOLD}SECURITY${NC}\n"
    printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

    printf "  %-20s │  " "Vulnerabilities"
    if [[ "$critical" -gt 0 ]]; then printf "${RED}%s critical${NC}  " "$critical"; fi
    if [[ "$high" -gt 0 ]]; then printf "${YELLOW}%s high${NC}  " "$high"; fi
    if [[ "$medium" -gt 0 ]]; then printf "%s medium  " "$medium"; fi
    if [[ "$critical" -eq 0 ]] && [[ "$high" -eq 0 ]] && [[ "$medium" -eq 0 ]]; then printf "${GREEN}None${NC}"; fi
    echo

    printf "  %-20s │  " "Exposed Secrets"
    if [[ "$secrets" -gt 0 ]]; then printf "${RED}%s${NC}\n" "$secrets"; else printf "${GREEN}0${NC}\n"; fi

    echo
    hr "$BOX_LINE" 68
    echo

    printf "  ${BOLD}DEPENDENCIES${NC}\n"
    printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

    printf "  %-20s │  %s\n" "Total" "$(format_number "$total_deps")"
    printf "  %-20s │  " "Abandoned"
    if [[ "$abandoned" -gt 0 ]]; then printf "${YELLOW}%s${NC}\n" "$abandoned"; else printf "0\n"; fi

    echo
    hr "$BOX_LINE" 68
    echo

    printf "  ${BOLD}COMPLIANCE${NC}\n"
    printf "  %s\n" "$(printf '%*s' 64 '' | tr ' ' '─')"

    printf "  %-20s │  " "License Status"
    case "$license_status" in
        pass) printf "${GREEN}Pass${NC}\n" ;;
        fail) printf "${RED}Fail${NC}\n" ;;
        *) printf "%s\n" "$license_status" ;;
    esac

    echo
    hr "$BOX_LINE" 68
    echo -e "  ${DIM}Generated by Phantom Report v${REPORT_VERSION}${NC}"
    echo
}

export -f format_report_output
export -f format_summary_terminal
export -f format_project_summary_terminal
export -f format_org_summary_terminal
export -f format_security_terminal
export -f format_licenses_terminal
export -f format_sbom_terminal
export -f format_compliance_terminal
export -f format_supply_chain_terminal
export -f format_dora_terminal
export -f format_full_terminal
