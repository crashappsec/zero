import { scannerData } from './load-data.js';

const { packageAnalysis, codeSecurity, devops, crypto } = scannerData;

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

// Crypto
if (crypto?.summary?.ciphers) {
  high += crypto.summary.ciphers.weak || 0;
}
if (crypto?.summary?.keys) {
  critical += crypto.summary.keys.hardcoded || 0;
}

export const data = [{ critical, high, medium, low, total: critical + high + medium + low }];
