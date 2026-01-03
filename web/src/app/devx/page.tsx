'use client';

import { useState, useMemo, Suspense, useEffect } from 'react';
import { MainLayout } from '@/components/layout/Sidebar';
import { Card, CardTitle, CardContent } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { useFetch } from '@/hooks/useApi';
import { api } from '@/lib/api';
import {
  Sparkles,
  BookOpen,
  Layers,
  Clock,
  CheckCircle,
  AlertTriangle,
  Workflow,
  Zap,
  ChevronDown,
  ChevronRight,
} from 'lucide-react';
import { ProjectFilter } from '@/components/ui/ProjectFilter';

interface ProjectDevX {
  projectId: string;
  onboarding: {
    score: number;
    has_readme: boolean;
    has_contributing: boolean;
    has_setup_docs: boolean;
    has_examples: boolean;
    estimated_setup_time: string;
  };
  sprawl: {
    tool_count: number;
    technology_count: number;
    redundant_tools: string[];
  };
  workflow: {
    has_ci: boolean;
    has_pre_commit: boolean;
    has_linting: boolean;
    has_formatting: boolean;
    automation_score: number;
  };
  overall_score: number;
}

function ScoreIndicator({ score }: { score: number }) {
  const color = score >= 70 ? 'text-green-500' :
                score >= 50 ? 'text-yellow-500' : 'text-red-500';
  return <span className={`font-bold ${color}`}>{score}</span>;
}

