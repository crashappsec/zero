'use client';

import { useState, Suspense } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { MainLayout } from '@/components/layout/Sidebar';
import { Card, CardTitle, CardContent } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge, StatusBadge } from '@/components/ui/Badge';
import { useFetch } from '@/hooks/useApi';
import { api } from '@/lib/api';
import type { Project } from '@/lib/types';
import {
  FileText,
  Download,
  ExternalLink,
  RefreshCw,
  CheckCircle,
  AlertTriangle,
  Clock,
  Loader2,
  FolderOpen,
} from 'lucide-react';

interface ReportInfo {
  project_id: string;
  generated_at: string;
  url: string;
  status: 'ready' | 'generating' | 'error';
  error?: string;
}

function ReportCard({
  project,
  report,
  onGenerate,
  generating,
}: {
  project: Project;
  report?: ReportInfo;
  onGenerate: () => void;
  generating: boolean;
}) {
  const hasReport = report?.status === 'ready';

  return (
    <Card className={hasReport ? 'border-green-700/50' : ''}>
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <div className={`p-2 rounded-lg ${hasReport ? 'bg-green-600/20' : 'bg-gray-700'}`}>
            <FileText className={`h-5 w-5 ${hasReport ? 'text-green-500' : 'text-gray-400'}`} />
          </div>
          <div>
            <h3 className="font-medium text-white">{project.id}</h3>
            <div className="flex items-center gap-2 mt-1">
              {hasReport ? (
                <>
                  <Badge variant="success">Ready</Badge>
                  <span className="text-xs text-gray-500">
                    Generated {new Date(report.generated_at).toLocaleDateString()}
                  </span>
                </>
              ) : report?.status === 'generating' || generating ? (
                <Badge variant="info">Generating...</Badge>
              ) : report?.status === 'error' ? (
                <Badge variant="error">Error</Badge>
              ) : (
                <Badge variant="default">No report</Badge>
              )}
            </div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {hasReport && (
            <>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => window.open(report.url, '_blank')}
                icon={<ExternalLink className="h-4 w-4" />}
              >
                View
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={onGenerate}
                loading={generating}
                icon={<RefreshCw className="h-4 w-4" />}
              >
                Regenerate
              </Button>
            </>
          )}
          {!hasReport && (
            <Button
              onClick={onGenerate}
              loading={generating || report?.status === 'generating'}
              icon={generating ? undefined : <FileText className="h-4 w-4" />}
            >
              {generating ? 'Generating...' : 'Generate Report'}
            </Button>
          )}
        </div>
      </div>

      {report?.error && (
        <p className="mt-3 text-sm text-red-400">{report.error}</p>
      )}

      {hasReport && (
        <div className="mt-4 grid grid-cols-4 gap-4 text-center">
          <div>
            <p className="text-lg font-medium text-white">Overview</p>
            <p className="text-xs text-gray-500">Executive summary</p>
          </div>
          <div>
            <p className="text-lg font-medium text-white">Security</p>
            <p className="text-xs text-gray-500">Vulns & secrets</p>
          </div>
          <div>
            <p className="text-lg font-medium text-white">Dependencies</p>
            <p className="text-xs text-gray-500">SBOM & licenses</p>
          </div>
          <div>
            <p className="text-lg font-medium text-white">DevOps</p>
            <p className="text-xs text-gray-500">DORA & IaC</p>
          </div>
        </div>
      )}
    </Card>
  );
}

