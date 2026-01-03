'use client';

import { useState, useMemo, Suspense } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
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
  Clock,
  TrendingUp,
  AlertTriangle,
  CheckCircle,
  FileCode,
  Shield,
} from 'lucide-react';

interface DORAMetrics {
  deployment_frequency: string;
  lead_time: string;
  change_failure_rate: number;
  mttr: string;
  performance_level: string;
}

interface IaCFinding {
  file: string;
  line: number;
  rule: string;
  severity: string;
  message: string;
}

interface ContainerFinding {
  image: string;
  file: string;
  issue: string;
  severity: string;
}

interface GHAFinding {
  workflow: string;
  job: string;
  issue: string;
  severity: string;
}

interface DevOpsData {
  dora: DORAMetrics | null;
  iac_findings: IaCFinding[];
  container_findings: ContainerFinding[];
  gha_findings: GHAFinding[];
  git_stats: {
    total_commits: number;
    branches: number;
    contributors: number;
  };
}

function DORACard({ metrics }: { metrics: DORAMetrics }) {
  const levelColors: Record<string, string> = {
    elite: 'text-green-500',
    high: 'text-blue-500',
    medium: 'text-yellow-500',
    low: 'text-red-500',
  };

  return (
    <Card>
      <CardTitle className="flex items-center gap-2">
        <TrendingUp className="h-5 w-5 text-blue-500" />
        DORA Metrics
        <Badge variant={metrics.performance_level === 'elite' ? 'success' :
                       metrics.performance_level === 'high' ? 'info' :
                       metrics.performance_level === 'medium' ? 'warning' : 'error'}>
          {metrics.performance_level} performer
        </Badge>
      </CardTitle>
      <CardContent className="mt-4">
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <div className="p-4 bg-gray-800/50 rounded-lg">
            <p className="text-sm text-gray-400">Deployment Frequency</p>
            <p className="text-xl font-bold text-white mt-1">{metrics.deployment_frequency}</p>
          </div>
          <div className="p-4 bg-gray-800/50 rounded-lg">
            <p className="text-sm text-gray-400">Lead Time for Changes</p>
            <p className="text-xl font-bold text-white mt-1">{metrics.lead_time}</p>
          </div>
          <div className="p-4 bg-gray-800/50 rounded-lg">
            <p className="text-sm text-gray-400">Change Failure Rate</p>
            <p className={`text-xl font-bold mt-1 ${
              metrics.change_failure_rate < 15 ? 'text-green-500' :
              metrics.change_failure_rate < 30 ? 'text-yellow-500' : 'text-red-500'
            }`}>
              {metrics.change_failure_rate}%
            </p>
          </div>
          <div className="p-4 bg-gray-800/50 rounded-lg">
            <p className="text-sm text-gray-400">Mean Time to Recovery</p>
            <p className="text-xl font-bold text-white mt-1">{metrics.mttr}</p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

function FindingsTable<T extends { severity: string }>({
  findings,
  columns,
  renderRow
}: {
  findings: T[];
  columns: string[];
  renderRow: (finding: T) => React.ReactNode;
}) {
  if (findings.length === 0) {
    return (
      <div className="p-8 text-center text-gray-400">
        <CheckCircle className="h-12 w-12 mx-auto mb-4 text-green-500" />
        No issues found
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full">
        <thead>
          <tr className="border-b border-gray-700">
            {columns.map((col) => (
              <th key={col} className="px-4 py-2 text-left text-xs font-medium text-gray-400 uppercase">
                {col}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {findings.map((finding, i) => (
            <tr key={i} className="border-b border-gray-700/50 hover:bg-gray-800/50">
              {renderRow(finding)}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function DevOpsContent() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const projectId = searchParams.get('project');
  const [activeTab, setActiveTab] = useState<'overview' | 'iac' | 'containers' | 'actions'>('overview');

  const { data: projectsData } = useFetch(() => api.projects.list(), []);
  const projects = projectsData?.data || [];

  const { data: devopsData, loading, error } = useFetch(
    () => projectId ? api.analysis.raw(projectId, 'devops') as Promise<any> : Promise.resolve(null),
    [projectId]
  );

  const devops = useMemo(() => {
    if (!devopsData?.findings) return null;
    const findings = devopsData.findings;
    return {
      dora: findings.dora || null,
      iac_findings: findings.iac?.findings || [],
      container_findings: findings.containers?.findings || [],
      gha_findings: findings.github_actions?.findings || [],
      git_stats: findings.git || { total_commits: 0, branches: 0, contributors: 0 },
    } as DevOpsData;
  }, [devopsData]);

  const handleProjectChange = (newProjectId: string) => {
    router.push(`/devops?project=${encodeURIComponent(newProjectId)}`);
  };

  const tabs = [
    { id: 'overview', label: 'Overview', icon: Server },
    { id: 'iac', label: `IaC (${devops?.iac_findings.length || 0})`, icon: FileCode },
    { id: 'containers', label: `Containers (${devops?.container_findings.length || 0})`, icon: Container },
    { id: 'actions', label: `Actions (${devops?.gha_findings.length || 0})`, icon: Workflow },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-white flex items-center gap-2">
          <Server className="h-6 w-6 text-green-500" />
          DevOps & Infrastructure
        </h1>
        <p className="mt-1 text-gray-400">
          DORA metrics, infrastructure as code, containers, and CI/CD analysis
        </p>
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
                <option key={p.id} value={p.id}>{p.id}</option>
              ))}
            </select>
          </div>
        </CardContent>
      </Card>

      {projectId && devops && (
        <>
          {/* Tabs */}
          <div className="flex gap-2 border-b border-gray-700">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id as any)}
                className={`flex items-center gap-2 px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === tab.id
                    ? 'text-green-500 border-green-500'
                    : 'text-gray-400 border-transparent hover:text-white'
                }`}
              >
                <tab.icon className="h-4 w-4" />
                {tab.label}
              </button>
            ))}
          </div>

          {/* Overview */}
          {activeTab === 'overview' && (
            <div className="space-y-6">
              {/* Stats */}
              <div className="grid gap-4 md:grid-cols-4">
                <Card>
                  <div className="flex items-center gap-3">
                    <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-blue-600/20">
                      <GitBranch className="h-6 w-6 text-blue-500" />
                    </div>
                    <div>
                      <p className="text-sm text-gray-400">Commits</p>
                      <p className="text-2xl font-bold text-white">{devops.git_stats.total_commits}</p>
                    </div>
                  </div>
                </Card>
                <Card>
                  <div className="flex items-center gap-3">
                    <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-purple-600/20">
                      <GitBranch className="h-6 w-6 text-purple-500" />
                    </div>
                    <div>
                      <p className="text-sm text-gray-400">Branches</p>
                      <p className="text-2xl font-bold text-white">{devops.git_stats.branches}</p>
                    </div>
                  </div>
                </Card>
                <Card>
                  <div className="flex items-center gap-3">
                    <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-yellow-600/20">
                      <AlertTriangle className="h-6 w-6 text-yellow-500" />
                    </div>
                    <div>
                      <p className="text-sm text-gray-400">IaC Issues</p>
                      <p className="text-2xl font-bold text-white">{devops.iac_findings.length}</p>
                    </div>
                  </div>
                </Card>
                <Card>
                  <div className="flex items-center gap-3">
                    <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-red-600/20">
                      <Container className="h-6 w-6 text-red-500" />
                    </div>
                    <div>
                      <p className="text-sm text-gray-400">Container Issues</p>
                      <p className="text-2xl font-bold text-white">{devops.container_findings.length}</p>
                    </div>
                  </div>
                </Card>
              </div>

              {/* DORA */}
              {devops.dora && <DORACard metrics={devops.dora} />}
            </div>
          )}

          {/* IaC Findings */}
          {activeTab === 'iac' && (
            <Card>
              <CardTitle className="flex items-center gap-2">
                <FileCode className="h-5 w-5 text-orange-500" />
                Infrastructure as Code Findings
              </CardTitle>
              <CardContent className="mt-4 p-0">
                <FindingsTable
                  findings={devops.iac_findings}
                  columns={['Severity', 'File', 'Rule', 'Message']}
                  renderRow={(f: IaCFinding) => (
                    <>
                      <td className="px-4 py-3"><SeverityBadge severity={f.severity} /></td>
                      <td className="px-4 py-3 text-sm text-white font-mono">{f.file}:{f.line}</td>
                      <td className="px-4 py-3 text-sm text-gray-400">{f.rule}</td>
                      <td className="px-4 py-3 text-sm text-gray-300">{f.message}</td>
                    </>
                  )}
                />
              </CardContent>
            </Card>
          )}

          {/* Container Findings */}
          {activeTab === 'containers' && (
            <Card>
              <CardTitle className="flex items-center gap-2">
                <Container className="h-5 w-5 text-blue-500" />
                Container Findings
              </CardTitle>
              <CardContent className="mt-4 p-0">
                <FindingsTable
                  findings={devops.container_findings}
                  columns={['Severity', 'Image', 'File', 'Issue']}
                  renderRow={(f: ContainerFinding) => (
                    <>
                      <td className="px-4 py-3"><SeverityBadge severity={f.severity} /></td>
                      <td className="px-4 py-3 text-sm text-white">{f.image}</td>
                      <td className="px-4 py-3 text-sm text-gray-400 font-mono">{f.file}</td>
                      <td className="px-4 py-3 text-sm text-gray-300">{f.issue}</td>
                    </>
                  )}
                />
              </CardContent>
            </Card>
          )}

          {/* GitHub Actions Findings */}
          {activeTab === 'actions' && (
            <Card>
              <CardTitle className="flex items-center gap-2">
                <Workflow className="h-5 w-5 text-purple-500" />
                GitHub Actions Findings
              </CardTitle>
              <CardContent className="mt-4 p-0">
                <FindingsTable
                  findings={devops.gha_findings}
                  columns={['Severity', 'Workflow', 'Job', 'Issue']}
                  renderRow={(f: GHAFinding) => (
                    <>
                      <td className="px-4 py-3"><SeverityBadge severity={f.severity} /></td>
                      <td className="px-4 py-3 text-sm text-white">{f.workflow}</td>
                      <td className="px-4 py-3 text-sm text-gray-400">{f.job}</td>
                      <td className="px-4 py-3 text-sm text-gray-300">{f.issue}</td>
                    </>
                  )}
                />
              </CardContent>
            </Card>
          )}
        </>
      )}

      {projectId && loading && (
        <Card className="p-8 text-center text-gray-400">Loading DevOps data...</Card>
      )}

      {projectId && error && (
        <Card className="p-8 text-center text-red-400">
          No DevOps data available. Run a scan with the devops scanner.
        </Card>
      )}

      {!projectId && (
        <Card className="text-center py-12">
          <Server className="h-12 w-12 text-gray-600 mx-auto mb-4" />
          <p className="text-gray-400">Select a project to view DevOps analysis</p>
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
