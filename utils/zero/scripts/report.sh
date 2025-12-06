#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Phantom Report Generator
# Generate analysis reports in various formats
#
# Usage: ./report.sh <org/repo> [options]
#        ./report.sh --org <name> [options]
#        ./report.sh --interactive
#############################################################################

set -e

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PHANTOM_DIR="$(dirname "$SCRIPT_DIR")"

# Load Phantom library
source "$PHANTOM_DIR/lib/phantom-lib.sh"

# Load report utilities
source "$PHANTOM_DIR/lib/report-common.sh"

#############################################################################
# Configuration
#############################################################################

TARGET=""
ORG=""
REPORT_TYPE="summary"
OUTPUT_FORMAT="terminal"
OUTPUT_FILE=""
INTERACTIVE=false
AUTO_NAME=true  # Use standardized naming by default

#############################################################################
# Usage
#############################################################################

usage() {
    cat << EOF
Phantom Report Generator - Generate analysis reports

Usage: $0 [options] [<org/repo>]

TARGETS:
    <org/repo>          Report for a specific project
    --org <name>        Aggregate report for all projects in an org

REPORT TYPES:
    -t, --type <type>   Report type (default: summary)
        summary         High-level overview of all findings
        security        Vulnerabilities, secrets, code security
        licenses        License compliance and dependency licenses
        compliance      SBOM, policy compliance
        sbom            Software Bill of Materials
        supply-chain    Dependencies, provenance, health
        dora            DevOps metrics and performance
        code-ownership  3-tier ownership analysis (basic/analysis/AI)
        full            Comprehensive report (all sections)

OUTPUT FORMATS:
    -f, --format <fmt>  Output format (default: terminal)
        terminal        Colored terminal output
        markdown        GitHub-flavored markdown
        json            Structured JSON
        html            Self-contained HTML
        csv             CSV data export

OPTIONS:
    -o, --output <file> Write to file instead of stdout
    --no-auto-name      Don't use standardized naming (use exact -o filename)
    -i, --interactive   Interactive mode (select options via menu)
    -h, --help          Show this help

REPORT NAMING:
    By default, reports are saved to the project's analysis/reports/ folder
    with standardized naming: <type>_<scanID>_<datetime>.<ext>

    Use -o to specify a custom output path (relative or absolute)

EXAMPLES:
    $0 expressjs/express                           # Summary report (terminal)
    $0 expressjs/express -t security -f markdown   # Security report in markdown
    $0 --org expressjs -t security -f html          # Org security report as HTML
    $0 -i                                          # Interactive mode

EOF
    exit 0
}

#############################################################################
# Argument Parsing
#############################################################################

parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -h|--help)
                usage
                ;;
            --org)
                ORG="$2"
                shift 2
                ;;
            -t|--type)
                REPORT_TYPE="$2"
                shift 2
                ;;
            -f|--format)
                OUTPUT_FORMAT="$2"
                shift 2
                ;;
            -o|--output)
                OUTPUT_FILE="$2"
                shift 2
                ;;
            -i|--interactive)
                INTERACTIVE=true
                shift
                ;;
            --no-auto-name)
                AUTO_NAME=false
                shift
                ;;
            --json)
                # Legacy support
                OUTPUT_FORMAT="json"
                shift
                ;;
            -*)
                echo -e "${RED}Error: Unknown option $1${NC}" >&2
                exit 1
                ;;
            *)
                if [[ -z "$TARGET" ]]; then
                    TARGET="$1"
                else
                    echo -e "${RED}Error: Multiple targets specified${NC}" >&2
                    exit 1
                fi
                shift
                ;;
        esac
    done

    # Validate report type
    if ! validate_report_type "$REPORT_TYPE"; then
        echo -e "${RED}Error: Invalid report type '$REPORT_TYPE'${NC}" >&2
        echo "Valid types: ${REPORT_TYPES[*]}"
        exit 1
    fi

    # Validate output format
    if ! validate_format "$OUTPUT_FORMAT"; then
        echo -e "${RED}Error: Invalid format '$OUTPUT_FORMAT'${NC}" >&2
        echo "Valid formats: ${REPORT_FORMATS[*]}"
        exit 1
    fi
}

