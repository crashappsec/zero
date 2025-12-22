// ML models from tech-id scanner
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

function loadData() {
  try {
    const filePath = path.join(__dirname, 'data', 'tech-id.json');
    if (fs.existsSync(filePath)) {
      const content = JSON.parse(fs.readFileSync(filePath, 'utf8'));
      const models = content?.findings?.models || [];
      if (models.length > 0) {
        return models.map(m => ({
          name: m.name || m.file || '',
          type: m.type || 'unknown',
          format: m.format || '',
          size: m.size || 0,
          risk: m.risk || 'unknown'
        }));
      }
    }
  } catch (e) {
    // Ignore errors
  }
  // Return placeholder row
  return [{ name: '', type: 'none', format: '', size: 0, risk: 'none' }];
}

export const data = loadData();
