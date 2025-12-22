// API findings data source - security and quality
import { scannerData } from './load-data.js';

const codeSecurity = scannerData.codeSecurity || {};
const apiFindings = codeSecurity?.findings?.api || [];
const apiSummary = codeSecurity?.summary?.api || {};

// Transform API findings for display
export const data = apiFindings.map(f => ({
  severity: f.severity || 'info',
  confidence: f.confidence || 'medium',
  category: f.category || 'unknown',
  owasp_api: f.owasp_api || '',
  title: f.title || f.rule_id || 'Unknown Issue',
  description: f.description || '',
  file: f.file || '',
  line: f.line || 0,
  snippet: f.snippet || '',
  http_method: f.http_method || '',
  endpoint: f.endpoint || '',
  framework: f.framework || '',
  cwe: (f.cwe || []).join(', '),
  remediation: f.remediation || ''
}));
