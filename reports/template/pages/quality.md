---
title: Code Quality & Ownership
---

# Code Quality & Ownership

<Alert status="info">
Code quality metrics, technology stack, and ownership analysis.
</Alert>

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
  <BigLink href="/">Back to Overview</BigLink>
  <BigLink href="/devops">DevOps</BigLink>
  <BigLink href="/security">Security</BigLink>
</ButtonGroup>
