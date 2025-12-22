---
title: DevOps
sidebar_position: 4
---

# DevOps

```sql dora_metrics
select * from zero.dora_metrics
```

```sql iac_findings
select * from zero.iac_findings
```

```sql github_actions_findings
select * from zero.github_actions_findings
```

```sql container_findings
select * from zero.container_findings
```

## DevOps Overview

<Grid cols=4>
<BigValue
  data={iac_findings}
  value={iac_findings.length}
  title="IaC Issues"
/>
<BigValue
  data={github_actions_findings}
  value={github_actions_findings.length}
  title="CI/CD Issues"
/>
<BigValue
  data={container_findings}
  value={container_findings.length}
  title="Container Issues"
/>
<BigValue
  data={dora_metrics}
  value=overall_class
  title="DORA Performance"
/>
</Grid>

---

## DORA Metrics

{#if dora_metrics.length > 0 && dora_metrics[0].deployment_frequency_class}

<Alert status="info">
DORA metrics measure software delivery performance across four key dimensions.
</Alert>

<Grid cols=4>
<BigValue
  data={dora_metrics}
  value=deployment_frequency_class
  title="Deploy Frequency"
/>
<BigValue
  data={dora_metrics}
  value=lead_time_class
  title="Lead Time"
/>
<BigValue
  data={dora_metrics}
  value=change_failure_class
  title="Change Failure Rate"
/>
<BigValue
  data={dora_metrics}
  value=mttr_class
  title="MTTR"
/>
</Grid>

{:else}

<Alert status="info">DORA metrics not available. Enable the dora feature in the devops scanner.</Alert>

{/if}

---

## Infrastructure as Code

{#if iac_findings.length > 0}

```sql iac_by_severity
select
  severity,
  count(*) as count
from zero.iac_findings
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

<Grid cols=2>
<BarChart
  data={iac_by_severity}
  x=severity
  y=count
  colorPalette={['#dc2626','#ea580c','#ca8a04','#22c55e']}
  title="IaC Findings by Severity"
/>

```sql iac_by_type
select type, count(*) as count from zero.iac_findings group by type order by count desc limit 10
```

<BarChart
  data={iac_by_type}
  x=type
  y=count
  swapXY=true
  title="IaC Findings by Type"
/>
</Grid>

<DataTable
  data={iac_findings}
  search=true
  rows=25
  rowShading=true
>
  <Column id=severity title="Severity" contentType=colorscale colorScale=red/>
  <Column id=type title="Type"/>
  <Column id=rule_id title="Rule"/>
  <Column id=title title="Title" wrap=true/>
  <Column id=file title="File"/>
  <Column id=resolution title="Resolution" wrap=true/>
</DataTable>

{:else}

<Alert status="success">No Infrastructure as Code issues detected.</Alert>

{/if}

---

## GitHub Actions Security

{#if github_actions_findings.length > 0}

```sql actions_by_category
select category, count(*) as count from zero.github_actions_findings group by category order by count desc
```

<BarChart
  data={actions_by_category}
  x=category
  y=count
  swapXY=true
  title="CI/CD Issues by Category"
/>

<DataTable
  data={github_actions_findings}
  search=true
  rows=25
  rowShading=true
>
  <Column id=severity title="Severity" contentType=colorscale colorScale=red/>
  <Column id=category title="Category"/>
  <Column id=workflow title="Workflow"/>
  <Column id=job title="Job"/>
  <Column id=description title="Description" wrap=true/>
</DataTable>

{:else}

<Alert status="success">No GitHub Actions issues detected.</Alert>

{/if}

---

## Container Security

{#if container_findings.length > 0}

<DataTable
  data={container_findings}
  search=true
  rows=25
  rowShading=true
>
  <Column id=severity title="Severity" contentType=colorscale colorScale=red/>
  <Column id=image title="Image"/>
  <Column id=rule_id title="Rule"/>
  <Column id=title title="Title" wrap=true/>
  <Column id=resolution title="Resolution" wrap=true/>
</DataTable>

{:else}

<Alert status="success">No container security issues detected.</Alert>

{/if}

---

<Grid cols=2>
  <BigLink url="/">Back to Dashboard</BigLink>
  <BigLink url="/quality">Code Quality</BigLink>
</Grid>
