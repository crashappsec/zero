'use client';

import { useState, useMemo, Suspense, useEffect } from 'react';
import { MainLayout } from '@/components/layout/Sidebar';
import { Card, CardTitle, CardContent } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { useFetch } from '@/hooks/useApi';
import { api } from '@/lib/api';
import {
  Cpu,
  Code,
  Database,
  Server,
  Layers,
  Brain,
  Package,
  ChevronDown,
  ChevronRight,
} from 'lucide-react';
import { ProjectFilter } from '@/components/ui/ProjectFilter';

interface ProjectTech {
  projectId: string;
  languages: Array<{ name: string; percentage: number; files: number }>;
  frameworks: Array<{ name: string; version?: string; type: string }>;
  databases: string[];
  infrastructure: string[];
  ai_ml: {
    models: Array<{ name: string; format: string; path: string }>;
    frameworks: string[];
  };
}

function LanguageBar({ languages }: { languages: Array<{ name: string; percentage: number }> }) {
  const colors = [
    'bg-blue-500', 'bg-green-500', 'bg-yellow-500', 'bg-purple-500',
    'bg-red-500', 'bg-cyan-500', 'bg-orange-500', 'bg-pink-500',
  ];

  return (
    <div className="flex h-3 rounded-full overflow-hidden">
      {languages.slice(0, 8).map((lang, i) => (
        <div
          key={lang.name}
          className={`${colors[i % colors.length]}`}
          style={{ width: `${lang.percentage}%` }}
          title={`${lang.name}: ${lang.percentage.toFixed(1)}%`}
        />
      ))}
    </div>
  );
}

