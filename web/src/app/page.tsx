'use client';

import { MainLayout } from '@/components/layout/Sidebar';
import { Card, CardTitle } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { ProjectCard, ProjectCardSkeleton } from '@/components/projects/ProjectCard';
import { useProjects, useQueueStats, useActiveScans } from '@/hooks/useApi';
import { Plus, Scan, Shield, AlertTriangle, CheckCircle, Clock } from 'lucide-react';
import Link from 'next/link';

function StatsCards() {
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
                  {scan.progress?.phase || scan.status} - {scan.profile}
                </p>
              </div>
            </div>
            <div className="text-right">
              <p className="text-sm text-gray-400">
                {scan.progress?.scanners_complete || 0} / {scan.progress?.scanners_total || '?'} scanners
              </p>
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
          <p className="mt-1 text-gray-400">Overview of your security analysis</p>
        </div>

        {/* Stats */}
        <StatsCards />

        {/* Active Scans */}
        <ActiveScansSection />

        {/* Projects */}
        <ProjectsSection />
      </div>
    </MainLayout>
  );
}
