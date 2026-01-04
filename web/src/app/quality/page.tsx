'use client';

import { useState, useMemo, Suspense, useEffect } from 'react';
import { MainLayout } from '@/components/layout/Sidebar';
import { Card, CardTitle, CardContent } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { useFetch } from '@/hooks/useApi';
import { api } from '@/lib/api';
import {
  BarChart3,
  TrendingUp,
  TrendingDown,
  AlertTriangle,
  CheckCircle,
  BookOpen,
  TestTube,
  Wrench,
  Gauge,
  ChevronDown,
  ChevronRight,
} from 'lucide-react';
import { ProjectFilter } from '@/components/ui/ProjectFilter';

interface ProjectQuality {
  projectId: string;
  tech_debt_score: number;
  tech_debt_issues: number;
  complexity_avg: number;
  complexity_max: number;
  test_coverage: number;
  documentation_score: number;
  overall_score: number;
}

function ScoreIndicator({ score }: { score: number }) {
  const color = score >= 80 ? 'text-green-500' :
                score >= 60 ? 'text-yellow-500' :
                score >= 40 ? 'text-orange-500' : 'text-red-500';
  return <span className={`font-bold ${color}`}>{score}</span>;
}

function ProjectQualityCard({ data, expanded, onToggle }: {
  data: ProjectQuality;
  expanded: boolean;
  onToggle: () => void;
}) {
  const health = data.overall_score >= 80 ? 'Healthy' :
                 data.overall_score >= 60 ? 'Needs Attention' : 'At Risk';

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
              <span>Score: <ScoreIndicator score={data.overall_score} /></span>
              <span>Coverage: {data.test_coverage}%</span>
              <span>Complexity: {data.complexity_avg.toFixed(1)}</span>
            </div>
          </div>
        </div>
        <Badge variant={data.overall_score >= 80 ? 'success' : data.overall_score >= 60 ? 'warning' : 'error'}>
          {health}
        </Badge>
      </button>

      {expanded && (
        <div className="border-t border-gray-700 p-4 space-y-4">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            <div className="p-3 bg-gray-800/50 rounded-lg text-center">
              <Wrench className="h-5 w-5 mx-auto mb-1 text-yellow-500" />
              <p className="text-lg font-bold text-white">{data.tech_debt_score}</p>
              <p className="text-xs text-gray-500">Tech Debt Score</p>
            </div>
            <div className="p-3 bg-gray-800/50 rounded-lg text-center">
              <AlertTriangle className="h-5 w-5 mx-auto mb-1 text-orange-500" />
              <p className="text-lg font-bold text-white">{data.complexity_avg.toFixed(1)}</p>
              <p className="text-xs text-gray-500">Avg Complexity</p>
            </div>
            <div className="p-3 bg-gray-800/50 rounded-lg text-center">
              <TestTube className="h-5 w-5 mx-auto mb-1 text-blue-500" />
              <p className="text-lg font-bold text-white">{data.test_coverage}%</p>
              <p className="text-xs text-gray-500">Test Coverage</p>
            </div>
            <div className="p-3 bg-gray-800/50 rounded-lg text-center">
              <BookOpen className="h-5 w-5 mx-auto mb-1 text-purple-500" />
              <p className="text-lg font-bold text-white">{data.documentation_score}</p>
              <p className="text-xs text-gray-500">Documentation</p>
            </div>
          </div>
        </div>
      )}
    </Card>
  );
}

