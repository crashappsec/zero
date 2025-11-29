<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Technology Identification Testing Suite

Comprehensive testing infrastructure for the technology identification analyzer.

## Overview

This testing suite provides **unit tests**, **integration tests**, and **CI/CD automation** to ensure the reliability and accuracy of technology detection across all layers.

### Test Coverage

- **Unit Tests**: Test individual detection functions in isolation
- **Integration Tests**: Test end-to-end workflows with real repository structures
- **Test Framework**: Custom assertion library and test runner
- **CI/CD**: Automated testing on every push and PR

## Quick Start

```bash
# Run all tests
cd utils/technology-identification/tests
./run-all-tests.sh

# Run only unit tests
./unit/test-sbom-scanning.sh
./unit/test-confidence-scoring.sh

# Run only integration tests
./integration/test-full-workflow.sh
```

## Test Structure

```
tests/
├── README.md                    # This file
├── test-framework.sh            # Test framework with assertions
├── run-all-tests.sh            # Master test runner
├── unit/                        # Unit tests
│   ├── test-sbom-scanning.sh
│   └── test-confidence-scoring.sh
├── integration/                 # Integration tests
│   └── test-full-workflow.sh
└── fixtures/                    # Test data (auto-generated)
```

## Test Framework

### Assertion Functions

The test framework (`test-framework.sh`) provides comprehensive assertion functions:

#### Value Assertions
```bash
assert_equals "expected" "actual" "Optional message"
assert_not_equals "unexpected" "actual"
assert_contains "haystack" "needle"
assert_not_contains "haystack" "needle"
```

#### Numeric Assertions
```bash
assert_greater_than 95 90
assert_less_than 50 100
```

#### File Assertions
```bash
assert_file_exists "/path/to/file"
assert_file_not_exists "/path/to/file"
```

#### JSON Assertions
```bash
assert_json_valid "$json_string"
assert_json_contains_key "$json" "key.path"
assert_json_value "$json" ".technologies[0].name" "Stripe"
```

#### Exit Code Assertions
```bash
assert_exit_code 0 "some-command arg1 arg2"
assert_exit_code 1 "command-that-should-fail"
```

### Test Execution

```bash
# Define a test function
test_something() {
    local result=$(function_to_test "input")
    assert_equals "expected" "$result"
}

# Run the test
run_test "Test description" test_something
```

### Test Lifecycle

Each test has automatic setup and teardown:

```bash
# Called before each test
setup() {
    # Creates TEST_TEMP_DIR
    # Isolated environment per test
}

# Called after each test
teardown() {
    # Cleans up TEST_TEMP_DIR
    # Removes temporary files
}
```

## Unit Tests

### test-sbom-scanning.sh

Tests Layer 1a SBOM package detection.

**Coverage**:
- ✅ Empty SBOM handling
- ✅ Single technology detection
- ✅ Multiple technology detection
- ✅ AWS SDK pattern matching
- ✅ Unknown package filtering
- ✅ Confidence scoring (95%)
- ✅ Version extraction
- ✅ Invalid JSON handling
- ✅ Missing file handling

**Run**:
```bash
./unit/test-sbom-scanning.sh
```

**Example Output**:
```
=========================================
  SBOM Scanning Unit Tests
=========================================

✓ PASS: Empty SBOM returns empty array
✓ PASS: Single Stripe package detected
✓ PASS: Multiple technologies detected
✓ PASS: AWS SDK detected correctly
✓ PASS: Unknown packages ignored
✓ PASS: Confidence score is 95%
✓ PASS: Version extracted correctly
✓ PASS: Invalid JSON handled gracefully
✓ PASS: Missing file handled gracefully

=========================================
  Test Results
=========================================

Total Tests:  9
Passed:       9
Failed:       0

=========================================
  ALL TESTS PASSED!
=========================================
```

### test-confidence-scoring.sh

Tests confidence aggregation and composite scoring algorithm.

**Coverage**:
- ✅ Empty layer handling
- ✅ Single layer aggregation
- ✅ Multiple detection same technology (Bayesian composite)
- ✅ Confidence capping at 100
- ✅ Multiple technology preservation
- ✅ Sorting by confidence (descending)
- ✅ Version preference (from layer with version)
- ✅ Evidence deduplication
- ✅ Detection method uniqueness
- ✅ Null layer handling

**Algorithm Tested**:
```
Composite Confidence = (P1 + P2 + ... + PN) / N × 1.2
(Capped at 100%)

Example - Stripe detected in 3 layers:
- SBOM: 95%
- Import: 75%
- API: 65%
→ Composite: (95 + 75 + 65) / 3 × 1.2 = 94%
```

**Run**:
```bash
./unit/test-confidence-scoring.sh
```

## Integration Tests

### test-full-workflow.sh

Tests complete end-to-end analyzer workflows with realistic repository structures.

