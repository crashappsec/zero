#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Certificate Format Parsers Library
# Parse certificates from various formats: PEM, DER, PKCS7, PKCS12
#############################################################################

# Source format detection if not already loaded
_FORMAT_PARSERS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
if ! type detect_cert_format &>/dev/null; then
    source "$_FORMAT_PARSERS_DIR/format-detection.sh"
fi

#############################################################################
# PEM Format Parsing
#############################################################################

# Extract certificates from PEM file
# Usage: parse_pem_certificates <file_path> <output_dir>
# Creates: cert1.pem, cert2.pem, etc. in output_dir
parse_pem_certificates() {
    local file="$1"
    local output_dir="$2"

    if [[ ! -d "$output_dir" ]]; then
        mkdir -p "$output_dir"
    fi

    # Split PEM file into individual certificates
    awk -v outdir="$output_dir" '
        /-----BEGIN CERTIFICATE-----/ {
            cert_num++
            in_cert = 1
        }
        in_cert {
            print > (outdir "/cert" cert_num ".pem")
        }
        /-----END CERTIFICATE-----/ {
            in_cert = 0
        }
    ' "$file"

    # Return count of certificates extracted
    local count=$(ls -1 "$output_dir"/cert*.pem 2>/dev/null | wc -l | tr -d ' ')
    echo "$count"
}

# Extract private key from PEM file (if present)
# Usage: extract_pem_private_key <file_path> <output_file>
extract_pem_private_key() {
    local file="$1"
    local output_file="$2"

    awk '
        /-----BEGIN.*PRIVATE KEY-----/,/-----END.*PRIVATE KEY-----/ { print }
    ' "$file" > "$output_file"

    if [[ -s "$output_file" ]]; then
        return 0
    else
        rm -f "$output_file"
        return 1
    fi
}

# Get all PEM blocks from file with their types
# Usage: list_pem_blocks <file_path>
# Output: TYPE|START_LINE|END_LINE for each block
list_pem_blocks() {
    local file="$1"

    awk '
        /-----BEGIN/ {
            start = NR
            match($0, /-----BEGIN ([^-]+)-----/, arr)
            type = arr[1]
        }
        /-----END/ {
            print type "|" start "|" NR
        }
    ' "$file"
}

#############################################################################
# DER Format Parsing
#############################################################################

# Convert DER certificate to PEM
# Usage: der_to_pem <der_file> <pem_file>
der_to_pem() {
    local der_file="$1"
    local pem_file="$2"

    if openssl x509 -inform DER -in "$der_file" -outform PEM -out "$pem_file" 2>/dev/null; then
        return 0
    fi

    return 1
}

# Parse DER certificate and extract to output directory
# Usage: parse_der_certificate <file_path> <output_dir>
parse_der_certificate() {
    local file="$1"
    local output_dir="$2"

    if [[ ! -d "$output_dir" ]]; then
        mkdir -p "$output_dir"
    fi

    # Convert DER to PEM
    if der_to_pem "$file" "$output_dir/cert1.pem"; then
        echo "1"
        return 0
    fi

    echo "0"
    return 1
}

#############################################################################
# PKCS7 Format Parsing
#############################################################################

# Extract certificates from PKCS7 bundle
# Usage: parse_pkcs7_certificates <file_path> <output_dir>
parse_pkcs7_certificates() {
    local file="$1"
    local output_dir="$2"

    if [[ ! -d "$output_dir" ]]; then
        mkdir -p "$output_dir"
    fi

    # Detect if PEM or DER encoded PKCS7
    local inform="PEM"
    if ! grep -q "^-----BEGIN" "$file" 2>/dev/null; then
        inform="DER"
    fi

    # Extract all certificates from PKCS7
    local temp_pem=$(mktemp)
    if ! openssl pkcs7 -inform "$inform" -in "$file" -print_certs -out "$temp_pem" 2>/dev/null; then
        rm -f "$temp_pem"
        echo "0"
        return 1
    fi

    # Split into individual certificates
    local count=$(parse_pem_certificates "$temp_pem" "$output_dir")
    rm -f "$temp_pem"

    echo "$count"
}

