#!/bin/bash
# Library Recommendation Engine
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0
#
# Recommends modern, secure library alternatives for outdated or deprecated packages.
# Part of the Developer Productivity module.

set -eo pipefail

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCANNERS_ROOT="$(dirname "$SCRIPT_DIR")"

# Load deps.dev client from shared libs (if not already loaded)
if ! command -v deps_dev_get_package_info &> /dev/null; then
    source "$UTILS_ROOT/lib/deps-dev-client.sh"
fi

# Cache for live package data
declare -A LIVE_PACKAGE_CACHE
LIVE_DATA_ENABLED=true

#############################################################################
# Live Package Data Fetching - Real-time Registry Queries
#############################################################################

# Fetch npm package info (deprecated status, downloads, latest version)
# Usage: fetch_npm_package_info <package>
fetch_npm_package_info() {
    local pkg="$1"
    local cache_key="npm:$pkg"

    # Check cache
    if [[ -n "${LIVE_PACKAGE_CACHE[$cache_key]:-}" ]]; then
        echo "${LIVE_PACKAGE_CACHE[$cache_key]}"
        return
    fi

    # URL encode package name for scoped packages
    local encoded_pkg="${pkg//@/%40}"
    encoded_pkg="${encoded_pkg//\//%2F}"

    local response=$(curl -s "https://registry.npmjs.org/${encoded_pkg}" 2>/dev/null)

    if [[ -z "$response" ]] || echo "$response" | jq -e '.error' >/dev/null 2>&1; then
        echo '{"error": "not_found"}'
        return
    fi

    # Extract relevant info
    local deprecated=$(echo "$response" | jq -r '.versions[.["dist-tags"].latest].deprecated // null')
    local latest_version=$(echo "$response" | jq -r '.["dist-tags"].latest // null')
    local repository=$(echo "$response" | jq -r '.repository.url // null')
    local homepage=$(echo "$response" | jq -r '.homepage // null')
    local last_publish=$(echo "$response" | jq -r '.time[.["dist-tags"].latest] // null')

    # Fetch download stats
    local downloads_response=$(curl -s "https://api.npmjs.org/downloads/point/last-month/${encoded_pkg}" 2>/dev/null)
    local monthly_downloads=$(echo "$downloads_response" | jq -r '.downloads // 0')

    local result=$(jq -n \
        --arg deprecated "$deprecated" \
        --arg latest "$latest_version" \
        --arg repo "$repository" \
        --arg homepage "$homepage" \
        --arg last_publish "$last_publish" \
        --argjson downloads "$monthly_downloads" \
        '{
            deprecated: (if $deprecated == "null" then null else $deprecated end),
            latest_version: $latest,
            repository: $repo,
            homepage: $homepage,
            last_publish: $last_publish,
            monthly_downloads: $downloads,
            is_deprecated: ($deprecated != "null" and $deprecated != null)
        }')

    LIVE_PACKAGE_CACHE[$cache_key]="$result"
    echo "$result"
}

