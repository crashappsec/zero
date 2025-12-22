import { scannerData } from './load-data.js';

const { codeOwnership } = scannerData;

export const data = codeOwnership?.summary ? [codeOwnership.summary] : [{ bus_factor: 0, total_contributors: 0 }];