function QualityContent() {
  const [expandedProjects, setExpandedProjects] = useState<Set<string>>(new Set());
  const [qualityData, setQualityData] = useState<ProjectQuality[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedProjects, setSelectedProjects] = useState<string[]>([]);

  const { data: projectsData } = useFetch(() => api.projects.list(), []);
  const projects = projectsData?.data || [];

  useEffect(() => {
    async function loadQualityData() {
      if (projects.length === 0) return;

      setLoading(true);
      const results: ProjectQuality[] = [];

      for (const project of projects) {
        try {
          const data = await api.analysis.raw(project.id, 'code-quality') as any;
          if (data?.findings) {
            const findings = data.findings;
            const techDebt = findings.tech_debt || { score: 0, issues: 0 };
            const complexity = findings.complexity || { average: 0, max: 0 };
            const coverage = findings.test_coverage || { percentage: 0 };
            const docs = findings.documentation || { score: 0 };

            const overallScore = Math.round(
              (techDebt.score +
               Math.max(0, 100 - complexity.average * 5) +
               coverage.percentage +
               docs.score) / 4
            );

            results.push({
              projectId: project.id,
              tech_debt_score: techDebt.score || 0,
              tech_debt_issues: techDebt.issues || 0,
              complexity_avg: complexity.average || 0,
              complexity_max: complexity.max || 0,
              test_coverage: coverage.percentage || 0,
              documentation_score: docs.score || 0,
              overall_score: overallScore,
            });
          }
        } catch {
          // Skip projects without quality data
        }
      }

      setQualityData(results);
      setLoading(false);
    }

    loadQualityData();
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
    if (selectedProjects.length === 0) return qualityData;
    return qualityData.filter(d => selectedProjects.includes(d.projectId));
  }, [qualityData, selectedProjects]);

  // Aggregate stats
  const stats = useMemo(() => {
    if (filteredData.length === 0) return null;

    const healthy = filteredData.filter(d => d.overall_score >= 80).length;
    const needsAttention = filteredData.filter(d => d.overall_score >= 60 && d.overall_score < 80).length;
    const atRisk = filteredData.filter(d => d.overall_score < 60).length;

    const avgScore = Math.round(
      filteredData.reduce((sum, d) => sum + d.overall_score, 0) / filteredData.length
    );
    const avgCoverage = Math.round(
      filteredData.reduce((sum, d) => sum + d.test_coverage, 0) / filteredData.length
    );
    const avgComplexity = (
      filteredData.reduce((sum, d) => sum + d.complexity_avg, 0) / filteredData.length
    ).toFixed(1);

    return {
      healthy,
      needsAttention,
      atRisk,
      avgScore,
      avgCoverage,
      avgComplexity,
      projectsAnalyzed: filteredData.length,
    };
  }, [filteredData]);

  // Sort by score (lowest first = highest priority)
  const sortedData = useMemo(() => {
    return [...filteredData].sort((a, b) => a.overall_score - b.overall_score);
  }, [filteredData]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            <BarChart3 className="h-6 w-6 text-indigo-500" />
            Code Quality
          </h1>
          <p className="mt-1 text-gray-400">
            Technical debt, complexity, test coverage, and documentation across all projects
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
          <div className="grid gap-4 md:grid-cols-3 lg:grid-cols-6">
            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-green-600/20">
                  <CheckCircle className="h-6 w-6 text-green-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Healthy</p>
                  <p className="text-2xl font-bold text-green-500">{stats.healthy}</p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-yellow-600/20">
                  <AlertTriangle className="h-6 w-6 text-yellow-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Needs Work</p>
                  <p className="text-2xl font-bold text-yellow-500">{stats.needsAttention}</p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-red-600/20">
                  <AlertTriangle className="h-6 w-6 text-red-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">At Risk</p>
                  <p className="text-2xl font-bold text-red-500">{stats.atRisk}</p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-indigo-600/20">
                  <Gauge className="h-6 w-6 text-indigo-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Avg Score</p>
                  <p className="text-2xl font-bold text-white">{stats.avgScore}</p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-blue-600/20">
                  <TestTube className="h-6 w-6 text-blue-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Avg Coverage</p>
                  <p className="text-2xl font-bold text-white">{stats.avgCoverage}%</p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-orange-600/20">
                  <TrendingUp className="h-6 w-6 text-orange-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Avg Complexity</p>
                  <p className="text-2xl font-bold text-white">{stats.avgComplexity}</p>
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
                <ProjectQualityCard
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
          <BarChart3 className="h-12 w-12 text-gray-600 mx-auto mb-4" />
          <p className="text-gray-400">No code quality data available</p>
          <p className="text-sm text-gray-500 mt-1">Run scans with the code-quality scanner</p>
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
