'use client';

import { useState, useMemo, Suspense, useEffect } from 'react';
import { MainLayout } from '@/components/layout/Sidebar';
import { Card, CardTitle, CardContent } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { useFetch } from '@/hooks/useApi';
import { api } from '@/lib/api';
import {
  Users,
  User,
  GitCommit,
  AlertTriangle,
  TrendingUp,
  FileCode,
  Activity,
  ChevronDown,
  ChevronRight,
} from 'lucide-react';
import { ProjectFilter } from '@/components/ui/ProjectFilter';

interface Contributor {
  name: string;
  email: string;
  commits: number;
  lines_added: number;
  lines_removed: number;
  first_commit: string;
  last_commit: string;
  files_touched: number;
}

interface ProjectOwnership {
  projectId: string;
  bus_factor: number;
  total_contributors: number;
  active_contributors: number;
  orphan_files: number;
  top_contributors: Contributor[];
  churn_hotspots: Array<{ file: string; changes: number }>;
}

function BusFactorIndicator({ busFactor }: { busFactor: number }) {
  const getColor = (bf: number) => {
    if (bf <= 1) return 'text-red-500';
    if (bf <= 2) return 'text-orange-500';
    if (bf <= 3) return 'text-yellow-500';
    return 'text-green-500';
  };

  return (
    <span className={`font-bold ${getColor(busFactor)}`}>{busFactor}</span>
  );
}

