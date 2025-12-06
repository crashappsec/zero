#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# 
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Config Loader Library
# Hierarchical configuration loading with global and module-specific configs
# Usage: source this file and call load_config <module_name>
#############################################################################

# Find utils root directory
UTILS_ROOT=""
CONFIG_DIR=""
find_utils_root() {
    local current_dir="$1"
    while [[ "$current_dir" != "/" ]]; do
        # Check if we're inside utils/zero/config
        if [[ -d "$current_dir/zero/config" ]] && [[ -f "$current_dir/zero/config/config.example.json" ]]; then
            UTILS_ROOT="$current_dir"
            CONFIG_DIR="$current_dir/zero/config"
            return 0
        fi
        # Check if current dir is the config dir itself
        if [[ "$(basename "$current_dir")" == "config" ]] && [[ -f "$current_dir/config.example.json" ]]; then
            CONFIG_DIR="$current_dir"
            UTILS_ROOT="$(dirname "$(dirname "$current_dir")")"
            return 0
        fi
        current_dir="$(dirname "$current_dir")"
    done
    return 1
}

# Initialize utils root
if ! find_utils_root "$(pwd)"; then
    if ! find_utils_root "$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"; then
        echo "Error: Cannot find utils root directory" >&2
        return 1 2>/dev/null || exit 1
    fi
fi

# Config paths (now in zero/config/)
GLOBAL_CONFIG="${CONFIG_DIR}/config.json"
GLOBAL_CONFIG_EXAMPLE="${CONFIG_DIR}/config.example.json"

# Load configuration with hierarchy
# Usage: load_config <module_name> [module_config_path]
load_config() {
    local module_name="$1"
    local module_config_path="$2"

    # Check if jq is available
    if ! command -v jq &> /dev/null; then
        echo "Error: jq is required for config loading" >&2
        return 1
    fi

    # Create global config from example if it doesn't exist
    if [[ ! -f "$GLOBAL_CONFIG" ]] && [[ -f "$GLOBAL_CONFIG_EXAMPLE" ]]; then
        cp "$GLOBAL_CONFIG_EXAMPLE" "$GLOBAL_CONFIG"
    fi

    # If no global config, return error
    if [[ ! -f "$GLOBAL_CONFIG" ]]; then
        return 1
    fi

    # Check if module configs should be ignored
    local ignore_module_configs=$(jq -r '.config_behavior.ignore_module_configs // false' "$GLOBAL_CONFIG" 2>/dev/null)

    # Load global config
    local config_data=$(cat "$GLOBAL_CONFIG")

    # If module config exists and not ignored, merge it
    if [[ "$ignore_module_configs" != "true" ]] && [[ -n "$module_config_path" ]] && [[ -f "$module_config_path" ]]; then
        local module_config_data=$(cat "$module_config_path")
        # Merge configs - module config overrides global
        config_data=$(jq -s '.[0] * .[1]' <(echo "$config_data") <(echo "$module_config_data") 2>/dev/null)
    fi

    # Export config as JSON string for easy access
    export CONFIG_JSON="$config_data"
    export CONFIG_MODULE="$module_name"

    return 0
}

# Get config value
# Usage: get_config <json_path>
# Example: get_config '.github.organizations[]'
get_config() {
    local path="$1"
    if [[ -z "$CONFIG_JSON" ]]; then
        echo "Error: Config not loaded. Call load_config first." >&2
        return 1
    fi
    echo "$CONFIG_JSON" | jq -r "$path" 2>/dev/null
}

# Get config value with default
# Usage: get_config_default <json_path> <default_value>
get_config_default() {
    local path="$1"
    local default="$2"
    local value=$(get_config "$path")
    if [[ -z "$value" ]] || [[ "$value" == "null" ]]; then
        echo "$default"
    else
        echo "$value"
    fi
}

# Get module-specific config
# Usage: get_module_config <key>
# Example: get_module_config 'default_modules[]'
get_module_config() {
    local key="$1"
    if [[ -z "$CONFIG_MODULE" ]]; then
        echo "Error: Module name not set" >&2
        return 1
    fi
    # Use module name as-is (don't convert hyphens to underscores)
    get_config ".modules.${CONFIG_MODULE}.${key}"
}

# Check if module is enabled
# Usage: is_module_enabled
is_module_enabled() {
    if [[ -z "$CONFIG_MODULE" ]]; then
        return 0  # Default to enabled if no module specified
    fi
    local enabled=$(get_module_config 'enabled' 2>/dev/null)
    [[ "$enabled" == "true" ]]
}

