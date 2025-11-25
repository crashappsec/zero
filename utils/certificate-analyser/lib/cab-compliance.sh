#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# CA/Browser Forum Baseline Requirements Compliance Library
# Validates certificates against CA/B Forum policies
# Reference: https://cabforum.org/baseline-requirements/
#############################################################################

# Policy dates and limits
CAB_MAX_VALIDITY_CURRENT=398      # Since Sept 1, 2020
CAB_MAX_VALIDITY_LEGACY=825       # Before Sept 1, 2020
CAB_POLICY_DATE_398="2020-09-01"  # Date when 398-day limit took effect
CAB_MIN_RSA_KEY_SIZE=2048         # Minimum RSA key size
CAB_MIN_ECC_KEY_SIZE=256          # Minimum ECC key size (P-256)

# Deprecated algorithms
CAB_DEPRECATED_HASH_ALGORITHMS="md5 md4 md2 sha1"
CAB_DEPRECATED_SIG_ALGORITHMS="md5WithRSAEncryption md4WithRSAEncryption md2WithRSAEncryption sha1WithRSAEncryption"

#############################################################################
# Validity Period Compliance
#############################################################################

# Check validity period compliance
# Usage: check_validity_compliance <cert_file>
# Returns JSON object with compliance details
check_validity_compliance() {
    local cert_file="$1"

    local not_before=$(openssl x509 -in "$cert_file" -noout -startdate 2>/dev/null | sed 's/notBefore=//')
    local not_after=$(openssl x509 -in "$cert_file" -noout -enddate 2>/dev/null | sed 's/notAfter=//')

    # Parse dates to epochs
    local start_epoch end_epoch
    start_epoch=$(parse_cert_date "$not_before")
    end_epoch=$(parse_cert_date "$not_after")

    if [[ "$start_epoch" == "0" ]] || [[ "$end_epoch" == "0" ]]; then
        echo '{"compliant": false, "error": "Unable to parse certificate dates"}'
        return 1
    fi

    local validity_days=$(( (end_epoch - start_epoch) / 86400 ))
    local policy_date_epoch=$(date -j -f "%Y-%m-%d" "$CAB_POLICY_DATE_398" +%s 2>/dev/null || date -d "$CAB_POLICY_DATE_398" +%s 2>/dev/null)

    local max_validity
    local policy_version
    if [[ $start_epoch -ge $policy_date_epoch ]]; then
        max_validity=$CAB_MAX_VALIDITY_CURRENT
        policy_version="BR 1.7.1+ (Sept 2020)"
    else
        max_validity=$CAB_MAX_VALIDITY_LEGACY
        policy_version="BR Pre-2020"
    fi

    local compliant="true"
    local message=""

    if [[ $validity_days -gt $max_validity ]]; then
        compliant="false"
        message="Validity period ($validity_days days) exceeds maximum ($max_validity days)"
    else
        message="Validity period ($validity_days days) within limits ($max_validity days max)"
    fi

    cat <<EOF
{
  "check": "validity_period",
  "compliant": $compliant,
  "validity_days": $validity_days,
  "max_allowed": $max_validity,
  "policy_version": "$policy_version",
  "not_before": "$not_before",
  "not_after": "$not_after",
  "message": "$message"
}
EOF
}

#############################################################################
# Key Size Compliance
#############################################################################

# Check key size compliance
# Usage: check_key_compliance <cert_file>
check_key_compliance() {
    local cert_file="$1"

    local key_info=$(openssl x509 -in "$cert_file" -noout -text 2>/dev/null | grep -A1 "Public Key Algorithm")
    local key_algo=$(echo "$key_info" | head -1 | sed 's/.*: //')
    local key_size=""
    local compliant="true"
    local message=""
    local min_size=0

    if echo "$key_algo" | grep -qi "rsa"; then
        key_size=$(openssl x509 -in "$cert_file" -noout -text 2>/dev/null | grep "Public-Key:" | grep -o '[0-9]*')
        min_size=$CAB_MIN_RSA_KEY_SIZE
        if [[ $key_size -lt $min_size ]]; then
            compliant="false"
            message="RSA key size ($key_size bits) below minimum ($min_size bits)"
        else
            message="RSA key size ($key_size bits) meets minimum ($min_size bits)"
        fi
    elif echo "$key_algo" | grep -qiE "ec|ecdsa"; then
        key_size=$(openssl x509 -in "$cert_file" -noout -text 2>/dev/null | grep "Public-Key:" | grep -o '[0-9]*')
        min_size=$CAB_MIN_ECC_KEY_SIZE
        if [[ $key_size -lt $min_size ]]; then
            compliant="false"
            message="ECC key size ($key_size bits) below minimum ($min_size bits)"
        else
            message="ECC key size ($key_size bits) meets minimum ($min_size bits)"
        fi
    else
        message="Unknown key algorithm: $key_algo"
    fi

    cat <<EOF
{
  "check": "key_size",
  "compliant": $compliant,
  "algorithm": "$key_algo",
  "key_size": $key_size,
  "min_required": $min_size,
  "message": "$message"
}
EOF
}