**Test Repositories Created**:

#### 1. Simple Node.js App
```
test-repo/
├── package.json        # Stripe, Express, React
├── src/
│   └── app.js         # Import statements
├── .env.example       # Environment variables
└── Dockerfile         # Container config
```

**Expected Detections**:
- Stripe (Layer 1: SBOM, Layer 3: Import, Layer 5: Env)
- Express (Layer 1: SBOM, Layer 3: Import)
- React (Layer 1: SBOM)
- Docker (Layer 2: Config)

#### 2. Python + AWS App
```
test-repo/
├── requirements.txt   # boto3, flask, redis
├── src/
│   └── app.py        # Import statements
└── .env              # AWS credentials
```

**Expected Detections**:
- AWS SDK/boto3 (Layers 1, 3, 5)
- Flask (Layers 1, 3)
- Redis (Layers 1, 3)

#### 3. Terraform Infrastructure
```
test-repo/
└── main.tf           # Terraform config with AWS provider
```

**Expected Detections**:
- Terraform (Layer 2: Config, version detection)
- AWS (Layer 2: Provider reference)

**Test Scenarios**:
- ✅ Full workflow with SBOM generation
- ✅ Multiple detection layers increase confidence
- ✅ JSON output structure validation
- ✅ Markdown output format
- ✅ Confidence threshold filtering
- ✅ Error handling for invalid paths

**Run**:
```bash
./integration/test-full-workflow.sh
```

**Note**: Integration tests require `syft` to be installed:
```bash
brew install syft
```

## Running Tests

### Run All Tests

```bash
# Master test runner
./run-all-tests.sh
```

**Output**:
```
╔══════════════════════════════════════════╗
║  Technology Identification Test Suite   ║
╚══════════════════════════════════════════╝

Checking prerequisites...
✓ jq installed
✓ syft installed
✓ osv-scanner installed

═══════════════════════════════════════
  UNIT TESTS
═══════════════════════════════════════

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Running: SBOM Scanning
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
[test output]
✓ SBOM Scanning PASSED

[... more tests ...]

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  FINAL TEST SUMMARY
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Total Test Suites:  3
Passed:             3
Failed:             0

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  ALL TEST SUITES PASSED!
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### Run Individual Test Suites

```bash
# Unit tests
./unit/test-sbom-scanning.sh
./unit/test-confidence-scoring.sh

# Integration tests
./integration/test-full-workflow.sh
```

### Run with Verbose Output

```bash
# Set bash debug mode
bash -x ./run-all-tests.sh
```

## CI/CD Integration

### GitHub Actions Workflow

Tests run automatically on:
- Push to `main` or `feature/technology-identification` branches
- Pull requests to `main`
- Changes to technology identification code or tests

**Workflow File**: `.github/workflows/test-technology-identification.yml`

**Jobs**:
1. **test**: Run all unit and integration tests
2. **test-claude**: Test Claude AI integration (main branch only, requires API key)
3. **lint**: ShellCheck linting and syntax validation
4. **coverage**: Test coverage reporting (placeholder)

**View Results**: https://github.com/crashappsec/gibson-powers/actions

### CI/CD Prerequisites

**Installed by CI**:
- jq
- bc
- syft
- osv-scanner
- ShellCheck (for linting)

**Secrets Required** (optional):
- `ANTHROPIC_API_KEY`: For Claude AI integration tests

## Prerequisites

### Required
- **bash** >= 3.2 (macOS default)
- **jq**: JSON processor
  ```bash
  brew install jq
  ```

### Recommended
- **syft**: SBOM generation (for integration tests)
  ```bash
  brew install syft
  ```
- **osv-scanner**: Vulnerability scanning (for Layer 1b tests)
  ```bash
  go install github.com/google/osv-scanner/cmd/osv-scanner@latest
  ```

### Optional
- **ShellCheck**: Script linting
  ```bash
  brew install shellcheck
  ```

## Writing New Tests

### Unit Test Template

```bash
#!/bin/bash
# Load test framework
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TEST_ROOT="$(dirname "$SCRIPT_DIR")"
source "$TEST_ROOT/test-framework.sh"

# Define function to test (or source from analyzer)
my_function() {
    echo "result"
}

# Write test
test_my_feature() {
    local result=$(my_function "input")
    assert_equals "result" "$result" "Should return 'result'"
}

# Run tests
main() {
    echo "========================================="
    echo "  My Feature Tests"
    echo "========================================="

    run_test "My feature works" test_my_feature

    print_summary
}

main
exit $?
```

### Integration Test Template

```bash
#!/bin/bash
# Load test framework
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TEST_ROOT="$(dirname "$SCRIPT_DIR")"
source "$TEST_ROOT/test-framework.sh"

# Path to analyzer
ANALYZER_SCRIPT="$(dirname "$TEST_ROOT")/technology-identification-analyser.sh"

