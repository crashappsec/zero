import { scannerData } from './load-data.js';

const { packageAnalysis } = scannerData;

let result = [{ license: 'No data', count: 0 }];

if (packageAnalysis?.findings?.licenses) {
  const counts = {};
  for (const pkg of packageAnalysis.findings.licenses) {
    const lic = pkg.license || 'Unknown';
    counts[lic] = (counts[lic] || 0) + 1;
  }

  result = Object.entries(counts)
    .map(([license, count]) => ({ license, count }))
    .sort((a, b) => b.count - a.count)
    .slice(0, 10);
}

export const data = result;
