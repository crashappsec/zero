#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Certificate Chain Validation Library
# Validates certificate chains against trust stores
#############################################################################

# Default trust store locations
SYSTEM_TRUST_STORE=""

# Detect system trust store location
detect_trust_store() {
    # macOS
    if [[ -f "/etc/ssl/cert.pem" ]]; then
        SYSTEM_TRUST_STORE="/etc/ssl/cert.pem"
    # Ubuntu/Debian
    elif [[ -f "/etc/ssl/certs/ca-certificates.crt" ]]; then
        SYSTEM_TRUST_STORE="/etc/ssl/certs/ca-certificates.crt"
    # RHEL/CentOS
    elif [[ -f "/etc/pki/tls/certs/ca-bundle.crt" ]]; then
        SYSTEM_TRUST_STORE="/etc/pki/tls/certs/ca-bundle.crt"
    # Alpine
    elif [[ -f "/etc/ssl/certs/ca-certificates.crt" ]]; then
        SYSTEM_TRUST_STORE="/etc/ssl/certs/ca-certificates.crt"
    # FreeBSD
    elif [[ -f "/usr/local/share/certs/ca-root-nss.crt" ]]; then
        SYSTEM_TRUST_STORE="/usr/local/share/certs/ca-root-nss.crt"
    fi

    echo "$SYSTEM_TRUST_STORE"
}

# Validate certificate chain
# Usage: validate_chain <leaf_cert> [intermediate_cert...] [--ca-file <ca_bundle>]
# Returns: 0 if valid, 1 if invalid
# Output: Validation result details
validate_chain() {
    local leaf_cert=""
    local intermediates=()
    local ca_file=""
    local verbose=false

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --ca-file)
                ca_file="$2"
                shift 2
                ;;
            --verbose)
                verbose=true
                shift
                ;;
            *)
                if [[ -z "$leaf_cert" ]]; then
                    leaf_cert="$1"
                else
                    intermediates+=("$1")
                fi
                shift
                ;;
        esac
    done

    if [[ -z "$leaf_cert" ]]; then
        echo "Error: No certificate provided" >&2
        return 1
    fi

    # Use system trust store if no CA file specified
    if [[ -z "$ca_file" ]]; then
        ca_file=$(detect_trust_store)
    fi

    # Build verification command
    local verify_cmd="openssl verify"

    if [[ -n "$ca_file" ]] && [[ -f "$ca_file" ]]; then
        verify_cmd="$verify_cmd -CAfile $ca_file"
    fi

    # Add untrusted intermediates
    for intermediate in "${intermediates[@]}"; do
        if [[ -f "$intermediate" ]]; then
            verify_cmd="$verify_cmd -untrusted $intermediate"
        fi
    done

    # Run verification
    local result
    result=$($verify_cmd "$leaf_cert" 2>&1)
    local status=$?

    if [[ $status -eq 0 ]]; then
        echo "VALID"
        if [[ "$verbose" == true ]]; then
            echo "Chain verification successful"
            echo "$result"
        fi
        return 0
    else
        echo "INVALID"
        echo "$result"
        return 1
    fi
}

