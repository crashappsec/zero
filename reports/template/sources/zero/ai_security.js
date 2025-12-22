// AI/ML security findings from tech-id scanner
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

function loadData() {
  try {
    const filePath = path.join(__dirname, 'data', 'tech-id.json');
    if (fs.existsSync(filePath)) {
      const content = JSON.parse(fs.readFileSync(filePath, 'utf8'));
      const findings = content?.findings?.ai_security || [];
      if (findings.length > 0) {
        return findings.map(f => ({
          severity: f.severity || 'medium',
          category: f.category || 'AI Security',
          title: f.title || f.rule || '',
          file: f.file || '',
          line: f.line || 0,
          description: f.description || f.message || ''
        }));
      }
    }
  } catch (e) {
    // Ignore errors
  }
  // Return placeholder row (filtered in SQL with severity != 'none')
  return [{ severity: 'none', category: '', title: '', file: '', line: 0, description: '' }];
}

export const data = loadData();
