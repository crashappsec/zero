// Package health from packages scanner
import { scannerData } from './load-data.js';

const { packageAnalysis } = scannerData;

function loadData() {
  const health = packageAnalysis?.findings?.health;
  if (Array.isArray(health) && health.length > 0) {
    return health.map(p => {
      if (!p) return null;
      return {
        package: p.package || p.name || '',
        version: p.version || '',
        ecosystem: p.ecosystem || '',
        health_score: p.health_score || p.score || 0,
        deprecated: p.deprecated || false,
        unmaintained: p.unmaintained || false,
        last_release: p.last_release || '',
        open_issues: p.open_issues || 0,
        stars: p.stars || 0
      };
    }).filter(Boolean);
  }
  // Return placeholder row
  return [{ package: '', version: '', ecosystem: 'none', health_score: -1, deprecated: false, unmaintained: false, last_release: '', open_issues: 0, stars: 0 }];
}

export const data = loadData();
