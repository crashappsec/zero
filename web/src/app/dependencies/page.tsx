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
import type { Dependency, Project } from '@/lib/types';
import { ExportButton } from '@/components/ui/ExportButton';
import { downloadCSV, downloadJSON } from '@/lib/export';
import {
  Search,
  Package,
  ChevronRight,
  ChevronDown,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Filter,
  LayoutList,
  Network,
  ArrowUpDown,
  ExternalLink,
  Scale,
} from 'lucide-react';

type ViewMode = 'list' | 'tree';
type SortField = 'name' | 'license' | 'vulns';

function HealthIndicator({ health }: { health?: Dependency['health'] }) {
  if (!health) {
    return <span className="text-gray-500 text-sm">-</span>;
  }

  const score = health.score;
  const color = score >= 80 ? 'text-green-500' : score >= 50 ? 'text-yellow-500' : 'text-red-500';

  return (
    <div className="flex items-center gap-2">
      <div className={`font-medium ${color}`}>{score}</div>
      {health.deprecated && (
        <Badge variant="error" className="text-xs">Deprecated</Badge>
      )}
    </div>
  );
}

function LicenseBadge({ license }: { license?: string }) {
  if (!license) return <span className="text-gray-500 text-sm">Unknown</span>;

  const isPermissive = ['MIT', 'ISC', 'BSD-2-Clause', 'BSD-3-Clause', 'Apache-2.0', 'Unlicense'].includes(license);
  const isRestrictive = ['GPL-2.0', 'GPL-3.0', 'AGPL-3.0', 'LGPL-2.1', 'LGPL-3.0'].includes(license);

  return (
    <Badge
      variant={isPermissive ? 'success' : isRestrictive ? 'warning' : 'default'}
      className="text-xs"
    >
      {license}
    </Badge>
  );
}

