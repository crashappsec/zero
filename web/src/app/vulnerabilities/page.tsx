'use client';

import { useState, useMemo, Suspense, useEffect } from 'react';
import { MainLayout } from '@/components/layout/Sidebar';
import { Card, CardTitle, CardContent } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Badge, SeverityBadge } from '@/components/ui/Badge';
import { useFetch } from '@/hooks/useApi';
import { api } from '@/lib/api';
import type { Vulnerability, Project } from '@/lib/types';
import { ExportButton } from '@/components/ui/ExportButton';
import { downloadCSV, downloadJSON } from '@/lib/export';
import { ProjectFilter } from '@/components/ui/ProjectFilter';
import {
  Search,
  Filter,
  ChevronDown,
  ChevronUp,
  ExternalLink,
  Package,
  AlertTriangle,
  Shield,
  ArrowUpDown,
  X,
} from 'lucide-react';

type SortField = 'severity' | 'package' | 'id' | 'project';
type SortDirection = 'asc' | 'desc';

const severityOrder: Record<string, number> = {
  critical: 0,
  high: 1,
  medium: 2,
  low: 3,
  unknown: 4,
};

interface VulnWithProject extends Vulnerability {
  projectId: string;
}

function VulnerabilityRow({ vuln, expanded, onToggle }: { vuln: VulnWithProject; expanded: boolean; onToggle: () => void }) {
  return (
    <div className="border-b border-gray-700 last:border-0">
      <button
        onClick={onToggle}
        className="w-full flex items-center gap-4 px-4 py-3 text-left hover:bg-gray-800/50 transition-colors"
      >
        <SeverityBadge severity={vuln.severity} />
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-mono text-sm text-blue-400">{vuln.id}</span>
            <span className="text-gray-500">in</span>
            <span className="font-medium text-white truncate">{vuln.package}@{vuln.version}</span>
          </div>
          <div className="flex items-center gap-2 mt-0.5">
            <p className="text-sm text-gray-400 truncate">{vuln.title}</p>
            <Badge variant="default" className="text-xs">{vuln.projectId}</Badge>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {vuln.fix_version && (
            <Badge variant="success" className="text-xs">
              Fix: {vuln.fix_version}
            </Badge>
          )}
          {expanded ? (
            <ChevronUp className="h-4 w-4 text-gray-500" />
          ) : (
            <ChevronDown className="h-4 w-4 text-gray-500" />
          )}
        </div>
      </button>

      {expanded && (
        <div className="px-4 py-3 bg-gray-800/30 border-t border-gray-700/50">
          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <h4 className="text-xs font-medium text-gray-500 uppercase mb-1">Description</h4>
              <p className="text-sm text-gray-300">{vuln.description || 'No description available'}</p>
            </div>
            <div className="space-y-3">
              <div>
                <h4 className="text-xs font-medium text-gray-500 uppercase mb-1">Package Details</h4>
                <div className="flex items-center gap-2">
                  <Package className="h-4 w-4 text-gray-500" />
                  <span className="text-sm text-white">{vuln.package}</span>
                  <span className="text-sm text-gray-500">version {vuln.version}</span>
                </div>
              </div>
              {vuln.fix_version && (
                <div>
                  <h4 className="text-xs font-medium text-gray-500 uppercase mb-1">Remediation</h4>
                  <p className="text-sm text-green-400">Upgrade to version {vuln.fix_version} or later</p>
                </div>
              )}
              <div>
                <h4 className="text-xs font-medium text-gray-500 uppercase mb-1">Source</h4>
                <span className="text-sm text-gray-400">{vuln.source}</span>
              </div>
            </div>
          </div>
          <div className="mt-4 flex gap-2">
            <a
              href={`https://osv.dev/vulnerability/${vuln.id}`}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1 text-sm text-blue-400 hover:text-blue-300"
            >
              View in OSV <ExternalLink className="h-3 w-3" />
            </a>
            {vuln.id.startsWith('CVE-') && (
              <a
                href={`https://nvd.nist.gov/vuln/detail/${vuln.id}`}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1 text-sm text-blue-400 hover:text-blue-300"
              >
                View in NVD <ExternalLink className="h-3 w-3" />
              </a>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

function SeverityFilter({
  selected,
  onChange
}: {
  selected: string[];
  onChange: (severities: string[]) => void;
}) {
  const severities = ['critical', 'high', 'medium', 'low'];

  const toggle = (sev: string) => {
    if (selected.includes(sev)) {
      onChange(selected.filter(s => s !== sev));
    } else {
      onChange([...selected, sev]);
    }
  };

  return (
    <div className="flex items-center gap-2">
      <Filter className="h-4 w-4 text-gray-500" />
      <span className="text-sm text-gray-400">Severity:</span>
      <div className="flex gap-1">
        {severities.map(sev => (
          <button
            key={sev}
            onClick={() => toggle(sev)}
            className={`px-2 py-1 text-xs rounded-md transition-colors ${
              selected.includes(sev)
                ? sev === 'critical' ? 'bg-red-600 text-white'
                : sev === 'high' ? 'bg-orange-600 text-white'
                : sev === 'medium' ? 'bg-yellow-600 text-white'
                : 'bg-blue-600 text-white'
                : 'bg-gray-700 text-gray-400 hover:bg-gray-600'
            }`}
          >
            {sev}
          </button>
        ))}
      </div>
      {selected.length > 0 && (
        <button
          onClick={() => onChange([])}
          className="text-xs text-gray-500 hover:text-gray-400"
        >
          Clear
        </button>
      )}
    </div>
  );
}

function VulnerabilitiesContent() {
  const [search, setSearch] = useState('');
  const [severityFilter, setSeverityFilter] = useState<string[]>([]);
  const [sortField, setSortField] = useState<SortField>('severity');
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc');
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());
  const [selectedProjects, setSelectedProjects] = useState<string[]>([]);
  const [allVulns, setAllVulns] = useState<VulnWithProject[]>([]);
  const [loading, setLoading] = useState(true);

  // Fetch projects
  const { data: projectsData } = useFetch(() => api.projects.list(), []);
  const projects = projectsData?.data || [];

  // Load vulnerabilities for all projects in parallel
  useEffect(() => {
    async function loadAllVulns() {
      if (projects.length === 0) return;

      setLoading(true);

      // Fetch all projects in parallel for better performance
      const results = await Promise.all(
        projects.map(async (project) => {
          try {
            const data = await api.analysis.vulnerabilities(project.id);
            if (data?.data) {
              return data.data.map(vuln => ({ ...vuln, projectId: project.id }));
            }
          } catch {
            // Skip projects without vulnerability data
          }
          return [];
        })
      );

      setAllVulns(results.flat());
      setLoading(false);
    }

    loadAllVulns();
  }, [projects]);

  // Filter by selected projects
  const projectFilteredVulns = useMemo(() => {
    if (selectedProjects.length === 0) return allVulns;
    return allVulns.filter(v => selectedProjects.includes(v.projectId));
  }, [allVulns, selectedProjects]);

  // Filter and sort
  const filteredVulns = useMemo(() => {
    let result = [...projectFilteredVulns];

    // Search filter
    if (search) {
      const lower = search.toLowerCase();
      result = result.filter(v =>
        v.id.toLowerCase().includes(lower) ||
        v.package.toLowerCase().includes(lower) ||
        v.title.toLowerCase().includes(lower) ||
        v.projectId.toLowerCase().includes(lower) ||
        (v.description?.toLowerCase().includes(lower))
      );
    }

    // Severity filter
    if (severityFilter.length > 0) {
      result = result.filter(v => severityFilter.includes(v.severity.toLowerCase()));
    }

    // Sort
    result.sort((a, b) => {
      let cmp = 0;
      switch (sortField) {
        case 'severity':
          cmp = (severityOrder[a.severity.toLowerCase()] || 99) - (severityOrder[b.severity.toLowerCase()] || 99);
          break;
        case 'package':
          cmp = a.package.localeCompare(b.package);
          break;
        case 'id':
          cmp = a.id.localeCompare(b.id);
          break;
        case 'project':
          cmp = a.projectId.localeCompare(b.projectId);
          break;
      }
      return sortDirection === 'asc' ? cmp : -cmp;
    });

    return result;
  }, [projectFilteredVulns, search, severityFilter, sortField, sortDirection]);

  // Stats
  const stats = useMemo(() => {
    const counts = { critical: 0, high: 0, medium: 0, low: 0 };
    projectFilteredVulns.forEach(v => {
      const sev = v.severity.toLowerCase() as keyof typeof counts;
      if (sev in counts) counts[sev]++;
    });
    return counts;
  }, [projectFilteredVulns]);

  const toggleExpanded = (id: string) => {
    const newSet = new Set(expandedIds);
    if (newSet.has(id)) {
      newSet.delete(id);
    } else {
      newSet.add(id);
    }
    setExpandedIds(newSet);
  };

  const toggleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(d => d === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('asc');
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            <Shield className="h-6 w-6 text-red-500" />
            Vulnerabilities
          </h1>
          <p className="mt-1 text-gray-400">
            Package and code vulnerabilities across all projects
          </p>
        </div>
        <div className="flex items-center gap-2">
          {projects.length > 0 && (
            <ProjectFilter
              projects={projects}
              selectedProjects={selectedProjects}
              onChange={setSelectedProjects}
            />
          )}
          {projectFilteredVulns.length > 0 && (
            <ExportButton
              onExport={(format) => {
                const data = filteredVulns.map(v => ({
                  project: v.projectId,
                  id: v.id,
                  package: v.package,
                  version: v.version,
                  severity: v.severity,
                  title: v.title,
                  description: v.description || '',
                  fix_version: v.fix_version || '',
                  source: v.source,
                }));
                if (format === 'csv') {
                  downloadCSV(data, 'vulnerabilities');
                } else {
                  downloadJSON(data, 'vulnerabilities');
                }
              }}
            />
          )}
        </div>
      </div>

      {loading ? (
        <div className="space-y-4">
          {[1, 2, 3].map((i) => (
            <Card key={i} className="animate-pulse h-24" />
          ))}
        </div>
      ) : allVulns.length > 0 ? (
        <>
          {/* Stats */}
          <div className="grid grid-cols-4 gap-4">
            <Card className="text-center">
              <p className="text-2xl font-bold text-red-500">{stats.critical}</p>
              <p className="text-sm text-gray-400">Critical</p>
            </Card>
            <Card className="text-center">
              <p className="text-2xl font-bold text-orange-500">{stats.high}</p>
              <p className="text-sm text-gray-400">High</p>
            </Card>
            <Card className="text-center">
              <p className="text-2xl font-bold text-yellow-500">{stats.medium}</p>
              <p className="text-sm text-gray-400">Medium</p>
            </Card>
            <Card className="text-center">
              <p className="text-2xl font-bold text-blue-500">{stats.low}</p>
              <p className="text-sm text-gray-400">Low</p>
            </Card>
          </div>

          {/* Filters */}
          <Card>
            <CardContent>
              <div className="flex flex-col md:flex-row gap-4">
                <div className="flex-1">
                  <Input
                    placeholder="Search vulnerabilities..."
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    icon={<Search className="h-4 w-4" />}
                  />
                </div>
                <SeverityFilter
                  selected={severityFilter}
                  onChange={setSeverityFilter}
                />
              </div>
            </CardContent>
          </Card>

          {/* Results */}
          <Card>
            <div className="flex items-center justify-between px-4 py-3 border-b border-gray-700">
              <div className="flex items-center gap-4">
                <span className="text-sm text-gray-400">
                  {filteredVulns.length} of {projectFilteredVulns.length} vulnerabilities
                </span>
                {(search || severityFilter.length > 0) && (
                  <button
                    onClick={() => {
                      setSearch('');
                      setSeverityFilter([]);
                    }}
                    className="flex items-center gap-1 text-xs text-gray-500 hover:text-gray-400"
                  >
                    <X className="h-3 w-3" />
                    Clear filters
                  </button>
                )}
              </div>
              <div className="flex items-center gap-2">
                <span className="text-xs text-gray-500">Sort by:</span>
                {(['severity', 'package', 'id', 'project'] as SortField[]).map((field) => (
                  <button
                    key={field}
                    onClick={() => toggleSort(field)}
                    className={`flex items-center gap-1 px-2 py-1 text-xs rounded ${
                      sortField === field ? 'bg-gray-700 text-white' : 'text-gray-400 hover:text-white'
                    }`}
                  >
                    {field.charAt(0).toUpperCase() + field.slice(1)}
                    <ArrowUpDown className="h-3 w-3" />
                  </button>
                ))}
              </div>
            </div>

            {filteredVulns.length === 0 ? (
              <div className="p-8 text-center">
                <AlertTriangle className="h-12 w-12 text-gray-600 mx-auto mb-4" />
                <p className="text-gray-400">No vulnerabilities match your filters</p>
              </div>
            ) : (
              <div>
                {filteredVulns.map((vuln) => (
                  <VulnerabilityRow
                    key={`${vuln.projectId}-${vuln.id}-${vuln.package}-${vuln.version}`}
                    vuln={vuln}
                    expanded={expandedIds.has(`${vuln.id}-${vuln.package}`)}
                    onToggle={() => toggleExpanded(`${vuln.id}-${vuln.package}`)}
                  />
                ))}
              </div>
            )}
          </Card>
        </>
      ) : (
        <Card className="text-center py-12">
          <Shield className="h-12 w-12 text-gray-600 mx-auto mb-4" />
          <p className="text-gray-400">No vulnerabilities found</p>
          <p className="text-sm text-gray-500 mt-1">Run scans with the package-analysis or code-security scanner</p>
        </Card>
      )}
    </div>
  );
}

export default function VulnerabilitiesPage() {
  return (
    <MainLayout>
      <Suspense fallback={<div className="animate-pulse h-96 bg-gray-800 rounded-lg" />}>
        <VulnerabilitiesContent />
      </Suspense>
    </MainLayout>
  );
}
