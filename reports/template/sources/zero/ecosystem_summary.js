import { scannerData } from './load-data.js';

function getEcosystemSummary() {
  const { packageAnalysis } = scannerData;
  const byEcosystem = {};

  // Get packages from package-analysis licenses
  const licenses = packageAnalysis?.findings?.licenses;
  if (Array.isArray(licenses)) {
    for (const p of licenses) {
      if (!p) continue;
      const eco = p.ecosystem || 'unknown';
      byEcosystem[eco] = (byEcosystem[eco] || 0) + 1;
    }
  }

  const result = Object.entries(byEcosystem).map(([ecosystem, count]) => ({
    ecosystem,
    count
  }));

  // Return at least one row for Evidence
  return result.length > 0 ? result : [{ ecosystem: 'none', count: 0 }];
}

export const data = getEcosystemSummary();
