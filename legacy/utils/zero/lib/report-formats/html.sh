#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# HTML Format Output
# Self-contained HTML output for reports
#############################################################################

# Format report output to HTML
format_report_output() {
    local json_data="$1"
    local target_id="$2"

    local report_type=$(echo "$json_data" | jq -r '.report_type')

    # Generate HTML with embedded CSS
    cat << 'HTMLHEAD'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Phantom Report</title>
    <style>
        :root {
            --bg-primary: #1a1a2e;
            --bg-secondary: #16213e;
            --bg-card: #0f3460;
            --text-primary: #eaeaea;
            --text-secondary: #a0a0a0;
            --accent: #e94560;
            --success: #00d9a0;
            --warning: #ffc107;
            --danger: #dc3545;
            --info: #17a2b8;
        }

        * { box-sizing: border-box; margin: 0; padding: 0; }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            background: var(--bg-primary);
            color: var(--text-primary);
            line-height: 1.6;
            padding: 2rem;
        }

        .container { max-width: 1200px; margin: 0 auto; }

        header {
            background: linear-gradient(135deg, var(--bg-secondary), var(--bg-card));
            padding: 2rem;
            border-radius: 12px;
            margin-bottom: 2rem;
            border-left: 4px solid var(--accent);
        }

        h1 { font-size: 2rem; margin-bottom: 0.5rem; }
        h2 { font-size: 1.5rem; margin: 1.5rem 0 1rem; color: var(--accent); }
        h3 { font-size: 1.2rem; margin: 1rem 0 0.5rem; }

        .meta { color: var(--text-secondary); font-size: 0.9rem; }
        .meta span { margin-right: 1.5rem; }

        .badge {
            display: inline-block;
            padding: 0.25rem 0.75rem;
            border-radius: 20px;
            font-size: 0.85rem;
            font-weight: 600;
            margin-right: 0.5rem;
        }
        .badge-critical { background: var(--danger); color: white; }
        .badge-high { background: #fd7e14; color: white; }
        .badge-medium { background: var(--warning); color: #333; }
        .badge-low { background: var(--success); color: white; }
        .badge-none { background: var(--success); color: white; }
        .badge-info { background: var(--info); color: white; }

        .card {
            background: var(--bg-secondary);
            border-radius: 8px;
            padding: 1.5rem;
            margin-bottom: 1.5rem;
        }

        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 1.5rem; }

        .stat-card {
            background: var(--bg-card);
            border-radius: 8px;
            padding: 1.5rem;
            text-align: center;
        }
        .stat-value { font-size: 2.5rem; font-weight: 700; }
        .stat-label { color: var(--text-secondary); font-size: 0.9rem; }

        .score-bar {
            height: 8px;
            background: var(--bg-primary);
            border-radius: 4px;
            overflow: hidden;
            margin-top: 0.5rem;
        }
        .score-fill {
            height: 100%;
            border-radius: 4px;
            transition: width 0.3s ease;
        }
        .score-fill.good { background: var(--success); }
        .score-fill.medium { background: var(--warning); }
        .score-fill.bad { background: var(--danger); }

        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 1rem;
        }
        th, td {
            padding: 0.75rem;
            text-align: left;
            border-bottom: 1px solid var(--bg-card);
        }
        th { color: var(--text-secondary); font-weight: 600; }
        tr:hover { background: rgba(255,255,255,0.05); }

        .risk-indicator {
            width: 12px;
            height: 12px;
            border-radius: 50%;
            display: inline-block;
            margin-right: 0.5rem;
        }
        .risk-critical { background: var(--danger); }
        .risk-high { background: #fd7e14; }
        .risk-medium { background: var(--warning); }
        .risk-low { background: var(--success); }
        .risk-none { background: var(--success); }

        .summary-box {
            background: rgba(233, 69, 96, 0.1);
            border-left: 4px solid var(--accent);
            padding: 1rem 1.5rem;
            border-radius: 0 8px 8px 0;
            margin-bottom: 1.5rem;
        }

        .recommendations li {
            margin-bottom: 0.75rem;
            padding-left: 1.5rem;
            position: relative;
        }
        .recommendations li::before {
            content: "→";
            position: absolute;
            left: 0;
            color: var(--accent);
        }

        footer {
            margin-top: 3rem;
            padding-top: 1.5rem;
            border-top: 1px solid var(--bg-card);
            color: var(--text-secondary);
            font-size: 0.85rem;
            text-align: center;
        }

        @media print {
            body { background: white; color: #333; }
            .card, .stat-card, header { background: #f5f5f5; }
        }
    </style>
</head>
<body>
<div class="container">
HTMLHEAD

    # Generate body based on report type
    case "$report_type" in
        summary) format_summary_html "$json_data" ;;
        security) format_security_html "$json_data" ;;
        licenses) format_licenses_html "$json_data" ;;
        compliance) format_compliance_html "$json_data" ;;
        supply-chain) format_supply_chain_html "$json_data" ;;
        dora) format_dora_html "$json_data" ;;
        sbom) format_sbom_html "$json_data" ;;
        full) format_full_html "$json_data" ;;
        code-ownership) format_code_ownership_html "$json_data" ;;
        *) format_summary_html "$json_data" ;;
    esac

    # Footer
    local generated_at=$(echo "$json_data" | jq -r '.generated_at // ""')
    cat << HTMLFOOT
