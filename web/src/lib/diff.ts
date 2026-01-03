// Diff utilities for comparing scan results

export interface DiffResult<T> {
  added: T[];
  removed: T[];
  unchanged: T[];
  changed: Array<{ before: T; after: T }>;
}

export interface VulnerabilityDiff {
  id: string;
  package: string;
  version: string;
  severity: string;
  title: string;
  status: 'added' | 'removed' | 'unchanged';
}

export interface SecretDiff {
  file: string;
  line: number;
  type: string;
  status: 'added' | 'removed' | 'unchanged';
}

export interface DependencyDiff {
  name: string;
  oldVersion?: string;
  newVersion?: string;
  status: 'added' | 'removed' | 'updated' | 'unchanged';
}

export interface ScanDiffSummary {
  vulnerabilities: {
    added: number;
    removed: number;
    unchanged: number;
  };
  secrets: {
    added: number;
    removed: number;
    unchanged: number;
  };
  dependencies: {
    added: number;
    removed: number;
    updated: number;
    unchanged: number;
  };
}

// Compare arrays using a key function
export function diffArrays<T>(
  before: T[],
  after: T[],
  getKey: (item: T) => string
): DiffResult<T> {
  const beforeMap = new Map(before.map(item => [getKey(item), item]));
  const afterMap = new Map(after.map(item => [getKey(item), item]));

  const added: T[] = [];
  const removed: T[] = [];
  const unchanged: T[] = [];
  const changed: Array<{ before: T; after: T }> = [];

  // Find added and unchanged/changed
  Array.from(afterMap.entries()).forEach(([key, afterItem]) => {
    const beforeItem = beforeMap.get(key);
    if (!beforeItem) {
      added.push(afterItem);
    } else if (JSON.stringify(beforeItem) === JSON.stringify(afterItem)) {
      unchanged.push(afterItem);
    } else {
      changed.push({ before: beforeItem, after: afterItem });
    }
  });

  // Find removed
  Array.from(beforeMap.entries()).forEach(([key, beforeItem]) => {
    if (!afterMap.has(key)) {
      removed.push(beforeItem);
    }
  });

  return { added, removed, unchanged, changed };
}

// Compare vulnerabilities between scans
export function diffVulnerabilities(
  before: Array<{ id: string; package: string; version: string; severity: string; title: string }>,
  after: Array<{ id: string; package: string; version: string; severity: string; title: string }>
): VulnerabilityDiff[] {
  const result = diffArrays(before, after, v => `${v.id}-${v.package}-${v.version}`);

  return [
    ...result.added.map(v => ({ ...v, status: 'added' as const })),
    ...result.removed.map(v => ({ ...v, status: 'removed' as const })),
    ...result.unchanged.map(v => ({ ...v, status: 'unchanged' as const })),
  ];
}

// Compare secrets between scans
export function diffSecrets(
  before: Array<{ file: string; line: number; type: string }>,
  after: Array<{ file: string; line: number; type: string }>
): SecretDiff[] {
  const result = diffArrays(before, after, s => `${s.file}:${s.line}:${s.type}`);

  return [
    ...result.added.map(s => ({ ...s, status: 'added' as const })),
    ...result.removed.map(s => ({ ...s, status: 'removed' as const })),
    ...result.unchanged.map(s => ({ ...s, status: 'unchanged' as const })),
  ];
}

// Compare dependencies between scans
export function diffDependencies(
  before: Array<{ name: string; version: string }>,
  after: Array<{ name: string; version: string }>
): DependencyDiff[] {
  const beforeMap = new Map(before.map(d => [d.name, d.version]));
  const afterMap = new Map(after.map(d => [d.name, d.version]));

  const result: DependencyDiff[] = [];

  // Find added, updated, unchanged
  Array.from(afterMap.entries()).forEach(([name, newVersion]) => {
    const oldVersion = beforeMap.get(name);
    if (!oldVersion) {
      result.push({ name, newVersion, status: 'added' });
    } else if (oldVersion !== newVersion) {
      result.push({ name, oldVersion, newVersion, status: 'updated' });
    } else {
      result.push({ name, oldVersion, newVersion, status: 'unchanged' });
    }
  });

  // Find removed
  Array.from(beforeMap.entries()).forEach(([name, oldVersion]) => {
    if (!afterMap.has(name)) {
      result.push({ name, oldVersion, status: 'removed' });
    }
  });

  return result;
}

// Generate summary of differences
export function createDiffSummary(
  vulnDiffs: VulnerabilityDiff[],
  secretDiffs: SecretDiff[],
  depDiffs: DependencyDiff[]
): ScanDiffSummary {
  return {
    vulnerabilities: {
      added: vulnDiffs.filter(v => v.status === 'added').length,
      removed: vulnDiffs.filter(v => v.status === 'removed').length,
      unchanged: vulnDiffs.filter(v => v.status === 'unchanged').length,
    },
    secrets: {
      added: secretDiffs.filter(s => s.status === 'added').length,
      removed: secretDiffs.filter(s => s.status === 'removed').length,
      unchanged: secretDiffs.filter(s => s.status === 'unchanged').length,
    },
    dependencies: {
      added: depDiffs.filter(d => d.status === 'added').length,
      removed: depDiffs.filter(d => d.status === 'removed').length,
      updated: depDiffs.filter(d => d.status === 'updated').length,
      unchanged: depDiffs.filter(d => d.status === 'unchanged').length,
    },
  };
}