# Validate chain from certificate directory
# Usage: validate_chain_from_dir <certs_dir> [--ca-file <ca_bundle>]
validate_chain_from_dir() {
    local certs_dir="$1"
    shift
    local ca_file=""

    # Parse additional arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --ca-file)
                ca_file="$2"
                shift 2
                ;;
            *)
                shift
                ;;
        esac
    done

    # Find all certificates
    local certs=()
    for cert_file in "$certs_dir"/cert*.pem; do
        [[ -f "$cert_file" ]] && certs+=("$cert_file")
    done

    if [[ ${#certs[@]} -eq 0 ]]; then
        echo "Error: No certificates found in $certs_dir" >&2
        return 1
    fi

    # First cert is leaf, rest are intermediates
    local leaf="${certs[0]}"
    local intermediates=("${certs[@]:1}")

    # Validate
    if [[ -n "$ca_file" ]]; then
        validate_chain "$leaf" "${intermediates[@]}" --ca-file "$ca_file"
    else
        validate_chain "$leaf" "${intermediates[@]}"
    fi
}

# Build and verify chain from remote server
# Usage: verify_remote_chain <host> [port]
verify_remote_chain() {
    local host="$1"
    local port="${2:-443}"

    # Get certificates from server
    local temp_dir=$(mktemp -d)
    local chain_pem="$temp_dir/chain.pem"

    # Retrieve full chain
    echo | openssl s_client -connect "$host:$port" -showcerts 2>/dev/null | \
        sed -n '/-----BEGIN CERTIFICATE-----/,/-----END CERTIFICATE-----/p' > "$chain_pem"

    if [[ ! -s "$chain_pem" ]]; then
        echo "Error: Could not retrieve certificates from $host:$port" >&2
        rm -rf "$temp_dir"
        return 1
    fi

    # Split into individual certs
    local cert_num=0
    local in_cert=false
    local current_cert=""

    while IFS= read -r line; do
        if [[ "$line" == "-----BEGIN CERTIFICATE-----" ]]; then
            in_cert=true
            current_cert="$line"$'\n'
        elif [[ "$line" == "-----END CERTIFICATE-----" ]]; then
            current_cert+="$line"$'\n'
            cert_num=$((cert_num + 1))
            echo "$current_cert" > "$temp_dir/cert${cert_num}.pem"
            in_cert=false
            current_cert=""
        elif [[ "$in_cert" == true ]]; then
            current_cert+="$line"$'\n'
        fi
    done < "$chain_pem"

    # Validate chain
    local result
    result=$(validate_chain_from_dir "$temp_dir")
    local status=$?

    # Cleanup
    rm -rf "$temp_dir"

    echo "$result"
    return $status
}

# Check if certificate is self-signed
# Usage: is_self_signed <cert_file>
is_self_signed() {
    local cert_file="$1"

    local subject=$(openssl x509 -in "$cert_file" -noout -subject -nameopt RFC2253 2>/dev/null | sed 's/subject=//')
    local issuer=$(openssl x509 -in "$cert_file" -noout -issuer -nameopt RFC2253 2>/dev/null | sed 's/issuer=//')

    [[ "$subject" == "$issuer" ]]
}

# Check if certificate is a CA
# Usage: is_ca_certificate <cert_file>
is_ca_certificate() {
    local cert_file="$1"

    openssl x509 -in "$cert_file" -noout -text 2>/dev/null | grep -q "CA:TRUE"
}

# Get certificate chain depth
# Usage: get_chain_depth <certs_dir>
get_chain_depth() {
    local certs_dir="$1"
    local count=0

    for cert_file in "$certs_dir"/cert*.pem; do
        [[ -f "$cert_file" ]] && count=$((count + 1))
    done

    echo "$count"
}

# Analyze chain structure
# Usage: analyze_chain_structure <certs_dir>
# Output: JSON with chain analysis
analyze_chain_structure() {
    local certs_dir="$1"
    local chain_info="["

    local first=true
    for cert_file in "$certs_dir"/cert*.pem; do
        [[ -f "$cert_file" ]] || continue

        local subject=$(openssl x509 -in "$cert_file" -noout -subject -nameopt RFC2253 2>/dev/null | sed 's/subject=//')
        local issuer=$(openssl x509 -in "$cert_file" -noout -issuer -nameopt RFC2253 2>/dev/null | sed 's/issuer=//')
        local is_ca=$(is_ca_certificate "$cert_file" && echo "true" || echo "false")
        local self_signed=$(is_self_signed "$cert_file" && echo "true" || echo "false")

        # Determine certificate type
        local cert_type="end-entity"
        if [[ "$self_signed" == "true" ]]; then
            cert_type="root"
        elif [[ "$is_ca" == "true" ]]; then
            cert_type="intermediate"
        fi

        [[ "$first" == true ]] || chain_info+=","
        first=false

        chain_info+=$(cat << EOF
{
    "file": "$(basename "$cert_file")",
    "type": "$cert_type",
    "subject": "$subject",
    "issuer": "$issuer",
    "is_ca": $is_ca,
    "self_signed": $self_signed
  }
EOF
)
    done

    chain_info+="]"
    echo "$chain_info"
}

# Verify certificate chain order
# Usage: verify_chain_order <certs_dir>
# Returns: 0 if correct order, 1 if incorrect
verify_chain_order() {
    local certs_dir="$1"
    local certs=()

    for cert_file in "$certs_dir"/cert*.pem; do
        [[ -f "$cert_file" ]] && certs+=("$cert_file")
    done

    if [[ ${#certs[@]} -lt 2 ]]; then
        # Single cert or empty - order is trivially correct
        return 0
    fi

    # Check that each cert's issuer matches the next cert's subject
    for ((i=0; i<${#certs[@]}-1; i++)); do
        local current_issuer=$(openssl x509 -in "${certs[$i]}" -noout -issuer -nameopt RFC2253 2>/dev/null | sed 's/issuer=//')
        local next_subject=$(openssl x509 -in "${certs[$((i+1))]}" -noout -subject -nameopt RFC2253 2>/dev/null | sed 's/subject=//')

        if [[ "$current_issuer" != "$next_subject" ]]; then
            echo "Chain order error: Certificate $((i+1)) issuer does not match certificate $((i+2)) subject"
            return 1
        fi
    done

    echo "Chain order is correct"
    return 0
}

# Get chain validation summary
# Usage: get_chain_validation_summary <certs_dir> [--ca-file <ca_bundle>]
get_chain_validation_summary() {
    local certs_dir="$1"
    shift
    local ca_file=""

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --ca-file)
                ca_file="$2"
                shift 2
                ;;
            *)
                shift
                ;;
        esac
    done

    local depth=$(get_chain_depth "$certs_dir")
    local order_check=$(verify_chain_order "$certs_dir")
    local order_valid=$?

    local validation_result
    if [[ -n "$ca_file" ]]; then
        validation_result=$(validate_chain_from_dir "$certs_dir" --ca-file "$ca_file")
    else
        validation_result=$(validate_chain_from_dir "$certs_dir")
    fi
    local chain_valid=$?

    cat << EOF
Chain Validation Summary
========================
Chain depth: $depth certificate(s)
Chain order: $([ $order_valid -eq 0 ] && echo "✓ Correct" || echo "✗ Incorrect")
Chain validation: $([ $chain_valid -eq 0 ] && echo "✓ Valid" || echo "✗ Invalid")

Details:
$validation_result
EOF
}

# Export functions
export -f detect_trust_store
export -f validate_chain
export -f validate_chain_from_dir
export -f verify_remote_chain
export -f is_self_signed
export -f is_ca_certificate
export -f get_chain_depth
export -f analyze_chain_structure
export -f verify_chain_order
export -f get_chain_validation_summary