#############################################################################
# Interactive Mode
#############################################################################

interactive_menu() {
    echo
    echo -e "${BOLD}PHANTOM REPORT GENERATOR${NC}"
    hr
    echo

    # Select target
    echo -e "${BOLD}Select target:${NC}"
    echo

    # List available projects
    local projects=()
    local idx=1

    if [[ -d "$GIBSON_PROJECTS_DIR" ]]; then
        while IFS= read -r org_dir; do
            local org_name=$(basename "$org_dir")
            while IFS= read -r repo_dir; do
                local repo_name=$(basename "$repo_dir")
                if [[ -f "$repo_dir/analysis/manifest.json" ]]; then
                    projects+=("$org_name/$repo_name")
                    printf "  ${CYAN}%2d${NC}  %s/%s\n" "$idx" "$org_name" "$repo_name"
                    ((idx++))
                fi
            done < <(find "$org_dir" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | sort)
        done < <(find "$GIBSON_PROJECTS_DIR" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | sort)
    fi

    if [[ ${#projects[@]} -eq 0 ]]; then
        echo -e "  ${DIM}No hydrated projects found${NC}"
        echo -e "  ${DIM}Run a hydration first: ./phantom.sh${NC}"
        echo
        exit 1
    fi

    echo
    read -p "Select project (1-${#projects[@]}): " project_choice

    if [[ "$project_choice" =~ ^[0-9]+$ ]] && [[ $project_choice -ge 1 ]] && [[ $project_choice -le ${#projects[@]} ]]; then
        TARGET="${projects[$((project_choice-1))]}"
    else
        echo -e "${RED}Invalid selection${NC}"
        exit 1
    fi

    echo
    echo -e "${BOLD}Select report type:${NC}"
    echo
    echo -e "  ${CYAN}1${NC}  Summary        High-level overview of all findings"
    echo -e "  ${CYAN}2${NC}  Security       Vulnerabilities, secrets, code security"
    echo -e "  ${CYAN}3${NC}  Licenses       License compliance and dependency licenses"
    echo -e "  ${CYAN}4${NC}  Compliance     SBOM, policy compliance"
    echo -e "  ${CYAN}5${NC}  SBOM           Software Bill of Materials"
    echo -e "  ${CYAN}6${NC}  Supply Chain   Dependencies, provenance, health"
    echo -e "  ${CYAN}7${NC}  DORA           DevOps metrics and performance"
    echo -e "  ${CYAN}8${NC}  Full           Comprehensive report (all sections)"
    echo
    read -p "Select type (1-8) [1]: " type_choice

    case "${type_choice:-1}" in
        1) REPORT_TYPE="summary" ;;
        2) REPORT_TYPE="security" ;;
        3) REPORT_TYPE="licenses" ;;
        4) REPORT_TYPE="compliance" ;;
        5) REPORT_TYPE="sbom" ;;
        6) REPORT_TYPE="supply-chain" ;;
        7) REPORT_TYPE="dora" ;;
        8) REPORT_TYPE="full" ;;
        *) REPORT_TYPE="summary" ;;
    esac

    echo
    echo -e "${BOLD}Select output format:${NC}"
    echo
    echo -e "  ${CYAN}t${NC}  Terminal      Colored terminal output"
    echo -e "  ${CYAN}m${NC}  Markdown      GitHub-flavored markdown"
    echo -e "  ${CYAN}j${NC}  JSON          Structured JSON"
    echo -e "  ${CYAN}h${NC}  HTML          Self-contained HTML"
    echo -e "  ${CYAN}c${NC}  CSV           CSV data export"
    echo
    read -p "Select format (t/m/j/h/c) [t]: " format_choice

    case "${format_choice:-t}" in
        t|T) OUTPUT_FORMAT="terminal" ;;
        m|M) OUTPUT_FORMAT="markdown" ;;
        j|J) OUTPUT_FORMAT="json" ;;
        h|H) OUTPUT_FORMAT="html" ;;
        c|C) OUTPUT_FORMAT="csv" ;;
        *) OUTPUT_FORMAT="terminal" ;;
    esac

    # For non-terminal output, reports are auto-saved with standardized names
    if [[ "$OUTPUT_FORMAT" != "terminal" ]]; then
        echo
        echo -e "${DIM}Reports will be saved to: ~/.phantom/projects/<org>/<repo>/analysis/reports/${NC}"
        echo -e "${DIM}Naming format: ${REPORT_TYPE}_<scanID>_<datetime>.<ext>${NC}"
    fi

    echo
}

