import { scannerData } from './load-data.js';

const { codeSecurity } = scannerData;

export const data = codeSecurity?.findings?.secrets || [{ type: 'No data', file: '', severity: 'N/A', line: 0 }];
