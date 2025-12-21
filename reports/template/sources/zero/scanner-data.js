import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const dataDir = path.join(__dirname, 'data');

// Load JSON file safely
function loadJSON(filename) {
  const filePath = path.join(dataDir, filename);
  try {
    if (fs.existsSync(filePath)) {
      return JSON.parse(fs.readFileSync(filePath, 'utf8'));
    }
  } catch (e) {
    console.warn(`Warning: Could not load ${filename}:`, e.message);
  }
  return null;
}

// Load all scanner outputs
const sbom = loadJSON('sbom.json');
const packageAnalysis = loadJSON('package-analysis.json');
const codeSecurity = loadJSON('code-security.json');
const crypto = loadJSON('crypto.json');
const devops = loadJSON('devops.json');
const codeQuality = loadJSON('code-quality.json');
const technology = loadJSON('technology.json');
const codeOwnership = loadJSON('code-ownership.json');
const devx = loadJSON('devx.json');

// Extract repository name
const repository = sbom?.repository || devops?.repository || packageAnalysis?.repository || 'Unknown';

// Metadata for the report
export const metadata = [
  {
    repository: repository.replace(/.*\/repos\//, '').replace('/repo', ''),
    timestamp: sbom?.timestamp || devops?.timestamp || new Date().toISOString(),
    scanners_run: [
      sbom && 'sbom',
      packageAnalysis && 'package-analysis',
      codeSecurity && 'code-security',
      crypto && 'crypto',
      devops && 'devops',
      codeQuality && 'code-quality',
      technology && 'tech-id',
      codeOwnership && 'code-ownership',
      devx && 'devx'
    ].filter(Boolean).join(', ')
  }
];

// Severity counts across all scanners
export const severity_counts = (() => {
  let critical = 0, high = 0, medium = 0, low = 0;

  // Package vulnerabilities
  if (packageAnalysis?.summary?.vulns) {
    critical += packageAnalysis.summary.vulns.critical || 0;
    high += packageAnalysis.summary.vulns.high || 0;
    medium += packageAnalysis.summary.vulns.medium || 0;
    low += packageAnalysis.summary.vulns.low || 0;
  }

  // Code security
  if (codeSecurity?.summary?.vulns) {
    critical += codeSecurity.summary.vulns.critical || 0;
    high += codeSecurity.summary.vulns.high || 0;
    medium += codeSecurity.summary.vulns.medium || 0;
    low += codeSecurity.summary.vulns.low || 0;
  }

  // DevOps IaC
  if (devops?.summary?.iac) {
    critical += devops.summary.iac.critical || 0;
    high += devops.summary.iac.high || 0;
    medium += devops.summary.iac.medium || 0;
    low += devops.summary.iac.low || 0;
  }

  // DevOps containers
  if (devops?.summary?.containers) {
    critical += devops.summary.containers.critical || 0;
    high += devops.summary.containers.high || 0;
    medium += devops.summary.containers.medium || 0;
    low += devops.summary.containers.low || 0;
  }

  // GitHub Actions
  if (devops?.summary?.github_actions) {
    critical += devops.summary.github_actions.critical || 0;
    high += devops.summary.github_actions.high || 0;
    medium += devops.summary.github_actions.medium || 0;
    low += devops.summary.github_actions.low || 0;
  }

  // Crypto findings
  if (crypto?.summary?.ciphers) {
    high += crypto.summary.ciphers.weak || 0;
  }
  if (crypto?.summary?.keys) {
    critical += crypto.summary.keys.hardcoded || 0;
  }

  return [{ critical, high, medium, low, total: critical + high + medium + low }];
})();

// Scanner summary table
export const scanner_summary = (() => {
  const rows = [];

  if (sbom) {
    rows.push({
      scanner: 'SBOM',
      status: sbom.summary?.generation?.error ? 'Error' : 'OK',
      findings: sbom.summary?.generation?.total_components || 0,
      summary: `${sbom.summary?.generation?.total_components || 0} packages`
    });
  }

  if (packageAnalysis) {
    const v = packageAnalysis.summary?.vulns || {};
    const total = (v.critical || 0) + (v.high || 0) + (v.medium || 0) + (v.low || 0);
    rows.push({
      scanner: 'Package Analysis',
      status: total > 0 ? 'Findings' : 'OK',
      findings: total,
      summary: `${v.critical || 0} critical, ${v.high || 0} high`
    });
  }

  if (codeSecurity) {
    const secrets = codeSecurity.summary?.secrets?.total || 0;
    const vulns = codeSecurity.summary?.vulns?.total || 0;
    rows.push({
      scanner: 'Code Security',
      status: secrets > 0 || vulns > 0 ? 'Findings' : 'OK',
      findings: secrets + vulns,
      summary: `${secrets} secrets, ${vulns} vulns`
    });
  }

  if (crypto) {
    const weak = crypto.summary?.ciphers?.weak || 0;
    const hardcoded = crypto.summary?.keys?.hardcoded || 0;
    rows.push({
      scanner: 'Cryptography',
      status: weak > 0 || hardcoded > 0 ? 'Findings' : 'OK',
      findings: weak + hardcoded,
      summary: `${weak} weak ciphers, ${hardcoded} hardcoded keys`
    });
  }

  if (devops) {
    const iac = devops.summary?.iac?.total_findings || 0;
    const actions = devops.summary?.github_actions?.total_findings || 0;
    rows.push({
      scanner: 'DevOps',
      status: iac > 0 || actions > 0 ? 'Findings' : 'OK',
      findings: iac + actions,
      summary: `${iac} IaC, ${actions} actions issues`
    });
  }

  if (codeQuality) {
    const score = codeQuality.summary?.overall_score || 0;
    rows.push({
      scanner: 'Code Quality',
      status: score < 50 ? 'Warning' : 'OK',
      findings: 100 - score,
      summary: `Score: ${score}/100`
    });
  }

  if (technology) {
    const techs = technology.summary?.technologies?.length || 0;
    rows.push({
      scanner: 'Tech Detection',
      status: 'OK',
      findings: techs,
      summary: `${techs} technologies detected`
    });
  }

  if (codeOwnership) {
    const busFactor = codeOwnership.summary?.bus_factor || 0;
    rows.push({
      scanner: 'Code Ownership',
      status: busFactor < 2 ? 'Warning' : 'OK',
      findings: busFactor,
      summary: `Bus factor: ${busFactor}`
    });
  }

  if (devx) {
    const score = devx.summary?.onboarding?.score || 0;
    rows.push({
      scanner: 'Developer Experience',
      status: score < 50 ? 'Warning' : 'OK',
      findings: score,
      summary: `Onboarding: ${score}%`
    });
  }

  return rows;
})();

// Findings by scanner for chart
export const findings_by_scanner = (() => {
  const data = [];

  if (packageAnalysis?.summary?.vulns) {
    const v = packageAnalysis.summary.vulns;
    if (v.critical) data.push({ scanner: 'Packages', severity: 'Critical', count: v.critical });
    if (v.high) data.push({ scanner: 'Packages', severity: 'High', count: v.high });
    if (v.medium) data.push({ scanner: 'Packages', severity: 'Medium', count: v.medium });
    if (v.low) data.push({ scanner: 'Packages', severity: 'Low', count: v.low });
  }

  if (codeSecurity?.summary?.vulns) {
    const v = codeSecurity.summary.vulns;
    if (v.critical) data.push({ scanner: 'Code', severity: 'Critical', count: v.critical });
    if (v.high) data.push({ scanner: 'Code', severity: 'High', count: v.high });
    if (v.medium) data.push({ scanner: 'Code', severity: 'Medium', count: v.medium });
    if (v.low) data.push({ scanner: 'Code', severity: 'Low', count: v.low });
  }

  if (devops?.summary?.iac) {
    const v = devops.summary.iac;
    if (v.critical) data.push({ scanner: 'IaC', severity: 'Critical', count: v.critical });
    if (v.high) data.push({ scanner: 'IaC', severity: 'High', count: v.high });
    if (v.medium) data.push({ scanner: 'IaC', severity: 'Medium', count: v.medium });
    if (v.low) data.push({ scanner: 'IaC', severity: 'Low', count: v.low });
  }

  if (crypto?.summary?.ciphers?.weak) {
    data.push({ scanner: 'Crypto', severity: 'High', count: crypto.summary.ciphers.weak });
  }

  return data.length > 0 ? data : [{ scanner: 'None', severity: 'None', count: 0 }];
})();

// Vulnerability details
export const vulnerabilities = (() => {
  const vulns = [];

  // Package vulnerabilities
  if (packageAnalysis?.findings?.vulns) {
    for (const v of packageAnalysis.findings.vulns) {
      vulns.push({
        source: 'Package',
        package: v.package || v.name,
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
        version: `Line ${v.line || '?'}`,
        severity: v.severity || 'unknown',
        cve: v.rule_id || '',
        title: v.title || v.message || '',
        fix_version: ''
      });
    }
  }

  return vulns;
})();

// DORA metrics
export const dora_metrics = devops?.summary?.dora ? [devops.summary.dora] : [];

// IaC findings
export const iac_findings = devops?.findings?.iac || [];

// GitHub Actions findings
export const github_actions_findings = devops?.findings?.github_actions || [];

// Container findings
export const container_findings = devops?.findings?.containers || [];

// Secrets detected
export const secrets = codeSecurity?.findings?.secrets || [];

// Crypto findings
export const crypto_findings = (() => {
  const findings = [];

  if (crypto?.findings?.ciphers) {
    for (const c of crypto.findings.ciphers) {
      findings.push({ type: 'Cipher', ...c });
    }
  }

  if (crypto?.findings?.keys) {
    for (const k of crypto.findings.keys) {
      findings.push({ type: 'Key', ...k });
    }
  }

  return findings;
})();

// License distribution
export const licenses = (() => {
  if (!packageAnalysis?.findings?.licenses) return [];

  const counts = {};
  for (const pkg of packageAnalysis.findings.licenses) {
    const lic = pkg.license || 'Unknown';
    counts[lic] = (counts[lic] || 0) + 1;
  }

  return Object.entries(counts)
    .map(([license, count]) => ({ license, count }))
    .sort((a, b) => b.count - a.count)
    .slice(0, 10);
})();

// Technologies detected
export const technologies = technology?.findings?.technologies || [];

// Code ownership summary
export const ownership_summary = codeOwnership?.summary ? [codeOwnership.summary] : [];

// Top contributors
export const contributors = codeOwnership?.findings?.contributors?.slice(0, 10) || [];
