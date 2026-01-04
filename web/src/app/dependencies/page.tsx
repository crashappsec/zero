'use client';

import { useState, useMemo, Suspense, useEffect } from 'react';
import { MainLayout } from '@/components/layout/Sidebar';
import { Card, CardTitle, CardContent } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import { useFetch } from '@/hooks/useApi';
import { api } from '@/lib/api';
import type { Dependency } from '@/lib/types';
import { ExportButton } from '@/components/ui/ExportButton';
import { downloadCSV, downloadJSON } from '@/lib/export';
import { ProjectFilter } from '@/components/ui/ProjectFilter';
import {
  Search,
  Package,
  ChevronRight,
  ChevronDown,
  AlertTriangle,
  CheckCircle,
  Filter,
  LayoutList,
  ArrowUpDown,
  Scale,
} from 'lucide-react';

type SortField = 'name' | 'license' | 'vulns' | 'project';

interface DepWithProject extends Dependency {
  projectId: string;
}

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

function DependencyRow({ dep, expanded, onToggle }: { dep: DepWithProject; expanded: boolean; onToggle: () => void }) {
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
          <Badge variant="default" className="text-xs">{dep.projectId}</Badge>
        </div>

        <div className="flex items-center gap-6">
          <div className="w-28">
            <LicenseBadge license={dep.license} />
          </div>
          <div className="w-20">
            <HealthIndicator health={dep.health} />
          </div>
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

function DependenciesContent() {
  const [search, setSearch] = useState('');
  const [showDirect, setShowDirect] = useState(false);
  const [showVulnerable, setShowVulnerable] = useState(false);
  const [sortField, setSortField] = useState<SortField>('name');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc');
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());
  const [selectedProjects, setSelectedProjects] = useState<string[]>([]);
  const [allDeps, setAllDeps] = useState<DepWithProject[]>([]);
  const [loading, setLoading] = useState(true);

  // Fetch projects
  const { data: projectsData } = useFetch(() => api.projects.list(), []);
  const projects = projectsData?.data || [];

  // Load dependencies for all projects in parallel
  useEffect(() => {
    async function loadAllDeps() {
      if (projects.length === 0) return;

      setLoading(true);

      // Fetch all projects in parallel for better performance
      const results = await Promise.all(
        projects.map(async (project) => {
          try {
            const data = await api.analysis.dependencies(project.id);
            if (data?.data) {
              return data.data.map(dep => ({ ...dep, projectId: project.id }));
            }
          } catch {
            // Skip projects without dependency data
          }
          return [];
        })
      );

      setAllDeps(results.flat());
      setLoading(false);
    }

    loadAllDeps();
  }, [projects]);

  // Filter by selected projects
  const projectFilteredDeps = useMemo(() => {
    if (selectedProjects.length === 0) return allDeps;
    return allDeps.filter(d => selectedProjects.includes(d.projectId));
  }, [allDeps, selectedProjects]);

  // Filter and sort
  const filteredDeps = useMemo(() => {
    let result = [...projectFilteredDeps];

    if (search) {
      const lower = search.toLowerCase();
      result = result.filter(d =>
        d.name.toLowerCase().includes(lower) ||
        d.projectId.toLowerCase().includes(lower) ||
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
        case 'project':
          cmp = a.projectId.localeCompare(b.projectId);
          break;
      }
      return sortDirection === 'asc' ? cmp : -cmp;
    });

    return result;
  }, [projectFilteredDeps, search, showDirect, showVulnerable, sortField, sortDirection]);

  // Stats
  const stats = useMemo(() => {
    const licenses = new Map<string, number>();
    let vulnCount = 0;
    let deprecatedCount = 0;
    let directCount = 0;

    projectFilteredDeps.forEach((d) => {
      if (d.license) {
        licenses.set(d.license, (licenses.get(d.license) || 0) + 1);
      }
      vulnCount += d.vulns_count || 0;
      if (d.health?.deprecated) deprecatedCount++;
      if (d.direct) directCount++;
    });

    return {
      total: projectFilteredDeps.length,
      direct: directCount,
      transitive: projectFilteredDeps.length - directCount,
      vulnerable: vulnCount,
      deprecated: deprecatedCount,
      topLicenses: Array.from(licenses.entries())
        .sort((a, b) => b[1] - a[1])
        .slice(0, 5),
    };
  }, [projectFilteredDeps]);

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

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            <Package className="h-6 w-6 text-blue-500" />
            Dependencies
          </h1>
          <p className="mt-1 text-gray-400">
            Package dependencies across all projects
          </p>
        </div>
        <div className="flex items-center gap-2">
          {projects.length > 0 && (
            <ProjectFilter
              projects={projects}
              selectedProjects={selectedProjects}
              onChange={setSelectedProjects}
            />
          )}
          {projectFilteredDeps.length > 0 && (
            <ExportButton
              onExport={(format) => {
                const data = filteredDeps.map(d => ({
                  project: d.projectId,
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
                  downloadCSV(data, 'dependencies');
                } else {
                  downloadJSON(data, 'dependencies');
                }
              }}
            />
          )}
        </div>
      </div>

      {loading ? (
        <div className="space-y-4">
          {[1, 2, 3].map((i) => (
            <Card key={i} className="animate-pulse h-24" />
          ))}
        </div>
      ) : allDeps.length > 0 ? (
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
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Results */}
          <Card>
            {/* List header */}
            <div className="flex items-center gap-4 px-4 py-2 border-b border-gray-700 bg-gray-800/50">
              <div className="w-5" />
              <div className="flex-1 text-sm font-medium text-gray-400">Package</div>
              <div className="flex items-center gap-6">
                <button
                  onClick={() => toggleSort('license')}
                  className="w-28 text-sm font-medium text-gray-400 text-left flex items-center gap-1"
                >
                  License
                  {sortField === 'license' && <ArrowUpDown className="h-3 w-3" />}
                </button>
                <div className="w-20 text-sm font-medium text-gray-400">Health</div>
                <button
                  onClick={() => toggleSort('vulns')}
                  className="w-16 text-sm font-medium text-gray-400 text-left flex items-center gap-1"
                >
                  Vulns
                  {sortField === 'vulns' && <ArrowUpDown className="h-3 w-3" />}
                </button>
              </div>
            </div>

            {filteredDeps.length === 0 ? (
              <div className="p-8 text-center">
                <Package className="h-12 w-12 text-gray-600 mx-auto mb-4" />
                <p className="text-gray-400">No dependencies match your filters</p>
              </div>
            ) : (
              <div>
                {filteredDeps.slice(0, 100).map((dep, i) => (
                  <DependencyRow
                    key={`${dep.projectId}-${dep.name}@${dep.version}-${i}`}
                    dep={dep}
                    expanded={expandedIds.has(`${dep.projectId}-${dep.name}`)}
                    onToggle={() => toggleExpanded(`${dep.projectId}-${dep.name}`)}
                  />
                ))}
                {filteredDeps.length > 100 && (
                  <div className="p-4 text-center text-gray-400 text-sm">
                    Showing 100 of {filteredDeps.length} dependencies. Use filters to narrow down.
                  </div>
                )}
              </div>
            )}
          </Card>
        </>
      ) : (
        <Card className="text-center py-12">
          <Package className="h-12 w-12 text-gray-600 mx-auto mb-4" />
          <p className="text-gray-400">No dependencies found</p>
          <p className="text-sm text-gray-500 mt-1">Run scans with the supply-chain scanner</p>
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
