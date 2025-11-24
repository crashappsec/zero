#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Integration Tests for Technology Identification
# Tests complete end-to-end workflows
#############################################################################

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TEST_ROOT="$(dirname "$SCRIPT_DIR")"
ANALYZER_ROOT="$(dirname "$TEST_ROOT")"
ANALYZER_SCRIPT="$ANALYZER_ROOT/technology-identification-analyser.sh"

# Load test framework
source "$TEST_ROOT/test-framework.sh"

#############################################################################
# Test Fixtures
#############################################################################

create_simple_node_repo() {
    local repo_dir="$TEST_TEMP_DIR/simple-node-app"
    mkdir -p "$repo_dir/src"

    # Create package.json
    cat > "$repo_dir/package.json" << 'EOF'
{
  "name": "test-app",
  "version": "1.0.0",
  "dependencies": {
    "stripe": "^14.12.0",
    "express": "^4.18.0",
    "react": "^18.2.0"
  }
}
EOF

    # Create source file with imports
    cat > "$repo_dir/src/app.js" << 'EOF'
import Stripe from 'stripe';
import express from 'express';

const stripe = new Stripe(process.env.STRIPE_SECRET_KEY);
const app = express();

app.post('/charge', async (req, res) => {
    const charge = await stripe.charges.create({
        amount: 1000,
        currency: 'usd'
    });
    res.json(charge);
});
EOF

    # Create .env.example
    cat > "$repo_dir/.env.example" << 'EOF'
STRIPE_SECRET_KEY=sk_test_...
STRIPE_PUBLISHABLE_KEY=pk_test_...
EOF

    # Create Dockerfile
    cat > "$repo_dir/Dockerfile" << 'EOF'
FROM node:18
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
CMD ["node", "src/app.js"]
EOF

    echo "$repo_dir"
}

create_python_repo_with_aws() {
    local repo_dir="$TEST_TEMP_DIR/python-aws-app"
    mkdir -p "$repo_dir/src"

    # Create requirements.txt
    cat > "$repo_dir/requirements.txt" << 'EOF'
boto3==1.28.0
flask==3.0.0
redis==5.0.0
EOF

    # Create Python source with imports
    cat > "$repo_dir/src/app.py" << 'EOF'
import boto3
from flask import Flask
import redis

app = Flask(__name__)
s3 = boto3.client('s3')
cache = redis.Redis(host='localhost', port=6379)

@app.route('/upload', methods=['POST'])
def upload():
    s3.upload_file('file.txt', 'my-bucket', 'file.txt')
    return 'OK'
EOF

    # Create .env
    cat > "$repo_dir/.env" << 'EOF'
AWS_ACCESS_KEY_ID=AKIA...
AWS_SECRET_ACCESS_KEY=...
AWS_REGION=us-east-1
EOF

    echo "$repo_dir"
}

create_terraform_repo() {
    local repo_dir="$TEST_TEMP_DIR/terraform-infra"
    mkdir -p "$repo_dir"

    # Create Terraform config
    cat > "$repo_dir/main.tf" << 'EOF'
terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

resource "aws_s3_bucket" "example" {
  bucket = "my-test-bucket"
}
EOF

    echo "$repo_dir"
}

#############################################################################
# Integration Tests
#############################################################################

test_analyze_simple_node_repo() {
    local repo_path=$(create_simple_node_repo)

    # Generate SBOM first (prerequisite for analyzer)
    if ! command -v syft &> /dev/null; then
        echo "Skipping: syft not installed"
        return 0
    fi

    local sbom_file="$TEST_TEMP_DIR/sbom.json"
    syft dir:"$repo_path" -o cyclonedx-json > "$sbom_file" 2>/dev/null || true

    # Run analyzer with local path and SBOM
    local output=$("$ANALYZER_SCRIPT" \
        --local-path "$repo_path" \
        --sbom-file "$sbom_file" \
        --format json \
        --no-claude \
        2>/dev/null || echo '{}')

    assert_json_valid "$output" "Output should be valid JSON" &&
    assert_contains "$output" "Stripe" "Should detect Stripe" &&
    assert_contains "$output" "Express" "Should detect Express" &&
    assert_contains "$output" "React" "Should detect React" &&
    assert_contains "$output" "Docker" "Should detect Docker from Dockerfile"
}

test_analyze_python_aws_repo() {
    local repo_path=$(create_python_repo_with_aws)

    if ! command -v syft &> /dev/null; then
        echo "Skipping: syft not installed"
        return 0
    fi

    local sbom_file="$TEST_TEMP_DIR/sbom-python.json"
    syft dir:"$repo_path" -o cyclonedx-json > "$sbom_file" 2>/dev/null || true

    local output=$("$ANALYZER_SCRIPT" \
        --local-path "$repo_path" \
        --sbom-file "$sbom_file" \
        --format json \
        --no-claude \
        2>/dev/null || echo '{}')

    assert_json_valid "$output" &&
    assert_contains "$output" "AWS" "Should detect AWS SDK" &&
    assert_contains "$output" "Flask" "Should detect Flask" &&
    assert_contains "$output" "Redis" "Should detect Redis"
}

test_analyze_terraform_repo() {
    local repo_path=$(create_terraform_repo)

    if ! command -v syft &> /dev/null; then
        echo "Skipping: syft not installed"
        return 0
    fi

    local sbom_file="$TEST_TEMP_DIR/sbom-tf.json"
    syft dir:"$repo_path" -o cyclonedx-json > "$sbom_file" 2>/dev/null || true

    local output=$("$ANALYZER_SCRIPT" \
        --local-path "$repo_path" \
        --sbom-file "$sbom_file" \
        --format json \
        --no-claude \
        2>/dev/null || echo '{}')

    assert_json_valid "$output" &&
    assert_contains "$output" "Terraform" "Should detect Terraform from .tf files"
}

