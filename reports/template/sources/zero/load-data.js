// Helper to load scanner JSON files
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const dataDir = path.join(__dirname, 'data');

export function loadJSON(filename) {
  const filePath = path.join(dataDir, filename);
  try {
    if (fs.existsSync(filePath)) {
      const content = fs.readFileSync(filePath, 'utf8');
      return JSON.parse(content);
    }
  } catch (e) {
    // Silently return empty object on error
  }
  return {};
}

// Load all scanner outputs once (use empty objects as defaults, never null)
export const scannerData = {
  sbom: loadJSON('sbom.json') || {},
  packageAnalysis: loadJSON('package-analysis.json') || {},
  codeSecurity: loadJSON('code-security.json') || {},
  crypto: loadJSON('crypto.json') || {},
  devops: loadJSON('devops.json') || {},
  codeQuality: loadJSON('code-quality.json') || {},
  technology: loadJSON('tech-id.json') || {},
  codeOwnership: loadJSON('code-ownership.json') || {},
  devx: loadJSON('devx.json') || {}
};

// Also export a data object for Evidence (required for JavaScript sources)
export const data = [{ loaded: true }];