<footer>
    Generated by Phantom Report v${REPORT_VERSION}<br>
    ${generated_at}
</footer>
</div>
</body>
</html>
HTMLFOOT
}

# Get score class based on value
get_score_class() {
    local score="$1"
    if [[ $score -ge 70 ]]; then
        echo "good"
    elif [[ $score -ge 40 ]]; then
        echo "medium"
    else
        echo "bad"
    fi
}

# Format summary report for HTML
format_summary_html() {
    local json="$1"

    local project_id=$(echo "$json" | jq -r '.project.id // .organization // "Unknown"')
    local risk_level=$(echo "$json" | jq -r '.risk.level // "unknown"')
    local profile=$(echo "$json" | jq -r '.project.profile // "N/A"')

    local critical=$(echo "$json" | jq -r '.risk.vulnerabilities.critical // 0')
    local high=$(echo "$json" | jq -r '.risk.vulnerabilities.high // 0')
    local medium=$(echo "$json" | jq -r '.risk.vulnerabilities.medium // 0')
    local total_deps=$(echo "$json" | jq -r '.dependencies.total // 0')
    local secrets=$(echo "$json" | jq -r '.secrets.exposed // 0')

    local risk_level_upper=$(echo "$risk_level" | tr '[:lower:]' '[:upper:]')
    cat << HTMLBODY
<header>
    <h1>Phantom Summary Report</h1>
    <div class="meta">
        <span><strong>Project:</strong> ${project_id}</span>
        <span><strong>Profile:</strong> ${profile}</span>
    </div>
</header>

<div class="grid">
    <div class="stat-card">
        <div class="stat-value"><span class="risk-indicator risk-${risk_level}"></span>${risk_level_upper}</div>
        <div class="stat-label">Risk Level</div>
    </div>
    <div class="stat-card">
        <div class="stat-value">${total_deps}</div>
        <div class="stat-label">Dependencies</div>
    </div>
    <div class="stat-card">
        <div class="stat-value" style="color: $([ $critical -gt 0 ] && echo 'var(--danger)' || echo 'var(--success)')">${critical}</div>
        <div class="stat-label">Critical Vulns</div>
    </div>
    <div class="stat-card">
        <div class="stat-value" style="color: $([ $secrets -gt 0 ] && echo 'var(--danger)' || echo 'var(--success)')">${secrets}</div>
        <div class="stat-label">Exposed Secrets</div>
    </div>
</div>

<div class="card">
    <h2>Vulnerability Summary</h2>
    <table>
        <tr><th>Severity</th><th>Count</th><th>Status</th></tr>
        <tr><td><span class="badge badge-critical">Critical</span></td><td>${critical}</td><td>$([ $critical -eq 0 ] && echo '✅' || echo '⚠️ Action Required')</td></tr>
        <tr><td><span class="badge badge-high">High</span></td><td>${high}</td><td>$([ $high -eq 0 ] && echo '✅' || echo '⚠️')</td></tr>
        <tr><td><span class="badge badge-medium">Medium</span></td><td>${medium}</td><td>$([ $medium -eq 0 ] && echo '✅' || echo 'Monitor')</td></tr>
    </table>
</div>
HTMLBODY
}

