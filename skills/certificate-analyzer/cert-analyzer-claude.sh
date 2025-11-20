#!/bin/bash
# Copyright (c) 2024 Crash Override Inc
# 101 Fulton St, 416, New York 10038
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Certificate Analysis Script with Claude AI Integration
# Analyzes digital certificates and uses Claude for intelligent recommendations
# Usage: ./cert-analyzer.sh <domain>
# 
# Environment Variables:
#   ANTHROPIC_API_KEY - Required for Claude AI analysis (or use --api-key flag)
#############################################################################

set -euo pipefail

# Load environment variables from .env file if it exists
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
if [ -f "$REPO_ROOT/.env" ]; then
    source "$REPO_ROOT/.env"
fi

# Colors for output
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
CURRENT_MAX_VALIDITY=398  # Current CA/B Forum limit (days)
EXPIRY_CRITICAL=7         # Critical if expires within X days
EXPIRY_WARNING=30         # Warning if expires within X days
CLAUDE_MODEL="claude-sonnet-4-20250514"
CLAUDE_API_URL="https://api.anthropic.com/v1/messages"

# Global variables
DOMAIN=""
WORK_DIR=""
REPORT_FILE=""
ANALYSIS_DATE=""
API_KEY=""
USE_CLAUDE=false

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

log_claude() {
    echo -e "${CYAN}[CLAUDE]${NC} $1"
}

cleanup() {
    if [[ -n "${WORK_DIR:-}" ]] && [[ -d "$WORK_DIR" ]]; then
        rm -rf "$WORK_DIR"
    fi
}

trap cleanup EXIT

show_usage() {
    cat << EOF
Certificate Analysis Script with Claude AI Integration

Usage: $0 [OPTIONS] <domain>

Options:
    --api-key KEY       Anthropic API key (or set ANTHROPIC_API_KEY env var)
    --no-claude         Skip Claude AI analysis (basic analysis only)
    -h, --help          Show this help message

Examples:
    # Using environment variable
    export ANTHROPIC_API_KEY="sk-ant-..."
    $0 example.com

    # Using command line flag
    $0 --api-key sk-ant-... example.com

    # Basic analysis without Claude
    $0 --no-claude example.com

Environment Variables:
    ANTHROPIC_API_KEY   API key for Claude integration

EOF
}

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

