#!/bin/bash
# Copyright (c) 2024 Crash Override Inc
# 101 Fulton St, 416, New York 10038
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Configuration System Library
# Hierarchical configuration with environment, global, and local overrides
#
# Configuration Priority (highest to lowest):
# 1. Command-line arguments
# 2. Environment variables (CODE_OWNERSHIP_*)
# 3. Local config (.code-ownership.conf in repo)
# 4. Global config (~/.config/code-ownership/config)
# 5. System config (/etc/code-ownership/config)
# 6. Built-in defaults
#############################################################################

# Default configuration values
declare -A CONFIG_DEFAULTS=(
    [analysis_method]="hybrid"
    [analysis_days]="90"
    [output_format]="json"
    [staleness_threshold_days]="90"
    [spof_score_threshold]="2"
    [health_score_weights_coverage]="0.35"
    [health_score_weights_distribution]="0.25"
    [health_score_weights_freshness]="0.20"
    [health_score_weights_engagement]="0.20"
    [bus_factor_threshold]="3"
    [coverage_target]="90"
    [gini_excellent]="0.3"
    [gini_good]="0.5"
    [ownership_recency_half_life]="90"
    [ownership_commit_weight]="0.6"
    [ownership_line_weight]="0.4"
    [github_api_enabled]="true"
    [github_cache_ttl]="3600"
    [codeowners_path]=".github/CODEOWNERS"
    [validate_codeowners]="false"
    [include_github_profiles]="true"
    [max_spof_display]="20"
    [verbose]="false"
)

# Active configuration (populated during load)
declare -A CONFIG

# Configuration file locations
SYSTEM_CONFIG="/etc/code-ownership/config"
GLOBAL_CONFIG="$HOME/.config/code-ownership/config"
LOCAL_CONFIG=".code-ownership.conf"

# Initialize configuration system
init_config() {
    # Start with defaults
    for key in "${!CONFIG_DEFAULTS[@]}"; do
        CONFIG[$key]="${CONFIG_DEFAULTS[$key]}"
    done

    # Load configurations in priority order (lowest to highest)
    load_config_file "$SYSTEM_CONFIG"
    load_config_file "$GLOBAL_CONFIG"

    # Local config is loaded per-repository
}

