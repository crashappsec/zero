---
title: AI/ML Security
---

# AI/ML Security

<Alert status="info">
Machine learning model security, AI framework analysis, and ML-BOM (Machine Learning Bill of Materials).
</Alert>

## AI Security Findings

```sql ai_security
select * from zero.ai_security
where severity != 'none' and title != ''
order by
  case severity
    when 'critical' then 1
    when 'high' then 2
    when 'medium' then 3
    when 'low' then 4
    else 5
  end
```

{#if ai_security.length > 0}

<Alert status="warning">
{ai_security.length} AI/ML security issues detected.
</Alert>

<DataTable
  data={ai_security}
  search=true
  rows=20
>
  <Column id=severity title="Severity"/>
  <Column id=category title="Category"/>
  <Column id=title title="Issue"/>
  <Column id=file title="File"/>
  <Column id=description title="Description"/>
</DataTable>

{:else}

<Alert status="positive">
No AI/ML security issues detected.
</Alert>

{/if}

## ML Models Detected

```sql ml_models
select * from zero.ml_models where type != 'none' and name != ''
```

{#if ml_models.length > 0}

<Alert status="info">
{ml_models.length} ML models found in the repository.
</Alert>

<DataTable
  data={ml_models}
  search=true
  rows=15
>
  <Column id=name title="Model"/>
  <Column id=type title="Type"/>
  <Column id=format title="Format"/>
  <Column id=risk title="Risk Level"/>
</DataTable>

{:else}

<Alert status="info">
No ML models detected in this repository.
</Alert>

{/if}

## ML Frameworks

```sql ml_frameworks
select * from zero.ml_frameworks where category != 'none' and name != ''
```

{#if ml_frameworks.length > 0}

<DataTable
  data={ml_frameworks}
  rows=10
>
  <Column id=name title="Framework"/>
  <Column id=version title="Version"/>
  <Column id=category title="Category"/>
</DataTable>

{:else}

<Alert status="info">
No ML frameworks detected.
</Alert>

{/if}

## Technologies

```sql technologies
select * from zero.technologies
```

{#if technologies.length > 0}

<DataTable
  data={technologies}
  search=true
  rows=15
>
  <Column id=name title="Technology"/>
  <Column id=category title="Category"/>
  <Column id=confidence title="Confidence"/>
  <Column id=version title="Version"/>
</DataTable>

{/if}

---

<ButtonGroup>
  <BigLink url="/">Back to Overview</BigLink>
  <BigLink url="/security">Security</BigLink>
  <BigLink url="/quality">Code Quality</BigLink>
</ButtonGroup>
