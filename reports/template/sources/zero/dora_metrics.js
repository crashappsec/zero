import { scannerData } from './load-data.js';

const { devops } = scannerData;

export const data = devops?.summary?.dora ? [devops.summary.dora] : [{
  deployment_frequency: 'N/A',
  lead_time: 'N/A',
  change_failure_rate: 'N/A',
  mttr: 'N/A'
}];
