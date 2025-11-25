#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# OCSP Verification Library
# Check OCSP stapling and verify certificate revocation status
#############################################################################

# OCSP response status codes
OCSP_GOOD="good"
OCSP_REVOKED="revoked"
OCSP_UNKNOWN="unknown"
OCSP_ERROR="error"

# Extract OCSP responder URI from certificate
# Usage: get_ocsp_uri <cert_file>
get_ocsp_uri() {
    local cert_file="$1"

    openssl x509 -in "$cert_file" -noout -ocsp_uri 2>/dev/null
}

# Check if server supports OCSP stapling
# Usage: check_ocsp_stapling <host> [port]
# Returns: 0 if stapling supported, 1 if not
check_ocsp_stapling() {
    local host="$1"
    local port="${2:-443}"

    # Request OCSP stapling from server
    local result=$(echo | openssl s_client -connect "$host:$port" -status 2>/dev/null)

    if echo "$result" | grep -q "OCSP Response Status: successful"; then
        echo "SUPPORTED"
        return 0
    elif echo "$result" | grep -q "OCSP response: no response sent"; then
        echo "NOT_SUPPORTED"
        return 1
    else
        echo "UNKNOWN"
        return 2
    fi
}

# Get OCSP stapling response from server
# Usage: get_ocsp_staple <host> [port]
get_ocsp_staple() {
    local host="$1"
    local port="${2:-443}"

    echo | openssl s_client -connect "$host:$port" -status 2>/dev/null | \
        sed -n '/OCSP Response Data/,/---/p'
}

# Verify certificate via OCSP
# Usage: verify_ocsp <cert_file> <issuer_cert_file> [ocsp_uri]
verify_ocsp() {
    local cert_file="$1"
    local issuer_cert="$2"
    local ocsp_uri="${3:-}"

    # Get OCSP URI from certificate if not provided
    if [[ -z "$ocsp_uri" ]]; then
        ocsp_uri=$(get_ocsp_uri "$cert_file")
    fi

    if [[ -z "$ocsp_uri" ]]; then
        echo "$OCSP_ERROR: No OCSP URI found"
        return 1
    fi

    # Make OCSP request
    local result
    result=$(openssl ocsp \
        -issuer "$issuer_cert" \
        -cert "$cert_file" \
        -url "$ocsp_uri" \
        -resp_text \
        2>&1)

    local status=$?

    if [[ $status -ne 0 ]]; then
        echo "$OCSP_ERROR: OCSP request failed"
        echo "$result"
        return 1
    fi

    # Parse response
    if echo "$result" | grep -q ": good"; then
        echo "$OCSP_GOOD"
        return 0
    elif echo "$result" | grep -q ": revoked"; then
        echo "$OCSP_REVOKED"
        # Extract revocation details
        local revocation_time=$(echo "$result" | grep "Revocation Time:" | sed 's/.*Revocation Time: //')
        local revocation_reason=$(echo "$result" | grep "Revocation Reason:" | sed 's/.*Revocation Reason: //')
        echo "Revocation Time: $revocation_time"
        [[ -n "$revocation_reason" ]] && echo "Revocation Reason: $revocation_reason"
        return 2
    else
        echo "$OCSP_UNKNOWN"
        return 3
    fi
}

# Verify OCSP with stapled response
# Usage: verify_ocsp_staple <host> [port]
verify_ocsp_staple() {
    local host="$1"
    local port="${2:-443}"

    local temp_dir=$(mktemp -d)

    # Get certificate chain from server
    echo | openssl s_client -connect "$host:$port" -showcerts 2>/dev/null | \
        sed -n '/-----BEGIN CERTIFICATE-----/,/-----END CERTIFICATE-----/p' > "$temp_dir/chain.pem"

    # Split certificates
    awk -v dir="$temp_dir" '
        /-----BEGIN CERTIFICATE-----/ { cert_num++; in_cert=1 }
        in_cert { print > (dir "/cert" cert_num ".pem") }
        /-----END CERTIFICATE-----/ { in_cert=0 }
    ' "$temp_dir/chain.pem"

    # Check if we have at least 2 certs (leaf + issuer)
    if [[ ! -f "$temp_dir/cert1.pem" ]] || [[ ! -f "$temp_dir/cert2.pem" ]]; then
        rm -rf "$temp_dir"
        echo "$OCSP_ERROR: Could not retrieve certificate chain"
        return 1
    fi

    # Check OCSP stapling support
    local stapling=$(check_ocsp_stapling "$host" "$port")

    echo "OCSP Stapling: $stapling"

    # If stapling is supported, verify the stapled response
    if [[ "$stapling" == "SUPPORTED" ]]; then
        echo ""
        echo "Stapled Response:"
        get_ocsp_staple "$host" "$port"
    fi

    # Also do direct OCSP check
    echo ""
    echo "Direct OCSP Verification:"
    local ocsp_result=$(verify_ocsp "$temp_dir/cert1.pem" "$temp_dir/cert2.pem")
    echo "$ocsp_result"

    rm -rf "$temp_dir"
}

