<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Certificate Expiry Monitoring Prompt

## Purpose

Monitor certificate expiration dates, assess renewal urgency, and plan certificate lifecycle management.

## Usage

### Basic Expiry Check

```
Check the certificate expiry status for [domain].

Report:
1. Expiration date
2. Days until expiry
3. Urgency level (Critical/Warning/OK)
4. Renewal recommendation
5. Automation readiness

Provide renewal timeline recommendations.
```

### Using Certificate Analyser

```bash
# Basic expiry check
./utils/certificate-analyser/cert-analyser.sh example.com

# Multiple domains
for domain in api.example.com www.example.com; do
    ./utils/certificate-analyser/cert-analyser.sh "$domain" | grep -A2 "Expiry"
done
```

## Example Output

### Single Domain Report

```
Certificate Expiry Report
=========================
Domain: example.com
Checked: 2024-11-25

Expiration: 2025-03-15 23:59:59 UTC
Days Remaining: 110 days
Status: ✓ OK

Renewal Timeline:
- Recommended renewal: 2025-02-15 (30 days before)
- Warning threshold: 2025-03-01 (14 days before)
- Critical threshold: 2025-03-08 (7 days before)
```

### Multi-Domain Summary

| Domain | Expires | Days | Status |
|--------|---------|------|--------|
| www.example.com | 2025-03-15 | 110 | ✓ OK |
| api.example.com | 2025-01-10 | 45 | ⚠️ Warning |
| mail.example.com | 2024-12-05 | 10 | ❌ Critical |
| admin.example.com | 2025-06-20 | 207 | ✓ OK |

### Urgency Levels

| Level | Days Remaining | Action |
|-------|----------------|--------|
| ❌ Critical | < 7 days | Immediate renewal required |
| ⚠️ Warning | 7-30 days | Schedule renewal now |
| ✓ OK | > 30 days | No immediate action |

## Variations

### Bulk Domain Monitoring

```
Generate an expiry report for all certificates:

Domains:
- api.example.com
- www.example.com
- mail.example.com
- cdn.example.com
- admin.example.com

Sort by expiration date (soonest first).
Highlight any within 30 days of expiry.
Calculate overall renewal workload.
```

### Renewal Planning

```
Create a certificate renewal plan for [domain]:

Certificate expires: [date]

Include:
1. Optimal renewal window
2. Pre-renewal checklist
3. Validation method recommendation
4. Deployment steps
5. Rollback plan
6. Verification steps

Consider automation opportunities.
```

### Automation Assessment

```
Assess certificate renewal automation readiness for [domain]:

Current state:
- Manual or automated renewal?
- What CA is used?
- How is deployment done?

Recommend:
- ACME/Let's Encrypt feasibility
- cert-manager (if Kubernetes)
- Certbot integration
- Cloud provider automation (ACM, Cloud Load Balancing)

Provide implementation steps.
```

### Chain Certificate Expiry

```
Check expiry for the entire certificate chain for [domain]:

Include:
- Leaf certificate expiry
- Intermediate certificate(s) expiry
- Root certificate expiry (for awareness)

Flag if any intermediate expires before leaf.
Note upcoming CA root expirations.
```

### Historical Tracking

```
Generate certificate history report for [domain]:

Track:
- Previous certificates (via CT logs)
- Validity period trends
- CA changes over time
- Key rotation history

Use crt.sh or CT log data if available.
```

## Monitoring Thresholds

### Recommended Alert Levels

| Alert | Days Before Expiry | Action |
|-------|-------------------|--------|
| Info | 90 days | Planning notice |
| Planning | 60 days | Begin renewal process |
| Warning | 30 days | Active renewal |
| Alert | 14 days | Urgent renewal |
| Critical | 7 days | Emergency renewal |
| Expired | 0 days | Incident response |

### By Certificate Type

**Production Certificates**:
- Renew at 30 days remaining
- Alert at 14 days

**Wildcard Certificates**:
- Renew at 45 days (wider impact)
- Extra validation for coverage

**Internal Certificates**:
- May have longer validity
- Check CA certificate expiry too

## Expiry Calculation

```bash
# Get expiry date
openssl x509 -in cert.pem -noout -enddate

# Calculate days remaining (macOS)
expiry=$(openssl x509 -in cert.pem -noout -enddate | cut -d= -f2)
expiry_epoch=$(date -jf "%b %d %H:%M:%S %Y %Z" "$expiry" +%s)
now_epoch=$(date +%s)
days_remaining=$(( (expiry_epoch - now_epoch) / 86400 ))
echo "$days_remaining days remaining"

# Calculate days remaining (Linux)
expiry=$(openssl x509 -in cert.pem -noout -enddate | cut -d= -f2)
expiry_epoch=$(date -d "$expiry" +%s)
now_epoch=$(date +%s)
days_remaining=$(( (expiry_epoch - now_epoch) / 86400 ))
echo "$days_remaining days remaining"
```

## Renewal Checklist

### Pre-Renewal
- [ ] Verify domain ownership still valid
- [ ] Check DNS records current
- [ ] Confirm CAA records allow CA
- [ ] Review SAN list for completeness
- [ ] Plan maintenance window (if needed)

### During Renewal
- [ ] Generate new CSR (consider key rotation)
- [ ] Submit to CA
- [ ] Complete domain validation
- [ ] Download certificate and chain

### Post-Renewal
- [ ] Verify certificate contents
- [ ] Deploy to server
- [ ] Test TLS connection
- [ ] Verify OCSP stapling
- [ ] Update monitoring with new expiry
- [ ] Archive old certificate

## Automation Options

### ACME / Let's Encrypt

```bash
# Certbot (standalone)
certbot renew --dry-run
certbot renew

# cert-manager (Kubernetes)
kubectl get certificate -A
kubectl describe certificate my-cert
```

### Cloud Provider

- **AWS ACM**: Auto-renews managed certificates
- **Google Cloud**: Auto-renews for Cloud Load Balancing
- **Azure**: Key Vault auto-renewal
- **Cloudflare**: Universal SSL auto-renewal

### Self-Hosted

- HashiCorp Vault PKI
- EJBCA
- StepCA

## Integration Examples

### Prometheus/Alertmanager

```yaml
# Prometheus alert rule
- alert: CertificateExpiringSoon
  expr: probe_ssl_earliest_cert_expiry - time() < 86400 * 30
  for: 1h
  labels:
    severity: warning
  annotations:
    summary: "Certificate expiring within 30 days"
```

### Nagios/Icinga

```bash
# check_http with certificate check
check_http -H example.com -C 30,14
# Warn at 30 days, critical at 14 days
```

### Datadog

```yaml
# Synthetic test with certificate check
type: api
subtype: ssl
config:
  certificate_expiration_alert_days: 30
```

## Related Prompts

- [security-audit.md](../security/security-audit.md) - Full security audit
- [cab-forum-compliance.md](cab-forum-compliance.md) - Compliance check
- [certificate-comparison.md](../operations/certificate-comparison.md) - Compare old/new

## Related RAG

- [X.509 Certificates](../../../rag/certificate-analysis/x509/x509-certificates.md) - Certificate structure
- [CA/B Forum Requirements](../../../rag/certificate-analysis/cab-forum/baseline-requirements.md) - Validity limits
- [TLS Security](../../../rag/certificate-analysis/tls-security/best-practices.md) - Best practices
