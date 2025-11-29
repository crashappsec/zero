#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Certificate Fingerprint Library
# Generate and verify certificate fingerprints (SHA-1, SHA-256, SHA-384, SHA-512)
#############################################################################

# Supported hash algorithms
HASH_SHA1="sha1"
HASH_SHA256="sha256"
HASH_SHA384="sha384"
HASH_SHA512="sha512"

# Generate certificate fingerprint
# Usage: get_cert_fingerprint <cert_file> [algorithm]
# Default algorithm: sha256
get_cert_fingerprint() {
    local cert_file="$1"
    local algorithm="${2:-$HASH_SHA256}"

    # Validate algorithm
    case "$algorithm" in
        sha1|sha256|sha384|sha512)
            ;;
        *)
            echo "Error: Unsupported algorithm: $algorithm" >&2
            return 1
            ;;
    esac

    # Generate fingerprint
    local fingerprint=$(openssl x509 -in "$cert_file" -noout -fingerprint -"$algorithm" 2>/dev/null)

    if [[ -z "$fingerprint" ]]; then
        return 1
    fi

    # Extract just the hash value (remove "SHA256 Fingerprint=" prefix)
    echo "$fingerprint" | sed 's/.*Fingerprint=//' | tr -d ':'
}

# Generate all fingerprints for a certificate
# Usage: get_all_fingerprints <cert_file>
# Output: JSON object with all fingerprints
get_all_fingerprints() {
    local cert_file="$1"

    local sha1=$(get_cert_fingerprint "$cert_file" "sha1")
    local sha256=$(get_cert_fingerprint "$cert_file" "sha256")
    local sha384=$(get_cert_fingerprint "$cert_file" "sha384")
    local sha512=$(get_cert_fingerprint "$cert_file" "sha512")

    cat << EOF
{
  "sha1": "$sha1",
  "sha256": "$sha256",
  "sha384": "$sha384",
  "sha512": "$sha512"
}
EOF
}

# Generate fingerprint with formatted output (colons)
# Usage: get_formatted_fingerprint <cert_file> [algorithm]
get_formatted_fingerprint() {
    local cert_file="$1"
    local algorithm="${2:-$HASH_SHA256}"

    openssl x509 -in "$cert_file" -noout -fingerprint -"$algorithm" 2>/dev/null | \
        sed 's/.*Fingerprint=//'
}

# Compare two certificate fingerprints
# Usage: compare_fingerprints <cert1_file> <cert2_file> [algorithm]
# Returns: 0 if match, 1 if different
compare_fingerprints() {
    local cert1="$1"
    local cert2="$2"
    local algorithm="${3:-$HASH_SHA256}"

    local fp1=$(get_cert_fingerprint "$cert1" "$algorithm")
    local fp2=$(get_cert_fingerprint "$cert2" "$algorithm")

    if [[ -z "$fp1" ]] || [[ -z "$fp2" ]]; then
        return 2
    fi

    [[ "$fp1" == "$fp2" ]]
}

# Verify certificate against known fingerprint
# Usage: verify_fingerprint <cert_file> <expected_fingerprint> [algorithm]
verify_fingerprint() {
    local cert_file="$1"
    local expected="$2"
    local algorithm="${3:-$HASH_SHA256}"

    # Normalize expected fingerprint (remove colons, convert to uppercase)
    expected=$(echo "$expected" | tr -d ':' | tr '[:lower:]' '[:upper:]')

    local actual=$(get_cert_fingerprint "$cert_file" "$algorithm" | tr '[:lower:]' '[:upper:]')

    [[ "$expected" == "$actual" ]]
}

# Generate public key fingerprint (SPKI)
# Usage: get_pubkey_fingerprint <cert_file> [algorithm]
get_pubkey_fingerprint() {
    local cert_file="$1"
    local algorithm="${2:-$HASH_SHA256}"

    openssl x509 -in "$cert_file" -pubkey -noout 2>/dev/null | \
        openssl pkey -pubin -outform DER 2>/dev/null | \
        openssl dgst -"$algorithm" 2>/dev/null | \
        awk '{print $2}' | \
        tr '[:lower:]' '[:upper:]'
}

# Generate HPKP pin (Base64-encoded SPKI hash)
# Usage: get_hpkp_pin <cert_file> [algorithm]
get_hpkp_pin() {
    local cert_file="$1"
    local algorithm="${2:-$HASH_SHA256}"

    openssl x509 -in "$cert_file" -pubkey -noout 2>/dev/null | \
        openssl pkey -pubin -outform DER 2>/dev/null | \
        openssl dgst -"$algorithm" -binary 2>/dev/null | \
        openssl enc -base64
}

# Format fingerprint for display (add colons)
# Usage: format_fingerprint <fingerprint>
format_fingerprint() {
    local fp="$1"
    echo "$fp" | sed 's/\(..\)/\1:/g' | sed 's/:$//'
}

# Get fingerprint output in various formats
# Usage: get_fingerprint_formatted <cert_file> <format> [algorithm]
# Formats: raw, colon, lower, upper
get_fingerprint_formatted() {
    local cert_file="$1"
    local format="$2"
    local algorithm="${3:-$HASH_SHA256}"

    local fp=$(get_cert_fingerprint "$cert_file" "$algorithm")

    case "$format" in
        raw)
            echo "$fp"
            ;;
        colon)
            format_fingerprint "$fp"
            ;;
        lower)
            echo "$fp" | tr '[:upper:]' '[:lower:]'
            ;;
        upper)
            echo "$fp" | tr '[:lower:]' '[:upper:]'
            ;;
        *)
            echo "$fp"
            ;;
    esac
}

# Generate fingerprint table for certificate
# Usage: print_fingerprint_table <cert_file>
print_fingerprint_table() {
    local cert_file="$1"

    echo "Certificate Fingerprints:"
    echo "========================="
    echo ""
    echo "SHA-1:   $(get_formatted_fingerprint "$cert_file" sha1)"
    echo "SHA-256: $(get_formatted_fingerprint "$cert_file" sha256)"
    echo "SHA-384: $(get_formatted_fingerprint "$cert_file" sha384)"
    echo "SHA-512: $(get_formatted_fingerprint "$cert_file" sha512)"
    echo ""
    echo "SPKI SHA-256 Pin (HPKP): $(get_hpkp_pin "$cert_file" sha256)"
}

# Export functions
export -f get_cert_fingerprint
export -f get_all_fingerprints
export -f get_formatted_fingerprint
export -f compare_fingerprints
export -f verify_fingerprint
export -f get_pubkey_fingerprint
export -f get_hpkp_pin
export -f format_fingerprint
export -f get_fingerprint_formatted
export -f print_fingerprint_table