# Get OCSP response details
# Usage: get_ocsp_response_details <host> [port]
get_ocsp_response_details() {
    local host="$1"
    local port="${2:-443}"

    local temp_dir=$(mktemp -d)

    # Get certificates
    echo | openssl s_client -connect "$host:$port" -showcerts 2>/dev/null | \
        sed -n '/-----BEGIN CERTIFICATE-----/,/-----END CERTIFICATE-----/p' > "$temp_dir/chain.pem"

    # Split
    awk -v dir="$temp_dir" '
        /-----BEGIN CERTIFICATE-----/ { cert_num++; in_cert=1 }
        in_cert { print > (dir "/cert" cert_num ".pem") }
        /-----END CERTIFICATE-----/ { in_cert=0 }
    ' "$temp_dir/chain.pem"

    if [[ ! -f "$temp_dir/cert1.pem" ]] || [[ ! -f "$temp_dir/cert2.pem" ]]; then
        rm -rf "$temp_dir"
        echo "Error: Could not retrieve certificates"
        return 1
    fi

    local ocsp_uri=$(get_ocsp_uri "$temp_dir/cert1.pem")

    if [[ -z "$ocsp_uri" ]]; then
        rm -rf "$temp_dir"
        echo "No OCSP URI found in certificate"
        return 1
    fi

    echo "OCSP URI: $ocsp_uri"
    echo ""

    # Get full OCSP response
    openssl ocsp \
        -issuer "$temp_dir/cert2.pem" \
        -cert "$temp_dir/cert1.pem" \
        -url "$ocsp_uri" \
        -resp_text \
        2>&1

    rm -rf "$temp_dir"
}

# Check OCSP must-staple extension
# Usage: has_ocsp_must_staple <cert_file>
has_ocsp_must_staple() {
    local cert_file="$1"

    # OCSP Must-Staple is OID 1.3.6.1.5.5.7.1.24
    if openssl x509 -in "$cert_file" -noout -text 2>/dev/null | grep -q "1.3.6.1.5.5.7.1.24"; then
        echo "YES"
        return 0
    else
        echo "NO"
        return 1
    fi
}

# Get OCSP verification summary
# Usage: get_ocsp_summary <cert_file> <issuer_cert> [host] [port]
get_ocsp_summary() {
    local cert_file="$1"
    local issuer_cert="$2"
    local host="${3:-}"
    local port="${4:-443}"

    local ocsp_uri=$(get_ocsp_uri "$cert_file")
    local must_staple=$(has_ocsp_must_staple "$cert_file")

    echo "OCSP Verification Summary"
    echo "========================="
    echo ""
    echo "OCSP URI: ${ocsp_uri:-Not found}"
    echo "OCSP Must-Staple: $must_staple"

    if [[ -n "$host" ]]; then
        local stapling=$(check_ocsp_stapling "$host" "$port")
        echo "Server OCSP Stapling: $stapling"

        if [[ "$must_staple" == "YES" ]] && [[ "$stapling" != "SUPPORTED" ]]; then
            echo ""
            echo "⚠️  WARNING: Certificate has OCSP Must-Staple but server does not support stapling!"
        fi
    fi

    if [[ -n "$ocsp_uri" ]] && [[ -f "$issuer_cert" ]]; then
        echo ""
        echo "OCSP Status:"
        verify_ocsp "$cert_file" "$issuer_cert" "$ocsp_uri"
    fi
}

# Parse OCSP response status
# Usage: parse_ocsp_status <ocsp_output>
parse_ocsp_status() {
    local ocsp_output="$1"

    if echo "$ocsp_output" | grep -q "Cert Status: good"; then
        echo "good"
    elif echo "$ocsp_output" | grep -q "Cert Status: revoked"; then
        echo "revoked"
    elif echo "$ocsp_output" | grep -q "Cert Status: unknown"; then
        echo "unknown"
    else
        echo "error"
    fi
}

# Check OCSP response validity period
# Usage: check_ocsp_response_validity <host> [port]
check_ocsp_response_validity() {
    local host="$1"
    local port="${2:-443}"

    local staple=$(get_ocsp_staple "$host" "$port")

    if [[ -z "$staple" ]]; then
        echo "No OCSP staple available"
        return 1
    fi

    # Extract validity times
    local produced=$(echo "$staple" | grep "Produced At:" | sed 's/.*Produced At: //')
    local this_update=$(echo "$staple" | grep "This Update:" | sed 's/.*This Update: //')
    local next_update=$(echo "$staple" | grep "Next Update:" | sed 's/.*Next Update: //')

    echo "OCSP Response Validity"
    echo "====================="
    echo "Produced At: $produced"
    echo "This Update: $this_update"
    echo "Next Update: $next_update"

    # Check if response is still valid (Next Update in future)
    if [[ -n "$next_update" ]]; then
        local next_epoch
        # Try GNU date first, then macOS
        next_epoch=$(date -d "$next_update" +%s 2>/dev/null || date -j -f "%b %d %H:%M:%S %Y %Z" "$next_update" +%s 2>/dev/null)
        local now_epoch=$(date +%s)

        if [[ -n "$next_epoch" ]] && [[ $next_epoch -gt $now_epoch ]]; then
            local remaining=$(( (next_epoch - now_epoch) / 3600 ))
            echo ""
            echo "Status: ✓ Valid (expires in ${remaining} hours)"
        else
            echo ""
            echo "Status: ✗ Expired or invalid"
        fi
    fi
}

# Export functions
export -f get_ocsp_uri
export -f check_ocsp_stapling
export -f get_ocsp_staple
export -f verify_ocsp
export -f verify_ocsp_staple
export -f get_ocsp_response_details
export -f has_ocsp_must_staple
export -f get_ocsp_summary
export -f parse_ocsp_status
export -f check_ocsp_response_validity