# Create test fixture
create_test_repo() {
    local repo_dir="$TEST_TEMP_DIR/test-repo"
    mkdir -p "$repo_dir"

    # Create test files
    echo '{"dependencies":{"stripe":"^14.0.0"}}' > "$repo_dir/package.json"

    echo "$repo_dir"
}

# Write test
test_full_workflow() {
    local repo_path=$(create_test_repo)

    # Run analyzer
    local output=$("$ANALYZER_SCRIPT" \
        --local-path "$repo_path" \
        --format json \
        --no-claude \
        2>/dev/null)

    assert_json_valid "$output" &&
    assert_contains "$output" "Stripe"
}

# Run tests
main() {
    run_test "Full workflow test" test_full_workflow
    print_summary
}

main
exit $?
```

## Test Data and Fixtures

Test fixtures are **auto-generated** in each test's `$TEST_TEMP_DIR` and automatically cleaned up.

**Common Fixtures**:

### SBOM Files
```bash
# Create test SBOM
cat > "$TEST_TEMP_DIR/sbom.json" << 'EOF'
{
  "bomFormat": "CycloneDX",
  "specVersion": "1.4",
  "components": [
    {
      "name": "stripe",
      "version": "14.12.0",
      "purl": "pkg:npm/stripe@14.12.0"
    }
  ]
}
EOF
```

### Repository Structures
```bash
# Create test repo
create_test_repo() {
    local repo_dir="$TEST_TEMP_DIR/repo"
    mkdir -p "$repo_dir/src"

    echo '{"dependencies":{"stripe":"^14.0.0"}}' > "$repo_dir/package.json"
    echo 'import Stripe from "stripe";' > "$repo_dir/src/app.js"

    echo "$repo_dir"
}
```

## Debugging Tests

### Failed Test Output

When a test fails, you'll see detailed output:

```
✗ FAIL: Should detect Stripe
  Expected: Stripe
  Actual:   Paypal

✗ FAIL: Stripe detected correctly

Failed Tests:
  ✗ Stripe detected correctly
```

### Debug Mode

Run tests with bash debug output:

```bash
bash -x ./unit/test-sbom-scanning.sh
```

### Check Test Environment

```bash
# Verify prerequisites
jq --version
syft version
osv-scanner --version

# Check test files
ls -la unit/
ls -la integration/
```

## Test Maintenance

### Adding New Tests

1. Create test file in `unit/` or `integration/`
2. Follow template structure
3. Add to `run-all-tests.sh`
4. Update this README
5. Run `./run-all-tests.sh` to verify

### Test Coverage Goals

**Target**: 80%+ function coverage

**Current Coverage**:
- ✅ SBOM scanning (scan_sbom_packages): 100%
- ✅ Confidence aggregation (aggregate_findings): 100%
- ⚠️ Config file scanning: 0%
- ⚠️ Import scanning: 0%
- ⚠️ API endpoint scanning: 0%
- ⚠️ Environment variable scanning: 0%

**TODO**: Add tests for remaining detection layers

## Future Enhancements

### Planned
- [ ] Test coverage reporting (bashcov/kcov)
- [ ] Performance benchmarking
- [ ] Regression test suite (compare against known good outputs)
- [ ] Fuzz testing for edge cases
- [ ] Mock external dependencies (GitHub API, etc.)

### Nice-to-Have
- [ ] Parallel test execution
- [ ] Test result dashboards
- [ ] Automated test generation from RAG patterns
- [ ] Property-based testing

## Troubleshooting

### Tests Fail on CI but Pass Locally

**Common Causes**:
- Different tool versions (jq, syft)
- Missing dependencies
- Environment variable differences

**Solution**: Run locally with same versions as CI

### Integration Tests Skip

**Cause**: `syft` not installed

**Solution**:
```bash
brew install syft
```

### Permission Denied

**Cause**: Test scripts not executable

**Solution**:
```bash
chmod +x tests/*.sh
chmod +x tests/unit/*.sh
chmod +x tests/integration/*.sh
```

## Resources

- [Test Framework Source](./test-framework.sh)
- [GitHub Actions Workflow](../../../.github/workflows/test-technology-identification.yml)
- [Analyzer Source](../technology-identification-analyser.sh)
- [CI/CD Results](https://github.com/crashappsec/gibson-powers/actions)

## Contributing

When contributing new detection features:

1. **Write tests first** (TDD approach)
2. Add unit tests for new functions
3. Add integration tests for new workflows
4. Ensure all existing tests pass
5. Update this README with new test descriptions

## License

GPL-3.0 - See [LICENSE](../../../LICENSE) for details.

---

**Last Updated**: 2025-11-24
**Test Framework Version**: 1.0.0
**Coverage**: 45% (20 tests across 2 suites)