function ProjectDevXCard({ data, expanded, onToggle }: {
  data: ProjectDevX;
  expanded: boolean;
  onToggle: () => void;
}) {
  const dxLevel = data.overall_score >= 70 ? 'Good' :
                  data.overall_score >= 50 ? 'Fair' : 'Needs Work';

  const checklistCount = [
    data.onboarding.has_readme,
    data.onboarding.has_contributing,
    data.onboarding.has_setup_docs,
    data.onboarding.has_examples,
  ].filter(Boolean).length;

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
              <span>DX: <ScoreIndicator score={data.overall_score} /></span>
              <span>{checklistCount}/4 docs</span>
              <span>{data.workflow.automation_score}% automated</span>
            </div>
          </div>
        </div>
        <Badge variant={data.overall_score >= 70 ? 'success' : data.overall_score >= 50 ? 'warning' : 'error'}>
          {dxLevel}
        </Badge>
      </button>

      {expanded && (
        <div className="border-t border-gray-700 p-4 space-y-4">
          {/* Onboarding */}
          <div>
            <h4 className="text-sm font-medium text-gray-400 mb-2">Onboarding ({data.onboarding.score}%)</h4>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-2">
              {[
                { label: 'README', present: data.onboarding.has_readme },
                { label: 'CONTRIBUTING', present: data.onboarding.has_contributing },
                { label: 'Setup Docs', present: data.onboarding.has_setup_docs },
                { label: 'Examples', present: data.onboarding.has_examples },
              ].map((item) => (
                <div key={item.label} className="flex items-center gap-2 text-sm">
                  {item.present ? (
                    <CheckCircle className="h-4 w-4 text-green-500" />
                  ) : (
                    <AlertTriangle className="h-4 w-4 text-yellow-500" />
                  )}
                  <span className={item.present ? 'text-white' : 'text-gray-500'}>
                    {item.label}
                  </span>
                </div>
              ))}
            </div>
          </div>

          {/* Workflow */}
          <div>
            <h4 className="text-sm font-medium text-gray-400 mb-2">Workflow ({data.workflow.automation_score}% automated)</h4>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-2">
              {[
                { label: 'CI/CD', present: data.workflow.has_ci },
                { label: 'Pre-commit', present: data.workflow.has_pre_commit },
                { label: 'Linting', present: data.workflow.has_linting },
                { label: 'Formatting', present: data.workflow.has_formatting },
              ].map((item) => (
                <div key={item.label} className="flex items-center gap-2 text-sm">
                  {item.present ? (
                    <CheckCircle className="h-4 w-4 text-green-500" />
                  ) : (
                    <AlertTriangle className="h-4 w-4 text-gray-500" />
                  )}
                  <span className={item.present ? 'text-white' : 'text-gray-500'}>
                    {item.label}
                  </span>
                </div>
              ))}
            </div>
          </div>

          {/* Tool Sprawl */}
          {data.sprawl.redundant_tools.length > 0 && (
            <div>
              <h4 className="text-sm font-medium text-gray-400 mb-2">Redundant Tools</h4>
              <div className="flex flex-wrap gap-2">
                {data.sprawl.redundant_tools.map((tool) => (
                  <Badge key={tool} variant="warning">{tool}</Badge>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </Card>
  );
}

function DevXContent() {
  const [expandedProjects, setExpandedProjects] = useState<Set<string>>(new Set());
  const [devxData, setDevxData] = useState<ProjectDevX[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedProjects, setSelectedProjects] = useState<string[]>([]);

  const { data: projectsData } = useFetch(() => api.projects.list(), []);
  const projects = projectsData?.data || [];

  useEffect(() => {
    async function loadDevXData() {
      if (projects.length === 0) return;

      setLoading(true);
      const results: ProjectDevX[] = [];

      for (const project of projects) {
        try {
          const data = await api.analysis.raw(project.id, 'developer-experience') as any;
          if (data?.findings) {
            const findings = data.findings;
            const onboarding = findings.onboarding || {
              score: 0,
              has_readme: false,
              has_contributing: false,
              has_setup_docs: false,
              has_examples: false,
              estimated_setup_time: 'Unknown',
            };
            const sprawl = findings.sprawl || {
              tool_count: 0,
              technology_count: 0,
              redundant_tools: [],
            };
            const workflow = findings.workflow || {
              has_ci: false,
              has_pre_commit: false,
              has_linting: false,
              has_formatting: false,
              automation_score: 0,
            };

            const overallScore = Math.round(
              (onboarding.score + workflow.automation_score) / 2
            );

            results.push({
              projectId: project.id,
              onboarding,
              sprawl,
              workflow,
              overall_score: overallScore,
            });
          }
        } catch {
          // Skip projects without devx data
        }
      }

      setDevxData(results);
      setLoading(false);
    }

    loadDevXData();
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
    if (selectedProjects.length === 0) return devxData;
    return devxData.filter(d => selectedProjects.includes(d.projectId));
  }, [devxData, selectedProjects]);

  // Aggregate stats
  const stats = useMemo(() => {
    if (filteredData.length === 0) return null;

    const good = filteredData.filter(d => d.overall_score >= 70).length;
    const fair = filteredData.filter(d => d.overall_score >= 50 && d.overall_score < 70).length;
    const needsWork = filteredData.filter(d => d.overall_score < 50).length;

    const avgScore = Math.round(
      filteredData.reduce((sum, d) => sum + d.overall_score, 0) / filteredData.length
    );
    const avgAutomation = Math.round(
      filteredData.reduce((sum, d) => sum + d.workflow.automation_score, 0) / filteredData.length
    );

    const withReadme = filteredData.filter(d => d.onboarding.has_readme).length;
    const withCI = filteredData.filter(d => d.workflow.has_ci).length;

    return {
      good,
      fair,
      needsWork,
      avgScore,
      avgAutomation,
      withReadme,
      withCI,
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
            <Sparkles className="h-6 w-6 text-yellow-500" />
            Developer Experience
          </h1>
          <p className="mt-1 text-gray-400">
            Onboarding, tool sprawl, and workflow automation across all projects
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
                  <p className="text-sm text-gray-400">Good DX</p>
                  <p className="text-2xl font-bold text-green-500">{stats.good}</p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-yellow-600/20">
                  <AlertTriangle className="h-6 w-6 text-yellow-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Fair DX</p>
                  <p className="text-2xl font-bold text-yellow-500">{stats.fair}</p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-red-600/20">
                  <AlertTriangle className="h-6 w-6 text-red-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Needs Work</p>
                  <p className="text-2xl font-bold text-red-500">{stats.needsWork}</p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-purple-600/20">
                  <Sparkles className="h-6 w-6 text-purple-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Avg DX Score</p>
                  <p className="text-2xl font-bold text-white">{stats.avgScore}</p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-blue-600/20">
                  <BookOpen className="h-6 w-6 text-blue-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Have README</p>
                  <p className="text-2xl font-bold text-white">{stats.withReadme}</p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-cyan-600/20">
                  <Workflow className="h-6 w-6 text-cyan-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Have CI/CD</p>
                  <p className="text-2xl font-bold text-white">{stats.withCI}</p>
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
                <ProjectDevXCard
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
          <Sparkles className="h-12 w-12 text-gray-600 mx-auto mb-4" />
          <p className="text-gray-400">No developer experience data available</p>
          <p className="text-sm text-gray-500 mt-1">Run scans with the developer-experience scanner</p>
        </Card>
      )}
    </div>
  );
}

export default function DevXPage() {
  return (
    <MainLayout>
      <Suspense fallback={<div className="animate-pulse h-96 bg-gray-800 rounded-lg" />}>
        <DevXContent />
      </Suspense>
    </MainLayout>
  );
}