# Format security report for HTML
format_security_html() {
    local json="$1"

    local project_id=$(echo "$json" | jq -r '.project.id // "Unknown"')
    local security_score=$(echo "$json" | jq -r '.security_score // 0')
    local critical=$(echo "$json" | jq -r '.vulnerabilities.summary.critical // 0')
    local high=$(echo "$json" | jq -r '.vulnerabilities.summary.high // 0')
    local secrets=$(echo "$json" | jq -r '.secrets.summary.total // 0')
    local code_sec=$(echo "$json" | jq -r '.code_security.summary.total // 0')
    local iac=$(echo "$json" | jq -r '.iac_security.summary.total // 0')

    cat << HTMLBODY
<header>
    <h1>Security Report</h1>
    <div class="meta"><span><strong>Project:</strong> ${project_id}</span></div>
</header>

<div class="card">
    <h2>Security Score</h2>
    <div class="stat-value">${security_score}/100</div>
    <div class="score-bar"><div class="score-fill $(get_score_class $security_score)" style="width: ${security_score}%"></div></div>
</div>

<div class="grid">
    <div class="stat-card">
        <div class="stat-value" style="color: var(--danger)">${critical}</div>
        <div class="stat-label">Critical Vulnerabilities</div>
    </div>
    <div class="stat-card">
        <div class="stat-value" style="color: #fd7e14">${high}</div>
        <div class="stat-label">High Vulnerabilities</div>
    </div>
    <div class="stat-card">
        <div class="stat-value">${secrets}</div>
        <div class="stat-label">Exposed Secrets</div>
    </div>
    <div class="stat-card">
        <div class="stat-value">${code_sec}</div>
        <div class="stat-label">Code Security Issues</div>
    </div>
</div>
HTMLBODY
}

# Format licenses report for HTML
format_licenses_html() {
    local json="$1"

    local project_id=$(echo "$json" | jq -r '.project.id // .organization // "Unknown"')
    local overall_status=$(echo "$json" | jq -r '.overall_status // "unknown"')
    local repo_license=$(echo "$json" | jq -r '.repository_license.license // "Not Found"')
    local repo_license_file=$(echo "$json" | jq -r '.repository_license.file // ""')
    local total_deps=$(echo "$json" | jq -r '.summary.total_dependencies_scanned // 0')
    local denied_count=$(echo "$json" | jq -r '.summary.denied_license_packages // 0')
    local review_count=$(echo "$json" | jq -r '.summary.review_required_packages // 0')

    local status_class="good"
    [[ "$overall_status" == "fail" ]] && status_class="critical"
    [[ "$overall_status" == "warning" ]] && status_class="warning"
    local overall_status_upper=$(echo "$overall_status" | tr '[:lower:]' '[:upper:]')

    cat << HTMLBODY
<header>
    <h1>License Report</h1>
    <div class="meta"><span><strong>Project:</strong> ${project_id}</span></div>
</header>

<div class="grid">
    <div class="stat-card">
        <div class="stat-value ${status_class}">${overall_status_upper}</div>
        <div class="stat-label">Overall Status</div>
    </div>
    <div class="stat-card">
        <div class="stat-value">${repo_license}</div>
        <div class="stat-label">Repository License</div>
        <div style="font-size: 0.8em; color: #666;">${repo_license_file}</div>
    </div>
    <div class="stat-card">
        <div class="stat-value">${total_deps}</div>
        <div class="stat-label">Dependencies Scanned</div>
    </div>
</div>

<div class="card">
    <h2>License Overview</h2>
    <table>
        <tr><th>Category</th><th>Count</th></tr>
        <tr><td>Denied Licenses (GPL, AGPL)</td><td class="${denied_count:+critical}">${denied_count}</td></tr>
        <tr><td>Review Required (LGPL, MPL)</td><td class="${review_count:+warning}">${review_count}</td></tr>
    </table>
</div>

<div class="card">
    <h2>Dependency Licenses</h2>
    <table>
        <tr><th>License</th><th>Package Count</th></tr>
$(echo "$json" | jq -r '.dependency_licenses.by_license | to_entries | sort_by(-.value.count) | .[:20][] | "<tr><td>\(.key)</td><td>\(.value.count)</td></tr>"' 2>/dev/null || echo "<tr><td colspan='2'>No license data available</td></tr>")
    </table>
</div>
HTMLBODY
}

