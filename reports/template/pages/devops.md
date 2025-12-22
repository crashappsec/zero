---
title: DevOps & Infrastructure
---

# DevOps & Infrastructure

<Alert status="info">
Infrastructure as Code, CI/CD, and container security analysis.
</Alert>

## DORA Metrics

```sql dora_metrics
select * from zero.dora_metrics
```

{#if dora_metrics.length > 0}

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

<BigValue
  data={dora_metrics}
  value=overall_class
  title="Overall Performance"
/>

{:else}

<Alert status="info">
DORA metrics not available.
</Alert>

{/if}

## Infrastructure as Code

```sql iac_findings
select * from zero.iac_findings
```

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

<BarChart
  data={iac_by_severity}
  x=severity
  y=count
  colorPalette={['#dc2626','#ea580c','#ca8a04','#22c55e']}
  title="IaC Findings by Severity"
/>

<DataTable
  data={iac_findings}
  search=true
  rows=20
>
  <Column id=severity title="Severity"/>
  <Column id=type title="Type"/>
  <Column id=rule_id title="Rule"/>
  <Column id=title title="Title"/>
  <Column id=file title="File"/>
  <Column id=resolution title="Resolution"/>
</DataTable>

{:else}

<Alert status="positive">
No IaC issues detected.
</Alert>

{/if}

## GitHub Actions

```sql github_actions_findings
select * from zero.github_actions_findings
```

{#if github_actions_findings.length > 0}

<DataTable
  data={github_actions_findings}
  search=true
  rows=15
>
  <Column id=severity title="Severity"/>
  <Column id=category title="Category"/>
  <Column id=workflow title="Workflow"/>
  <Column id=job title="Job"/>
  <Column id=description title="Description"/>
</DataTable>

{:else}

<Alert status="positive">
No GitHub Actions issues detected.
</Alert>

{/if}

## Container Security

```sql container_findings
select * from zero.container_findings
```

{#if container_findings.length > 0}

<DataTable
  data={container_findings}
  search=true
  rows=15
>
  <Column id=severity title="Severity"/>
  <Column id=image title="Image"/>
  <Column id=rule_id title="Rule"/>
  <Column id=title title="Title"/>
  <Column id=resolution title="Resolution"/>
</DataTable>

{:else}

<Alert status="positive">
No container issues detected.
</Alert>

{/if}

---

<ButtonGroup>
  <BigLink url="/">Back to Overview</BigLink>
  <BigLink url="/dependencies">Dependencies</BigLink>
  <BigLink url="/quality">Code Quality</BigLink>
</ButtonGroup>
