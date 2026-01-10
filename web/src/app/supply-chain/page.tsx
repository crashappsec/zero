'use client';

import { useState, Suspense, useEffect } from 'react';
import { MainLayout } from '@/components/layout/Sidebar';
import { Card, CardTitle, CardContent } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { BenchmarkTier, SUPPLY_CHAIN_BENCHMARKS } from '@/components/ui/BenchmarkTier';
import { useFetch } from '@/hooks/useApi';
import { api } from '@/lib/api';
import { ProjectFilter } from '@/components/ui/ProjectFilter';
import {
  Package,
  AlertTriangle,
  FileText,
  ChevronRight,
  Info,
} from 'lucide-react';
import Link from 'next/link';

interface SupplyChainStats {
  totalPackages: number;
  vulnerablePackages: number;
  healthyPackages: number;
  licenses: Map<string, number>;
  criticalVulns: number;
  highVulns: number;
}

function SupplyChainContent() {
  const [selectedProjects, setSelectedProjects] = useState<string[]>([]);
  const [stats, setStats] = useState<SupplyChainStats>({
    totalPackages: 0,
    vulnerablePackages: 0,
    healthyPackages: 0,
    licenses: new Map(),
    criticalVulns: 0,
    highVulns: 0,
  });
  const [loading, setLoading] = useState(true);
  const [topLicenses, setTopLicenses] = useState<{ name: string; count: number }[]>([]);

  // Fetch projects
  const { data: projectsData } = useFetch(() => api.projects.list(), []);
  const projects = projectsData?.data || [];

  // Load supply chain data
  useEffect(() => {
    async function loadSupplyChainData() {
      if (projects.length === 0) return;
      setLoading(true);

      let totalPkgs = 0;
      let vulnPkgs = 0;
      let critVulns = 0;
      let highVulns = 0;
      const licenseMap = new Map<string, number>();

      const filteredProjects = selectedProjects.length > 0
        ? projects.filter(p => selectedProjects.includes(p.id))
        : projects;

      await Promise.all(
        filteredProjects.map(async (project) => {
          try {
            // Fetch dependencies
            const depsData = await api.analysis.dependencies(project.id);
            if (depsData?.data) {
              totalPkgs += depsData.data.length;
              depsData.data.forEach(dep => {
                if (dep.license) {
                  licenseMap.set(dep.license, (licenseMap.get(dep.license) || 0) + 1);
                }
              });
            }

            // Fetch vulnerabilities
            const vulnsData = await api.analysis.vulnerabilities(project.id);
            if (vulnsData?.data) {
              const vulnPackages = new Set(vulnsData.data.map(v => v.package));
              vulnPkgs += vulnPackages.size;
              vulnsData.data.forEach(v => {
                const sev = v.severity.toLowerCase();
                if (sev === 'critical') critVulns++;
                if (sev === 'high') highVulns++;
              });
            }
          } catch {
            // Skip projects without data
          }
        })
      );

      setStats({
        totalPackages: totalPkgs,
        vulnerablePackages: vulnPkgs,
        healthyPackages: Math.max(0, totalPkgs - vulnPkgs),
        licenses: licenseMap,
        criticalVulns: critVulns,
        highVulns: highVulns,
      });

      // Get top 5 licenses
      const sortedLicenses = Array.from(licenseMap.entries())
        .sort((a, b) => b[1] - a[1])
        .slice(0, 5)
        .map(([name, count]) => ({ name, count }));
      setTopLicenses(sortedLicenses);

      setLoading(false);
    }

    loadSupplyChainData();
  }, [projects, selectedProjects]);

  const healthScore = stats.totalPackages > 0
    ? Math.round((stats.healthyPackages / stats.totalPackages) * 100)
    : 100;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            <Package className="h-6 w-6 text-blue-500" />
            Supply Chain
          </h1>
          <p className="mt-1 text-gray-400">
            Dependencies, licenses, and package health
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
              value={healthScore}
              label="Package Health"
              tiers={SUPPLY_CHAIN_BENCHMARKS.packageHealth}
              unit="%"
              lowerIsBetter={false}
            />
            <BenchmarkTier
              value={stats.totalPackages > 0 ? Math.round((stats.vulnerablePackages / stats.totalPackages) * 100) : 0}
              label="Vulnerable Deps"
              tiers={SUPPLY_CHAIN_BENCHMARKS.vulnerableDeps}
              unit="%"
              lowerIsBetter={true}
            />
            <BenchmarkTier
              value={stats.criticalVulns + stats.highVulns}
              label="Critical + High Vulns"
              tiers={SUPPLY_CHAIN_BENCHMARKS.licenseViolations}
              lowerIsBetter={true}
            />
            <Card className="p-4">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm text-gray-400">Total Packages</span>
              </div>
              <p className="text-2xl font-bold text-blue-500">{stats.totalPackages}</p>
              <p className="text-xs text-gray-500 mt-2">
                {stats.licenses.size} license types
              </p>
            </Card>
          </div>

          {/* Vulnerabilities Summary */}
          <Card>
            <CardContent>
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-2">
                  <AlertTriangle className="h-5 w-5 text-red-500" />
                  <CardTitle>Vulnerable Dependencies</CardTitle>
                </div>
                <Link
                  href="/vulnerabilities"
                  className="flex items-center gap-1 text-sm text-blue-400 hover:text-blue-300"
                >
                  View details <ChevronRight className="h-4 w-4" />
                </Link>
              </div>

              <div className="grid grid-cols-3 gap-4">
                <div className="text-center p-4 bg-red-500/10 rounded-lg">
                  <p className="text-2xl font-bold text-red-500">{stats.criticalVulns}</p>
                  <p className="text-sm text-gray-400">Critical</p>
                </div>
                <div className="text-center p-4 bg-orange-500/10 rounded-lg">
                  <p className="text-2xl font-bold text-orange-500">{stats.highVulns}</p>
                  <p className="text-sm text-gray-400">High</p>
                </div>
                <div className="text-center p-4 bg-green-500/10 rounded-lg">
                  <p className="text-2xl font-bold text-green-500">{stats.healthyPackages}</p>
                  <p className="text-sm text-gray-400">Healthy</p>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* License Distribution */}
          <Card>
            <CardContent>
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-2">
                  <FileText className="h-5 w-5 text-purple-500" />
                  <CardTitle>License Distribution</CardTitle>
                </div>
                <Link
                  href="/dependencies"
                  className="flex items-center gap-1 text-sm text-blue-400 hover:text-blue-300"
                >
                  View all <ChevronRight className="h-4 w-4" />
                </Link>
              </div>

              {topLicenses.length > 0 ? (
                <div className="space-y-3">
                  {topLicenses.map((license) => {
                    const percentage = stats.totalPackages > 0
                      ? Math.round((license.count / stats.totalPackages) * 100)
                      : 0;
                    return (
                      <div key={license.name} className="space-y-1">
                        <div className="flex items-center justify-between text-sm">
                          <span className="text-white">{license.name || 'Unknown'}</span>
                          <span className="text-gray-400">{license.count} ({percentage}%)</span>
                        </div>
                        <div className="h-2 bg-gray-700 rounded-full overflow-hidden">
                          <div
                            className="h-full bg-purple-500 rounded-full"
                            style={{ width: `${percentage}%` }}
                          />
                        </div>
                      </div>
                    );
                  })}
                </div>
              ) : (
                <p className="text-gray-500 text-center py-4">No license data available</p>
              )}
            </CardContent>
          </Card>

          {/* Quick Links */}
          <div className="grid grid-cols-2 gap-4">
            <Link href="/dependencies">
              <Card className="hover:bg-gray-800/50 transition-colors cursor-pointer">
                <CardContent className="flex items-center gap-4">
                  <Package className="h-8 w-8 text-blue-500" />
                  <div>
                    <p className="font-medium text-white">Dependencies</p>
                    <p className="text-sm text-gray-400">View full SBOM and dependency tree</p>
                  </div>
                  <ChevronRight className="h-5 w-5 text-gray-500 ml-auto" />
                </CardContent>
              </Card>
            </Link>
            <Link href="/vulnerabilities">
              <Card className="hover:bg-gray-800/50 transition-colors cursor-pointer">
                <CardContent className="flex items-center gap-4">
                  <AlertTriangle className="h-8 w-8 text-red-500" />
                  <div>
                    <p className="font-medium text-white">Vulnerabilities</p>
                    <p className="text-sm text-gray-400">View CVEs and remediation</p>
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

export default function SupplyChainPage() {
  return (
    <MainLayout>
      <Suspense fallback={<div className="animate-pulse h-96 bg-gray-800 rounded-lg" />}>
        <SupplyChainContent />
      </Suspense>
    </MainLayout>
  );
}
