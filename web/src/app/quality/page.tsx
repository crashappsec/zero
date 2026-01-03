'use client';

import { useState, useMemo, Suspense } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { MainLayout } from '@/components/layout/Sidebar';
import { Card, CardTitle, CardContent } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { useFetch } from '@/hooks/useApi';
import { api } from '@/lib/api';
import {
  BarChart3,
  TrendingUp,
  TrendingDown,
  FileCode,
  AlertTriangle,
  CheckCircle,
  BookOpen,
  TestTube,
  Wrench,
  Gauge,
} from 'lucide-react';

interface QualityMetrics {
  tech_debt: {
    score: number;
    issues: number;
    hotspots: Array<{ file: string; debt_minutes: number }>;
  };
  complexity: {
    average: number;
    max: number;
    high_complexity_files: Array<{ file: string; complexity: number }>;
  };
  test_coverage: {
    percentage: number;
    covered_lines: number;
    total_lines: number;
  };
  documentation: {
    score: number;
    documented_functions: number;
    total_functions: number;
  };
}

function ScoreGauge({ score, label, description }: { score: number; label: string; description?: string }) {
  const getColor = (s: number) => {
    if (s >= 80) return 'text-green-500';
    if (s >= 60) return 'text-yellow-500';
    if (s >= 40) return 'text-orange-500';
    return 'text-red-500';
  };

  const getBgColor = (s: number) => {
    if (s >= 80) return 'bg-green-600/20';
    if (s >= 60) return 'bg-yellow-600/20';
    if (s >= 40) return 'bg-orange-600/20';
    return 'bg-red-600/20';
  };

  return (
    <Card>
      <div className="flex items-center gap-4">
        <div className={`flex h-16 w-16 items-center justify-center rounded-full ${getBgColor(score)}`}>
          <span className={`text-2xl font-bold ${getColor(score)}`}>{score}</span>
        </div>
        <div>
          <p className="text-sm text-gray-400">{label}</p>
          <div className="flex items-center gap-2">
            <p className={`text-xl font-bold ${getColor(score)}`}>
              {score >= 80 ? 'Good' : score >= 60 ? 'Fair' : score >= 40 ? 'Needs Work' : 'Poor'}
            </p>
            {score >= 60 ? (
              <TrendingUp className="h-4 w-4 text-green-500" />
            ) : (
              <TrendingDown className="h-4 w-4 text-red-500" />
            )}
          </div>
          {description && <p className="text-xs text-gray-500">{description}</p>}
        </div>
      </div>
    </Card>
  );
}

function ComplexityChart({ files }: { files: Array<{ file: string; complexity: number }> }) {
  const maxComplexity = Math.max(...files.map(f => f.complexity), 10);

  return (
    <div className="space-y-3">
      {files.map((file) => (
        <div key={file.file}>
          <div className="flex items-center justify-between text-sm mb-1">
            <span className="text-gray-400 truncate max-w-[70%]">{file.file}</span>
            <span className={`font-medium ${
              file.complexity > 20 ? 'text-red-500' :
              file.complexity > 10 ? 'text-yellow-500' : 'text-green-500'
            }`}>
              {file.complexity}
            </span>
          </div>
          <div className="h-2 bg-gray-700 rounded-full overflow-hidden">
            <div
              className={`h-full transition-all duration-500 ${
                file.complexity > 20 ? 'bg-red-500' :
                file.complexity > 10 ? 'bg-yellow-500' : 'bg-green-500'
              }`}
              style={{ width: `${(file.complexity / maxComplexity) * 100}%` }}
            />
          </div>
        </div>
      ))}
    </div>
  );
}

