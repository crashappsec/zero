---
title: Security Report
---

```sql metadata
select * from zero.metadata
```

# {metadata[0].repository}

<Alert status="info">
Last scanned: {new Date(metadata[0].timestamp).toLocaleString()} | Scanners: {metadata[0].scanners_run}
</Alert>

## Security Overview

```sql severity_counts
select * from zero.severity_counts
```

<BigValue
  data={severity_counts}
  value=critical
  title="Critical"
  fmt="num0"
/>

<BigValue
  data={severity_counts}
  value=high
  title="High"
  fmt="num0"
/>

<BigValue
  data={severity_counts}
  value=medium
  title="Medium"
  fmt="num0"
/>

<BigValue
  data={severity_counts}
  value=low
  title="Low"
  fmt="num0"
/>

<BigValue
  data={severity_counts}
  value=total
  title="Total Findings"
  fmt="num0"
/>

## Findings by Scanner

```sql findings_by_scanner
select * from zero.findings_by_scanner
```

<BarChart
  data={findings_by_scanner}
  x=scanner
  y=count
  series=severity
  type=stacked
  colorPalette={['#dc2626','#ea580c','#ca8a04','#22c55e']}
  title="Findings Distribution"
/>

## Scanner Results

```sql scanner_summary
select * from zero.scanner_summary
```

<DataTable
  data={scanner_summary}
  rows=15
>
  <Column id=scanner title="Scanner"/>
  <Column id=status title="Status"/>
  <Column id=findings title="Findings" fmt="num0"/>
  <Column id=summary title="Summary"/>
</DataTable>

---

<Grid cols=2>
  <BigLink href="/security">
    Security Findings
  </BigLink>
  <BigLink href="/dependencies">
    Dependencies & SBOM
  </BigLink>
  <BigLink href="/devops">
    DevOps & Infrastructure
  </BigLink>
  <BigLink href="/quality">
    Code Quality
  </BigLink>
</Grid>