# Fetch PyPI package info
# Usage: fetch_pypi_package_info <package>
fetch_pypi_package_info() {
    local pkg="$1"
    local cache_key="pypi:$pkg"

    # Check cache
    if [[ -n "${LIVE_PACKAGE_CACHE[$cache_key]:-}" ]]; then
        echo "${LIVE_PACKAGE_CACHE[$cache_key]}"
        return
    fi

    local response=$(curl -s "https://pypi.org/pypi/${pkg}/json" 2>/dev/null)

    if [[ -z "$response" ]] || echo "$response" | jq -e '.message' >/dev/null 2>&1; then
        echo '{"error": "not_found"}'
        return
    fi

    # Extract relevant info
    local latest_version=$(echo "$response" | jq -r '.info.version // null')
    local summary=$(echo "$response" | jq -r '.info.summary // null')
    local homepage=$(echo "$response" | jq -r '.info.home_page // null')
    local project_url=$(echo "$response" | jq -r '.info.project_url // null')

    # Check for development status classifiers (deprecated indicators)
    local dev_status=$(echo "$response" | jq -r '.info.classifiers[] | select(startswith("Development Status"))' 2>/dev/null | head -1)
    local is_inactive="false"
    if [[ "$dev_status" == *"Inactive"* ]] || [[ "$dev_status" == *"1 - Planning"* ]]; then
        is_inactive="true"
    fi

    # Check for "yanked" releases or maintenance warnings
    local yanked=$(echo "$response" | jq -r '[.releases[][] | select(.yanked == true)] | length > 0')

    local result=$(jq -n \
        --arg latest "$latest_version" \
        --arg summary "$summary" \
        --arg homepage "$homepage" \
        --arg project_url "$project_url" \
        --arg dev_status "$dev_status" \
        --argjson is_inactive "$is_inactive" \
        --argjson has_yanked "$yanked" \
        '{
            latest_version: $latest,
            summary: $summary,
            homepage: $homepage,
            project_url: $project_url,
            development_status: $dev_status,
            is_inactive: $is_inactive,
            has_yanked_releases: $has_yanked
        }')

    LIVE_PACKAGE_CACHE[$cache_key]="$result"
    echo "$result"
}

# Fetch Go package info from pkg.go.dev
# Usage: fetch_go_package_info <package>
fetch_go_package_info() {
    local pkg="$1"
    local cache_key="go:$pkg"

    # Check cache
    if [[ -n "${LIVE_PACKAGE_CACHE[$cache_key]:-}" ]]; then
        echo "${LIVE_PACKAGE_CACHE[$cache_key]}"
        return
    fi

    # pkg.go.dev doesn't have a public API, use proxy.golang.org for version info
    local encoded_pkg="${pkg}"
    local response=$(curl -s "https://proxy.golang.org/${encoded_pkg}/@latest" 2>/dev/null)

    if [[ -z "$response" ]] || [[ "$response" == *"not found"* ]]; then
        echo '{"error": "not_found"}'
        return
    fi

    local latest_version=$(echo "$response" | jq -r '.Version // null')
    local timestamp=$(echo "$response" | jq -r '.Time // null')

    # Check if module is deprecated via go.mod retract or module deprecation
    local mod_response=$(curl -s "https://proxy.golang.org/${encoded_pkg}/@v/${latest_version}.mod" 2>/dev/null)
    local is_deprecated="false"
    if [[ "$mod_response" == *"// Deprecated:"* ]] || [[ "$mod_response" == *"retract"* ]]; then
        is_deprecated="true"
    fi

    local result=$(jq -n \
        --arg latest "$latest_version" \
        --arg timestamp "$timestamp" \
        --argjson is_deprecated "$is_deprecated" \
        '{
            latest_version: $latest,
            last_publish: $timestamp,
            is_deprecated: $is_deprecated
        }')

    LIVE_PACKAGE_CACHE[$cache_key]="$result"
    echo "$result"
}

# Get live package info based on ecosystem
# Usage: get_live_package_info <package> <ecosystem>
get_live_package_info() {
    local pkg="$1"
    local ecosystem="${2:-npm}"

    if [[ "$LIVE_DATA_ENABLED" != "true" ]]; then
        echo '{"live_data": false}'
        return
    fi

    case "$ecosystem" in
        npm|node)
            fetch_npm_package_info "$pkg"
            ;;
        pypi|python)
            fetch_pypi_package_info "$pkg"
            ;;
        go|golang)
            fetch_go_package_info "$pkg"
            ;;
        *)
            echo '{"error": "unsupported_ecosystem"}'
            ;;
    esac
}

# Check if package is deprecated using live data
# Usage: is_package_deprecated <package> <ecosystem>
is_package_deprecated() {
    local pkg="$1"
    local ecosystem="${2:-npm}"

    local live_info=$(get_live_package_info "$pkg" "$ecosystem")

    case "$ecosystem" in
        npm|node)
            echo "$live_info" | jq -r '.is_deprecated // false'
            ;;
        pypi|python)
            echo "$live_info" | jq -r '.is_inactive // false'
            ;;
        go|golang)
            echo "$live_info" | jq -r '.is_deprecated // false'
            ;;
        *)
            echo "false"
            ;;
    esac
}