#############################################################################
# Signature Algorithm Compliance
#############################################################################

# Check signature algorithm compliance
# Usage: check_signature_compliance <cert_file>
check_signature_compliance() {
    local cert_file="$1"

    local sig_algo=$(openssl x509 -in "$cert_file" -noout -text 2>/dev/null | grep "Signature Algorithm:" | head -1 | sed 's/.*: //')
    local compliant="true"
    local severity="none"
    local message=""

    # Check for deprecated algorithms
    local sig_lower=$(echo "$sig_algo" | tr '[:upper:]' '[:lower:]')

    if echo "$sig_lower" | grep -qiE "md5|md4|md2"; then
        compliant="false"
        severity="critical"
        message="MD5/MD4/MD2 signatures are prohibited"
    elif echo "$sig_lower" | grep -qi "sha1"; then
        compliant="false"
        severity="high"
        message="SHA-1 signatures deprecated since 2017"
    elif echo "$sig_lower" | grep -qiE "sha256|sha384|sha512"; then
        message="SHA-2 family signature algorithm is compliant"
    elif echo "$sig_lower" | grep -qi "ecdsa"; then
        message="ECDSA signature algorithm is compliant"
    else
        message="Signature algorithm: $sig_algo"
    fi

    cat <<EOF
{
  "check": "signature_algorithm",
  "compliant": $compliant,
  "algorithm": "$sig_algo",
  "severity": "$severity",
  "message": "$message"
}
EOF
}

#############################################################################
# Subject Alternative Name (SAN) Compliance
#############################################################################

# Check SAN compliance
# Usage: check_san_compliance <cert_file>
check_san_compliance() {
    local cert_file="$1"

    local san_ext=$(openssl x509 -in "$cert_file" -noout -text 2>/dev/null | grep -A1 "Subject Alternative Name" | tail -1)
    local cn=$(openssl x509 -in "$cert_file" -noout -subject -nameopt RFC2253 2>/dev/null | sed -n 's/.*CN=\([^,]*\).*/\1/p' || echo "")

    local has_san="false"
    local san_count=0
    local compliant="true"
    local message=""
    local warnings=""

    if [[ -n "$san_ext" ]] && [[ "$san_ext" != *"Subject Alternative Name"* ]]; then
        has_san="true"
        # Count SANs
        san_count=$(echo "$san_ext" | tr ',' '\n' | wc -l | tr -d ' ')
    fi

    # BR 7.1.4.2.1: Subject Alternative Name MUST be present
    if [[ "$has_san" == "false" ]]; then
        compliant="false"
        message="Subject Alternative Name extension is required but not present"
    else
        message="Subject Alternative Name extension present with $san_count entries"

        # Check if CN is in SAN (required for public trust)
        if [[ -n "$cn" ]]; then
            if ! echo "$san_ext" | grep -qi "$cn"; then
                warnings="Warning: CN value should be included in SAN"
            fi
        fi

        # Check for wildcard usage
        if echo "$san_ext" | grep -q '\*\.'; then
            warnings="${warnings:+$warnings; }Note: Wildcard certificate detected"
        fi

        # Check for internal names (not allowed for public trust)
        if echo "$san_ext" | grep -qiE 'localhost|\.local|\.internal|\.corp|\.lan'; then
            warnings="${warnings:+$warnings; }Warning: Internal/private domain names detected"
        fi

        # Check for IP addresses
        if echo "$san_ext" | grep -qE 'IP Address:[0-9]'; then
            warnings="${warnings:+$warnings; }Note: IP address in SAN"
        fi
    fi

    cat <<EOF
{
  "check": "subject_alternative_name",
  "compliant": $compliant,
  "has_san": $has_san,
  "san_count": $san_count,
  "san_values": "$(echo "$san_ext" | tr -d '\n' | sed 's/"/\\"/g')",
  "message": "$message",
  "warnings": "$warnings"
}
EOF
}

