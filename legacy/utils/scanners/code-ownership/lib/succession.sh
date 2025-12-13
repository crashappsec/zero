#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# 
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Succession Planning Library
# Automated successor identification and knowledge transfer planning
#
# Key Features:
# - Identify potential successors based on contribution patterns
# - Calculate readiness scores for knowledge transfer
# - Generate prioritized transition plans
# - Detect areas with no succession coverage
# - Recommend mentorship pairings
#############################################################################

# Identify primary owner and potential successors for a file
# Returns: primary_owner|ownership_score|successor1|readiness1|successor2|readiness2|...
identify_successors() {
    local repo_path="$1"
    local file_path="$2"
    local since_date="$3"
    local min_contribution="${4:-5}"  # Minimum contribution threshold (commits)

    cd "$repo_path" || return 1

    # Get all contributors to this file with their commit counts
    local contributors=$(git log --since="$since_date" --format="%ae" -- "$file_path" 2>/dev/null | \
        sort | uniq -c | sort -rn)

    if [[ -z "$contributors" ]]; then
        return 1
    fi

    # Parse contributors
    local -a contributor_emails=()
    local -a contributor_commits=()

    while read -r commits email; do
        contributor_emails+=("$email")
        contributor_commits+=("$commits")
    done <<< "$contributors"

    # Primary owner is the top contributor
    local primary_owner="${contributor_emails[0]}"
    local primary_commits="${contributor_commits[0]}"

    # Calculate primary owner's ownership score
    local primary_score=$(echo "scale=2; ($primary_commits / $(git log --since="$since_date" --oneline -- "$file_path" | wc -l | tr -d ' ')) * 100" | bc -l)

    # Identify potential successors (contributors with >=min_contribution commits)
    local result="$primary_owner|$primary_score"

    for ((i=1; i<${#contributor_emails[@]}; i++)); do
        local successor_email="${contributor_emails[i]}"
        local successor_commits="${contributor_commits[i]}"

        # Filter by minimum contribution threshold
        if [[ $successor_commits -lt $min_contribution ]]; then
            continue
        fi

        # Calculate readiness score (0-100)
        local readiness=$(calculate_successor_readiness \
            "$repo_path" \
            "$file_path" \
            "$successor_email" \
            "$primary_commits" \
            "$since_date")

        result="$result|$successor_email|$readiness"
    done

    echo "$result"
}

# Calculate successor readiness score
# Based on: contribution frequency, recency, code familiarity, collaboration
calculate_successor_readiness() {
    local repo_path="$1"
    local file_path="$2"
    local successor_email="$3"
    local primary_commits="$4"
    local since_date="$5"

    cd "$repo_path" || return 1

    # Factor 1: Contribution frequency (0-30 points)
    local commits=$(git log --since="$since_date" --author="$successor_email" --oneline -- "$file_path" 2>/dev/null | wc -l | tr -d ' ')
    local frequency_score=$(echo "scale=2; ($commits / $primary_commits) * 30" | bc -l)
    if (( $(echo "$frequency_score > 30" | bc -l) )); then
        frequency_score="30"
    fi

    # Factor 2: Recency (0-25 points)
    local last_commit=$(git log --since="$since_date" --author="$successor_email" --format="%ad" --date=short -- "$file_path" 2>/dev/null | head -1)
    local recency_score=0
    if [[ -n "$last_commit" ]]; then
        local days_since=$(( ($(date +%s) - $(date -j -f "%Y-%m-%d" "$last_commit" +%s 2>/dev/null || date -d "$last_commit" +%s)) / 86400 ))

        if [[ $days_since -lt 30 ]]; then
            recency_score=25
        elif [[ $days_since -lt 60 ]]; then
            recency_score=20
        elif [[ $days_since -lt 90 ]]; then
            recency_score=15
        else
            recency_score=10
        fi
    fi

    # Factor 3: Code familiarity (0-25 points)
    # Based on lines currently authored
    local lines_owned=0
    if file "$repo_path/$file_path" 2>/dev/null | grep -q "text"; then
        lines_owned=$(git blame --line-porcelain "$file_path" 2>/dev/null | \
            grep -c "author-mail <$successor_email>" || echo "0")
    fi
    local total_lines=$(wc -l < "$repo_path/$file_path" 2>/dev/null || echo "1")
    local familiarity_score=$(echo "scale=2; ($lines_owned / $total_lines) * 25" | bc -l)

    # Factor 4: Collaboration history (0-20 points)
    # Check if successor has worked with primary owner on this file
    local primary_owner=$(git log --since="$since_date" --format="%ae" -- "$file_path" 2>/dev/null | head -1)
    local collaboration_score=0

    # Find commits where both contributed within 7 days
    local successor_dates=$(git log --since="$since_date" --author="$successor_email" --format="%ad" --date=short -- "$file_path" 2>/dev/null)
    local primary_dates=$(git log --since="$since_date" --author="$primary_owner" --format="%ad" --date=short -- "$file_path" 2>/dev/null)

    local overlaps=0
    while IFS= read -r s_date; do
        if echo "$primary_dates" | grep -q "$s_date"; then
            ((overlaps++))
        fi
    done <<< "$successor_dates"

    if [[ $overlaps -gt 0 ]]; then
        collaboration_score=$(echo "scale=2; ($overlaps * 5)" | bc -l)
        if (( $(echo "$collaboration_score > 20" | bc -l) )); then
            collaboration_score="20"
        fi
    fi

    # Calculate total readiness score
    local total=$(echo "scale=2; $frequency_score + $recency_score + $familiarity_score + $collaboration_score" | bc -l)
    echo "${total%.*}"  # Return as integer
}

# Generate succession plan for repository
# Returns structured data with priority rankings
generate_succession_plan() {
    local repo_path="$1"
    local since_date="$2"
    local output_file="$3"

    cd "$repo_path" || return 1

    # Get all files in repository
    git ls-files | while read -r file; do
        # Identify successors for this file
        local succession_info=$(identify_successors "$repo_path" "$file" "$since_date")

        if [[ -n "$succession_info" ]]; then
            # Parse succession info
            IFS='|' read -ra parts <<< "$succession_info"
            local primary="${parts[0]}"
            local ownership_score="${parts[1]}"

            # Count successors
            local successor_count=$(( (${#parts[@]} - 2) / 2 ))

            # Determine priority
            local priority="Low"
            if [[ $successor_count -eq 0 ]]; then
                priority="Critical"
            elif [[ $successor_count -eq 1 ]]; then
                priority="High"
            elif [[ $(echo "$ownership_score > 80" | bc -l) -eq 1 ]]; then
                priority="Medium"
            fi

            # Output: file|primary_owner|ownership_score|successor_count|priority|successors...
            echo "$file|$primary|$ownership_score|$successor_count|$priority|${succession_info#*|*|}" >> "$output_file"
        fi
    done
}

# Identify high-risk areas (no successors available)
detect_succession_risks() {
    local repo_path="$1"
    local since_date="$2"
    local output_file="$3"

    cd "$repo_path" || return 1

    # Get all files
    git ls-files | while read -r file; do
        local succession_info=$(identify_successors "$repo_path" "$file" "$since_date" 5)

        if [[ -n "$succession_info" ]]; then
            IFS='|' read -ra parts <<< "$succession_info"
            local successor_count=$(( (${#parts[@]} - 2) / 2 ))

            # Flag files with no successors or single low-readiness successor
            if [[ $successor_count -eq 0 ]]; then
                local primary="${parts[0]}"
                local ownership_score="${parts[1]}"
                echo "$file|$primary|$ownership_score|no_successor|Critical" >> "$output_file"
            elif [[ $successor_count -eq 1 ]]; then
                local successor_readiness="${parts[3]}"
                if [[ $(echo "$successor_readiness < 40" | bc -l) -eq 1 ]]; then
                    local primary="${parts[0]}"
                    echo "$file|$primary|${parts[1]}|low_readiness|High" >> "$output_file"
                fi
            fi
        fi
    done
}

# Recommend mentorship pairings
# Identifies primary owners who should mentor potential successors
recommend_mentorships() {
    local repo_path="$1"
    local since_date="$2"
    local output_file="$3"

    cd "$repo_path" || return 1

    # Track mentor-mentee pairs with shared files
    declare -A mentorship_pairs
    declare -A pair_file_counts

    git ls-files | while read -r file; do
        local succession_info=$(identify_successors "$repo_path" "$file" "$since_date" 3)

        if [[ -n "$succession_info" ]]; then
            IFS='|' read -ra parts <<< "$succession_info"
            local primary="${parts[0]}"

            # For each potential successor
            for ((i=2; i<${#parts[@]}; i+=2)); do
                local successor="${parts[i]}"
                local readiness="${parts[i+1]}"

                # Only recommend if readiness is 20-60 (has potential but needs guidance)
                if [[ $(echo "$readiness >= 20 && $readiness <= 60" | bc -l) -eq 1 ]]; then
                    local pair_key="$primary|$successor"

                    # Track files where they should collaborate
                    if [[ -z "${mentorship_pairs[$pair_key]}" ]]; then
                        mentorship_pairs[$pair_key]="$file"
                        pair_file_counts[$pair_key]=1
                    else
                        mentorship_pairs[$pair_key]="${mentorship_pairs[$pair_key]},$file"
                        ((pair_file_counts[$pair_key]++))
                    fi
                fi
            done
        fi
    done

    # Output mentorship recommendations sorted by priority (file count)
    for pair in "${!pair_file_counts[@]}"; do
        IFS='|' read -r mentor mentee <<< "$pair"
        local file_count="${pair_file_counts[$pair]}"
        local files="${mentorship_pairs[$pair]}"

        echo "$mentor|$mentee|$file_count|$files"
    done | sort -t'|' -k3 -rn > "$output_file"
}

# Calculate succession coverage for repository
# Returns percentage of files with adequate succession planning
calculate_succession_coverage() {
    local repo_path="$1"
    local since_date="$2"
    local min_successors="${3:-1}"

    cd "$repo_path" || return 1

    local total_files=0
    local covered_files=0

    while read -r file; do
        ((total_files++))

        local succession_info=$(identify_successors "$repo_path" "$file" "$since_date" 3)
        if [[ -n "$succession_info" ]]; then
            IFS='|' read -ra parts <<< "$succession_info"
            local successor_count=$(( (${#parts[@]} - 2) / 2 ))

            if [[ $successor_count -ge $min_successors ]]; then
                ((covered_files++))
            fi
        fi
    done < <(git ls-files)

    if [[ $total_files -eq 0 ]]; then
        echo "0"
    else
        echo "scale=2; ($covered_files / $total_files) * 100" | bc -l
    fi
}

# Generate comprehensive succession report
generate_succession_report() {
    local repo_path="$1"
    local since_date="$2"
    local format="${3:-text}"

    local plan_file=$(mktemp)
    local risks_file=$(mktemp)
    local mentorships_file=$(mktemp)

    # Generate data
    generate_succession_plan "$repo_path" "$since_date" "$plan_file"
    detect_succession_risks "$repo_path" "$since_date" "$risks_file"
    recommend_mentorships "$repo_path" "$since_date" "$mentorships_file"

    # Calculate coverage
    local coverage=$(calculate_succession_coverage "$repo_path" "$since_date" 1)

    # Count risks
    local critical_risks=$(grep -c "Critical" "$risks_file" 2>/dev/null || echo "0")
    local high_risks=$(grep -c "High" "$risks_file" 2>/dev/null || echo "0")

    if [[ "$format" == "json" ]]; then
        # Build JSON arrays
        local plans_json=$(awk -F'|' '
        {
            if (NR>1) printf ","
            printf "{\"file\":\"%s\",\"primary_owner\":\"%s\",\"ownership_score\":%s,\"successor_count\":%d,\"priority\":\"%s\"}",
                $1, $2, $3, $4, $5
        }
        BEGIN { printf "[" }
        END { printf "]" }
        ' "$plan_file")

        local risks_json=$(awk -F'|' '
        {
            if (NR>1) printf ","
            printf "{\"file\":\"%s\",\"primary_owner\":\"%s\",\"ownership_score\":%s,\"risk_type\":\"%s\",\"priority\":\"%s\"}",
                $1, $2, $3, $4, $5
        }
        BEGIN { printf "[" }
        END { printf "]" }
        ' "$risks_file")

        local mentorships_json=$(awk -F'|' '
        {
            if (NR>1) printf ","
            split($4, files, ",")
            file_array = "["
            for (i in files) {
                if (i>1) file_array = file_array ","
                file_array = file_array "\"" files[i] "\""
            }
            file_array = file_array "]"
            printf "{\"mentor\":\"%s\",\"mentee\":\"%s\",\"shared_files\":%d,\"files\":%s}",
                $1, $2, $3, file_array
        }
        BEGIN { printf "[" }
        END { printf "]" }
        ' "$mentorships_file")

        jq -n \
            --arg coverage "$coverage" \
            --arg critical "$critical_risks" \
            --arg high "$high_risks" \
            --argjson plans "$plans_json" \
            --argjson risks "$risks_json" \
            --argjson mentorships "$mentorships_json" \
            '{
                succession_coverage: ($coverage | tonumber),
                risk_summary: {
                    critical_risks: ($critical | tonumber),
                    high_risks: ($high | tonumber),
                    total_risks: (($critical | tonumber) + ($high | tonumber))
                },
                succession_plans: $plans,
                succession_risks: $risks,
                mentorship_recommendations: $mentorships,
                recommendations: {
                    status: (
                        if ($critical | tonumber) > 0 then "Critical: Immediate succession planning needed"
                        elif ($high | tonumber) > 5 then "Warning: Multiple high-risk areas"
                        elif ($coverage | tonumber) < 50 then "Warning: Low succession coverage"
                        else "Good: Adequate succession planning"
                        end
                    )
                }
            }'
    else
        # Text format
        cat << EOF
========================================
Succession Planning Report
========================================

Repository: $repo_path
Analysis Date: $(date +%Y-%m-%d)

Summary:
--------
Succession Coverage: ${coverage}%
Critical Risks: $critical_risks files
High Risks: $high_risks files

Top Priority Files (No Successors):
----------------------------------
EOF
        if [[ -s "$risks_file" ]]; then
            grep "Critical" "$risks_file" | head -10 | awk -F'|' '{printf "%-50s %s\n", $1, $2}'
        else
            echo "None"
        fi

        cat << EOF

Mentorship Recommendations:
--------------------------
EOF
        if [[ -s "$mentorships_file" ]]; then
            head -10 "$mentorships_file" | awk -F'|' '{printf "%s → %s (%d shared files)\n", $1, $2, $3}'
        else
            echo "None"
        fi

        echo ""
        echo "========================================"
    fi

    # Cleanup
    rm -f "$plan_file" "$risks_file" "$mentorships_file"
}

# Export functions
export -f identify_successors
export -f calculate_successor_readiness
export -f generate_succession_plan
export -f detect_succession_risks
export -f recommend_mentorships
export -f calculate_succession_coverage
export -f generate_succession_report