# Get deprecation message if available
# Usage: get_deprecation_message <package> <ecosystem>
get_deprecation_message() {
    local pkg="$1"
    local ecosystem="${2:-npm}"

    local live_info=$(get_live_package_info "$pkg" "$ecosystem")

    case "$ecosystem" in
        npm|node)
            local msg=$(echo "$live_info" | jq -r '.deprecated // null')
            if [[ "$msg" != "null" && -n "$msg" ]]; then
                echo "$msg"
            fi
            ;;
        pypi|python)
            local status=$(echo "$live_info" | jq -r '.development_status // null')
            if [[ "$status" == *"Inactive"* ]]; then
                echo "Project marked as Inactive"
            fi
            ;;
        go|golang)
            local dep=$(echo "$live_info" | jq -r '.is_deprecated // false')
            if [[ "$dep" == "true" ]]; then
                echo "Module is deprecated or retracted"
            fi
            ;;
    esac
}

#############################################################################
# Library Replacement Database
# Format: old_package|replacement|reason|migration_effort
# Migration effort: trivial, easy, moderate, significant, major
#############################################################################

# NPM Library Replacements
NPM_REPLACEMENTS="request|axios|request is deprecated, axios is actively maintained|easy
request|got|request is deprecated, got has modern API|easy
request|node-fetch|request is deprecated, node-fetch is minimal|easy
moment|dayjs|moment is in maintenance mode, dayjs is smaller|trivial
moment|date-fns|moment is in maintenance mode, date-fns is tree-shakeable|moderate
moment|luxon|moment is in maintenance mode, luxon is from Moment team|moderate
underscore|lodash|lodash is more performant and feature-rich|easy
underscore|ramda|ramda is better for functional programming|moderate
colors|chalk|colors had supply chain incident|trivial
colors|picocolors|picocolors is minimal and fast|trivial
faker|@faker-js/faker|faker was abandoned, community fork available|trivial
left-pad|native String.padStart|left-pad incident, use native|trivial
event-stream|readable-stream|event-stream compromised|easy
uuid|nanoid|nanoid is smaller and faster|easy
express-validator|zod|zod provides better TypeScript integration|moderate
joi|zod|zod provides better TypeScript integration|moderate
yup|zod|zod provides better TypeScript integration|moderate
commander|yargs|yargs has more features|easy
minimist|yargs|yargs is more full-featured|easy
mocha|jest|jest has better DX and parallel execution|moderate
jasmine|jest|jest has better DX and parallel execution|moderate
karma|jest|jest eliminates need for browser runner|significant
webpack|vite|vite is faster for development|significant
webpack|esbuild|esbuild is much faster|moderate
gulp|npm scripts|modern npm has adequate task running|moderate
grunt|npm scripts|grunt is outdated|moderate
bower|npm|bower is deprecated|significant
tslint|eslint|tslint is deprecated|easy
node-sass|sass|node-sass is deprecated|trivial
node-gyp|prebuild|prebuild avoids native compilation|moderate
bcrypt|argon2|argon2 is more secure|easy
crypto-js|native crypto|Node has built-in crypto module|moderate
bluebird|native Promise|Native Promises are adequate now|moderate
q|native Promise|Native Promises are adequate now|moderate
async|native async/await|Modern JS has async/await|moderate
lodash|native Array methods|Many lodash methods have native equivalents|varies"

