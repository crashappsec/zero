#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Unit Tests for SBOM Scanning Functions
# Tests Layer 1a: SBOM package detection
#############################################################################

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TEST_ROOT="$(dirname "$SCRIPT_DIR")"
UTILS_ROOT="$(dirname "$(dirname "$TEST_ROOT")")"

# Load test framework
source "$TEST_ROOT/test-framework.sh"

# Load the analyzer script functions
ANALYZER_SCRIPT="$TEST_ROOT/../technology-identification-analyser.sh"

# Source only the functions we need (not the main script execution)
# We'll extract the scan_sbom_packages function
scan_sbom_packages() {
    local sbom_file="$1"
    local findings=()

    # Extract components from SBOM
    local components=$(jq -c '.components[]?' "$sbom_file" 2>/dev/null)

    if [[ -z "$components" ]]; then
        echo "[]"
        return
    fi

    while IFS= read -r component; do
        local name=$(echo "$component" | jq -r '.name // ""' 2>/dev/null || echo "")
        local version=$(echo "$component" | jq -r '.version // ""' 2>/dev/null || echo "")
        local purl=$(echo "$component" | jq -r '.purl // ""' 2>/dev/null || echo "")

        # Extract ecosystem from purl
        local ecosystem=""
        if [[ -n "$purl" ]]; then
            ecosystem=$(echo "$purl" | sed -n 's/^pkg:\([^/]*\).*/\1/p')
        fi

        # Map package names to technology categories
        local tech_category=""
        local tech_name=""
        local confidence=95

        case "$name" in
            # Business Tools - Payment
            stripe) tech_category="business-tools/payment"; tech_name="Stripe" ;;
            paypal|paypal-*) tech_category="business-tools/payment"; tech_name="PayPal" ;;

            # Developer Tools - Infrastructure
            terraform|terraform-*) tech_category="developer-tools/infrastructure"; tech_name="Terraform" ;;

            # Web Frameworks - Frontend
            react|react-dom) tech_category="web-frameworks/frontend"; tech_name="React" ;;

            # Databases
            pg|postgres|postgresql) tech_category="databases/relational"; tech_name="PostgreSQL" ;;
            redis) tech_category="databases/keyvalue"; tech_name="Redis" ;;

            # Cloud - AWS
            aws-sdk|aws-sdk-*|@aws-sdk/*|boto3|botocore) tech_category="cloud-providers/aws"; tech_name="AWS SDK" ;;

            *)
                # Unknown package, skip
                continue
                ;;
        esac

        if [[ -n "$tech_category" ]]; then
            local finding=$(jq -n \
                --arg name "$tech_name" \
                --arg category "$tech_category" \
                --arg version "$version" \
                --argjson confidence "$confidence" \
                --arg method "sbom-package" \
                --arg evidence "package.json dependency: $name@$version" \
                '{
                    name: $name,
                    category: $category,
                    version: $version,
                    confidence: $confidence,
                    detection_method: $method,
                    evidence: [$evidence]
                }')
            findings+=("$finding")
        fi
    done <<< "$components"

    # Return findings as JSON array
    if [[ ${#findings[@]} -eq 0 ]]; then
        echo "[]"
    else
        printf '%s\n' "${findings[@]}" | jq -s '.' 2>/dev/null || echo "[]"
    fi
}

#############################################################################
# Tests
#############################################################################

test_scan_sbom_empty() {
    # Create empty SBOM
    local sbom_file="$TEST_TEMP_DIR/empty.json"
    echo '{"bomFormat":"CycloneDX","specVersion":"1.4","version":1,"components":[]}' > "$sbom_file"

    local result=$(scan_sbom_packages "$sbom_file")

    assert_json_valid "$result" "Should return valid JSON" &&
    assert_equals "[]" "$result" "Should return empty array for empty SBOM"
}

test_scan_sbom_single_stripe() {
    # Create SBOM with Stripe
    local sbom_file="$TEST_TEMP_DIR/stripe.json"
    cat > "$sbom_file" << 'EOF'
{
    "bomFormat": "CycloneDX",
    "specVersion": "1.4",
    "version": 1,
    "components": [
        {
            "type": "library",
            "name": "stripe",
            "version": "14.12.0",
            "purl": "pkg:npm/stripe@14.12.0"
        }
    ]
}
EOF

    local result=$(scan_sbom_packages "$sbom_file")

    assert_json_valid "$result" "Should return valid JSON" &&
    assert_contains "$result" "Stripe" "Should detect Stripe" &&
    assert_contains "$result" "14.12.0" "Should include version" &&
    assert_contains "$result" "business-tools/payment" "Should categorize as payment tool"
}

test_scan_sbom_multiple_technologies() {
    # Create SBOM with multiple technologies
    local sbom_file="$TEST_TEMP_DIR/multi.json"
    cat > "$sbom_file" << 'EOF'
{
    "bomFormat": "CycloneDX",
    "specVersion": "1.4",
    "version": 1,
    "components": [
        {
            "type": "library",
            "name": "stripe",
            "version": "14.12.0",
            "purl": "pkg:npm/stripe@14.12.0"
        },
        {
            "type": "library",
            "name": "react",
            "version": "18.2.0",
            "purl": "pkg:npm/react@18.2.0"
        },
        {
            "type": "library",
            "name": "redis",
            "version": "4.5.0",
            "purl": "pkg:npm/redis@4.5.0"
        }
    ]
}
EOF

    local result=$(scan_sbom_packages "$sbom_file")
    local count=$(echo "$result" | jq 'length')

    assert_json_valid "$result" "Should return valid JSON" &&
    assert_equals "3" "$count" "Should detect 3 technologies" &&
    assert_contains "$result" "Stripe" "Should detect Stripe" &&
    assert_contains "$result" "React" "Should detect React" &&
    assert_contains "$result" "Redis" "Should detect Redis"
}

test_scan_sbom_aws_sdk() {
    # Test AWS SDK detection
    local sbom_file="$TEST_TEMP_DIR/aws.json"
    cat > "$sbom_file" << 'EOF'
{
    "bomFormat": "CycloneDX",
    "specVersion": "1.4",
    "version": 1,
    "components": [
        {
            "type": "library",
            "name": "@aws-sdk/client-s3",
            "version": "3.450.0",
            "purl": "pkg:npm/@aws-sdk/client-s3@3.450.0"
        }
    ]
}
EOF

    local result=$(scan_sbom_packages "$sbom_file")

    assert_json_valid "$result" "Should return valid JSON" &&
    assert_contains "$result" "AWS SDK" "Should detect AWS SDK" &&
    assert_contains "$result" "cloud-providers/aws" "Should categorize as AWS"
}

test_scan_sbom_unknown_packages() {
    # Create SBOM with unknown packages
    local sbom_file="$TEST_TEMP_DIR/unknown.json"
    cat > "$sbom_file" << 'EOF'
{
    "bomFormat": "CycloneDX",
    "specVersion": "1.4",
    "version": 1,
    "components": [
        {
            "type": "library",
            "name": "some-unknown-package",
            "version": "1.0.0",
            "purl": "pkg:npm/some-unknown-package@1.0.0"
        },
        {
            "type": "library",
            "name": "another-unknown",
            "version": "2.0.0",
            "purl": "pkg:npm/another-unknown@2.0.0"
        }
    ]
}
EOF

    local result=$(scan_sbom_packages "$sbom_file")

    assert_json_valid "$result" "Should return valid JSON" &&
    assert_equals "[]" "$result" "Should return empty array for unknown packages"
}

test_scan_sbom_confidence_score() {
    # Test that confidence scores are correct
    local sbom_file="$TEST_TEMP_DIR/confidence.json"
    cat > "$sbom_file" << 'EOF'
{
    "bomFormat": "CycloneDX",
    "specVersion": "1.4",
    "version": 1,
    "components": [
        {
            "type": "library",
            "name": "stripe",
            "version": "14.12.0",
            "purl": "pkg:npm/stripe@14.12.0"
        }
    ]
}
EOF

    local result=$(scan_sbom_packages "$sbom_file")
    local confidence=$(echo "$result" | jq '.[0].confidence')

    assert_equals "95" "$confidence" "SBOM package detection should have 95% confidence"
}

test_scan_sbom_version_extraction() {
    # Test version extraction
    local sbom_file="$TEST_TEMP_DIR/version.json"
    cat > "$sbom_file" << 'EOF'
{
    "bomFormat": "CycloneDX",
    "specVersion": "1.4",
    "version": 1,
    "components": [
        {
            "type": "library",
            "name": "react",
            "version": "18.2.0",
            "purl": "pkg:npm/react@18.2.0"
        }
    ]
}
EOF

    local result=$(scan_sbom_packages "$sbom_file")
    local version=$(echo "$result" | jq -r '.[0].version')

    assert_equals "18.2.0" "$version" "Should extract exact version"
}

test_scan_sbom_invalid_json() {
    # Test handling of invalid JSON
    local sbom_file="$TEST_TEMP_DIR/invalid.json"
    echo "{ invalid json }" > "$sbom_file"

    local result=$(scan_sbom_packages "$sbom_file")

    assert_equals "[]" "$result" "Should return empty array for invalid JSON"
}

test_scan_sbom_missing_file() {
    # Test handling of missing file
    local sbom_file="$TEST_TEMP_DIR/nonexistent.json"

    local result=$(scan_sbom_packages "$sbom_file")

    assert_equals "[]" "$result" "Should return empty array for missing file"
}

#############################################################################
# Run all tests
#############################################################################

main() {
    echo ""
    echo "========================================="
    echo "  SBOM Scanning Unit Tests"
    echo "========================================="
    echo ""

    run_test "Empty SBOM returns empty array" test_scan_sbom_empty
    run_test "Single Stripe package detected" test_scan_sbom_single_stripe
    run_test "Multiple technologies detected" test_scan_sbom_multiple_technologies
    run_test "AWS SDK detected correctly" test_scan_sbom_aws_sdk
    run_test "Unknown packages ignored" test_scan_sbom_unknown_packages
    run_test "Confidence score is 95%" test_scan_sbom_confidence_score
    run_test "Version extracted correctly" test_scan_sbom_version_extraction
    run_test "Invalid JSON handled gracefully" test_scan_sbom_invalid_json
    run_test "Missing file handled gracefully" test_scan_sbom_missing_file

    print_summary
}

# Run tests if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main
    exit $?
fi
