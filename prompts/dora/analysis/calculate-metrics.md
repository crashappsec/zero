<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Calculate DORA Metrics

## Purpose

Calculate the four key DORA metrics from deployment data and classify performance level.

## Prompt Template

```
I need you to calculate DORA metrics from this deployment data.

Data Source: [GitHub Actions/GitLab CI/Jenkins/Manual entry]
Time Period: [Last week/month/quarter/custom dates]

Deployment Data:
[Paste JSON deployment data OR provide summary]

Please calculate:

1. Deployment Frequency
   - Total number of deployments
   - Average deploys per day/week
   - Performance classification (Elite/High/Medium/Low)

2. Lead Time for Changes
   - Median time from commit to production
   - Breakdown by phase (code review, CI, deployment)
   - Performance classification

3. Change Failure Rate
   - Number of failed deployments
   - Percentage of total deployments
   - Performance classification

4. Time to Restore Service
   - Median recovery time for incidents
   - Breakdown by phase (detection, diagnosis, resolution)
   - Performance classification

Also provide:
- Overall performance level
- Comparison to DORA research benchmarks
- Key strengths and areas for improvement
```

## Example Input

```json
{
  "period": "2024-11",
  "total_deployments": 45,
  "successful_deployments": 41,
  "failed_deployments": 4,
  "total_days": 30,
  "deployments": [
    {
      "date": "2024-11-01",
      "commit_time": "2024-11-01T09:00:00Z",
      "production_time": "2024-11-01T10:30:00Z",
      "status": "success"
    }
    // ... more deployments
  ],
  "incidents": [
    {
      "detected_at": "2024-11-05T14:00:00Z",
      "resolved_at": "2024-11-05T14:45:00Z",
      "severity": "high"
    }
    // ... more incidents
  ]
}
```

## Example Usage

**Basic calculation:**
```
Calculate DORA metrics for our Platform team for November 2024:
- 45 deployments in 30 days
- 4 failures (8.9% failure rate)
- Median lead time: 2.5 hours
- Median MTTR: 42 minutes
```

**With detailed data:**
```
Calculate DORA metrics from this CI/CD data: [paste JSON from example above]
Include trend analysis if previous month data is available.
```

## Expected Output

The skill should provide:
- Calculated values for all four metrics
- Performance classification for each metric
- Overall performance level
- Benchmark comparison
- Trend analysis (if historical data provided)
- Key insights and recommendations

## Related Prompts

- [Analyze Trends](./trend-analysis.md) - Track metrics over time
- [Team Comparison](./team-comparison.md) - Compare multiple teams
- [Performance Classification](./classify-performance.md) - Determine performance level
