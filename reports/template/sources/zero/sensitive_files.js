import { scannerData } from './load-data.js';

const { codeSecurity } = scannerData;

// Extract git history security data from code-security scanner metadata
const gitHistorySecurity = codeSecurity?.metadata?.git_history_security || {};

// Sensitive files found in git history
const sensitiveFiles = (gitHistorySecurity.sensitive_files || []).map(f => ({
  file: f.file,
  category: f.category,
  severity: f.severity,
  description: f.description,
  first_commit_hash: f.first_commit?.short_hash || '',
  first_commit_author: f.first_commit?.author || '',
  first_commit_date: f.first_commit?.date || '',
  still_exists: f.still_exists,
  was_removed: f.was_removed,
  size_bytes: f.size_bytes || 0
}));

export const data = sensitiveFiles.length > 0 ? sensitiveFiles : [{
  file: '',
  category: '',
  severity: '',
  description: 'No sensitive files found',
  first_commit_hash: '',
  first_commit_author: '',
  first_commit_date: '',
  still_exists: false,
  was_removed: false,
  size_bytes: 0
}];