function ProjectOwnershipCard({ data, expanded, onToggle }: {
  data: ProjectOwnership;
  expanded: boolean;
  onToggle: () => void;
}) {
  const riskLevel = data.bus_factor <= 1 ? 'Critical' :
                    data.bus_factor <= 2 ? 'High' :
                    data.bus_factor <= 3 ? 'Medium' : 'Low';

  return (
    <Card className="overflow-hidden">
      <button
        onClick={onToggle}
        className="w-full flex items-center justify-between p-4 hover:bg-gray-800/50 transition-colors"
      >
        <div className="flex items-center gap-4">
          {expanded ? (
            <ChevronDown className="h-4 w-4 text-gray-500" />
          ) : (
            <ChevronRight className="h-4 w-4 text-gray-500" />
          )}
          <div>
            <h3 className="font-medium text-white text-left">{data.projectId}</h3>
            <div className="flex items-center gap-4 mt-1 text-sm text-gray-400">
              <span>Bus Factor: <BusFactorIndicator busFactor={data.bus_factor} /></span>
              <span>{data.total_contributors} contributors</span>
              <span>{data.active_contributors} active (90d)</span>
            </div>
          </div>
        </div>
        <Badge variant={data.bus_factor <= 2 ? 'error' : data.bus_factor <= 3 ? 'warning' : 'success'}>
          {riskLevel} Risk
        </Badge>
      </button>

      {expanded && (
        <div className="border-t border-gray-700 p-4 space-y-4">
          {/* Top Contributors */}
          {data.top_contributors.length > 0 && (
            <div>
              <h4 className="text-sm font-medium text-gray-400 mb-2">Top Contributors</h4>
              <div className="space-y-2">
                {data.top_contributors.slice(0, 5).map((c, i) => (
                  <div key={c.email} className="flex items-center justify-between text-sm">
                    <div className="flex items-center gap-2">
                      <span className="text-gray-500 w-4">{i + 1}.</span>
                      <User className="h-4 w-4 text-gray-500" />
                      <span className="text-white">{c.name}</span>
                    </div>
                    <div className="flex items-center gap-4 text-gray-400">
                      <span>{c.commits} commits</span>
                      <span className="text-green-500">+{c.lines_added.toLocaleString()}</span>
                      <span className="text-red-500">-{c.lines_removed.toLocaleString()}</span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Churn Hotspots */}
          {data.churn_hotspots.length > 0 && (
            <div>
              <h4 className="text-sm font-medium text-gray-400 mb-2">Churn Hotspots</h4>
              <div className="flex flex-wrap gap-2">
                {data.churn_hotspots.slice(0, 5).map((h) => (
                  <Badge key={h.file} variant="warning">
                    {h.file.split('/').pop()} ({h.changes} changes)
                  </Badge>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </Card>
  );
}

function OwnershipContent() {
  const [expandedProjects, setExpandedProjects] = useState<Set<string>>(new Set());
  const [ownershipData, setOwnershipData] = useState<ProjectOwnership[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedProjects, setSelectedProjects] = useState<string[]>([]);

  const { data: projectsData } = useFetch(() => api.projects.list(), []);
  const projects = projectsData?.data || [];

  // Load ownership data for all projects in parallel
  useEffect(() => {
    async function loadOwnershipData() {
      if (projects.length === 0) return;

      setLoading(true);

      // Fetch all projects in parallel for better performance
      const results = await Promise.all(
        projects.map(async (project) => {
          try {
            const data = await api.analysis.raw(project.id, 'code-ownership') as any;
            if (data?.findings) {
              const findings = data.findings;
              return {
                projectId: project.id,
                bus_factor: findings.bus_factor?.bus_factor || 0,
                total_contributors: findings.contributors?.total || 0,
                active_contributors: findings.contributors?.active_last_90_days || 0,
                orphan_files: findings.orphans?.count || 0,
                top_contributors: findings.contributors?.top || [],
                churn_hotspots: findings.churn?.hotspots || [],
              } as ProjectOwnership;
            }
          } catch {
            // Skip projects without ownership data
          }
          return null;
        })
      );

      setOwnershipData(results.filter((r): r is ProjectOwnership => r !== null));
      setLoading(false);
    }

    loadOwnershipData();
  }, [projects]);

  const toggleProject = (id: string) => {
    const newSet = new Set(expandedProjects);
    if (newSet.has(id)) {
      newSet.delete(id);
    } else {
      newSet.add(id);
    }
    setExpandedProjects(newSet);
  };

  // Filter data based on selected projects
  const filteredData = useMemo(() => {
    if (selectedProjects.length === 0) return ownershipData;
    return ownershipData.filter(d => selectedProjects.includes(d.projectId));
  }, [ownershipData, selectedProjects]);

  // Aggregate stats
  const stats = useMemo(() => {
    if (filteredData.length === 0) return null;

    const criticalRisk = filteredData.filter(d => d.bus_factor <= 1).length;
    const highRisk = filteredData.filter(d => d.bus_factor === 2).length;
    const avgBusFactor = filteredData.reduce((sum, d) => sum + d.bus_factor, 0) / filteredData.length;
    const totalContributors = new Set(
      filteredData.flatMap(d => d.top_contributors.map(c => c.email))
    ).size;
    const totalOrphans = filteredData.reduce((sum, d) => sum + d.orphan_files, 0);

    return {
      criticalRisk,
      highRisk,
      avgBusFactor: avgBusFactor.toFixed(1),
      totalContributors,
      totalOrphans,
      projectsAnalyzed: filteredData.length,
    };
  }, [filteredData]);

  // Sort by bus factor (lowest first = highest risk)
  const sortedData = useMemo(() => {
    return [...filteredData].sort((a, b) => a.bus_factor - b.bus_factor);
  }, [filteredData]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            <Users className="h-6 w-6 text-purple-500" />
            Code Ownership
          </h1>
          <p className="mt-1 text-gray-400">
            Bus factor, contributor analysis, and ownership distribution across all projects
          </p>
        </div>
        {projects.length > 0 && (
          <ProjectFilter
            projects={projects}
            selectedProjects={selectedProjects}
            onChange={setSelectedProjects}
          />
        )}
      </div>

      {loading ? (
        <div className="space-y-4">
          {[1, 2, 3].map((i) => (
            <Card key={i} className="animate-pulse h-24" />
          ))}
        </div>
      ) : stats ? (
        <>
          {/* Aggregate Stats */}
          <div className="grid gap-4 md:grid-cols-5">
            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-red-600/20">
                  <AlertTriangle className="h-6 w-6 text-red-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Critical Risk</p>
                  <p className="text-2xl font-bold text-red-500">{stats.criticalRisk}</p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-orange-600/20">
                  <AlertTriangle className="h-6 w-6 text-orange-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">High Risk</p>
                  <p className="text-2xl font-bold text-orange-500">{stats.highRisk}</p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-purple-600/20">
                  <TrendingUp className="h-6 w-6 text-purple-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Avg Bus Factor</p>
                  <p className="text-2xl font-bold text-white">{stats.avgBusFactor}</p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-blue-600/20">
                  <Users className="h-6 w-6 text-blue-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Contributors</p>
                  <p className="text-2xl font-bold text-white">{stats.totalContributors}</p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-yellow-600/20">
                  <FileCode className="h-6 w-6 text-yellow-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Orphan Files</p>
                  <p className="text-2xl font-bold text-white">{stats.totalOrphans}</p>
                </div>
              </div>
            </Card>
          </div>

          {/* Projects List */}
          <div>
            <h2 className="text-lg font-semibold text-white mb-4">
              Projects ({stats.projectsAnalyzed} analyzed)
            </h2>
            <div className="space-y-3">
              {sortedData.map((data) => (
                <ProjectOwnershipCard
                  key={data.projectId}
                  data={data}
                  expanded={expandedProjects.has(data.projectId)}
                  onToggle={() => toggleProject(data.projectId)}
                />
              ))}
            </div>
          </div>
        </>
      ) : (
        <Card className="text-center py-12">
          <Users className="h-12 w-12 text-gray-600 mx-auto mb-4" />
          <p className="text-gray-400">No code ownership data available</p>
          <p className="text-sm text-gray-500 mt-1">Run scans with the code-ownership scanner</p>
        </Card>
      )}
    </div>
  );
}

export default function OwnershipPage() {
  return (
    <MainLayout>
      <Suspense fallback={<div className="animate-pulse h-96 bg-gray-800 rounded-lg" />}>
        <OwnershipContent />
      </Suspense>
    </MainLayout>
  );
}