# Simplified implementations for other report types
format_compliance_html() {
    format_summary_html "$1"
}

format_supply_chain_html() {
    format_summary_html "$1"
}

format_dora_html() {
    local json="$1"
    local project_id=$(echo "$json" | jq -r '.project.id // .organization // "Unknown"')
    local perf=$(echo "$json" | jq -r '.dora.overall_performance // "N/A"')
    local desc=$(echo "$json" | jq -r '.dora.description // ""')

    cat << HTMLBODY
<header>
    <h1>DORA Metrics Report</h1>
    <div class="meta"><span><strong>Project:</strong> ${project_id}</span></div>
</header>

<div class="stat-card" style="max-width: 400px; margin: 2rem auto;">
    <div class="stat-value">${perf}</div>
    <div class="stat-label">DORA Performance Level</div>
</div>

<div class="summary-box">
    <p>${desc}</p>
</div>
HTMLBODY
}

format_sbom_html() {
    local json="$1"
    local project_id=$(echo "$json" | jq -r '.project.id // "Unknown"')
    local total=$(echo "$json" | jq -r '.summary.total_components // 0')
    local direct=$(echo "$json" | jq -r '.summary.direct_dependencies // 0')

    cat << HTMLBODY
<header>
    <h1>Software Bill of Materials</h1>
    <div class="meta"><span><strong>Project:</strong> ${project_id}</span></div>
</header>

<div class="grid">
    <div class="stat-card">
        <div class="stat-value">${total}</div>
        <div class="stat-label">Total Components</div>
    </div>
    <div class="stat-card">
        <div class="stat-value">${direct}</div>
        <div class="stat-label">Direct Dependencies</div>
    </div>
</div>

<div class="card">
    <h2>Components</h2>
    <table>
        <tr><th>Name</th><th>Version</th><th>Ecosystem</th><th>License</th></tr>
$(echo "$json" | jq -r '.components[0:50][]? | "<tr><td>\(.name)</td><td>\(.version)</td><td>\(.ecosystem)</td><td>\(.license)</td></tr>"' 2>/dev/null || echo "<tr><td colspan='4'>No components data available</td></tr>")
    </table>
</div>
HTMLBODY
}

