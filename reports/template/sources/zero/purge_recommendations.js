import { scannerData } from './load-data.js';

const { codeSecurity } = scannerData;

// Extract git history security data from code-security scanner metadata
const gitHistorySecurity = codeSecurity?.metadata?.git_history_security || {};

// Purge recommendations
const purgeRecommendations = (gitHistorySecurity.purge_recommendations || []).map(r => ({
  file: r.file,
  reason: r.reason,
  severity: r.severity,
  priority: r.priority,
  command: r.command,
  alternative: r.alternative || '',
  affected_commits: r.affected_commits
}));

export const data = purgeRecommendations.length > 0 ? purgeRecommendations : [{
  file: '',
  reason: 'No files to purge',
  severity: '',
  priority: 0,
  command: '',
  alternative: '',
  affected_commits: 0
}];
