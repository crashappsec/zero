#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Certificate Comparison Library
# Compare multiple certificates to identify differences
#############################################################################

# Compare two certificates
# Usage: compare_certificates <cert1_file> <cert2_file>
compare_certificates() {
    local cert1="$1"
    local cert2="$2"

    if [[ ! -f "$cert1" ]] || [[ ! -f "$cert2" ]]; then
        echo "Error: Both certificate files must exist" >&2
        return 1
    fi

    echo "Certificate Comparison"
    echo "======================"
    echo ""

    # Get basic info for both certs
    local subject1=$(openssl x509 -in "$cert1" -noout -subject -nameopt RFC2253 2>/dev/null | sed 's/subject=//')
    local subject2=$(openssl x509 -in "$cert2" -noout -subject -nameopt RFC2253 2>/dev/null | sed 's/subject=//')

    local issuer1=$(openssl x509 -in "$cert1" -noout -issuer -nameopt RFC2253 2>/dev/null | sed 's/issuer=//')
    local issuer2=$(openssl x509 -in "$cert2" -noout -issuer -nameopt RFC2253 2>/dev/null | sed 's/issuer=//')

    local serial1=$(openssl x509 -in "$cert1" -noout -serial 2>/dev/null | sed 's/serial=//')
    local serial2=$(openssl x509 -in "$cert2" -noout -serial 2>/dev/null | sed 's/serial=//')

    local notbefore1=$(openssl x509 -in "$cert1" -noout -startdate 2>/dev/null | sed 's/notBefore=//')
    local notbefore2=$(openssl x509 -in "$cert2" -noout -startdate 2>/dev/null | sed 's/notBefore=//')

    local notafter1=$(openssl x509 -in "$cert1" -noout -enddate 2>/dev/null | sed 's/notAfter=//')
    local notafter2=$(openssl x509 -in "$cert2" -noout -enddate 2>/dev/null | sed 's/notAfter=//')

    local sig1=$(openssl x509 -in "$cert1" -noout -text 2>/dev/null | grep "Signature Algorithm:" | head -1 | sed 's/.*: //')
    local sig2=$(openssl x509 -in "$cert2" -noout -text 2>/dev/null | grep "Signature Algorithm:" | head -1 | sed 's/.*: //')

    local key1=$(openssl x509 -in "$cert1" -noout -text 2>/dev/null | grep "Public-Key:" | sed 's/.*: //')
    local key2=$(openssl x509 -in "$cert2" -noout -text 2>/dev/null | grep "Public-Key:" | sed 's/.*: //')

    local fp1=$(openssl x509 -in "$cert1" -noout -fingerprint -sha256 2>/dev/null | sed 's/.*=//')
    local fp2=$(openssl x509 -in "$cert2" -noout -fingerprint -sha256 2>/dev/null | sed 's/.*=//')

    # Print comparison table
    printf "%-25s %-40s %-40s %s\n" "Attribute" "Certificate 1" "Certificate 2" "Match"
    printf "%-25s %-40s %-40s %s\n" "------------------------" "----------------------------------------" "----------------------------------------" "-----"

    # Subject
    local subj_match="✗"
    [[ "$subject1" == "$subject2" ]] && subj_match="✓"
    printf "%-25s %-40.40s %-40.40s %s\n" "Subject" "$subject1" "$subject2" "$subj_match"

    # Issuer
    local issuer_match="✗"
    [[ "$issuer1" == "$issuer2" ]] && issuer_match="✓"
    printf "%-25s %-40.40s %-40.40s %s\n" "Issuer" "$issuer1" "$issuer2" "$issuer_match"

    # Serial Number
    local serial_match="✗"
    [[ "$serial1" == "$serial2" ]] && serial_match="✓"
    printf "%-25s %-40.40s %-40.40s %s\n" "Serial Number" "$serial1" "$serial2" "$serial_match"

    # Not Before
    local nb_match="✗"
    [[ "$notbefore1" == "$notbefore2" ]] && nb_match="✓"
    printf "%-25s %-40.40s %-40.40s %s\n" "Not Before" "$notbefore1" "$notbefore2" "$nb_match"

    # Not After
    local na_match="✗"
    [[ "$notafter1" == "$notafter2" ]] && na_match="✓"
    printf "%-25s %-40.40s %-40.40s %s\n" "Not After" "$notafter1" "$notafter2" "$na_match"

    # Signature Algorithm
    local sig_match="✗"
    [[ "$sig1" == "$sig2" ]] && sig_match="✓"
    printf "%-25s %-40.40s %-40.40s %s\n" "Signature Algorithm" "$sig1" "$sig2" "$sig_match"

    # Key Size
    local key_match="✗"
    [[ "$key1" == "$key2" ]] && key_match="✓"
    printf "%-25s %-40.40s %-40.40s %s\n" "Public Key" "$key1" "$key2" "$key_match"

    # SHA-256 Fingerprint
    local fp_match="✗"
    [[ "$fp1" == "$fp2" ]] && fp_match="✓"
    printf "%-25s %-40.40s %-40.40s %s\n" "SHA-256 Fingerprint" "${fp1:0:40}..." "${fp2:0:40}..." "$fp_match"

    echo ""

    # Overall result
    if [[ "$fp1" == "$fp2" ]]; then
        echo "Result: ✓ Certificates are IDENTICAL"
        return 0
    elif [[ "$subject1" == "$subject2" ]] && [[ "$issuer1" == "$issuer2" ]]; then
        echo "Result: ⚠ Certificates have SAME SUBJECT/ISSUER but different content"
        echo "        Likely different versions of the same certificate"
        return 1
    else
        echo "Result: ✗ Certificates are DIFFERENT"
        return 2
    fi
}

