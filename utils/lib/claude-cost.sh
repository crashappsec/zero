#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Claude API Cost Tracking Library
# Tracks Claude API usage and calculates costs
#############################################################################

# Cost tracking file (session-based)
COST_TRACKING_FILE="/tmp/claude_cost_$$.tmp"

# Claude pricing (per million tokens) - Updated as of 2024
# https://www.anthropic.com/api
declare -A INPUT_COSTS=(
    ["claude-3-opus-20240229"]=15.00
    ["claude-3-sonnet-20240229"]=3.00
    ["claude-3-haiku-20240307"]=0.25
    ["claude-3-5-sonnet-20241022"]=3.00
    ["claude-sonnet-4-5-20250929"]=3.00
)

declare -A OUTPUT_COSTS=(
    ["claude-3-opus-20240229"]=75.00
    ["claude-3-sonnet-20240229"]=15.00
    ["claude-3-haiku-20240307"]=1.25
    ["claude-3-5-sonnet-20241022"]=15.00
    ["claude-sonnet-4-5-20250929"]=15.00
)

# Initialize cost tracking
init_cost_tracking() {
    echo "timestamp,model,input_tokens,output_tokens,input_cost,output_cost,total_cost" > "$COST_TRACKING_FILE"
}

# Clean up cost tracking file
cleanup_cost_tracking() {
    rm -f "$COST_TRACKING_FILE"
}

# Calculate cost for token usage
calculate_cost() {
    local model="$1"
    local input_tokens="$2"
    local output_tokens="$3"

    # Get pricing for model (default to Sonnet if not found)
    local input_price_per_million="${INPUT_COSTS[$model]:-3.00}"
    local output_price_per_million="${OUTPUT_COSTS[$model]:-15.00}"

    # Calculate costs
    local input_cost=$(echo "scale=6; ($input_tokens / 1000000) * $input_price_per_million" | bc -l)
    local output_cost=$(echo "scale=6; ($output_tokens / 1000000) * $output_price_per_million" | bc -l)
    local total_cost=$(echo "scale=6; $input_cost + $output_cost" | bc -l)

    echo "$input_cost|$output_cost|$total_cost"
}

# Record API usage from Claude response
# Usage: record_api_usage "$response" "$model"
record_api_usage() {
    local response="$1"
    local model="$2"

    # Extract token usage from response
    local input_tokens=$(echo "$response" | jq -r '.usage.input_tokens // 0' 2>/dev/null)
    local output_tokens=$(echo "$response" | jq -r '.usage.output_tokens // 0' 2>/dev/null)

    # Skip if no usage data
    if [[ "$input_tokens" == "0" && "$output_tokens" == "0" ]]; then
        return 0
    fi

    # Calculate costs
    local costs=$(calculate_cost "$model" "$input_tokens" "$output_tokens")
    IFS='|' read -r input_cost output_cost total_cost <<< "$costs"

    # Record to tracking file
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    echo "$timestamp,$model,$input_tokens,$output_tokens,$input_cost,$output_cost,$total_cost" >> "$COST_TRACKING_FILE"

    # Return total cost for immediate use
    echo "$total_cost"
}

# Get total cost for current session
get_session_cost() {
    if [[ ! -f "$COST_TRACKING_FILE" ]]; then
        echo "0.000000"
        return
    fi

    # Sum total costs (skip header)
    awk -F',' 'NR>1 {sum+=$7} END {printf "%.6f", sum}' "$COST_TRACKING_FILE"
}

# Get session statistics
get_session_stats() {
    if [[ ! -f "$COST_TRACKING_FILE" ]]; then
        echo "No API calls recorded"
        return
    fi

    local total_calls=$(tail -n +2 "$COST_TRACKING_FILE" | wc -l | tr -d ' ')
    local total_input_tokens=$(awk -F',' 'NR>1 {sum+=$3} END {print sum+0}' "$COST_TRACKING_FILE")
    local total_output_tokens=$(awk -F',' 'NR>1 {sum+=$4} END {print sum+0}' "$COST_TRACKING_FILE")
    local total_cost=$(get_session_cost)

    cat << EOF
API Usage Statistics:
  Total API Calls: $total_calls
  Input Tokens: $(printf "%'d" $total_input_tokens)
  Output Tokens: $(printf "%'d" $total_output_tokens)
  Total Cost: \$$total_cost
EOF
}

# Display cost summary (called at end of analysis)
display_cost_summary() {
    if [[ ! -f "$COST_TRACKING_FILE" ]]; then
        return
    fi

    echo ""
    echo "========================================="
    echo "  Claude API Usage Summary"
    echo "========================================="
    get_session_stats
    echo "========================================="
}

# Export cost data to JSON
export_cost_data_json() {
    local output_file="$1"

    if [[ ! -f "$COST_TRACKING_FILE" ]]; then
        echo '{"calls":[],"summary":{"total_calls":0,"total_input_tokens":0,"total_output_tokens":0,"total_cost":0}}' > "$output_file"
        return
    fi

    # Build JSON array of calls
    local calls_json=$(tail -n +2 "$COST_TRACKING_FILE" | awk -F',' '{
        printf "{\"timestamp\":\"%s\",\"model\":\"%s\",\"input_tokens\":%s,\"output_tokens\":%s,\"input_cost\":%s,\"output_cost\":%s,\"total_cost\":%s},",
        $1, $2, $3, $4, $5, $6, $7
    }' | sed 's/,$//')

    # Calculate summary
    local total_calls=$(tail -n +2 "$COST_TRACKING_FILE" | wc -l | tr -d ' ')
    local total_input_tokens=$(awk -F',' 'NR>1 {sum+=$3} END {print sum+0}' "$COST_TRACKING_FILE")
    local total_output_tokens=$(awk -F',' 'NR>1 {sum+=$4} END {print sum+0}' "$COST_TRACKING_FILE")
    local total_cost=$(get_session_cost)

    # Build complete JSON
    cat > "$output_file" << EOF
{
  "calls": [$calls_json],
  "summary": {
    "total_calls": $total_calls,
    "total_input_tokens": $total_input_tokens,
    "total_output_tokens": $total_output_tokens,
    "total_cost": $total_cost
  }
}
EOF
}

# Export functions
export -f init_cost_tracking
export -f cleanup_cost_tracking
export -f calculate_cost
export -f record_api_usage
export -f get_session_cost
export -f get_session_stats
export -f display_cost_summary
export -f export_cost_data_json
