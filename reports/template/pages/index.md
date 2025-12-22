---
title: Engineering Insights
---

```sql org_summary
select * from zero.org_summary
```

```sql repos
select * from zero.repos order by critical desc, vulns desc
```

```sql severity_counts
select * from zero.severity_counts
```

```sql metadata
select * from zero.metadata
```

{#if org_summary[0].is_org_mode === true}

# Organization Overview

<Alert status="info">
Engineering insights across <b>{org_summary[0].total_repos}</b> repositories
</Alert>

## Overview

<Grid cols=6>
<BigValue
  data={org_summary}
  value=total_repos
  title="Repositories"
/>
<BigValue
  data={org_summary}
  value=critical_vulns
  title="Critical"
  comparison=repos_with_critical
  comparisonTitle="repos affected"
/>
<BigValue
  data={org_summary}
  value=high_vulns
  title="High"
/>
<BigValue
  data={org_summary}
  value=medium_vulns
  title="Medium"
/>
<BigValue
  data={org_summary}
  value=low_vulns
  title="Low"
/>
<BigValue
  data={org_summary}
  value=total_secrets
  title="Secrets"
  comparison=repos_with_secrets
  comparisonTitle="repos affected"
/>
</Grid>

## Severity Distribution

```sql severity_pie
select 'Critical' as severity, critical_vulns as count from zero.org_summary where critical_vulns > 0
union all
select 'High' as severity, high_vulns as count from zero.org_summary where high_vulns > 0
union all
select 'Medium' as severity, medium_vulns as count from zero.org_summary where medium_vulns > 0
union all
select 'Low' as severity, low_vulns as count from zero.org_summary where low_vulns > 0
```

{#if severity_pie.length > 0}
<BarChart
  data={severity_pie}
  x=severity
  y=count
  colorPalette={['#dc2626','#ea580c','#ca8a04','#22c55e']}
  swapXY=true
  title="Vulnerabilities by Severity"
/>
{:else}
<Alert status="success">No vulnerabilities detected across the organization.</Alert>
{/if}

## Repository Summary

<DataTable
  data={repos}
  rows=all
  rowShading=true
  search=true
>
  <Column id=name title="Repository"/>
  <Column id=packages title="Packages" fmt=num0/>
  <Column id=vulns title="Total Vulns" fmt=num0/>
  <Column id=critical title="Critical" fmt=num0/>
  <Column id=high title="High" fmt=num0/>
  <Column id=secrets title="Secrets" fmt=num0/>
</DataTable>

{:else}

# {metadata[0].repository}

<Alert status="info">
Last scanned: {new Date(metadata[0].timestamp).toLocaleString()}
</Alert>

## Overview

<Grid cols=5>
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
<BigValue
  data={severity_counts}
  value=total
  title="Total"
/>
</Grid>

## Findings by Scanner

```sql findings_by_scanner
select * from zero.findings_by_scanner
```

{#if findings_by_scanner.length > 0}
<BarChart
  data={findings_by_scanner}
  x=scanner
  y=count
  series=severity
  type=stacked
  colorPalette={['#dc2626','#ea580c','#ca8a04','#22c55e']}
  title="Findings Distribution"
/>
{:else}
<Alert status="success">No findings detected.</Alert>
{/if}

## Scanner Results

```sql scanner_summary
select * from zero.scanner_summary
```

<DataTable data={scanner_summary} rows=15>
  <Column id=scanner title="Scanner"/>
  <Column id=status title="Status"/>
  <Column id=findings title="Findings" fmt=num0/>
  <Column id=summary title="Summary"/>
</DataTable>

{/if}

---

## Explore

<Grid cols=3>
  <BigLink url="/security">Security</BigLink>
  <BigLink url="/dependencies">Dependencies</BigLink>
  <BigLink url="/supply-chain">Supply Chain</BigLink>
  <BigLink url="/devops">DevOps</BigLink>
  <BigLink url="/quality">Quality & Ownership</BigLink>
  <BigLink url="/ai-security">AI/ML</BigLink>
</Grid>
