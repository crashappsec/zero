import { scannerData } from './load-data.js';

const { codeSecurity } = scannerData;

// Extract git history security data from code-security scanner metadata
const gitHistorySecurity = codeSecurity?.metadata?.git_history_security || {};

// Build summary data
const summary = gitHistorySecurity.summary || {};

export const data = [{
  total_violations: summary.total_violations || 0,
  gitignore_violations: summary.gitignore_violations || 0,
  sensitive_files_found: summary.sensitive_files_found || 0,
  files_to_purge: summary.files_to_purge || 0,
  commits_scanned: summary.commits_scanned || 0,
  risk_score: summary.risk_score || 100,
  risk_level: summary.risk_level || 'excellent',
  note: summary.note || null
}];
