#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Claude AI Analysis Library for Certificate Analyser
# Provides intelligent, RAG-enhanced certificate security analysis
#############################################################################

# Default model for analysis
CLAUDE_MODEL="${CLAUDE_MODEL:-claude-sonnet-4-20250514}"

# RAG Server configuration (future enhancement)
# When RAG_SERVER_URL is set, will use semantic search via RAG server
# Falls back to local filesystem when not configured or unreachable
RAG_SERVER_URL="${RAG_SERVER_URL:-}"
RAG_API_KEY="${RAG_API_KEY:-}"

#############################################################################
# Load RAG Knowledge Base
#############################################################################

# Load certificate analysis RAG documentation
# Currently uses local filesystem; future versions will support:
# - Pinecone, Weaviate, ChromaDB, Qdrant vector databases
# - Semantic search for more relevant context selection
# - Hybrid retrieval combining vector + keyword search
load_certificate_rag() {
    local rag_context=""
    local repo_root="$1"
    local rag_dir="$repo_root/rag/certificate-analysis"

    # Load X.509 certificate knowledge
    if [[ -f "$rag_dir/x509/x509-certificates.md" ]]; then
        rag_context+="# X.509 Certificate Knowledge\n\n"
        # Get key sections: structure, extensions, algorithms
        rag_context+=$(sed -n '/^## Certificate Structure/,/^## [A-Z]/p' "$rag_dir/x509/x509-certificates.md" | head -100)
        rag_context+="\n\n"
        rag_context+=$(sed -n '/^## Certificate Extensions/,/^## [A-Z]/p' "$rag_dir/x509/x509-certificates.md" | head -80)
        rag_context+="\n\n"
    fi

    # Load CA/Browser Forum requirements
    if [[ -f "$rag_dir/cab-forum/baseline-requirements.md" ]]; then
        rag_context+="# CA/Browser Forum Baseline Requirements\n\n"
        # Get key requirements section
        rag_context+=$(sed -n '/^## Key Requirements/,/^## Validation Types/p' "$rag_dir/cab-forum/baseline-requirements.md" | head -150)
        rag_context+="\n\n"
    fi

    # Load TLS security best practices
    if [[ -f "$rag_dir/tls-security/best-practices.md" ]]; then
        rag_context+="# TLS Security Best Practices\n\n"
        # Get cipher suite and configuration sections
        rag_context+=$(sed -n '/^## Cipher Suite Selection/,/^## Server Configuration/p' "$rag_dir/tls-security/best-practices.md" | head -60)
        rag_context+="\n\n"
        rag_context+=$(sed -n '/^## Best Practices Summary/,/^## References/p' "$rag_dir/tls-security/best-practices.md" | head -50)
        rag_context+="\n\n"
    fi

    # Load revocation knowledge
    if [[ -f "$rag_dir/revocation/ocsp-crl.md" ]]; then
        rag_context+="# Certificate Revocation (OCSP/CRL)\n\n"
        rag_context+=$(sed -n '/^## Best Practices/,/^## Troubleshooting/p' "$rag_dir/revocation/ocsp-crl.md" | head -40)
        rag_context+="\n\n"
    fi

    echo -e "$rag_context"
}

#############################################################################
# Build Analysis Prompt
#############################################################################

build_certificate_analysis_prompt() {
    local cert_data="$1"
    local rag_context="$2"
    local analysis_type="${3:-comprehensive}"  # comprehensive, quick, compliance

    local prompt=""

    case "$analysis_type" in
        quick)
            prompt="You are a certificate security expert. Provide a QUICK SECURITY ASSESSMENT of this certificate.

# Certificate Knowledge Base
$rag_context

# Focus Areas
1. Critical issues requiring immediate attention
2. Key compliance gaps
3. Top 3 priority actions

Keep the response concise (under 500 words).

# Certificate Data:
$cert_data"
            ;;

        compliance)
            prompt="You are a CA/Browser Forum compliance expert. Analyze this certificate for BASELINE REQUIREMENTS COMPLIANCE.

# CA/Browser Forum Knowledge Base
$rag_context

# Compliance Analysis Requirements

For each requirement, provide:
1. **Requirement**: Name and description
2. **Status**: ✓ Pass, ⚠️ Warning, or ❌ Fail
3. **Evidence**: Certificate values that demonstrate compliance/non-compliance
4. **Remediation**: If non-compliant, specific steps to fix

