---
title: Quality
sidebar_position: 5
---

# Code Quality

<Alert status="info">
Code health metrics including tech debt, complexity, test coverage, and documentation.
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

---

<Grid cols=3>
  <BigLink url="/">Back to Dashboard</BigLink>
  <BigLink url="/team">Team & Ownership</BigLink>
  <BigLink url="/ai-security">Technology</BigLink>
</Grid>
