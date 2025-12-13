#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# JSON Format Output
# Structured JSON output for reports (passthrough with formatting)
#############################################################################

# Format report output to JSON
# Usage: format_report_output <json_data> <target_id>
format_report_output() {
    local json_data="$1"
    local target_id="$2"

    # Pretty print the JSON
    echo "$json_data" | jq '.'
}

export -f format_report_output
