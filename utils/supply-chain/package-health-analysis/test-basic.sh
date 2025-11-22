#!/bin/bash
# Quick test of package health analyser

set -e

echo "Testing Package Health Analyser..."
echo "=================================="
echo ""

# Test 1: Check if scripts are executable
echo "Test 1: Checking script permissions..."
if [ -x "package-health-analyser.sh" ]; then
    echo "✓ Base analyser is executable"
else
    echo "✗ Base analyser is not executable"
    exit 1
fi

if [ -x "package-health-analyser-claude.sh" ]; then
    echo "✓ AI analyser is executable"
else
    echo "✗ AI analyser is not executable"
    exit 1
fi

# Test 2: Check dependencies
echo ""
echo "Test 2: Checking dependencies..."
if command -v jq &> /dev/null; then
    echo "✓ jq is installed"
else
    echo "✗ jq is missing"
    exit 1
fi

if command -v curl &> /dev/null; then
    echo "✓ curl is installed"
else
    echo "✗ curl is missing"
    exit 1
fi

if command -v syft &> /dev/null; then
    echo "✓ syft is installed"
else
    echo "⚠ syft is missing (required for repo scanning)"
fi

if command -v gh &> /dev/null; then
    echo "✓ gh is installed"
else
    echo "⚠ gh is missing (required for repo scanning)"
fi

# Test 3: Test deps.dev API directly
echo ""
echo "Test 3: Testing deps.dev API connection..."
response=$(curl -s "https://api.deps.dev/v3alpha/systems/npm/packages/request")
if echo "$response" | jq -e '.packageKey.name' > /dev/null 2>&1; then
    deprecated=$(echo "$response" | jq -r '.deprecated // false')
    echo "✓ deps.dev API is accessible"
    echo "  Package 'request' deprecated: $deprecated"
else
    echo "✗ deps.dev API is not accessible"
    exit 1
fi

# Test 4: Create a minimal test SBOM
echo ""
echo "Test 4: Creating test SBOM..."
cat > test-sbom.json <<'EOF'
{
  "bomFormat": "CycloneDX",
  "specVersion": "1.4",
  "version": 1,
  "components": [
    {
      "type": "library",
      "name": "request",
      "version": "2.88.0",
      "purl": "pkg:npm/request@2.88.0"
    },
    {
      "type": "library",
      "name": "axios",
      "version": "1.6.0",
      "purl": "pkg:npm/axios@1.6.0"
    },
    {
      "type": "library",
      "name": "lodash",
      "version": "4.17.21",
      "purl": "pkg:npm/lodash@4.17.21"
    }
  ]
}
EOF
echo "✓ Test SBOM created"

# Test 5: Run base analyser on test SBOM
echo ""
echo "Test 5: Running base analyser on test SBOM..."
if ./package-health-analyser.sh --sbom test-sbom.json --format json > test-results.json 2>&1; then
    echo "✓ Base analyser executed successfully"

    # Check results
    total_packages=$(jq -r '.summary.total_packages' test-results.json 2>/dev/null || echo "0")
    deprecated=$(jq -r '.summary.deprecated_packages' test-results.json 2>/dev/null || echo "0")

    echo "  Packages analyzed: $total_packages"
    echo "  Deprecated packages: $deprecated"

    if [ "$total_packages" -gt 0 ]; then
        echo "✓ Analysis produced results"
    else
        echo "⚠ No packages were analyzed"
    fi
else
    echo "✗ Base analyser failed"
    cat test-results.json
    exit 1
fi

# Test 6: Check for API key (for AI analyser)
echo ""
echo "Test 6: Checking for AI analyser requirements..."
if [ -n "${ANTHROPIC_API_KEY:-}" ]; then
    echo "✓ ANTHROPIC_API_KEY is set"
    echo "  AI-enhanced analyser can be tested"
else
    echo "⚠ ANTHROPIC_API_KEY not set"
    echo "  AI-enhanced analyser will not be tested"
fi

# Cleanup
echo ""
echo "Cleanup..."
rm -f test-sbom.json test-results.json
echo "✓ Test files cleaned up"

echo ""
echo "=================================="
echo "Basic tests completed successfully!"
echo ""
echo "Next steps:"
echo "1. Test with real repository (requires gh CLI)"
echo "2. Test AI-enhanced analyser (requires ANTHROPIC_API_KEY)"
echo "3. Test organization-wide scanning"
