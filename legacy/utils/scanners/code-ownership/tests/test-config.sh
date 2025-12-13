#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# 
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Unit Tests for Configuration Library
# Tests configuration loading, validation, and management
#############################################################################

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LIB_DIR="$SCRIPT_DIR/../lib"

# Load library
source "$LIB_DIR/config.sh"

# Test framework
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

assert_equals() {
    local expected="$1"
    local actual="$2"
    local test_name="$3"

    ((TESTS_RUN++))

    if [[ "$expected" == "$actual" ]]; then
        echo "✓ PASS: $test_name"
        ((TESTS_PASSED++))
        return 0
    else
        echo "✗ FAIL: $test_name"
        echo "  Expected: $expected"
        echo "  Actual:   $actual"
        ((TESTS_FAILED++))
        return 1
    fi
}

assert_file_exists() {
    local file="$1"
    local test_name="$2"

    ((TESTS_RUN++))

    if [[ -f "$file" ]]; then
        echo "✓ PASS: $test_name"
        ((TESTS_PASSED++))
        return 0
    else
        echo "✗ FAIL: $test_name"
        echo "  File not found: $file"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Test: init_config
test_init_config() {
    echo ""
    echo "Testing init_config..."

    init_config

    # Check that defaults are loaded
    local analysis_method=$(get_config "analysis_method")
    assert_equals "hybrid" "$analysis_method" "Default analysis_method should be 'hybrid'"

    local analysis_days=$(get_config "analysis_days")
    assert_equals "90" "$analysis_days" "Default analysis_days should be '90'"

    local output_format=$(get_config "output_format")
    assert_equals "json" "$output_format" "Default output_format should be 'json'"
}

# Test: get_config and set_config
test_get_set_config() {
    echo ""
    echo "Testing get_config and set_config..."

    init_config

    # Test get_config
    local value=$(get_config "analysis_method")
    assert_equals "hybrid" "$value" "get_config should return default value"

    # Test set_config
    set_config "analysis_method" "commit"
    value=$(get_config "analysis_method")
    assert_equals "commit" "$value" "set_config should update value"

    # Test get_config with default
    value=$(get_config "nonexistent_key" "default_value")
    assert_equals "default_value" "$value" "get_config should return default for missing key"
}

# Test: load_config_file
test_load_config_file() {
    echo ""
    echo "Testing load_config_file..."

    init_config

    # Create test config file
    local test_config=$(mktemp)
    cat > "$test_config" << EOF
# Test configuration
analysis_method=line
analysis_days=180
coverage_target=85
EOF

    # Load config file
    load_config_file "$test_config"

    # Verify values loaded
    local method=$(get_config "analysis_method")
    assert_equals "line" "$method" "Config file should override analysis_method"

    local days=$(get_config "analysis_days")
    assert_equals "180" "$days" "Config file should override analysis_days"

    local target=$(get_config "coverage_target")
    assert_equals "85" "$target" "Config file should set coverage_target"

    # Cleanup
    rm -f "$test_config"
}

# Test: load_env_config
test_load_env_config() {
    echo ""
    echo "Testing load_env_config..."

    init_config

    # Set environment variables
    export CODE_OWNERSHIP_ANALYSIS_METHOD="commit"
    export CODE_OWNERSHIP_ANALYSIS_DAYS="120"
    export CODE_OWNERSHIP_BUS_FACTOR_THRESHOLD="5"

    # Load environment config
    load_env_config

    # Verify values loaded
    local method=$(get_config "analysis_method")
    assert_equals "commit" "$method" "Environment should override analysis_method"

    local days=$(get_config "analysis_days")
    assert_equals "120" "$days" "Environment should override analysis_days"

    local threshold=$(get_config "bus_factor_threshold")
    assert_equals "5" "$threshold" "Environment should override bus_factor_threshold"

    # Cleanup
    unset CODE_OWNERSHIP_ANALYSIS_METHOD
    unset CODE_OWNERSHIP_ANALYSIS_DAYS
    unset CODE_OWNERSHIP_BUS_FACTOR_THRESHOLD
}

# Test: validate_config
test_validate_config() {
    echo ""
    echo "Testing validate_config..."

    # Test valid configuration
    init_config
    set_config "analysis_method" "hybrid"
    set_config "analysis_days" "90"
    set_config "output_format" "json"

    if validate_config 2>/dev/null; then
        echo "✓ PASS: Valid configuration should pass validation"
        ((TESTS_RUN++))
        ((TESTS_PASSED++))
    else
        echo "✗ FAIL: Valid configuration should pass validation"
        ((TESTS_RUN++))
        ((TESTS_FAILED++))
    fi

    # Test invalid analysis_method
    set_config "analysis_method" "invalid"
    if ! validate_config 2>/dev/null; then
        echo "✓ PASS: Invalid analysis_method should fail validation"
        ((TESTS_RUN++))
        ((TESTS_PASSED++))
    else
        echo "✗ FAIL: Invalid analysis_method should fail validation"
        ((TESTS_RUN++))
        ((TESTS_FAILED++))
    fi

    # Reset to valid
    set_config "analysis_method" "hybrid"

    # Test invalid output_format
    set_config "output_format" "xml"
    if ! validate_config 2>/dev/null; then
        echo "✓ PASS: Invalid output_format should fail validation"
        ((TESTS_RUN++))
        ((TESTS_PASSED++))
    else
        echo "✗ FAIL: Invalid output_format should fail validation"
        ((TESTS_RUN++))
        ((TESTS_FAILED++))
    fi
}

# Test: generate_default_config
test_generate_default_config() {
    echo ""
    echo "Testing generate_default_config..."

    local test_config=$(mktemp)

    generate_default_config "$test_config"

    assert_file_exists "$test_config" "Config file should be generated"

    # Check contents
    if grep -q "analysis_method=" "$test_config"; then
        echo "✓ PASS: Config file should contain analysis_method"
        ((TESTS_RUN++))
        ((TESTS_PASSED++))
    else
        echo "✗ FAIL: Config file should contain analysis_method"
        ((TESTS_RUN++))
        ((TESTS_FAILED++))
    fi

    # Cleanup
    rm -f "$test_config"
}

# Test: get_config_int, get_config_float, get_config_bool
test_type_conversions() {
    echo ""
    echo "Testing type conversion functions..."

    init_config

    # Test get_config_int
    set_config "test_int" "42"
    local int_val=$(get_config_int "test_int")
    assert_equals "42" "$int_val" "get_config_int should return integer"

    # Test get_config_float
    set_config "test_float" "3.14159"
    local float_val=$(get_config_float "test_float")
    assert_equals "3.14159" "$float_val" "get_config_float should return float"

    # Test get_config_bool (true values)
    for value in "true" "yes" "1" "on"; do
        set_config "test_bool" "$value"
        local bool_val=$(get_config_bool "test_bool")
        assert_equals "true" "$bool_val" "get_config_bool should return true for '$value'"
    done

    # Test get_config_bool (false values)
    for value in "false" "no" "0" "off"; do
        set_config "test_bool" "$value"
        local bool_val=$(get_config_bool "test_bool")
        assert_equals "false" "$bool_val" "get_config_bool should return false for '$value'"
    done
}

# Run all tests
main() {
    echo "========================================="
    echo "Configuration Library Unit Tests"
    echo "========================================="

    test_init_config
    test_get_set_config
    test_load_config_file
    test_load_env_config
    test_validate_config
    test_generate_default_config
    test_type_conversions

    echo ""
    echo "========================================="
    echo "Test Results:"
    echo "  Total:  $TESTS_RUN"
    echo "  Passed: $TESTS_PASSED"
    echo "  Failed: $TESTS_FAILED"
    echo "========================================="

    if [[ $TESTS_FAILED -eq 0 ]]; then
        echo "✓ All tests passed!"
        exit 0
    else
        echo "✗ Some tests failed"
        exit 1
    fi
}

main "$@"
