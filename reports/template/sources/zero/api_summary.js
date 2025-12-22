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
  endpoints_found: apiSummary.endpoints_found || 0
}];
