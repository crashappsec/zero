import { scannerData } from './load-data.js';

const { devops } = scannerData;

export const data = devops?.findings?.containers || [{ image: 'No data', severity: 'N/A', issue: '' }];
