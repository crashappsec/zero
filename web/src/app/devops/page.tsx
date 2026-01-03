'use client';

import { useState, useMemo, Suspense, useEffect } from 'react';
import { MainLayout } from '@/components/layout/Sidebar';
import { Card, CardTitle, CardContent } from '@/components/ui/Card';
import { Badge, SeverityBadge } from '@/components/ui/Badge';
import { useFetch } from '@/hooks/useApi';
import { api } from '@/lib/api';
import {
  Server,
  Container,
  GitBranch,
  Workflow,
  TrendingUp,
  AlertTriangle,
  CheckCircle,
  FileCode,
  ChevronDown,
  ChevronRight,
} from 'lucide-react';
import { ProjectFilter } from '@/components/ui/ProjectFilter';

interface ProjectDevOps {
  projectId: string;
  dora: {
    deployment_frequency: string;
    lead_time: string;
    change_failure_rate: number;
    mttr: string;
    performance_level: string;
  } | null;
  iac_count: number;
  container_count: number;
  gha_count: number;
  git_stats: {
    total_commits: number;
    branches: number;
    contributors: number;
  };
}

function DORABadge({ level }: { level: string }) {
  const variant = level === 'elite' ? 'success' :
                  level === 'high' ? 'info' :
                  level === 'medium' ? 'warning' : 'error';
  return <Badge variant={variant}>{level} performer</Badge>;
}

function ProjectDevOpsCard({ data, expanded, onToggle }: {
  data: ProjectDevOps;
  expanded: boolean;
  onToggle: () => void;
}) {
  const totalIssues = data.iac_count + data.container_count + data.gha_count;

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
              <span>{data.git_stats.total_commits} commits</span>
              <span>{data.git_stats.branches} branches</span>
              {totalIssues > 0 && (
                <span className="text-yellow-500">{totalIssues} issues</span>
              )}
            </div>
          </div>
        </div>
        {data.dora && <DORABadge level={data.dora.performance_level} />}
      </button>

      {expanded && (
        <div className="border-t border-gray-700 p-4 space-y-4">
          {/* DORA Metrics */}
          {data.dora && (
            <div>
              <h4 className="text-sm font-medium text-gray-400 mb-2">DORA Metrics</h4>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
                <div className="p-2 bg-gray-800/50 rounded">
                  <p className="text-xs text-gray-500">Deploy Freq</p>
                  <p className="text-sm text-white">{data.dora.deployment_frequency}</p>
                </div>
                <div className="p-2 bg-gray-800/50 rounded">
                  <p className="text-xs text-gray-500">Lead Time</p>
                  <p className="text-sm text-white">{data.dora.lead_time}</p>
                </div>
                <div className="p-2 bg-gray-800/50 rounded">
                  <p className="text-xs text-gray-500">Failure Rate</p>
                  <p className={`text-sm ${data.dora.change_failure_rate < 15 ? 'text-green-500' : 'text-yellow-500'}`}>
                    {data.dora.change_failure_rate}%
                  </p>
                </div>
                <div className="p-2 bg-gray-800/50 rounded">
                  <p className="text-xs text-gray-500">MTTR</p>
                  <p className="text-sm text-white">{data.dora.mttr}</p>
                </div>
              </div>
            </div>
          )}

          {/* Issues Summary */}
          {totalIssues > 0 && (
            <div>
              <h4 className="text-sm font-medium text-gray-400 mb-2">Issues Found</h4>
              <div className="flex flex-wrap gap-2">
                {data.iac_count > 0 && (
                  <Badge variant="warning">{data.iac_count} IaC issues</Badge>
                )}
                {data.container_count > 0 && (
                  <Badge variant="warning">{data.container_count} container issues</Badge>
                )}
                {data.gha_count > 0 && (
                  <Badge variant="warning">{data.gha_count} GitHub Actions issues</Badge>
                )}
              </div>
            </div>
          )}
        </div>
      )}
    </Card>
  );
}