#############################################################################
# Basic Constraints Compliance
#############################################################################

# Check Basic Constraints compliance
# Usage: check_basic_constraints <cert_file>
check_basic_constraints() {
    local cert_file="$1"

    local bc_ext=$(openssl x509 -in "$cert_file" -noout -text 2>/dev/null | grep -A2 "Basic Constraints")
    local is_ca="false"
    local path_length=""
    local critical="false"
    local compliant="true"
    local message=""

    if echo "$bc_ext" | grep -q "critical"; then
        critical="true"
    fi

    if echo "$bc_ext" | grep -q "CA:TRUE"; then
        is_ca="true"
        path_length=$(echo "$bc_ext" | grep -o "pathlen:[0-9]*" | grep -o '[0-9]*' || echo "unlimited")

        # CA certificates MUST have Basic Constraints marked critical
        if [[ "$critical" != "true" ]]; then
            compliant="false"
            message="CA certificate must have Basic Constraints marked critical"
        else
            message="CA certificate with Basic Constraints critical"
        fi
    else
        # End-entity certificates
        if echo "$bc_ext" | grep -q "Basic Constraints"; then
            message="End-entity certificate with Basic Constraints present"
        else
            message="End-entity certificate (no Basic Constraints or CA:FALSE)"
        fi
    fi

    cat <<EOF
{
  "check": "basic_constraints",
  "compliant": $compliant,
  "is_ca": $is_ca,
  "critical": $critical,
  "path_length": "$path_length",
  "message": "$message"
}
EOF
}

#############################################################################
# Key Usage Compliance
#############################################################################

# Check Key Usage compliance
# Usage: check_key_usage <cert_file>
check_key_usage() {
    local cert_file="$1"

    local ku_ext=$(openssl x509 -in "$cert_file" -noout -text 2>/dev/null | grep -A1 "Key Usage")
    local eku_ext=$(openssl x509 -in "$cert_file" -noout -text 2>/dev/null | grep -A3 "Extended Key Usage")

    local has_ku="false"
    local has_eku="false"
    local ku_critical="false"
    local compliant="true"
    local message=""
    local key_usage=""
    local extended_key_usage=""

    if echo "$ku_ext" | grep -q "Key Usage"; then
        has_ku="true"
        if echo "$ku_ext" | grep -q "critical"; then
            ku_critical="true"
        fi
        key_usage=$(echo "$ku_ext" | tail -1 | tr -d ' ')
    fi

    if echo "$eku_ext" | grep -q "Extended Key Usage"; then
        has_eku="true"
        extended_key_usage=$(echo "$eku_ext" | grep -v "Extended Key Usage" | grep -v "critical" | tr '\n' ',' | sed 's/,$//')
    fi

    # For TLS server certificates, check appropriate key usage
    if echo "$extended_key_usage" | grep -qi "TLS Web Server Authentication"; then
        # Digital Signature should be present for RSA-based TLS
        if ! echo "$key_usage" | grep -qi "Digital Signature"; then
            message="TLS server cert should have Digital Signature key usage"
        else
            message="TLS server certificate with appropriate key usage"
        fi
    elif [[ "$has_ku" == "true" ]]; then
        message="Key Usage extension present"
    else
        message="No Key Usage extension (may be required depending on certificate type)"
    fi

    cat <<EOF
{
  "check": "key_usage",
  "compliant": $compliant,
  "has_key_usage": $has_ku,
  "has_extended_key_usage": $has_eku,
  "key_usage_critical": $ku_critical,
  "key_usage": "$key_usage",
  "extended_key_usage": "$(echo "$extended_key_usage" | tr -d '\n' | sed 's/"/\\"/g')",
  "message": "$message"
}
EOF
}

#############################################################################
# Certificate Transparency Compliance
#############################################################################

