#!/bin/bash
# Copyright (c) 2024 Gibson Powers Contributors
# 
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Certificate Analysis Script
# Analyzes digital certificates for compliance with CA/Browser Forum policies
# Usage: ./cert-analyzer.sh <domain>
#############################################################################

set -euo pipefail

# Colors for output
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

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

# Load cost tracking if using Claude
if [[ "$USE_CLAUDE" == "true" ]]; then
    REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
    if [ -f "$REPO_ROOT/utils/lib/claude-cost.sh" ]; then
        source "$REPO_ROOT/utils/lib/claude-cost.sh"
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
# Certificate Retrieval
#############################################################################

retrieve_certificates() {
    local domain=$1
    local output_dir=$2
    
    log_info "Retrieving certificates for ${domain}..."
    
    # Retrieve full certificate chain
    if ! echo | timeout 10 openssl s_client -showcerts -servername "$domain" \
        -connect "${domain}:443" 2>/dev/null > "${output_dir}/chain.txt"; then
        log_error "Failed to retrieve certificates from ${domain}:443"
        log_error "Possible issues: domain unreachable, no HTTPS, firewall blocking"
        return 1
    fi
    
    # Split certificates into individual files
    awk '/BEGIN CERTIFICATE/,/END CERTIFICATE/{ 
        if(/BEGIN CERTIFICATE/){a++}; 
        print > "'"${output_dir}"'/cert" a ".pem"
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
    openssl x509 -in "$cert_file" -noout -"${date_type}" 2>/dev/null | sed "s/${date_type}=//"
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
# Date Calculation Functions
#############################################################################

days_until_expiry() {
    local cert_file=$1
    local end_date=$(get_cert_dates "$cert_file" "enddate")
    local end_epoch=$(date -d "$end_date" +%s 2>/dev/null || echo "0")
    local now_epoch=$(date +%s)
    local diff_seconds=$((end_epoch - now_epoch))
    local diff_days=$((diff_seconds / 86400))
    echo "$diff_days"
}

cert_validity_period() {
    local cert_file=$1
    local start_date=$(get_cert_dates "$cert_file" "startdate")
    local end_date=$(get_cert_dates "$cert_file" "enddate")
    local start_epoch=$(date -d "$start_date" +%s 2>/dev/null || echo "0")
    local end_epoch=$(date -d "$end_date" +%s 2>/dev/null || echo "0")
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
    elif [[ $(date -d "$issue_date" +%s) -lt $(date -d "2020-09-01" +%s) ]]; then
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
**Analyst**: Certificate Security Analyzer v1.0

---

## Executive Summary

EOF
}

write_executive_summary() {
    local total_certs=$1
    local days_remaining=$2
    local validity_status=$3
    
    cat >> "$REPORT_FILE" <<EOF
This report analyzes the complete certificate chain for **${DOMAIN}**, comprising ${total_certs} certificate(s). 
The analysis focuses on compliance with current CA/Browser Forum policies, cryptographic strength, 
and operational security posture.

**Key Findings:**
- Certificate expires in **${days_remaining} days** ($(date -d "+${days_remaining} days" +%Y-%m-%d))
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
# Claude AI Analysis
#############################################################################

analyze_with_claude() {
    local report_content="$1"
    local model="claude-sonnet-4-20250514"

    if [[ -z "$ANTHROPIC_API_KEY" ]]; then
        log_error "ANTHROPIC_API_KEY required for --claude mode"
        exit 1
    fi

    log_info "Analyzing with Claude AI..."

    local prompt="Analyze this SSL/TLS certificate analysis report and provide enhanced security insights. Focus on:
1. Risk prioritization and immediate actions required
2. Security posture assessment (critical, high, medium, low risk)
3. Certificate management best practices compliance
4. Cryptographic strength evaluation and recommendations
5. Expiration management and renewal planning
6. Industry standards compliance (CA/B Forum, NIST, etc.)
7. Specific remediation steps prioritized by impact
8. Long-term certificate management strategy
9. Common pitfalls and how to avoid them
10. Automation opportunities for certificate lifecycle management

Certificate Analysis Report:
$report_content"

    local response=$(curl -s https://api.anthropic.com/v1/messages \
        -H "content-type: application/json" \
        -H "x-api-key: $ANTHROPIC_API_KEY" \
        -H "anthropic-version: 2023-06-01" \
        -d "{
            \"model\": \"$model\",
            \"max_tokens\": 4096,
            \"messages\": [{
                \"role\": \"user\",
                \"content\": $(echo "$prompt" | jq -Rs .)
            }]
        }")

    if command -v record_api_usage &> /dev/null; then
        record_api_usage "$response" "$model" > /dev/null
    fi

    echo "$response" | jq -r '.content[0].text // empty'
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
# Main Entry Point
#############################################################################

main() {
    # Parse arguments
    local domain=""
    while [[ $# -gt 0 ]]; do
        case $1 in
            --claude)
                USE_CLAUDE=true
                shift
                ;;
            -h|--help)
                echo "Usage: $0 [OPTIONS] <domain>"
                echo ""
                echo "OPTIONS:"
                echo "  --claude    Use Claude AI for enhanced analysis (requires ANTHROPIC_API_KEY)"
                echo "  -h, --help  Show this help message"
                echo ""
                echo "Example: $0 example.com"
                echo "Example: $0 --claude example.com"
                echo ""
                echo "Analyzes digital certificates for compliance with CA/Browser Forum policies"
                exit 0
                ;;
            *)
                domain="$1"
                shift
                ;;
        esac
    done

    if [[ -z "$domain" ]]; then
        echo "Usage: $0 [OPTIONS] <domain>"
        echo ""
        echo "Use --help for more information"
        exit 1
    fi

    # Remove protocol if present
    domain=$(echo "$domain" | sed 's|^https\?://||' | sed 's|/.*||')

    # Check dependencies
    if ! command -v openssl &> /dev/null; then
        log_error "openssl is required but not installed"
        exit 1
    fi

    # Run analysis
    analyze_domain "$domain"

    # If Claude mode, analyze the report
    if [[ "$USE_CLAUDE" == "true" ]]; then
        if [[ -f "$REPORT_FILE" ]]; then
            local report_content=$(cat "$REPORT_FILE")

            echo ""
            echo "═══════════════════════════════════════════════════════════"
            echo "  Claude AI Enhanced Analysis"
            echo "═══════════════════════════════════════════════════════════"
            echo ""

            local claude_analysis=$(analyze_with_claude "$report_content")
            echo "$claude_analysis"

            echo ""

            # Display cost summary
            if command -v display_api_cost_summary &> /dev/null; then
                display_api_cost_summary
                echo ""
            fi
        fi
    fi
}

main "$@"