# Compare certificate with remote server
# Usage: compare_cert_with_remote <cert_file> <host> [port]
compare_cert_with_remote() {
    local cert_file="$1"
    local host="$2"
    local port="${3:-443}"

    if [[ ! -f "$cert_file" ]]; then
        echo "Error: Certificate file not found: $cert_file" >&2
        return 1
    fi

    local temp_file=$(mktemp)

    # Fetch certificate from remote
    echo | openssl s_client -connect "$host:$port" -servername "$host" 2>/dev/null | \
        sed -n '/-----BEGIN CERTIFICATE-----/,/-----END CERTIFICATE-----/p' > "$temp_file"

    if [[ ! -s "$temp_file" ]]; then
        echo "Error: Could not retrieve certificate from $host:$port" >&2
        rm -f "$temp_file"
        return 1
    fi

    echo "Comparing local certificate with $host:$port"
    echo ""

    compare_certificates "$cert_file" "$temp_file"
    local result=$?

    rm -f "$temp_file"
    return $result
}

# Compare multiple certificates in a chain
# Usage: compare_chain <dir1> <dir2>
compare_chains() {
    local dir1="$1"
    local dir2="$2"

    if [[ ! -d "$dir1" ]] || [[ ! -d "$dir2" ]]; then
        echo "Error: Both directories must exist" >&2
        return 1
    fi

    echo "Chain Comparison"
    echo "================"
    echo ""

    local certs1=()
    local certs2=()

    for f in "$dir1"/cert*.pem; do
        [[ -f "$f" ]] && certs1+=("$f")
    done

    for f in "$dir2"/cert*.pem; do
        [[ -f "$f" ]] && certs2+=("$f")
    done

    echo "Chain 1: ${#certs1[@]} certificate(s)"
    echo "Chain 2: ${#certs2[@]} certificate(s)"
    echo ""

    if [[ ${#certs1[@]} -ne ${#certs2[@]} ]]; then
        echo "⚠ Chains have different lengths"
        echo ""
    fi

    local max_certs=${#certs1[@]}
    [[ ${#certs2[@]} -gt $max_certs ]] && max_certs=${#certs2[@]}

    for ((i=0; i<max_certs; i++)); do
        local cert1="${certs1[$i]:-}"
        local cert2="${certs2[$i]:-}"

        echo "--- Certificate $((i+1)) ---"

        if [[ -z "$cert1" ]]; then
            echo "Missing in Chain 1"
        elif [[ -z "$cert2" ]]; then
            echo "Missing in Chain 2"
        else
            compare_certificates "$cert1" "$cert2"
        fi
        echo ""
    done
}

# Generate comparison summary as JSON
# Usage: compare_certificates_json <cert1_file> <cert2_file>
compare_certificates_json() {
    local cert1="$1"
    local cert2="$2"

    local subject1=$(openssl x509 -in "$cert1" -noout -subject -nameopt RFC2253 2>/dev/null | sed 's/subject=//')
    local subject2=$(openssl x509 -in "$cert2" -noout -subject -nameopt RFC2253 2>/dev/null | sed 's/subject=//')

    local fp1=$(openssl x509 -in "$cert1" -noout -fingerprint -sha256 2>/dev/null | sed 's/.*=//')
    local fp2=$(openssl x509 -in "$cert2" -noout -fingerprint -sha256 2>/dev/null | sed 's/.*=//')

    local serial1=$(openssl x509 -in "$cert1" -noout -serial 2>/dev/null | sed 's/serial=//')
    local serial2=$(openssl x509 -in "$cert2" -noout -serial 2>/dev/null | sed 's/serial=//')

    local notafter1=$(openssl x509 -in "$cert1" -noout -enddate 2>/dev/null | sed 's/notAfter=//')
    local notafter2=$(openssl x509 -in "$cert2" -noout -enddate 2>/dev/null | sed 's/notAfter=//')

    local identical="false"
    [[ "$fp1" == "$fp2" ]] && identical="true"

    local same_subject="false"
    [[ "$subject1" == "$subject2" ]] && same_subject="true"

    cat <<EOF
{
  "comparison": {
    "identical": $identical,
    "same_subject": $same_subject,
    "cert1": {
      "subject": "$(echo "$subject1" | sed 's/"/\\"/g')",
      "serial": "$serial1",
      "expires": "$notafter1",
      "fingerprint_sha256": "$fp1"
    },
    "cert2": {
      "subject": "$(echo "$subject2" | sed 's/"/\\"/g')",
      "serial": "$serial2",
      "expires": "$notafter2",
      "fingerprint_sha256": "$fp2"
    }
  }
}
EOF
}

# Check if certificate matches a domain
# Usage: check_cert_domain_match <cert_file> <domain>
check_cert_domain_match() {
    local cert_file="$1"
    local domain="$2"

    local san=$(openssl x509 -in "$cert_file" -noout -text 2>/dev/null | grep -A1 "Subject Alternative Name" | tail -1)
    local cn=$(openssl x509 -in "$cert_file" -noout -subject -nameopt RFC2253 2>/dev/null | sed -n 's/.*CN=\([^,]*\).*/\1/p')

    echo "Domain Match Check"
    echo "=================="
    echo ""
    echo "Domain: $domain"
    echo "CN:     $cn"
    echo "SAN:    $(echo "$san" | tr ',' '\n' | head -5 | tr '\n' ', ')..."
    echo ""

    # Check exact match in CN
    if [[ "$cn" == "$domain" ]]; then
        echo "✓ Exact match in CN"
        return 0
    fi

    # Check exact match in SAN
    if echo "$san" | grep -qi "DNS:$domain"; then
        echo "✓ Exact match in SAN"
        return 0
    fi

    # Check wildcard match
    local parent_domain=$(echo "$domain" | sed 's/^[^.]*\.//')
    if echo "$san" | grep -qi "DNS:\*\.$parent_domain"; then
        echo "✓ Wildcard match (*.$parent_domain) in SAN"
        return 0
    fi

    if [[ "$cn" == "*.$parent_domain" ]]; then
        echo "✓ Wildcard match (*.$parent_domain) in CN"
        return 0
    fi

    echo "✗ No match found for domain: $domain"
    return 1
}

# Compare certificate expiry dates
# Usage: compare_expiry <cert1> <cert2>
compare_expiry() {
    local cert1="$1"
    local cert2="$2"

    local expiry1=$(openssl x509 -in "$cert1" -noout -enddate 2>/dev/null | sed 's/notAfter=//')
    local expiry2=$(openssl x509 -in "$cert2" -noout -enddate 2>/dev/null | sed 's/notAfter=//')

    # Parse to epochs
    local epoch1 epoch2
    epoch1=$(date -d "$expiry1" +%s 2>/dev/null || date -j -f "%b %d %H:%M:%S %Y %Z" "$expiry1" +%s 2>/dev/null || echo "0")
    epoch2=$(date -d "$expiry2" +%s 2>/dev/null || date -j -f "%b %d %H:%M:%S %Y %Z" "$expiry2" +%s 2>/dev/null || echo "0")

    echo "Expiry Comparison"
    echo "================="
    echo ""
    echo "Certificate 1: $expiry1"
    echo "Certificate 2: $expiry2"
    echo ""

    if [[ "$epoch1" == "0" ]] || [[ "$epoch2" == "0" ]]; then
        echo "Unable to parse dates for comparison"
        return 1
    fi

    local diff_seconds=$((epoch2 - epoch1))
    local diff_days=$((diff_seconds / 86400))

    if [[ $diff_days -eq 0 ]]; then
        echo "Result: Certificates expire at the same time"
    elif [[ $diff_days -gt 0 ]]; then
        echo "Result: Certificate 2 expires $diff_days days LATER than Certificate 1"
    else
        echo "Result: Certificate 2 expires ${diff_days#-} days EARLIER than Certificate 1"
    fi
}

# Export functions
export -f compare_certificates
export -f compare_cert_with_remote
export -f compare_chains
export -f compare_certificates_json
export -f check_cert_domain_match
export -f compare_expiry
