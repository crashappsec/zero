# Test file: False positives that should NOT be detected as secrets
# These patterns look like secrets but are not

# Placeholder values (should NOT detect)
AWS_ACCESS_KEY = "AKIAIOSFODNN7EXAMPLE"  # AWS documentation example
AWS_SECRET_KEY = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"  # AWS example

# Environment variable references (should NOT detect)
api_key = os.environ.get("API_KEY")
secret = os.getenv("SECRET_KEY")
token = config["GITHUB_TOKEN"]

# Placeholder strings (should NOT detect)
password = "<YOUR_PASSWORD_HERE>"
api_key = "your-api-key-here"
token = "INSERT_TOKEN_HERE"
secret = "${SECRET}"
key = "{{API_KEY}}"

# Test data markers (should NOT detect)
test_password = "test123"  # Only used in tests
mock_key = "mock-api-key-for-testing"

# Hash values that look like keys (should NOT detect as API keys)
sha256_hash = "a5b9d7c8e3f2a1b0c9d8e7f6a5b4c3d2e1f0a9b8c7d6e5f4a3b2c1d0e9f8a7b6"
md5_hash = "d41d8cd98f00b204e9800998ecf8427e"

# Base64 encoded non-secrets (should NOT detect)
base64_data = "SGVsbG8gV29ybGQh"  # Just "Hello World!"

# UUIDs (should NOT detect as secrets)
user_id = "550e8400-e29b-41d4-a716-446655440000"
request_id = "123e4567-e89b-12d3-a456-426614174000"

# Version strings that look like tokens (should NOT detect)
version = "v1.2.3-beta.4"

# Comment with example key (should still be safe as it's clearly an example)
# Example: api_key = "sk-example-key-do-not-use"
