#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# 
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Certificate Analysis Script
# Analyzes digital certificates for compliance with CA/Browser Forum policies
# Supports multiple formats: PEM, DER, PKCS7, PKCS12
# Usage: ./cert-analyser.sh [OPTIONS] <target>
#############################################################################

set -euo pipefail

# Script version
SCRIPT_VERSION="2.0.0"

# Colors for output
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
UTILS_ROOT="$(dirname "$SCRIPT_DIR")"
REPO_ROOT="$(dirname "$UTILS_ROOT")"

# Load .env file if it exists in repository root
if [[ -f "$REPO_ROOT/.env" ]]; then
    set -a  # automatically export all variables
    source "$REPO_ROOT/.env"
    set +a  # stop automatically exporting
fi

# Load local libraries
source "$SCRIPT_DIR/lib/format-detection.sh"
source "$SCRIPT_DIR/lib/format-parsers.sh"
source "$SCRIPT_DIR/lib/fingerprint.sh"
source "$SCRIPT_DIR/lib/chain-validation.sh"
source "$SCRIPT_DIR/lib/ocsp-verification.sh"
source "$SCRIPT_DIR/lib/starttls.sh"
source "$SCRIPT_DIR/lib/cab-compliance.sh"
source "$SCRIPT_DIR/lib/cert-compare.sh"
source "$SCRIPT_DIR/lib/claude-analysis.sh"

# Load global libraries
if [[ -f "$UTILS_ROOT/lib/claude-cost.sh" ]]; then
    source "$UTILS_ROOT/lib/claude-cost.sh"
fi

# Configuration
CURRENT_MAX_VALIDITY=398  # Current CA/B Forum limit (days)
EXPIRY_CRITICAL=7         # Critical if expires within X days
EXPIRY_WARNING=30         # Warning if expires within X days

# Global variables
DOMAIN=""
WORK_DIR=""
REPORT_FILE=""
ANALYSIS_DATE=""
USE_CLAUDE=false
ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY:-}"

# New options for Phase 1
INPUT_FILE=""
INPUT_FORMAT=""
INPUT_SOURCE="domain"  # domain|file|stdin
OUTPUT_FORMAT="markdown"  # markdown|json|text
OUTPUT_FILE=""
PKCS12_PASSWORD="${PKCS12_PASSWORD:-}"
PASSWORD_FILE=""
SHOW_FINGERPRINTS=false
FINGERPRINT_ALGO="sha256"
PORT=443
PORT_EXPLICIT=false
STARTTLS_PROTO=""
VERBOSE=false

# Phase 2 options
VERIFY_CHAIN=false
CHECK_OCSP=false
CHECK_CT=false
ALL_CHECKS=false
CA_FILE=""

# Phase 3 options
CAB_COMPLIANCE=false

# Claude AI options
CLAUDE_ANALYSIS_TYPE="comprehensive"  # comprehensive, quick, compliance

# Load cost tracking if using Claude
if [[ "$USE_CLAUDE" == "true" ]]; then
    REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
    if [ -f "$REPO_ROOT/lib/claude-cost.sh" ]; then
        source "$REPO_ROOT/lib/claude-cost.sh"
        init_cost_tracking
    fi
fi

#############################################################################
# Utility Functions
#############################################################################

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

cleanup() {
    if [[ -n "${WORK_DIR:-}" ]] && [[ -d "$WORK_DIR" ]]; then
        rm -rf "$WORK_DIR"
    fi
}

trap cleanup EXIT

#############################################################################
# Cross-platform timeout function
#############################################################################

# Detect if timeout command is available (Linux) or use gtimeout (macOS with coreutils)
run_with_timeout() {
    local timeout_duration=$1
    shift

    if command -v timeout >/dev/null 2>&1; then
        # GNU timeout (Linux)
        timeout "$timeout_duration" "$@"
    elif command -v gtimeout >/dev/null 2>&1; then
        # GNU coreutils timeout on macOS (brew install coreutils)
        gtimeout "$timeout_duration" "$@"
    else
        # Fallback: Use Perl-based timeout for macOS without GNU coreutils
        perl -e '
            my $timeout = shift @ARGV;
            my $pid = fork();
            if ($pid == 0) {
                exec @ARGV;
                exit 1;
            }
            eval {
                local $SIG{ALRM} = sub { die "timeout\n" };
                alarm $timeout;
                waitpid($pid, 0);
                alarm 0;
            };
            if ($@ eq "timeout\n") {
                kill 9, $pid;
                exit 124;
            }
            exit $? >> 8;
        ' "$timeout_duration" "$@"
    fi
}

#############################################################################
# Certificate Retrieval
#############################################################################

retrieve_certificates() {
    local domain=$1
    local output_dir=$2
    
    log_info "Retrieving certificates for ${domain}..."

    # Retrieve full certificate chain
    if ! echo | run_with_timeout 10 openssl s_client -showcerts -servername "$domain" \
        -connect "${domain}:443" 2>/dev/null > "${output_dir}/chain.txt"; then
        log_error "Failed to retrieve certificates from ${domain}:443"
        log_error "Possible issues: domain unreachable, no HTTPS, firewall blocking"
        return 1
    fi
    
    # Split certificates into individual files
    awk -v outdir="${output_dir}" '/BEGIN CERTIFICATE/,/END CERTIFICATE/{
        if(/BEGIN CERTIFICATE/){a++};
        print > (outdir "/cert" a ".pem")
    }' "${output_dir}/chain.txt"
    
    # Count certificates
    local cert_count=$(ls -1 "${output_dir}"/cert*.pem 2>/dev/null | wc -l)
    
    if [[ $cert_count -eq 0 ]]; then
        log_error "No certificates found in chain"
        return 1
    fi
    
    log_success "Retrieved ${cert_count} certificate(s) from chain"
    return 0
}

#############################################################################
# Certificate Parsing Functions
#############################################################################

get_cert_subject() {
    local cert_file=$1
    openssl x509 -in "$cert_file" -noout -subject -nameopt RFC2253 2>/dev/null | sed 's/subject=//'
}

get_cert_issuer() {
    local cert_file=$1
    openssl x509 -in "$cert_file" -noout -issuer -nameopt RFC2253 2>/dev/null | sed 's/issuer=//'
}

get_cert_serial() {
    local cert_file=$1
    openssl x509 -in "$cert_file" -noout -serial 2>/dev/null | sed 's/serial=//'
}

get_cert_dates() {
    local cert_file=$1
    local date_type=$2  # startdate or enddate
    openssl x509 -in "$cert_file" -noout -"${date_type}" 2>/dev/null | sed 's/.*=//'
}

get_cert_sans() {
    local cert_file=$1
    openssl x509 -in "$cert_file" -noout -ext subjectAltName 2>/dev/null | \
        grep -v "subject" || echo "None"
}

get_cert_signature_algorithm() {
    local cert_file=$1
    openssl x509 -in "$cert_file" -noout -text 2>/dev/null | \
        grep "Signature Algorithm" | head -1 | sed 's/.*: //'
}

