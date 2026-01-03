'use client';

import { useState, useMemo, Suspense } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { MainLayout } from '@/components/layout/Sidebar';
import { Card, CardTitle, CardContent } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { useFetch } from '@/hooks/useApi';
import { api } from '@/lib/api';
import {
  Users,
  User,
  GitCommit,
  AlertTriangle,
  TrendingUp,
  Calendar,
  FileCode,
  Activity,
} from 'lucide-react';

interface Contributor {
  name: string;
  email: string;
  commits: number;
  lines_added: number;
  lines_removed: number;
  first_commit: string;
  last_commit: string;
  files_touched: number;
}

interface OwnershipData {
  bus_factor: number;
  total_contributors: number;
  active_contributors: number;
  orphan_files: number;
  top_contributors: Contributor[];
  ownership_distribution: Record<string, number>;
  churn_hotspots: Array<{ file: string; changes: number }>;
}

function BusFactorCard({ busFactor }: { busFactor: number }) {
  const getColor = (bf: number) => {
    if (bf <= 1) return 'text-red-500';
    if (bf <= 2) return 'text-orange-500';
    if (bf <= 3) return 'text-yellow-500';
    return 'text-green-500';
  };

  const getRisk = (bf: number) => {
    if (bf <= 1) return { level: 'Critical', desc: 'Single point of failure' };
    if (bf <= 2) return { level: 'High', desc: 'Limited knowledge distribution' };
    if (bf <= 3) return { level: 'Medium', desc: 'Moderate risk' };
    return { level: 'Low', desc: 'Good knowledge distribution' };
  };

  const risk = getRisk(busFactor);

  return (
    <Card>
      <div className="flex items-center gap-4">
        <div className="flex h-14 w-14 items-center justify-center rounded-lg bg-purple-600/20">
          <Users className="h-7 w-7 text-purple-500" />
        </div>
        <div>
          <p className="text-sm text-gray-400">Bus Factor</p>
          <p className={`text-3xl font-bold ${getColor(busFactor)}`}>{busFactor}</p>
          <div className="flex items-center gap-2 mt-1">
            <Badge variant={busFactor <= 2 ? 'error' : busFactor <= 3 ? 'warning' : 'success'}>
              {risk.level} Risk
            </Badge>
            <span className="text-xs text-gray-500">{risk.desc}</span>
          </div>
        </div>
      </div>
    </Card>
  );
}

function ContributorRow({ contributor, rank }: { contributor: Contributor; rank: number }) {
  return (
    <div className="flex items-center gap-4 px-4 py-3 border-b border-gray-700 last:border-0 hover:bg-gray-800/50">
      <div className="flex h-8 w-8 items-center justify-center rounded-full bg-gray-700 text-sm font-medium text-white">
        {rank}
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <User className="h-4 w-4 text-gray-500" />
          <span className="font-medium text-white truncate">{contributor.name}</span>
        </div>
        <p className="text-sm text-gray-500 truncate">{contributor.email}</p>
      </div>
      <div className="flex items-center gap-6 text-sm">
        <div className="text-center">
          <p className="font-medium text-white">{contributor.commits}</p>
          <p className="text-xs text-gray-500">commits</p>
        </div>
        <div className="text-center">
          <p className="font-medium text-green-500">+{contributor.lines_added.toLocaleString()}</p>
          <p className="text-xs text-gray-500">added</p>
        </div>
        <div className="text-center">
          <p className="font-medium text-red-500">-{contributor.lines_removed.toLocaleString()}</p>
          <p className="text-xs text-gray-500">removed</p>
        </div>
        <div className="text-center">
          <p className="font-medium text-white">{contributor.files_touched}</p>
          <p className="text-xs text-gray-500">files</p>
        </div>
      </div>
    </div>
  );
}

