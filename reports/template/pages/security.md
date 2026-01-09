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
  <Column id=severity title="Severity" contentType=colorscale colorScale=negative/>
  <Column id=source title="Source"/>
  <Column id=package title="Package/File"/>
  <Column id=cve title="CVE/Rule"/>
  <Column id=title title="Description" wrap=true/>
  <Column id=fix_version title="Fix"/>
</DataTable>

{:else}

<Alert status="positive">No code vulnerabilities detected.</Alert>

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

<Alert status="positive">No secrets detected.</Alert>

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

<Alert status="positive">No cryptographic issues detected.</Alert>

{/if}

---

## Git History Security

```sql git_history_summary
select * from zero.git_history_summary
```

```sql sensitive_files
select * from zero.sensitive_files order by
  case severity
    when 'critical' then 1
    when 'high' then 2
    when 'medium' then 3
    when 'low' then 4
    else 5
  end
```

```sql purge_recommendations
select * from zero.purge_recommendations order by priority
```

{#if git_history_summary[0].total_violations > 0}

<Alert status="warning">
<b>{git_history_summary[0].total_violations}</b> security issues found in git history. These files may have leaked sensitive data even if later removed.
</Alert>

<Grid cols=4>
<BigValue
  data={git_history_summary}
  value=sensitive_files_found
  title="Sensitive Files"
/>
<BigValue
  data={git_history_summary}
  value=gitignore_violations
  title="Gitignore Violations"
/>
<BigValue
  data={git_history_summary}
  value=files_to_purge
  title="Files to Purge"
/>
<BigValue
  data={git_history_summary}
  value=commits_scanned
  title="Commits Scanned"
/>
</Grid>

### Sensitive Files in History

Files containing credentials, keys, or other sensitive data found in git history:

<DataTable
  data={sensitive_files}
  search=true
  rows=25
  rowShading=true
>
  <Column id=severity title="Severity" contentType=colorscale colorScale=negative/>
  <Column id=category title="Category"/>
  <Column id=file title="File" wrap=true/>
  <Column id=description title="Description" wrap=true/>
  <Column id=first_commit_date title="First Committed"/>
  <Column id=still_exists title="Still Exists"/>
</DataTable>

### Purge Recommendations

Files that should be removed from git history using BFG or git-filter-repo:

<DataTable
  data={purge_recommendations}
  search=true
  rows=25
  rowShading=true
>
  <Column id=priority title="Priority"/>
  <Column id=severity title="Severity" contentType=colorscale colorScale=negative/>
  <Column id=file title="File" wrap=true/>
  <Column id=reason title="Reason" wrap=true/>
  <Column id=command title="Purge Command" wrap=true/>
  <Column id=affected_commits title="Commits"/>
</DataTable>

{:else if git_history_summary[0].note}

<Alert status="info">{git_history_summary[0].note}</Alert>

{:else}

<Alert status="positive">No sensitive files found in git history.</Alert>

{/if}

---

<Grid cols=2>
  <BigLink url="/">Back to Dashboard</BigLink>
  <BigLink url="/dependencies">Dependencies & SBOM</BigLink>
</Grid>