get_cert_pubkey_info() {
    local cert_file=$1
    openssl x509 -in "$cert_file" -noout -text 2>/dev/null | \
        grep -A 1 "Public Key Algorithm" | tail -1 | sed 's/^[[:space:]]*//'
}

get_cert_key_usage() {
    local cert_file=$1
    openssl x509 -in "$cert_file" -noout -ext keyUsage 2>/dev/null | \
        grep -v "Key Usage" || echo "Not specified"
}

get_cert_ext_key_usage() {
    local cert_file=$1
    openssl x509 -in "$cert_file" -noout -ext extendedKeyUsage 2>/dev/null | \
        grep -v "Extended Key Usage" || echo "Not specified"
}

#############################################################################
# Date Calculation Functions (Cross-platform)
#############################################################################

# Convert OpenSSL date format to epoch (works on both macOS and Linux)
# OpenSSL format: "Nov 25 05:11:26 2025 GMT"
parse_openssl_date() {
    local date_str="$1"

    # Try GNU date first (Linux)
    if date -d "$date_str" +%s 2>/dev/null; then
        return 0
    fi

    # Fall back to macOS date
    # macOS date expects: -j -f format datestr +output
    # OpenSSL format: "Nov 25 05:11:26 2025 GMT"
    date -j -f "%b %d %H:%M:%S %Y %Z" "$date_str" +%s 2>/dev/null && return 0

    # If all else fails, use openssl directly for epoch
    echo "0"
}

# Get future date string (works on both macOS and Linux)
# Usage: get_future_date <days> [format]
get_future_date() {
    local days="$1"
    local format="${2:-%Y-%m-%d}"

    # Try GNU date first (Linux)
    date -d "+${days} days" +"$format" 2>/dev/null && return 0

    # Fall back to macOS date
    date -v "+${days}d" +"$format" 2>/dev/null && return 0

    # Fallback: calculate manually
    local now_epoch=$(date +%s)
    local future_epoch=$((now_epoch + (days * 86400)))
    # Try to format the epoch - GNU first, then macOS
    date -d "@$future_epoch" +"$format" 2>/dev/null || date -r "$future_epoch" +"$format" 2>/dev/null
}

# Compare a date to a reference date (cross-platform)
# Usage: is_date_before <date_str> <reference_date>
# Returns: 0 if date_str is before reference_date, 1 otherwise
is_date_before() {
    local date_str="$1"
    local ref_date="$2"

    local date_epoch=$(parse_openssl_date "$date_str")
    local ref_epoch

    # Try to parse reference date
    ref_epoch=$(date -d "$ref_date" +%s 2>/dev/null || date -j -f "%Y-%m-%d" "$ref_date" +%s 2>/dev/null)

    [[ "$date_epoch" -lt "$ref_epoch" ]]
}

days_until_expiry() {
    local cert_file=$1

    # Use openssl to get the epoch directly (most reliable)
    local end_epoch=$(openssl x509 -in "$cert_file" -noout -enddate 2>/dev/null | \
        sed 's/.*=//' | xargs -I {} sh -c 'date -j -f "%b %d %H:%M:%S %Y %Z" "{}" +%s 2>/dev/null || date -d "{}" +%s 2>/dev/null')

    if [[ -z "$end_epoch" ]] || [[ "$end_epoch" == "0" ]]; then
        # Fallback: parse dates manually
        local end_date=$(get_cert_dates "$cert_file" "enddate")
        end_epoch=$(parse_openssl_date "$end_date")
    fi

    local now_epoch=$(date +%s)
    local diff_seconds=$((end_epoch - now_epoch))
    local diff_days=$((diff_seconds / 86400))
    echo "$diff_days"
}

cert_validity_period() {
    local cert_file=$1

    # Use openssl to get epochs directly
    local start_date=$(get_cert_dates "$cert_file" "startdate")
    local end_date=$(get_cert_dates "$cert_file" "enddate")

    local start_epoch=$(parse_openssl_date "$start_date")
    local end_epoch=$(parse_openssl_date "$end_date")

    if [[ "$start_epoch" == "0" ]] || [[ "$end_epoch" == "0" ]]; then
        # Fallback: calculate from openssl -text output
        echo "0"
        return
    fi

    local diff_seconds=$((end_epoch - start_epoch))
    local diff_days=$((diff_seconds / 86400))
    echo "$diff_days"
}

#############################################################################
# Certificate Analysis Functions
#############################################################################

analyze_validity_compliance() {
    local validity_days=$1
    local issue_date=$2

    # Check against current CA/B Forum policy (398 days)
    if [[ $validity_days -le $CURRENT_MAX_VALIDITY ]]; then
        echo "✓ Compliant"
    elif is_date_before "$issue_date" "2020-09-01"; then
        echo "⚠️ Legacy (issued before policy change)"
    else
        echo "❌ Non-compliant"
    fi
}

analyze_signature_algorithm() {
    local sig_alg=$1
    
    if [[ $sig_alg =~ sha1 ]] || [[ $sig_alg =~ SHA1 ]]; then
        echo "❌ SHA-1 (deprecated)"
    elif [[ $sig_alg =~ sha256 ]] || [[ $sig_alg =~ sha384 ]] || [[ $sig_alg =~ sha512 ]]; then
        echo "✓ SHA-256+ (compliant)"
    else
        echo "⚠️ Unknown algorithm"
    fi
}

analyze_public_key() {
    local pubkey_info=$1
    
    if [[ $pubkey_info =~ "4096 bit" ]] || [[ $pubkey_info =~ "3072 bit" ]]; then
        echo "✓ RSA 3072+/4096 (strong)"
    elif [[ $pubkey_info =~ "2048 bit" ]]; then
        echo "⚠️ RSA 2048 (minimum, consider upgrade)"
    elif [[ $pubkey_info =~ "1024 bit" ]] || [[ $pubkey_info =~ "512 bit" ]]; then
        echo "❌ RSA <2048 (weak)"
    elif [[ $pubkey_info =~ "256 bit" ]] && [[ $pubkey_info =~ "EC" ]]; then
        echo "✓ ECC P-256+ (strong)"
    else
        echo "ℹ️ $pubkey_info"
    fi
}

check_certificate_transparency() {
    local cert_file=$1
    if openssl x509 -in "$cert_file" -noout -text 2>/dev/null | grep -q "CT Precertificate SCTs"; then
        echo "✓ Present"
    else
        echo "⚠️ Not detected"
    fi
}

check_ocsp_uri() {
    local cert_file=$1
    local ocsp=$(openssl x509 -in "$cert_file" -noout -ocsp_uri 2>/dev/null || echo "")
    if [[ -n "$ocsp" ]]; then
        echo "✓ $ocsp"
    else
        echo "⚠️ Not available"
    fi
}

get_crl_uri() {
    local cert_file=$1
    local crl=$(openssl x509 -in "$cert_file" -noout -text 2>/dev/null | \
        grep -A 4 "CRL Distribution" | grep "URI:" | sed 's/.*URI://' | tr -d '[:space:]')
    if [[ -n "$crl" ]]; then
        echo "✓ $crl"
    else
        echo "⚠️ Not available"
    fi
}

