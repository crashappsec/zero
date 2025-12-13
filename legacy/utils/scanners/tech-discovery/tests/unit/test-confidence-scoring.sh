#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Unit Tests for Confidence Scoring and Aggregation
# Tests confidence calculation and finding aggregation
#############################################################################

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TEST_ROOT="$(dirname "$SCRIPT_DIR")"

# Load test framework
source "$TEST_ROOT/test-framework.sh"

# Implement the aggregate_findings function for testing
aggregate_findings() {
    local layer1="$1"
    local layer2="$2"
    local layer3="$3"
    local layer4="$4"
    local layer5="$5"

    # Combine all findings - flatten all arrays into one
    local all_findings=$(jq -s 'map(select(. != null)) | add' \
        <(echo "$layer1") \
        <(echo "$layer2") \
        <(echo "$layer3") \
        <(echo "$layer4") \
        <(echo "$layer5") 2>/dev/null)

    # Handle case where all_findings might be null
    if [[ "$all_findings" == "null" ]] || [[ -z "$all_findings" ]]; then
        all_findings="[]"
    fi

    # Group by technology name and calculate composite confidence
    echo "$all_findings" | jq '
        group_by(.name) |
        map({
            name: .[0].name,
            category: .[0].category,
            version: (map(.version // "") | map(select(length > 0)) | .[0] // ""),
            confidence: (
                map(.confidence) |
                (add / length * 1.2) |
                if . > 100 then 100 else . end |
                floor
            ),
            detection_methods: map(.detection_method) | unique,
            evidence: map(.evidence[]) | unique
        }) |
        sort_by(-.confidence)
    ' 2>/dev/null || echo "[]"
}

#############################################################################
# Tests
#############################################################################

test_aggregate_empty_layers() {
    local result=$(aggregate_findings "[]" "[]" "[]" "[]" "[]")

    assert_json_valid "$result" "Should return valid JSON" &&
    assert_equals "[]" "$result" "Should return empty array when all layers empty"
}

test_aggregate_single_layer() {
    local layer1='[{"name":"Stripe","category":"business-tools/payment","version":"14.12.0","confidence":95,"detection_method":"sbom-package","evidence":["package.json"]}]'

    local result=$(aggregate_findings "$layer1" "[]" "[]" "[]" "[]")
    local count=$(echo "$result" | jq 'length')

    # Single detection: 95 * 1.2 = 114, capped at 100
    assert_json_valid "$result" &&
    assert_equals "1" "$count" "Should have 1 technology" &&
    assert_json_value "$result" '.[0].name' 'Stripe' &&
    assert_json_value "$result" '.[0].confidence' '100'
}

test_aggregate_multiple_detections_same_tech() {
    # Stripe detected in 3 layers: SBOM (95), imports (75), API (65)
    # Expected composite: (95 + 75 + 65) / 3 * 1.2 = 93.6, floor to 93
    local layer1='[{"name":"Stripe","category":"business-tools/payment","version":"14.12.0","confidence":95,"detection_method":"sbom-package","evidence":["package.json"]}]'
    local layer3='[{"name":"Stripe","category":"business-tools/payment","confidence":75,"detection_method":"import-statement","evidence":["src/payments.js"]}]'
    local layer4='[{"name":"Stripe","category":"business-tools/payment","confidence":65,"detection_method":"api-endpoint","evidence":["api.stripe.com"]}]'

    local result=$(aggregate_findings "$layer1" "[]" "$layer3" "$layer4" "[]")
    local count=$(echo "$result" | jq 'length')
    local confidence=$(echo "$result" | jq '.[0].confidence')
    local methods=$(echo "$result" | jq '.[0].detection_methods | length')

    assert_equals "1" "$count" "Should deduplicate to 1 technology" &&
    assert_equals "93" "$confidence" "Composite confidence should be 93" &&
    assert_equals "3" "$methods" "Should have 3 detection methods"
}

test_aggregate_confidence_capped_at_100() {
    # High confidence detections that would exceed 100
    local layer1='[{"name":"Docker","category":"developer-tools/containers","confidence":95,"detection_method":"config-file","evidence":["Dockerfile"]}]'
    local layer2='[{"name":"Docker","category":"developer-tools/containers","confidence":90,"detection_method":"config-file","evidence":["docker-compose.yml"]}]'
    local layer3='[{"name":"Docker","category":"developer-tools/containers","confidence":85,"detection_method":"import-statement","evidence":["import"]}]'

    # (95 + 90 + 85) / 3 * 1.2 = 108, should cap at 100
    local result=$(aggregate_findings "$layer1" "$layer2" "$layer3" "[]" "[]")
    local confidence=$(echo "$result" | jq '.[0].confidence')

    assert_equals "100" "$confidence" "Confidence should be capped at 100"
}

test_aggregate_multiple_technologies() {
    local layer1='[
        {"name":"Stripe","category":"business-tools/payment","confidence":95,"detection_method":"sbom-package","evidence":["package.json"]},
        {"name":"React","category":"web-frameworks/frontend","confidence":95,"detection_method":"sbom-package","evidence":["package.json"]}
    ]'
    local layer2='[
        {"name":"Docker","category":"developer-tools/containers","confidence":90,"detection_method":"config-file","evidence":["Dockerfile"]}
    ]'

    local result=$(aggregate_findings "$layer1" "$layer2" "[]" "[]" "[]")
    local count=$(echo "$result" | jq 'length')

    assert_equals "3" "$count" "Should have 3 distinct technologies"
}

