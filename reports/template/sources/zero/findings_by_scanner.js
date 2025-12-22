import { scannerData } from './load-data.js';

const { packageAnalysis, codeSecurity, devops, crypto } = scannerData;

const rows = [];

if (packageAnalysis?.summary?.vulns) {
  const v = packageAnalysis.summary.vulns;
  if (v.critical) rows.push({ scanner: 'Packages', severity: 'Critical', count: v.critical });
  if (v.high) rows.push({ scanner: 'Packages', severity: 'High', count: v.high });
  if (v.medium) rows.push({ scanner: 'Packages', severity: 'Medium', count: v.medium });
  if (v.low) rows.push({ scanner: 'Packages', severity: 'Low', count: v.low });
}

if (codeSecurity?.summary?.vulns) {
  const v = codeSecurity.summary.vulns;
  if (v.critical) rows.push({ scanner: 'Code', severity: 'Critical', count: v.critical });
  if (v.high) rows.push({ scanner: 'Code', severity: 'High', count: v.high });
  if (v.medium) rows.push({ scanner: 'Code', severity: 'Medium', count: v.medium });
  if (v.low) rows.push({ scanner: 'Code', severity: 'Low', count: v.low });
}

if (devops?.summary?.iac) {
  const v = devops.summary.iac;
  if (v.critical) rows.push({ scanner: 'IaC', severity: 'Critical', count: v.critical });
  if (v.high) rows.push({ scanner: 'IaC', severity: 'High', count: v.high });
  if (v.medium) rows.push({ scanner: 'IaC', severity: 'Medium', count: v.medium });
  if (v.low) rows.push({ scanner: 'IaC', severity: 'Low', count: v.low });
}

if (crypto?.summary?.ciphers?.weak) {
  rows.push({ scanner: 'Crypto', severity: 'High', count: crypto.summary.ciphers.weak });
}

export const data = rows.length > 0 ? rows : [{ scanner: 'None', severity: 'None', count: 0 }];