# Get GitHub organizations from config
# Usage: get_organizations
get_organizations() {
    get_config '.github.organizations[]?' | grep -v '^$' || true
}

# Get GitHub repositories from config
# Usage: get_repositories
get_repositories() {
    get_config '.github.repositories[]?' | grep -v '^$' || true
}

# Get default modules for current module
# Usage: get_default_modules
get_default_modules() {
    get_module_config 'default_modules[]?' | grep -v '^$' || true
}

# Load GitHub token from config if not already set
# Usage: load_github_token
# Sets GITHUB_TOKEN environment variable from config.json if:
# 1. GITHUB_TOKEN is not already set
# 2. config.json exists and has github.pat value
load_github_token() {
    # If GITHUB_TOKEN already set, don't override
    if [[ -n "${GITHUB_TOKEN:-}" ]]; then
        return 0
    fi

    # Try to load from config
    if [[ -f "$GLOBAL_CONFIG" ]]; then
        local github_pat=$(jq -r '.github.pat // ""' "$GLOBAL_CONFIG" 2>/dev/null)
        if [[ -n "$github_pat" ]] && [[ "$github_pat" != "null" ]]; then
            export GITHUB_TOKEN="$github_pat"
            return 0
        fi
    fi

    return 1
}

#############################################################################
# Profile Functions
# Dynamically load scan profiles from zero.config.json
#############################################################################

PHANTOM_CONFIG="${CONFIG_DIR}/zero.config.json"

# Get list of available profile names
# Usage: get_available_profiles
get_available_profiles() {
    if [[ ! -f "$PHANTOM_CONFIG" ]]; then
        echo "quick standard advanced deep security" # fallback
        return
    fi
    jq -r '.profiles | keys[]' "$PHANTOM_CONFIG" 2>/dev/null | tr '\n' ' '
}

# Get scanners for a profile
# Usage: get_profile_scanners <profile_name>
get_profile_scanners() {
    local profile="$1"
    if [[ ! -f "$PHANTOM_CONFIG" ]]; then
        # Fallback hardcoded profiles
        case "$profile" in
            quick)    echo "package-sbom tech-discovery package-vulns licenses" ;;
            standard) echo "package-sbom tech-discovery package-vulns licenses code-security code-secrets tech-debt code-ownership dora" ;;
            advanced) echo "package-sbom tech-discovery package-vulns package-health licenses code-security iac-security code-secrets tech-debt documentation git test-coverage code-ownership dora package-provenance" ;;
            deep)     echo "package-sbom tech-discovery package-vulns package-health licenses code-security iac-security code-secrets tech-debt documentation git test-coverage code-ownership dora package-provenance" ;;
            security) echo "package-sbom package-vulns licenses code-security iac-security code-secrets" ;;
            *)        echo "package-sbom tech-discovery package-vulns licenses code-security code-secrets tech-debt code-ownership dora" ;;
        esac
        return
    fi
    jq -r --arg p "$profile" '.profiles[$p].scanners // [] | .[]' "$PHANTOM_CONFIG" 2>/dev/null | tr '\n' ' '
}

# Get profile metadata
# Usage: get_profile_name <profile_name>
get_profile_name() {
    local profile="$1"
    if [[ ! -f "$PHANTOM_CONFIG" ]]; then
        echo "${profile^}"
        return
    fi
    jq -r --arg p "$profile" '.profiles[$p].name // $p' "$PHANTOM_CONFIG" 2>/dev/null
}

# Get profile description
# Usage: get_profile_description <profile_name>
get_profile_description() {
    local profile="$1"
    if [[ ! -f "$PHANTOM_CONFIG" ]]; then
        echo ""
        return
    fi
    jq -r --arg p "$profile" '.profiles[$p].description // ""' "$PHANTOM_CONFIG" 2>/dev/null
}

# Get profile estimated time
# Usage: get_profile_estimated_time <profile_name>
get_profile_estimated_time() {
    local profile="$1"
    if [[ ! -f "$PHANTOM_CONFIG" ]]; then
        echo "~2 minutes"
        return
    fi
    jq -r --arg p "$profile" '.profiles[$p].estimated_time // "~2 minutes"' "$PHANTOM_CONFIG" 2>/dev/null
}

# Check if profile uses Claude
# Usage: profile_uses_claude <profile_name>
profile_uses_claude() {
    local profile="$1"
    if [[ ! -f "$PHANTOM_CONFIG" ]]; then
        [[ "$profile" == "deep" ]]
        return
    fi
    local mode=$(jq -r --arg p "$profile" '.profiles[$p].claude_mode // "none"' "$PHANTOM_CONFIG" 2>/dev/null)
    [[ "$mode" == "enabled" || "$mode" == "required" ]]
}

