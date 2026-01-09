// ML frameworks from tech-id scanner
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

function loadData() {
  try {
    const filePath = path.join(__dirname, 'data', 'tech-id.json');
    if (fs.existsSync(filePath)) {
      const content = JSON.parse(fs.readFileSync(filePath, 'utf8'));
      const frameworks = (content?.findings?.frameworks || [])
        .filter(f => f.category === 'ml' || f.category === 'ai');
      if (frameworks.length > 0) {
        return frameworks.map(f => ({
          name: f.name || '',
          version: f.version || '',
          category: f.category || ''
        }));
      }
    }
  } catch (e) {
    // Ignore errors
  }
  // Return placeholder row
  return [{ name: '', version: '', category: 'none' }];
}

export const data = loadData();