# Python Library Replacements
PYTHON_REPLACEMENTS="urllib2|requests|urllib2 is Python 2 only|easy
urllib2|httpx|httpx supports async|easy
requests|httpx|httpx supports async and HTTP/2|easy
optparse|argparse|optparse is deprecated|easy
optparse|click|click has better UX|moderate
optparse|typer|typer uses type hints|moderate
nose|pytest|nose is unmaintained|moderate
unittest|pytest|pytest has better features|moderate
mock|unittest.mock|mock is in stdlib now|trivial
fabric|invoke|fabric 1.x is deprecated|moderate
pycrypto|cryptography|pycrypto is unmaintained|moderate
pycryptodome|cryptography|cryptography is more maintained|moderate
PIL|pillow|PIL is unmaintained|trivial
mysql-python|mysqlclient|mysql-python is unmaintained|easy
MySQLdb|mysqlclient|MySQLdb is Python 2 only|easy
pymongo|motor|motor for async MongoDB|moderate
redis-py|redis|package renamed|trivial
python-dateutil|pendulum|pendulum has better API|moderate
pytz|zoneinfo|zoneinfo is in stdlib 3.9+|easy
six|native|Python 2 compatibility not needed|moderate
future|native|Python 2 compatibility not needed|moderate
typing_extensions|native|Most features in stdlib now|varies
flask-restful|flask-smorest|flask-restful is maintenance mode|moderate
django-rest-framework|ninja|ninja is faster and simpler|significant"

# Go Library Replacements
GO_REPLACEMENTS="github.com/gorilla/mux|github.com/go-chi/chi|chi is lighter and faster|easy
github.com/gorilla/mux|github.com/gin-gonic/gin|gin is more full-featured|moderate
github.com/pkg/errors|errors|stdlib errors has wrapping now|easy
github.com/sirupsen/logrus|go.uber.org/zap|zap is more performant|moderate
github.com/sirupsen/logrus|log/slog|slog is in stdlib 1.21+|easy
gopkg.in/yaml.v2|gopkg.in/yaml.v3|v3 is current|trivial
github.com/dgrijalva/jwt-go|github.com/golang-jwt/jwt|dgrijalva is unmaintained|easy
github.com/jinzhu/gorm|gorm.io/gorm|gorm v2 is current|moderate"

#############################################################################
# Recommendation Functions
#############################################################################

# Get replacement recommendations for a package
# Usage: get_replacements <package> <ecosystem>
get_replacements() {
    local pkg="$1"
    local ecosystem="${2:-npm}"
    local replacements_list=""

    case "$ecosystem" in
        npm|node)
            replacements_list="$NPM_REPLACEMENTS"
            ;;
        pypi|python)
            replacements_list="$PYTHON_REPLACEMENTS"
            ;;
        go|golang)
            replacements_list="$GO_REPLACEMENTS"
            ;;
        *)
            echo "[]"
            return
            ;;
    esac

    local results="[]"

    while IFS= read -r line; do
        [[ -z "$line" ]] && continue
        local old_pkg=$(echo "$line" | cut -d'|' -f1)
        local new_pkg=$(echo "$line" | cut -d'|' -f2)
        local reason=$(echo "$line" | cut -d'|' -f3)
        local effort=$(echo "$line" | cut -d'|' -f4)

        if [[ "$pkg" == "$old_pkg" ]]; then
            results=$(echo "$results" | jq --arg new "$new_pkg" --arg reason "$reason" --arg effort "$effort" \
                '. + [{"replacement": $new, "reason": $reason, "migration_effort": $effort}]')
        fi
    done <<< "$replacements_list"

    echo "$results"
}

# Check if a package has known replacements
# Usage: has_replacement <package> <ecosystem>
has_replacement() {
    local pkg="$1"
    local ecosystem="${2:-npm}"

    local replacements=$(get_replacements "$pkg" "$ecosystem")
    local count=$(echo "$replacements" | jq 'length')

    if [[ $count -gt 0 ]]; then
        echo "true"
    else
        echo "false"
    fi
}

# Get package health score from deps.dev
# Usage: get_health_score <package> <ecosystem>
get_health_score() {
    local pkg="$1"
    local ecosystem="$2"

    if ! type get_package_info &>/dev/null; then
        echo '{"error": "deps_dev_client_not_loaded"}'
        return
    fi

    local pkg_info=$(get_package_info "$ecosystem" "$pkg" 2>/dev/null)

    if [[ -z "$pkg_info" || "$pkg_info" == *"error"* ]]; then
        echo '{"score": null, "error": "package_not_found"}'
        return
    fi

    local scorecard=$(echo "$pkg_info" | jq -r '.scorecard.score // null')
    local dependent_count=$(echo "$pkg_info" | jq -r '.dependentCount // 0')

    echo "{
        \"openssf_score\": $scorecard,
        \"dependent_count\": $dependent_count
    }"
}