function ProjectTechCard({ data, expanded, onToggle }: {
  data: ProjectTech;
  expanded: boolean;
  onToggle: () => void;
}) {
  const topLangs = data.languages.slice(0, 3).map(l => l.name).join(', ');

  return (
    <Card className="overflow-hidden">
      <button
        onClick={onToggle}
        className="w-full flex items-center justify-between p-4 hover:bg-gray-800/50 transition-colors"
      >
        <div className="flex items-center gap-4 flex-1">
          {expanded ? (
            <ChevronDown className="h-4 w-4 text-gray-500" />
          ) : (
            <ChevronRight className="h-4 w-4 text-gray-500" />
          )}
          <div className="flex-1 min-w-0">
            <h3 className="font-medium text-white text-left">{data.projectId}</h3>
            <div className="flex items-center gap-4 mt-1 text-sm text-gray-400">
              <span>{data.languages.length} languages</span>
              <span>{data.frameworks.length} frameworks</span>
              {data.ai_ml.models.length > 0 && (
                <Badge variant="info" className="text-xs">AI/ML</Badge>
              )}
            </div>
          </div>
          {data.languages.length > 0 && (
            <div className="w-32 hidden md:block">
              <LanguageBar languages={data.languages} />
            </div>
          )}
        </div>
      </button>

      {expanded && (
        <div className="border-t border-gray-700 p-4 space-y-4">
          {/* Languages */}
          {data.languages.length > 0 && (
            <div>
              <h4 className="text-sm font-medium text-gray-400 mb-2">Languages</h4>
              <div className="flex flex-wrap gap-2">
                {data.languages.map((lang) => (
                  <Badge key={lang.name} variant="default">
                    {lang.name} ({lang.percentage.toFixed(1)}%)
                  </Badge>
                ))}
              </div>
            </div>
          )}

          {/* Frameworks */}
          {data.frameworks.length > 0 && (
            <div>
              <h4 className="text-sm font-medium text-gray-400 mb-2">Frameworks</h4>
              <div className="flex flex-wrap gap-2">
                {data.frameworks.map((fw) => (
                  <Badge key={fw.name} variant="info">
                    {fw.name}{fw.version ? ` v${fw.version}` : ''}
                  </Badge>
                ))}
              </div>
            </div>
          )}

          {/* AI/ML */}
          {(data.ai_ml.models.length > 0 || data.ai_ml.frameworks.length > 0) && (
            <div>
              <h4 className="text-sm font-medium text-gray-400 mb-2">AI/ML</h4>
              <div className="flex flex-wrap gap-2">
                {data.ai_ml.frameworks.map((fw) => (
                  <Badge key={fw} variant="warning">{fw}</Badge>
                ))}
                {data.ai_ml.models.map((m) => (
                  <Badge key={m.path} variant="success">{m.name} ({m.format})</Badge>
                ))}
              </div>
            </div>
          )}

          {/* Infrastructure */}
          {(data.databases.length > 0 || data.infrastructure.length > 0) && (
            <div>
              <h4 className="text-sm font-medium text-gray-400 mb-2">Infrastructure</h4>
              <div className="flex flex-wrap gap-2">
                {data.databases.map((db) => (
                  <Badge key={db} variant="success">{db}</Badge>
                ))}
                {data.infrastructure.map((inf) => (
                  <Badge key={inf} variant="default">{inf}</Badge>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </Card>
  );
}

function TechnologyContent() {
  const [expandedProjects, setExpandedProjects] = useState<Set<string>>(new Set());
  const [techData, setTechData] = useState<ProjectTech[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedProjects, setSelectedProjects] = useState<string[]>([]);

  const { data: projectsData } = useFetch(() => api.projects.list(), []);
  const projects = projectsData?.data || [];

  useEffect(() => {
    async function loadTechData() {
      if (projects.length === 0) return;

      setLoading(true);
      const results: ProjectTech[] = [];

      for (const project of projects) {
        try {
          const data = await api.analysis.raw(project.id, 'tech-id') as any;
          if (data?.findings) {
            const findings = data.findings;
            results.push({
              projectId: project.id,
              languages: findings.detection?.languages || [],
              frameworks: findings.frameworks?.detected || [],
              databases: findings.infrastructure?.databases || [],
              infrastructure: findings.infrastructure?.platforms || [],
              ai_ml: {
                models: findings.models?.detected || [],
                frameworks: findings.ai_security?.frameworks || [],
              },
            });
          }
        } catch {
          // Skip projects without tech data
        }
      }

      setTechData(results);
      setLoading(false);
    }

    loadTechData();
  }, [projects]);

  const toggleProject = (id: string) => {
    const newSet = new Set(expandedProjects);
    if (newSet.has(id)) {
      newSet.delete(id);
    } else {
      newSet.add(id);
    }
    setExpandedProjects(newSet);
  };

  // Filter data based on selected projects
  const filteredData = useMemo(() => {
    if (selectedProjects.length === 0) return techData;
    return techData.filter(d => selectedProjects.includes(d.projectId));
  }, [techData, selectedProjects]);

  // Aggregate stats
  const stats = useMemo(() => {
    if (filteredData.length === 0) return null;

    const allLanguages = new Map<string, number>();
    const allFrameworks = new Set<string>();
    const allDatabases = new Set<string>();
    let aiMlProjects = 0;

    filteredData.forEach((p) => {
      p.languages.forEach((l) => {
        allLanguages.set(l.name, (allLanguages.get(l.name) || 0) + l.files);
      });
      p.frameworks.forEach((f) => allFrameworks.add(f.name));
      p.databases.forEach((d) => allDatabases.add(d));
      if (p.ai_ml.models.length > 0 || p.ai_ml.frameworks.length > 0) {
        aiMlProjects++;
      }
    });

    const topLanguages = Array.from(allLanguages.entries())
      .sort((a, b) => b[1] - a[1])
      .slice(0, 10)
      .map(([name, files]) => ({ name, files }));

    return {
      totalLanguages: allLanguages.size,
      totalFrameworks: allFrameworks.size,
      totalDatabases: allDatabases.size,
      aiMlProjects,
      topLanguages,
      projectsAnalyzed: filteredData.length,
    };
  }, [filteredData]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            <Cpu className="h-6 w-6 text-cyan-500" />
            Technology Stack
          </h1>
          <p className="mt-1 text-gray-400">
            Languages, frameworks, and infrastructure across all projects
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
      ) : stats ? (
        <>
          {/* Aggregate Stats */}
          <div className="grid gap-4 md:grid-cols-4">
            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-blue-600/20">
                  <Code className="h-6 w-6 text-blue-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Languages</p>
                  <p className="text-2xl font-bold text-white">{stats.totalLanguages}</p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-purple-600/20">
                  <Layers className="h-6 w-6 text-purple-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Frameworks</p>
                  <p className="text-2xl font-bold text-white">{stats.totalFrameworks}</p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-green-600/20">
                  <Database className="h-6 w-6 text-green-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">Databases</p>
                  <p className="text-2xl font-bold text-white">{stats.totalDatabases}</p>
                </div>
              </div>
            </Card>

            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-pink-600/20">
                  <Brain className="h-6 w-6 text-pink-500" />
                </div>
                <div>
                  <p className="text-sm text-gray-400">AI/ML Projects</p>
                  <p className="text-2xl font-bold text-white">{stats.aiMlProjects}</p>
                </div>
              </div>
            </Card>
          </div>

          {/* Top Languages */}
          {stats.topLanguages.length > 0 && (
            <Card>
              <CardTitle className="flex items-center gap-2">
                <Code className="h-5 w-5 text-blue-500" />
                Top Languages (by file count)
              </CardTitle>
              <CardContent className="mt-4">
                <div className="flex flex-wrap gap-2">
                  {stats.topLanguages.map((lang) => (
                    <Badge key={lang.name} variant="default" className="text-sm">
                      {lang.name} ({lang.files} files)
                    </Badge>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

          {/* Projects List */}
          <div>
            <h2 className="text-lg font-semibold text-white mb-4">
              Projects ({stats.projectsAnalyzed} analyzed)
            </h2>
            <div className="space-y-3">
              {filteredData.map((data) => (
                <ProjectTechCard
                  key={data.projectId}
                  data={data}
                  expanded={expandedProjects.has(data.projectId)}
                  onToggle={() => toggleProject(data.projectId)}
                />
              ))}
            </div>
          </div>
        </>
      ) : (
        <Card className="text-center py-12">
          <Cpu className="h-12 w-12 text-gray-600 mx-auto mb-4" />
          <p className="text-gray-400">No technology data available</p>
          <p className="text-sm text-gray-500 mt-1">Run scans with the tech-id scanner</p>
        </Card>
      )}
    </div>
  );
}

export default function TechnologyPage() {
  return (
    <MainLayout>
      <Suspense fallback={<div className="animate-pulse h-96 bg-gray-800 rounded-lg" />}>
        <TechnologyContent />
      </Suspense>
    </MainLayout>
  );
}