#############################################################################
# Report Naming
#############################################################################

# Generate standardized report filename
# Format: <type>.<ext> (single file per report type, overwrites on regeneration)
generate_report_filename() {
    local project_id="$1"
    local report_type="$2"
    local format="$3"

    # Get extension
    local ext
    case "$format" in
        markdown) ext="md" ;;
        json) ext="json" ;;
        html) ext="html" ;;
        csv) ext="csv" ;;
        terminal) ext="txt" ;;
        *) ext="txt" ;;
    esac

    echo "${report_type}.${ext}"
}

# Get reports directory for a project
get_reports_dir() {
    local project_id="$1"
    local analysis_path=$(gibson_project_analysis_path "$project_id")
    echo "$analysis_path/reports"
}

#############################################################################
# Report Generation
#############################################################################

# Load report type module
load_report_type() {
    local type="$1"
    local module="$PHANTOM_DIR/lib/report-types/${type}.sh"

    if [[ -f "$module" ]]; then
        source "$module"
        return 0
    else
        echo -e "${YELLOW}Warning: Report type '$type' not yet implemented${NC}" >&2
        return 1
    fi
}

# Load format module
load_format_module() {
    local format="$1"
    local module="$PHANTOM_DIR/lib/report-formats/${format}.sh"

    if [[ -f "$module" ]]; then
        source "$module"
        return 0
    else
        echo -e "${YELLOW}Warning: Format '$format' not yet implemented${NC}" >&2
        return 1
    fi
}

# Generate report for a single project
generate_project_report() {
    local project_id="$1"
    local analysis_path=$(gibson_project_analysis_path "$project_id")

    if [[ ! -d "$analysis_path" ]]; then
        echo -e "${RED}Error: No analysis found for '$project_id'${NC}" >&2
        exit 1
    fi

    # Load report type module
    if ! load_report_type "$REPORT_TYPE"; then
        # Fallback to summary for unimplemented types
        if [[ "$REPORT_TYPE" != "summary" ]]; then
            echo -e "${YELLOW}Falling back to summary report${NC}" >&2
            REPORT_TYPE="summary"
            load_report_type "summary" || exit 1
        else
            exit 1
        fi
    fi

    # Load format module
    if ! load_format_module "$OUTPUT_FORMAT"; then
        # Fallback to terminal for unimplemented formats
        if [[ "$OUTPUT_FORMAT" != "terminal" ]]; then
            echo -e "${YELLOW}Falling back to terminal output${NC}" >&2
            OUTPUT_FORMAT="terminal"
            load_format_module "terminal" || exit 1
        else
            exit 1
        fi
    fi

    # Generate the report
    # Report type modules export: generate_<type>_data()
    # Format modules export: format_<format>_output()

    local report_data
    report_data=$(generate_report_data "$project_id" "$analysis_path")

    # Determine output destination
    local output_dest=""
    if [[ "$OUTPUT_FORMAT" != "terminal" ]]; then
        if [[ -n "$OUTPUT_FILE" ]]; then
            # User specified a file
            if [[ "$AUTO_NAME" == "true" ]] && [[ ! "$OUTPUT_FILE" =~ ^/ ]] && [[ ! "$OUTPUT_FILE" =~ ^\.\. ]]; then
                # Relative path without explicit path - use standardized naming in reports dir
                local reports_dir=$(get_reports_dir "$project_id")
                mkdir -p "$reports_dir"
                output_dest="$reports_dir/$(generate_report_filename "$project_id" "$REPORT_TYPE" "$OUTPUT_FORMAT")"
            else
                # Absolute path or explicit relative path - use as-is
                output_dest="$OUTPUT_FILE"
            fi
        else
            # No file specified - use standardized naming
            local reports_dir=$(get_reports_dir "$project_id")
            mkdir -p "$reports_dir"
            output_dest="$reports_dir/$(generate_report_filename "$project_id" "$REPORT_TYPE" "$OUTPUT_FORMAT")"
        fi
    fi

    if [[ -n "$output_dest" ]]; then
        format_report_output "$report_data" "$project_id" > "$output_dest"
        echo -e "${GREEN}Report saved to: $output_dest${NC}"
    else
        format_report_output "$report_data" "$project_id"
    fi
}