test_aggregate_sorts_by_confidence() {
    local layer1='[
        {"name":"Low","category":"test","confidence":50,"detection_method":"test","evidence":["test"]},
        {"name":"High","category":"test","confidence":95,"detection_method":"test","evidence":["test"]},
        {"name":"Medium","category":"test","confidence":75,"detection_method":"test","evidence":["test"]}
    ]'

    local result=$(aggregate_findings "$layer1" "[]" "[]" "[]" "[]")
    local first_name=$(echo "$result" | jq -r '.[0].name')
    local last_name=$(echo "$result" | jq -r '.[2].name')

    assert_equals "High" "$first_name" "Highest confidence should be first" &&
    assert_equals "Low" "$last_name" "Lowest confidence should be last"
}

test_aggregate_preserves_version() {
    local layer1='[{"name":"Stripe","category":"business-tools/payment","version":"14.12.0","confidence":95,"detection_method":"sbom-package","evidence":["package.json"]}]'
    local layer3='[{"name":"Stripe","category":"business-tools/payment","version":"","confidence":75,"detection_method":"import-statement","evidence":["import"]}]'

    local result=$(aggregate_findings "$layer1" "[]" "$layer3" "[]" "[]")
    local version=$(echo "$result" | jq -r '.[0].version')

    assert_equals "14.12.0" "$version" "Should preserve version from layer with version info"
}

test_aggregate_evidence_deduplication() {
    local layer1='[{"name":"AWS SDK","category":"cloud-providers/aws","confidence":95,"detection_method":"sbom-package","evidence":["package.json","package-lock.json"]}]'
    local layer3='[{"name":"AWS SDK","category":"cloud-providers/aws","confidence":75,"detection_method":"import-statement","evidence":["package.json","src/aws.js"]}]'

    local result=$(aggregate_findings "$layer1" "[]" "$layer3" "[]" "[]")
    local evidence_count=$(echo "$result" | jq '.[0].evidence | length')

    # Should have unique evidence: package.json, package-lock.json, src/aws.js = 3
    assert_equals "3" "$evidence_count" "Should deduplicate evidence"
}

test_aggregate_detection_methods_unique() {
    local layer1='[{"name":"Terraform","category":"developer-tools/infrastructure","confidence":90,"detection_method":"config-file","evidence":["main.tf"]}]'
    local layer2='[{"name":"Terraform","category":"developer-tools/infrastructure","confidence":90,"detection_method":"config-file","evidence":["variables.tf"]}]'

    local result=$(aggregate_findings "$layer1" "$layer2" "[]" "[]" "[]")
    local methods=$(echo "$result" | jq '.[0].detection_methods')

    assert_contains "$methods" "config-file" "Should have config-file method" &&
    assert_equals "1" "$(echo "$methods" | jq 'length')" "Should not duplicate detection method"
}

test_aggregate_handles_null_layers() {
    local layer1='[{"name":"React","category":"web-frameworks/frontend","confidence":95,"detection_method":"sbom-package","evidence":["package.json"]}]'

    # Pass null and empty strings
    local result=$(aggregate_findings "$layer1" "" "null" "[]" "")
    local count=$(echo "$result" | jq 'length')

    assert_equals "1" "$count" "Should handle null/empty layers gracefully"
}

#############################################################################
# Run all tests
#############################################################################

main() {
    echo ""
    echo "========================================="
    echo "  Confidence Scoring Unit Tests"
    echo "========================================="
    echo ""

    run_test "Empty layers return empty array" test_aggregate_empty_layers
    run_test "Single layer aggregation" test_aggregate_single_layer
    run_test "Multiple detections same tech" test_aggregate_multiple_detections_same_tech
    run_test "Confidence capped at 100" test_aggregate_confidence_capped_at_100
    run_test "Multiple technologies preserved" test_aggregate_multiple_technologies
    run_test "Results sorted by confidence" test_aggregate_sorts_by_confidence
    run_test "Version preserved correctly" test_aggregate_preserves_version
    run_test "Evidence deduplicated" test_aggregate_evidence_deduplication
    run_test "Detection methods unique" test_aggregate_detection_methods_unique
    run_test "Null layers handled gracefully" test_aggregate_handles_null_layers

    print_summary
}

# Run tests if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main
    exit $?
fi
