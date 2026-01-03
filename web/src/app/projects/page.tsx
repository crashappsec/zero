'use client';

import { MainLayout } from '@/components/layout/Sidebar';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { ProjectCard, ProjectCardSkeleton } from '@/components/projects/ProjectCard';
import { useProjects } from '@/hooks/useApi';
import { Plus, Search, RefreshCw } from 'lucide-react';
import Link from 'next/link';
import { useState, useMemo } from 'react';

export default function ProjectsPage() {
  const { data: projects, loading, error, refetch } = useProjects();
  const [search, setSearch] = useState('');

  const filteredProjects = useMemo(() => {
    if (!projects) return [];
    if (!search) return projects;
    const query = search.toLowerCase();
    return projects.filter(
      (p) =>
        p.id.toLowerCase().includes(query) ||
        p.owner?.toLowerCase().includes(query) ||
        p.repo?.toLowerCase().includes(query)
    );
  }, [projects, search]);

  return (
    <MainLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-white">Projects</h1>
            <p className="mt-1 text-gray-400">
              {projects?.length || 0} hydrated repositories
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="secondary" size="sm" onClick={refetch} icon={<RefreshCw className="h-4 w-4" />}>
              Refresh
            </Button>
            <Link href="/scans">
              <Button variant="primary" size="sm" icon={<Plus className="h-4 w-4" />}>
                New Scan
              </Button>
            </Link>
          </div>
        </div>

        {/* Search */}
        <Input
          placeholder="Search projects..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          icon={<Search className="h-4 w-4" />}
        />

        {/* Error */}
        {error && (
          <Card className="border-red-700 bg-red-900/20">
            <p className="text-red-400">Failed to load projects: {error.message}</p>
          </Card>
        )}

        {/* Projects Grid */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {loading ? (
            <>
              <ProjectCardSkeleton />
              <ProjectCardSkeleton />
              <ProjectCardSkeleton />
              <ProjectCardSkeleton />
              <ProjectCardSkeleton />
              <ProjectCardSkeleton />
            </>
          ) : filteredProjects.length > 0 ? (
            filteredProjects.map((project) => (
              <ProjectCard key={project.id} project={project} />
            ))
          ) : (
            <Card className="col-span-full text-center py-12">
              <p className="text-gray-400">
                {search ? 'No projects match your search' : 'No projects found'}
              </p>
            </Card>
          )}
        </div>
      </div>
    </MainLayout>
  );
}