function OwnershipContent() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const projectId = searchParams.get('project');

  const { data: projectsData } = useFetch(() => api.projects.list(), []);
  const projects = projectsData?.data || [];

  const { data: ownershipData, loading, error } = useFetch(
    () => projectId ? api.analysis.raw(projectId, 'code-ownership') as Promise<any> : Promise.resolve(null),
    [projectId]
  );

  const ownership = useMemo(() => {
    if (!ownershipData?.findings) return null;
    const findings = ownershipData.findings;
    return {
      bus_factor: findings.bus_factor?.bus_factor || 0,
      total_contributors: findings.contributors?.total || 0,
      active_contributors: findings.contributors?.active_last_90_days || 0,
      orphan_files: findings.orphans?.count || 0,
      top_contributors: findings.contributors?.top || [],
      ownership_distribution: findings.ownership_distribution || {},
      churn_hotspots: findings.churn?.hotspots || [],
    } as OwnershipData;
  }, [ownershipData]);

  const handleProjectChange = (newProjectId: string) => {
    router.push(`/ownership?project=${encodeURIComponent(newProjectId)}`);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-white flex items-center gap-2">
          <Users className="h-6 w-6 text-purple-500" />
          Code Ownership
        </h1>
        <p className="mt-1 text-gray-400">
          Contributor analysis, bus factor, and ownership distribution
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

      {projectId && ownership && (
        <>
          {/* Stats */}
          <div className="grid gap-4 md:grid-cols-4">
            <BusFactorCard busFactor={ownership.bus_factor} />
            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-blue-600/20">
                  <Users className="h-6 w-6 text-blue-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Total Contributors</p>
                  <p className="text-2xl font-bold text-white">{ownership.total_contributors}</p>
                </div>
              </div>
            </Card>
            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-green-600/20">
                  <Activity className="h-6 w-6 text-green-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Active (90 days)</p>
                  <p className="text-2xl font-bold text-white">{ownership.active_contributors}</p>
                </div>
              </div>
            </Card>
            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-yellow-600/20">
                  <AlertTriangle className="h-6 w-6 text-yellow-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Orphan Files</p>
                  <p className="text-2xl font-bold text-white">{ownership.orphan_files}</p>
                </div>
              </div>
            </Card>
          </div>

          {/* Top Contributors */}
          <Card>
            <CardTitle className="flex items-center gap-2">
              <TrendingUp className="h-5 w-5 text-blue-500" />
              Top Contributors
            </CardTitle>
            <CardContent className="mt-4 p-0">
              {ownership.top_contributors.length === 0 ? (
                <div className="p-8 text-center text-gray-400">
                  No contributor data available
                </div>
              ) : (
                <div>
                  {ownership.top_contributors.slice(0, 10).map((contributor, i) => (
                    <ContributorRow key={contributor.email} contributor={contributor} rank={i + 1} />
                  ))}
                </div>
              )}
            </CardContent>
          </Card>

          {/* Churn Hotspots */}
          {ownership.churn_hotspots.length > 0 && (
            <Card>
              <CardTitle className="flex items-center gap-2">
                <GitCommit className="h-5 w-5 text-orange-500" />
                Churn Hotspots
              </CardTitle>
              <CardContent className="mt-4">
                <p className="text-sm text-gray-400 mb-4">
                  Files with the most changes - potential complexity or stability issues
                </p>
                <div className="space-y-2">
                  {ownership.churn_hotspots.slice(0, 10).map((hotspot) => (
                    <div key={hotspot.file} className="flex items-center justify-between p-2 bg-gray-800/50 rounded">
                      <div className="flex items-center gap-2 min-w-0">
                        <FileCode className="h-4 w-4 text-gray-500 shrink-0" />
                        <span className="text-sm text-white truncate">{hotspot.file}</span>
                      </div>
                      <Badge variant="warning">{hotspot.changes} changes</Badge>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}
        </>
      )}

      {projectId && loading && (
        <Card className="p-8 text-center text-gray-400">Loading ownership data...</Card>
      )}

      {projectId && error && (
        <Card className="p-8 text-center text-red-400">
          No code ownership data available. Run a scan with the code-ownership scanner.
        </Card>
      )}

      {!projectId && (
        <Card className="text-center py-12">
          <Users className="h-12 w-12 text-gray-600 mx-auto mb-4" />
          <p className="text-gray-400">Select a project to view code ownership analysis</p>
        </Card>
      )}
    </div>
  );
}

export default function OwnershipPage() {
  return (
    <MainLayout>
      <Suspense fallback={<div className="animate-pulse h-96 bg-gray-800 rounded-lg" />}>
        <OwnershipContent />
      </Suspense>
    </MainLayout>
  );
}