# Load configuration from file
load_config_file() {
    local config_file="$1"

    if [[ ! -f "$config_file" ]]; then
        return 0
    fi

    while IFS='=' read -r key value; do
        # Skip comments and empty lines
        [[ "$key" =~ ^#.*$ ]] && continue
        [[ -z "$key" ]] && continue

        # Remove leading/trailing whitespace
        key=$(echo "$key" | xargs)
        value=$(echo "$value" | xargs)

        # Remove quotes from value
        value="${value%\"}"
        value="${value#\"}"
        value="${value%\'}"
        value="${value#\'}"

        # Store in config
        CONFIG[$key]="$value"
    done < "$config_file"
}

# Load local repository configuration
load_local_config() {
    local repo_path="$1"

    if [[ -f "$repo_path/$LOCAL_CONFIG" ]]; then
        load_config_file "$repo_path/$LOCAL_CONFIG"
    fi
}

# Load configuration from environment variables
# Environment variables use CODE_OWNERSHIP_ prefix
load_env_config() {
    for key in "${!CONFIG_DEFAULTS[@]}"; do
        # Convert config key to env var name
        # Example: analysis_method -> CODE_OWNERSHIP_ANALYSIS_METHOD
        local env_var="CODE_OWNERSHIP_${key^^}"

        if [[ -n "${!env_var}" ]]; then
            CONFIG[$key]="${!env_var}"
        fi
    done
}

# Get configuration value
get_config() {
    local key="$1"
    local default="${2:-}"

    if [[ -n "${CONFIG[$key]}" ]]; then
        echo "${CONFIG[$key]}"
    else
        echo "$default"
    fi
}

# Set configuration value (runtime override)
set_config() {
    local key="$1"
    local value="$2"

    CONFIG[$key]="$value"
}

# Validate configuration values
validate_config() {
    local errors=0

    # Validate analysis_method
    local method="${CONFIG[analysis_method]}"
    if [[ ! "$method" =~ ^(commit|line|hybrid)$ ]]; then
        echo "Error: Invalid analysis_method '$method'. Must be: commit, line, or hybrid" >&2
        ((errors++))
    fi

    # Validate analysis_days
    local days="${CONFIG[analysis_days]}"
    if ! [[ "$days" =~ ^[0-9]+$ ]] || [[ $days -lt 1 ]]; then
        echo "Error: Invalid analysis_days '$days'. Must be positive integer" >&2
        ((errors++))
    fi

    # Validate output_format
    local format="${CONFIG[output_format]}"
    if [[ ! "$format" =~ ^(json|text|markdown)$ ]]; then
        echo "Error: Invalid output_format '$format'. Must be: json, text, or markdown" >&2
        ((errors++))
    fi

    # Validate weights sum to 1.0
    local weight_sum=$(echo "scale=2; ${CONFIG[health_score_weights_coverage]} + ${CONFIG[health_score_weights_distribution]} + ${CONFIG[health_score_weights_freshness]} + ${CONFIG[health_score_weights_engagement]}" | bc -l)
    if [[ $(echo "$weight_sum != 1.0" | bc -l) -eq 1 ]]; then
        echo "Warning: Health score weights sum to $weight_sum, not 1.0" >&2
    fi

    # Validate percentage values (0-1)
    for key in health_score_weights_coverage health_score_weights_distribution health_score_weights_freshness health_score_weights_engagement ownership_commit_weight ownership_line_weight; do
        local value="${CONFIG[$key]}"
        if [[ $(echo "$value < 0 || $value > 1" | bc -l) -eq 1 ]]; then
            echo "Error: $key must be between 0 and 1, got $value" >&2
            ((errors++))
        fi
    done

    return $errors
}

# Generate default configuration file
generate_default_config() {
    local output_file="${1:-config}"

    cat > "$output_file" << 'EOF'
# Code Ownership Analyzer Configuration
# This file uses simple key=value format
# Lines starting with # are comments

# Analysis Settings
analysis_method=hybrid           # commit, line, or hybrid
analysis_days=90                 # Number of days to analyze
output_format=json               # json, text, or markdown

# Ownership Thresholds
staleness_threshold_days=90      # Days before owner considered stale
spof_score_threshold=2           # Minimum score to flag as SPOF (0-6)
bus_factor_threshold=3           # Target minimum bus factor
coverage_target=90               # Target ownership coverage percentage

# Health Score Weights (must sum to 1.0)
health_score_weights_coverage=0.35
health_score_weights_distribution=0.25
health_score_weights_freshness=0.20
health_score_weights_engagement=0.20

# Distribution Metrics
gini_excellent=0.3               # Gini coefficient for excellent distribution
gini_good=0.5                    # Gini coefficient for good distribution

# Ownership Calculation
ownership_recency_half_life=90   # Days for recency decay half-life
ownership_commit_weight=0.6      # Weight for commit-based ownership
ownership_line_weight=0.4        # Weight for line-based ownership

# GitHub Integration
github_api_enabled=true          # Enable GitHub API lookups
github_cache_ttl=3600            # Cache TTL in seconds
include_github_profiles=true     # Include GitHub profile links

# CODEOWNERS
codeowners_path=.github/CODEOWNERS  # Path to CODEOWNERS file
validate_codeowners=false        # Validate CODEOWNERS by default

# Display Options
max_spof_display=20              # Maximum SPOFs to display
verbose=false                    # Enable verbose output

EOF
}

# Export configuration as environment variables
# Useful for passing to child processes
export_config() {
    for key in "${!CONFIG[@]}"; do
        local env_var="CODE_OWNERSHIP_${key^^}"
        export "$env_var=${CONFIG[$key]}"
    done
}

# Print current configuration
print_config() {
    local format="${1:-text}"

    if [[ "$format" == "json" ]]; then
        echo "{"
        local first=true
        for key in "${!CONFIG[@]}"; do
            if [[ "$first" != "true" ]]; then
                echo ","
            fi
            first=false
            printf "  \"%s\": \"%s\"" "$key" "${CONFIG[$key]}"
        done
        echo ""
        echo "}"
    else
        echo "Current Configuration:"
        echo "===================="
        for key in "${!CONFIG[@]}" | sort; do
            printf "%-40s = %s\n" "$key" "${CONFIG[$key]}"
        done
    fi
}

# Save current configuration to file
save_config() {
    local output_file="$1"
    local config_dir=$(dirname "$output_file")

    # Create directory if it doesn't exist
    mkdir -p "$config_dir"

    cat > "$output_file" << EOF
# Code Ownership Analyzer Configuration
# Generated: $(date -u +"%Y-%m-%d %H:%M:%S UTC")

EOF

    for key in "${!CONFIG[@]}" | sort; do
        echo "$key=${CONFIG[$key]}"
    done >> "$output_file"
}

# Get config value with type conversion
get_config_int() {
    local value=$(get_config "$1" "${2:-0}")
    echo "${value%.*}"  # Remove decimal if present
}

get_config_float() {
    get_config "$1" "${2:-0.0}"
}

get_config_bool() {
    local value=$(get_config "$1" "${2:-false}")
    case "${value,,}" in
        true|yes|1|on)
            echo "true"
            ;;
        *)
            echo "false"
            ;;
    esac
}

# Export functions
export -f init_config
export -f load_config_file
export -f load_local_config
export -f load_env_config
export -f get_config
export -f set_config
export -f validate_config
export -f generate_default_config
export -f export_config
export -f print_config
export -f save_config
export -f get_config_int
export -f get_config_float
export -f get_config_bool
