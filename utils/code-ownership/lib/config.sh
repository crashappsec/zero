#!/bin/bash
# Copyright (c) 2024 Gibson Powers Contributors
# 
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Configuration System Library
# Hierarchical configuration with environment, global, and local overrides
# Compatible with bash 3.2+ (macOS default)
#
# Configuration Priority (highest to lowest):
# 1. Command-line arguments
# 2. Environment variables (CODE_OWNERSHIP_*)
# 3. Local config (.code-ownership.conf in repo)
# 4. Global config (~/.config/code-ownership/config)
# 5. System config (/etc/code-ownership/config)
# 6. Built-in defaults
#############################################################################

# Configuration storage (bash 3.2 compatible - uses temp file)
CONFIG_CACHE_FILE="${CONFIG_CACHE_FILE:-/tmp/code_ownership_config_$$.tmp}"

# Configuration file locations
SYSTEM_CONFIG="/etc/code-ownership/config"
GLOBAL_CONFIG="$HOME/.config/code-ownership/config"
LOCAL_CONFIG=".code-ownership.conf"

# Initialize configuration with defaults
init_config() {
    # Create config cache file with defaults
    cat > "$CONFIG_CACHE_FILE" << 'EOF'
analysis_method=hybrid
analysis_days=90
output_format=markdown
staleness_threshold_days=90
spof_score_threshold=2
health_score_weights_coverage=0.35
health_score_weights_distribution=0.25
health_score_weights_freshness=0.20
health_score_weights_engagement=0.20
bus_factor_threshold=3
coverage_target=90
gini_excellent=0.3
gini_good=0.5
ownership_recency_half_life=90
ownership_commit_weight=0.6
ownership_line_weight=0.4
github_api_enabled=true
github_cache_ttl=3600
codeowners_path=.github/CODEOWNERS
validate_codeowners=false
include_github_profiles=true
max_spof_display=20
verbose=false
EOF

    # Load configurations in priority order (lowest to highest)
    load_config_file "$SYSTEM_CONFIG"
    load_config_file "$GLOBAL_CONFIG"
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
        key=$(echo "$key" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
        value=$(echo "$value" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')

        # Remove quotes from value
        value=$(echo "$value" | sed 's/^["'"'"']//;s/["'"'"']$//')

        # Update config cache
        set_config "$key" "$value"
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
    # Read all keys from config cache
    while IFS='=' read -r key value; do
        [[ -z "$key" ]] && continue

        # Convert config key to env var name (bash 3.2 compatible)
        local env_var="CODE_OWNERSHIP_$(echo "$key" | tr '[:lower:]' '[:upper:]')"

        # Check if environment variable is set
        if [[ -n "${!env_var:-}" ]]; then
            set_config "$key" "${!env_var}"
        fi
    done < "$CONFIG_CACHE_FILE"
}

# Get configuration value
get_config() {
    local key="$1"
    local default="${2:-}"

    if [[ ! -f "$CONFIG_CACHE_FILE" ]]; then
        echo "$default"
        return 1
    fi

    local value=$(grep "^${key}=" "$CONFIG_CACHE_FILE" 2>/dev/null | cut -d'=' -f2-)

    if [[ -n "$value" ]]; then
        echo "$value"
    else
        echo "$default"
    fi
}

# Set configuration value (runtime override)
set_config() {
    local key="$1"
    local value="$2"

    if [[ ! -f "$CONFIG_CACHE_FILE" ]]; then
        echo "${key}=${value}" > "$CONFIG_CACHE_FILE"
        return
    fi

    # Check if key exists
    if grep -q "^${key}=" "$CONFIG_CACHE_FILE" 2>/dev/null; then
        # Update existing key (macOS sed compatible)
        sed -i '' "s|^${key}=.*|${key}=${value}|" "$CONFIG_CACHE_FILE" 2>/dev/null || \
        sed -i "s|^${key}=.*|${key}=${value}|" "$CONFIG_CACHE_FILE"
    else
        # Add new key
        echo "${key}=${value}" >> "$CONFIG_CACHE_FILE"
    fi
}

# Validate configuration values
validate_config() {
    local errors=0

    # Validate analysis_method
    local method=$(get_config "analysis_method")
    if [[ ! "$method" =~ ^(commit|line|hybrid)$ ]]; then
        echo "Error: Invalid analysis_method '$method'. Must be: commit, line, or hybrid" >&2
        ((errors++))
    fi

    # Validate analysis_days
    local days=$(get_config "analysis_days")
    if ! [[ "$days" =~ ^[0-9]+$ ]] || [[ $days -lt 1 ]]; then
        echo "Error: Invalid analysis_days '$days'. Must be positive integer" >&2
        ((errors++))
    fi

    # Validate output_format
    local format=$(get_config "output_format")
    if [[ ! "$format" =~ ^(json|text|markdown)$ ]]; then
        echo "Error: Invalid output_format '$format'. Must be: json, text, or markdown" >&2
        ((errors++))
    fi

    # Validate weights sum to 1.0
    local coverage=$(get_config "health_score_weights_coverage")
    local distribution=$(get_config "health_score_weights_distribution")
    local freshness=$(get_config "health_score_weights_freshness")
    local engagement=$(get_config "health_score_weights_engagement")
    local weight_sum=$(echo "scale=2; $coverage + $distribution + $freshness + $engagement" | bc -l)

    if [[ $(echo "$weight_sum != 1.0" | bc -l) -eq 1 ]]; then
        echo "Warning: Health score weights sum to $weight_sum, not 1.0" >&2
    fi

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
export_config() {
    if [[ ! -f "$CONFIG_CACHE_FILE" ]]; then
        return 1
    fi

    while IFS='=' read -r key value; do
        [[ -z "$key" ]] && continue
        local env_var="CODE_OWNERSHIP_$(echo "$key" | tr '[:lower:]' '[:upper:]')"
        export "$env_var=$value"
    done < "$CONFIG_CACHE_FILE"
}

# Print current configuration
print_config() {
    local format="${1:-text}"

    if [[ ! -f "$CONFIG_CACHE_FILE" ]]; then
        echo "Config not initialized"
        return 1
    fi

    if [[ "$format" == "json" ]]; then
        echo "{"
        local first=true
        while IFS='=' read -r key value; do
            [[ -z "$key" ]] && continue
            if [[ "$first" != "true" ]]; then
                echo ","
            fi
            first=false
            printf "  \"%s\": \"%s\"" "$key" "$value"
        done < "$CONFIG_CACHE_FILE"
        echo ""
        echo "}"
    else
        echo "Current Configuration:"
        echo "===================="
        sort "$CONFIG_CACHE_FILE" | while IFS='=' read -r key value; do
            [[ -z "$key" ]] && continue
            printf "%-40s = %s\n" "$key" "$value"
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

    if [[ -f "$CONFIG_CACHE_FILE" ]]; then
        sort "$CONFIG_CACHE_FILE" >> "$output_file"
    fi
}

# Cleanup function (call on exit)
cleanup_config() {
    rm -f "$CONFIG_CACHE_FILE"
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
    case "${value}" in
        true|TRUE|yes|YES|1|on|ON)
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
export -f cleanup_config
export -f get_config_int
export -f get_config_float
export -f get_config_bool
