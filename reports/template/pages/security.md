---
title: Security Findings
---

# Security Findings

<Alert status="warning">
This page shows code vulnerabilities, secrets, and cryptographic issues detected in the codebase.
</Alert>

## Vulnerabilities

```sql vulnerabilities
select * from zero.vulnerabilities order by
  case severity
    when 'critical' then 1
    when 'high' then 2
    when 'medium' then 3
    when 'low' then 4
    else 5
  end
```

{#if vulnerabilities.length > 0}

<DataTable
  data={vulnerabilities}
  search=true
  rows=20
>
  <Column id=severity title="Severity"/>
  <Column id=source title="Source"/>
  <Column id=package title="Package/File"/>
  <Column id=cve title="CVE/Rule"/>
  <Column id=title title="Description"/>
  <Column id=fix_version title="Fix Version"/>
</DataTable>

{:else}

<Alert status="success">
No vulnerabilities detected.
</Alert>

{/if}

## Secrets Detected

```sql secrets
select * from zero.secrets
```

{#if secrets.length > 0}

<Alert status="error">
{secrets.length} potential secrets found in the codebase. These should be rotated and removed.
</Alert>

<DataTable
  data={secrets}
  search=true
  rows=20
>
  <Column id=type title="Type"/>
  <Column id=file title="File"/>
  <Column id=line title="Line"/>
  <Column id=confidence title="Confidence"/>
</DataTable>

{:else}

<Alert status="success">
No secrets detected.
</Alert>

{/if}

## Cryptographic Issues

```sql crypto_findings
select * from zero.crypto_findings
```

{#if crypto_findings.length > 0}

<DataTable
  data={crypto_findings}
  search=true
  rows=20
>
  <Column id=type title="Type"/>
  <Column id=algorithm title="Algorithm"/>
  <Column id=file title="File"/>
  <Column id=line title="Line"/>
  <Column id=severity title="Severity"/>
  <Column id=recommendation title="Recommendation"/>
</DataTable>

{:else}

<Alert status="success">
No cryptographic issues detected.
</Alert>

{/if}

---

<ButtonGroup>
  <BigLink href="/">Back to Overview</BigLink>
  <BigLink href="/dependencies">Dependencies</BigLink>
</ButtonGroup>
