# Test file: Patterns that should NOT trigger secret detection
# These are NOT in secret formats, just look suspicious

import os

# Environment variable references (NOT secrets - just references)
api_key = os.environ.get("API_KEY")
secret = os.getenv("SECRET_KEY")
token = config["GITHUB_TOKEN"]

# Placeholder strings with obvious markers (NOT secrets)
password = "<YOUR_PASSWORD_HERE>"
api_key = "your-api-key-here"
token = "INSERT_TOKEN_HERE"
secret = "${SECRET}"
key = "{{API_KEY}}"

# Short/invalid length strings (NOT valid API key formats)
short_key = "abc123"
empty_key = ""

# Hash values (NOT API keys - just hashes)
sha256_hash = "a5b9d7c8e3f2a1b0c9d8e7f6a5b4c3d2e1f0a9b8c7d6e5f4a3b2c1d0e9f8a7b6"
md5_hash = "d41d8cd98f00b204e9800998ecf8427e"

# Random strings that don't match patterns
random_string = "hello_world_this_is_not_a_secret"

# Version strings
version = "v1.2.3-beta.4"

# File paths that look like they might contain secrets but don't
config_path = "/etc/secrets/config.json"
key_file = "~/.ssh/id_rsa"