# Get PKCS7 content type
# Usage: get_pkcs7_content_type <file_path>
get_pkcs7_content_type() {
    local file="$1"

    local inform="PEM"
    if ! grep -q "^-----BEGIN" "$file" 2>/dev/null; then
        inform="DER"
    fi

    openssl pkcs7 -inform "$inform" -in "$file" -print -noout 2>/dev/null | grep -i "content:" | head -1 | sed 's/.*: //'
}

#############################################################################
# PKCS12 Format Parsing
#############################################################################

# Extract certificates from PKCS12 keystore
# Usage: parse_pkcs12_certificates <file_path> <output_dir> [password]
# Note: Password can be provided via argument, PKCS12_PASSWORD env var, or password file
parse_pkcs12_certificates() {
    local file="$1"
    local output_dir="$2"
    local password="${3:-${PKCS12_PASSWORD:-}}"

    if [[ ! -d "$output_dir" ]]; then
        mkdir -p "$output_dir"
    fi

    # Build password argument
    local pass_arg=""
    if [[ -n "$password" ]]; then
        pass_arg="-passin pass:$password"
    else
        pass_arg="-passin pass:"
    fi

    # Extract certificates
    local temp_pem=$(mktemp)
    if ! openssl pkcs12 -in "$file" $pass_arg -nokeys -out "$temp_pem" 2>/dev/null; then
        rm -f "$temp_pem"
        echo "0"
        return 1
    fi

    # Split into individual certificates
    local count=$(parse_pem_certificates "$temp_pem" "$output_dir")
    rm -f "$temp_pem"

    echo "$count"
}

# Extract private key from PKCS12 keystore
# Usage: extract_pkcs12_private_key <file_path> <output_file> [password]
extract_pkcs12_private_key() {
    local file="$1"
    local output_file="$2"
    local password="${3:-${PKCS12_PASSWORD:-}}"

    local pass_arg=""
    if [[ -n "$password" ]]; then
        pass_arg="-passin pass:$password"
    else
        pass_arg="-passin pass:"
    fi

    if openssl pkcs12 -in "$file" $pass_arg -nocerts -nodes -out "$output_file" 2>/dev/null; then
        if [[ -s "$output_file" ]]; then
            return 0
        fi
    fi

    rm -f "$output_file"
    return 1
}

# Get PKCS12 info without extracting
# Usage: get_pkcs12_info <file_path> [password]
get_pkcs12_info() {
    local file="$1"
    local password="${2:-${PKCS12_PASSWORD:-}}"

    local pass_arg=""
    if [[ -n "$password" ]]; then
        pass_arg="-passin pass:$password"
    else
        pass_arg="-passin pass:"
    fi

    openssl pkcs12 -in "$file" $pass_arg -info -noout 2>&1
}

# Check if PKCS12 password is correct
# Usage: verify_pkcs12_password <file_path> [password]
verify_pkcs12_password() {
    local file="$1"
    local password="${2:-${PKCS12_PASSWORD:-}}"

    local pass_arg=""
    if [[ -n "$password" ]]; then
        pass_arg="-passin pass:$password"
    else
        pass_arg="-passin pass:"
    fi

    if openssl pkcs12 -in "$file" $pass_arg -info -noout 2>&1 | grep -q "MAC:"; then
        return 0
    fi

    return 1
}

#############################################################################
# Generic Certificate Parsing
#############################################################################

# Parse certificates from any supported format
# Usage: parse_certificates <file_path> <output_dir> [format] [password]
# Returns: Number of certificates extracted
parse_certificates() {
    local file="$1"
    local output_dir="$2"
    local format="${3:-}"
    local password="${4:-${PKCS12_PASSWORD:-}}"

    # Auto-detect format if not specified
    if [[ -z "$format" ]]; then
        format=$(detect_cert_format "$file")
    fi

    case "$format" in
        pem)
            parse_pem_certificates "$file" "$output_dir"
            ;;
        der)
            parse_der_certificate "$file" "$output_dir"
            ;;
        pkcs7)
            parse_pkcs7_certificates "$file" "$output_dir"
            ;;
        pkcs12)
            parse_pkcs12_certificates "$file" "$output_dir" "$password"
            ;;
        *)
            echo "0"
            return 1
            ;;
    esac
}

