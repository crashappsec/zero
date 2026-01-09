import { scannerData } from './load-data.js';

const { codeQuality } = scannerData;

const metrics = [];

if (codeQuality?.summary) {
  const s = codeQuality.summary;

  metrics.push({
    metric: 'Overall Score',
    value: s.overall_score || 0,
    rating: s.overall_score >= 80 ? 'Good' : s.overall_score >= 50 ? 'Fair' : 'Needs Work'
  });

  if (s.tech_debt) {
    metrics.push({
      metric: 'Tech Debt',
      value: s.tech_debt.total_issues || 0,
      rating: s.tech_debt.severity || 'unknown'
    });
  }

  if (s.complexity) {
    metrics.push({
      metric: 'Avg Complexity',
      value: s.complexity.average || 0,
      rating: s.complexity.average < 10 ? 'Good' : s.complexity.average < 20 ? 'Fair' : 'High'
    });
  }

  if (s.test_coverage) {
    metrics.push({
      metric: 'Test Coverage',
      value: s.test_coverage.percentage || 0,
      rating: s.test_coverage.percentage >= 80 ? 'Good' : s.test_coverage.percentage >= 50 ? 'Fair' : 'Low'
    });
  }

  if (s.documentation) {
    metrics.push({
      metric: 'Documentation',
      value: s.documentation.score || 0,
      rating: s.documentation.score >= 80 ? 'Good' : s.documentation.score >= 50 ? 'Fair' : 'Low'
    });
  }
}

export const data = metrics.length > 0 ? metrics : [{ metric: 'No data', value: 0, rating: 'N/A' }];