# Generate report for an organization
generate_org_report() {
    local org="$1"
    local projects=$(gibson_list_org_projects "$org")

    if [[ -z "$projects" ]]; then
        echo -e "${RED}Error: No projects found for org '$org'${NC}" >&2
        exit 1
    fi

    # Load modules
    load_report_type "$REPORT_TYPE" || exit 1
    load_format_module "$OUTPUT_FORMAT" || exit 1

    # Generate org aggregate report
    local report_data
    report_data=$(generate_org_report_data "$org" "$projects")

    # Determine output destination for org reports
    local output_dest=""
    if [[ "$OUTPUT_FORMAT" != "terminal" ]]; then
        if [[ -n "$OUTPUT_FILE" ]]; then
            if [[ "$AUTO_NAME" == "true" ]] && [[ ! "$OUTPUT_FILE" =~ ^/ ]] && [[ ! "$OUTPUT_FILE" =~ ^\.\. ]]; then
                # Use org-level reports directory
                local org_reports_dir="$GIBSON_PROJECTS_DIR/$org/reports"
                mkdir -p "$org_reports_dir"
                local datetime=$(date +"%Y%m%d-%H%M%S")
                local ext
                case "$OUTPUT_FORMAT" in
                    markdown) ext="md" ;;
                    json) ext="json" ;;
                    html) ext="html" ;;
                    csv) ext="csv" ;;
                    *) ext="txt" ;;
                esac
                output_dest="$org_reports_dir/${REPORT_TYPE}_org_${datetime}.${ext}"
            else
                output_dest="$OUTPUT_FILE"
            fi
        else
            # No file specified - use standardized naming
            local org_reports_dir="$GIBSON_PROJECTS_DIR/$org/reports"
            mkdir -p "$org_reports_dir"
            local datetime=$(date +"%Y%m%d-%H%M%S")
            local ext
            case "$OUTPUT_FORMAT" in
                markdown) ext="md" ;;
                json) ext="json" ;;
                html) ext="html" ;;
                csv) ext="csv" ;;
                *) ext="txt" ;;
            esac
            output_dest="$org_reports_dir/${REPORT_TYPE}_org_${datetime}.${ext}"
        fi
    fi

    if [[ -n "$output_dest" ]]; then
        format_report_output "$report_data" "$org" > "$output_dest"
        echo -e "${GREEN}Report saved to: $output_dest${NC}"
    else
        format_report_output "$report_data" "$org"
    fi
}

#############################################################################
# Main
#############################################################################

main() {
    # Ensure Gibson is initialized
    gibson_ensure_initialized

    # Parse args first (if any)
    if [[ $# -gt 0 ]]; then
        parse_args "$@"
    fi

    # Interactive mode - either explicitly requested or no target specified
    if [[ "$INTERACTIVE" == "true" ]] || [[ -z "$TARGET" && -z "$ORG" ]]; then
        interactive_menu
    fi

    # Validate we have a target
    if [[ -z "$TARGET" ]] && [[ -z "$ORG" ]]; then
        echo -e "${RED}Error: No target specified${NC}" >&2
        echo "Usage: $0 <org/repo> or $0 --org <name>"
        exit 1
    fi

    # Generate report
    if [[ -n "$ORG" ]]; then
        generate_org_report "$ORG"
    else
        local project_id=$(gibson_project_id "$TARGET")
        generate_project_report "$project_id"
    fi
}

main "$@"
