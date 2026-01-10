// API Types

// Repo (formerly Project) represents a hydrated repository
export interface Repo {
  id: string;
  name: string;
  owner: string;
  repo: string;
  path: string;
  status: string;
  scanners: string[];
  last_scan: string;
  freshness?: FreshnessInfo;
}

// Backwards compatibility alias
export type Project = Repo;

export interface FreshnessInfo {
  level: 'fresh' | 'stale' | 'very_stale' | 'expired';
  level_string: string;
  age_string: string;
  needs_refresh: boolean;
}

export interface ScannerInfo {
  name: string;
  description: string;
  features: string[];
}

export interface AgentInfo {
  id: string;
  name: string;
  persona: string;
  description: string;
  scanner: string;
}

export interface ProfileInfo {
  name: string;
  description: string;
  estimated_time: string;
  scanners: string[];
}

// Scan types
export interface ScanJob {
  job_id: string;
  target: string;
  is_org?: boolean;
  profile: string;
  status: 'queued' | 'cloning' | 'scanning' | 'complete' | 'failed' | 'canceled';
  started_at: string;
  finished_at?: string;
  duration_seconds?: number;
  progress?: ScanProgress;
  project_ids?: string[];
  error?: string;
}

export interface ScanProgress {
  phase: string;
  repos_total: number;
  repos_complete: number;
  current_repo?: string;
  scanners_total: number;
  scanners_complete: number;
  current_scanner?: string;
  scanner_statuses?: Record<string, ScannerState>;
}

export interface ScannerState {
  status: string;
  summary?: string;
  duration_seconds?: number;
}

export interface QueueStats {
  total_jobs: number;
  queued_jobs: number;
  running_jobs: number;
  completed_jobs: number;
  failed_jobs: number;
  canceled_jobs: number;
}

// Chat types
export interface ChatMessage {
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: string;
}

export interface ChatSession {
  session_id: string;
  agent_id: string;
  project_id?: string;
  messages: ChatMessage[];
  created_at: string;
  updated_at: string;
}

export interface StreamChunk {
  type: 'start' | 'delta' | 'done' | 'error' | 'tool_call' | 'tool_result' | 'delegation';
  session_id: string;
  agent_id: string;
  content?: string;
  error?: string;
  // Tool call fields
  tool_name?: string;
  tool_input?: Record<string, unknown>;
  // Tool result fields
  is_error?: boolean;
  // Delegation fields (sub-agent progress)
  delegated_agent?: string;
  delegated_event?: 'start' | 'text' | 'tool_call' | 'tool_result' | 'done';
}

// Tool call tracking for threaded display
export interface ToolCallInfo {
  id: string;
  name: string;
  input: Record<string, unknown>;
  status: 'running' | 'complete' | 'error';
  startTime: number;
  endTime?: number;
}

// Analysis types
export interface AnalysisSummary {
  project_id: string;
  scanners: Record<string, ScannerSummary>;
  totals: {
    critical: number;
    high: number;
    medium: number;
    low: number;
  };
}

export interface ScannerSummary {
  name: string;
  status: string;
  findings_count: number;
  last_run: string;
}

export interface Vulnerability {
  id: string;
  package: string;
  version: string;
  severity: string;
  title: string;
  description?: string;
  fix_version?: string;
  source: string;
}

export interface Secret {
  file: string;
  line: number;
  type: string;
  severity: string;
  description: string;
  redacted_match?: string;
}

// Dependency types (from SBOM)
export interface Dependency {
  name: string;
  version: string;
  type: 'library' | 'framework' | 'application';
  purl?: string;
  license?: string;
  licenses?: string[];
  direct?: boolean;
  scope?: 'runtime' | 'development' | 'optional';
  dependencies?: string[];
  health?: DependencyHealth;
  vulns_count?: number;
}

export interface DependencyHealth {
  score: number;
  maintenance: number;
  popularity: number;
  quality: number;
  deprecated?: boolean;
  last_publish?: string;
}

export interface DependencyTree {
  root: string;
  nodes: Record<string, DependencyNode>;
}

export interface DependencyNode {
  name: string;
  version: string;
  children: string[];
  depth: number;
}

// Aggregate stats
export interface AggregateStats {
  total_projects: number;
  total_vulns: number;
  total_secrets: number;
  total_deps: number;
  vulns_by_severity: Record<string, number>;
  project_stats: ProjectStats[];
}

export interface ProjectStats {
  id: string;
  vulns: number;
  secrets: number;
  deps: number;
  severity: Record<string, number>;
}

// API Response wrappers
export interface ListResponse<T> {
  data: T[];
  total: number;
}

export interface HealthResponse {
  status: string;
  version: string;
  timestamp: string;
}

// Configuration types
export interface Settings {
  default_profile: string;
  storage_path: string;
  parallel_repos: number;
  parallel_scanners: number;
  scanner_timeout_seconds: number;
  cache_ttl_hours: number;
}

export interface ScannerConfig {
  name: string;
  description: string;
  estimated_time: string;
  output_file: string;
  features: Record<string, FeatureConfig>;
}

export interface FeatureConfig {
  enabled: boolean;
  [key: string]: unknown;
}

// Scanner metadata with features
export interface ScannerMeta {
  name: string;
  displayName: string;
  description: string;
  features: string[];
  icon?: string;
}