# Convert any format to PEM
# Usage: convert_to_pem <input_file> <output_file> [format] [password]
convert_to_pem() {
    local input_file="$1"
    local output_file="$2"
    local format="${3:-}"
    local password="${4:-${PKCS12_PASSWORD:-}}"

    # Auto-detect format if not specified
    if [[ -z "$format" ]]; then
        format=$(detect_cert_format "$input_file")
    fi

    case "$format" in
        pem)
            # Already PEM, just copy certificates
            grep -A 1000 "^-----BEGIN CERTIFICATE-----" "$input_file" | \
                grep -B 1000 "^-----END CERTIFICATE-----" > "$output_file"
            ;;
        der)
            der_to_pem "$input_file" "$output_file"
            ;;
        pkcs7)
            local inform="PEM"
            if ! grep -q "^-----BEGIN" "$input_file" 2>/dev/null; then
                inform="DER"
            fi
            openssl pkcs7 -inform "$inform" -in "$input_file" -print_certs -out "$output_file" 2>/dev/null
            ;;
        pkcs12)
            local pass_arg=""
            if [[ -n "$password" ]]; then
                pass_arg="-passin pass:$password"
            else
                pass_arg="-passin pass:"
            fi
            openssl pkcs12 -in "$input_file" $pass_arg -nokeys -out "$output_file" 2>/dev/null
            ;;
        *)
            return 1
            ;;
    esac

    [[ -s "$output_file" ]]
}

# Read password from file
# Usage: read_password_file <file_path>
read_password_file() {
    local file="$1"

    if [[ -f "$file" ]] && [[ -r "$file" ]]; then
        # Read first line, strip newlines
        head -1 "$file" | tr -d '\n\r'
    fi
}

# Get certificate chain from parsed certificates
# Usage: get_certificate_chain <certs_dir>
# Returns: Ordered list of cert files (leaf first, root last)
get_certificate_chain() {
    local certs_dir="$1"
    local -a chain=()
    local -a certs=()

    # Get all cert files
    for cert_file in "$certs_dir"/cert*.pem; do
        [[ -f "$cert_file" ]] && certs+=("$cert_file")
    done

    if [[ ${#certs[@]} -eq 0 ]]; then
        return 1
    fi

    if [[ ${#certs[@]} -eq 1 ]]; then
        echo "${certs[0]}"
        return 0
    fi

    # Build chain by matching issuer to subject
    # Find leaf certificate (subject != issuer and not CA)
    local leaf=""
    for cert in "${certs[@]}"; do
        local subject=$(openssl x509 -in "$cert" -noout -subject -nameopt RFC2253 2>/dev/null | sed 's/subject=//')
        local issuer=$(openssl x509 -in "$cert" -noout -issuer -nameopt RFC2253 2>/dev/null | sed 's/issuer=//')
        local is_ca=$(openssl x509 -in "$cert" -noout -text 2>/dev/null | grep -c "CA:TRUE")

        if [[ "$subject" != "$issuer" ]] && [[ "$is_ca" -eq 0 ]]; then
            leaf="$cert"
            break
        fi
    done

    # If no leaf found, use first cert
    if [[ -z "$leaf" ]]; then
        leaf="${certs[0]}"
    fi

    echo "$leaf"
    chain+=("$leaf")

    # Follow chain upward
    local current="$leaf"
    while true; do
        local current_issuer=$(openssl x509 -in "$current" -noout -issuer -nameopt RFC2253 2>/dev/null | sed 's/issuer=//')
        local found_next=false

        for cert in "${certs[@]}"; do
            [[ "$cert" == "$current" ]] && continue
            [[ " ${chain[*]} " =~ " $cert " ]] && continue

            local cert_subject=$(openssl x509 -in "$cert" -noout -subject -nameopt RFC2253 2>/dev/null | sed 's/subject=//')

            if [[ "$cert_subject" == "$current_issuer" ]]; then
                echo "$cert"
                chain+=("$cert")
                current="$cert"
                found_next=true
                break
            fi
        done

        if [[ "$found_next" == false ]]; then
            break
        fi
    done
}

# Export functions
export -f parse_pem_certificates
export -f extract_pem_private_key
export -f list_pem_blocks
export -f der_to_pem
export -f parse_der_certificate
export -f parse_pkcs7_certificates
export -f get_pkcs7_content_type
export -f parse_pkcs12_certificates
export -f extract_pkcs12_private_key
export -f get_pkcs12_info
export -f verify_pkcs12_password
export -f parse_certificates
export -f convert_to_pem
export -f read_password_file
export -f get_certificate_chain
