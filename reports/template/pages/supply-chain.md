---
title: Supply Chain Security
---

# Supply Chain Security

<Alert status="warning">
Supply chain analysis including malicious package detection, dependency health, and SBOM integrity.
</Alert>

## Package Overview

```sql sbom_packages
select * from zero.sbom_packages limit 100
```

```sql ecosystem_summary
select * from zero.ecosystem_summary
```

{#if ecosystem_summary.length > 0}

<BarChart
  data={ecosystem_summary}
  x=ecosystem
  y=count
  title="Packages by Ecosystem"
/>

{/if}

## Malicious Package Detection

```sql malcontent
select * from zero.malcontent
where severity != 'none' and package != ''
order by
  case severity
    when 'critical' then 1
    when 'high' then 2
    when 'medium' then 3
    when 'low' then 4
    else 5
  end
```

{#if malcontent.length > 0}

<Alert status="error">
{malcontent.length} suspicious behaviors detected in dependencies. Review these carefully.
</Alert>

<DataTable
  data={malcontent}
  search=true
  rows=20
>
  <Column id=severity title="Severity"/>
  <Column id=package title="Package"/>
  <Column id=category title="Category"/>
  <Column id=rule title="Rule"/>
  <Column id=description title="Description"/>
</DataTable>

{:else}

<Alert status="positive">
No malicious package behaviors detected.
</Alert>

{/if}

## Package Health

```sql package_health
select * from zero.package_health
where ecosystem != 'none' and health_score >= 0
order by health_score asc limit 20
```

{#if package_health.length > 0}

<Alert status="info">
Showing packages with lowest health scores. Consider updating or replacing unmaintained packages.
</Alert>

<DataTable
  data={package_health}
  search=true
  rows=20
>
  <Column id=package title="Package"/>
  <Column id=version title="Version"/>
  <Column id=ecosystem title="Ecosystem"/>
  <Column id=health_score title="Health Score"/>
  <Column id=deprecated title="Deprecated"/>
  <Column id=unmaintained title="Unmaintained"/>
</DataTable>

{:else}

<Alert status="info">
No package health data available.
</Alert>

{/if}

## License Compliance

```sql licenses
select * from zero.licenses
```

{#if licenses.length > 0}

<BarChart
  data={licenses}
  x=license
  y=count
  swapXY=true
  title="License Distribution"
/>

{/if}

---

<ButtonGroup>
  <BigLink url="/">Back to Overview</BigLink>
  <BigLink url="/dependencies">Dependencies</BigLink>
  <BigLink url="/security">Security</BigLink>
</ButtonGroup>
