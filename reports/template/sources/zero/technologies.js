import { scannerData } from './load-data.js';

const { technology } = scannerData;

export const data = technology?.findings?.technologies || [{ name: 'No data', category: 'N/A', version: '' }];
