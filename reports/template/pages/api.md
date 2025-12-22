---
title: API Analysis
sidebar_position: 7
---

# API Analysis

<Alert status="info">
API security and quality analysis - REST, GraphQL, and OpenAPI validation.
</Alert>

```sql api_findings
select * from zero.api_findings
order by
  case severity
    when 'critical' then 1
    when 'high' then 2
    when 'medium' then 3
    when 'low' then 4
    else 5
  end
```

```sql api_summary
select * from zero.api_summary
```

## Overview

<Grid cols=4>
<BigValue
  data={api_summary}
  value=total_findings
  title="Total Issues"
/>
<BigValue
  data={api_summary}
  value=critical
  title="Critical"
/>
<BigValue
  data={api_summary}
  value=high
  title="High"
/>
<BigValue
  data={api_summary}
  value=endpoints_found
  title="Endpoints Found"
/>
</Grid>

---

## Security Findings

```sql security_findings
select * from zero.api_findings
where category like 'api-auth%'
   or category like 'api-injection%'
   or category like 'api-ssrf%'
   or category like 'api-data%'
   or category like 'api-rate%'
   or category like 'api-mass%'
   or category = 'api-misconfiguration'
order by
  case severity
    when 'critical' then 1
    when 'high' then 2
    when 'medium' then 3
    when 'low' then 4
    else 5
  end
```

{#if security_findings.length > 0}

<Alert status="warning">
{security_findings.length} API security issues detected.
</Alert>

```sql security_by_category
select category, count(*) as count from zero.api_findings
where category like 'api-auth%'
   or category like 'api-injection%'
   or category like 'api-ssrf%'
   or category like 'api-data%'
   or category like 'api-rate%'
   or category like 'api-mass%'
   or category = 'api-misconfiguration'
group by category order by count desc
```

<BarChart
  data={security_by_category}
  x=category
  y=count
  swapXY=true
  title="Security Issues by Category"
/>

<DataTable
  data={security_findings}
  search=true
  rows=20
  rowShading=true
>
  <Column id=severity title="Severity"/>
  <Column id=category title="Category"/>
  <Column id=title title="Issue"/>
  <Column id=file title="File"/>
  <Column id=line title="Line" fmt=num0/>
  <Column id=owasp_api title="OWASP API"/>
  <Column id=remediation title="Remediation" wrap=true/>
</DataTable>

{:else}

<Alert status="positive">
No API security issues detected.
</Alert>

{/if}

---

## Quality Findings

```sql quality_findings
select * from zero.api_findings
where category like 'api-design%'
   or category like 'api-performance%'
   or category like 'api-observability%'
   or category like 'api-documentation%'
order by
  case severity
    when 'high' then 1
    when 'medium' then 2
    when 'low' then 3
    else 4
  end
```

{#if quality_findings.length > 0}

<Alert status="info">
{quality_findings.length} API quality recommendations found.
</Alert>

```sql quality_by_category
select category, count(*) as count from zero.api_findings
where category like 'api-design%'
   or category like 'api-performance%'
   or category like 'api-observability%'
   or category like 'api-documentation%'
group by category order by count desc
```

<BarChart
  data={quality_by_category}
  x=category
  y=count
  swapXY=true
  title="Quality Issues by Category"
/>

<DataTable
  data={quality_findings}
  search=true
  rows=20
  rowShading=true
>
  <Column id=severity title="Severity"/>
  <Column id=category title="Category"/>
  <Column id=title title="Issue"/>
  <Column id=file title="File"/>
  <Column id=line title="Line" fmt=num0/>
  <Column id=remediation title="Recommendation" wrap=true/>
</DataTable>

{:else}

<Alert status="positive">
No API quality issues found.
</Alert>

{/if}

---

## GraphQL

```sql graphql_findings
select * from zero.api_findings
where framework = 'graphql'
   or category like '%graphql%'
   or endpoint = '/graphql'
order by
  case severity
    when 'critical' then 1
    when 'high' then 2
    when 'medium' then 3
    when 'low' then 4
    else 5
  end
```

{#if graphql_findings.length > 0}

<Alert status="warning">
{graphql_findings.length} GraphQL issues detected.
</Alert>

<DataTable
  data={graphql_findings}
  search=true
  rows=15
  rowShading=true
>
  <Column id=severity title="Severity"/>
  <Column id=title title="Issue"/>
  <Column id=file title="File"/>
  <Column id=line title="Line" fmt=num0/>
  <Column id=remediation title="Remediation" wrap=true/>
</DataTable>

{:else}

<Alert status="info">
No GraphQL issues detected (or GraphQL not used).
</Alert>

{/if}

---

## OpenAPI/Swagger

```sql openapi_findings
select * from zero.api_findings
where category = 'api-misconfiguration'
   or title like '%OpenAPI%'
   or title like '%Swagger%'
   or file like '%openapi%'
   or file like '%swagger%'
order by
  case severity
    when 'critical' then 1
    when 'high' then 2
    when 'medium' then 3
    when 'low' then 4
    else 5
  end
```

{#if openapi_findings.length > 0}

<Alert status="warning">
{openapi_findings.length} OpenAPI specification issues found.
</Alert>

<DataTable
  data={openapi_findings}
  search=true
  rows=15
  rowShading=true
>
  <Column id=severity title="Severity"/>
  <Column id=title title="Issue"/>
  <Column id=file title="Spec File"/>
  <Column id=endpoint title="Endpoint"/>
  <Column id=http_method title="Method"/>
  <Column id=remediation title="Remediation" wrap=true/>
</DataTable>

{:else}

<Alert status="info">
No OpenAPI specification issues found.
</Alert>

{/if}

---

## By Framework

```sql by_framework
select framework, count(*) as count from zero.api_findings
where framework != ''
group by framework order by count desc
```

{#if by_framework.length > 0}

<BarChart
  data={by_framework}
  x=framework
  y=count
  swapXY=true
  title="Issues by Framework"
/>

{/if}

---

<Grid cols=3>
  <BigLink url="/">Back to Dashboard</BigLink>
  <BigLink url="/security">Security</BigLink>
  <BigLink url="/quality">Code Quality</BigLink>
</Grid>
