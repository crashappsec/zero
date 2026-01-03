'use client';

import { MainLayout } from '@/components/layout/Sidebar';
import { Card, CardTitle, CardContent } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { ProjectCard, ProjectCardSkeleton } from '@/components/projects/ProjectCard';
import { ActivityFeed, useActivityFeed, ActivityEvent } from '@/components/ActivityFeed';
import { useProjects, useQueueStats, useActiveScans, useFetch } from '@/hooks/useApi';
import { api } from '@/lib/api';
import type { AggregateStats } from '@/lib/types';
import {
  Plus,
  Scan,
  Shield,
  AlertTriangle,
  CheckCircle,
  Clock,
  Package,
  Key,
  TrendingUp,
  BarChart3,
} from 'lucide-react';
import Link from 'next/link';
import { useEffect } from 'react';

function AggregateStatsCards() {
  const { data: stats, loading } = useFetch(() => api.analysis.stats(), []);

  if (loading || !stats) {
    return (
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {[1, 2, 3, 4].map((i) => (
          <Card key={i} className="animate-pulse">
            <div className="h-20 bg-gray-700 rounded" />
          </Card>
        ))}
      </div>
    );
  }

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Card>
        <div className="flex items-center gap-3">
          <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-red-600/20">
            <AlertTriangle className="h-6 w-6 text-red-500" />
          </div>
          <div>
            <p className="text-sm text-gray-400">Total Vulnerabilities</p>
            <div className="flex items-baseline gap-2">
              <p className="text-3xl font-bold text-white">{stats.total_vulns}</p>
              {stats.vulns_by_severity.critical > 0 && (
                <Badge variant="error" className="text-xs">
                  {stats.vulns_by_severity.critical} critical
                </Badge>
              )}
            </div>
          </div>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-3">
          <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-yellow-600/20">
            <Key className="h-6 w-6 text-yellow-500" />
          </div>
          <div>
            <p className="text-sm text-gray-400">Secrets Detected</p>
            <p className="text-3xl font-bold text-white">{stats.total_secrets}</p>
          </div>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-3">
          <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-blue-600/20">
            <Package className="h-6 w-6 text-blue-500" />
          </div>
          <div>
            <p className="text-sm text-gray-400">Dependencies</p>
            <p className="text-3xl font-bold text-white">{stats.total_deps}</p>
          </div>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-3">
          <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-green-600/20">
            <Shield className="h-6 w-6 text-green-500" />
          </div>
          <div>
            <p className="text-sm text-gray-400">Projects Scanned</p>
            <p className="text-3xl font-bold text-white">{stats.total_projects}</p>
          </div>
        </div>
      </Card>
    </div>
  );
}