format_full_html() {
    local json="$1"
    local project_id=$(echo "$json" | jq -r '.project.id // "Unknown"')
    local overall=$(echo "$json" | jq -r '.scores.overall // 0')
    local security=$(echo "$json" | jq -r '.scores.security // 0')
    local compliance=$(echo "$json" | jq -r '.scores.compliance // 0')
    local supply_chain=$(echo "$json" | jq -r '.scores.supply_chain // 0')
    local risk_level=$(echo "$json" | jq -r '.risk.level // "unknown"')
    local risk_level_upper=$(echo "$risk_level" | tr '[:lower:]' '[:upper:]')

    cat << HTMLBODY
<header>
    <h1>Full Analysis Report</h1>
    <div class="meta"><span><strong>Project:</strong> ${project_id}</span></div>
</header>

<div class="grid">
    <div class="stat-card">
        <div class="stat-value">${overall}</div>
        <div class="stat-label">Overall Score</div>
        <div class="score-bar"><div class="score-fill $(get_score_class $overall)" style="width: ${overall}%"></div></div>
    </div>
    <div class="stat-card">
        <div class="stat-value"><span class="risk-indicator risk-${risk_level}"></span>${risk_level_upper}</div>
        <div class="stat-label">Risk Level</div>
    </div>
</div>

<h2>Score Breakdown</h2>
<div class="grid">
    <div class="stat-card">
        <div class="stat-value">${security}</div>
        <div class="stat-label">Security</div>
        <div class="score-bar"><div class="score-fill $(get_score_class $security)" style="width: ${security}%"></div></div>
    </div>
    <div class="stat-card">
        <div class="stat-value">${compliance}</div>
        <div class="stat-label">Compliance</div>
        <div class="score-bar"><div class="score-fill $(get_score_class $compliance)" style="width: ${compliance}%"></div></div>
    </div>
    <div class="stat-card">
        <div class="stat-value">${supply_chain}</div>
        <div class="stat-label">Supply Chain</div>
        <div class="score-bar"><div class="score-fill $(get_score_class $supply_chain)" style="width: ${supply_chain}%"></div></div>
    </div>
</div>

<div class="card">
    <h2>Security Overview</h2>
    <table>
        <tr><th>Metric</th><th>Value</th></tr>
        <tr><td>Critical Vulnerabilities</td><td>$(echo "$json" | jq -r '.security.vulnerabilities.critical')</td></tr>
        <tr><td>High Vulnerabilities</td><td>$(echo "$json" | jq -r '.security.vulnerabilities.high')</td></tr>
        <tr><td>Exposed Secrets</td><td>$(echo "$json" | jq -r '.security.secrets.total')</td></tr>
        <tr><td>Code Security Issues</td><td>$(echo "$json" | jq -r '.security.code_security_findings')</td></tr>
    </table>
</div>

<div class="card">
    <h2>Dependencies</h2>
    <table>
        <tr><th>Metric</th><th>Value</th></tr>
        <tr><td>Total Dependencies</td><td>$(echo "$json" | jq -r '.dependencies.total')</td></tr>
        <tr><td>Direct Dependencies</td><td>$(echo "$json" | jq -r '.dependencies.direct')</td></tr>
        <tr><td>Abandoned Packages</td><td>$(echo "$json" | jq -r '.dependencies.health.abandoned')</td></tr>
        <tr><td>Typosquatting Suspects</td><td>$(echo "$json" | jq -r '.dependencies.health.typosquatting_suspects')</td></tr>
    </table>
</div>
HTMLBODY
}

