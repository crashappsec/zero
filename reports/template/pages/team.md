---
title: Team
sidebar_position: 6
---

# Team & Ownership

<Alert status="info">
Code ownership analysis, contributor health, and team productivity metrics.
</Alert>

## Ownership Health

```sql ownership_summary
select * from zero.ownership_summary
```

{#if ownership_summary.length > 0}

<Grid cols=4>
<BigValue
  data={ownership_summary}
  value=bus_factor
  title="Bus Factor"
  fmt="num0"
/>
<BigValue
  data={ownership_summary}
  value=bus_factor_risk
  title="Risk Level"
/>
<BigValue
  data={ownership_summary}
  value=total_contributors
  title="All-Time Contributors"
  fmt="num0"
/>
<BigValue
  data={ownership_summary}
  value=active_contributors_90d
  title="Active (90 days)"
  fmt="num0"
/>
</Grid>

### What is Bus Factor?

The **bus factor** is the minimum number of team members that would need to leave before the project becomes unmaintainable. A bus factor of 1 means a single person leaving could critically impact the project.

| Bus Factor | Risk Level | Recommendation |
|------------|------------|----------------|
| 1 | 🔴 Critical | Immediate knowledge sharing needed |
| 2 | 🟠 High | Encourage pair programming and documentation |
| 3-4 | 🟡 Medium | Good, but could improve |
| 5+ | 🟢 Low | Healthy distribution of knowledge |

{:else}

<Alert status="info">
No ownership data available. Run with code-ownership scanner enabled.
</Alert>

{/if}

---

## Top Contributors

```sql contributors
select * from zero.contributors
```

{#if contributors.length > 0}

<DataTable
  data={contributors}
  rows=10
  rowShading=true
>
  <Column id=name title="Contributor"/>
  <Column id=email title="Email"/>
  <Column id=commits title="Commits" fmt="num0"/>
  <Column id=lines_added title="Lines Added" fmt="num0"/>
  <Column id=lines_removed title="Lines Removed" fmt="num0"/>
  <Column id=contribution_pct title="Contribution %" fmt="pct1"/>
</DataTable>

{:else}

<Alert status="info">
No contributor data available.
</Alert>

{/if}

---

## Activity Trends

```sql activity_by_month
select * from zero.activity_by_month order by month desc limit 12
```

{#if activity_by_month.length > 0}

<LineChart
  data={activity_by_month}
  x=month
  y=commits
  title="Commit Activity (Last 12 Months)"
/>

{:else}

<Alert status="info">
No activity trend data available.
</Alert>

{/if}

---

## CODEOWNERS Analysis

```sql codeowners
select * from zero.codeowners
```

{#if codeowners.length > 0}

<DataTable
  data={codeowners}
  rows=20
  search=true
>
  <Column id=path title="Path Pattern"/>
  <Column id=owners title="Owners"/>
  <Column id=coverage title="Coverage"/>
</DataTable>

{:else}

<Alert status="info">
No CODEOWNERS file found. Consider adding one to define code ownership explicitly.
</Alert>

{/if}

---

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
  <BigLink url="/quality">Quality</BigLink>
  <BigLink url="/devops">DevOps</BigLink>
</Grid>
