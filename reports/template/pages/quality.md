---
title: Code Quality & Ownership
---

# Code Quality & Ownership

<Alert status="info">
Code quality metrics, developer experience, technology stack, and ownership analysis.
</Alert>

## Quality Metrics

```sql code_quality
select * from zero.code_quality
```

{#if code_quality.length > 0 && code_quality[0].metric !== 'No data'}

<DataTable
  data={code_quality}
  rows=10
>
  <Column id=metric title="Metric"/>
  <Column id=value title="Value"/>
  <Column id=rating title="Rating"/>
</DataTable>

{:else}

<Alert status="info">
No code quality metrics available.
</Alert>

{/if}

## Developer Experience

```sql devx_metrics
select * from zero.devx_metrics
```

{#if devx_metrics.length > 0 && devx_metrics[0].metric !== 'No data'}

<DataTable
  data={devx_metrics}
  rows=10
>
  <Column id=category title="Category"/>
  <Column id=metric title="Metric"/>
  <Column id=value title="Value"/>
  <Column id=status title="Status"/>
</DataTable>

{:else}

<Alert status="info">
No developer experience metrics available.
</Alert>

{/if}

## Technology Stack

```sql technologies
select * from zero.technologies
```

{#if technologies.length > 0}

<DataTable
  data={technologies}
  search=true
  rows=20
>
  <Column id=name title="Technology"/>
  <Column id=category title="Category"/>
  <Column id=confidence title="Confidence"/>
  <Column id=version title="Version"/>
</DataTable>

{:else}

<Alert status="info">
No technologies detected.
</Alert>

{/if}

## Code Ownership

```sql ownership_summary
select * from zero.ownership_summary
```

{#if ownership_summary.length > 0}

<BigValue
  data={ownership_summary}
  value=bus_factor
  title="Bus Factor"
  fmt="num0"
/>

<BigValue
  data={ownership_summary}
  value=total_contributors
  title="Total Contributors"
  fmt="num0"
/>

<BigValue
  data={ownership_summary}
  value=active_contributors_90d
  title="Active (90 days)"
  fmt="num0"
/>

{/if}

## Top Contributors

```sql contributors
select * from zero.contributors
```

{#if contributors.length > 0}

<DataTable
  data={contributors}
  rows=10
>
  <Column id=name title="Contributor"/>
  <Column id=email title="Email"/>
  <Column id=commits title="Commits" fmt="num0"/>
  <Column id=lines_added title="Lines Added" fmt="num0"/>
  <Column id=lines_removed title="Lines Removed" fmt="num0"/>
</DataTable>

{:else}

<Alert status="info">
No contributor data available.
</Alert>

{/if}

---

<ButtonGroup>
  <BigLink url="/">Back to Overview</BigLink>
  <BigLink url="/devops">DevOps</BigLink>
  <BigLink url="/security">Security</BigLink>
</ButtonGroup>
