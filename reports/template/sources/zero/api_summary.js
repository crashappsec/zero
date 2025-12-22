// API summary data source
import { scannerData } from './load-data.js';

const codeSecurity = scannerData.codeSecurity || {};
const apiSummary = codeSecurity?.summary?.api || {};

// Export summary metrics
export const data = [{
  total_findings: apiSummary.total_findings || 0,
  critical: apiSummary.critical || 0,
  high: apiSummary.high || 0,
  medium: apiSummary.medium || 0,
  low: apiSummary.low || 0,
  endpoints_found: apiSummary.endpoints_found || 0,
  // Flatten category counts
  ...Object.fromEntries(
    Object.entries(apiSummary.by_category || {}).map(([k, v]) => [`cat_${k.replace(/-/g, '_')}`, v])
  ),
  // Flatten OWASP API counts
  ...Object.fromEntries(
    Object.entries(apiSummary.by_owasp_api || {}).map(([k, v]) => [`owasp_${k.replace(/[:\s]/g, '_')}`, v])
  ),
  // Flatten framework counts
  ...Object.fromEntries(
    Object.entries(apiSummary.by_framework || {}).map(([k, v]) => [`framework_${k}`, v])
  )
}];