function QualityContent() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const projectId = searchParams.get('project');

  const { data: projectsData } = useFetch(() => api.projects.list(), []);
  const projects = projectsData?.data || [];

  const { data: qualityData, loading, error } = useFetch(
    () => projectId ? api.analysis.raw(projectId, 'code-quality') as Promise<any> : Promise.resolve(null),
    [projectId]
  );

  const quality = useMemo(() => {
    if (!qualityData?.findings) return null;
    const findings = qualityData.findings;
    return {
      tech_debt: findings.tech_debt || { score: 0, issues: 0, hotspots: [] },
      complexity: findings.complexity || { average: 0, max: 0, high_complexity_files: [] },
      test_coverage: findings.test_coverage || { percentage: 0, covered_lines: 0, total_lines: 0 },
      documentation: findings.documentation || { score: 0, documented_functions: 0, total_functions: 0 },
    } as QualityMetrics;
  }, [qualityData]);

  const handleProjectChange = (newProjectId: string) => {
    router.push(`/quality?project=${encodeURIComponent(newProjectId)}`);
  };

  const overallScore = useMemo(() => {
    if (!quality) return 0;
    return Math.round(
      (quality.tech_debt.score +
       Math.max(0, 100 - quality.complexity.average * 5) +
       quality.test_coverage.percentage +
       quality.documentation.score) / 4
    );
  }, [quality]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-white flex items-center gap-2">
          <BarChart3 className="h-6 w-6 text-indigo-500" />
          Code Quality
        </h1>
        <p className="mt-1 text-gray-400">
          Technical debt, complexity, test coverage, and documentation metrics
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

      {projectId && quality && (
        <>
          {/* Overall Score */}
          <Card>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <div className={`flex h-20 w-20 items-center justify-center rounded-full ${
                  overallScore >= 80 ? 'bg-green-600/20' :
                  overallScore >= 60 ? 'bg-yellow-600/20' : 'bg-red-600/20'
                }`}>
                  <Gauge className={`h-10 w-10 ${
                    overallScore >= 80 ? 'text-green-500' :
                    overallScore >= 60 ? 'text-yellow-500' : 'text-red-500'
                  }`} />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Overall Code Quality Score</p>
                  <p className={`text-4xl font-bold ${
                    overallScore >= 80 ? 'text-green-500' :
                    overallScore >= 60 ? 'text-yellow-500' : 'text-red-500'
                  }`}>
                    {overallScore}/100
                  </p>
                </div>
              </div>
              <div className="text-right">
                <Badge variant={overallScore >= 80 ? 'success' : overallScore >= 60 ? 'warning' : 'error'}>
                  {overallScore >= 80 ? 'Healthy' : overallScore >= 60 ? 'Needs Attention' : 'At Risk'}
                </Badge>
              </div>
            </div>
          </Card>

          {/* Metric Cards */}
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <ScoreGauge
              score={quality.tech_debt.score}
              label="Technical Debt"
              description={`${quality.tech_debt.issues} issues`}
            />
            <ScoreGauge
              score={Math.max(0, 100 - quality.complexity.average * 5)}
              label="Complexity"
              description={`Avg: ${quality.complexity.average.toFixed(1)}`}
            />
            <ScoreGauge
              score={quality.test_coverage.percentage}
              label="Test Coverage"
              description={`${quality.test_coverage.covered_lines}/${quality.test_coverage.total_lines} lines`}
            />
            <ScoreGauge
              score={quality.documentation.score}
              label="Documentation"
              description={`${quality.documentation.documented_functions}/${quality.documentation.total_functions} functions`}
            />
          </div>

          {/* Complexity Hotspots */}
          {quality.complexity.high_complexity_files.length > 0 && (
            <Card>
              <CardTitle className="flex items-center gap-2">
                <AlertTriangle className="h-5 w-5 text-orange-500" />
                High Complexity Files
              </CardTitle>
              <CardContent className="mt-4">
                <p className="text-sm text-gray-400 mb-4">
                  Files with cyclomatic complexity above recommended thresholds
                </p>
                <ComplexityChart files={quality.complexity.high_complexity_files.slice(0, 10)} />
              </CardContent>
            </Card>
          )}

          {/* Tech Debt Hotspots */}
          {quality.tech_debt.hotspots.length > 0 && (
            <Card>
              <CardTitle className="flex items-center gap-2">
                <Wrench className="h-5 w-5 text-yellow-500" />
                Technical Debt Hotspots
              </CardTitle>
              <CardContent className="mt-4">
                <p className="text-sm text-gray-400 mb-4">
                  Files with the most accumulated technical debt
                </p>
                <div className="space-y-2">
                  {quality.tech_debt.hotspots.slice(0, 10).map((hotspot) => (
                    <div key={hotspot.file} className="flex items-center justify-between p-2 bg-gray-800/50 rounded">
                      <div className="flex items-center gap-2 min-w-0">
                        <FileCode className="h-4 w-4 text-gray-500 shrink-0" />
                        <span className="text-sm text-white truncate">{hotspot.file}</span>
                      </div>
                      <Badge variant="warning">
                        {hotspot.debt_minutes < 60
                          ? `${hotspot.debt_minutes}m`
                          : `${Math.round(hotspot.debt_minutes / 60)}h`} debt
                      </Badge>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}
        </>
      )}

      {projectId && loading && (
        <Card className="p-8 text-center text-gray-400">Loading quality data...</Card>
      )}

      {projectId && error && (
        <Card className="p-8 text-center text-red-400">
          No code quality data available. Run a scan with the code-quality scanner.
        </Card>
      )}

      {!projectId && (
        <Card className="text-center py-12">
          <BarChart3 className="h-12 w-12 text-gray-600 mx-auto mb-4" />
          <p className="text-gray-400">Select a project to view code quality metrics</p>
        </Card>
      )}
    </div>
  );
}

export default function QualityPage() {
  return (
    <MainLayout>
      <Suspense fallback={<div className="animate-pulse h-96 bg-gray-800 rounded-lg" />}>
        <QualityContent />
      </Suspense>
    </MainLayout>
  );
}