# Analyze a package and provide recommendations
# Usage: analyze_package <package> <ecosystem> <version>
analyze_package() {
    local pkg="$1"
    local ecosystem="${2:-npm}"
    local version="${3:-}"

    local recommendations=()
    local risk_level="low"
    local action_required="false"
    local deprecation_source=""
    local deprecation_message=""

    # Get live package info first (checks deprecation status from registry)
    local live_info=$(get_live_package_info "$pkg" "$ecosystem")
    local live_deprecated=$(is_package_deprecated "$pkg" "$ecosystem")

    # Check for LIVE deprecation from registry
    if [[ "$live_deprecated" == "true" ]]; then
        action_required="true"
        risk_level="high"
        deprecation_source="registry"
        deprecation_message=$(get_deprecation_message "$pkg" "$ecosystem")
        recommendations+=("⚠️ DEPRECATED: Package is marked deprecated in the registry")
        if [[ -n "$deprecation_message" ]]; then
            recommendations+=("Registry message: $deprecation_message")
        fi
    fi

    # Check for known replacements from our database
    local replacements=$(get_replacements "$pkg" "$ecosystem")
    local replacement_count=$(echo "$replacements" | jq 'length')

    if [[ $replacement_count -gt 0 ]]; then
        action_required="true"
        if [[ "$risk_level" == "low" ]]; then
            risk_level="medium"
        fi
        recommendations+=("This package has recommended replacements available")
    fi

    # Get health metrics from deps.dev
    local health=$(get_health_score "$pkg" "$ecosystem")
    local openssf_score=$(echo "$health" | jq -r '.openssf_score // null')
    local dependent_count=$(echo "$health" | jq -r '.dependent_count // 0')

    # Analyze health
    if [[ "$openssf_score" != "null" ]]; then
        if [[ $(echo "$openssf_score < 4" | bc -l 2>/dev/null || echo "0") == "1" ]]; then
            if [[ "$risk_level" != "high" ]]; then
                risk_level="high"
            fi
            recommendations+=("Low OpenSSF Scorecard score ($openssf_score) indicates maintenance concerns")
        elif [[ $(echo "$openssf_score < 6" | bc -l 2>/dev/null || echo "0") == "1" ]]; then
            if [[ "$risk_level" == "low" ]]; then
                risk_level="medium"
            fi
            recommendations+=("Moderate OpenSSF Scorecard score ($openssf_score) - monitor for issues")
        fi
    fi

    # Check adoption from live data
    local monthly_downloads=$(echo "$live_info" | jq -r '.monthly_downloads // 0')
    if [[ "$monthly_downloads" != "null" && "$monthly_downloads" -gt 0 ]]; then
        if [[ $monthly_downloads -lt 1000 ]]; then
            recommendations+=("Low download count ($monthly_downloads/month) - consider more established alternatives")
        fi
    elif [[ $dependent_count -lt 100 ]]; then
        recommendations+=("Low adoption ($dependent_count dependents) - consider more established alternatives")
    fi

    # Check last publish date for staleness
    local last_publish=$(echo "$live_info" | jq -r '.last_publish // null')
    if [[ "$last_publish" != "null" && -n "$last_publish" ]]; then
        local publish_epoch=$(date -j -f "%Y-%m-%dT%H:%M:%S" "$(echo "$last_publish" | cut -d'.' -f1)" "+%s" 2>/dev/null || echo "0")
        local now_epoch=$(date "+%s")
        local days_since=$(( (now_epoch - publish_epoch) / 86400 ))

        if [[ $days_since -gt 730 ]]; then
            if [[ "$risk_level" == "low" ]]; then
                risk_level="medium"
            fi
            recommendations+=("No updates in over 2 years ($days_since days) - may be abandoned")
        elif [[ $days_since -gt 365 ]]; then
            recommendations+=("No updates in over 1 year ($days_since days) - check if actively maintained")
        fi
    fi

    local recommendations_json=$(printf '%s\n' "${recommendations[@]}" 2>/dev/null | jq -R . | jq -s '.' || echo "[]")

    # Build live_info for output
    local live_info_output=$(echo "$live_info" | jq '{
        latest_version: .latest_version,
        monthly_downloads: .monthly_downloads,
        last_publish: .last_publish,
        is_deprecated: .is_deprecated
    }' 2>/dev/null || echo '{}')

    echo "{
        \"package\": \"$pkg\",
        \"ecosystem\": \"$ecosystem\",
        \"version\": \"$version\",
        \"risk_level\": \"$risk_level\",
        \"action_required\": $action_required,
        \"is_deprecated\": $live_deprecated,
        \"deprecation_message\": $(echo "$deprecation_message" | jq -Rs '.'),
        \"replacements\": $replacements,
        \"health_metrics\": $health,
        \"live_info\": $live_info_output,
        \"recommendations\": $recommendations_json
    }" | jq '.'
}

