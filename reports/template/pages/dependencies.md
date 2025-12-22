---
title: Dependencies & SBOM
---

# Dependencies & SBOM

<Alert status="info">
Software Bill of Materials and dependency analysis results.
</Alert>

## Vulnerability Summary

```sql vulnerabilities
select * from zero.vulnerabilities where source = 'Package'
```

```sql vuln_by_severity
select
  severity,
  count(*) as count
from zero.vulnerabilities
where source = 'Package'
group by severity
order by
  case severity
    when 'critical' then 1
    when 'high' then 2
    when 'medium' then 3
    when 'low' then 4
    else 5
  end
```

{#if vuln_by_severity.length > 0}

<BarChart
  data={vuln_by_severity}
  x=severity
  y=count
  colorPalette={['#dc2626','#ea580c','#ca8a04','#22c55e']}
  title="Vulnerabilities by Severity"
/>

{/if}

## Vulnerable Packages

{#if vulnerabilities.length > 0}

<DataTable
  data={vulnerabilities}
  search=true
  rows=25
>
  <Column id=package title="Package"/>
  <Column id=version title="Version"/>
  <Column id=severity title="Severity"/>
  <Column id=cve title="CVE"/>
  <Column id=title title="Description"/>
  <Column id=fix_version title="Fix Available"/>
</DataTable>

{:else}

<Alert status="positive">
No vulnerable packages detected.
</Alert>

{/if}

## License Distribution

```sql licenses
select * from zero.licenses
```

{#if licenses.length > 0}

<BarChart
  data={licenses}
  x=license
  y=count
  swapXY=true
  title="Top 10 Licenses"
/>

<DataTable
  data={licenses}
  rows=15
>
  <Column id=license title="License"/>
  <Column id=count title="Packages" fmt="num0"/>
</DataTable>

{:else}

<Alert status="info">
No license data available.
</Alert>

{/if}

---

<ButtonGroup>
  <BigLink url="/">Back to Overview</BigLink>
  <BigLink url="/security">Security</BigLink>
  <BigLink url="/devops">DevOps</BigLink>
</ButtonGroup>
