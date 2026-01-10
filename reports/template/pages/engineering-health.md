---
title: Engineering Health
sidebar_position: 1
---

# Engineering Health Dashboard

<Alert status="info">
A holistic view of your repository's engineering health across all six dimensions.
</Alert>

```sql severity_counts
select * from zero.severity_counts
```

```sql metadata
select * from zero.metadata
```

```sql ownership_summary
select * from zero.ownership_summary
```

```sql scanner_summary
select * from zero.scanner_summary
```

## Health Score Overview

<Grid cols=6>
<BigValue
  data={severity_counts}
  value=critical
  title="Critical Issues"
/>
<BigValue
  data={severity_counts}
  value=high
  title="High Issues"
/>
<BigValue
  data={severity_counts}
  value=medium
  title="Medium Issues"
/>
<BigValue
  data={severity_counts}
  value=low
  title="Low Issues"
/>
<BigValue
  data={ownership_summary}
  value=bus_factor
  title="Bus Factor"
  fmt="num0"
/>
<BigValue
  data={ownership_summary}
  value=active_contributors_90d
  title="Active Contributors"
  fmt="num0"
/>
</Grid>

---

## Six Dimensions of Engineering Health

### 🔒 Security

```sql security_summary
select
  count(*) as total_findings,
  sum(case when severity = 'critical' then 1 else 0 end) as critical,
  sum(case when severity = 'high' then 1 else 0 end) as high
from zero.vulnerabilities
where source = 'Code'
```

{#if security_summary[0].total_findings > 0}
<Alert status="warning">
{security_summary[0].total_findings} code security findings ({security_summary[0].critical} critical, {security_summary[0].high} high)
</Alert>
{:else}
<Alert status="positive">No code security issues detected.</Alert>
{/if}

<BigLink url="/security">View Security Details →</BigLink>

---

### 📦 Supply Chain

```sql supply_chain_summary
select
  count(*) as total_vulns,
  sum(case when severity = 'critical' then 1 else 0 end) as critical,
  sum(case when severity = 'high' then 1 else 0 end) as high
from zero.vulnerabilities
where source = 'Package'
```

```sql package_count
select count(*) as total from zero.sbom_packages where name != ''
```

{#if supply_chain_summary[0].total_vulns > 0}
<Alert status="warning">
{supply_chain_summary[0].total_vulns} vulnerable packages ({supply_chain_summary[0].critical} critical, {supply_chain_summary[0].high} high)
</Alert>
{:else}
<Alert status="positive">No vulnerable packages detected in {package_count[0].total} dependencies.</Alert>
{/if}

<BigLink url="/supply-chain">View Supply Chain Details →</BigLink>

---

### 📊 Quality

```sql code_quality
select * from zero.code_quality limit 3
```

{#if code_quality.length > 0 && code_quality[0].metric !== 'No data'}
<DataTable data={code_quality} rows=3>
  <Column id=metric title="Metric"/>
  <Column id=value title="Value"/>
  <Column id=rating title="Rating"/>
</DataTable>
{:else}
<Alert status="info">No code quality metrics available.</Alert>
{/if}

<BigLink url="/quality">View Quality Details →</BigLink>

---

### ⚙️ DevOps

```sql dora_metrics
select * from zero.dora_metrics limit 1
```

{#if dora_metrics.length > 0 && dora_metrics[0].metric !== 'No data'}
<Grid cols=4>
<BigValue data={dora_metrics} value=deployment_frequency title="Deploy Frequency"/>
<BigValue data={dora_metrics} value=lead_time title="Lead Time"/>
<BigValue data={dora_metrics} value=change_failure_rate title="Change Failure Rate"/>
<BigValue data={dora_metrics} value=mttr title="MTTR"/>
</Grid>
{:else}
<Alert status="info">No DORA metrics available.</Alert>
{/if}

<BigLink url="/devops">View DevOps Details →</BigLink>

---

### 🤖 Technology

```sql tech_summary
select count(*) as total_tech from zero.technologies
```

```sql ml_summary
select count(*) as total_models from zero.ml_models where type != 'none' and name != ''
```

<Grid cols=2>
<BigValue data={tech_summary} value=total_tech title="Technologies Detected"/>
<BigValue data={ml_summary} value=total_models title="ML Models"/>
</Grid>

<BigLink url="/ai-security">View Technology Details →</BigLink>

---

### 👥 Team

{#if ownership_summary.length > 0}
<Grid cols=3>
<BigValue data={ownership_summary} value=bus_factor title="Bus Factor" fmt="num0"/>
<BigValue data={ownership_summary} value=total_contributors title="All-Time Contributors" fmt="num0"/>
<BigValue data={ownership_summary} value=active_contributors_90d title="Active (90 days)" fmt="num0"/>
</Grid>
{:else}
<Alert status="info">No ownership data available.</Alert>
{/if}

<BigLink url="/team">View Team Details →</BigLink>

---

## Scanner Status

<DataTable data={scanner_summary} rows=10>
  <Column id=scanner title="Scanner"/>
  <Column id=status title="Status"/>
  <Column id=findings title="Findings" fmt=num0/>
  <Column id=summary title="Summary"/>
</DataTable>

---

<Grid cols=3>
  <BigLink url="/">Overview</BigLink>
  <BigLink url="/security">Security</BigLink>
  <BigLink url="/supply-chain">Supply Chain</BigLink>
</Grid>