get_cert_full_text() {
    local cert_file=$1
    openssl x509 -in "$cert_file" -noout -text 2>/dev/null
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
# Claude API Integration
#############################################################################

call_claude_api() {
    local prompt=$1
    local max_tokens=${2:-4000}
    
    if [[ -z "$API_KEY" ]]; then
        log_error "No API key provided. Use --api-key flag or set ANTHROPIC_API_KEY"
        return 1
    fi
    
    log_claude "Analyzing with Claude AI..."
    
    # Create JSON payload with proper escaping
    local json_prompt=$(echo "$prompt" | jq -Rs .)
    
    local payload=$(cat <<EOF
{
  "model": "${CLAUDE_MODEL}",
  "max_tokens": ${max_tokens},
  "messages": [
    {
      "role": "user",
      "content": ${json_prompt}
    }
  ]
}
EOF
)
    
    # Call API
    local response=$(curl -s -w "\n%{http_code}" "${CLAUDE_API_URL}" \
        -H "Content-Type: application/json" \
        -H "x-api-key: ${API_KEY}" \
        -H "anthropic-version: 2023-06-01" \
        -d "$payload")
    
    # Extract HTTP status code and response body
    local http_code=$(echo "$response" | tail -n1)
    local response_body=$(echo "$response" | sed '$d')
    
    if [[ "$http_code" != "200" ]]; then
        log_error "Claude API returned status code: $http_code"
        log_error "Response: $response_body"
        return 1
    fi
    
    # Extract text content from response
    echo "$response_body" | jq -r '.content[0].text' 2>/dev/null
    
    if [[ $? -ne 0 ]]; then
        log_error "Failed to parse Claude API response"
        return 1
    fi
    
    return 0
}

build_claude_analysis_prompt() {
    local domain=$1
    local cert_count=$2
    local work_dir=$3
    
    # Gather certificate data
    local leaf_cert="${work_dir}/cert1.pem"
    local days_remaining=$(days_until_expiry "$leaf_cert")
    local validity_days=$(cert_validity_period "$leaf_cert")
    local subject=$(get_cert_subject "$leaf_cert")
    local issuer=$(get_cert_issuer "$leaf_cert")
    local not_before=$(get_cert_dates "$leaf_cert" "startdate")
    local not_after=$(get_cert_dates "$leaf_cert" "enddate")
    local sig_alg=$(get_cert_signature_algorithm "$leaf_cert")
    local pubkey=$(get_cert_pubkey_info "$leaf_cert")
    local sans=$(get_cert_sans "$leaf_cert")
    local ct_status=$(check_certificate_transparency "$leaf_cert")
    local ocsp=$(check_ocsp_uri "$leaf_cert")
    local crl=$(get_crl_uri "$leaf_cert")
    
    # Get full certificate details
    local cert_details=""
    for i in $(seq 1 "$cert_count"); do
        if [[ -f "${work_dir}/cert${i}.pem" ]]; then
            cert_details+="
=== Certificate ${i} ===
$(get_cert_full_text "${work_dir}/cert${i}.pem")

"
        fi
    done
    
    # Build prompt
    cat <<EOF
You are a digital certificate security expert analyzing certificates for compliance with CA/Browser Forum policies.

## Domain Analysis Request

**Domain**: ${domain}
**Analysis Date**: ${ANALYSIS_DATE}

## Certificate Chain Summary

**Total Certificates**: ${cert_count}
**Days Until Expiration**: ${days_remaining}
**Total Validity Period**: ${validity_days} days

### Leaf Certificate Quick Facts
- **Subject**: ${subject}
- **Issuer**: ${issuer}
- **Valid From**: ${not_before}
- **Valid Until**: ${not_after}
- **Signature Algorithm**: ${sig_alg}
- **Public Key**: ${pubkey}
- **SANs**: ${sans}
- **Certificate Transparency**: ${ct_status}
- **OCSP**: ${ocsp}
- **CRL**: ${crl}

## Complete Certificate Chain Details

${cert_details}

## Analysis Requirements

Provide a comprehensive security analysis covering:

1. **CA/Browser Forum Compliance Assessment**
   - Current 398-day policy compliance
   - Preparation for future 90-180 day policies
   - Any legacy certificate concerns

2. **Cryptographic Security Analysis**
   - Signature algorithm strength and currency
   - Public key algorithm and size adequacy
   - Any deprecated or weak cryptography concerns

3. **Operational Security Posture**
   - Certificate Transparency compliance
   - Revocation checking mechanisms (OCSP, CRL)
   - Chain validation status
   - Domain coverage (SANs analysis)

4. **Risk Assessment** (Categorize by severity)
   - **Critical Issues (❌)**: Require immediate action
   - **Warnings (⚠️)**: Should be addressed soon
   - **Informational (ℹ️)**: Best practices and opportunities

5. **Actionable Recommendations** (Prioritize by timeline)
   - **Immediate Actions (0-7 days)**: Urgent items
   - **Short-term (7-30 days)**: Near-term planning
   - **Medium-term (30-90 days)**: Implementation phase
   - **Long-term (90+ days)**: Strategic initiatives

## Special Considerations

- This domain is ${domain} - consider its use case and security requirements
- Days until expiration: ${days_remaining} (Critical if ≤${EXPIRY_CRITICAL}, Warning if ≤${EXPIRY_WARNING})
- Current validity period: ${validity_days} days (Policy limit: ${CURRENT_MAX_VALIDITY} days)

## Output Format

Provide your analysis in clear markdown format with:
- Concise executive summary (2-3 sentences)
- Detailed findings organized by the sections above
- Specific, actionable recommendations with clear priorities
- Technical justification for each recommendation

Focus on practical, actionable guidance that a security or ops team can implement.
EOF
}

#############################################################################
# Report Generation Functions
#############################################################################

write_report_header() {
    cat >> "$REPORT_FILE" <<EOF
# Certificate Analysis Report

**Domain**: ${DOMAIN}  
**Analysis Date**: ${ANALYSIS_DATE}  
**Analysis Method**: $(if [[ "$USE_CLAUDE" == true ]]; then echo "Claude AI Enhanced Analysis"; else echo "Basic Automated Analysis"; fi)

---

EOF
}

write_basic_summary() {
    local total_certs=$1
    local days_remaining=$2
    local validity_status=$3
    
    cat >> "$REPORT_FILE" <<EOF
## Executive Summary

This report analyzes the complete certificate chain for **${DOMAIN}**, comprising ${total_certs} certificate(s).

**Key Findings:**
- Certificate expires in **${days_remaining} days** ($(date -d "+${days_remaining} days" +%Y-%m-%d))
- Validity period compliance: **${validity_status}**

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
    local cert_type=$2
    
    local subject=$(get_cert_subject "$cert_file")
    local issuer=$(get_cert_issuer "$cert_file")
    local serial=$(get_cert_serial "$cert_file")
    local not_before=$(get_cert_dates "$cert_file" "startdate")
    local not_after=$(get_cert_dates "$cert_file" "enddate")
    local days_left=$(days_until_expiry "$cert_file")
    local validity_days=$(cert_validity_period "$cert_file")
    local sig_alg=$(get_cert_signature_algorithm "$cert_file")
    local pubkey=$(get_cert_pubkey_info "$cert_file")
    
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

write_claude_analysis() {
    local claude_output=$1
    
    cat >> "$REPORT_FILE" <<EOF
## Claude AI Security Analysis

${claude_output}

---

EOF
}

write_certificate_chain_section() {
    local cert_count=$1
    
    echo "## Certificate Chain Overview" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    if [[ -f "${WORK_DIR}/cert1.pem" ]]; then
        write_certificate_details "${WORK_DIR}/cert1.pem" "End-Entity Certificate (Leaf)"
    fi
    
    if [[ $cert_count -gt 1 ]]; then
        for i in $(seq 2 $((cert_count - 1))); do
            if [[ -f "${WORK_DIR}/cert${i}.pem" ]]; then
                write_certificate_details "${WORK_DIR}/cert${i}.pem" "Intermediate Certificate $((i-1))"
            fi
        done
    fi
    
    if [[ $cert_count -gt 1 ]] && [[ -f "${WORK_DIR}/cert${cert_count}.pem" ]]; then
        write_certificate_details "${WORK_DIR}/cert${cert_count}.pem" "Root Certificate"
    fi
    
    echo "---" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
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
$(get_cert_full_text "${WORK_DIR}/cert${i}.pem")
\`\`\`

EOF
        fi
    done
    
    cat >> "$REPORT_FILE" <<EOF
---

## Methodology Notes

**Tools Used**: OpenSSL $(openssl version | awk '{print $2}')  
**Claude Model**: ${CLAUDE_MODEL}  
**Analysis Date**: ${ANALYSIS_DATE}  
**Script Version**: 2.0 (Claude-Enhanced)

---

*This analysis combines automated certificate parsing with Claude AI's security expertise for comprehensive, actionable recommendations.*
EOF
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
    
    # Count certificates
    local cert_count=$(ls -1 "${WORK_DIR}"/cert*.pem 2>/dev/null | wc -l)
    local leaf_cert="${WORK_DIR}/cert1.pem"
    local days_remaining=$(days_until_expiry "$leaf_cert")
    local validity_days=$(cert_validity_period "$leaf_cert")
    local not_before=$(get_cert_dates "$leaf_cert" "startdate")
    local validity_status=$(analyze_validity_compliance "$validity_days" "$not_before")
    
    # Generate report
    log_info "Generating analysis report..."
    
    write_report_header
    write_basic_summary "$cert_count" "$days_remaining" "$validity_status"
    write_certificate_chain_section "$cert_count"
    
    # Claude AI Analysis
    if [[ "$USE_CLAUDE" == true ]]; then
        log_claude "Requesting Claude AI analysis..."
        local prompt=$(build_claude_analysis_prompt "$domain" "$cert_count" "$WORK_DIR")
        local claude_response=$(call_claude_api "$prompt" 8000)
        
        if [[ $? -eq 0 ]] && [[ -n "$claude_response" ]]; then
            log_success "Claude AI analysis complete"
            write_claude_analysis "$claude_response"
        else
            log_warning "Claude AI analysis failed, report contains basic analysis only"
        fi
    else
        log_info "Skipping Claude AI analysis (use --api-key to enable)"
    fi
    
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
    if [[ "$USE_CLAUDE" == true ]]; then
        echo "  Claude AI Analysis: ✓ Enabled"
    else
        echo "  Claude AI Analysis: ✗ Disabled"
    fi
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
    while [[ $# -gt 0 ]]; do
        case $1 in
            --api-key)
                API_KEY="$2"
                USE_CLAUDE=true
                shift 2
                ;;
            --no-claude)
                USE_CLAUDE=false
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            -*)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
            *)
                DOMAIN="$1"
                shift
                ;;
        esac
    done
    
    # Check for API key in environment if not provided via flag
    if [[ -z "$API_KEY" ]] && [[ -n "${ANTHROPIC_API_KEY:-}" ]]; then
        API_KEY="$ANTHROPIC_API_KEY"
        USE_CLAUDE=true
    fi
    
    # Validate domain provided
    if [[ -z "$DOMAIN" ]]; then
        log_error "No domain specified"
        show_usage
        exit 1
    fi
    
    # Remove protocol if present
    DOMAIN=$(echo "$DOMAIN" | sed 's|^https\?://||' | sed 's|/.*||')
    
    # Check dependencies
    if ! command -v openssl &> /dev/null; then
        log_error "openssl is required but not installed"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        log_error "jq is required for Claude API integration but not installed"
        log_error "Install with: apt-get install jq (Ubuntu) or brew install jq (macOS)"
        exit 1
    fi
    
    if ! command -v curl &> /dev/null; then
        log_error "curl is required for Claude API integration but not installed"
        exit 1
    fi
    
    # Warn if no API key and claude not explicitly disabled
    if [[ -z "$API_KEY" ]] && [[ "$USE_CLAUDE" != false ]]; then
        log_warning "No API key provided - Claude AI analysis disabled"
        log_warning "Set ANTHROPIC_API_KEY or use --api-key to enable AI-powered analysis"
        USE_CLAUDE=false
    fi
    
    analyze_domain "$DOMAIN"
}

main "$@"
