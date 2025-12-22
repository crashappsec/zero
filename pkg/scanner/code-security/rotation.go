package codesecurity

// RotationDatabase provides rotation guidance for various secret types
type RotationDatabase struct {
	guides map[string]*RotationGuide
}

// NewRotationDatabase creates a new rotation database with built-in guides
func NewRotationDatabase() *RotationDatabase {
	db := &RotationDatabase{
		guides: make(map[string]*RotationGuide),
	}
	db.initializeGuides()
	return db
}

// GetGuide returns rotation guidance for a secret type
func (db *RotationDatabase) GetGuide(secretType string) *RotationGuide {
	if guide, ok := db.guides[secretType]; ok {
		return guide
	}
	// Return generic guidance for unknown types
	return db.guides["generic_secret"]
}

// GetServiceProvider returns the service provider for a secret type
func GetServiceProvider(secretType string) string {
	providers := map[string]string{
		"aws_access_key":        "aws",
		"aws_secret_key":        "aws",
		"github_token":          "github",
		"github_app_key":        "github",
		"stripe_secret_key":     "stripe",
		"stripe_publishable_key": "stripe",
		"slack_token":           "slack",
		"slack_webhook":         "slack",
		"openai_api_key":        "openai",
		"anthropic_api_key":     "anthropic",
		"google_api_key":        "google",
		"gcp_service_account":   "gcp",
		"azure_secret":          "azure",
		"database_credential":   "database",
		"mysql_password":        "database",
		"postgres_password":     "database",
		"mongodb_uri":           "database",
		"redis_password":        "database",
		"private_key":           "crypto",
		"ssh_private_key":       "crypto",
		"jwt_secret":            "jwt",
		"jwt_token":             "jwt",
		"twilio_auth_token":     "twilio",
		"sendgrid_api_key":      "sendgrid",
		"mailchimp_api_key":     "mailchimp",
		"npm_token":             "npm",
		"pypi_token":            "pypi",
		"docker_auth":           "docker",
		"heroku_api_key":        "heroku",
		"vercel_token":          "vercel",
		"netlify_token":         "netlify",
		"datadog_api_key":       "datadog",
		"newrelic_license_key":  "newrelic",
		"sentry_dsn":            "sentry",
	}
	if provider, ok := providers[secretType]; ok {
		return provider
	}
	return "unknown"
}

