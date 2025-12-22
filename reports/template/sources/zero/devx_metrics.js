import { scannerData } from './load-data.js';

const { devx, technology } = scannerData;

const metrics = [];

if (devx?.summary) {
  const s = devx.summary;

  if (s.onboarding) {
    metrics.push({
      category: 'Onboarding',
      metric: 'Overall Score',
      value: s.onboarding.score || 0,
      status: s.onboarding.score >= 70 ? 'Good' : s.onboarding.score >= 40 ? 'Fair' : 'Needs Work'
    });

    if (s.onboarding.readme) {
      metrics.push({
        category: 'Onboarding',
        metric: 'README Quality',
        value: s.onboarding.readme.score || 0,
        status: s.onboarding.readme.exists ? 'Present' : 'Missing'
      });
    }

    if (s.onboarding.contributing) {
      metrics.push({
        category: 'Onboarding',
        metric: 'Contributing Guide',
        value: s.onboarding.contributing.score || 0,
        status: s.onboarding.contributing.exists ? 'Present' : 'Missing'
      });
    }
  }

  if (s.sprawl) {
    metrics.push({
      category: 'Sprawl',
      metric: 'Tool Count',
      value: s.sprawl.tool_count || 0,
      status: s.sprawl.tool_count < 10 ? 'Good' : s.sprawl.tool_count < 20 ? 'Moderate' : 'High'
    });

    metrics.push({
      category: 'Sprawl',
      metric: 'Technology Count',
      value: s.sprawl.tech_count || 0,
      status: s.sprawl.tech_count < 5 ? 'Focused' : s.sprawl.tech_count < 10 ? 'Moderate' : 'Diverse'
    });
  }
}

export const data = metrics.length > 0 ? metrics : [{ category: 'N/A', metric: 'No data', value: 0, status: 'N/A' }];
