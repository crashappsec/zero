import { scannerData } from './load-data.js';

function getSbomPackages() {
  const { packageAnalysis } = scannerData;
  const packages = [];

  // Get packages from package-analysis licenses
  const licenses = packageAnalysis?.findings?.licenses;
  if (Array.isArray(licenses)) {
    for (const p of licenses) {
      if (!p) continue;
      const licenseList = Array.isArray(p.licenses) ? p.licenses : [];
      packages.push({
        name: p.package || p.name || '',
        version: p.version || '',
        ecosystem: p.ecosystem || '',
        license: licenseList.join(', ') || 'Unknown',
        license_status: p.status || 'unknown'
      });
    }
  }

  // Return at least one row for Evidence
  return packages.length > 0 ? packages : [{ name: '', version: '', ecosystem: '', license: '', license_status: '' }];
}

export const data = getSbomPackages();