#############################################################################
# Report Generation Functions
#############################################################################

write_report_header() {
    cat >> "$REPORT_FILE" <<EOF
# Certificate Analysis Report

**Domain**: ${DOMAIN}  
**Analysis Date**: ${ANALYSIS_DATE}  
**Analyst**: Certificate Security Analyser v1.0

---

## Executive Summary

EOF
}

write_executive_summary() {
    local total_certs=$1
    local days_remaining=$2
    local validity_status=$3

    # Calculate expiry date (cross-platform)
    local expiry_date=$(get_future_date "$days_remaining" "%Y-%m-%d")

    cat >> "$REPORT_FILE" <<EOF
This report analyzes the complete certificate chain for **${DOMAIN}**, comprising ${total_certs} certificate(s).
The analysis focuses on compliance with current CA/Browser Forum policies, cryptographic strength,
and operational security posture.

**Key Findings:**
- Certificate expires in **${days_remaining} days** (${expiry_date})
- Validity period compliance: **${validity_status}**
- Chain validation and detailed security analysis below

EOF

    if [[ $days_remaining -le $EXPIRY_CRITICAL ]]; then
        cat >> "$REPORT_FILE" <<EOF
🚨 **CRITICAL**: Certificate expires within ${EXPIRY_CRITICAL} days. Immediate renewal required.

EOF
    elif [[ $days_remaining -le $EXPIRY_WARNING ]]; then
        cat >> "$REPORT_FILE" <<EOF
⚠️ **WARNING**: Certificate expires within ${EXPIRY_WARNING} days. Plan renewal soon.

EOF
    fi
    
    echo "---" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
}