# Check CT compliance (SCT presence)
# Usage: check_ct_compliance <cert_file>
check_ct_compliance() {
    local cert_file="$1"

    # Check for CT extension by OID (1.3.6.1.4.1.11129.2.4.2) or name
    local cert_text=$(openssl x509 -in "$cert_file" -noout -text 2>/dev/null)
    local sct_ext=$(echo "$cert_text" | grep -iE "CT Precertificate SCTs|Signed Certificate Timestamp|1\.3\.6\.1\.4\.1\.11129\.2\.4\.2")
    local has_sct="false"
    local sct_count=0
    local compliant="true"
    local message=""

    if [[ -n "$sct_ext" ]]; then
        has_sct="true"
        # Count SCTs (rough estimate - look for CT log entries or OID occurrences)
        sct_count=$(echo "$cert_text" | grep -c "Signed Certificate Timestamp" 2>/dev/null || echo "0")
        if [[ "$sct_count" == "0" ]]; then
            # If we found the extension but not specific SCT entries, at least one exists
            sct_count=1
        fi
        message="Certificate Transparency SCTs present"
    else
        # Check if it's a public CA cert (would need SCTs)
        local issuer=$(openssl x509 -in "$cert_file" -noout -issuer 2>/dev/null)
        if echo "$issuer" | grep -qiE "Let's Encrypt|DigiCert|Comodo|Sectigo|GlobalSign|GoDaddy|Amazon|Google Trust"; then
            compliant="false"
            message="Public CA certificate missing required SCTs"
        else
            message="No SCTs found (may be private/internal CA)"
        fi
    fi

    cat <<EOF
{
  "check": "certificate_transparency",
  "compliant": $compliant,
  "has_sct": $has_sct,
  "sct_count": $sct_count,
  "message": "$message"
}
EOF
}

#############################################################################
# Full Compliance Check
#############################################################################

# Run all compliance checks
# Usage: run_cab_compliance_check <cert_file>
run_cab_compliance_check() {
    local cert_file="$1"

    echo "{"
    echo '  "certificate_file": "'"$cert_file"'",'
    echo '  "checks": ['

    # Run all checks
    echo "    $(check_validity_compliance "$cert_file"),"
    echo "    $(check_key_compliance "$cert_file"),"
    echo "    $(check_signature_compliance "$cert_file"),"
    echo "    $(check_san_compliance "$cert_file"),"
    echo "    $(check_basic_constraints "$cert_file"),"
    echo "    $(check_key_usage "$cert_file"),"
    echo "    $(check_ct_compliance "$cert_file")"

    echo "  ]"
    echo "}"
}