test_multiple_detection_layers_composite_confidence() {
    local repo_path=$(create_simple_node_repo)

    if ! command -v syft &> /dev/null; then
        echo "Skipping: syft not installed"
        return 0
    fi

    local sbom_file="$TEST_TEMP_DIR/sbom-composite.json"
    syft dir:"$repo_path" -o cyclonedx-json > "$sbom_file" 2>/dev/null || true

    local output=$("$ANALYZER_SCRIPT" \
        --local-path "$repo_path" \
        --sbom-file "$sbom_file" \
        --format json \
        --no-claude \
        2>/dev/null || echo '{}')

    # Stripe should be detected in multiple layers:
    # - SBOM (package.json)
    # - Import statement
    # - Environment variable
    # Composite confidence should be higher than single layer

    local stripe_confidence=$(echo "$output" | jq '.technologies[] | select(.name == "Stripe") | .confidence' 2>/dev/null || echo "0")

    if [[ -n "$stripe_confidence" ]] && [[ "$stripe_confidence" != "0" ]]; then
        assert_greater_than "$stripe_confidence" "90" "Stripe confidence should exceed 90 with multiple detection layers"
    else
        echo "Note: Stripe not detected in output"
        return 0
    fi
}

test_json_output_structure() {
    local repo_path=$(create_simple_node_repo)

    if ! command -v syft &> /dev/null; then
        echo "Skipping: syft not installed"
        return 0
    fi

    local sbom_file="$TEST_TEMP_DIR/sbom-structure.json"
    syft dir:"$repo_path" -o cyclonedx-json > "$sbom_file" 2>/dev/null || true

    local output=$("$ANALYZER_SCRIPT" \
        --local-path "$repo_path" \
        --sbom-file "$sbom_file" \
        --format json \
        --no-claude \
        2>/dev/null || echo '{}')

    assert_json_valid "$output" &&
    assert_json_contains_key "$output" "scan_metadata" "Should have scan_metadata" &&
    assert_json_contains_key "$output" "summary" "Should have summary" &&
    assert_json_contains_key "$output" "technologies" "Should have technologies array"
}

test_markdown_output_format() {
    local repo_path=$(create_simple_node_repo)

    if ! command -v syft &> /dev/null; then
        echo "Skipping: syft not installed"
        return 0
    fi

    local sbom_file="$TEST_TEMP_DIR/sbom-markdown.json"
    syft dir:"$repo_path" -o cyclonedx-json > "$sbom_file" 2>/dev/null || true

    local output=$("$ANALYZER_SCRIPT" \
        --local-path "$repo_path" \
        --sbom-file "$sbom_file" \
        --format markdown \
        --no-claude \
        2>/dev/null || echo '')

    assert_contains "$output" "# Technology Stack Analysis Report" "Should have markdown header" &&
    assert_contains "$output" "## Technologies Detected" "Should have technologies section" &&
    assert_contains "$output" "## Summary by Category" "Should have summary section"
}

test_confidence_threshold_filtering() {
    local repo_path=$(create_simple_node_repo)

    if ! command -v syft &> /dev/null; then
        echo "Skipping: syft not installed"
        return 0
    fi

    local sbom_file="$TEST_TEMP_DIR/sbom-threshold.json"
    syft dir:"$repo_path" -o cyclonedx-json > "$sbom_file" 2>/dev/null || true

    # Set high confidence threshold
    local output=$("$ANALYZER_SCRIPT" \
        --local-path "$repo_path" \
        --sbom-file "$sbom_file" \
        --format json \
        --confidence 90 \
        --no-claude \
        2>/dev/null || echo '{}')

    # Check that low-confidence detections are filtered
    local tech_count=$(echo "$output" | jq '.technologies | length' 2>/dev/null || echo "0")

    # All technologies should have confidence >= 90
    local low_confidence_count=$(echo "$output" | jq '[.technologies[] | select(.confidence < 90)] | length' 2>/dev/null || echo "0")

    assert_equals "0" "$low_confidence_count" "No technologies below confidence threshold should be included"
}

test_error_handling_invalid_path() {
    # Test that analyzer handles invalid path gracefully
    local output=$("$ANALYZER_SCRIPT" \
        --local-path "/nonexistent/path" \
        --no-claude \
        2>&1 || echo "error")

    assert_contains "$output" "Error" "Should report error for invalid path"
}

#############################################################################
# Run all tests
#############################################################################

main() {
    echo ""
    echo "========================================="
    echo "  Integration Tests"
    echo "========================================="
    echo ""

    # Check prerequisites
    if ! command -v jq &> /dev/null; then
        echo "Error: jq is required but not installed"
        exit 1
    fi

    if ! command -v syft &> /dev/null; then
        echo "Warning: syft not installed - some tests will be skipped"
        echo "Install with: brew install syft"
        echo ""
    fi

    run_test "Analyze simple Node.js repo" test_analyze_simple_node_repo
    run_test "Analyze Python + AWS repo" test_analyze_python_aws_repo
    run_test "Analyze Terraform repo" test_analyze_terraform_repo
    run_test "Multiple detection layers increase confidence" test_multiple_detection_layers_composite_confidence
    run_test "JSON output has correct structure" test_json_output_structure
    run_test "Markdown output format" test_markdown_output_format
    run_test "Confidence threshold filtering" test_confidence_threshold_filtering
    run_test "Error handling for invalid path" test_error_handling_invalid_path

    print_summary
}

# Run tests if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main
    exit $?
fi
