#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Certificate Format Detection Library
# Auto-detects certificate formats: PEM, DER, PKCS7, PKCS12
#############################################################################

# Format constants
FORMAT_PEM="pem"
FORMAT_DER="der"
FORMAT_PKCS7="pkcs7"
FORMAT_PKCS12="pkcs12"
FORMAT_UNKNOWN="unknown"

# Detect certificate format from file
# Usage: detect_cert_format <file_path>
# Returns: pem|der|pkcs7|pkcs12|unknown
detect_cert_format() {
    local file="$1"

    if [[ ! -f "$file" ]]; then
        echo "$FORMAT_UNKNOWN"
        return 1
    fi

    # First, check by file extension
    local ext_format=$(detect_format_by_extension "$file")

    # Then verify by content (magic bytes / structure)
    local content_format=$(detect_format_by_content "$file")

    # If content detection succeeded, use that (more reliable)
    if [[ "$content_format" != "$FORMAT_UNKNOWN" ]]; then
        echo "$content_format"
        return 0
    fi

    # Fall back to extension-based detection
    if [[ "$ext_format" != "$FORMAT_UNKNOWN" ]]; then
        echo "$ext_format"
        return 0
    fi

    echo "$FORMAT_UNKNOWN"
    return 1
}

# Detect format by file extension
# Usage: detect_format_by_extension <file_path>
detect_format_by_extension() {
    local file="$1"
    local filename=$(basename "$file")
    local ext="${filename##*.}"

    # Convert to lowercase
    ext=$(echo "$ext" | tr '[:upper:]' '[:lower:]')

    case "$ext" in
        pem|crt|cer|cert)
            echo "$FORMAT_PEM"
            ;;
        der)
            echo "$FORMAT_DER"
            ;;
        p7b|p7c|pkcs7)
            echo "$FORMAT_PKCS7"
            ;;
        p12|pfx|pkcs12)
            echo "$FORMAT_PKCS12"
            ;;
        *)
            echo "$FORMAT_UNKNOWN"
            ;;
    esac
}

# Detect format by file content (magic bytes / structure)
# Usage: detect_format_by_content <file_path>
detect_format_by_content() {
    local file="$1"

    # Check for PEM format (text-based with headers)
    if is_pem_format "$file"; then
        echo "$FORMAT_PEM"
        return 0
    fi

    # Check for PKCS12 format (specific ASN.1 structure)
    if is_pkcs12_format "$file"; then
        echo "$FORMAT_PKCS12"
        return 0
    fi

    # Check for PKCS7 format
    if is_pkcs7_format "$file"; then
        echo "$FORMAT_PKCS7"
        return 0
    fi

    # Check for DER format (binary ASN.1)
    if is_der_format "$file"; then
        echo "$FORMAT_DER"
        return 0
    fi

    echo "$FORMAT_UNKNOWN"
    return 1
}

# Check if file is PEM format (certificate, not PKCS7)
# PEM files contain Base64-encoded data between -----BEGIN and -----END markers
is_pem_format() {
    local file="$1"

    # Check for PEM certificate headers (but not PKCS7)
    if grep -q "^-----BEGIN CERTIFICATE-----" "$file" 2>/dev/null; then
        return 0
    fi

    # Also check for other PEM types (private key, public key)
    if grep -q "^-----BEGIN.*PRIVATE KEY-----" "$file" 2>/dev/null; then
        return 0
    fi

    if grep -q "^-----BEGIN PUBLIC KEY-----" "$file" 2>/dev/null; then
        return 0
    fi

    return 1
}

# Check if file is DER format (binary X.509 certificate)
# DER files start with ASN.1 SEQUENCE tag (0x30)
is_der_format() {
    local file="$1"

    # Get first byte
    local first_byte=$(xxd -p -l 1 "$file" 2>/dev/null)

    # ASN.1 SEQUENCE tag is 0x30
    if [[ "$first_byte" == "30" ]]; then
        # Try to parse as X.509 certificate
        if openssl x509 -inform DER -in "$file" -noout 2>/dev/null; then
            return 0
        fi
    fi

    return 1
}

# Check if file is PKCS7 format
# PKCS7 files have specific OID in ASN.1 structure
is_pkcs7_format() {
    local file="$1"

    # First check if it's PEM-encoded PKCS7
    if grep -q "^-----BEGIN PKCS7-----" "$file" 2>/dev/null; then
        return 0
    fi

    if grep -q "^-----BEGIN CERTIFICATE-----" "$file" 2>/dev/null; then
        # Could be a certificate bundle - not PKCS7
        return 1
    fi

    # Check for DER-encoded PKCS7
    # PKCS7 OID: 1.2.840.113549.1.7 (starts with 06 09 2a 86 48 86 f7 0d 01 07)
    local first_bytes=$(xxd -p -l 20 "$file" 2>/dev/null | tr -d '\n')

    # Check for ASN.1 SEQUENCE containing PKCS7 OID
    if [[ "$first_bytes" == "30"* ]]; then
        # Try to parse as PKCS7
        if openssl pkcs7 -inform DER -in "$file" -print_certs -noout 2>/dev/null; then
            return 0
        fi
    fi

    return 1
}

