'use client';

import { useState, Suspense, useEffect } from 'react';
import { MainLayout } from '@/components/layout/Sidebar';
import { Card, CardTitle, CardContent } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { useFetch } from '@/hooks/useApi';
import { api } from '@/lib/api';
import type { Project } from '@/lib/types';
import { ProjectFilter } from '@/components/ui/ProjectFilter';
import {
  Users,
  User,
  AlertTriangle,
  TrendingUp,
  ChevronRight,
  GitCommit,
  UserMinus,
  Clock,
} from 'lucide-react';
import Link from 'next/link';

interface TeamStats {
  totalContributors: number;
  busFactor: number;
  busFactorRisk: string;
  topContributors: { name: string; commits: number; percentage: number }[];
  orphanedFiles: number;
  lastCommitDays: number;
}

function TeamContent() {
  const [selectedProjects, setSelectedProjects] = useState<string[]>([]);
  const [stats, setStats] = useState<TeamStats>({
    totalContributors: 0,
    busFactor: 0,
    busFactorRisk: 'unknown',
    topContributors: [],
    orphanedFiles: 0,
    lastCommitDays: 0,
  });
  const [loading, setLoading] = useState(true);

  // Fetch projects
  const { data: projectsData } = useFetch(() => api.projects.list(), []);
  const projects = projectsData?.data || [];

  // Load team data
  useEffect(() => {
    async function loadTeamData() {
      if (projects.length === 0) return;
      setLoading(true);

      let totalContribs = 0;
      let avgBusFactor = 0;
      let projectCount = 0;
      let orphaned = 0;
      let minLastCommit = Infinity;
      const allContributors: { name: string; commits: number }[] = [];

      const filteredProjects = selectedProjects.length > 0
        ? projects.filter(p => selectedProjects.includes(p.id))
        : projects;

      await Promise.all(
        filteredProjects.map(async (project) => {
          try {
            const ownershipData = await api.analysis.ownership(project.id);
            if (ownershipData?.data) {
              const data = ownershipData.data;

              // Contributors
              if (data.contributors) {
                totalContribs += data.contributors.length;
                data.contributors.forEach((c: any) => {
                  allContributors.push({ name: c.name || c.email, commits: c.commits || 0 });
                });
              }

              // Bus factor
              if (data.bus_factor) {
                avgBusFactor += data.bus_factor.factor || 0;
                projectCount++;
              }

              // Orphaned files
              if (data.orphans) {
                orphaned += data.orphans.length;
              }

              // Last commit
              if (data.last_commit_days !== undefined) {
                minLastCommit = Math.min(minLastCommit, data.last_commit_days);
              }
            }
          } catch {
            // Skip projects without data
          }
        })
      );

      // Calculate average bus factor
      const finalBusFactor = projectCount > 0 ? Math.round(avgBusFactor / projectCount) : 0;

      // Aggregate and sort contributors
      const contribMap = new Map<string, number>();
      allContributors.forEach(c => {
        contribMap.set(c.name, (contribMap.get(c.name) || 0) + c.commits);
      });
      const totalCommits = Array.from(contribMap.values()).reduce((a, b) => a + b, 0);
      const topContribs = Array.from(contribMap.entries())
        .sort((a, b) => b[1] - a[1])
        .slice(0, 5)
        .map(([name, commits]) => ({
          name,
          commits,
          percentage: totalCommits > 0 ? Math.round((commits / totalCommits) * 100) : 0,
        }));

      // Determine bus factor risk
      let risk = 'low';
      if (finalBusFactor <= 1) risk = 'critical';
      else if (finalBusFactor <= 2) risk = 'high';
      else if (finalBusFactor <= 3) risk = 'medium';

      setStats({
        totalContributors: new Set(allContributors.map(c => c.name)).size,
        busFactor: finalBusFactor,
        busFactorRisk: risk,
        topContributors: topContribs,
        orphanedFiles: orphaned,
        lastCommitDays: minLastCommit === Infinity ? 0 : minLastCommit,
      });

      setLoading(false);
    }

    loadTeamData();
  }, [projects, selectedProjects]);

  const busFactorColor = {
    critical: 'text-red-500',
    high: 'text-orange-500',
    medium: 'text-yellow-500',
    low: 'text-green-500',
    unknown: 'text-gray-500',
  }[stats.busFactorRisk];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            <Users className="h-6 w-6 text-purple-500" />
            Team
          </h1>
          <p className="mt-1 text-gray-400">
            Code ownership, bus factor, and contributor analysis
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
      ) : (
        <>
          {/* Overview Stats */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <Card className="text-center">
              <p className="text-3xl font-bold text-purple-500">{stats.totalContributors}</p>
              <p className="text-sm text-gray-400">Contributors</p>
            </Card>
            <Card className="text-center">
              <p className={`text-3xl font-bold ${busFactorColor}`}>{stats.busFactor}</p>
              <p className="text-sm text-gray-400">Bus Factor</p>
            </Card>
            <Card className="text-center">
              <p className="text-3xl font-bold text-orange-500">{stats.orphanedFiles}</p>
              <p className="text-sm text-gray-400">Orphaned Files</p>
            </Card>
            <Card className="text-center">
              <p className="text-3xl font-bold text-blue-500">{stats.lastCommitDays}</p>
              <p className="text-sm text-gray-400">Days Since Commit</p>
            </Card>
          </div>

          {/* Bus Factor Alert */}
          {stats.busFactorRisk === 'critical' || stats.busFactorRisk === 'high' ? (
            <Card className="bg-red-500/10 border-red-500/20">
              <CardContent className="flex items-center gap-4">
                <AlertTriangle className="h-8 w-8 text-red-500" />
                <div>
                  <p className="font-medium text-white">
                    {stats.busFactorRisk === 'critical' ? 'Critical' : 'High'} Bus Factor Risk
                  </p>
                  <p className="text-sm text-gray-400">
                    Only {stats.busFactor} contributor{stats.busFactor !== 1 ? 's' : ''} control most of the codebase.
                    Consider knowledge sharing and documentation.
                  </p>
                </div>
              </CardContent>
            </Card>
          ) : null}

          {/* Top Contributors */}
          <Card>
            <CardContent>
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-2">
                  <User className="h-5 w-5 text-purple-500" />
                  <CardTitle>Top Contributors</CardTitle>
                </div>
                <Link
                  href="/ownership"
                  className="flex items-center gap-1 text-sm text-blue-400 hover:text-blue-300"
                >
                  View all <ChevronRight className="h-4 w-4" />
                </Link>
              </div>

              {stats.topContributors.length > 0 ? (
                <div className="space-y-3">
                  {stats.topContributors.map((contrib, i) => (
                    <div key={contrib.name} className="flex items-center gap-3">
                      <span className="text-sm text-gray-500 w-6">#{i + 1}</span>
                      <div className="flex-1">
                        <div className="flex items-center justify-between mb-1">
                          <span className="text-white font-medium">{contrib.name}</span>
                          <span className="text-sm text-gray-400">
                            {contrib.commits} commits ({contrib.percentage}%)
                          </span>
                        </div>
                        <div className="h-2 bg-gray-700 rounded-full overflow-hidden">
                          <div
                            className="h-full bg-purple-500 rounded-full"
                            style={{ width: `${contrib.percentage}%` }}
                          />
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-gray-500 text-center py-4">No contributor data available</p>
              )}
            </CardContent>
          </Card>

          {/* Quick Links */}
          <div className="grid grid-cols-2 gap-4">
            <Link href="/ownership">
              <Card className="hover:bg-gray-800/50 transition-colors cursor-pointer">
                <CardContent className="flex items-center gap-4">
                  <GitCommit className="h-8 w-8 text-purple-500" />
                  <div>
                    <p className="font-medium text-white">Code Ownership</p>
                    <p className="text-sm text-gray-400">View detailed ownership analysis</p>
                  </div>
                  <ChevronRight className="h-5 w-5 text-gray-500 ml-auto" />
                </CardContent>
              </Card>
            </Link>
            <Link href="/devx">
              <Card className="hover:bg-gray-800/50 transition-colors cursor-pointer">
                <CardContent className="flex items-center gap-4">
                  <TrendingUp className="h-8 w-8 text-green-500" />
                  <div>
                    <p className="font-medium text-white">Developer Experience</p>
                    <p className="text-sm text-gray-400">Onboarding and workflow metrics</p>
                  </div>
                  <ChevronRight className="h-5 w-5 text-gray-500 ml-auto" />
                </CardContent>
              </Card>
            </Link>
          </div>
        </>
      )}
    </div>
  );
}

export default function TeamPage() {
  return (
    <MainLayout>
      <Suspense fallback={<div className="animate-pulse h-96 bg-gray-800 rounded-lg" />}>
        <TeamContent />
      </Suspense>
    </MainLayout>
  );
}
