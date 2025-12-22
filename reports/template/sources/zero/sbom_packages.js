import { scannerData } from './load-data.js';

const { packageAnalysis } = scannerData;

const packages = [];

// Get packages from package-analysis licenses
if (packageAnalysis?.findings?.licenses) {
  for (const p of packageAnalysis.findings.licenses) {
    packages.push({
      name: p.package || p.name || '',
      version: p.version || '',
      ecosystem: p.ecosystem || '',
      license: (p.licenses || []).join(', ') || 'Unknown',
      license_status: p.status || 'unknown'
    });
  }
}

export const data = packages;