function ReportsContent() {
  const [generating, setGenerating] = useState<Record<string, boolean>>({});
  const [reports, setReports] = useState<Record<string, ReportInfo>>({});

  // Fetch projects
  const { data: projectsData, loading } = useFetch(() => api.projects.list(), []);
  const projects = projectsData?.data || [];

  const handleGenerate = async (projectId: string) => {
    setGenerating((prev) => ({ ...prev, [projectId]: true }));

    try {
      // In a real implementation, this would call an API endpoint
      // For now, we'll simulate report generation
      setReports((prev) => ({
        ...prev,
        [projectId]: {
          project_id: projectId,
          generated_at: new Date().toISOString(),
          url: `/api/reports/${encodeURIComponent(projectId)}`,
          status: 'generating',
        },
      }));

      // Simulate generation time
      await new Promise((resolve) => setTimeout(resolve, 2000));

      setReports((prev) => ({
        ...prev,
        [projectId]: {
          project_id: projectId,
          generated_at: new Date().toISOString(),
          url: `/api/reports/${encodeURIComponent(projectId)}`,
          status: 'ready',
        },
      }));
    } catch (err) {
      setReports((prev) => ({
        ...prev,
        [projectId]: {
          project_id: projectId,
          generated_at: new Date().toISOString(),
          url: '',
          status: 'error',
          error: err instanceof Error ? err.message : 'Failed to generate report',
        },
      }));
    } finally {
      setGenerating((prev) => ({ ...prev, [projectId]: false }));
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            <FileText className="h-6 w-6 text-purple-500" />
            Reports
          </h1>
          <p className="mt-1 text-gray-400">
            Generate and view analysis reports for your projects
          </p>
        </div>
      </div>

      {/* Info Card */}
      <Card className="border-blue-700/50 bg-blue-900/10">
        <CardContent>
          <div className="flex items-start gap-3">
            <div className="p-2 rounded-lg bg-blue-600/20">
              <AlertTriangle className="h-5 w-5 text-blue-400" />
            </div>
            <div>
              <h3 className="font-medium text-white">About Reports</h3>
              <p className="mt-1 text-sm text-gray-400">
                Reports are generated using Evidence.dev and provide interactive HTML dashboards
                with charts and detailed analysis. Each report includes sections for security,
                dependencies, supply chain, DevOps, and quality metrics.
              </p>
              <p className="mt-2 text-sm text-gray-500">
                Reports require HTTP to render properly. The generated URL will open in your browser.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Projects List */}
      {loading ? (
        <div className="text-center py-12 text-gray-400">
          <Loader2 className="h-8 w-8 animate-spin mx-auto mb-4" />
          Loading projects...
        </div>
      ) : projects.length === 0 ? (
        <Card className="text-center py-12">
          <FolderOpen className="h-12 w-12 text-gray-600 mx-auto mb-4" />
          <p className="text-gray-400">No projects found</p>
          <p className="text-sm text-gray-500 mt-1">
            Run a scan first to analyze a repository
          </p>
        </Card>
      ) : (
        <div className="space-y-4">
          <h2 className="text-lg font-semibold text-white">Available Projects</h2>
          {projects.map((project) => (
            <ReportCard
              key={project.id}
              project={project}
              report={reports[project.id]}
              onGenerate={() => handleGenerate(project.id)}
              generating={generating[project.id] || false}
            />
          ))}
        </div>
      )}

      {/* Report Types Reference */}
      <Card>
        <CardTitle>Report Sections</CardTitle>
        <CardContent className="mt-4">
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            <div className="p-4 bg-gray-800/50 rounded-lg">
              <h4 className="font-medium text-white mb-2">Overview</h4>
              <p className="text-sm text-gray-400">
                Executive summary with key metrics, severity counts, and engineering insights.
              </p>
            </div>
            <div className="p-4 bg-gray-800/50 rounded-lg">
              <h4 className="font-medium text-white mb-2">Security</h4>
              <p className="text-sm text-gray-400">
                Vulnerabilities, secrets detection, and cryptographic security issues.
              </p>
            </div>
            <div className="p-4 bg-gray-800/50 rounded-lg">
              <h4 className="font-medium text-white mb-2">Dependencies</h4>
              <p className="text-sm text-gray-400">
                SBOM inventory, license distribution, and package health analysis.
              </p>
            </div>
            <div className="p-4 bg-gray-800/50 rounded-lg">
              <h4 className="font-medium text-white mb-2">Supply Chain</h4>
              <p className="text-sm text-gray-400">
                Malcontent detection, package health scores, and supply chain threats.
              </p>
            </div>
            <div className="p-4 bg-gray-800/50 rounded-lg">
              <h4 className="font-medium text-white mb-2">DevOps</h4>
              <p className="text-sm text-gray-400">
                DORA metrics, IaC security, GitHub Actions, and container analysis.
              </p>
            </div>
            <div className="p-4 bg-gray-800/50 rounded-lg">
              <h4 className="font-medium text-white mb-2">Quality & Ownership</h4>
              <p className="text-sm text-gray-400">
                Code quality metrics, developer experience, technologies, and code ownership.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

export default function ReportsPage() {
  return (
    <MainLayout>
      <Suspense fallback={<div className="animate-pulse h-96 bg-gray-800 rounded-lg" />}>
        <ReportsContent />
      </Suspense>
    </MainLayout>
  );
}
