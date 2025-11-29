# API Keys

## AI/ML Services

### OpenAI
```
Pattern: sk-[A-Za-z0-9]{48}
Pattern (project): sk-proj-[A-Za-z0-9]{48}
Example: sk-abc123...
Severity: high
```

OpenAI API keys start with `sk-` followed by 48 alphanumeric characters.

### Anthropic
```
Pattern: sk-ant-[A-Za-z0-9_-]{90,}
Example: sk-ant-api03-...
Severity: high
```

Anthropic API keys start with `sk-ant-` followed by ~95 characters.

### Cohere
```
Pattern: [A-Za-z0-9]{40}
Context: COHERE_API_KEY or near "cohere" imports
Severity: high
```

40-character alphanumeric strings.

### Hugging Face
```
Pattern: hf_[A-Za-z0-9]{34}
Example: hf_abc123...
Severity: medium
```

Hugging Face tokens start with `hf_`.

---

## Payment Services

### Stripe
```
Pattern (Live): sk_live_[0-9a-zA-Z]{24,}
Pattern (Test): sk_test_[0-9a-zA-Z]{24,}
Pattern (Restricted): rk_live_[0-9a-zA-Z]{24,}
Severity: critical (live), low (test)
```

Stripe secret keys. Live keys (`sk_live_`) are critical; test keys are low severity.

### PayPal
```
Pattern: access_token\$[a-zA-Z0-9]{32}\$[a-zA-Z0-9]{64}
Pattern (Client ID): A[a-zA-Z0-9_-]{79}
Severity: critical
```

PayPal OAuth tokens and client credentials.

### Square
```
Pattern: sq0[a-z]{3}-[A-Za-z0-9_-]{22}
Pattern (Sandbox): sandbox-sq0[a-z]{3}-[A-Za-z0-9_-]{22}
Severity: critical (production), low (sandbox)
```

Square API credentials.

---

## Communication Services

### Twilio
```
Pattern (Account SID): AC[a-f0-9]{32}
Pattern (Auth Token): [a-f0-9]{32}
Context: Near TWILIO_ environment variables
Severity: high
```

Twilio credentials come in pairs.

### SendGrid
```
Pattern: SG\.[A-Za-z0-9_-]{22}\.[A-Za-z0-9_-]{43}
Example: SG.abc123...xyz789
Severity: medium
```

SendGrid API keys have a distinctive dot-separated format.

### Mailgun
```
Pattern: key-[a-f0-9]{32}
Example: key-abc123...
Severity: medium
```

Mailgun API keys start with `key-`.

### Mailchimp
```
Pattern: [a-f0-9]{32}-us[0-9]{1,2}
Example: abc123...-us14
Severity: low
```

Mailchimp API keys include datacenter suffix.

---

## Developer Platforms

### GitHub
```
Pattern (Classic): ghp_[A-Za-z0-9_]{36,}
Pattern (Fine-grained): github_pat_[A-Za-z0-9_]{22,}
Pattern (OAuth): gho_[A-Za-z0-9_]{36,}
Pattern (App): ghs_[A-Za-z0-9_]{36,}
Pattern (Refresh): ghr_[A-Za-z0-9_]{36,}
Severity: high
```

GitHub tokens use prefixes to indicate type:
- `ghp_` - Personal access tokens
- `github_pat_` - Fine-grained PATs
- `gho_` - OAuth tokens
- `ghs_` - App installation tokens
- `ghr_` - Refresh tokens

### GitLab
```
Pattern: glpat-[A-Za-z0-9-]{20,}
Pattern (Runner): GR1348941[A-Za-z0-9_-]{20}
Severity: high
```

GitLab personal access tokens and runner tokens.

### NPM
```
Pattern: npm_[A-Za-z0-9]{36}
Severity: high
```

NPM automation tokens.

### PyPI
```
Pattern: pypi-AgEIcHlwaS5vcmc[A-Za-z0-9_-]+
Severity: high
```

PyPI API tokens (base64-encoded).

---

## Monitoring & Analytics

### Datadog
```
Pattern: [a-f0-9]{32}
Context: DD_API_KEY, DD_APP_KEY
Severity: medium
```

32-character hex strings.

### Sentry
```
Pattern: [a-f0-9]{32}
Context: SENTRY_DSN, sentry.io URLs
Severity: medium
```

Part of Sentry DSN URLs.

### New Relic
```
Pattern: NRAK-[A-Z0-9]{27}
Pattern (License): [a-f0-9]{40}NRAL
Severity: medium
```

New Relic API and license keys.

### Segment
```
Pattern: [A-Za-z0-9]{32}
Context: analytics.identify, SEGMENT_WRITE_KEY
Severity: medium
```

---

## Social Platforms

### Slack
```
Pattern (Bot): xoxb-[0-9]{10,13}-[0-9]{10,13}[a-zA-Z0-9-]*
Pattern (User): xoxp-[0-9]{10,13}-[0-9]{10,13}[a-zA-Z0-9-]*
Pattern (App): xoxa-[0-9]{10,13}-[0-9]{10,13}[a-zA-Z0-9-]*
Pattern (Webhook): https://hooks\.slack\.com/services/T[A-Z0-9]+/B[A-Z0-9]+/[A-Za-z0-9]+
Severity: medium
```

Slack tokens use `xox` prefix variants.

### Discord
```
Pattern (Bot): [MN][A-Za-z0-9_-]{23,}\.[A-Za-z0-9_-]{6}\.[A-Za-z0-9_-]{27}
Pattern (Webhook): https://discord(app)?\.com/api/webhooks/[0-9]+/[A-Za-z0-9_-]+
Severity: medium
```

Discord bot tokens and webhook URLs.

### Twitter/X
```
Pattern (API Key): [A-Za-z0-9]{25}
Pattern (Bearer): AAAAAAAAAAAAA[A-Za-z0-9%]+
Context: TWITTER_API_KEY, TWITTER_BEARER_TOKEN
Severity: medium
```

---

## Detection Notes

### Context Clues
- Environment variable assignments
- Configuration file keys
- SDK initialization calls
- Header values in HTTP code

### Common False Positives
- Placeholder strings (YOUR_API_KEY)
- Example documentation values
- Test/mock values in test files
- Base64 data that matches patterns

### Severity Considerations
- **Production vs Test**: Test keys are lower severity
- **Scope**: Keys with broad permissions are higher severity
- **Exposure**: Public repos vs private repos
