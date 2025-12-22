import { scannerData } from './load-data.js';

const { packageAnalysis } = scannerData;

const byEcosystem = {};

// Get packages from package-analysis licenses
if (packageAnalysis?.findings?.licenses) {
  for (const p of packageAnalysis.findings.licenses) {
    const eco = p.ecosystem || 'unknown';
    byEcosystem[eco] = (byEcosystem[eco] || 0) + 1;
  }
}

export const data = Object.entries(byEcosystem).map(([ecosystem, count]) => ({
  ecosystem,
  count
}));
