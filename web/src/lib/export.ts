// Export utilities for downloading data as CSV or JSON

export function downloadJSON<T>(data: T, filename: string): void {
  const json = JSON.stringify(data, null, 2);
  const blob = new Blob([json], { type: 'application/json' });
  downloadBlob(blob, `${filename}.json`);
}

export function downloadCSV(data: Record<string, unknown>[], filename: string): void {
  if (data.length === 0) return;

  // Get all unique keys across all objects
  const headers = Array.from(
    new Set(data.flatMap(row => Object.keys(row)))
  );

  // Create CSV content
  const csvRows: string[] = [];

  // Header row
  csvRows.push(headers.map(escapeCSV).join(','));

  // Data rows
  for (const row of data) {
    const values = headers.map(header => {
      const value = row[header];
      if (value === null || value === undefined) return '';
      if (typeof value === 'object') return escapeCSV(JSON.stringify(value));
      return escapeCSV(String(value));
    });
    csvRows.push(values.join(','));
  }

  const csv = csvRows.join('\n');
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
  downloadBlob(blob, `${filename}.csv`);
}

function escapeCSV(value: string): string {
  // If value contains comma, newline, or quote, wrap in quotes and escape internal quotes
  if (value.includes(',') || value.includes('\n') || value.includes('"')) {
    return `"${value.replace(/"/g, '""')}"`;
  }
  return value;
}

function downloadBlob(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

// Pre-configured export functions for common data types
export function exportVulnerabilities(data: Record<string, unknown>[], format: 'csv' | 'json', projectId?: string): void {
  const filename = projectId ? `vulnerabilities-${projectId}` : 'vulnerabilities';
  if (format === 'csv') {
    downloadCSV(data, filename);
  } else {
    downloadJSON(data, filename);
  }
}

export function exportSecrets(data: Record<string, unknown>[], format: 'csv' | 'json', projectId?: string): void {
  const filename = projectId ? `secrets-${projectId}` : 'secrets';
  if (format === 'csv') {
    downloadCSV(data, filename);
  } else {
    downloadJSON(data, filename);
  }
}

export function exportDependencies(data: Record<string, unknown>[], format: 'csv' | 'json', projectId?: string): void {
  const filename = projectId ? `dependencies-${projectId}` : 'dependencies';
  if (format === 'csv') {
    downloadCSV(data, filename);
  } else {
    downloadJSON(data, filename);
  }
}

export function exportProjects(data: Record<string, unknown>[], format: 'csv' | 'json'): void {
  if (format === 'csv') {
    downloadCSV(data, 'projects');
  } else {
    downloadJSON(data, 'projects');
  }
}
