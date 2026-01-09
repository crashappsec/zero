import { scannerData } from './load-data.js';

const { sbom, packageAnalysis, codeSecurity, crypto, devops, codeQuality, technology, codeOwnership, devx } = scannerData;

const rows = [];

if (sbom) {
  const pkgCount = sbom.summary?.generation?.total_components || 0;
  rows.push({
    scanner: 'SBOM',
    status: sbom.summary?.generation?.error ? 'Error' : 'OK',
    findings: pkgCount,
    summary: pkgCount + ' packages'
  });
}

if (packageAnalysis) {
  const v = packageAnalysis.summary?.vulns || {};
  const total = (v.critical || 0) + (v.high || 0) + (v.medium || 0) + (v.low || 0);
  rows.push({
    scanner: 'Package Analysis',
    status: total > 0 ? 'Findings' : 'OK',
    findings: total,
    summary: (v.critical || 0) + ' critical, ' + (v.high || 0) + ' high'
  });
}

if (codeSecurity) {
  const secrets = codeSecurity.summary?.secrets?.total || 0;
  const vulns = codeSecurity.summary?.vulns?.total || 0;
  rows.push({
    scanner: 'Code Security',
    status: secrets > 0 || vulns > 0 ? 'Findings' : 'OK',
    findings: secrets + vulns,
    summary: secrets + ' secrets, ' + vulns + ' vulns'
  });
}

if (crypto) {
  const weak = crypto.summary?.ciphers?.weak || 0;
  const hardcoded = crypto.summary?.keys?.hardcoded || 0;
  rows.push({
    scanner: 'Cryptography',
    status: weak > 0 || hardcoded > 0 ? 'Findings' : 'OK',
    findings: weak + hardcoded,
    summary: weak + ' weak ciphers, ' + hardcoded + ' hardcoded keys'
  });
}

if (devops) {
  const iac = devops.summary?.iac?.total_findings || 0;
  const actions = devops.summary?.github_actions?.total_findings || 0;
  rows.push({
    scanner: 'DevOps',
    status: iac > 0 || actions > 0 ? 'Findings' : 'OK',
    findings: iac + actions,
    summary: iac + ' IaC, ' + actions + ' actions issues'
  });
}

if (codeQuality) {
  const score = codeQuality.summary?.overall_score || 0;
  rows.push({
    scanner: 'Code Quality',
    status: score < 50 ? 'Warning' : 'OK',
    findings: 100 - score,
    summary: 'Score: ' + score + '/100'
  });
}

if (technology) {
  const techs = technology.summary?.technologies?.length || 0;
  rows.push({
    scanner: 'Tech Detection',
    status: 'OK',
    findings: techs,
    summary: techs + ' technologies detected'
  });
}

if (codeOwnership) {
  const busFactor = codeOwnership.summary?.bus_factor || 0;
  rows.push({
    scanner: 'Code Ownership',
    status: busFactor < 2 ? 'Warning' : 'OK',
    findings: busFactor,
    summary: 'Bus factor: ' + busFactor
  });
}

if (devx) {
  const score = devx.summary?.onboarding?.score || 0;
  rows.push({
    scanner: 'Developer Experience',
    status: score < 50 ? 'Warning' : 'OK',
    findings: score,
    summary: 'Onboarding: ' + score + '%'
  });
}

export const data = rows.length > 0 ? rows : [{ scanner: 'No data', status: 'N/A', findings: 0, summary: 'No scanner data available' }];
