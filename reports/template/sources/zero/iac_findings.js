import { scannerData } from './load-data.js';

const { devops } = scannerData;

export const data = devops?.findings?.iac || [{ type: 'No data', severity: 'N/A', file: '', message: 'No IaC findings' }];
