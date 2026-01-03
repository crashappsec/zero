'use client';

import { useState, useMemo, Suspense } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
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
  FileText,
  Zap,
  Target,
} from 'lucide-react';

interface DevXMetrics {
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
    recommendation: string;
  };
  workflow: {
    has_ci: boolean;
    has_pre_commit: boolean;
    has_linting: boolean;
    has_formatting: boolean;
    automation_score: number;
  };
}

function OnboardingChecklist({ onboarding }: { onboarding: DevXMetrics['onboarding'] }) {
  const items = [
    { label: 'README.md', present: onboarding.has_readme },
    { label: 'CONTRIBUTING.md', present: onboarding.has_contributing },
    { label: 'Setup Documentation', present: onboarding.has_setup_docs },
    { label: 'Examples/Tutorials', present: onboarding.has_examples },
  ];

  return (
    <div className="space-y-3">
      {items.map((item) => (
        <div key={item.label} className="flex items-center gap-3">
          {item.present ? (
            <CheckCircle className="h-5 w-5 text-green-500" />
          ) : (
            <AlertTriangle className="h-5 w-5 text-yellow-500" />
          )}
          <span className={item.present ? 'text-white' : 'text-gray-400'}>
            {item.label}
          </span>
          {item.present ? (
            <Badge variant="success" className="ml-auto">Present</Badge>
          ) : (
            <Badge variant="warning" className="ml-auto">Missing</Badge>
          )}
        </div>
      ))}
    </div>
  );
}

function WorkflowChecklist({ workflow }: { workflow: DevXMetrics['workflow'] }) {
  const items = [
    { label: 'CI/CD Pipeline', present: workflow.has_ci, icon: Workflow },
    { label: 'Pre-commit Hooks', present: workflow.has_pre_commit, icon: Zap },
    { label: 'Code Linting', present: workflow.has_linting, icon: CheckCircle },
    { label: 'Code Formatting', present: workflow.has_formatting, icon: FileText },
  ];

  return (
    <div className="grid gap-3 md:grid-cols-2">
      {items.map((item) => (
        <div
          key={item.label}
          className={`flex items-center gap-3 p-3 rounded-lg ${
            item.present ? 'bg-green-600/10' : 'bg-gray-800/50'
          }`}
        >
          <item.icon className={`h-5 w-5 ${item.present ? 'text-green-500' : 'text-gray-500'}`} />
          <span className={item.present ? 'text-white' : 'text-gray-400'}>
            {item.label}
          </span>
          {item.present && <CheckCircle className="h-4 w-4 text-green-500 ml-auto" />}
        </div>
      ))}
    </div>
  );
}