## Required Checks:
1. Validity Period (≤398 days since Sept 2020)
2. Key Size (RSA ≥2048 bits, ECC ≥P-256)
3. Signature Algorithm (SHA-256 or stronger)
4. Subject Alternative Name (must be present)
5. Basic Constraints (correct for certificate type)
6. Key Usage (appropriate for TLS server)
7. Certificate Transparency (SCTs present)

## Output Format:

### Compliance Summary
| Requirement | Status | Notes |
|-------------|--------|-------|
| ... | ... | ... |

### Detailed Findings
[For each non-compliant item, provide remediation steps]

### Audit-Ready Documentation
[Summary suitable for compliance evidence]

# Certificate Data:
$cert_data"
            ;;

        comprehensive|*)
            prompt="You are a senior certificate security expert. Analyze this certificate data and provide a COMPREHENSIVE SECURITY ASSESSMENT with actionable recommendations.

# Certificate Security Knowledge Base
$rag_context

# Analysis Requirements

You MUST evaluate the following areas:

## 1. Security Posture Assessment
- Overall risk level: CRITICAL / HIGH / MEDIUM / LOW
- Key vulnerabilities and exposures
- Attack surface analysis

## 2. Cryptographic Strength Analysis
- Public key algorithm and size assessment
- Signature algorithm security
- Comparison to current NIST/NSA recommendations
- Future-proofing considerations (post-quantum readiness)

## 3. CA/Browser Forum Compliance
- Validity period compliance
- Required extensions present
- Certificate Transparency status
- OCSP/CRL availability

## 4. Operational Security
- Days until expiration and renewal urgency
- Certificate chain completeness
- Trust anchor validation
- OCSP stapling recommendation

## 5. Configuration Recommendations
- Immediate actions (0-7 days)
- Short-term improvements (7-30 days)
- Long-term strategic changes (30+ days)

# Output Format

## 🔒 Security Assessment Summary

| Metric | Value | Status |
|--------|-------|--------|
| Overall Risk | [CRITICAL/HIGH/MEDIUM/LOW] | [emoji] |
| Key Strength | [algorithm/size] | [✓/⚠️/❌] |
| Signature | [algorithm] | [✓/⚠️/❌] |
| Validity | [days] | [✓/⚠️/❌] |
| Compliance | [status] | [✓/⚠️/❌] |
| CT Status | [present/missing] | [✓/⚠️/❌] |

## 🔴 Critical Issues (Immediate Action Required)
[List any critical issues with specific remediation steps]

## 🟠 Warnings (Address Within 30 Days)
[List warnings with context and recommendations]

## 🟢 Informational (Best Practices)
[Suggestions for improvement]

## 📋 Prioritized Action Plan

### Immediate (0-7 days)
1. [Action with specific command or step]
2. ...

### Short-term (7-30 days)
1. [Action with specific guidance]
2. ...

### Long-term (30+ days)
1. [Strategic improvement]
2. ...

## 🔧 Technical Recommendations

### Certificate Configuration
[Specific recommendations for certificate settings]

### Server Configuration
[TLS/SSL server configuration recommendations]

### Automation Opportunities
[ACME, cert-manager, or other automation suggestions]

## 📊 Compliance Status

| Requirement | Status | Evidence |
|-------------|--------|----------|
| Validity ≤398 days | [✓/❌] | [actual value] |
| Key Size | [✓/❌] | [actual value] |
| Signature Algorithm | [✓/❌] | [actual value] |
| SAN Present | [✓/❌] | [yes/no] |
| CT SCTs | [✓/❌] | [count] |

# Certificate Data:
$cert_data"
            ;;
    esac

    echo "$prompt"
}

#############################################################################
# Call Claude API
#############################################################################

call_claude_api() {
    local prompt="$1"
    local model="${2:-$CLAUDE_MODEL}"
    local max_tokens="${3:-4096}"

    if [[ -z "$ANTHROPIC_API_KEY" ]]; then
        echo "Error: ANTHROPIC_API_KEY is required for Claude analysis" >&2
        return 1
    fi

    local response
    response=$(curl -s https://api.anthropic.com/v1/messages \
        -H "content-type: application/json" \
        -H "x-api-key: $ANTHROPIC_API_KEY" \
        -H "anthropic-version: 2023-06-01" \
        -d "{
            \"model\": \"$model\",
            \"max_tokens\": $max_tokens,
            \"messages\": [{
                \"role\": \"user\",
                \"content\": $(echo "$prompt" | jq -Rs .)
            }]
        }")

    # Check for errors
    local error=$(echo "$response" | jq -r '.error.message // empty')
    if [[ -n "$error" ]]; then
        echo "API Error: $error" >&2
        return 1
    fi

    # Record API usage if cost tracking is available
    if command -v record_api_usage &>/dev/null; then
        record_api_usage "$response" "$model" >/dev/null 2>&1 || true
    fi

    # Extract and return the response text
    echo "$response" | jq -r '.content[0].text // empty'
}