function SeverityBreakdown() {
  const { data: stats, loading } = useFetch(() => api.analysis.stats(), []);

  if (loading || !stats || stats.total_vulns === 0) {
    return null;
  }

  const severities = [
    { name: 'Critical', count: stats.vulns_by_severity.critical || 0, color: 'bg-red-500' },
    { name: 'High', count: stats.vulns_by_severity.high || 0, color: 'bg-orange-500' },
    { name: 'Medium', count: stats.vulns_by_severity.medium || 0, color: 'bg-yellow-500' },
    { name: 'Low', count: stats.vulns_by_severity.low || 0, color: 'bg-blue-500' },
  ];

  const total = stats.total_vulns || 1;

  return (
    <Card>
      <CardTitle className="flex items-center gap-2">
        <BarChart3 className="h-5 w-5 text-purple-500" />
        Vulnerability Breakdown
      </CardTitle>
      <CardContent className="mt-4">
        <div className="space-y-3">
          {severities.map((sev) => (
            <div key={sev.name}>
              <div className="flex items-center justify-between text-sm mb-1">
                <span className="text-gray-400">{sev.name}</span>
                <span className="text-white font-medium">{sev.count}</span>
              </div>
              <div className="h-2 bg-gray-700 rounded-full overflow-hidden">
                <div
                  className={`h-full ${sev.color} transition-all duration-500`}
                  style={{ width: `${(sev.count / total) * 100}%` }}
                />
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

function TopIssuesCard() {
  const { data: stats, loading } = useFetch(() => api.analysis.stats(), []);

  if (loading || !stats) {
    return null;
  }

  // Sort projects by vulnerabilities
  const topProjects = [...stats.project_stats]
    .sort((a, b) => b.vulns - a.vulns)
    .slice(0, 5)
    .filter((p) => p.vulns > 0);

  if (topProjects.length === 0) {
    return null;
  }

  return (
    <Card>
      <CardTitle className="flex items-center gap-2">
        <TrendingUp className="h-5 w-5 text-orange-500" />
        Projects with Most Issues
      </CardTitle>
      <CardContent className="mt-4">
        <div className="space-y-3">
          {topProjects.map((project) => (
            <Link
              key={project.id}
              href={`/projects/${encodeURIComponent(project.id)}`}
              className="flex items-center justify-between p-2 rounded-lg hover:bg-gray-800 transition-colors"
            >
              <span className="text-white truncate">{project.id}</span>
              <div className="flex items-center gap-2">
                {project.severity.critical > 0 && (
                  <Badge variant="error" className="text-xs">
                    {project.severity.critical} critical
                  </Badge>
                )}
                <span className="text-gray-400 text-sm">{project.vulns} vulns</span>
              </div>
            </Link>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

function ScanStatsCards() {
  const stats = useQueueStats();
  const activeScans = useActiveScans();

  return (
    <div className="grid gap-4 md:grid-cols-4">
      <Card>
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-green-600/20">
            <CheckCircle className="h-5 w-5 text-green-500" />
          </div>
          <div>
            <p className="text-sm text-gray-400">Completed Scans</p>
            <p className="text-2xl font-bold text-white">{stats?.completed_jobs ?? '-'}</p>
          </div>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-600/20">
            <Scan className="h-5 w-5 text-blue-500" />
          </div>
          <div>
            <p className="text-sm text-gray-400">Active Scans</p>
            <p className="text-2xl font-bold text-white">{activeScans.length}</p>
          </div>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-yellow-600/20">
            <Clock className="h-5 w-5 text-yellow-500" />
          </div>
          <div>
            <p className="text-sm text-gray-400">Queued</p>
            <p className="text-2xl font-bold text-white">{stats?.queued_jobs ?? '-'}</p>
          </div>
        </div>
      </Card>

      <Card>
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-red-600/20">
            <AlertTriangle className="h-5 w-5 text-red-500" />
          </div>
          <div>
            <p className="text-sm text-gray-400">Failed</p>
            <p className="text-2xl font-bold text-white">{stats?.failed_jobs ?? '-'}</p>
          </div>
        </div>
      </Card>
    </div>
  );
}

function ActiveScansSection() {
  const activeScans = useActiveScans();
  const { addEvent } = useActivityFeed();

  // Track scan changes for activity feed
  useEffect(() => {
    activeScans.forEach((scan) => {
      if (scan.status === 'scanning' && scan.progress?.current_scanner) {
        // This would normally be tracked via WebSocket
      }
    });
  }, [activeScans, addEvent]);

  if (activeScans.length === 0) return null;

  return (
    <section>
      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-lg font-semibold text-white">Active Scans</h2>
        <Link href="/scans">
          <Button variant="ghost" size="sm">
            View All
          </Button>
        </Link>
      </div>
      <div className="space-y-2">
        {activeScans.map((scan) => (
          <Card key={scan.job_id} className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="relative">
                <Scan className="h-5 w-5 text-blue-500" />
                <span className="absolute -right-1 -top-1 h-2 w-2 rounded-full bg-blue-500 animate-pulse" />
              </div>
              <div>
                <p className="font-medium text-white">{scan.target}</p>
                <p className="text-sm text-gray-500">
                  {scan.progress?.current_scanner || scan.progress?.phase || scan.status} - {scan.profile}
                </p>
              </div>
            </div>
            <div className="text-right">
              <p className="text-sm text-gray-400">
                {scan.progress?.scanners_complete || 0} / {scan.progress?.scanners_total || '?'} scanners
              </p>
              {scan.progress?.scanners_total && (
                <div className="w-24 h-1.5 bg-gray-700 rounded-full mt-1 overflow-hidden">
                  <div
                    className="h-full bg-blue-500 transition-all"
                    style={{
                      width: `${((scan.progress.scanners_complete || 0) / scan.progress.scanners_total) * 100}%`
                    }}
                  />
                </div>
              )}
            </div>
          </Card>
        ))}
      </div>
    </section>
  );
}

function ProjectsSection() {
  const { data: projects, loading, error } = useProjects();

  return (
    <section>
      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-lg font-semibold text-white">Recent Projects</h2>
        <Link href="/scans">
          <Button variant="primary" size="sm" icon={<Plus className="h-4 w-4" />}>
            New Scan
          </Button>
        </Link>
      </div>

      {error && (
        <Card className="border-red-700 bg-red-900/20">
          <p className="text-red-400">Failed to load projects: {error.message}</p>
        </Card>
      )}

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {loading ? (
          <>
            <ProjectCardSkeleton />
            <ProjectCardSkeleton />
            <ProjectCardSkeleton />
          </>
        ) : projects && projects.length > 0 ? (
          projects.slice(0, 6).map((project) => (
            <ProjectCard key={project.id} project={project} />
          ))
        ) : (
          <Card className="col-span-full text-center py-12">
            <Shield className="mx-auto h-12 w-12 text-gray-600" />
            <h3 className="mt-4 text-lg font-medium text-white">No projects yet</h3>
            <p className="mt-2 text-gray-400">Start by scanning a repository</p>
            <Link href="/scans" className="mt-4 inline-block">
              <Button variant="primary" icon={<Plus className="h-4 w-4" />}>
                New Scan
              </Button>
            </Link>
          </Card>
        )}
      </div>

      {projects && projects.length > 6 && (
        <div className="mt-4 text-center">
          <Link href="/projects">
            <Button variant="ghost">View All Projects</Button>
          </Link>
        </div>
      )}
    </section>
  );
}

export default function DashboardPage() {
  return (
    <MainLayout>
      <div className="space-y-8">
        {/* Header */}
        <div>
          <h1 className="text-2xl font-bold text-white">Dashboard</h1>
          <p className="mt-1 text-gray-400">Overview of your repository analysis</p>
        </div>

        {/* Aggregate Stats */}
        <AggregateStatsCards />

        {/* Analytics Grid */}
        <div className="grid gap-6 lg:grid-cols-2">
          <SeverityBreakdown />
          <TopIssuesCard />
        </div>

        {/* Scan Stats */}
        <div>
          <h2 className="text-lg font-semibold text-white mb-4">Scan Activity</h2>
          <ScanStatsCards />
        </div>

        {/* Active Scans */}
        <ActiveScansSection />

        {/* Projects */}
        <ProjectsSection />
      </div>
    </MainLayout>
  );
}
