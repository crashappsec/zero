---
title: Dependencies
sidebar_position: 2
---

# Dependencies

```sql sbom_packages
select * from zero.sbom_packages where name != ''
```

```sql ecosystem_summary
select * from zero.ecosystem_summary where ecosystem != 'none'
```

```sql licenses
select * from zero.licenses
```

```sql vulnerabilities
select * from zero.vulnerabilities where source = 'Package'
```

## Package Overview

<Grid cols=3>
<BigValue
  data={sbom_packages}
  value={sbom_packages.length}
  title="Total Packages"
/>
<BigValue
  data={ecosystem_summary}
  value={ecosystem_summary.length}
  title="Ecosystems"
/>
<BigValue
  data={vulnerabilities}
  value={vulnerabilities.length}
  title="Vulnerable Packages"
/>
</Grid>

---

## Ecosystem Distribution

{#if ecosystem_summary.length > 0}

<BarChart
  data={ecosystem_summary}
  x=ecosystem
  y=count
  swapXY=true
  title="Packages by Ecosystem"
/>

{:else}
<Alert status="info">No ecosystem data available.</Alert>
{/if}

---

## Vulnerable Packages

{#if vulnerabilities.length > 0}

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

<BarChart
  data={vuln_by_severity}
  x=severity
  y=count
  colorPalette={['#dc2626','#ea580c','#ca8a04','#22c55e']}
  title="Vulnerable Packages by Severity"
/>

<DataTable
  data={vulnerabilities}
  search=true
  rows=25
  rowShading=true
>
  <Column id=package title="Package"/>
  <Column id=severity title="Severity" contentType=colorscale colorScale=negative/>
  <Column id=cve title="CVE"/>
  <Column id=title title="Description" wrap=true/>
  <Column id=fix_version title="Fix Available"/>
</DataTable>

{:else}

<Alert status="positive">No vulnerable packages detected.</Alert>

{/if}

---

## License Distribution

{#if licenses.length > 0}

<Grid cols=3>
<BarChart
  data={licenses}
  x=license
  y=count
  swapXY=true
  title="License Distribution"
/>

<DataTable
  data={licenses}
  rows=15
  rowShading=true
>
  <Column id=license title="License"/>
  <Column id=count title="Packages" fmt=num0/>
</DataTable>
</Grid>

{:else}
<Alert status="info">No license data available.</Alert>
{/if}

---

## All Packages

{#if sbom_packages.length > 0}

<DataTable
  data={sbom_packages}
  search=true
  rows=50
  rowShading=true
>
  <Column id=name title="Package"/>
  <Column id=version title="Version"/>
  <Column id=ecosystem title="Ecosystem"/>
  <Column id=license title="License"/>
</DataTable>

{:else}
<Alert status="info">No package data available.</Alert>
{/if}

---

<Grid cols=3>
  <BigLink url="/">Back to Dashboard</BigLink>
  <BigLink url="/supply-chain">Supply Chain</BigLink>
  <BigLink url="/security">Security</BigLink>
</Grid>
