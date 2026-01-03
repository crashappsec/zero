// Scanner metadata for the web UI
// Maps scanner names to display names, features, and descriptions

export interface ScannerFeature {
  key: string;
  name: string;
  description: string;
}

export interface ScannerMetadata {
  name: string;
  displayName: string;
  description: string;
  icon: string;
  features: ScannerFeature[];
}

// Scanner → Features mapping (from CLAUDE.md)
export const scannerMetadata: Record<string, ScannerMetadata> = {
  sbom: {
    name: 'sbom',
    displayName: 'SBOM',
    description: 'Software Bill of Materials generation and integrity verification',
    icon: 'Package',
    features: [
      { key: 'generation', name: 'Generation', description: 'SBOM generation in CycloneDX format' },
      { key: 'integrity', name: 'Integrity', description: 'SBOM integrity verification' },
    ],
  },
  packages: {
    name: 'packages',
    displayName: 'Package Analysis',
    description: 'Comprehensive package and dependency analysis',
    icon: 'Package',
    features: [
      { key: 'vulns', name: 'Vulnerabilities', description: 'Known vulnerability detection via OSV.dev' },
      { key: 'health', name: 'Health', description: 'Package health and maintenance scores' },
      { key: 'licenses', name: 'Licenses', description: 'License detection and compatibility analysis' },
      { key: 'malcontent', name: 'Malcontent', description: 'Malicious package behavior detection' },
      { key: 'confusion', name: 'Confusion', description: 'Dependency confusion detection' },
      { key: 'typosquats', name: 'Typosquats', description: 'Typosquatting package detection' },
      { key: 'deprecations', name: 'Deprecations', description: 'Deprecated package detection' },
      { key: 'duplicates', name: 'Duplicates', description: 'Duplicate dependency detection' },
      { key: 'reachability', name: 'Reachability', description: 'Vulnerability reachability analysis' },
      { key: 'provenance', name: 'Provenance', description: 'Package provenance verification' },
      { key: 'bundle', name: 'Bundle', description: 'Bundle size analysis' },
      { key: 'recommendations', name: 'Recommendations', description: 'Update recommendations' },
    ],
  },
  'code-crypto': {
    name: 'code-crypto',
    displayName: 'Code Crypto',
    description: 'Cryptographic security analysis',
    icon: 'Lock',
    features: [
      { key: 'ciphers', name: 'Ciphers', description: 'Cipher usage and strength analysis' },
      { key: 'keys', name: 'Keys', description: 'Cryptographic key detection' },
      { key: 'random', name: 'Random', description: 'Random number generation security' },
      { key: 'tls', name: 'TLS', description: 'TLS configuration analysis' },
      { key: 'certificates', name: 'Certificates', description: 'Certificate security analysis' },
    ],
  },
  'code-security': {
    name: 'code-security',
    displayName: 'Code Security',
    description: 'Static analysis and secret detection',
    icon: 'Shield',
    features: [
      { key: 'vulns', name: 'Vulnerabilities', description: 'Code vulnerability detection (SAST)' },
      { key: 'secrets', name: 'Secrets', description: 'Secret and credential detection' },
      { key: 'api', name: 'API', description: 'API security analysis' },
      { key: 'git_history_security', name: 'Git History', description: 'Historical secret detection in git history' },
    ],
  },
  'code-quality': {
    name: 'code-quality',
    displayName: 'Code Quality',
    description: 'Code quality metrics and analysis',
    icon: 'BarChart3',
    features: [
      { key: 'tech_debt', name: 'Tech Debt', description: 'Technical debt assessment' },
      { key: 'complexity', name: 'Complexity', description: 'Code complexity metrics' },
      { key: 'test_coverage', name: 'Test Coverage', description: 'Test coverage analysis' },
      { key: 'documentation', name: 'Documentation', description: 'Documentation quality' },
    ],
  },
  devops: {
    name: 'devops',
    displayName: 'DevOps',
    description: 'DevOps and CI/CD security analysis',
    icon: 'Server',
    features: [
      { key: 'iac', name: 'IaC', description: 'Infrastructure as Code security' },
      { key: 'containers', name: 'Containers', description: 'Container security analysis' },
      { key: 'github_actions', name: 'GitHub Actions', description: 'CI/CD pipeline security' },
      { key: 'dora', name: 'DORA', description: 'DORA metrics (deployment frequency, lead time, etc.)' },
      { key: 'git', name: 'Git', description: 'Git repository metrics' },
    ],
  },
  'tech-id': {
    name: 'tech-id',
    displayName: 'Tech ID',
    description: 'Technology detection and ML-BOM generation',
    icon: 'Cpu',
    features: [
      { key: 'detection', name: 'Detection', description: 'Technology stack detection' },
      { key: 'models', name: 'Models', description: 'ML model detection' },
      { key: 'frameworks', name: 'Frameworks', description: 'Framework detection' },
      { key: 'datasets', name: 'Datasets', description: 'Dataset detection' },
      { key: 'ai_security', name: 'AI Security', description: 'AI/ML security analysis' },
      { key: 'ai_governance', name: 'AI Governance', description: 'AI governance compliance' },
      { key: 'infrastructure', name: 'Infrastructure', description: 'Infrastructure detection' },
    ],
  },
  'code-ownership': {
    name: 'code-ownership',
    displayName: 'Code Ownership',
    description: 'Code ownership and contributor analysis',
    icon: 'Users',
    features: [
      { key: 'contributors', name: 'Contributors', description: 'Contributor analysis' },
      { key: 'bus_factor', name: 'Bus Factor', description: 'Bus factor risk assessment' },
      { key: 'codeowners', name: 'CODEOWNERS', description: 'CODEOWNERS coverage analysis' },
      { key: 'orphans', name: 'Orphans', description: 'Orphaned code detection' },
      { key: 'churn', name: 'Churn', description: 'Code churn analysis' },
      { key: 'patterns', name: 'Patterns', description: 'Contribution patterns' },
    ],
  },
  devx: {
    name: 'devx',
    displayName: 'Developer Experience',
    description: 'Developer experience analysis',
    icon: 'Sparkles',
    features: [
      { key: 'onboarding', name: 'Onboarding', description: 'Developer onboarding assessment' },
      { key: 'sprawl', name: 'Sprawl', description: 'Tool and technology sprawl analysis' },
      { key: 'workflow', name: 'Workflow', description: 'Developer workflow analysis' },
    ],
  },
};

// Get scanner metadata by name
export function getScannerMetadata(scanner: string): ScannerMetadata | undefined {
  return scannerMetadata[scanner];
}

// Get all scanner names
export function getAllScannerNames(): string[] {
  return Object.keys(scannerMetadata);
}

// Get display name for a scanner
export function getScannerDisplayName(scanner: string): string {
  return scannerMetadata[scanner]?.displayName || scanner;
}