function DevOpsContent() {
  const [expandedProjects, setExpandedProjects] = useState<Set<string>>(new Set());
  const [devopsData, setDevopsData] = useState<ProjectDevOps[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedProjects, setSelectedProjects] = useState<string[]>([]);

  const { data: projectsData } = useFetch(() => api.projects.list(), []);
  const projects = projectsData?.data || [];

  useEffect(() => {
    async function loadDevOpsData() {
      if (projects.length === 0) return;

      setLoading(true);
      const results: ProjectDevOps[] = [];

      for (const project of projects) {
        try {
          const data = await api.analysis.raw(project.id, 'devops') as any;
          if (data?.findings) {
            const findings = data.findings;
            results.push({
              projectId: project.id,
              dora: findings.dora || null,
              iac_count: findings.iac?.findings?.length || 0,
              container_count: findings.containers?.findings?.length || 0,
              gha_count: findings.github_actions?.findings?.length || 0,
              git_stats: findings.git || { total_commits: 0, branches: 0, contributors: 0 },
            });
          }
        } catch {
          // Skip projects without devops data
        }
      }

      setDevopsData(results);
      setLoading(false);
    }

    loadDevOpsData();
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
    if (selectedProjects.length === 0) return devopsData;
    return devopsData.filter(d => selectedProjects.includes(d.projectId));
  }, [devopsData, selectedProjects]);

  // Aggregate stats
  const stats = useMemo(() => {
    if (filteredData.length === 0) return null;

    const doraLevels = { elite: 0, high: 0, medium: 0, low: 0 };
    let totalIaC = 0;
    let totalContainer = 0;
    let totalGHA = 0;
    let totalCommits = 0;

    filteredData.forEach((p) => {
      if (p.dora) {
        const level = p.dora.performance_level as keyof typeof doraLevels;
        if (level in doraLevels) doraLevels[level]++;
      }
      totalIaC += p.iac_count;
      totalContainer += p.container_count;
      totalGHA += p.gha_count;
      totalCommits += p.git_stats.total_commits;
    });

    return {
      doraLevels,
      totalIaC,
      totalContainer,
      totalGHA,
      totalCommits,
      projectsAnalyzed: filteredData.length,
    };
  }, [filteredData]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            <Server className="h-6 w-6 text-green-500" />
            DevOps & Infrastructure
          </h1>
          <p className="mt-1 text-gray-400">
            DORA metrics, IaC, containers, and CI/CD analysis across all projects
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
          <div className="grid gap-4 md:grid-cols-4">
            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-green-600/20">
                  <TrendingUp className="h-6 w-6 text-green-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Elite Performers</p>
                  <p className="text-2xl font-bold text-green-500">{stats.doraLevels.elite}</p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-orange-600/20">
                  <FileCode className="h-6 w-6 text-orange-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">IaC Issues</p>
                  <p className="text-2xl font-bold text-white">{stats.totalIaC}</p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-blue-600/20">
                  <Container className="h-6 w-6 text-blue-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Container Issues</p>
                  <p className="text-2xl font-bold text-white">{stats.totalContainer}</p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-purple-600/20">
                  <Workflow className="h-6 w-6 text-purple-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">GHA Issues</p>
                  <p className="text-2xl font-bold text-white">{stats.totalGHA}</p>
                </div>
              </div>
            </Card>
          </div>

          {/* DORA Distribution */}
          <Card>
            <CardTitle className="flex items-center gap-2">
              <TrendingUp className="h-5 w-5 text-blue-500" />
              DORA Performance Distribution
            </CardTitle>
            <CardContent className="mt-4">
              <div className="grid grid-cols-4 gap-4 text-center">
                <div className="p-4 bg-green-600/10 rounded-lg">
                  <p className="text-2xl font-bold text-green-500">{stats.doraLevels.elite}</p>
                  <p className="text-sm text-gray-400">Elite</p>
                </div>
                <div className="p-4 bg-blue-600/10 rounded-lg">
                  <p className="text-2xl font-bold text-blue-500">{stats.doraLevels.high}</p>
                  <p className="text-sm text-gray-400">High</p>
                </div>
                <div className="p-4 bg-yellow-600/10 rounded-lg">
                  <p className="text-2xl font-bold text-yellow-500">{stats.doraLevels.medium}</p>
                  <p className="text-sm text-gray-400">Medium</p>
                </div>
                <div className="p-4 bg-red-600/10 rounded-lg">
                  <p className="text-2xl font-bold text-red-500">{stats.doraLevels.low}</p>
                  <p className="text-sm text-gray-400">Low</p>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Projects List */}
          <div>
            <h2 className="text-lg font-semibold text-white mb-4">
              Projects ({stats.projectsAnalyzed} analyzed)
            </h2>
            <div className="space-y-3">
              {filteredData.map((data) => (
                <ProjectDevOpsCard
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
          <Server className="h-12 w-12 text-gray-600 mx-auto mb-4" />
          <p className="text-gray-400">No DevOps data available</p>
          <p className="text-sm text-gray-500 mt-1">Run scans with the devops scanner</p>
        </Card>
      )}
    </div>
  );
}

export default function DevOpsPage() {
  return (
    <MainLayout>
      <Suspense fallback={<div className="animate-pulse h-96 bg-gray-800 rounded-lg" />}>
        <DevOpsContent />
      </Suspense>
    </MainLayout>
  );
}
