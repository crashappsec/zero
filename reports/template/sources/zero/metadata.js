import { scannerData } from './load-data.js';

const { sbom, devops, packageAnalysis } = scannerData;
const repository = sbom?.repository || devops?.repository || packageAnalysis?.repository || 'Unknown';

export const data = [{
  repository: repository.replace(/.*\/repos\//, '').replace('/repo', ''),
  timestamp: sbom?.timestamp || devops?.timestamp || new Date().toISOString(),
  scanners_run: Object.entries(scannerData)
    .filter(([_, v]) => v !== null)
    .map(([k, _]) => k)
    .join(', ')
}];
