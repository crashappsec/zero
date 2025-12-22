import { scannerData } from './load-data.js';

const { codeOwnership } = scannerData;

export const data = codeOwnership?.findings?.contributors?.slice(0, 10) || [{ name: 'No data', commits: 0, lines_added: 0 }];