# Check if file is PKCS12 format
# PKCS12 files have specific structure with version and auth safe
is_pkcs12_format() {
    local file="$1"

    # Get first few bytes
    local first_bytes=$(xxd -p -l 4 "$file" 2>/dev/null | tr -d '\n')

    # PKCS12 starts with ASN.1 SEQUENCE (0x30)
    if [[ "$first_bytes" != "30"* ]]; then
        return 1
    fi

    # Try to parse as PKCS12 (will fail with bad password but confirms format)
    local result=$(openssl pkcs12 -in "$file" -info -noout -passin pass: 2>&1)

    # Check for PKCS12-specific errors or success indicators
    if echo "$result" | grep -qE "MAC:|PKCS7|shrouded|bag"; then
        return 0
    fi

    # Also check for password-related errors (confirms it's PKCS12)
    if echo "$result" | grep -qE "mac verify failure|invalid password"; then
        return 0
    fi

    return 1
}

# Get PEM certificate type from header
# Usage: get_pem_type <file_path>
# Returns: CERTIFICATE|PRIVATE KEY|PUBLIC KEY|etc.
get_pem_type() {
    local file="$1"

    grep "^-----BEGIN" "$file" 2>/dev/null | head -1 | sed 's/-----BEGIN \(.*\)-----/\1/'
}

# Count certificates in a file
# Usage: count_certificates <file_path> [format]
count_certificates() {
    local file="$1"
    local format="${2:-}"

    # Auto-detect format if not specified
    if [[ -z "$format" ]]; then
        format=$(detect_cert_format "$file")
    fi

    case "$format" in
        pem)
            grep -c "^-----BEGIN CERTIFICATE-----" "$file" 2>/dev/null || echo "0"
            ;;
        der)
            # DER format contains exactly one certificate
            echo "1"
            ;;
        pkcs7)
            # Count certificates in PKCS7 bundle
            local count=$(openssl pkcs7 -in "$file" -print_certs 2>/dev/null | grep -c "^-----BEGIN CERTIFICATE-----")
            echo "${count:-0}"
            ;;
        pkcs12)
            # Count certificates in PKCS12 (requires password)
            # This is a rough estimate - actual count needs password
            echo "1+"
            ;;
        *)
            echo "0"
            ;;
    esac
}

# Detect if input is from stdin
# Usage: is_stdin_input
is_stdin_input() {
    [[ ! -t 0 ]]
}

# Read certificate from stdin and save to temp file
# Usage: read_stdin_to_temp
# Returns: path to temp file
read_stdin_to_temp() {
    local temp_file=$(mktemp)
    cat > "$temp_file"
    echo "$temp_file"
}

# Validate that file contains valid certificate data
# Usage: validate_cert_file <file_path> [format]
validate_cert_file() {
    local file="$1"
    local format="${2:-}"

    if [[ ! -f "$file" ]]; then
        echo "Error: File not found: $file" >&2
        return 1
    fi

    if [[ ! -r "$file" ]]; then
        echo "Error: File not readable: $file" >&2
        return 1
    fi

    if [[ ! -s "$file" ]]; then
        echo "Error: File is empty: $file" >&2
        return 1
    fi

    # Auto-detect format if not specified
    if [[ -z "$format" ]]; then
        format=$(detect_cert_format "$file")
    fi

    if [[ "$format" == "$FORMAT_UNKNOWN" ]]; then
        echo "Error: Unable to detect certificate format" >&2
        return 1
    fi

    return 0
}

# Get human-readable format name
# Usage: format_name <format_code>
format_name() {
    local format="$1"

    case "$format" in
        pem)
            echo "PEM (Base64-encoded)"
            ;;
        der)
            echo "DER (Binary ASN.1)"
            ;;
        pkcs7)
            echo "PKCS#7 Bundle"
            ;;
        pkcs12)
            echo "PKCS#12 Keystore"
            ;;
        *)
            echo "Unknown"
            ;;
    esac
}

# Export functions
export -f detect_cert_format
export -f detect_format_by_extension
export -f detect_format_by_content
export -f is_pem_format
export -f is_der_format
export -f is_pkcs7_format
export -f is_pkcs12_format
export -f get_pem_type
export -f count_certificates
export -f is_stdin_input
export -f read_stdin_to_temp
export -f validate_cert_file
export -f format_name
