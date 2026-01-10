'use client';

import { useState, useMemo, Suspense, useEffect } from 'react';
import { MainLayout } from '@/components/layout/Sidebar';
import { Card, CardTitle, CardContent } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { BenchmarkTier, LINEARB_BENCHMARKS, getTier, TierLevel } from '@/components/ui/BenchmarkTier';
import { useFetch } from '@/hooks/useApi';
import { api } from '@/lib/api';
import { ProjectFilter } from '@/components/ui/ProjectFilter';
import {
  Zap,
  TrendingUp,
  Clock,
  GitPullRequest,
  Rocket,
  AlertTriangle,
  ChevronRight,
  Server,
  Container,
  Workflow,
} from 'lucide-react';
import Link from 'next/link';

interface ProjectSpeed {
  projectId: string;
  dora: {
    deployment_frequency: string;
    lead_time: string;
    change_failure_rate: number;
    mttr: string;
    performance_level: string;
    // PR-level metrics (Phase 3)
    avg_pickup_hours?: number;
    pickup_class?: string;
    avg_review_hours?: number;
    review_class?: string;
    avg_merge_hours?: number;
    merge_class?: string;
    avg_pr_size?: number;
    pr_size_class?: string;
    total_prs?: number;
    // Rework rate (DORA 2025)
    rework_rate?: number;
    rework_class?: string;
    refactor_rate?: number;
    refactor_class?: string;
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

interface AggregateStats {
  doraLevels: Record<string, number>;
  avgChangeFailureRate: number;
  totalProjects: number;
  totalIssues: number;
  elitePerformers: number;
  // PR-level metrics (Phase 3)
  avgPickupHours: number;
  avgReviewHours: number;
  avgMergeHours: number;
  avgPRSize: number;
  totalPRs: number;
  // Rework rate (DORA 2025)
  avgReworkRate: number;
  avgRefactorRate: number;
}

function DORALevelBadge({ level }: { level: string }) {
  const tierMap: Record<string, TierLevel> = {
    elite: 'elite',
    high: 'good',
    medium: 'fair',
    low: 'needs_focus',
  };
  const tier = tierMap[level.toLowerCase()] || 'unknown';
  const colorMap: Record<TierLevel, string> = {
    elite: 'bg-green-500/20 text-green-500',
    good: 'bg-blue-500/20 text-blue-500',
    fair: 'bg-yellow-500/20 text-yellow-500',
    needs_focus: 'bg-red-500/20 text-red-500',
    unknown: 'bg-gray-500/20 text-gray-500',
  };

  return (
    <span className={`px-2 py-1 rounded text-xs font-medium ${colorMap[tier]}`}>
      {level.charAt(0).toUpperCase() + level.slice(1)} Performer
    </span>
  );
}

function SpeedContent() {
  const [speedData, setSpeedData] = useState<ProjectSpeed[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedProjects, setSelectedProjects] = useState<string[]>([]);

  const { data: projectsData } = useFetch(() => api.projects.list(), []);
  const projects = projectsData?.data || [];

  useEffect(() => {
    async function loadSpeedData() {
      if (projects.length === 0) return;

      setLoading(true);
      const results: ProjectSpeed[] = [];

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

      setSpeedData(results);
      setLoading(false);
    }

    loadSpeedData();
  }, [projects]);

  // Filter data based on selected projects
  const filteredData = useMemo(() => {
    if (selectedProjects.length === 0) return speedData;
    return speedData.filter(d => selectedProjects.includes(d.projectId));
  }, [speedData, selectedProjects]);

  // Aggregate stats
  const stats = useMemo((): AggregateStats | null => {
    if (filteredData.length === 0) return null;

    const doraLevels: Record<string, number> = { elite: 0, high: 0, medium: 0, low: 0 };
    let totalCFR = 0;
    let cfrCount = 0;
    let totalIssues = 0;
    // PR metrics
    let totalPickup = 0, totalReview = 0, totalMerge = 0, totalSize = 0, totalPRs = 0;
    let prMetricsCount = 0;
    // Rework rate
    let totalRework = 0, totalRefactor = 0;
    let reworkCount = 0;

    filteredData.forEach((p) => {
      if (p.dora) {
        const level = p.dora.performance_level.toLowerCase();
        if (level in doraLevels) doraLevels[level]++;
        if (!isNaN(p.dora.change_failure_rate)) {
          totalCFR += p.dora.change_failure_rate;
          cfrCount++;
        }
        // Aggregate PR metrics
        if (p.dora.total_prs && p.dora.total_prs > 0) {
          totalPickup += p.dora.avg_pickup_hours || 0;
          totalReview += p.dora.avg_review_hours || 0;
          totalMerge += p.dora.avg_merge_hours || 0;
          totalSize += p.dora.avg_pr_size || 0;
          totalPRs += p.dora.total_prs;
          prMetricsCount++;
        }
        // Aggregate rework rate
        if (p.dora.rework_rate !== undefined) {
          totalRework += p.dora.rework_rate;
          totalRefactor += p.dora.refactor_rate || 0;
          reworkCount++;
        }
      }
      totalIssues += p.iac_count + p.container_count + p.gha_count;
    });

    return {
      doraLevels,
      avgChangeFailureRate: cfrCount > 0 ? totalCFR / cfrCount : 0,
      totalProjects: filteredData.length,
      totalIssues,
      elitePerformers: doraLevels.elite,
      // PR metrics averages
      avgPickupHours: prMetricsCount > 0 ? totalPickup / prMetricsCount : 0,
      avgReviewHours: prMetricsCount > 0 ? totalReview / prMetricsCount : 0,
      avgMergeHours: prMetricsCount > 0 ? totalMerge / prMetricsCount : 0,
      avgPRSize: prMetricsCount > 0 ? Math.round(totalSize / prMetricsCount) : 0,
      totalPRs,
      // Rework rate averages
      avgReworkRate: reworkCount > 0 ? totalRework / reworkCount : 0,
      avgRefactorRate: reworkCount > 0 ? totalRefactor / reworkCount : 0,
    };
  }, [filteredData]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            <Zap className="h-6 w-6 text-yellow-500" />
            Speed
          </h1>
          <p className="mt-1 text-gray-400">
            DORA metrics, cycle time, and delivery performance
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
          {/* Key Metrics with Benchmarks */}
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <BenchmarkTier
              value={stats.avgChangeFailureRate}
              label="Change Failure Rate"
              tiers={LINEARB_BENCHMARKS.changeFailureRate}
              unit="%"
              lowerIsBetter={true}
            />
            <Card className="p-4">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm text-gray-400">Elite Performers</span>
                <span className="text-xs font-medium px-2 py-0.5 rounded bg-green-500/20 text-green-500">
                  DORA
                </span>
              </div>
              <p className="text-2xl font-bold text-green-500">
                {stats.elitePerformers}
                <span className="text-sm font-normal text-gray-400 ml-1">
                  / {stats.totalProjects}
                </span>
              </p>
              <p className="text-xs text-gray-500 mt-2">
                {Math.round((stats.elitePerformers / stats.totalProjects) * 100)}% of projects
              </p>
            </Card>
            <Card className="p-4">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm text-gray-400">Projects Analyzed</span>
              </div>
              <p className="text-2xl font-bold text-blue-500">{stats.totalProjects}</p>
              <p className="text-xs text-gray-500 mt-2">
                With DORA metrics
              </p>
            </Card>
            <Card className="p-4">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm text-gray-400">DevOps Issues</span>
              </div>
              <p className="text-2xl font-bold text-orange-500">{stats.totalIssues}</p>
              <p className="text-xs text-gray-500 mt-2">
                IaC, Container, CI/CD
              </p>
            </Card>
          </div>

          {/* DORA Performance Distribution */}
          <Card>
            <CardContent>
              <div className="flex items-center gap-2 mb-4">
                <TrendingUp className="h-5 w-5 text-blue-500" />
                <CardTitle>DORA Performance Distribution</CardTitle>
              </div>
              <div className="grid grid-cols-4 gap-4 text-center">
                <div className="p-4 bg-green-500/10 rounded-lg border border-green-500/20">
                  <p className="text-2xl font-bold text-green-500">{stats.doraLevels.elite}</p>
                  <p className="text-sm text-gray-400">Elite</p>
                  <p className="text-xs text-gray-500 mt-1">On-demand deploys</p>
                </div>
                <div className="p-4 bg-blue-500/10 rounded-lg border border-blue-500/20">
                  <p className="text-2xl font-bold text-blue-500">{stats.doraLevels.high}</p>
                  <p className="text-sm text-gray-400">High</p>
                  <p className="text-xs text-gray-500 mt-1">Daily to weekly</p>
                </div>
                <div className="p-4 bg-yellow-500/10 rounded-lg border border-yellow-500/20">
                  <p className="text-2xl font-bold text-yellow-500">{stats.doraLevels.medium}</p>
                  <p className="text-sm text-gray-400">Medium</p>
                  <p className="text-xs text-gray-500 mt-1">Weekly to monthly</p>
                </div>
                <div className="p-4 bg-red-500/10 rounded-lg border border-red-500/20">
                  <p className="text-2xl font-bold text-red-500">{stats.doraLevels.low}</p>
                  <p className="text-sm text-gray-400">Low</p>
                  <p className="text-xs text-gray-500 mt-1">Monthly or less</p>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* PR Cycle Time Metrics (Phase 3 - LinearB alignment) */}
          {stats.totalPRs > 0 && (
            <Card>
              <CardContent>
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center gap-2">
                    <GitPullRequest className="h-5 w-5 text-purple-500" />
                    <CardTitle>PR Cycle Time Breakdown</CardTitle>
                  </div>
                  <Badge variant="default">{stats.totalPRs} PRs analyzed</Badge>
                </div>
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                  <BenchmarkTier
                    value={stats.avgPickupHours.toFixed(1)}
                    label="Pickup Time"
                    tiers={LINEARB_BENCHMARKS.pickupTime}
                    unit="h"
                    lowerIsBetter={true}
                  />
                  <BenchmarkTier
                    value={stats.avgReviewHours.toFixed(1)}
                    label="Review Time"
                    tiers={LINEARB_BENCHMARKS.reviewTime}
                    unit="h"
                    lowerIsBetter={true}
                  />
                  <BenchmarkTier
                    value={stats.avgMergeHours.toFixed(1)}
                    label="Merge Time"
                    tiers={LINEARB_BENCHMARKS.mergeTime}
                    unit="h"
                    lowerIsBetter={true}
                  />
                  <BenchmarkTier
                    value={stats.avgPRSize}
                    label="Avg PR Size"
                    tiers={LINEARB_BENCHMARKS.prSize}
                    unit=" lines"
                    lowerIsBetter={true}
                  />
                </div>
              </CardContent>
            </Card>
          )}

          {/* Rework & Refactor Rate (DORA 2025) */}
          {(stats.avgReworkRate > 0 || stats.avgRefactorRate > 0) && (
            <Card>
              <CardContent>
                <div className="flex items-center gap-2 mb-4">
                  <Clock className="h-5 w-5 text-orange-500" />
                  <CardTitle>Code Churn Analysis (DORA 2025)</CardTitle>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <BenchmarkTier
                    value={stats.avgReworkRate.toFixed(1)}
                    label="Rework Rate"
                    tiers={LINEARB_BENCHMARKS.reworkRate}
                    unit="%"
                    lowerIsBetter={true}
                  />
                  <BenchmarkTier
                    value={stats.avgRefactorRate.toFixed(1)}
                    label="Refactor Rate"
                    tiers={LINEARB_BENCHMARKS.refactorRate}
                    unit="%"
                    lowerIsBetter={true}
                  />
                </div>
                <p className="text-xs text-gray-500 mt-3">
                  Rework: commits fixing recent changes | Refactor: commits improving existing code
                </p>
              </CardContent>
            </Card>
          )}

          {/* Project Details */}
          <Card>
            <CardContent>
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-2">
                  <Rocket className="h-5 w-5 text-purple-500" />
                  <CardTitle>Project Delivery Metrics</CardTitle>
                </div>
              </div>
              <div className="space-y-4">
                {filteredData.map((data) => (
                  <div
                    key={data.projectId}
                    className="p-4 bg-gray-800/50 rounded-lg border border-gray-700"
                  >
                    <div className="flex items-center justify-between mb-3">
                      <div>
                        <h3 className="font-medium text-white">{data.projectId}</h3>
                        <p className="text-sm text-gray-500">
                          {data.git_stats.total_commits} commits | {data.git_stats.branches} branches
                        </p>
                      </div>
                      {data.dora && <DORALevelBadge level={data.dora.performance_level} />}
                    </div>

                    {data.dora ? (
                      <>
                        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
                          <div className="p-2 bg-gray-900/50 rounded">
                            <p className="text-xs text-gray-500">Deploy Frequency</p>
                            <p className="text-sm font-medium text-white">{data.dora.deployment_frequency}</p>
                          </div>
                          <div className="p-2 bg-gray-900/50 rounded">
                            <p className="text-xs text-gray-500">Lead Time</p>
                            <p className="text-sm font-medium text-white">{data.dora.lead_time}</p>
                          </div>
                          <div className="p-2 bg-gray-900/50 rounded">
                            <p className="text-xs text-gray-500">Change Failure Rate</p>
                            <p className={`text-sm font-medium ${data.dora.change_failure_rate <= 4 ? 'text-green-500' : data.dora.change_failure_rate <= 17 ? 'text-yellow-500' : 'text-red-500'}`}>
                              {data.dora.change_failure_rate}%
                            </p>
                          </div>
                          <div className="p-2 bg-gray-900/50 rounded">
                            <p className="text-xs text-gray-500">MTTR</p>
                            <p className="text-sm font-medium text-white">{data.dora.mttr}</p>
                          </div>
                        </div>

                        {/* PR Cycle Time (if available) */}
                        {data.dora.total_prs && data.dora.total_prs > 0 && (
                          <div className="grid grid-cols-2 md:grid-cols-5 gap-3 mt-3">
                            <div className="p-2 bg-purple-900/20 rounded border border-purple-500/20">
                              <p className="text-xs text-gray-500">Pickup</p>
                              <p className={`text-sm font-medium ${data.dora.pickup_class === 'elite' ? 'text-green-500' : data.dora.pickup_class === 'good' ? 'text-blue-500' : 'text-yellow-500'}`}>
                                {data.dora.avg_pickup_hours?.toFixed(1)}h
                              </p>
                            </div>
                            <div className="p-2 bg-purple-900/20 rounded border border-purple-500/20">
                              <p className="text-xs text-gray-500">Review</p>
                              <p className={`text-sm font-medium ${data.dora.review_class === 'elite' ? 'text-green-500' : data.dora.review_class === 'good' ? 'text-blue-500' : 'text-yellow-500'}`}>
                                {data.dora.avg_review_hours?.toFixed(1)}h
                              </p>
                            </div>
                            <div className="p-2 bg-purple-900/20 rounded border border-purple-500/20">
                              <p className="text-xs text-gray-500">Merge</p>
                              <p className={`text-sm font-medium ${data.dora.merge_class === 'elite' ? 'text-green-500' : data.dora.merge_class === 'good' ? 'text-blue-500' : 'text-yellow-500'}`}>
                                {data.dora.avg_merge_hours?.toFixed(1)}h
                              </p>
                            </div>
                            <div className="p-2 bg-purple-900/20 rounded border border-purple-500/20">
                              <p className="text-xs text-gray-500">PR Size</p>
                              <p className={`text-sm font-medium ${data.dora.pr_size_class === 'elite' ? 'text-green-500' : data.dora.pr_size_class === 'good' ? 'text-blue-500' : 'text-yellow-500'}`}>
                                {data.dora.avg_pr_size} lines
                              </p>
                            </div>
                            <div className="p-2 bg-purple-900/20 rounded border border-purple-500/20">
                              <p className="text-xs text-gray-500">PRs</p>
                              <p className="text-sm font-medium text-purple-400">{data.dora.total_prs}</p>
                            </div>
                          </div>
                        )}

                        {/* Rework Rate (if available) */}
                        {data.dora.rework_rate !== undefined && (
                          <div className="grid grid-cols-2 gap-3 mt-3">
                            <div className="p-2 bg-orange-900/20 rounded border border-orange-500/20">
                              <p className="text-xs text-gray-500">Rework Rate</p>
                              <p className={`text-sm font-medium ${data.dora.rework_class === 'elite' ? 'text-green-500' : data.dora.rework_class === 'good' ? 'text-blue-500' : 'text-yellow-500'}`}>
                                {data.dora.rework_rate.toFixed(1)}%
                              </p>
                            </div>
                            <div className="p-2 bg-orange-900/20 rounded border border-orange-500/20">
                              <p className="text-xs text-gray-500">Refactor Rate</p>
                              <p className={`text-sm font-medium ${data.dora.refactor_class === 'elite' ? 'text-green-500' : data.dora.refactor_class === 'good' ? 'text-blue-500' : 'text-yellow-500'}`}>
                                {data.dora.refactor_rate?.toFixed(1)}%
                              </p>
                            </div>
                          </div>
                        )}
                      </>
                    ) : (
                      <p className="text-sm text-gray-500">No DORA metrics available</p>
                    )}

                    {(data.iac_count > 0 || data.container_count > 0 || data.gha_count > 0) && (
                      <div className="flex flex-wrap gap-2 mt-3">
                        {data.iac_count > 0 && (
                          <Badge variant="warning">{data.iac_count} IaC issues</Badge>
                        )}
                        {data.container_count > 0 && (
                          <Badge variant="warning">{data.container_count} container issues</Badge>
                        )}
                        {data.gha_count > 0 && (
                          <Badge variant="warning">{data.gha_count} CI/CD issues</Badge>
                        )}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>

          {/* Quick Links */}
          <div className="grid grid-cols-3 gap-4">
            <Link href="/devops">
              <Card className="hover:bg-gray-800/50 transition-colors cursor-pointer">
                <CardContent className="flex items-center gap-4">
                  <Server className="h-8 w-8 text-orange-500" />
                  <div>
                    <p className="font-medium text-white">IaC Security</p>
                    <p className="text-sm text-gray-400">Terraform, CloudFormation</p>
                  </div>
                  <ChevronRight className="h-5 w-5 text-gray-500 ml-auto" />
                </CardContent>
              </Card>
            </Link>
            <Link href="/devops">
              <Card className="hover:bg-gray-800/50 transition-colors cursor-pointer">
                <CardContent className="flex items-center gap-4">
                  <Container className="h-8 w-8 text-blue-500" />
                  <div>
                    <p className="font-medium text-white">Container Security</p>
                    <p className="text-sm text-gray-400">Docker, Kubernetes</p>
                  </div>
                  <ChevronRight className="h-5 w-5 text-gray-500 ml-auto" />
                </CardContent>
              </Card>
            </Link>
            <Link href="/devops">
              <Card className="hover:bg-gray-800/50 transition-colors cursor-pointer">
                <CardContent className="flex items-center gap-4">
                  <Workflow className="h-8 w-8 text-purple-500" />
                  <div>
                    <p className="font-medium text-white">CI/CD Security</p>
                    <p className="text-sm text-gray-400">GitHub Actions</p>
                  </div>
                  <ChevronRight className="h-5 w-5 text-gray-500 ml-auto" />
                </CardContent>
              </Card>
            </Link>
          </div>

          {/* Benchmark Reference */}
          <Card className="bg-gray-800/30 border-gray-700">
            <CardContent>
              <div className="flex items-center gap-2 mb-3">
                <AlertTriangle className="h-4 w-4 text-yellow-500" />
                <span className="text-sm font-medium text-gray-300">LinearB 2026 Benchmarks</span>
              </div>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-xs text-gray-400">
                <div>
                  <p className="font-medium text-gray-300">Cycle Time</p>
                  <p>Elite: &lt;25h | Good: 25-72h</p>
                </div>
                <div>
                  <p className="font-medium text-gray-300">Deploy Frequency</p>
                  <p>Elite: &gt;1.2/day | Good: 0.5-1.2</p>
                </div>
                <div>
                  <p className="font-medium text-gray-300">Change Failure Rate</p>
                  <p>Elite: &lt;1% | Good: 1-4%</p>
                </div>
                <div>
                  <p className="font-medium text-gray-300">Rework Rate</p>
                  <p>Elite: &lt;3% | Good: 3-5%</p>
                </div>
              </div>
              <p className="text-xs text-gray-500 mt-3">
                Based on 8.1M PRs, 4,813 teams, 163,820 contributors
              </p>
            </CardContent>
          </Card>
        </>
      ) : (
        <Card className="text-center py-12">
          <Zap className="h-12 w-12 text-gray-600 mx-auto mb-4" />
          <p className="text-gray-400">No speed/delivery data available</p>
          <p className="text-sm text-gray-500 mt-1">Run scans with the devops scanner</p>
        </Card>
      )}
    </div>
  );
}

export default function SpeedPage() {
  return (
    <MainLayout>
      <Suspense fallback={<div className="animate-pulse h-96 bg-gray-800 rounded-lg" />}>
        <SpeedContent />
      </Suspense>
    </MainLayout>
  );
}
