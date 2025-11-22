#!/bin/bash
# Copyright (c) 2024 Gibson Powers Contributors
# 
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# CSV Export Library
# Export analysis data to CSV format for Excel, Google Sheets, etc.
#
# Key Features:
# - Multiple export formats (ownership, metrics, SPOFs, trends)
# - Proper CSV escaping (quotes, commas, newlines)
# - Header generation
# - Multi-sheet export support (creates multiple CSVs)
#############################################################################

# Escape CSV field
csv_escape() {
    local field="$1"

    # If field contains comma, quote, or newline, wrap in quotes and escape internal quotes
    if [[ "$field" =~ [,\"$'\n'] ]]; then
        field=$(echo "$field" | sed 's/"/""/g')
        echo "\"$field\""
    else
        echo "$field"
    fi
}

# Generate CSV header
csv_header() {
    local -a fields=("$@")
    local first=true

    for field in "${fields[@]}"; do
        if [[ "$first" != "true" ]]; then
            printf ","
        fi
        first=false
        printf "%s" "$(csv_escape "$field")"
    done
    echo ""
}

# Generate CSV row
csv_row() {
    local -a values=("$@")
    local first=true

    for value in "${values[@]}"; do
        if [[ "$first" != "true" ]]; then
            printf ","
        fi
        first=false
        printf "%s" "$(csv_escape "$value")"
    done
    echo ""
}

# Export ownership data to CSV
export_ownership_csv() {
    local json_data="$1"
    local output_file="$2"

    # Header
    csv_header "Email" "Name" "Files Owned" "Percentage" > "$output_file"

    # Calculate total files
    local total_files=$(echo "$json_data" | jq -r '.repository_metrics.total_files')

    # Data rows
    echo "$json_data" | jq -r '.contributors[] | "\(.email)|\(.name // .email)|\(.files_owned)"' | \
    while IFS='|' read -r email name files_owned; do
        local percentage=$(echo "scale=2; ($files_owned / $total_files) * 100" | bc -l)
        csv_row "$email" "$name" "$files_owned" "${percentage}%"
    done >> "$output_file"
}

# Export metrics to CSV
export_metrics_csv() {
    local json_data="$1"
    local output_file="$2"

    # Header
    csv_header "Metric" "Value" "Status" > "$output_file"

    # Extract metrics
    local health_score=$(echo "$json_data" | jq -r '.ownership_health.health_score')
    local health_grade=$(echo "$json_data" | jq -r '.ownership_health.health_grade')
    local coverage=$(echo "$json_data" | jq -r '.ownership_health.coverage_percentage')
    local gini=$(echo "$json_data" | jq -r '.ownership_health.gini_coefficient')
    local bus_factor=$(echo "$json_data" | jq -r '.ownership_health.bus_factor')
    local total_files=$(echo "$json_data" | jq -r '.repository_metrics.total_files')
    local total_commits=$(echo "$json_data" | jq -r '.repository_metrics.total_commits')
    local active_contributors=$(echo "$json_data" | jq -r '.repository_metrics.active_contributors')

    # Data rows
    csv_row "Health Score" "$health_score" "$health_grade"
    csv_row "Coverage Percentage" "$coverage" "$(if (( $(echo "$coverage >= 80" | bc -l) )); then echo "Good"; else echo "Needs Improvement"; fi)"
    csv_row "Gini Coefficient" "$gini" "$(if (( $(echo "$gini <= 0.5" | bc -l) )); then echo "Well Distributed"; else echo "Concentrated"; fi)"
    csv_row "Bus Factor" "$bus_factor" "$(if [[ $bus_factor -ge 3 ]]; then echo "Healthy"; else echo "At Risk"; fi)"
    csv_row "Total Files" "$total_files" ""
    csv_row "Total Commits" "$total_commits" ""
    csv_row "Active Contributors" "$active_contributors" ""
} >> "$output_file"

# Export SPOFs to CSV
export_spof_csv() {
    local json_data="$1"
    local output_file="$2"

    # Header
    csv_header "File" "Risk Level" "Score" "Contributors" "Critical" "Complex" "Has Backup" "Has Tests" "Has Docs" "Lines of Code" > "$output_file"

    # Data rows
    echo "$json_data" | jq -r '
        .single_points_of_failure[]
        | "\(.file)|\(.risk)|\(.score)|\(.contributors)|\(.critical)|\(.complex)|\(.has_backup)|\(.has_tests)|\(.has_docs)|\(.loc)"
    ' | while IFS='|' read -r file risk score contributors critical complex has_backup has_tests has_docs loc; do
        csv_row "$file" "$risk" "$score" "$contributors" "$critical" "$complex" "$has_backup" "$has_tests" "$has_docs" "$loc"
    done >> "$output_file"
}

# Export contributors with detailed stats to CSV
export_contributors_detailed_csv() {
    local json_data="$1"
    local output_file="$2"

    # Header
    csv_header "Email" "Name" "Files Owned" "Percentage" "Rank" "Status" > "$output_file"

    local total_files=$(echo "$json_data" | jq -r '.repository_metrics.total_files')

    # Data rows with ranking
    echo "$json_data" | jq -r '
        .contributors
        | sort_by(-.files_owned)
        | to_entries[]
        | "\(.key + 1)|\(.value.email)|\(.value.name // .value.email)|\(.value.files_owned)"
    ' | while IFS='|' read -r rank email name files_owned; do
        local percentage=$(echo "scale=2; ($files_owned / $total_files) * 100" | bc -l)

        # Determine status
        local status="Active"
        if (( $(echo "$percentage >= 20" | bc -l) )); then
            status="Major Contributor"
        elif (( $(echo "$percentage >= 10" | bc -l) )); then
            status="Regular Contributor"
        elif (( $(echo "$percentage >= 5" | bc -l) )); then
            status="Active"
        else
            status="Minor Contributor"
        fi

        csv_row "$email" "$name" "$files_owned" "${percentage}%" "$rank" "$status"
    done >> "$output_file"
}

# Export time-series data to CSV
export_timeseries_csv() {
    local repo_path="$1"
    local output_file="$2"

    # Need to source trends.sh for this
    if ! type list_snapshots &>/dev/null; then
        echo "Error: trends.sh must be sourced" >&2
        return 1
    fi

    # Header
    csv_header "Date" "Health Score" "Coverage %" "Gini Coefficient" "Bus Factor" "Active Contributors" "Total Files" > "$output_file"

    # Get all snapshots
    local snapshots=($(list_snapshots "$repo_path" 2>/dev/null))

    for snapshot in "${snapshots[@]}"; do
        if [[ ! -f "$snapshot" ]]; then
            continue
        fi

        local date=$(jq -r '.snapshot_date' "$snapshot")
        local health=$(jq -r '.ownership_health.health_score' "$snapshot")
        local coverage=$(jq -r '.ownership_health.coverage_percentage' "$snapshot")
        local gini=$(jq -r '.ownership_health.gini_coefficient' "$snapshot")
        local bus_factor=$(jq -r '.ownership_health.bus_factor' "$snapshot")
        local contributors=$(jq -r '.repository_metrics.active_contributors' "$snapshot")
        local total_files=$(jq -r '.repository_metrics.total_files' "$snapshot")

        csv_row "$date" "$health" "$coverage" "$gini" "$bus_factor" "$contributors" "$total_files"
    done >> "$output_file"
}

# Export file-level ownership details to CSV
export_file_ownership_csv() {
    local repo_path="$1"
    local output_file="$2"
    local since_date="$3"

    # Header
    csv_header "File" "Primary Owner" "Owner Email" "Commits" "Last Modified" "Contributors" "Ownership %" > "$output_file"

    cd "$repo_path" || return 1

    # Analyze each file
    git ls-files | while read -r file; do
        # Get primary owner (most commits)
        local owner_data=$(git log --since="$since_date" --format="%an|%ae" -- "$file" 2>/dev/null | \
            sort | uniq -c | sort -rn | head -1)

        if [[ -z "$owner_data" ]]; then
            continue
        fi

        local commits=$(echo "$owner_data" | awk '{print $1}')
        local owner_name=$(echo "$owner_data" | awk '{print $2}' | cut -d'|' -f1)
        local owner_email=$(echo "$owner_data" | awk '{print $2}' | cut -d'|' -f2)

        # Get total commits for this file
        local total_commits=$(git log --since="$since_date" --oneline -- "$file" 2>/dev/null | wc -l | tr -d ' ')
        local ownership_pct=$(echo "scale=1; ($commits / $total_commits) * 100" | bc -l)

        # Get last modified date
        local last_modified=$(git log -1 --format="%ad" --date=short -- "$file" 2>/dev/null)

        # Get total contributors
        local total_contributors=$(git log --since="$since_date" --format="%ae" -- "$file" 2>/dev/null | sort -u | wc -l | tr -d ' ')

        csv_row "$file" "$owner_name" "$owner_email" "$commits" "$last_modified" "$total_contributors" "${ownership_pct}%"
    done >> "$output_file"
}

# Export all data (multi-file export)
export_all_csv() {
    local json_data="$1"
    local output_prefix="$2"
    local repo_path="${3:-}"

    echo "Exporting to CSV..."

    # Export ownership
    export_ownership_csv "$json_data" "${output_prefix}_ownership.csv"
    echo "✓ Exported ownership: ${output_prefix}_ownership.csv"

    # Export metrics
    export_metrics_csv "$json_data" "${output_prefix}_metrics.csv"
    echo "✓ Exported metrics: ${output_prefix}_metrics.csv"

    # Export SPOFs
    local spof_count=$(echo "$json_data" | jq -r '.single_points_of_failure | length')
    if [[ $spof_count -gt 0 ]]; then
        export_spof_csv "$json_data" "${output_prefix}_spofs.csv"
        echo "✓ Exported SPOFs: ${output_prefix}_spofs.csv"
    fi

    # Export detailed contributors
    export_contributors_detailed_csv "$json_data" "${output_prefix}_contributors_detailed.csv"
    echo "✓ Exported detailed contributors: ${output_prefix}_contributors_detailed.csv"

    # Export time-series if trends exist
    if [[ -n "$repo_path" ]] && type list_snapshots &>/dev/null; then
        if list_snapshots "$repo_path" &>/dev/null; then
            export_timeseries_csv "$repo_path" "${output_prefix}_timeseries.csv"
            echo "✓ Exported time-series: ${output_prefix}_timeseries.csv"
        fi
    fi

    # Export file-level data if repo_path provided
    if [[ -n "$repo_path" ]]; then
        local since_date=$(date -v-90d +%Y-%m-%d 2>/dev/null || date -d "90 days ago" +%Y-%m-%d)
        export_file_ownership_csv "$repo_path" "${output_prefix}_files.csv" "$since_date"
        echo "✓ Exported file ownership: ${output_prefix}_files.csv"
    fi

    echo ""
    echo "All CSV files exported with prefix: $output_prefix"
}

# Create CSV manifest (index of all exported files)
create_csv_manifest() {
    local output_prefix="$1"
    local manifest_file="${output_prefix}_manifest.txt"

    cat > "$manifest_file" << EOF
Code Ownership Analysis - CSV Export Manifest
Generated: $(date)

Available Files:
EOF

    for csv_file in "${output_prefix}"_*.csv; do
        if [[ -f "$csv_file" ]]; then
            local row_count=$(wc -l < "$csv_file" | tr -d ' ')
            echo "- $(basename "$csv_file") ($row_count rows)" >> "$manifest_file"
        fi
    done

    echo ""
    echo "Created manifest: $manifest_file"
}

# Export functions
export -f csv_escape
export -f csv_header
export -f csv_row
export -f export_ownership_csv
export -f export_metrics_csv
export -f export_spof_csv
export -f export_contributors_detailed_csv
export -f export_timeseries_csv
export -f export_file_ownership_csv
export -f export_all_csv
export -f create_csv_manifest
