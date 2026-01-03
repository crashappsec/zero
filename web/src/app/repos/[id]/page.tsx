'use client';

import { useParams } from 'next/navigation';
import { MainLayout } from '@/components/layout/Sidebar';
import { Card, CardTitle, CardContent } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge, SeverityBadge, StatusBadge } from '@/components/ui/Badge';
import { useFetch } from '@/hooks/useApi';
import { api } from '@/lib/api';
import { formatRelativeTime } from '@/lib/utils';
import type { Repo, AnalysisSummary } from '@/lib/types';
import {
  ArrowLeft,
  AlertTriangle,
  Key,
  Package,
  GitBranch,
  Clock,
  RefreshCw,
  MessageSquare,
} from 'lucide-react';
import Link from 'next/link';

function SeverityStats({ summary }: { summary: AnalysisSummary }) {
  const totals = summary.totals || { critical: 0, high: 0, medium: 0, low: 0 };

  return (
    <div className="grid grid-cols-4 gap-4">
      <Card className="text-center">
        <p className="text-3xl font-bold text-red-500">{totals.critical}</p>
        <p className="text-sm text-gray-500 dark:text-gray-400">Critical</p>
      </Card>
      <Card className="text-center">
        <p className="text-3xl font-bold text-orange-500">{totals.high}</p>
        <p className="text-sm text-gray-500 dark:text-gray-400">High</p>
      </Card>
      <Card className="text-center">
        <p className="text-3xl font-bold text-yellow-500">{totals.medium}</p>
        <p className="text-sm text-gray-500 dark:text-gray-400">Medium</p>
      </Card>
      <Card className="text-center">
        <p className="text-3xl font-bold text-blue-500">{totals.low}</p>
        <p className="text-sm text-gray-500 dark:text-gray-400">Low</p>
      </Card>
    </div>
  );
}

