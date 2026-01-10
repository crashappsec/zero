---
title: Supply Chain
sidebar_position: 3
---

# Supply Chain

```sql sbom_packages
select * from zero.sbom_packages where name != ''
```

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

```sql package_health
select * from zero.package_health
where package != '' and health_score >= 0
order by health_score asc
```

## Supply Chain Risk Overview

<Grid cols=4>
<BigValue
  data={sbom_packages}
  value={sbom_packages.length}
  title="Total Packages"
/>
<BigValue
  data={malcontent}
  value={malcontent.length}
  title="Suspicious Behaviors"
/>
<BigValue
  data={package_health}
  value={package_health.filter(p => p.deprecated === true).length}
  title="Deprecated"
/>
<BigValue
  data={package_health}
  value={package_health.filter(p => p.unmaintained === true).length}
  title="Unmaintained"
/>
</Grid>

---

## Malicious Package Detection

{#if malcontent.length > 0}

<Alert status="warning">
<b>{malcontent.length}</b> suspicious behaviors detected in dependencies. Review these carefully for potential supply chain attacks.
</Alert>

```sql malcontent_by_severity
select severity, count(*) as count from zero.malcontent
where severity != 'none' and package != ''
group by severity
order by case severity when 'critical' then 1 when 'high' then 2 when 'medium' then 3 else 4 end
```

<Grid cols=3>
<BarChart
  data={malcontent_by_severity}
  x=severity
  y=count
  colorPalette={['#dc2626','#ea580c','#ca8a04','#22c55e']}
  title="Suspicious Behaviors by Severity"
/>

```sql malcontent_by_category
select category, count(*) as count from zero.malcontent
where package != ''
group by category order by count desc limit 10
```

<BarChart
  data={malcontent_by_category}
  x=category
  y=count
  swapXY=true
  title="Behaviors by Category"
/>
</Grid>

<DataTable
  data={malcontent}
  search=true
  rows=25
  rowShading=true
>
  <Column id=severity title="Severity" contentType=colorscale colorScale=negative/>
  <Column id=package title="Package"/>
  <Column id=category title="Category"/>
  <Column id=rule title="Rule"/>
  <Column id=description title="Description" wrap=true/>
</DataTable>

{:else}

<Alert status="positive">No malicious package behaviors detected.</Alert>

{/if}

---

## Package Health

{#if package_health.length > 0 && package_health[0].package != ''}

<Alert status="info">
Packages sorted by health score (lowest first). Consider updating or replacing unmaintained packages.
</Alert>

<DataTable
  data={package_health}
  search=true
  rows=25
  rowShading=true
>
  <Column id=package title="Package"/>
  <Column id=version title="Version"/>
  <Column id=ecosystem title="Ecosystem"/>
  <Column id=health_score title="Health Score" fmt=num0 contentType=colorscale colorScale=green/>
  <Column id=deprecated title="Deprecated"/>
  <Column id=unmaintained title="Unmaintained"/>
</DataTable>

{:else}

<Alert status="info">No package health data available. Enable the health feature in the packages scanner.</Alert>

{/if}

---

<Grid cols=3>
  <BigLink url="/">Back to Dashboard</BigLink>
  <BigLink url="/dependencies">Dependencies</BigLink>
  <BigLink url="/security">Security</BigLink>
</Grid>
