// Malcontent findings from packages scanner
import { scannerData } from './load-data.js';

const { packageAnalysis } = scannerData;

function loadData() {
  const findings = packageAnalysis?.findings?.malcontent;
  if (Array.isArray(findings) && findings.length > 0) {
    return findings.map(f => {
      if (!f) return null;
      return {
        package: f.package || '',
        file: f.file || '',
        severity: f.severity || 'unknown',
        category: f.category || '',
        rule: f.rule || '',
        description: f.description || '',
        risk_score: f.risk_score || 0
      };
    }).filter(Boolean);
  }
  // Return placeholder row (filtered in SQL)
  return [{ package: '', file: '', severity: 'none', category: '', rule: '', description: '', risk_score: 0 }];
}

export const data = loadData();