# Analyze multiple packages from a manifest
# Usage: analyze_manifest <manifest_path>
analyze_manifest() {
    local manifest_path="$1"

    if [[ ! -f "$manifest_path" ]]; then
        echo '{"error": "manifest_not_found"}'
        return 1
    fi

    local ecosystem=""
    local packages="[]"

    # Detect manifest type and extract packages
    local filename=$(basename "$manifest_path")

    case "$filename" in
        package.json)
            ecosystem="npm"
            # Extract dependencies and devDependencies
            local deps=$(jq -r '.dependencies // {} | keys[]' "$manifest_path" 2>/dev/null)
            local dev_deps=$(jq -r '.devDependencies // {} | keys[]' "$manifest_path" 2>/dev/null)
            packages=$(echo -e "$deps\n$dev_deps" | grep -v '^$' | jq -R . | jq -s '.')
            ;;
        requirements.txt)
            ecosystem="python"
            # Extract package names (without versions)
            packages=$(cut -d'=' -f1 "$manifest_path" 2>/dev/null | cut -d'>' -f1 | cut -d'<' -f1 | cut -d'~' -f1 | cut -d'[' -f1 | grep -v '^#' | grep -v '^$' | tr '[:upper:]' '[:lower:]' | jq -R . | jq -s '.')
            ;;
        pyproject.toml)
            ecosystem="python"
            # Extract from dependencies section (basic parsing)
            packages=$(grep -E '^\s*"?[a-zA-Z]' "$manifest_path" 2>/dev/null | grep -v '\[' | cut -d'=' -f1 | cut -d'"' -f2 | tr '[:upper:]' '[:lower:]' | jq -R . | jq -s '.' || echo "[]")
            ;;
        go.mod)
            ecosystem="go"
            # Extract require statements
            packages=$(grep -E '^\s+[a-z]' "$manifest_path" 2>/dev/null | awk '{print $1}' | jq -R . | jq -s '.' || echo "[]")
            ;;
        Gemfile)
            ecosystem="rubygems"
            # Extract gem names
            packages=$(grep -E "^\s*gem\s+" "$manifest_path" 2>/dev/null | sed "s/.*gem ['\"]\\([^'\"]*\\).*/\\1/" | jq -R . | jq -s '.' || echo "[]")
            ;;
        *)
            echo '{"error": "unsupported_manifest_type", "filename": "'$filename'"}'
            return 1
            ;;
    esac

    local results="[]"
    local total=0
    local with_replacements=0
    local high_risk=0
    local medium_risk=0

    while IFS= read -r pkg; do
        [[ -z "$pkg" || "$pkg" == "null" ]] && continue
        total=$((total + 1))

        local analysis=$(analyze_package "$pkg" "$ecosystem")
        results=$(echo "$results" | jq --argjson a "$analysis" '. + [$a]')

        # Count statistics
        if [[ $(echo "$analysis" | jq -r '.action_required') == "true" ]]; then
            with_replacements=$((with_replacements + 1))
        fi
        local pkg_risk=$(echo "$analysis" | jq -r '.risk_level')
        if [[ "$pkg_risk" == "high" ]]; then
            high_risk=$((high_risk + 1))
        elif [[ "$pkg_risk" == "medium" ]]; then
            medium_risk=$((medium_risk + 1))
        fi
    done < <(echo "$packages" | jq -r '.[]')

    echo "{
        \"manifest\": \"$manifest_path\",
        \"ecosystem\": \"$ecosystem\",
        \"summary\": {
            \"total_packages\": $total,
            \"packages_with_replacements\": $with_replacements,
            \"risk_breakdown\": {
                \"high\": $high_risk,
                \"medium\": $medium_risk,
                \"low\": $((total - high_risk - medium_risk))
            }
        },
        \"packages\": $results
    }" | jq '.'
}

