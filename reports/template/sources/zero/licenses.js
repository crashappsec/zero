import { scannerData } from './load-data.js';

const { packageAnalysis } = scannerData;

let result = [{ license: 'No data', count: 0 }];

const licenses = packageAnalysis?.findings?.licenses;
if (Array.isArray(licenses) && licenses.length > 0) {
  const counts = {};
  for (const pkg of licenses) {
    if (!pkg) continue;
    const lic = pkg.license || 'Unknown';
    counts[lic] = (counts[lic] || 0) + 1;
  }

  const entries = Object.entries(counts);
  if (entries.length > 0) {
    result = entries
      .map(([license, count]) => ({ license, count }))
      .sort((a, b) => b.count - a.count)
      .slice(0, 10);
  }
}

export const data = result;