#############################################################################
# Main Analysis Function
#############################################################################

# Enhanced Claude analysis with RAG
# Usage: analyze_certificate_with_claude <cert_data> <repo_root> [analysis_type]
analyze_certificate_with_claude() {
    local cert_data="$1"
    local repo_root="$2"
    local analysis_type="${3:-comprehensive}"

    # Load RAG knowledge base
    echo -e "${BLUE:-}Loading certificate security knowledge base...${NC:-}" >&2
    local rag_context
    rag_context=$(load_certificate_rag "$repo_root")

    # Build the analysis prompt
    local prompt
    prompt=$(build_certificate_analysis_prompt "$cert_data" "$rag_context" "$analysis_type")

    # Call Claude API
    echo -e "${BLUE:-}Analyzing with Claude AI ($CLAUDE_MODEL)...${NC:-}" >&2
    local result
    result=$(call_claude_api "$prompt")

    if [[ -z "$result" ]]; then
        echo "Error: No response from Claude API" >&2
        return 1
    fi

    echo "$result"
}

#############################################################################
# Batch Analysis Functions
#############################################################################

# Analyze multiple certificates and provide comparison
analyze_certificates_batch() {
    local certs_dir="$1"
    local repo_root="$2"

    local batch_data=""
    local cert_count=0

    # Collect certificate data
    for cert_file in "$certs_dir"/cert*.pem; do
        if [[ -f "$cert_file" ]]; then
            cert_count=$((cert_count + 1))
            batch_data+="## Certificate $cert_count\n"
            batch_data+="$(openssl x509 -in "$cert_file" -noout -text 2>/dev/null)\n\n"
        fi
    done

    if [[ $cert_count -eq 0 ]]; then
        echo "No certificates found for batch analysis" >&2
        return 1
    fi

    # Load RAG and build prompt
    local rag_context
    rag_context=$(load_certificate_rag "$repo_root")

    local prompt="You are a certificate security expert. Analyze this certificate chain and provide insights.

# Certificate Knowledge Base
$rag_context

# Chain Analysis Requirements
1. Validate chain structure and order
2. Identify any weak links in the chain
3. Check intermediate certificate compliance
4. Assess overall chain security
5. Provide chain-specific recommendations

# Certificate Chain Data ($cert_count certificates):
$batch_data"

    call_claude_api "$prompt"
}

#############################################################################
# Specialized Analysis Functions
#############################################################################

# Quick risk assessment
quick_risk_assessment() {
    local cert_data="$1"
    local repo_root="$2"

    analyze_certificate_with_claude "$cert_data" "$repo_root" "quick"
}

# Compliance-focused analysis
compliance_analysis() {
    local cert_data="$1"
    local repo_root="$2"

    analyze_certificate_with_claude "$cert_data" "$repo_root" "compliance"
}

# Generate executive summary
generate_executive_summary() {
    local cert_data="$1"
    local repo_root="$2"

    local rag_context
    rag_context=$(load_certificate_rag "$repo_root")

    local prompt="You are a security consultant preparing a brief for executives. Summarize this certificate analysis in 3-4 paragraphs suitable for non-technical leadership.

Focus on:
1. Overall security status (good/needs attention/critical)
2. Business impact of any issues
3. Key actions required and timeline
4. Cost/risk of inaction

Keep technical jargon to a minimum. Be concise.

# Certificate Data:
$cert_data"

    call_claude_api "$prompt" "$CLAUDE_MODEL" 1024
}

#############################################################################
# Report Enhancement Functions
#############################################################################

# Append Claude analysis to markdown report
append_claude_analysis_to_report() {
    local report_file="$1"
    local claude_analysis="$2"

    cat >> "$report_file" <<EOF

---

## Claude AI Security Analysis

$claude_analysis

---

*Analysis generated by Claude AI using certificate security knowledge base.*
*Model: $CLAUDE_MODEL*
*Generated: $(date -u +"%Y-%m-%d %H:%M:%S UTC")*

EOF
}

# Print Claude analysis to console with formatting
print_claude_analysis() {
    local analysis="$1"

    echo ""
    echo "═══════════════════════════════════════════════════════════"
    echo "  Claude AI Enhanced Security Analysis"
    echo "═══════════════════════════════════════════════════════════"
    echo ""
    echo "$analysis"
    echo ""
    echo "═══════════════════════════════════════════════════════════"
    echo ""
}
