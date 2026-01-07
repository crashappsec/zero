# Test file: Known secrets that SHOULD be detected
# These are intentionally invalid/test credentials

# AWS Access Key (should detect)
AWS_ACCESS_KEY_ID = "AKIAIOSFODNN7EXAMPLE"
AWS_SECRET_ACCESS_KEY = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"

# GitHub Token (should detect)
GITHUB_TOKEN = "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

# Slack Signing Secret (should detect) - using signing secret instead of bot token
SLACK_SIGNING_SECRET = "a0b1c2d3e4f5a0b1c2d3e4f5a0b1c2d3"

# Generic API Key patterns (should detect)
api_key = "sk-proj-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

# JWT Token (should detect)
jwt_token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"

# Private Key (should detect)
private_key = """-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA0Z3...EXAMPLE...
-----END RSA PRIVATE KEY-----"""

# Database connection string with password (should detect)
DATABASE_URL = "postgresql://user:password123@localhost:5432/mydb"

# Stripe Test Key (should detect) - using test key prefix
STRIPE_KEY = "sk_test_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

# SendGrid API Key (should detect)
SENDGRID_KEY = "SG.xxxxxxxxxxxxxxxxxxxxxx.xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