function DevXContent() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const projectId = searchParams.get('project');

  const { data: projectsData } = useFetch(() => api.projects.list(), []);
  const projects = projectsData?.data || [];

  const { data: devxData, loading, error } = useFetch(
    () => projectId ? api.analysis.raw(projectId, 'developer-experience') as Promise<any> : Promise.resolve(null),
    [projectId]
  );

  const devx = useMemo(() => {
    if (!devxData?.findings) return null;
    const findings = devxData.findings;
    return {
      onboarding: findings.onboarding || {
        score: 0,
        has_readme: false,
        has_contributing: false,
        has_setup_docs: false,
        has_examples: false,
        estimated_setup_time: 'Unknown',
      },
      sprawl: findings.sprawl || {
        tool_count: 0,
        technology_count: 0,
        redundant_tools: [],
        recommendation: '',
      },
      workflow: findings.workflow || {
        has_ci: false,
        has_pre_commit: false,
        has_linting: false,
        has_formatting: false,
        automation_score: 0,
      },
    } as DevXMetrics;
  }, [devxData]);

  const handleProjectChange = (newProjectId: string) => {
    router.push(`/devx?project=${encodeURIComponent(newProjectId)}`);
  };

  const overallScore = useMemo(() => {
    if (!devx) return 0;
    return Math.round(
      (devx.onboarding.score + devx.workflow.automation_score) / 2
    );
  }, [devx]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-white flex items-center gap-2">
          <Sparkles className="h-6 w-6 text-yellow-500" />
          Developer Experience
        </h1>
        <p className="mt-1 text-gray-400">
          Onboarding experience, tool sprawl, and workflow automation analysis
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

      {projectId && devx && (
        <>
          {/* Stats Overview */}
          <div className="grid gap-4 md:grid-cols-4">
            <Card>
              <div className="flex items-center gap-3">
                <div className={`flex h-12 w-12 items-center justify-center rounded-lg ${
                  overallScore >= 70 ? 'bg-green-600/20' : overallScore >= 50 ? 'bg-yellow-600/20' : 'bg-red-600/20'
                }`}>
                  <Target className={`h-6 w-6 ${
                    overallScore >= 70 ? 'text-green-500' : overallScore >= 50 ? 'text-yellow-500' : 'text-red-500'
                  }`} />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Overall DX Score</p>
                  <p className={`text-2xl font-bold ${
                    overallScore >= 70 ? 'text-green-500' : overallScore >= 50 ? 'text-yellow-500' : 'text-red-500'
                  }`}>{overallScore}</p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-blue-600/20">
                  <BookOpen className="h-6 w-6 text-blue-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Onboarding Score</p>
                  <p className="text-2xl font-bold text-white">{devx.onboarding.score}</p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-purple-600/20">
                  <Layers className="h-6 w-6 text-purple-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Tools</p>
                  <p className="text-2xl font-bold text-white">{devx.sprawl.tool_count}</p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-green-600/20">
                  <Clock className="h-6 w-6 text-green-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Est. Setup Time</p>
                  <p className="text-xl font-bold text-white">{devx.onboarding.estimated_setup_time}</p>
                </div>
              </div>
            </Card>
          </div>

          {/* Onboarding */}
          <Card>
            <CardTitle className="flex items-center gap-2">
              <BookOpen className="h-5 w-5 text-blue-500" />
              Onboarding Experience
            </CardTitle>
            <CardContent className="mt-4">
              <p className="text-sm text-gray-400 mb-4">
                Documentation and resources available for new contributors
              </p>
              <OnboardingChecklist onboarding={devx.onboarding} />
            </CardContent>
          </Card>

          {/* Workflow Automation */}
          <Card>
            <CardTitle className="flex items-center gap-2">
              <Workflow className="h-5 w-5 text-green-500" />
              Workflow Automation
              <Badge variant={devx.workflow.automation_score >= 75 ? 'success' :
                            devx.workflow.automation_score >= 50 ? 'warning' : 'error'}>
                {devx.workflow.automation_score}% automated
              </Badge>
            </CardTitle>
            <CardContent className="mt-4">
              <p className="text-sm text-gray-400 mb-4">
                Development workflow tools and automation
              </p>
              <WorkflowChecklist workflow={devx.workflow} />
            </CardContent>
          </Card>

          {/* Tool Sprawl */}
          <Card>
            <CardTitle className="flex items-center gap-2">
              <Layers className="h-5 w-5 text-purple-500" />
              Tool Sprawl Analysis
            </CardTitle>
            <CardContent className="mt-4">
              <div className="grid gap-4 md:grid-cols-2 mb-4">
                <div className="p-4 bg-gray-800/50 rounded-lg">
                  <p className="text-sm text-gray-400">Total Tools</p>
                  <p className="text-2xl font-bold text-white">{devx.sprawl.tool_count}</p>
                </div>
                <div className="p-4 bg-gray-800/50 rounded-lg">
                  <p className="text-sm text-gray-400">Technologies</p>
                  <p className="text-2xl font-bold text-white">{devx.sprawl.technology_count}</p>
                </div>
              </div>

              {devx.sprawl.redundant_tools.length > 0 && (
                <div className="mt-4">
                  <h4 className="text-sm font-medium text-gray-400 mb-2 flex items-center gap-2">
                    <AlertTriangle className="h-4 w-4 text-yellow-500" />
                    Potentially Redundant Tools
                  </h4>
                  <div className="flex flex-wrap gap-2">
                    {devx.sprawl.redundant_tools.map((tool) => (
                      <Badge key={tool} variant="warning">{tool}</Badge>
                    ))}
                  </div>
                </div>
              )}

              {devx.sprawl.recommendation && (
                <div className="mt-4 p-4 bg-blue-600/10 rounded-lg">
                  <p className="text-sm text-blue-400">
                    <strong>Recommendation:</strong> {devx.sprawl.recommendation}
                  </p>
                </div>
              )}
            </CardContent>
          </Card>
        </>
      )}

      {projectId && loading && (
        <Card className="p-8 text-center text-gray-400">Loading developer experience data...</Card>
      )}

      {projectId && error && (
        <Card className="p-8 text-center text-red-400">
          No developer experience data available. Run a scan with the developer-experience scanner.
        </Card>
      )}

      {!projectId && (
        <Card className="text-center py-12">
          <Sparkles className="h-12 w-12 text-gray-600 mx-auto mb-4" />
          <p className="text-gray-400">Select a project to view developer experience analysis</p>
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
