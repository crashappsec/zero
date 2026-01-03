'use client';

import { useState, useMemo, Suspense } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { MainLayout } from '@/components/layout/Sidebar';
import { Card, CardTitle, CardContent } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Badge, SeverityBadge } from '@/components/ui/Badge';
import { useFetch } from '@/hooks/useApi';
import { api } from '@/lib/api';
import type { Secret, Project } from '@/lib/types';
import {
  Search,
  Filter,
  ChevronDown,
  ChevronUp,
  Key,
  FileCode,
  AlertTriangle,
  Lock,
  Eye,
  EyeOff,
  ArrowUpDown,
  X,
  Copy,
} from 'lucide-react';

type SortField = 'severity' | 'type' | 'file';
type SortDirection = 'asc' | 'desc';

const severityOrder: Record<string, number> = {
  critical: 0,
  high: 1,
  medium: 2,
  low: 3,
  unknown: 4,
};

const typeIcons: Record<string, typeof Key> = {
  api_key: Key,
  password: Lock,
  token: Key,
  secret: Lock,
  private_key: FileCode,
};

function SecretRow({ secret, expanded, onToggle }: { secret: Secret; expanded: boolean; onToggle: () => void }) {
  const [showMatch, setShowMatch] = useState(false);
  const Icon = typeIcons[secret.type.toLowerCase()] || Key;

  return (
    <div className="border-b border-gray-700 last:border-0">
      <button
        onClick={onToggle}
        className="w-full flex items-center gap-4 px-4 py-3 text-left hover:bg-gray-800/50 transition-colors"
      >
        <SeverityBadge severity={secret.severity} />
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <Icon className="h-4 w-4 text-yellow-500" />
            <span className="font-medium text-white">{secret.type}</span>
          </div>
          <div className="flex items-center gap-2 mt-0.5 text-sm text-gray-400">
            <FileCode className="h-3 w-3" />
            <span className="truncate">{secret.file}</span>
            <span className="text-gray-600">:</span>
            <span className="text-blue-400">{secret.line}</span>
          </div>
        </div>
        {expanded ? (
          <ChevronUp className="h-4 w-4 text-gray-500" />
        ) : (
          <ChevronDown className="h-4 w-4 text-gray-500" />
        )}
      </button>

      {expanded && (
        <div className="px-4 py-3 bg-gray-800/30 border-t border-gray-700/50">
          <div className="space-y-4">
            <div>
              <h4 className="text-xs font-medium text-gray-500 uppercase mb-1">Description</h4>
              <p className="text-sm text-gray-300">{secret.description}</p>
            </div>

            <div className="grid gap-4 md:grid-cols-2">
              <div>
                <h4 className="text-xs font-medium text-gray-500 uppercase mb-1">Location</h4>
                <div className="flex items-center gap-2 text-sm">
                  <code className="px-2 py-1 bg-gray-900 rounded text-blue-400 font-mono">
                    {secret.file}:{secret.line}
                  </code>
                </div>
              </div>

              <div>
                <h4 className="text-xs font-medium text-gray-500 uppercase mb-1">Type</h4>
                <Badge variant="warning">{secret.type}</Badge>
              </div>
            </div>

            {secret.redacted_match && (
              <div>
                <div className="flex items-center justify-between mb-1">
                  <h4 className="text-xs font-medium text-gray-500 uppercase">Match</h4>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      setShowMatch(!showMatch);
                    }}
                    className="flex items-center gap-1 text-xs text-gray-500 hover:text-gray-400"
                  >
                    {showMatch ? (
                      <>
                        <EyeOff className="h-3 w-3" />
                        Hide
                      </>
                    ) : (
                      <>
                        <Eye className="h-3 w-3" />
                        Show (redacted)
                      </>
                    )}
                  </button>
                </div>
                {showMatch && (
                  <code className="block px-3 py-2 bg-gray-900 rounded text-sm font-mono text-gray-300 break-all">
                    {secret.redacted_match}
                  </code>
                )}
              </div>
            )}

            <div className="pt-2 border-t border-gray-700">
              <h4 className="text-xs font-medium text-gray-500 uppercase mb-2">Remediation</h4>
              <ul className="text-sm text-gray-300 space-y-1">
                <li>1. Remove the secret from the codebase</li>
                <li>2. Rotate the secret immediately</li>
                <li>3. Check git history for exposure</li>
                <li>4. Consider using environment variables or a secrets manager</li>
              </ul>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function TypeFilter({
  types,
  selected,
  onChange
}: {
  types: string[];
  selected: string[];
  onChange: (types: string[]) => void;
}) {
  const toggle = (type: string) => {
    if (selected.includes(type)) {
      onChange(selected.filter(t => t !== type));
    } else {
      onChange([...selected, type]);
    }
  };

  return (
    <div className="flex items-center gap-2 flex-wrap">
      <Filter className="h-4 w-4 text-gray-500" />
      <span className="text-sm text-gray-400">Type:</span>
      {types.map(type => (
        <button
          key={type}
          onClick={() => toggle(type)}
          className={`px-2 py-1 text-xs rounded-md transition-colors ${
            selected.includes(type)
              ? 'bg-yellow-600 text-white'
              : 'bg-gray-700 text-gray-400 hover:bg-gray-600'
          }`}
        >
          {type}
        </button>
      ))}
      {selected.length > 0 && (
        <button
          onClick={() => onChange([])}
          className="text-xs text-gray-500 hover:text-gray-400"
        >
          Clear
        </button>
      )}
    </div>
  );
}

function SecretsContent() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const projectId = searchParams.get('project');

  const [search, setSearch] = useState('');
  const [typeFilter, setTypeFilter] = useState<string[]>([]);
  const [sortField, setSortField] = useState<SortField>('severity');
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc');
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());

  // Fetch projects
  const { data: projectsData } = useFetch(() => api.projects.list(), []);
  const projects = projectsData?.data || [];

  // Fetch secrets
  const { data: secretsData, loading, error } = useFetch(
    () => projectId ? api.analysis.secrets(projectId) : Promise.resolve({ data: [], total: 0 }),
    [projectId]
  );
  const secrets = secretsData?.data || [];

  // Get unique types
  const secretTypes = useMemo(() => {
    const types = new Set<string>();
    secrets.forEach(s => types.add(s.type));
    return Array.from(types).sort();
  }, [secrets]);

  // Filter and sort
  const filteredSecrets = useMemo(() => {
    let result = [...secrets];

    if (search) {
      const lower = search.toLowerCase();
      result = result.filter(s =>
        s.file.toLowerCase().includes(lower) ||
        s.type.toLowerCase().includes(lower) ||
        s.description.toLowerCase().includes(lower)
      );
    }

    if (typeFilter.length > 0) {
      result = result.filter(s => typeFilter.includes(s.type));
    }

    result.sort((a, b) => {
      let cmp = 0;
      switch (sortField) {
        case 'severity':
          cmp = (severityOrder[a.severity.toLowerCase()] || 99) - (severityOrder[b.severity.toLowerCase()] || 99);
          break;
        case 'type':
          cmp = a.type.localeCompare(b.type);
          break;
        case 'file':
          cmp = a.file.localeCompare(b.file);
          break;
      }
      return sortDirection === 'asc' ? cmp : -cmp;
    });

    return result;
  }, [secrets, search, typeFilter, sortField, sortDirection]);

  // Stats
  const stats = useMemo(() => {
    const counts = { critical: 0, high: 0, medium: 0, low: 0 };
    const typeCount = new Map<string, number>();

    secrets.forEach(s => {
      const sev = s.severity.toLowerCase() as keyof typeof counts;
      if (sev in counts) counts[sev]++;
      typeCount.set(s.type, (typeCount.get(s.type) || 0) + 1);
    });

    return { ...counts, byType: typeCount };
  }, [secrets]);

  const toggleExpanded = (id: string) => {
    const newSet = new Set(expandedIds);
    if (newSet.has(id)) {
      newSet.delete(id);
    } else {
      newSet.add(id);
    }
    setExpandedIds(newSet);
  };

  const toggleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(d => d === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('asc');
    }
  };

  const handleProjectChange = (newProjectId: string) => {
    router.push(`/secrets?project=${encodeURIComponent(newProjectId)}`);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            <Key className="h-6 w-6 text-yellow-500" />
            Secrets Detection
          </h1>
          <p className="mt-1 text-gray-400">
            Detected secrets and credentials in your codebase
          </p>
        </div>
      </div>

      {/* Project Selector */}
      <Card>
        <CardContent>
          <div className="flex items-center gap-4">
            <label className="text-sm font-medium text-gray-300">Project:</label>
            <select
              value={projectId || ''}
              onChange={(e) => handleProjectChange(e.target.value)}
              className="flex-1 max-w-md rounded-md border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 focus:border-green-500 focus:outline-none focus:ring-1 focus:ring-green-500"
            >
              <option value="">Select a project...</option>
              {projects.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.id}
                </option>
              ))}
            </select>
          </div>
        </CardContent>
      </Card>

      {projectId && (
        <>
          {/* Stats */}
          <div className="grid grid-cols-4 gap-4">
            <Card className="text-center">
              <p className="text-2xl font-bold text-red-500">{stats.critical}</p>
              <p className="text-sm text-gray-400">Critical</p>
            </Card>
            <Card className="text-center">
              <p className="text-2xl font-bold text-orange-500">{stats.high}</p>
              <p className="text-sm text-gray-400">High</p>
            </Card>
            <Card className="text-center">
              <p className="text-2xl font-bold text-yellow-500">{stats.medium}</p>
              <p className="text-sm text-gray-400">Medium</p>
            </Card>
            <Card className="text-center">
              <p className="text-2xl font-bold text-blue-500">{stats.low}</p>
              <p className="text-sm text-gray-400">Low</p>
            </Card>
          </div>

          {/* Filters */}
          <Card>
            <CardContent>
              <div className="flex flex-col gap-4">
                <div className="flex-1">
                  <Input
                    placeholder="Search secrets..."
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    icon={<Search className="h-4 w-4" />}
                  />
                </div>
                {secretTypes.length > 0 && (
                  <TypeFilter
                    types={secretTypes}
                    selected={typeFilter}
                    onChange={setTypeFilter}
                  />
                )}
              </div>
            </CardContent>
          </Card>

          {/* Results */}
          <Card>
            <div className="flex items-center justify-between px-4 py-3 border-b border-gray-700">
              <div className="flex items-center gap-4">
                <span className="text-sm text-gray-400">
                  {filteredSecrets.length} of {secrets.length} secrets
                </span>
                {(search || typeFilter.length > 0) && (
                  <button
                    onClick={() => {
                      setSearch('');
                      setTypeFilter([]);
                    }}
                    className="flex items-center gap-1 text-xs text-gray-500 hover:text-gray-400"
                  >
                    <X className="h-3 w-3" />
                    Clear filters
                  </button>
                )}
              </div>
              <div className="flex items-center gap-2">
                <span className="text-xs text-gray-500">Sort by:</span>
                <button
                  onClick={() => toggleSort('severity')}
                  className={`flex items-center gap-1 px-2 py-1 text-xs rounded ${
                    sortField === 'severity' ? 'bg-gray-700 text-white' : 'text-gray-400 hover:text-white'
                  }`}
                >
                  Severity
                  <ArrowUpDown className="h-3 w-3" />
                </button>
                <button
                  onClick={() => toggleSort('type')}
                  className={`flex items-center gap-1 px-2 py-1 text-xs rounded ${
                    sortField === 'type' ? 'bg-gray-700 text-white' : 'text-gray-400 hover:text-white'
                  }`}
                >
                  Type
                  <ArrowUpDown className="h-3 w-3" />
                </button>
                <button
                  onClick={() => toggleSort('file')}
                  className={`flex items-center gap-1 px-2 py-1 text-xs rounded ${
                    sortField === 'file' ? 'bg-gray-700 text-white' : 'text-gray-400 hover:text-white'
                  }`}
                >
                  File
                  <ArrowUpDown className="h-3 w-3" />
                </button>
              </div>
            </div>

            {loading ? (
              <div className="p-8 text-center text-gray-400">Loading secrets...</div>
            ) : error ? (
              <div className="p-8 text-center text-red-400">Failed to load secrets</div>
            ) : filteredSecrets.length === 0 ? (
              <div className="p-8 text-center">
                <Lock className="h-12 w-12 text-green-600 mx-auto mb-4" />
                <p className="text-gray-400">
                  {secrets.length === 0
                    ? 'No secrets detected in this project'
                    : 'No secrets match your filters'}
                </p>
              </div>
            ) : (
              <div>
                {filteredSecrets.map((secret, i) => (
                  <SecretRow
                    key={`${secret.file}-${secret.line}-${i}`}
                    secret={secret}
                    expanded={expandedIds.has(`${secret.file}-${secret.line}`)}
                    onToggle={() => toggleExpanded(`${secret.file}-${secret.line}`)}
                  />
                ))}
              </div>
            )}
          </Card>
        </>
      )}

      {!projectId && (
        <Card className="text-center py-12">
          <Key className="h-12 w-12 text-gray-600 mx-auto mb-4" />
          <p className="text-gray-400">Select a project to view detected secrets</p>
        </Card>
      )}
    </div>
  );
}

export default function SecretsPage() {
  return (
    <MainLayout>
      <Suspense fallback={<div className="animate-pulse h-96 bg-gray-800 rounded-lg" />}>
        <SecretsContent />
      </Suspense>
    </MainLayout>
  );
}