// initializeGuides populates the rotation database with guides
func (db *RotationDatabase) initializeGuides() {
	// AWS Access Keys
	db.guides["aws_access_key"] = &RotationGuide{
		Priority: "immediate",
		Steps: []string{
			"1. Log into AWS Console or use AWS CLI",
			"2. Navigate to IAM > Users > Security credentials",
			"3. Create a new access key pair",
			"4. Update all applications using the old key",
			"5. Test applications with new credentials",
			"6. Deactivate and then delete the old access key",
			"7. Review CloudTrail logs for unauthorized usage",
		},
		RotationURL:    "https://console.aws.amazon.com/iam/home#/security_credentials",
		CLICommand:     "aws iam create-access-key --user-name USERNAME && aws iam delete-access-key --user-name USERNAME --access-key-id OLD_KEY_ID",
		AutomationHint: "Use AWS Secrets Manager with automatic rotation: aws secretsmanager rotate-secret",
	}

	db.guides["aws_secret_key"] = db.guides["aws_access_key"]

	// GitHub Tokens
	db.guides["github_token"] = &RotationGuide{
		Priority: "immediate",
		Steps: []string{
			"1. Go to GitHub Settings > Developer settings > Personal access tokens",
			"2. Click 'Regenerate token' or create a new token",
			"3. Select appropriate scopes (follow least privilege)",
			"4. Update all applications and CI/CD pipelines",
			"5. Revoke the old token",
			"6. Review GitHub audit log for unauthorized access",
		},
		RotationURL:    "https://github.com/settings/tokens",
		CLICommand:     "gh auth refresh --scopes SCOPES",
		AutomationHint: "Consider using GitHub Apps with short-lived installation tokens instead of PATs",
	}

	db.guides["github_app_key"] = &RotationGuide{
		Priority: "immediate",
		Steps: []string{
			"1. Go to GitHub > Settings > Developer settings > GitHub Apps",
			"2. Select your app and go to Private keys section",
			"3. Generate a new private key",
			"4. Update your application with the new key",
			"5. Delete the old private key from GitHub",
		},
		RotationURL:    "https://github.com/settings/apps",
		AutomationHint: "Store private keys in a secrets manager like HashiCorp Vault",
	}

	// Stripe Keys
	db.guides["stripe_secret_key"] = &RotationGuide{
		Priority: "immediate",
		Steps: []string{
			"1. Log into Stripe Dashboard",
			"2. Go to Developers > API keys",
			"3. Click 'Roll secret key' to generate new key",
			"4. Update all applications with new key",
			"5. The old key will be invalidated after rolling",
			"6. Review Stripe logs for unauthorized transactions",
		},
		RotationURL:    "https://dashboard.stripe.com/apikeys",
		AutomationHint: "Use Stripe's restricted API keys with limited permissions",
		ExpiresIn:      "Rolling invalidates old key immediately",
	}

	db.guides["stripe_publishable_key"] = &RotationGuide{
		Priority: "high",
		Steps: []string{
			"1. Log into Stripe Dashboard",
			"2. Go to Developers > API keys",
			"3. Click 'Roll publishable key'",
			"4. Update frontend applications with new key",
			"5. Deploy updated frontend code",
		},
		RotationURL: "https://dashboard.stripe.com/apikeys",
	}

	// Slack Tokens
	db.guides["slack_token"] = &RotationGuide{
		Priority: "immediate",
		Steps: []string{
			"1. Go to api.slack.com/apps and select your app",
			"2. Navigate to OAuth & Permissions",
			"3. Reinstall app to workspace to get new tokens",
			"4. Update your application with new tokens",
			"5. Review Slack audit logs for unauthorized access",
		},
		RotationURL:    "https://api.slack.com/apps",
		AutomationHint: "Use Slack's token rotation feature if available for your app type",
	}

	db.guides["slack_webhook"] = &RotationGuide{
		Priority: "high",
		Steps: []string{
			"1. Go to api.slack.com/apps and select your app",
			"2. Navigate to Incoming Webhooks",
			"3. Remove the compromised webhook URL",
			"4. Create a new webhook URL",
			"5. Update your application with new webhook URL",
		},
		RotationURL: "https://api.slack.com/apps",
	}

	// OpenAI API Keys
	db.guides["openai_api_key"] = &RotationGuide{
		Priority: "immediate",
		Steps: []string{
			"1. Log into OpenAI platform",
			"2. Go to API keys section",
			"3. Create a new secret key",
			"4. Update all applications with new key",
			"5. Delete the compromised key",
			"6. Review usage to check for unauthorized API calls",
		},
		RotationURL:    "https://platform.openai.com/api-keys",
		AutomationHint: "Use environment variables and secrets managers, never hardcode",
	}

	// Anthropic API Keys
	db.guides["anthropic_api_key"] = &RotationGuide{
		Priority: "immediate",
		Steps: []string{
			"1. Log into Anthropic Console",
			"2. Go to API keys section",
			"3. Create a new API key",
			"4. Update all applications with new key",
			"5. Revoke the compromised key",
			"6. Review usage logs for unauthorized access",
		},
		RotationURL:    "https://console.anthropic.com/settings/keys",
		AutomationHint: "Store API keys in environment variables or secrets manager",
	}

	// Database Credentials
	db.guides["database_credential"] = &RotationGuide{
		Priority: "immediate",
		Steps: []string{
			"1. Generate a new strong password",
			"2. Create new database user with same permissions or update password",
			"3. Update connection strings in all applications",
			"4. Test database connectivity",
			"5. Deploy updated applications",
			"6. Remove or disable old credentials",
			"7. Review database audit logs",
		},
		CLICommand:     "ALTER USER username IDENTIFIED BY 'new_password';",
		AutomationHint: "Use AWS Secrets Manager or HashiCorp Vault for automatic rotation",
	}

	db.guides["mysql_password"] = db.guides["database_credential"]
	db.guides["postgres_password"] = db.guides["database_credential"]
	db.guides["mongodb_uri"] = db.guides["database_credential"]
	db.guides["redis_password"] = db.guides["database_credential"]

	// Private Keys
	db.guides["private_key"] = &RotationGuide{
		Priority: "immediate",
		Steps: []string{
			"1. Generate a new key pair",
			"2. Update public key in all authorized locations",
			"3. Update applications with new private key",
			"4. Securely delete old private key",
			"5. Revoke old public key from all authorized_keys files",
		},
		CLICommand:     "ssh-keygen -t ed25519 -C 'your_email@example.com'",
		AutomationHint: "Use SSH certificates or short-lived keys with automatic rotation",
	}

	db.guides["ssh_private_key"] = db.guides["private_key"]

	// JWT Secrets
	db.guides["jwt_secret"] = &RotationGuide{
		Priority: "high",
		Steps: []string{
			"1. Generate a new secure random secret",
			"2. Update JWT signing configuration",
			"3. Consider using key rotation with multiple valid keys",
			"4. Invalidate all existing tokens (users will need to re-authenticate)",
			"5. Deploy updated application",
		},
		CLICommand:     "openssl rand -base64 64",
		AutomationHint: "Use asymmetric keys (RS256) for easier rotation without invalidating tokens",
	}

	db.guides["jwt_token"] = &RotationGuide{
		Priority: "medium",
		Steps: []string{
			"1. Tokens expire automatically based on exp claim",
			"2. If token is long-lived, add to blacklist/revocation list",
			"3. Request a new token through normal authentication flow",
		},
		AutomationHint: "Use short-lived tokens with refresh token rotation",
	}

	// Google/GCP
	db.guides["google_api_key"] = &RotationGuide{
		Priority: "immediate",
		Steps: []string{
			"1. Go to Google Cloud Console > APIs & Services > Credentials",
			"2. Create a new API key",
			"3. Apply appropriate restrictions to new key",
			"4. Update applications with new key",
			"5. Delete the compromised key",
		},
		RotationURL:    "https://console.cloud.google.com/apis/credentials",
		AutomationHint: "Use service accounts with short-lived tokens instead of API keys",
	}

	db.guides["gcp_service_account"] = &RotationGuide{
		Priority: "immediate",
		Steps: []string{
			"1. Go to Google Cloud Console > IAM & Admin > Service Accounts",
			"2. Select the service account",
			"3. Go to Keys tab and create new key",
			"4. Update applications with new key file",
			"5. Delete the old key",
			"6. Review Cloud Audit Logs",
		},
		RotationURL:    "https://console.cloud.google.com/iam-admin/serviceaccounts",
		CLICommand:     "gcloud iam service-accounts keys create KEY_FILE --iam-account=SA_EMAIL",
		AutomationHint: "Use Workload Identity Federation to eliminate service account keys",
	}

	// Azure
	db.guides["azure_secret"] = &RotationGuide{
		Priority: "immediate",
		Steps: []string{
			"1. Go to Azure Portal > Azure Active Directory > App registrations",
			"2. Select your application",
			"3. Go to Certificates & secrets",
			"4. Create a new client secret",
			"5. Update applications with new secret",
			"6. Delete the old secret",
		},
		RotationURL:    "https://portal.azure.com/#blade/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/RegisteredApps",
		CLICommand:     "az ad app credential reset --id APP_ID",
		AutomationHint: "Use Managed Identities to eliminate secrets where possible",
	}

	// Third-party Services
	db.guides["twilio_auth_token"] = &RotationGuide{
		Priority: "immediate",
		Steps: []string{
			"1. Log into Twilio Console",
			"2. Go to Account > API keys & tokens",
			"3. Create a new API key (preferred) or request secondary auth token",
			"4. Update applications",
			"5. Revoke old credentials",
		},
		RotationURL: "https://console.twilio.com/us1/account/keys-credentials/api-keys",
	}

	db.guides["sendgrid_api_key"] = &RotationGuide{
		Priority: "immediate",
		Steps: []string{
			"1. Log into SendGrid",
			"2. Go to Settings > API Keys",
			"3. Create a new API key with same permissions",
			"4. Update applications",
			"5. Delete the old API key",
		},
		RotationURL: "https://app.sendgrid.com/settings/api_keys",
	}

	db.guides["mailchimp_api_key"] = &RotationGuide{
		Priority: "immediate",
		Steps: []string{
			"1. Log into Mailchimp",
			"2. Go to Account > Extras > API keys",
			"3. Create a new API key",
			"4. Update applications",
			"5. Disable the old API key",
		},
		RotationURL: "https://us1.admin.mailchimp.com/account/api/",
	}

	// Package Registry Tokens
	db.guides["npm_token"] = &RotationGuide{
		Priority: "immediate",
		Steps: []string{
			"1. Log into npmjs.com",
			"2. Go to Access Tokens",
			"3. Create a new token with appropriate scope",
			"4. Update CI/CD pipelines and .npmrc files",
			"5. Revoke the old token",
			"6. Check if any malicious packages were published",
		},
		RotationURL:    "https://www.npmjs.com/settings/tokens",
		CLICommand:     "npm token create --read-only",
		AutomationHint: "Use granular access tokens with minimal required permissions",
	}

	db.guides["pypi_token"] = &RotationGuide{
		Priority: "immediate",
		Steps: []string{
			"1. Log into pypi.org",
			"2. Go to Account settings > API tokens",
			"3. Create a new token with project scope",
			"4. Update CI/CD pipelines",
			"5. Remove the old token",
		},
		RotationURL: "https://pypi.org/manage/account/token/",
	}

	// Platform Tokens
	db.guides["heroku_api_key"] = &RotationGuide{
		Priority: "immediate",
		Steps: []string{
			"1. Log into Heroku Dashboard",
			"2. Go to Account Settings",
			"3. Regenerate API Key",
			"4. Update Heroku CLI and applications",
		},
		RotationURL: "https://dashboard.heroku.com/account",
		CLICommand:  "heroku authorizations:create",
	}

	db.guides["vercel_token"] = &RotationGuide{
		Priority: "immediate",
		Steps: []string{
			"1. Log into Vercel Dashboard",
			"2. Go to Settings > Tokens",
			"3. Create a new token",
			"4. Update CI/CD pipelines",
			"5. Delete the old token",
		},
		RotationURL: "https://vercel.com/account/tokens",
	}

	db.guides["netlify_token"] = &RotationGuide{
		Priority: "immediate",
		Steps: []string{
			"1. Log into Netlify",
			"2. Go to User settings > Applications > Personal access tokens",
			"3. Create a new token",
			"4. Update applications and CI/CD",
			"5. Delete the old token",
		},
		RotationURL: "https://app.netlify.com/user/applications",
	}

	// Monitoring Services
	db.guides["datadog_api_key"] = &RotationGuide{
		Priority: "high",
		Steps: []string{
			"1. Log into Datadog",
			"2. Go to Organization Settings > API Keys",
			"3. Create a new API key",
			"4. Update monitoring agents and applications",
			"5. Revoke the old key",
		},
		RotationURL: "https://app.datadoghq.com/organization-settings/api-keys",
	}

	db.guides["newrelic_license_key"] = &RotationGuide{
		Priority: "high",
		Steps: []string{
			"1. Log into New Relic",
			"2. Go to API keys",
			"3. Create a new license key (requires account admin)",
			"4. Update all New Relic agents",
			"5. Delete the old key",
		},
		RotationURL: "https://one.newrelic.com/launcher/api-keys-ui.api-keys-launcher",
	}

	db.guides["sentry_dsn"] = &RotationGuide{
		Priority: "medium",
		Steps: []string{
			"1. Log into Sentry",
			"2. Go to Project Settings > Client Keys (DSN)",
			"3. Create a new key",
			"4. Update applications with new DSN",
			"5. Revoke the old key",
		},
		RotationURL: "https://sentry.io/settings/",
	}

	// Generic fallback
	db.guides["generic_secret"] = &RotationGuide{
		Priority: "high",
		Steps: []string{
			"1. Identify the service or system the secret belongs to",
			"2. Generate a new secret/credential through the service's admin interface",
			"3. Update all applications and services using this secret",
			"4. Test that applications work with new credentials",
			"5. Revoke or delete the old secret",
			"6. Review logs for any unauthorized access",
		},
		AutomationHint: "Store secrets in a secrets manager (HashiCorp Vault, AWS Secrets Manager, etc.)",
	}

	db.guides["api_key"] = db.guides["generic_secret"]
	db.guides["high_entropy_string"] = db.guides["generic_secret"]
}

// EnrichWithRotation adds rotation guidance to a slice of secret findings
func EnrichWithRotation(findings []SecretFinding, db *RotationDatabase) []SecretFinding {
	for i := range findings {
		if findings[i].Rotation == nil {
			guide := db.GetGuide(findings[i].Type)
			findings[i].Rotation = guide
			findings[i].ServiceProvider = GetServiceProvider(findings[i].Type)
		}
	}
	return findings
}
