import { scannerData } from './load-data.js';

const { packageAnalysis, codeSecurity } = scannerData;

const vulns = [];

// Package vulnerabilities
if (packageAnalysis?.findings?.vulns) {
  for (const v of packageAnalysis.findings.vulns) {
    vulns.push({
      source: 'Package',
      package: v.package || v.name || '',
      version: v.version || '',
      severity: v.severity || 'unknown',
      cve: v.cve || v.id || '',
      title: v.title || v.description || '',
      fix_version: v.fix_version || v.patched_versions || ''
    });
  }
}

// Code vulnerabilities
if (codeSecurity?.findings?.vulns) {
  for (const v of codeSecurity.findings.vulns) {
    vulns.push({
      source: 'Code',
      package: v.file || '',
      version: 'Line ' + (v.line || '?'),
      severity: v.severity || 'unknown',
      cve: v.rule_id || '',
      title: v.title || v.message || '',
      fix_version: ''
    });
  }
}

export const data = vulns.length > 0 ? vulns : [{ source: 'None', package: '', version: '', severity: '', cve: '', title: 'No vulnerabilities found', fix_version: '' }];