# Get scanners that should use Claude for a profile
# Usage: get_profile_claude_scanners <profile_name>
get_profile_claude_scanners() {
    local profile="$1"
    if [[ ! -f "$PHANTOM_CONFIG" ]]; then
        echo ""
        return
    fi
    jq -r --arg p "$profile" '.profiles[$p].claude_scanners // [] | .[]' "$PHANTOM_CONFIG" 2>/dev/null | tr '\n' ' '
}

# Get default profile from config
# Usage: get_default_profile
get_default_profile() {
    if [[ ! -f "$PHANTOM_CONFIG" ]]; then
        echo "standard"
        return
    fi
    jq -r '.settings.default_profile // "standard"' "$PHANTOM_CONFIG" 2>/dev/null
}

# Get all scanner names from config
# Usage: get_all_scanners
get_all_scanners() {
    if [[ ! -f "$PHANTOM_CONFIG" ]]; then
        echo "package-sbom tech-discovery package-vulns package-health licenses code-security iac-security code-secrets tech-debt documentation git test-coverage code-ownership dora package-provenance"
        return
    fi
    jq -r '.scanners | keys[]' "$PHANTOM_CONFIG" 2>/dev/null | tr '\n' ' '
}

# Get scanner display name
# Usage: get_scanner_name <scanner_id>
get_scanner_name() {
    local scanner="$1"
    if [[ ! -f "$PHANTOM_CONFIG" ]]; then
        echo "$scanner"
        return
    fi
    jq -r --arg s "$scanner" '.scanners[$s].name // $s' "$PHANTOM_CONFIG" 2>/dev/null
}

# Get scanner description
# Usage: get_scanner_description <scanner_id>
get_scanner_description() {
    local scanner="$1"
    if [[ ! -f "$PHANTOM_CONFIG" ]]; then
        echo ""
        return
    fi
    jq -r --arg s "$scanner" '.scanners[$s].description // ""' "$PHANTOM_CONFIG" 2>/dev/null
}

# Get scanner script path
# Usage: get_scanner_script <scanner_id>
get_scanner_script() {
    local scanner="$1"
    if [[ ! -f "$PHANTOM_CONFIG" ]]; then
        echo "utils/scanners/${scanner}/${scanner}.sh"
        return
    fi
    jq -r --arg s "$scanner" '.scanners[$s].script // "utils/scanners/\($s)/\($s).sh"' "$PHANTOM_CONFIG" 2>/dev/null
}

# Get scanner output filename
# Usage: get_scanner_output_file <scanner_id>
get_scanner_output_file() {
    local scanner="$1"
    if [[ ! -f "$PHANTOM_CONFIG" ]]; then
        echo "${scanner}.json"
        return
    fi
    jq -r --arg s "$scanner" '.scanners[$s].output_file // "\($s).json"' "$PHANTOM_CONFIG" 2>/dev/null
}

# Check if scanner in profile
# Usage: scanner_in_profile <scanner_id> <profile_name>
scanner_in_profile() {
    local scanner="$1"
    local profile="$2"
    local profile_scanners=$(get_profile_scanners "$profile")
    [[ " $profile_scanners " =~ " $scanner " ]]
}

# Print profile list for help
# Usage: print_profile_help
print_profile_help() {
    local profiles=$(get_available_profiles)
    for profile in $profiles; do
        local name=$(get_profile_name "$profile")
        local desc=$(get_profile_description "$profile")
        local time=$(get_profile_estimated_time "$profile")
        printf "    --%-12s %s (%s)\n" "$profile" "$desc" "$time"
    done
}

# Export functions for use in other scripts
export -f load_config
export -f get_config
export -f get_config_default
export -f get_module_config
export -f is_module_enabled
export -f get_organizations
export -f get_repositories
export -f get_default_modules
export -f load_github_token
export -f get_available_profiles
export -f get_profile_scanners
export -f get_profile_name
export -f get_profile_description
export -f get_profile_estimated_time
export -f profile_uses_claude
export -f get_profile_claude_scanners
export -f get_default_profile
export -f get_all_scanners
export -f get_scanner_name
export -f get_scanner_description
export -f get_scanner_script
export -f get_scanner_output_file
export -f scanner_in_profile
export -f print_profile_help
export PHANTOM_CONFIG