# Generate migration plan for a project
# Usage: generate_migration_plan <project_dir>
generate_migration_plan() {
    local project_dir="$1"

    if [[ ! -d "$project_dir" ]]; then
        echo '{"error": "directory_not_found"}'
        return 1
    fi

    local manifests_analyzed="[]"
    local total_packages=0
    local total_replacements=0
    local priority_migrations="[]"

    # Find and analyze manifests
    for manifest in package.json requirements.txt pyproject.toml go.mod Gemfile; do
        local manifest_path="$project_dir/$manifest"
        if [[ -f "$manifest_path" ]]; then
            local analysis=$(analyze_manifest "$manifest_path")
            manifests_analyzed=$(echo "$manifests_analyzed" | jq --argjson a "$analysis" '. + [$a]')

            # Aggregate stats
            local pkg_count=$(echo "$analysis" | jq -r '.summary.total_packages')
            local rep_count=$(echo "$analysis" | jq -r '.summary.packages_with_replacements')
            total_packages=$((total_packages + pkg_count))
            total_replacements=$((total_replacements + rep_count))

            # Extract high-priority migrations
            local high_priority=$(echo "$analysis" | jq '[.packages[] | select(.risk_level == "high" and .action_required == true)]')
            priority_migrations=$(echo "$priority_migrations" | jq --argjson hp "$high_priority" '. + $hp')
        fi
    done

    # Estimate total effort
    local effort_score=0
    local effort_map='{"trivial": 1, "easy": 2, "moderate": 4, "significant": 8, "major": 16}'

    while IFS= read -r pkg; do
        [[ -z "$pkg" || "$pkg" == "null" ]] && continue
        local effort=$(echo "$pkg" | jq -r '.replacements[0].migration_effort // "moderate"')
        local points=$(echo "$effort_map" | jq -r ".[\"$effort\"] // 4")
        effort_score=$((effort_score + points))
    done < <(echo "$priority_migrations" | jq -c '.[]')

    local effort_rating="low"
    if [[ $effort_score -gt 50 ]]; then
        effort_rating="high"
    elif [[ $effort_score -gt 20 ]]; then
        effort_rating="medium"
    fi

    echo "{
        \"project_dir\": \"$project_dir\",
        \"summary\": {
            \"total_packages\": $total_packages,
            \"packages_needing_migration\": $total_replacements,
            \"priority_migrations\": $(echo "$priority_migrations" | jq 'length'),
            \"estimated_effort\": \"$effort_rating\",
            \"effort_score\": $effort_score
        },
        \"priority_migrations\": $priority_migrations,
        \"manifests_analyzed\": $manifests_analyzed
    }" | jq '.'
}

#############################################################################
# Export Functions
#############################################################################

export -f fetch_npm_package_info
export -f fetch_pypi_package_info
export -f fetch_go_package_info
export -f get_live_package_info
export -f is_package_deprecated
export -f get_deprecation_message
export -f get_replacements
export -f has_replacement
export -f get_health_score
export -f analyze_package
export -f analyze_manifest
export -f generate_migration_plan
