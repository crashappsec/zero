# Cloud Provider Credentials

## AWS (Amazon Web Services)

### Access Keys
```
Pattern: AKIA[0-9A-Z]{16}
Example: AKIAIOSFODNN7EXAMPLE
Severity: critical
```

AWS Access Key IDs always start with `AKIA` followed by 16 alphanumeric characters.

### Secret Access Keys
```
Pattern: [A-Za-z0-9/+=]{40}
Context: Often paired with AKIA keys
Severity: critical
```

40-character base64-encoded strings, usually found near access key IDs.

### Session Tokens
```
Pattern: FwoGZXIvYXdzE[A-Za-z0-9/+=]+
Severity: high
```

Temporary credentials from STS, start with `FwoGZXIvYXdzE`.

### ARN Patterns (informational)
```
Pattern: arn:aws:[a-z0-9-]+:[a-z0-9-]*:[0-9]*:[a-zA-Z0-9-_/:.]+
Severity: informational
```

Not secrets themselves but may expose infrastructure details.

---

## GCP (Google Cloud Platform)

### Service Account Keys
```
Pattern: "type": "service_account"
File: *.json files containing service account credentials
Severity: critical
```

JSON files containing:
- `type`: "service_account"
- `project_id`: Project identifier
- `private_key_id`: Key identifier
- `private_key`: RSA private key (PEM format)
- `client_email`: Service account email

### API Keys
```
Pattern: AIza[0-9A-Za-z_-]{35}
Example: AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe
Severity: high
```

GCP API keys start with `AIza` followed by 35 characters.

### OAuth Client Secrets
```
Pattern: [0-9]+-[0-9A-Za-z_]{32}\.apps\.googleusercontent\.com
Severity: high
```

OAuth 2.0 client IDs for Google APIs.

---

## Azure

### Storage Account Keys
```
Pattern: [A-Za-z0-9+/]{86}==
Context: Near "AccountKey=" or storage connection strings
Severity: critical
```

Base64-encoded 64-byte keys.

### Connection Strings
```
Pattern: DefaultEndpointsProtocol=https;AccountName=[^;]+;AccountKey=[A-Za-z0-9+/]{86}==
Severity: critical
```

Full Azure Storage connection strings.

### Service Principal Secrets
```
Pattern: [a-zA-Z0-9_~.-]{34}
Context: Near "client_secret" or "AZURE_CLIENT_SECRET"
Severity: critical
```

Azure AD application secrets.

### SAS Tokens
```
Pattern: sv=[0-9-]+&s[a-z]=[a-z]+&[a-z]+=[^&]+(&[a-z]+=[^&]+)*&sig=[A-Za-z0-9%/+=]+
Severity: high
```

Shared Access Signature tokens for Azure resources.

---

## DigitalOcean

### Personal Access Tokens
```
Pattern: dop_v1_[a-f0-9]{64}
Example: dop_v1_abc123...
Severity: high
```

DigitalOcean API tokens start with `dop_v1_`.

### Spaces Keys
```
Pattern: DO[0-9A-Z]{18}
Severity: high
```

DigitalOcean Spaces access keys.

---

## Heroku

### API Keys
```
Pattern: [0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}
Context: HEROKU_API_KEY environment variable
Severity: high
```

UUID-formatted API keys.

---

## Detection Notes

### Critical Indicators
- Files named `credentials`, `secrets`, `.aws/credentials`
- Environment variables: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`
- Configuration files with embedded credentials

### Common False Positives
- Example/placeholder values: `AKIAIOSFODNN7EXAMPLE`
- Documentation strings
- Test fixtures with mock credentials
- Base64-encoded non-secret data matching patterns

### Remediation Priority
1. **Critical**: Rotate immediately, revoke old credentials
2. **High**: Rotate within 24 hours
3. **Medium**: Schedule rotation, assess exposure
