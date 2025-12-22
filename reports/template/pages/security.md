---
title: Security
sidebar_position: 1
---

# Security

```sql severity_counts
select * from zero.severity_counts
```

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

```sql secrets
select * from zero.secrets
```

```sql crypto_findings
select * from zero.crypto_findings
```

## Overview

<Grid cols=4>
<BigValue
  data={severity_counts}
  value=critical
  title="Critical"
/>
<BigValue
  data={severity_counts}
  value=high
  title="High"
/>
<BigValue
  data={severity_counts}
  value=medium
  title="Medium"
/>
<BigValue
  data={severity_counts}
  value=low
  title="Low"
/>
</Grid>

---

## Code Vulnerabilities

{#if vulnerabilities.length > 0}

```sql vuln_by_severity
select severity, count(*) as count from zero.vulnerabilities group by severity
order by case severity when 'critical' then 1 when 'high' then 2 when 'medium' then 3 when 'low' then 4 else 5 end
```

<BarChart
  data={vuln_by_severity}
  x=severity
  y=count
  colorPalette={['#dc2626','#ea580c','#ca8a04','#22c55e']}
  swapXY=true
  title="Vulnerabilities by Severity"
/>

<DataTable
  data={vulnerabilities}
  search=true
  rows=25
  rowShading=true
>
  <Column id=severity title="Severity" contentType=colorscale colorScale=red/>
  <Column id=source title="Source"/>
  <Column id=package title="Package/File"/>
  <Column id=cve title="CVE/Rule"/>
  <Column id=title title="Description" wrap=true/>
  <Column id=fix_version title="Fix"/>
</DataTable>

{:else}

<Alert status="success">No code vulnerabilities detected.</Alert>

{/if}

---

## Secrets Detected

{#if secrets.length > 0}

<Alert status="warning">
<b>{secrets.length}</b> potential secrets found. These should be rotated and removed from the codebase.
</Alert>

```sql secrets_by_type
select type, count(*) as count from zero.secrets group by type order by count desc limit 10
```

<BarChart
  data={secrets_by_type}
  x=type
  y=count
  swapXY=true
  title="Secrets by Type"
/>

<DataTable
  data={secrets}
  search=true
  rows=25
  rowShading=true
>
  <Column id=type title="Type"/>
  <Column id=file title="File" wrap=true/>
  <Column id=line title="Line"/>
  <Column id=confidence title="Confidence"/>
</DataTable>

{:else}

<Alert status="success">No secrets detected.</Alert>

{/if}

---

## Cryptographic Issues

{#if crypto_findings.length > 0}

```sql crypto_by_type
select type, count(*) as count from zero.crypto_findings group by type order by count desc
```

<BarChart
  data={crypto_by_type}
  x=type
  y=count
  swapXY=true
  title="Crypto Issues by Type"
/>

<DataTable
  data={crypto_findings}
  search=true
  rows=25
  rowShading=true
>
  <Column id=type title="Type"/>
  <Column id=algorithm title="Algorithm"/>
  <Column id=file title="File" wrap=true/>
  <Column id=line title="Line"/>
  <Column id=severity title="Severity"/>
  <Column id=recommendation title="Recommendation" wrap=true/>
</DataTable>

{:else}

<Alert status="success">No cryptographic issues detected.</Alert>

{/if}

---

<Grid cols=2>
  <BigLink url="/">Back to Dashboard</BigLink>
  <BigLink url="/dependencies">Dependencies & SBOM</BigLink>
</Grid>