format_code_ownership_html() {
    local json="$1"
    local project_id=$(echo "$json" | jq -r '.project.id // "Unknown"')
    local tier1=$(echo "$json" | jq -r '.tiers.basic // false')
    local tier2=$(echo "$json" | jq -r '.tiers.analysis // false')
    local tier3=$(echo "$json" | jq -r '.tiers.ai_insights // false')

    cat << HTMLBODY
<header>
    <h1>Code Ownership Report</h1>
    <div class="meta"><span><strong>Project:</strong> ${project_id}</span></div>
</header>

<div class="card">
    <h2>Tier 1: Basic (CODEOWNERS Detection)</h2>
HTMLBODY

    if [[ "$tier1" == "true" ]]; then
        local codeowners_exists=$(echo "$json" | jq -r '.tier1_basic.codeowners.exists // false')
        if [[ "$codeowners_exists" == "true" ]]; then
            local codeowners_path=$(echo "$json" | jq -r '.tier1_basic.codeowners.path // ""')
            local codeowners_valid=$(echo "$json" | jq -r '.tier1_basic.codeowners.valid // "unknown"')
            echo "    <p><strong>CODEOWNERS File:</strong> ✅ Present (<code>${codeowners_path}</code>)</p>"
            echo "    <p><strong>Syntax:</strong> ${codeowners_valid}</p>"
        else
            echo "    <p><strong>CODEOWNERS File:</strong> ⚠️ Not Found</p>"
        fi
    else
        echo "    <p>No CODEOWNERS data available</p>"
    fi

    echo "</div>"

    cat << HTMLBODY
<div class="card">
    <h2>Tier 2: Analysis (Bus Factor & Concentration)</h2>
HTMLBODY

    if [[ "$tier2" == "true" ]]; then
        local bus_factor=$(echo "$json" | jq -r '.tier2_analysis.bus_factor.value // 0')
        local risk_level=$(echo "$json" | jq -r '.tier2_analysis.bus_factor.risk_level // "unknown"')
        local risk_desc=$(echo "$json" | jq -r '.tier2_analysis.bus_factor.risk_description // ""')
        local gini=$(echo "$json" | jq -r '.tier2_analysis.concentration.gini_coefficient // 0')
        local top1_pct=$(echo "$json" | jq -r '.tier2_analysis.concentration.top_contributor_percentage // 0')
        local top3_pct=$(echo "$json" | jq -r '.tier2_analysis.concentration.top_3_contributors_percentage // 0')
        local risk_level_upper=$(echo "$risk_level" | tr '[:lower:]' '[:upper:]')

        local risk_class="badge-low"
        [[ "$risk_level" == "medium" ]] && risk_class="badge-medium"
        [[ "$risk_level" == "high" ]] && risk_class="badge-high"
        [[ "$risk_level" == "critical" ]] && risk_class="badge-critical"

        cat << HTMLBODY
    <div class="grid">
        <div class="stat-card">
            <div class="stat-value">${bus_factor}</div>
            <div class="stat-label">Bus Factor</div>
            <span class="badge ${risk_class}">${risk_level_upper}</span>
        </div>
        <div class="stat-card">
            <div class="stat-value">${gini}</div>
            <div class="stat-label">Gini Coefficient</div>
        </div>
        <div class="stat-card">
            <div class="stat-value">${top1_pct}%</div>
            <div class="stat-label">Top Contributor</div>
        </div>
    </div>
    <div class="summary-box">
        <p>${risk_desc}</p>
    </div>
    <h3>Top Contributors</h3>
    <table>
        <tr><th>Contributor</th><th>Commits</th><th>Ownership</th></tr>
$(echo "$json" | jq -r '.tier2_analysis.contributors[:5][]? | "<tr><td>\(.name)</td><td>\(.commits)</td><td>\(.ownership_percentage)%</td></tr>"' 2>/dev/null)
    </table>
HTMLBODY
    else
        echo "    <p>No bus factor analysis available - run bus-factor scanner</p>"
    fi

    echo "</div>"

    cat << HTMLBODY
<div class="card">
    <h2>Tier 3: AI Insights (Claude Analysis)</h2>
HTMLBODY

    if [[ "$tier3" == "true" ]]; then
        echo "    <h3>Key Insights</h3>"
        echo "    <ul>"
        echo "$json" | jq -r '.tier3_ai_insights.insights[]? | "<li>\(.)</li>"' 2>/dev/null
        echo "    </ul>"
        echo "    <h3>Recommendations</h3>"
        echo "    <ul>"
        echo "$json" | jq -r '.tier3_ai_insights.recommendations[]? | "<li>\(.)</li>"' 2>/dev/null
        echo "    </ul>"
    else
        cat << HTMLBODY
    <p><em>AI analysis not available</em></p>
    <p>Use <code>--deep</code> profile or run with Claude for insights like:</p>
    <ul>
        <li>Critical risk areas and succession planning</li>
        <li>Knowledge transfer recommendations</li>
        <li>Auto-generated optimal CODEOWNERS file</li>
    </ul>
HTMLBODY
    fi

    echo "</div>"
}

export -f format_report_output
export -f format_summary_html
export -f format_security_html
export -f format_licenses_html
export -f format_compliance_html
export -f format_supply_chain_html
export -f format_dora_html
export -f format_sbom_html
export -f format_full_html
export -f format_code_ownership_html
export -f get_score_class
