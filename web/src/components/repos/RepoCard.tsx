'use client';

import Link from 'next/link';
import { Card } from '@/components/ui/Card';
import { Badge, StatusBadge } from '@/components/ui/Badge';
import { formatRelativeTime, getFreshnessIndicator } from '@/lib/utils';
import type { Repo } from '@/lib/types';
import { GitBranch, Clock } from 'lucide-react';

interface RepoCardProps {
  repo: Repo;
}

export function RepoCard({ repo }: RepoCardProps) {
  const freshness = repo.freshness
    ? getFreshnessIndicator(repo.freshness.level)
    : { color: 'bg-gray-500', label: 'Unknown' };

  return (
    <Link href={`/repos/${encodeURIComponent(repo.id)}`}>
      <Card hover className="h-full">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-gray-100 dark:bg-gray-700">
              <GitBranch className="h-5 w-5 text-gray-500 dark:text-gray-400" />
            </div>
            <div>
              <h3 className="font-semibold text-gray-900 dark:text-white">{repo.id}</h3>
              <p className="text-sm text-gray-500">{repo.owner}/{repo.repo}</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <span className={`h-2 w-2 rounded-full ${freshness.color}`} title={freshness.label} />
          </div>
        </div>

        <div className="mt-4 flex flex-wrap gap-2">
          {repo.scanners?.slice(0, 4).map((scanner) => (
            <Badge key={scanner} variant="default" size="sm">
              {scanner}
            </Badge>
          ))}
          {repo.scanners && repo.scanners.length > 4 && (
            <Badge variant="default" size="sm">
              +{repo.scanners.length - 4}
            </Badge>
          )}
        </div>

        <div className="mt-4 flex items-center justify-between text-sm text-gray-500">
          <div className="flex items-center gap-1">
            <Clock className="h-4 w-4" />
            <span>{repo.last_scan ? formatRelativeTime(repo.last_scan) : 'Never scanned'}</span>
          </div>
          {repo.status && <StatusBadge status={repo.status} />}
        </div>
      </Card>
    </Link>
  );
}

export function RepoCardSkeleton() {
  return (
    <Card className="h-full animate-pulse">
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <div className="h-10 w-10 rounded-lg bg-gray-200 dark:bg-gray-700" />
          <div>
            <div className="h-5 w-32 rounded bg-gray-200 dark:bg-gray-700" />
            <div className="mt-1 h-4 w-24 rounded bg-gray-200 dark:bg-gray-700" />
          </div>
        </div>
      </div>
      <div className="mt-4 flex gap-2">
        <div className="h-5 w-16 rounded bg-gray-200 dark:bg-gray-700" />
        <div className="h-5 w-20 rounded bg-gray-200 dark:bg-gray-700" />
      </div>
      <div className="mt-4 h-4 w-24 rounded bg-gray-200 dark:bg-gray-700" />
    </Card>
  );
}
