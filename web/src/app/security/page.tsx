'use client';

import { useState, Suspense, useEffect } from 'react';
import { MainLayout } from '@/components/layout/Sidebar';
import { Card, CardTitle, CardContent } from '@/components/ui/Card';
import { Badge, SeverityBadge } from '@/components/ui/Badge';
import { BenchmarkTier, SECURITY_BENCHMARKS } from '@/components/ui/BenchmarkTier';
import { useFetch } from '@/hooks/useApi';
import { api } from '@/lib/api';
import type { Vulnerability, Secret } from '@/lib/types';
import { ProjectFilter } from '@/components/ui/ProjectFilter';
import {
  Shield,
  Key,
  Lock,
  AlertTriangle,
  ChevronRight,
  Info,
} from 'lucide-react';
import Link from 'next/link';

interface SecurityStats {
  vulnerabilities: { critical: number; high: number; medium: number; low: number; total: number };
  secrets: { critical: number; high: number; medium: number; low: number; total: number };
}

function SecurityContent() {
  const [selectedProjects, setSelectedProjects] = useState<string[]>([]);
  const [stats, setStats] = useState<SecurityStats>({
    vulnerabilities: { critical: 0, high: 0, medium: 0, low: 0, total: 0 },
    secrets: { critical: 0, high: 0, medium: 0, low: 0, total: 0 },
  });
  const [loading, setLoading] = useState(true);
  const [recentVulns, setRecentVulns] = useState<(Vulnerability & { projectId: string })[]>([]);
  const [recentSecrets, setRecentSecrets] = useState<(Secret & { projectId: string })[]>([]);

  // Fetch projects
  const { data: projectsData } = useFetch(() => api.projects.list(), []);
  const projects = projectsData?.data || [];

  // Load all security data
  useEffect(() => {
    async function loadSecurityData() {
      if (projects.length === 0) return;
      setLoading(true);

      const vulnStats = { critical: 0, high: 0, medium: 0, low: 0, total: 0 };
      const secretStats = { critical: 0, high: 0, medium: 0, low: 0, total: 0 };
      const allVulns: (Vulnerability & { projectId: string })[] = [];
      const allSecrets: (Secret & { projectId: string })[] = [];

      const filteredProjects = selectedProjects.length > 0
        ? projects.filter(p => selectedProjects.includes(p.id))
        : projects;

      await Promise.all(
        filteredProjects.map(async (project) => {
          try {
            // Fetch vulnerabilities
            const vulnsData = await api.analysis.vulnerabilities(project.id);
            if (vulnsData?.data) {
              vulnsData.data.forEach(v => {
                const sev = v.severity.toLowerCase() as keyof typeof vulnStats;
                if (sev in vulnStats && sev !== 'total') vulnStats[sev]++;
                vulnStats.total++;
                allVulns.push({ ...v, projectId: project.id });
              });
            }

            // Fetch secrets
            const secretsData = await api.analysis.secrets(project.id);
            if (secretsData?.data) {
              secretsData.data.forEach(s => {
                const sev = s.severity.toLowerCase() as keyof typeof secretStats;
                if (sev in secretStats && sev !== 'total') secretStats[sev]++;
                secretStats.total++;
                allSecrets.push({ ...s, projectId: project.id });
              });
            }
          } catch {
            // Skip projects without data
          }
        })
      );

      setStats({ vulnerabilities: vulnStats, secrets: secretStats });

      // Sort by severity and take top 5
      const severityOrder: Record<string, number> = { critical: 0, high: 1, medium: 2, low: 3 };
      allVulns.sort((a, b) =>
        (severityOrder[a.severity.toLowerCase()] || 99) - (severityOrder[b.severity.toLowerCase()] || 99)
      );
      allSecrets.sort((a, b) =>
        (severityOrder[a.severity.toLowerCase()] || 99) - (severityOrder[b.severity.toLowerCase()] || 99)
      );

      setRecentVulns(allVulns.slice(0, 5));
      setRecentSecrets(allSecrets.slice(0, 5));
      setLoading(false);
    }

    loadSecurityData();
  }, [projects, selectedProjects]);

  const totalCritical = stats.vulnerabilities.critical + stats.secrets.critical;
  const totalHigh = stats.vulnerabilities.high + stats.secrets.high;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            <Shield className="h-6 w-6 text-red-500" />
            Security
          </h1>
          <p className="mt-1 text-gray-400">
            Vulnerabilities, secrets, and cryptographic issues
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
          {/* Benchmark Metrics */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <BenchmarkTier
              value={totalCritical}
              label="Critical Issues"
              tiers={SECURITY_BENCHMARKS.criticalVulns}
              lowerIsBetter={true}
            />
            <BenchmarkTier
              value={totalHigh}
              label="High Severity"
              tiers={SECURITY_BENCHMARKS.highVulns}
              lowerIsBetter={true}
            />
            <BenchmarkTier
              value={stats.secrets.total}
              label="Secrets Exposed"
              tiers={SECURITY_BENCHMARKS.secretsExposed}
              lowerIsBetter={true}
            />
            <Card className="p-4">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm text-gray-400">Total Findings</span>
              </div>
              <p className="text-2xl font-bold text-blue-500">
                {stats.vulnerabilities.total + stats.secrets.total}
              </p>
              <p className="text-xs text-gray-500 mt-2">
                {stats.vulnerabilities.total} vulns, {stats.secrets.total} secrets
              </p>
            </Card>
          </div>

          {/* Vulnerabilities Section */}
          <Card>
            <CardContent>
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-2">
                  <AlertTriangle className="h-5 w-5 text-red-500" />
                  <CardTitle>Vulnerabilities</CardTitle>
                </div>
                <Link
                  href="/vulnerabilities"
                  className="flex items-center gap-1 text-sm text-blue-400 hover:text-blue-300"
                >
                  View all <ChevronRight className="h-4 w-4" />
                </Link>
              </div>

              <div className="grid grid-cols-4 gap-4 mb-4">
                <div className="text-center p-2 bg-red-500/10 rounded">
                  <p className="text-xl font-bold text-red-500">{stats.vulnerabilities.critical}</p>
                  <p className="text-xs text-gray-400">Critical</p>
                </div>
                <div className="text-center p-2 bg-orange-500/10 rounded">
                  <p className="text-xl font-bold text-orange-500">{stats.vulnerabilities.high}</p>
                  <p className="text-xs text-gray-400">High</p>
                </div>
                <div className="text-center p-2 bg-yellow-500/10 rounded">
                  <p className="text-xl font-bold text-yellow-500">{stats.vulnerabilities.medium}</p>
                  <p className="text-xs text-gray-400">Medium</p>
                </div>
                <div className="text-center p-2 bg-blue-500/10 rounded">
                  <p className="text-xl font-bold text-blue-500">{stats.vulnerabilities.low}</p>
                  <p className="text-xs text-gray-400">Low</p>
                </div>
              </div>

              {recentVulns.length > 0 ? (
                <div className="space-y-2">
                  {recentVulns.map((vuln, i) => (
                    <div
                      key={`${vuln.projectId}-${vuln.id}-${i}`}
                      className="flex items-center gap-3 p-2 bg-gray-800/50 rounded"
                    >
                      <SeverityBadge severity={vuln.severity} />
                      <div className="flex-1 min-w-0">
                        <span className="font-mono text-sm text-blue-400">{vuln.id}</span>
                        <span className="text-gray-500 mx-2">in</span>
                        <span className="text-white">{vuln.package}@{vuln.version}</span>
                      </div>
                      <Badge variant="default" className="text-xs">{vuln.projectId}</Badge>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-gray-500 text-center py-4">No vulnerabilities found</p>
              )}
            </CardContent>
          </Card>

          {/* Secrets Section */}
          <Card>
            <CardContent>
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-2">
                  <Key className="h-5 w-5 text-yellow-500" />
                  <CardTitle>Secrets</CardTitle>
                </div>
                <Link
                  href="/secrets"
                  className="flex items-center gap-1 text-sm text-blue-400 hover:text-blue-300"
                >
                  View all <ChevronRight className="h-4 w-4" />
                </Link>
              </div>

              <div className="grid grid-cols-4 gap-4 mb-4">
                <div className="text-center p-2 bg-red-500/10 rounded">
                  <p className="text-xl font-bold text-red-500">{stats.secrets.critical}</p>
                  <p className="text-xs text-gray-400">Critical</p>
                </div>
                <div className="text-center p-2 bg-orange-500/10 rounded">
                  <p className="text-xl font-bold text-orange-500">{stats.secrets.high}</p>
                  <p className="text-xs text-gray-400">High</p>
                </div>
                <div className="text-center p-2 bg-yellow-500/10 rounded">
                  <p className="text-xl font-bold text-yellow-500">{stats.secrets.medium}</p>
                  <p className="text-xs text-gray-400">Medium</p>
                </div>
                <div className="text-center p-2 bg-blue-500/10 rounded">
                  <p className="text-xl font-bold text-blue-500">{stats.secrets.low}</p>
                  <p className="text-xs text-gray-400">Low</p>
                </div>
              </div>

              {recentSecrets.length > 0 ? (
                <div className="space-y-2">
                  {recentSecrets.map((secret, i) => (
                    <div
                      key={`${secret.projectId}-${secret.file}-${i}`}
                      className="flex items-center gap-3 p-2 bg-gray-800/50 rounded"
                    >
                      <SeverityBadge severity={secret.severity} />
                      <Lock className="h-4 w-4 text-yellow-500" />
                      <div className="flex-1 min-w-0">
                        <span className="font-medium text-white">{secret.type}</span>
                        <span className="text-gray-500 mx-2">in</span>
                        <span className="text-gray-400 truncate">{secret.file}</span>
                      </div>
                      <Badge variant="default" className="text-xs">{secret.projectId}</Badge>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-gray-500 text-center py-4">No secrets detected</p>
              )}
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}

export default function SecurityPage() {
  return (
    <MainLayout>
      <Suspense fallback={<div className="animate-pulse h-96 bg-gray-800 rounded-lg" />}>
        <SecurityContent />
      </Suspense>
    </MainLayout>
  );
}
