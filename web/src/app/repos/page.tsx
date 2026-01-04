'use client';

import { MainLayout } from '@/components/layout/Sidebar';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { RepoCard, RepoCardSkeleton } from '@/components/repos/RepoCard';
import { useRepos } from '@/hooks/useApi';
import { Plus, Search, RefreshCw } from 'lucide-react';
import Link from 'next/link';
import { useState, useMemo } from 'react';

export default function ReposPage() {
  const { data: repos, loading, error, refetch } = useRepos();
  const [search, setSearch] = useState('');

  const filteredRepos = useMemo(() => {
    if (!repos) return [];
    if (!search) return repos;
    const query = search.toLowerCase();
    return repos.filter(
      (r) =>
        r.id.toLowerCase().includes(query) ||
        r.owner?.toLowerCase().includes(query) ||
        r.repo?.toLowerCase().includes(query)
    );
  }, [repos, search]);

  return (
    <MainLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Repos</h1>
            <p className="mt-1 text-gray-500 dark:text-gray-400">
              {repos?.length || 0} hydrated repositories
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
          placeholder="Search repos..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          icon={<Search className="h-4 w-4" />}
        />

        {/* Error */}
        {error && (
          <Card className="border-red-500 dark:border-red-700 bg-red-50 dark:bg-red-900/20">
            <p className="text-red-600 dark:text-red-400">Failed to load repos: {error.message}</p>
          </Card>
        )}

        {/* Repos Grid */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {loading ? (
            <>
              <RepoCardSkeleton />
              <RepoCardSkeleton />
              <RepoCardSkeleton />
              <RepoCardSkeleton />
              <RepoCardSkeleton />
              <RepoCardSkeleton />
            </>
          ) : filteredRepos.length > 0 ? (
            filteredRepos.map((repo) => (
              <RepoCard key={repo.id} repo={repo} />
            ))
          ) : (
            <Card className="col-span-full text-center py-12">
              <p className="text-gray-500 dark:text-gray-400">
                {search ? 'No repos match your search' : 'No repos found'}
              </p>
            </Card>
          )}
        </div>
      </div>
    </MainLayout>
  );
}
