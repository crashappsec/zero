// Malcontent findings from package-analysis scanner
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

function loadData() {
  try {
    const filePath = path.join(__dirname, 'data', 'package-analysis.json');
    if (fs.existsSync(filePath)) {
      const content = JSON.parse(fs.readFileSync(filePath, 'utf8'));
      const findings = content?.findings?.malcontent || [];
      if (findings.length > 0) {
        return findings.map(f => ({
          package: f.package || '',
          file: f.file || '',
          severity: f.severity || 'unknown',
          category: f.category || '',
          rule: f.rule || '',
          description: f.description || '',
          risk_score: f.risk_score || 0
        }));
      }
    }
  } catch (e) {
    // Ignore errors
  }
  // Return placeholder row (filtered in SQL)
  return [{ package: '', file: '', severity: 'none', category: '', rule: '', description: '', risk_score: 0 }];
}

export const data = loadData();
