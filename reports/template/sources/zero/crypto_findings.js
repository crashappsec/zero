import { scannerData } from './load-data.js';

const { crypto } = scannerData;

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

export const data = findings.length > 0 ? findings : [{ type: 'No data', file: '', severity: 'N/A' }];