function VulnerabilitiesSection({ repoId }: { repoId: string }) {
  const { data, loading } = useFetch(
    () => api.analysis.vulnerabilities(repoId),
    [repoId]
  );

  if (loading) {
    return (
      <Card>
        <CardTitle>Vulnerabilities</CardTitle>
        <div className="mt-4 animate-pulse space-y-2">
          <div className="h-12 rounded bg-gray-200 dark:bg-gray-700" />
          <div className="h-12 rounded bg-gray-200 dark:bg-gray-700" />
          <div className="h-12 rounded bg-gray-200 dark:bg-gray-700" />
        </div>
      </Card>
    );
  }

  const vulns = data?.data || [];

  return (
    <Card>
      <div className="flex items-center justify-between">
        <CardTitle className="flex items-center gap-2">
          <AlertTriangle className="h-5 w-5 text-orange-500" />
          Vulnerabilities ({vulns.length})
        </CardTitle>
      </div>
      <CardContent className="mt-4">
        {vulns.length === 0 ? (
          <p className="text-gray-500 dark:text-gray-400 text-center py-4">No vulnerabilities found</p>
        ) : (
          <div className="space-y-2 max-h-96 overflow-y-auto">
            {vulns.slice(0, 20).map((vuln, i) => (
              <div
                key={`${vuln.id}-${i}`}
                className="flex items-center justify-between rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50 p-3"
              >
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <SeverityBadge severity={vuln.severity} />
                    <span className="font-mono text-sm text-gray-700 dark:text-gray-300 truncate">
                      {vuln.id}
                    </span>
                  </div>
                  <p className="mt-1 text-sm text-gray-500 dark:text-gray-400 truncate">{vuln.title}</p>
                  <p className="text-xs text-gray-400 dark:text-gray-500">
                    {vuln.package}@{vuln.version}
                    {vuln.fix_version && ` → ${vuln.fix_version}`}
                  </p>
                </div>
              </div>
            ))}
            {vulns.length > 20 && (
              <p className="text-center text-sm text-gray-400 dark:text-gray-500">
                +{vulns.length - 20} more vulnerabilities
              </p>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function SecretsSection({ repoId }: { repoId: string }) {
  const { data, loading } = useFetch(
    () => api.analysis.secrets(repoId),
    [repoId]
  );

  if (loading) {
    return (
      <Card>
        <CardTitle>Secrets</CardTitle>
        <div className="mt-4 animate-pulse space-y-2">
          <div className="h-12 rounded bg-gray-200 dark:bg-gray-700" />
          <div className="h-12 rounded bg-gray-200 dark:bg-gray-700" />
        </div>
      </Card>
    );
  }

  const secrets = data?.data || [];

  return (
    <Card>
      <CardTitle className="flex items-center gap-2">
        <Key className="h-5 w-5 text-yellow-500" />
        Secrets ({secrets.length})
      </CardTitle>
      <CardContent className="mt-4">
        {secrets.length === 0 ? (
          <p className="text-gray-500 dark:text-gray-400 text-center py-4">No secrets detected</p>
        ) : (
          <div className="space-y-2 max-h-64 overflow-y-auto">
            {secrets.slice(0, 10).map((secret, i) => (
              <div
                key={`${secret.file}-${secret.line}-${i}`}
                className="rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50 p-3"
              >
                <div className="flex items-center gap-2">
                  <SeverityBadge severity={secret.severity} />
                  <span className="text-sm text-gray-700 dark:text-gray-300">{secret.type}</span>
                </div>
                <p className="mt-1 font-mono text-xs text-gray-400 dark:text-gray-500">
                  {secret.file}:{secret.line}
                </p>
              </div>
            ))}
            {secrets.length > 10 && (
              <p className="text-center text-sm text-gray-400 dark:text-gray-500">
                +{secrets.length - 10} more secrets
              </p>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function ScannersSection({ summary }: { summary: AnalysisSummary }) {
  const scanners = Object.entries(summary.scanners || {});

  return (
    <Card>
      <CardTitle className="flex items-center gap-2">
        <Package className="h-5 w-5 text-blue-500" />
        Scanners ({scanners.length})
      </CardTitle>
      <CardContent className="mt-4">
        <div className="space-y-2">
          {scanners.map(([name, data]) => (
            <div
              key={name}
              className="flex items-center justify-between rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50 p-3"
            >
              <div>
                <p className="font-medium text-gray-900 dark:text-white">{name}</p>
                <p className="text-xs text-gray-400 dark:text-gray-500">
                  {data.findings_count} findings
                </p>
              </div>
              <StatusBadge status={data.status} />
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

export default function RepoDetailPage() {
  const params = useParams();
  const repoId = decodeURIComponent(params.id as string);

  const { data: repo, loading: repoLoading } = useFetch(
    () => api.repos.get(repoId),
    [repoId]
  );

  const { data: summary, loading: summaryLoading } = useFetch(
    () => api.analysis.summary(repoId),
    [repoId]
  );

  const loading = repoLoading || summaryLoading;

  if (loading) {
    return (
      <MainLayout>
        <div className="animate-pulse space-y-6">
          <div className="h-8 w-48 rounded bg-gray-200 dark:bg-gray-700" />
          <div className="h-24 rounded bg-gray-200 dark:bg-gray-700" />
          <div className="grid grid-cols-4 gap-4">
            <div className="h-24 rounded bg-gray-200 dark:bg-gray-700" />
            <div className="h-24 rounded bg-gray-200 dark:bg-gray-700" />
            <div className="h-24 rounded bg-gray-200 dark:bg-gray-700" />
            <div className="h-24 rounded bg-gray-200 dark:bg-gray-700" />
          </div>
        </div>
      </MainLayout>
    );
  }

  return (
    <MainLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link href="/repos">
              <Button variant="ghost" size="sm" icon={<ArrowLeft className="h-4 w-4" />}>
                Back
              </Button>
            </Link>
            <div>
              <h1 className="text-2xl font-bold text-gray-900 dark:text-white flex items-center gap-2">
                <GitBranch className="h-6 w-6 text-gray-400" />
                {repoId}
              </h1>
              {repo?.last_scan && (
                <p className="mt-1 text-sm text-gray-500 dark:text-gray-400 flex items-center gap-1">
                  <Clock className="h-4 w-4" />
                  Last scanned {formatRelativeTime(repo.last_scan)}
                </p>
              )}
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Link href={`/chat?repo=${encodeURIComponent(repoId)}`}>
              <Button variant="secondary" size="sm" icon={<MessageSquare className="h-4 w-4" />}>
                Ask Agent
              </Button>
            </Link>
            <Link href={`/scans?target=${encodeURIComponent(repoId)}`}>
              <Button variant="primary" size="sm" icon={<RefreshCw className="h-4 w-4" />}>
                Rescan
              </Button>
            </Link>
          </div>
        </div>

        {/* Severity Stats */}
        {summary && <SeverityStats summary={summary} />}

        {/* Details Grid */}
        <div className="grid gap-6 lg:grid-cols-2">
          <VulnerabilitiesSection repoId={repoId} />
          <div className="space-y-6">
            <SecretsSection repoId={repoId} />
            {summary && <ScannersSection summary={summary} />}
          </div>
        </div>
      </div>
    </MainLayout>
  );
}