write_certificate_details() {
    local cert_file=$1
    local cert_num=$2
    local cert_type=$3
    
    local subject=$(get_cert_subject "$cert_file")
    local issuer=$(get_cert_issuer "$cert_file")
    local serial=$(get_cert_serial "$cert_file")
    local not_before=$(get_cert_dates "$cert_file" "startdate")
    local not_after=$(get_cert_dates "$cert_file" "enddate")
    local days_left=$(days_until_expiry "$cert_file")
    local validity_days=$(cert_validity_period "$cert_file")
    local sig_alg=$(get_cert_signature_algorithm "$cert_file")
    local pubkey=$(get_cert_pubkey_info "$cert_file")
    
    # Status indicator for days remaining
    local expiry_status="✓"
    if [[ $days_left -le $EXPIRY_CRITICAL ]]; then
        expiry_status="❌"
    elif [[ $days_left -le $EXPIRY_WARNING ]]; then
        expiry_status="⚠️"
    fi
    
    cat >> "$REPORT_FILE" <<EOF
### ${cert_type}

- **Subject**: ${subject}
- **Issuer**: ${issuer}
- **Serial Number**: ${serial}
- **Valid From**: ${not_before}
- **Valid Until**: ${not_after}
- **Days Remaining**: ${days_left} days ${expiry_status}
- **Total Validity Period**: ${validity_days} days
- **Signature Algorithm**: ${sig_alg}
- **Public Key**: ${pubkey}

EOF

    # Add SANs for leaf certificate
    if [[ "$cert_type" == "End-Entity Certificate (Leaf)" ]]; then
        local sans=$(get_cert_sans "$cert_file")
        cat >> "$REPORT_FILE" <<EOF
**Subject Alternative Names**:
\`\`\`
${sans}
\`\`\`

EOF
    fi
}

write_detailed_analysis() {
    local leaf_cert=$1
    
    local validity_days=$(cert_validity_period "$leaf_cert")
    local not_before=$(get_cert_dates "$leaf_cert" "startdate")
    local compliance=$(analyze_validity_compliance "$validity_days" "$not_before")
    local sig_alg=$(get_cert_signature_algorithm "$leaf_cert")
    local sig_status=$(analyze_signature_algorithm "$sig_alg")
    local pubkey=$(get_cert_pubkey_info "$leaf_cert")
    local pubkey_status=$(analyze_public_key "$pubkey")
    local ct_status=$(check_certificate_transparency "$leaf_cert")
    local ocsp=$(check_ocsp_uri "$leaf_cert")
    local crl=$(get_crl_uri "$leaf_cert")
    
    cat >> "$REPORT_FILE" <<EOF
## Detailed Analysis

### 1. Validity Period Compliance

| Metric | Value | Status |
|--------|-------|--------|
| Total Validity Period | ${validity_days} days | ${compliance} |
| Current Policy Limit | ${CURRENT_MAX_VALIDITY} days | Reference |
| Future Expected Limit | ~90-180 days | Planning |

**Analysis**: $(
    if [[ $validity_days -le $CURRENT_MAX_VALIDITY ]]; then
        echo "Certificate complies with current CA/Browser Forum policy (max 398 days)."
    else
        echo "Certificate exceeds current policy limits. May be a legacy certificate issued before 2020."
    fi
) Industry is moving toward shorter validity periods (90-180 days expected by 2027) to improve security posture.

### 2. Cryptographic Strength

**Signature Algorithm**: ${sig_alg}
- ${sig_status}

**Public Key**: ${pubkey}
- ${pubkey_status}

**Analysis**: $(
    if [[ $sig_status =~ "✓" ]] && [[ $pubkey_status =~ "✓" ]]; then
        echo "Cryptographic parameters meet current security standards."
    elif [[ $sig_status =~ "⚠️" ]] || [[ $pubkey_status =~ "⚠️" ]]; then
        echo "While currently acceptable, consider upgrading to stronger algorithms (RSA 3072+ or ECC P-256+) during next renewal."
    else
        echo "CRITICAL: Weak cryptographic parameters detected. Immediate upgrade required."
    fi
)

### 3. Subject Alternative Names (SANs)

**Covered Domains**:
\`\`\`
$(get_cert_sans "$leaf_cert")
\`\`\`

**Analysis**: Certificate should cover all domains and subdomains that will use it. Wildcard certificates (*.example.com) provide flexibility but should be managed carefully.

### 4. Certificate Transparency

**CT Log Status**: ${ct_status}

**Analysis**: Certificate Transparency has been mandatory since April 2018. CT logs provide public, append-only records of certificates, enabling detection of mis-issuance.

### 5. Revocation Checking

**OCSP**: ${ocsp}
**CRL**: ${crl}

**Analysis**: $(
    if [[ $ocsp =~ "✓" ]]; then
        echo "OCSP (Online Certificate Status Protocol) provides real-time revocation checking. Consider enabling OCSP stapling for improved performance and privacy."
    else
        echo "No OCSP responder detected. Revocation checking may rely solely on CRL, which is less efficient."
    fi
)

### 6. Chain Validation

EOF

    # Attempt chain validation
    if [[ -f "${WORK_DIR}/cert2.pem" ]]; then
        cat "${WORK_DIR}"/cert*.pem > "${WORK_DIR}/chain-bundle.pem"
        if openssl verify -CAfile "${WORK_DIR}/chain-bundle.pem" "$leaf_cert" 2>&1 | grep -q "OK"; then
            echo "**Status**: ✓ Valid" >> "$REPORT_FILE"
        else
            echo "**Status**: ⚠️ Issues detected (may be due to missing root CA in local trust store)" >> "$REPORT_FILE"
        fi
    else
        echo "**Status**: ℹ️ Single certificate - root validation requires system trust store" >> "$REPORT_FILE"
    fi
    
    echo "" >> "$REPORT_FILE"
    echo "---" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
}

write_risk_assessment() {
    local days_left=$1
    local validity_days=$2
    local sig_status=$3
    local pubkey_status=$4
    
    cat >> "$REPORT_FILE" <<EOF
## Risk Assessment

EOF

    # Critical issues
    local has_critical=false
    echo "### Critical Issues ❌" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    if [[ $days_left -le $EXPIRY_CRITICAL ]]; then
        echo "1. **Imminent Expiration**: Certificate expires in ${days_left} days" >> "$REPORT_FILE"
        has_critical=true
    fi
    
    if [[ $sig_status =~ "❌" ]]; then
        echo "1. **Weak Signature Algorithm**: Using deprecated SHA-1" >> "$REPORT_FILE"
        has_critical=true
    fi
    
    if [[ $pubkey_status =~ "❌" ]]; then
        echo "1. **Weak Public Key**: Key size below minimum requirements" >> "$REPORT_FILE"
        has_critical=true
    fi
    
    if [[ $has_critical == false ]]; then
        echo "*No critical issues detected.*" >> "$REPORT_FILE"
    fi
    echo "" >> "$REPORT_FILE"
    
    # Warnings
    echo "### Warnings ⚠️" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    local has_warning=false
    if [[ $days_left -le $EXPIRY_WARNING ]] && [[ $days_left -gt $EXPIRY_CRITICAL ]]; then
        echo "1. **Upcoming Expiration**: Certificate expires in ${days_left} days - plan renewal" >> "$REPORT_FILE"
        has_warning=true
    fi
    
    if [[ $pubkey_status =~ "⚠️" ]]; then
        echo "1. **Minimal Key Strength**: RSA 2048 is minimum acceptable - consider 3072+ or ECC" >> "$REPORT_FILE"
        has_warning=true
    fi
    
    if [[ $validity_days -gt $CURRENT_MAX_VALIDITY ]]; then
        echo "1. **Long Validity Period**: Certificate has ${validity_days}-day validity (legacy)" >> "$REPORT_FILE"
        has_warning=true
    fi
    
    if [[ $has_warning == false ]]; then
        echo "*No warnings.*" >> "$REPORT_FILE"
    fi
    echo "" >> "$REPORT_FILE"
    
    # Informational
    cat >> "$REPORT_FILE" <<EOF
### Informational ℹ️

1. **Industry Trend**: Certificate lifespans are decreasing to improve security
2. **Automation**: Shorter certificates require automated renewal processes
3. **Best Practice**: Implement monitoring 30/14/7 days before expiration

---

EOF
}

write_recommendations() {
    local days_left=$1
    
    cat >> "$REPORT_FILE" <<EOF
## Recommendations

### Immediate Actions (0-30 days)

EOF

    if [[ $days_left -le $EXPIRY_WARNING ]]; then
        cat >> "$REPORT_FILE" <<EOF
1. **Renew Certificate**: Expiration is imminent - initiate renewal process immediately
2. **Verify Domain Control**: Ensure domain validation methods (DNS, HTTP, email) are accessible
3. **Test Renewal Process**: Verify certificate deployment pipeline is functional

EOF
    else
        cat >> "$REPORT_FILE" <<EOF
1. **Monitor Expiration**: Set up alerts for 30, 14, and 7 days before expiration
2. **Document Renewal Process**: Ensure team knows how to renew certificates
3. **Review Certificate Inventory**: Catalog all certificates and their expiration dates

EOF
    fi
    
    cat >> "$REPORT_FILE" <<EOF
### Short-term Planning (30-90 days)

1. **Implement Automated Renewal**
   - Adopt ACME protocol (Let's Encrypt, commercial CAs supporting ACME)
   - Deploy automation tools (certbot, cert-manager, acme.sh)
   - Test automated renewal in staging environment

2. **Enable Certificate Monitoring**
   - Deploy certificate expiration monitoring (e.g., cert-manager, commercial services)
   - Set up alerting to multiple channels (email, Slack, PagerDuty)
   - Create runbooks for renewal procedures

3. **Security Enhancements**
   - Enable OCSP stapling on web servers
   - Implement Certificate Transparency monitoring
   - Review and update cipher suites

### Long-term Strategy (90+ days)

1. **Prepare for Shorter Certificates**
   - Future CA/Browser Forum policies will reduce validity to 90-180 days
   - Manual processes will not scale - automation is mandatory
   - Budget for automation tools and engineering time

2. **Infrastructure as Code**
   - Treat certificate management as code (Terraform, CloudFormation, Kubernetes)
   - Version control certificate configurations
   - Implement CI/CD for certificate rotation

3. **Consider Modern Alternatives**
   - Evaluate ECC certificates (smaller, faster, equally secure)
   - Investigate managed certificate services (AWS ACM, Google-managed certs)
   - For Kubernetes: implement cert-manager with automatic rotation

4. **Certificate Lifecycle Management**
   - Establish certificate inventory and tracking system
   - Define ownership and responsibilities for certificate renewals
   - Create disaster recovery procedures for certificate issues
   - Regular security audits and compliance checks

---

EOF
}

write_appendix() {
    local cert_count=$1
    
    cat >> "$REPORT_FILE" <<EOF
## Appendix: Raw Certificate Data

EOF

    for i in $(seq 1 "$cert_count"); do
        if [[ -f "${WORK_DIR}/cert${i}.pem" ]]; then
            cat >> "$REPORT_FILE" <<EOF
### Certificate ${i}

\`\`\`
$(openssl x509 -in "${WORK_DIR}/cert${i}.pem" -noout -text 2>/dev/null)
\`\`\`

EOF
        fi
    done
    
    cat >> "$REPORT_FILE" <<EOF
---

## Methodology Notes

**Tools Used**: OpenSSL $(openssl version | awk '{print $2}')  
**Standards Referenced**: 
- CA/Browser Forum Baseline Requirements
- RFC 5280 (X.509 PKI Certificate and CRL Profile)
- RFC 6960 (OCSP)

**Analysis Date**: ${ANALYSIS_DATE}  
**Script Version**: 1.0

---

*This analysis was performed using automated tooling and should be reviewed by qualified security personnel for production environments.*
EOF
}

#############################################################################
# Claude AI Analysis (wrapper for lib/claude-analysis.sh)
#############################################################################

# Main Claude analysis function - uses the enhanced library
analyze_with_claude() {
    local report_content="$1"
    local analysis_type="${CLAUDE_ANALYSIS_TYPE:-comprehensive}"

    if [[ -z "$ANTHROPIC_API_KEY" ]]; then
        log_error "ANTHROPIC_API_KEY required for --claude mode"
        log_info "Set via: export ANTHROPIC_API_KEY='sk-ant-...'"
        log_info "Or add to .env file in repository root"
        exit 1
    fi

    # Use the enhanced RAG-based analysis from claude-analysis.sh library
    analyze_certificate_with_claude "$report_content" "$REPO_ROOT" "$analysis_type"
}

#############################################################################
# Main Analysis Function
#############################################################################

analyze_domain() {
    local domain=$1
    
    DOMAIN="$domain"
    WORK_DIR=$(mktemp -d)
    ANALYSIS_DATE=$(date -u +"%Y-%m-%d %H:%M:%S UTC")
    REPORT_FILE="certificate-analysis-${domain}-$(date +%Y%m%d-%H%M%S).md"
    
    log_info "Starting certificate analysis for ${domain}"
    log_info "Working directory: ${WORK_DIR}"
    
    # Retrieve certificates
    if ! retrieve_certificates "$domain" "$WORK_DIR"; then
        log_error "Certificate retrieval failed"
        return 1
    fi
    
    # Count certificates in chain
    local cert_count=$(ls -1 "${WORK_DIR}"/cert*.pem 2>/dev/null | wc -l)
    
    # Get info from leaf certificate
    local leaf_cert="${WORK_DIR}/cert1.pem"
    local days_remaining=$(days_until_expiry "$leaf_cert")
    local validity_days=$(cert_validity_period "$leaf_cert")
    local not_before=$(get_cert_dates "$leaf_cert" "startdate")
    local validity_status=$(analyze_validity_compliance "$validity_days" "$not_before")
    
    # Generate report
    log_info "Generating analysis report..."
    
    write_report_header
    write_executive_summary "$cert_count" "$days_remaining" "$validity_status"
    
    echo "## Certificate Chain Overview" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    # Write certificate details
    if [[ -f "${WORK_DIR}/cert1.pem" ]]; then
        write_certificate_details "${WORK_DIR}/cert1.pem" 1 "End-Entity Certificate (Leaf)"
    fi
    
    # Intermediate certificates
    if [[ $cert_count -gt 1 ]]; then
        for i in $(seq 2 $((cert_count - 1))); do
            if [[ -f "${WORK_DIR}/cert${i}.pem" ]]; then
                write_certificate_details "${WORK_DIR}/cert${i}.pem" "$i" "Intermediate Certificate $((i-1))"
            fi
        done
    fi
    
    # Root certificate
    if [[ $cert_count -gt 1 ]] && [[ -f "${WORK_DIR}/cert${cert_count}.pem" ]]; then
        write_certificate_details "${WORK_DIR}/cert${cert_count}.pem" "$cert_count" "Root Certificate"
    fi
    
    echo "---" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    # Write analysis sections
    write_detailed_analysis "$leaf_cert"
    
    local sig_status=$(analyze_signature_algorithm "$(get_cert_signature_algorithm "$leaf_cert")")
    local pubkey_status=$(analyze_public_key "$(get_cert_pubkey_info "$leaf_cert")")
    
    write_risk_assessment "$days_remaining" "$validity_days" "$sig_status" "$pubkey_status"
    write_recommendations "$days_remaining"
    write_appendix "$cert_count"
    
    log_success "Report generated: ${REPORT_FILE}"
    
    # Summary output
    echo ""
    echo "═══════════════════════════════════════════════════════════"
    echo "  Certificate Analysis Summary"
    echo "═══════════════════════════════════════════════════════════"
    echo "  Domain: ${domain}"
    echo "  Certificates in chain: ${cert_count}"
    echo "  Days until expiration: ${days_remaining}"
    echo "  Validity period: ${validity_days} days"
    echo "  Compliance: ${validity_status}"
    echo "═══════════════════════════════════════════════════════════"
    echo ""
    echo "Full report: ${REPORT_FILE}"
    echo ""
    
    return 0
}

#############################################################################
# File-based Certificate Analysis
#############################################################################

# Analyze certificates from a file
analyze_file() {
    local file="$1"
    local format="${INPUT_FORMAT:-}"
    local password="${PKCS12_PASSWORD:-}"

    WORK_DIR=$(mktemp -d)
    ANALYSIS_DATE=$(date -u +"%Y-%m-%d %H:%M:%S UTC")

    # Auto-detect format if not specified
    if [[ -z "$format" ]]; then
        format=$(detect_cert_format "$file")
        if [[ "$format" == "unknown" ]]; then
            log_error "Unable to detect certificate format for: $file"
            return 1
        fi
        log_info "Detected format: $(format_name "$format")"
    fi

    # Read password from file if specified
    if [[ -n "$PASSWORD_FILE" ]] && [[ -f "$PASSWORD_FILE" ]]; then
        password=$(read_password_file "$PASSWORD_FILE")
    fi

    # Parse certificates based on format
    log_info "Parsing certificates..."
    local cert_count=$(parse_certificates "$file" "$WORK_DIR" "$format" "$password")

    if [[ "$cert_count" -eq 0 ]]; then
        log_error "No certificates found in file"
        return 1
    fi

    log_success "Found $cert_count certificate(s)"

    # Get the leaf certificate (first one)
    local leaf_cert="$WORK_DIR/cert1.pem"

    if [[ ! -f "$leaf_cert" ]]; then
        log_error "Failed to extract certificates"
        return 1
    fi

    # Set up report file
    local basename=$(basename "$file" | sed 's/\.[^.]*$//')
    REPORT_FILE="certificate-analysis-${basename}-$(date +%Y%m%d-%H%M%S).md"
    DOMAIN="$file"

    # Show fingerprints if requested
    if [[ "$SHOW_FINGERPRINTS" == "true" ]]; then
        echo ""
        print_fingerprint_table "$leaf_cert"
        echo ""
    fi

    # Run validation checks if requested
    if [[ "$VERIFY_CHAIN" == "true" ]] || [[ "$CHECK_OCSP" == "true" ]] || [[ "$CHECK_CT" == "true" ]] || [[ "$CAB_COMPLIANCE" == "true" ]]; then
        run_validation_checks "$WORK_DIR"
    fi

    # Generate report based on output format
    case "$OUTPUT_FORMAT" in
        json)
            generate_json_output "$leaf_cert" "$cert_count"
            ;;
        text)
            generate_text_output "$leaf_cert" "$cert_count"
            ;;
        markdown|*)
            generate_full_report "$leaf_cert" "$cert_count"
            ;;
    esac

    return 0
}

# Generate JSON output
generate_json_output() {
    local leaf_cert="$1"
    local cert_count="$2"

    local subject=$(get_cert_subject "$leaf_cert")
    local issuer=$(get_cert_issuer "$leaf_cert")
    local serial=$(get_cert_serial "$leaf_cert")
    local not_before=$(get_cert_dates "$leaf_cert" "startdate")
    local not_after=$(get_cert_dates "$leaf_cert" "enddate")
    local days_remaining=$(days_until_expiry "$leaf_cert")
    local validity_days=$(cert_validity_period "$leaf_cert")
    local sig_alg=$(get_cert_signature_algorithm "$leaf_cert")
    local pubkey=$(get_cert_pubkey_info "$leaf_cert")
    local sans=$(get_cert_sans "$leaf_cert" | tr '\n' ',' | sed 's/,$//')

    # Get fingerprints
    local fp_sha256=$(get_cert_fingerprint "$leaf_cert" "sha256")
    local fp_sha1=$(get_cert_fingerprint "$leaf_cert" "sha1")

    cat << EOF
{
  "analysis_date": "$ANALYSIS_DATE",
  "source": "$DOMAIN",
  "certificate_count": $cert_count,
  "certificate": {
    "subject": "$subject",
    "issuer": "$issuer",
    "serial_number": "$serial",
    "not_before": "$not_before",
    "not_after": "$not_after",
    "days_remaining": $days_remaining,
    "validity_period_days": $validity_days,
    "signature_algorithm": "$sig_alg",
    "public_key": "$pubkey",
    "subject_alt_names": "$sans"
  },
  "fingerprints": {
    "sha256": "$fp_sha256",
    "sha1": "$fp_sha1"
  },
  "compliance": {
    "validity_compliant": $([ $validity_days -le $CURRENT_MAX_VALIDITY ] && echo "true" || echo "false"),
    "max_validity_days": $CURRENT_MAX_VALIDITY
  },
  "warnings": {
    "expiring_critical": $([ $days_remaining -le $EXPIRY_CRITICAL ] && echo "true" || echo "false"),
    "expiring_warning": $([ $days_remaining -le $EXPIRY_WARNING ] && echo "true" || echo "false")
  }
}
EOF
}

# Generate text output (simple summary)
generate_text_output() {
    local leaf_cert="$1"
    local cert_count="$2"

    local subject=$(get_cert_subject "$leaf_cert")
    local issuer=$(get_cert_issuer "$leaf_cert")
    local not_before=$(get_cert_dates "$leaf_cert" "startdate")
    local not_after=$(get_cert_dates "$leaf_cert" "enddate")
    local days_remaining=$(days_until_expiry "$leaf_cert")
    local validity_days=$(cert_validity_period "$leaf_cert")
    local sig_alg=$(get_cert_signature_algorithm "$leaf_cert")

    echo "Certificate Analysis Summary"
    echo "============================"
    echo ""
    echo "Source: $DOMAIN"
    echo "Certificates: $cert_count"
    echo ""
    echo "Subject: $subject"
    echo "Issuer: $issuer"
    echo ""
    echo "Valid From: $not_before"
    echo "Valid Until: $not_after"
    echo "Days Remaining: $days_remaining"
    echo "Validity Period: $validity_days days"
    echo ""
    echo "Signature Algorithm: $sig_alg"
    echo ""

    # Compliance status
    if [[ $validity_days -le $CURRENT_MAX_VALIDITY ]]; then
        echo "Compliance: ✓ Compliant (${validity_days} days ≤ ${CURRENT_MAX_VALIDITY} days max)"
    else
        echo "Compliance: ✗ Non-compliant (${validity_days} days > ${CURRENT_MAX_VALIDITY} days max)"
    fi

    # Expiry warnings
    if [[ $days_remaining -le $EXPIRY_CRITICAL ]]; then
        echo ""
        echo "⚠️  CRITICAL: Certificate expires in $days_remaining days!"
    elif [[ $days_remaining -le $EXPIRY_WARNING ]]; then
        echo ""
        echo "⚠️  WARNING: Certificate expires in $days_remaining days"
    fi
}

# Generate full markdown report (calls existing functions)
generate_full_report() {
    local leaf_cert="$1"
    local cert_count="$2"

    local days_remaining=$(days_until_expiry "$leaf_cert")
    local validity_days=$(cert_validity_period "$leaf_cert")
    local not_before=$(get_cert_dates "$leaf_cert" "startdate")
    local validity_status=$(analyze_validity_compliance "$validity_days" "$not_before")

    # Generate report
    log_info "Generating analysis report..."

    write_report_header
    write_executive_summary "$cert_count" "$days_remaining" "$validity_status"

    echo "## Certificate Chain Overview" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"

    # Write certificate details
    write_certificate_details "$leaf_cert" 1 "End-Entity Certificate (Leaf)"

    # Intermediate certificates
    if [[ $cert_count -gt 1 ]]; then
        for i in $(seq 2 $((cert_count - 1))); do
            if [[ -f "$WORK_DIR/cert${i}.pem" ]]; then
                write_certificate_details "$WORK_DIR/cert${i}.pem" "$i" "Intermediate Certificate $((i-1))"
            fi
        done
    fi

    # Root certificate
    if [[ $cert_count -gt 1 ]] && [[ -f "$WORK_DIR/cert${cert_count}.pem" ]]; then
        write_certificate_details "$WORK_DIR/cert${cert_count}.pem" "$cert_count" "Root Certificate"
    fi

    echo "---" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"

    # Write analysis sections
    write_detailed_analysis "$leaf_cert"

    local sig_status=$(analyze_signature_algorithm "$(get_cert_signature_algorithm "$leaf_cert")")
    local pubkey_status=$(analyze_public_key "$(get_cert_pubkey_info "$leaf_cert")")

    write_risk_assessment "$days_remaining" "$validity_days" "$sig_status" "$pubkey_status"
    write_recommendations "$days_remaining"
    write_appendix "$cert_count"

    log_success "Report generated: ${REPORT_FILE}"

    # Summary output
    echo ""
    echo "═══════════════════════════════════════════════════════════"
    echo "  Certificate Analysis Summary"
    echo "═══════════════════════════════════════════════════════════"
    echo "  Source: ${DOMAIN}"
    echo "  Certificates in chain: ${cert_count}"
    echo "  Days until expiration: ${days_remaining}"
    echo "  Validity period: ${validity_days} days"
    echo "  Compliance: ${validity_status}"
    echo "═══════════════════════════════════════════════════════════"
    echo ""
    echo "Full report: ${REPORT_FILE}"
    echo ""
}

#############################################################################
# Validation Functions (Phase 2)
#############################################################################

# Run validation checks on certificates
# Usage: run_validation_checks <certs_dir> [host] [port]
run_validation_checks() {
    local certs_dir="$1"
    local host="${2:-}"
    local port="${3:-443}"

    local leaf_cert="$certs_dir/cert1.pem"
    local issuer_cert="$certs_dir/cert2.pem"

    echo ""
    echo "═══════════════════════════════════════════════════════════"
    echo "  Validation Checks"
    echo "═══════════════════════════════════════════════════════════"
    echo ""

    # Chain Validation
    if [[ "$VERIFY_CHAIN" == "true" ]]; then
        echo "Chain Validation"
        echo "----------------"
        # Use || true to prevent script exit on validation failure (invalid chain is valid result)
        if [[ -n "$CA_FILE" ]]; then
            get_chain_validation_summary "$certs_dir" --ca-file "$CA_FILE" || true
        else
            get_chain_validation_summary "$certs_dir" || true
        fi
        echo ""
    fi

    # OCSP Check
    if [[ "$CHECK_OCSP" == "true" ]]; then
        echo "OCSP Verification"
        echo "-----------------"
        local ocsp_uri
        ocsp_uri=$(get_ocsp_uri "$leaf_cert" 2>/dev/null) || true
        echo "OCSP URI: ${ocsp_uri:-Not found}"

        local must_staple
        must_staple=$(has_ocsp_must_staple "$leaf_cert" 2>/dev/null) || true
        echo "Must-Staple Extension: ${must_staple:-NO}"

        if [[ -n "$host" ]]; then
            local stapling
            stapling=$(check_ocsp_stapling "$host" "$port" 2>/dev/null) || true
            echo "Server OCSP Stapling: ${stapling:-UNKNOWN}"

            if [[ "${must_staple:-NO}" == "YES" ]] && [[ "${stapling:-UNKNOWN}" != "SUPPORTED" ]]; then
                echo ""
                echo "⚠️  WARNING: Certificate has OCSP Must-Staple but server does not support stapling!"
            fi
        fi

        if [[ -n "$ocsp_uri" ]] && [[ -f "$issuer_cert" ]]; then
            echo ""
            echo "OCSP Status Check:"
            verify_ocsp "$leaf_cert" "$issuer_cert" "$ocsp_uri" || true
        fi
        echo ""
    fi

    # Certificate Transparency Check
    if [[ "$CHECK_CT" == "true" ]]; then
        echo "Certificate Transparency"
        echo "------------------------"
        local ct_status
        ct_status=$(check_certificate_transparency "$leaf_cert" 2>/dev/null) || echo "Unknown"
        echo "SCT Status: ${ct_status:-Unknown}"
        echo ""
    fi

    # CA/Browser Forum Compliance Check
    if [[ "$CAB_COMPLIANCE" == "true" ]]; then
        echo ""
        print_cab_compliance_report "$leaf_cert" || true
        echo ""
    fi
}

# Analyze domain with StartTLS
# Usage: analyze_domain_starttls <host> <protocol> [port]
analyze_domain_starttls() {
    local host="$1"
    local protocol="$2"
    local port="${3:-}"

    # Get default port if not specified
    if [[ -z "$port" ]]; then
        port=$(get_starttls_port "$protocol")
    fi

    log_info "Connecting to $host:$port using StartTLS ($protocol)..."

    WORK_DIR=$(mktemp -d)
    local chain_pem="$WORK_DIR/chain.pem"

    # Get certificates via StartTLS
    if ! starttls_get_certs "$protocol" "$host" "$port" "$chain_pem"; then
        log_error "Failed to retrieve certificates via StartTLS"
        return 1
    fi

    if [[ ! -s "$chain_pem" ]]; then
        log_error "No certificates retrieved from $host:$port"
        return 1
    fi

    # Parse certificates
    local cert_count=$(parse_pem_certificates "$chain_pem" "$WORK_DIR")
    log_success "Retrieved $cert_count certificate(s) via StartTLS"

    # Set up for analysis
    DOMAIN="$host:$port ($protocol)"
    ANALYSIS_DATE=$(date -u +"%Y-%m-%d %H:%M:%S UTC")
    REPORT_FILE="certificate-analysis-${host}-$(date +%Y%m%d-%H%M%S).md"

    local leaf_cert="$WORK_DIR/cert1.pem"

    # Show fingerprints if requested
    if [[ "$SHOW_FINGERPRINTS" == "true" ]]; then
        echo ""
        print_fingerprint_table "$leaf_cert"
        echo ""
    fi

    # Run validation checks if requested
    if [[ "$VERIFY_CHAIN" == "true" ]] || [[ "$CHECK_OCSP" == "true" ]] || [[ "$CHECK_CT" == "true" ]] || [[ "$CAB_COMPLIANCE" == "true" ]]; then
        run_validation_checks "$WORK_DIR" "$host" "$port"
    fi

    # Generate output
    case "$OUTPUT_FORMAT" in
        json)
            generate_json_output "$leaf_cert" "$cert_count"
            ;;
        text)
            generate_text_output "$leaf_cert" "$cert_count"
            ;;
        markdown|*)
            generate_full_report "$leaf_cert" "$cert_count"
            ;;
    esac

    return 0
}

#############################################################################
# Help and Version
#############################################################################

show_help() {
    cat << EOF
Certificate Analyser v${SCRIPT_VERSION}
Analyzes digital certificates for compliance with CA/Browser Forum policies

Usage: $0 [OPTIONS] <target>

TARGET:
  <domain>              Domain name to analyze (e.g., example.com)
  --file <path>         Certificate file to analyze
  --stdin               Read certificate from stdin

INPUT OPTIONS:
  --format <fmt>        Force input format: pem|der|pkcs7|pkcs12 (default: auto-detect)
  --password <pass>     Password for PKCS12 files
  --password-file <f>   Read password from file
  --port <port>         Port for domain connections (default: 443)
  --starttls <proto>    Use StartTLS: smtp|mysql|postgres|ldap|imap|ftp|pop3|xmpp

OUTPUT OPTIONS:
  -o, --output <fmt>    Output format: markdown|json|text (default: markdown)
  --output-file <file>  Write output to file
  --fingerprint [algo]  Show fingerprints (sha1|sha256|sha384|sha512|all)

VALIDATION OPTIONS:
  --verify-chain        Validate certificate chain against trust store
  --check-ocsp          Check OCSP status and stapling support
  --check-ct            Verify Certificate Transparency (SCT)
  --compliance          Run CA/Browser Forum Baseline Requirements compliance check
  --all-checks          Run all validation checks (chain, OCSP, CT, compliance)
  --ca-file <file>      Custom CA bundle for chain validation

CLAUDE AI OPTIONS:
  --claude              Enable Claude AI enhanced analysis (comprehensive)
  --claude-quick        Quick risk assessment (concise output)
  --claude-compliance   Compliance-focused analysis (CA/B Forum audit)
  --advanced            Full analysis: --claude + --all-checks combined
  --no-claude           Disable Claude AI (even if API key set)

  Claude AI uses RAG knowledge base for context-aware analysis.
  Requires: ANTHROPIC_API_KEY environment variable or .env file

GENERAL OPTIONS:
  -v, --verbose         Verbose output
  --version             Show version
  -h, --help            Show this help

EXAMPLES:
  # Analyze domain certificate
  $0 example.com

  # Analyze with chain validation and OCSP check
  $0 --verify-chain --check-ocsp example.com

  # Analyze local PEM file
  $0 --file certificate.pem

  # Analyze PKCS12 keystore
  $0 --file keystore.p12 --password mypassword

  # Analyze with all validation checks
  $0 --all-checks example.com

  # Get JSON output with fingerprints
  $0 --file cert.pem -o json --fingerprint sha256

  # Use StartTLS for SMTP server
  $0 --starttls smtp mail.example.com

  # Read from stdin
  cat cert.pem | $0 --stdin

  # With Claude AI analysis (comprehensive)
  $0 --claude example.com

  # Quick Claude risk assessment
  $0 --claude-quick example.com

  # Claude compliance-focused audit
  $0 --claude-compliance example.com

  # Full advanced analysis (all checks + Claude)
  $0 --advanced example.com

EOF
}

show_version() {
    echo "Certificate Analyser v${SCRIPT_VERSION}"
    echo "Copyright (c) 2025 Crash Override Inc."
}

#############################################################################
# Main Entry Point
#############################################################################

main() {
    # Parse arguments
    local target=""

    while [[ $# -gt 0 ]]; do
        case $1 in
            --file)
                INPUT_FILE="$2"
                INPUT_SOURCE="file"
                shift 2
                ;;
            --stdin)
                INPUT_SOURCE="stdin"
                shift
                ;;
            --format)
                INPUT_FORMAT="$2"
                shift 2
                ;;
            --password)
                PKCS12_PASSWORD="$2"
                shift 2
                ;;
            --password-file)
                PASSWORD_FILE="$2"
                shift 2
                ;;
            --port)
                PORT="$2"
                PORT_EXPLICIT=true
                shift 2
                ;;
            --starttls)
                STARTTLS_PROTO="$2"
                shift 2
                ;;
            --verify-chain)
                VERIFY_CHAIN=true
                shift
                ;;
            --check-ocsp)
                CHECK_OCSP=true
                shift
                ;;
            --check-ct)
                CHECK_CT=true
                shift
                ;;
            --all-checks)
                ALL_CHECKS=true
                VERIFY_CHAIN=true
                CHECK_OCSP=true
                CHECK_CT=true
                CAB_COMPLIANCE=true
                shift
                ;;
            --compliance)
                CAB_COMPLIANCE=true
                shift
                ;;
            --ca-file)
                CA_FILE="$2"
                shift 2
                ;;
            -o|--output)
                OUTPUT_FORMAT="$2"
                shift 2
                ;;
            --output-file)
                OUTPUT_FILE="$2"
                shift 2
                ;;
            --fingerprint)
                SHOW_FINGERPRINTS=true
                if [[ -n "${2:-}" ]] && [[ ! "$2" =~ ^- ]]; then
                    FINGERPRINT_ALGO="$2"
                    shift
                fi
                shift
                ;;
            --claude)
                USE_CLAUDE=true
                shift
                ;;
            --claude-quick)
                USE_CLAUDE=true
                CLAUDE_ANALYSIS_TYPE="quick"
                shift
                ;;
            --claude-compliance)
                USE_CLAUDE=true
                CLAUDE_ANALYSIS_TYPE="compliance"
                shift
                ;;
            --advanced)
                USE_CLAUDE=true
                CLAUDE_ANALYSIS_TYPE="comprehensive"
                ALL_CHECKS=true
                VERIFY_CHAIN=true
                CHECK_OCSP=true
                CHECK_CT=true
                CAB_COMPLIANCE=true
                shift
                ;;
            --no-claude)
                USE_CLAUDE=false
                shift
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            --version)
                show_version
                exit 0
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            -*)
                log_error "Unknown option: $1"
                echo "Use --help for usage information"
                exit 1
                ;;
            *)
                target="$1"
                shift
                ;;
        esac
    done

    # Check dependencies
    if ! command -v openssl &> /dev/null; then
        log_error "openssl is required but not installed"
        exit 1
    fi

    # Initialize Claude cost tracking if enabled
    if [[ "$USE_CLAUDE" == "true" ]] && type init_cost_tracking &>/dev/null; then
        init_cost_tracking
    fi

    # Determine input source and run analysis
    case "$INPUT_SOURCE" in
        stdin)
            if is_stdin_input; then
                local temp_file=$(read_stdin_to_temp)
                INPUT_FILE="$temp_file"
                analyze_file "$INPUT_FILE"
                rm -f "$temp_file"
            else
                log_error "No input provided on stdin"
                exit 1
            fi
            ;;
        file)
            if [[ -z "$INPUT_FILE" ]]; then
                log_error "No file specified. Use --file <path>"
                exit 1
            fi
            if [[ ! -f "$INPUT_FILE" ]]; then
                log_error "File not found: $INPUT_FILE"
                exit 1
            fi
            analyze_file "$INPUT_FILE"
            ;;
        domain|*)
            if [[ -z "$target" ]]; then
                echo "Usage: $0 [OPTIONS] <target>"
                echo ""
                echo "Use --help for more information"
                exit 1
            fi

            # Remove protocol if present
            target=$(echo "$target" | sed 's|^https\?://||' | sed 's|/.*||')
            DOMAIN="$target"

            # Check if StartTLS mode
            if [[ -n "$STARTTLS_PROTO" ]]; then
                # Only pass port if explicitly set, otherwise let StartTLS use protocol default
                if [[ "$PORT_EXPLICIT" == "true" ]]; then
                    analyze_domain_starttls "$target" "$STARTTLS_PROTO" "$PORT"
                else
                    analyze_domain_starttls "$target" "$STARTTLS_PROTO" ""
                fi
            else
                # Run domain analysis (existing functionality)
                analyze_domain "$target"

                # Run validation checks if requested (for domain mode)
                if [[ "$VERIFY_CHAIN" == "true" ]] || [[ "$CHECK_OCSP" == "true" ]] || [[ "$CHECK_CT" == "true" ]] || [[ "$CAB_COMPLIANCE" == "true" ]]; then
                    if [[ -d "$WORK_DIR" ]]; then
                        run_validation_checks "$WORK_DIR" "$target" "$PORT"
                    fi
                fi
            fi
            ;;
    esac

    # If Claude mode, analyze the report
    if [[ "$USE_CLAUDE" == "true" ]]; then
        if [[ -f "$REPORT_FILE" ]]; then
            local report_content=$(cat "$REPORT_FILE")

            echo ""
            echo "═══════════════════════════════════════════════════════════"
            echo "  Claude AI Enhanced Analysis (${CLAUDE_ANALYSIS_TYPE})"
            echo "═══════════════════════════════════════════════════════════"
            echo ""

            local claude_analysis
            claude_analysis=$(analyze_with_claude "$report_content")

            if [[ -n "$claude_analysis" ]]; then
                # Display to console
                echo "$claude_analysis"

                # Append to report file
                append_claude_analysis_to_report "$REPORT_FILE" "$claude_analysis"
                log_success "Claude analysis appended to report: ${REPORT_FILE}"
            else
                log_error "Claude analysis returned empty - check API key and connectivity"
            fi

            echo ""

            # Display cost summary
            if type display_cost_summary &>/dev/null; then
                display_cost_summary
                echo ""
            fi
        else
            log_warning "No report file found for Claude analysis"
        fi
    fi
}

main "$@"