function DependencyRow({ dep, expanded, onToggle }: { dep: Dependency; expanded: boolean; onToggle: () => void }) {
  const hasChildren = dep.dependencies && dep.dependencies.length > 0;
  const hasVulns = (dep.vulns_count || 0) > 0;

  return (
    <div className="border-b border-gray-700 last:border-0">
      <div
        className="flex items-center gap-4 px-4 py-3 hover:bg-gray-800/50 transition-colors cursor-pointer"
        onClick={onToggle}
      >
        <button className="w-5">
          {hasChildren ? (
            expanded ? (
              <ChevronDown className="h-4 w-4 text-gray-500" />
            ) : (
              <ChevronRight className="h-4 w-4 text-gray-500" />
            )
          ) : null}
        </button>

        <div className="flex items-center gap-2 min-w-0 flex-1">
          <Package className="h-4 w-4 text-gray-500 shrink-0" />
          <span className="font-medium text-white truncate">{dep.name}</span>
          <span className="text-gray-500 text-sm">@{dep.version}</span>
          {dep.direct && (
            <Badge variant="info" className="text-xs">direct</Badge>
          )}
          {dep.scope === 'development' && (
            <Badge variant="default" className="text-xs">dev</Badge>
          )}
        </div>

        <div className="flex items-center gap-6">
          {/* License */}
          <div className="w-28">
            <LicenseBadge license={dep.license} />
          </div>

          {/* Health */}
          <div className="w-20">
            <HealthIndicator health={dep.health} />
          </div>

          {/* Vulnerabilities */}
          <div className="w-16">
            {hasVulns ? (
              <span className="flex items-center gap-1 text-red-400">
                <AlertTriangle className="h-4 w-4" />
                {dep.vulns_count}
              </span>
            ) : (
              <span className="flex items-center gap-1 text-green-500">
                <CheckCircle className="h-4 w-4" />
              </span>
            )}
          </div>
        </div>
      </div>

      {expanded && hasChildren && (
        <div className="bg-gray-800/30 border-t border-gray-700/50 px-4 py-3">
          <h4 className="text-xs font-medium text-gray-500 uppercase mb-2">
            Dependencies ({dep.dependencies?.length})
          </h4>
          <div className="flex flex-wrap gap-1">
            {dep.dependencies?.map((d) => (
              <Badge key={d} variant="default" className="text-xs">
                {d}
              </Badge>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function TreeNode({
  dep,
  depth,
  allDeps,
  expanded,
  onToggle,
}: {
  dep: Dependency;
  depth: number;
  allDeps: Map<string, Dependency>;
  expanded: Set<string>;
  onToggle: (name: string) => void;
}) {
  const isExpanded = expanded.has(dep.name);
  const hasChildren = dep.dependencies && dep.dependencies.length > 0;
  const hasVulns = (dep.vulns_count || 0) > 0;

  return (
    <div>
      <div
        className="flex items-center gap-2 py-1.5 hover:bg-gray-800/50 rounded cursor-pointer"
        style={{ paddingLeft: `${depth * 20 + 8}px` }}
        onClick={() => onToggle(dep.name)}
      >
        <button className="w-4">
          {hasChildren ? (
            isExpanded ? (
              <ChevronDown className="h-3 w-3 text-gray-500" />
            ) : (
              <ChevronRight className="h-3 w-3 text-gray-500" />
            )
          ) : (
            <span className="w-3" />
          )}
        </button>

        <Package className={`h-4 w-4 ${hasVulns ? 'text-red-400' : 'text-gray-500'}`} />

        <span className="text-sm text-white">{dep.name}</span>
        <span className="text-xs text-gray-500">@{dep.version}</span>

        {hasVulns && (
          <Badge variant="error" className="text-xs ml-2">
            {dep.vulns_count} vuln{dep.vulns_count !== 1 ? 's' : ''}
          </Badge>
        )}
        {dep.health?.deprecated && (
          <Badge variant="warning" className="text-xs">deprecated</Badge>
        )}
      </div>

      {isExpanded && hasChildren && (
        <div>
          {dep.dependencies?.map((childName) => {
            const childDep = allDeps.get(childName);
            if (!childDep) {
              return (
                <div
                  key={childName}
                  className="flex items-center gap-2 py-1.5 text-gray-500 text-sm"
                  style={{ paddingLeft: `${(depth + 1) * 20 + 8}px` }}
                >
                  <span className="w-4" />
                  <Package className="h-4 w-4" />
                  <span>{childName}</span>
                  <span className="text-xs">(transitive)</span>
                </div>
              );
            }
            return (
              <TreeNode
                key={childName}
                dep={childDep}
                depth={depth + 1}
                allDeps={allDeps}
                expanded={expanded}
                onToggle={onToggle}
              />
            );
          })}
        </div>
      )}
    </div>
  );
}

function DependenciesContent() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const projectId = searchParams.get('project');

  const [search, setSearch] = useState('');
  const [viewMode, setViewMode] = useState<ViewMode>('list');
  const [showDirect, setShowDirect] = useState(false);
  const [showVulnerable, setShowVulnerable] = useState(false);
  const [sortField, setSortField] = useState<SortField>('name');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc');
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());

  // Fetch projects
  const { data: projectsData } = useFetch(() => api.projects.list(), []);
  const projects = projectsData?.data || [];

  // Fetch dependencies
  const { data: depsData, loading, error } = useFetch(
    () => projectId ? api.analysis.dependencies(projectId) : Promise.resolve({ data: [], total: 0 }),
    [projectId]
  );
  const dependencies = depsData?.data || [];

  // Create dependency map for tree view
  const depMap = useMemo(() => {
    const map = new Map<string, Dependency>();
    dependencies.forEach((d) => map.set(d.name, d));
    return map;
  }, [dependencies]);

  // Filter and sort
  const filteredDeps = useMemo(() => {
    let result = [...dependencies];

    if (search) {
      const lower = search.toLowerCase();
      result = result.filter(d =>
        d.name.toLowerCase().includes(lower) ||
        (d.license?.toLowerCase().includes(lower))
      );
    }

    if (showDirect) {
      result = result.filter(d => d.direct);
    }

    if (showVulnerable) {
      result = result.filter(d => (d.vulns_count || 0) > 0);
    }

    result.sort((a, b) => {
      let cmp = 0;
      switch (sortField) {
        case 'name':
          cmp = a.name.localeCompare(b.name);
          break;
        case 'license':
          cmp = (a.license || '').localeCompare(b.license || '');
          break;
        case 'vulns':
          cmp = (b.vulns_count || 0) - (a.vulns_count || 0);
          break;
      }
      return sortDirection === 'asc' ? cmp : -cmp;
    });

    return result;
  }, [dependencies, search, showDirect, showVulnerable, sortField, sortDirection]);

  // Direct dependencies for tree view
  const directDeps = useMemo(() => {
    return dependencies.filter(d => d.direct);
  }, [dependencies]);

  // Stats
  const stats = useMemo(() => {
    const licenses = new Map<string, number>();
    let vulnCount = 0;
    let deprecatedCount = 0;
    let directCount = 0;

    dependencies.forEach((d) => {
      if (d.license) {
        licenses.set(d.license, (licenses.get(d.license) || 0) + 1);
      }
      vulnCount += d.vulns_count || 0;
      if (d.health?.deprecated) deprecatedCount++;
      if (d.direct) directCount++;
    });

    return {
      total: dependencies.length,
      direct: directCount,
      transitive: dependencies.length - directCount,
      vulnerable: vulnCount,
      deprecated: deprecatedCount,
      topLicenses: Array.from(licenses.entries())
        .sort((a, b) => b[1] - a[1])
        .slice(0, 5),
    };
  }, [dependencies]);

  const toggleExpanded = (id: string) => {
    const newSet = new Set(expandedIds);
    if (newSet.has(id)) {
      newSet.delete(id);
    } else {
      newSet.add(id);
    }
    setExpandedIds(newSet);
  };

  const handleProjectChange = (newProjectId: string) => {
    router.push(`/dependencies?project=${encodeURIComponent(newProjectId)}`);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            <Package className="h-6 w-6 text-blue-500" />
            Dependencies
          </h1>
          <p className="mt-1 text-gray-400">
            View and analyze project dependencies
          </p>
        </div>
        {projectId && dependencies.length > 0 && (
          <ExportButton
            onExport={(format) => {
              const data = filteredDeps.map(d => ({
                name: d.name,
                version: d.version,
                license: d.license || '',
                direct: d.direct ? 'Yes' : 'No',
                scope: d.scope || '',
                vulns_count: d.vulns_count || 0,
                health_score: d.health?.score || '',
                deprecated: d.health?.deprecated ? 'Yes' : 'No',
              }));
              if (format === 'csv') {
                downloadCSV(data, `dependencies-${projectId.replace('/', '-')}`);
              } else {
                downloadJSON(data, `dependencies-${projectId.replace('/', '-')}`);
              }
            }}
          />
        )}
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
          <div className="grid grid-cols-5 gap-4">
            <Card className="text-center">
              <p className="text-2xl font-bold text-white">{stats.total}</p>
              <p className="text-sm text-gray-400">Total Packages</p>
            </Card>
            <Card className="text-center">
              <p className="text-2xl font-bold text-blue-500">{stats.direct}</p>
              <p className="text-sm text-gray-400">Direct</p>
            </Card>
            <Card className="text-center">
              <p className="text-2xl font-bold text-gray-400">{stats.transitive}</p>
              <p className="text-sm text-gray-400">Transitive</p>
            </Card>
            <Card className="text-center">
              <p className="text-2xl font-bold text-red-500">{stats.vulnerable}</p>
              <p className="text-sm text-gray-400">Vulnerable</p>
            </Card>
            <Card className="text-center">
              <p className="text-2xl font-bold text-yellow-500">{stats.deprecated}</p>
              <p className="text-sm text-gray-400">Deprecated</p>
            </Card>
          </div>

          {/* License Distribution */}
          {stats.topLicenses.length > 0 && (
            <Card>
              <CardTitle className="flex items-center gap-2">
                <Scale className="h-5 w-5 text-purple-500" />
                License Distribution
              </CardTitle>
              <CardContent className="mt-4">
                <div className="flex flex-wrap gap-3">
                  {stats.topLicenses.map(([license, count]) => (
                    <div key={license} className="flex items-center gap-2">
                      <LicenseBadge license={license} />
                      <span className="text-sm text-gray-400">({count})</span>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

          {/* Controls */}
          <Card>
            <CardContent>
              <div className="flex flex-col md:flex-row gap-4 items-start md:items-center justify-between">
                <div className="flex-1 max-w-md">
                  <Input
                    placeholder="Search packages..."
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    icon={<Search className="h-4 w-4" />}
                  />
                </div>

                <div className="flex items-center gap-4">
                  {/* Filters */}
                  <div className="flex items-center gap-2">
                    <Filter className="h-4 w-4 text-gray-500" />
                    <button
                      onClick={() => setShowDirect(!showDirect)}
                      className={`px-2 py-1 text-xs rounded-md transition-colors ${
                        showDirect ? 'bg-blue-600 text-white' : 'bg-gray-700 text-gray-400 hover:bg-gray-600'
                      }`}
                    >
                      Direct only
                    </button>
                    <button
                      onClick={() => setShowVulnerable(!showVulnerable)}
                      className={`px-2 py-1 text-xs rounded-md transition-colors ${
                        showVulnerable ? 'bg-red-600 text-white' : 'bg-gray-700 text-gray-400 hover:bg-gray-600'
                      }`}
                    >
                      Vulnerable
                    </button>
                  </div>

                  {/* View Mode */}
                  <div className="flex items-center gap-1 border border-gray-700 rounded-md p-1">
                    <button
                      onClick={() => setViewMode('list')}
                      className={`p-1.5 rounded ${viewMode === 'list' ? 'bg-gray-700' : 'hover:bg-gray-800'}`}
                      title="List View"
                    >
                      <LayoutList className="h-4 w-4 text-gray-400" />
                    </button>
                    <button
                      onClick={() => setViewMode('tree')}
                      className={`p-1.5 rounded ${viewMode === 'tree' ? 'bg-gray-700' : 'hover:bg-gray-800'}`}
                      title="Tree View"
                    >
                      <Network className="h-4 w-4 text-gray-400" />
                    </button>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Results */}
          <Card>
            {viewMode === 'list' ? (
              <>
                {/* List header */}
                <div className="flex items-center gap-4 px-4 py-2 border-b border-gray-700 bg-gray-800/50">
                  <div className="w-5" />
                  <div className="flex-1 text-sm font-medium text-gray-400">Package</div>
                  <div className="flex items-center gap-6">
                    <button
                      onClick={() => {
                        if (sortField === 'license') {
                          setSortDirection(d => d === 'asc' ? 'desc' : 'asc');
                        } else {
                          setSortField('license');
                          setSortDirection('asc');
                        }
                      }}
                      className="w-28 text-sm font-medium text-gray-400 text-left flex items-center gap-1"
                    >
                      License
                      {sortField === 'license' && <ArrowUpDown className="h-3 w-3" />}
                    </button>
                    <div className="w-20 text-sm font-medium text-gray-400">Health</div>
                    <button
                      onClick={() => {
                        if (sortField === 'vulns') {
                          setSortDirection(d => d === 'asc' ? 'desc' : 'asc');
                        } else {
                          setSortField('vulns');
                          setSortDirection('desc');
                        }
                      }}
                      className="w-16 text-sm font-medium text-gray-400 text-left flex items-center gap-1"
                    >
                      Vulns
                      {sortField === 'vulns' && <ArrowUpDown className="h-3 w-3" />}
                    </button>
                  </div>
                </div>

                {loading ? (
                  <div className="p-8 text-center text-gray-400">Loading dependencies...</div>
                ) : error ? (
                  <div className="p-8 text-center text-red-400">Failed to load dependencies</div>
                ) : filteredDeps.length === 0 ? (
                  <div className="p-8 text-center">
                    <Package className="h-12 w-12 text-gray-600 mx-auto mb-4" />
                    <p className="text-gray-400">
                      {dependencies.length === 0
                        ? 'No dependencies found for this project'
                        : 'No dependencies match your filters'}
                    </p>
                  </div>
                ) : (
                  <div>
                    {filteredDeps.map((dep) => (
                      <DependencyRow
                        key={`${dep.name}@${dep.version}`}
                        dep={dep}
                        expanded={expandedIds.has(dep.name)}
                        onToggle={() => toggleExpanded(dep.name)}
                      />
                    ))}
                  </div>
                )}
              </>
            ) : (
              // Tree View
              <div className="p-4">
                <h3 className="text-sm font-medium text-gray-400 mb-4">
                  Dependency Tree ({directDeps.length} direct dependencies)
                </h3>
                {loading ? (
                  <div className="text-center text-gray-400 py-8">Loading...</div>
                ) : directDeps.length === 0 ? (
                  <div className="text-center text-gray-400 py-8">No direct dependencies found</div>
                ) : (
                  <div className="font-mono text-sm">
                    {directDeps.map((dep) => (
                      <TreeNode
                        key={dep.name}
                        dep={dep}
                        depth={0}
                        allDeps={depMap}
                        expanded={expandedIds}
                        onToggle={toggleExpanded}
                      />
                    ))}
                  </div>
                )}
              </div>
            )}
          </Card>
        </>
      )}

      {!projectId && (
        <Card className="text-center py-12">
          <Package className="h-12 w-12 text-gray-600 mx-auto mb-4" />
          <p className="text-gray-400">Select a project to view its dependencies</p>
        </Card>
      )}
    </div>
  );
}

export default function DependenciesPage() {
  return (
    <MainLayout>
      <Suspense fallback={<div className="animate-pulse h-96 bg-gray-800 rounded-lg" />}>
        <DependenciesContent />
      </Suspense>
    </MainLayout>
  );
}
