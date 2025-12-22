import { scannerData } from './load-data.js';

const { devops } = scannerData;

export const data = devops?.findings?.github_actions || [{ workflow: 'No data', severity: 'N/A', issue: '' }];