# Print human-readable compliance report
# Usage: print_cab_compliance_report <cert_file>
print_cab_compliance_report() {
    local cert_file="$1"

    echo "CA/Browser Forum Baseline Requirements Compliance"
    echo "=================================================="
    echo ""

    local total_checks=0
    local passed_checks=0

    # Validity Period
    echo "1. Validity Period"
    echo "   ---------------"
    local validity_result=$(check_validity_compliance "$cert_file")
    local validity_compliant=$(echo "$validity_result" | grep -o '"compliant": [^,]*' | cut -d' ' -f2)
    local validity_msg=$(echo "$validity_result" | grep -o '"message": "[^"]*"' | cut -d'"' -f4)
    total_checks=$((total_checks + 1))
    if [[ "$validity_compliant" == "true" ]]; then
        echo "   ✓ PASS: $validity_msg"
        passed_checks=$((passed_checks + 1))
    else
        echo "   ✗ FAIL: $validity_msg"
    fi
    echo ""

    # Key Size
    echo "2. Key Size"
    echo "   --------"
    local key_result=$(check_key_compliance "$cert_file")
    local key_compliant=$(echo "$key_result" | grep -o '"compliant": [^,]*' | cut -d' ' -f2)
    local key_msg=$(echo "$key_result" | grep -o '"message": "[^"]*"' | cut -d'"' -f4)
    total_checks=$((total_checks + 1))
    if [[ "$key_compliant" == "true" ]]; then
        echo "   ✓ PASS: $key_msg"
        passed_checks=$((passed_checks + 1))
    else
        echo "   ✗ FAIL: $key_msg"
    fi
    echo ""

    # Signature Algorithm
    echo "3. Signature Algorithm"
    echo "   -------------------"
    local sig_result=$(check_signature_compliance "$cert_file")
    local sig_compliant=$(echo "$sig_result" | grep -o '"compliant": [^,]*' | cut -d' ' -f2)
    local sig_msg=$(echo "$sig_result" | grep -o '"message": "[^"]*"' | cut -d'"' -f4)
    total_checks=$((total_checks + 1))
    if [[ "$sig_compliant" == "true" ]]; then
        echo "   ✓ PASS: $sig_msg"
        passed_checks=$((passed_checks + 1))
    else
        echo "   ✗ FAIL: $sig_msg"
    fi
    echo ""

    # Subject Alternative Name
    echo "4. Subject Alternative Name (SAN)"
    echo "   -------------------------------"
    local san_result=$(check_san_compliance "$cert_file")
    local san_compliant=$(echo "$san_result" | grep -o '"compliant": [^,]*' | cut -d' ' -f2)
    local san_msg=$(echo "$san_result" | grep -o '"message": "[^"]*"' | cut -d'"' -f4)
    local san_warnings=$(echo "$san_result" | grep -o '"warnings": "[^"]*"' | cut -d'"' -f4)
    total_checks=$((total_checks + 1))
    if [[ "$san_compliant" == "true" ]]; then
        echo "   ✓ PASS: $san_msg"
        passed_checks=$((passed_checks + 1))
    else
        echo "   ✗ FAIL: $san_msg"
    fi
    [[ -n "$san_warnings" ]] && echo "   ⚠ $san_warnings"
    echo ""

    # Basic Constraints
    echo "5. Basic Constraints"
    echo "   -----------------"
    local bc_result=$(check_basic_constraints "$cert_file")
    local bc_compliant=$(echo "$bc_result" | grep -o '"compliant": [^,]*' | cut -d' ' -f2)
    local bc_msg=$(echo "$bc_result" | grep -o '"message": "[^"]*"' | cut -d'"' -f4)
    total_checks=$((total_checks + 1))
    if [[ "$bc_compliant" == "true" ]]; then
        echo "   ✓ PASS: $bc_msg"
        passed_checks=$((passed_checks + 1))
    else
        echo "   ✗ FAIL: $bc_msg"
    fi
    echo ""

    # Key Usage
    echo "6. Key Usage"
    echo "   ---------"
    local ku_result=$(check_key_usage "$cert_file")
    local ku_compliant=$(echo "$ku_result" | grep -o '"compliant": [^,]*' | cut -d' ' -f2)
    local ku_msg=$(echo "$ku_result" | grep -o '"message": "[^"]*"' | cut -d'"' -f4)
    total_checks=$((total_checks + 1))
    if [[ "$ku_compliant" == "true" ]]; then
        echo "   ✓ PASS: $ku_msg"
        passed_checks=$((passed_checks + 1))
    else
        echo "   ✗ FAIL: $ku_msg"
    fi
    echo ""

    # Certificate Transparency
    echo "7. Certificate Transparency"
    echo "   ------------------------"
    local ct_result=$(check_ct_compliance "$cert_file")
    local ct_compliant=$(echo "$ct_result" | grep -o '"compliant": [^,]*' | cut -d' ' -f2)
    local ct_msg=$(echo "$ct_result" | grep -o '"message": "[^"]*"' | cut -d'"' -f4)
    total_checks=$((total_checks + 1))
    if [[ "$ct_compliant" == "true" ]]; then
        echo "   ✓ PASS: $ct_msg"
        passed_checks=$((passed_checks + 1))
    else
        echo "   ✗ FAIL: $ct_msg"
    fi
    echo ""

    # Summary
    echo "=================================================="
    echo "Summary: $passed_checks/$total_checks checks passed"
    if [[ $passed_checks -eq $total_checks ]]; then
        echo "Status: ✓ FULLY COMPLIANT"
    else
        echo "Status: ✗ NON-COMPLIANT"
    fi
}

# Helper function to parse certificate dates (cross-platform)
parse_cert_date() {
    local date_str="$1"

    # Try GNU date first (Linux)
    date -d "$date_str" +%s 2>/dev/null && return 0

    # Try macOS date
    date -j -f "%b %d %H:%M:%S %Y %Z" "$date_str" +%s 2>/dev/null && return 0

    echo "0"
}

# Export functions
export -f check_validity_compliance
export -f check_key_compliance
export -f check_signature_compliance
export -f check_san_compliance
export -f check_basic_constraints
export -f check_key_usage
export -f check_ct_compliance
export -f run_cab_compliance_check
export -f print_cab_compliance_report
export -f parse_cert_date
