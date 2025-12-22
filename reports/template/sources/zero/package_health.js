// Package health from package-analysis scanner
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

function loadData() {
  try {
    const filePath = path.join(__dirname, 'data', 'package-analysis.json');
    if (fs.existsSync(filePath)) {
      const content = JSON.parse(fs.readFileSync(filePath, 'utf8'));
      const health = content?.findings?.health || [];
      if (health.length > 0) {
        return health.map(p => ({
          package: p.package || p.name || '',
          version: p.version || '',
          ecosystem: p.ecosystem || '',
          health_score: p.health_score || p.score || 0,
          deprecated: p.deprecated || false,
          unmaintained: p.unmaintained || false,
          last_release: p.last_release || '',
          open_issues: p.open_issues || 0,
          stars: p.stars || 0
        }));
      }
    }
  } catch (e) {
    // Ignore errors
  }
  // Return placeholder row
  return [{ package: '', version: '', ecosystem: 'none', health_score: -1, deprecated: false, unmaintained: false, last_release: '', open_issues: 0, stars: 0 }];
}

export const data = loadData();
