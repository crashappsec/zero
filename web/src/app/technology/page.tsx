'use client';

import { useState, useMemo, Suspense } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { MainLayout } from '@/components/layout/Sidebar';
import { Card, CardTitle, CardContent } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { useFetch } from '@/hooks/useApi';
import { api } from '@/lib/api';
import {
  Cpu,
  Code,
  Database,
  Globe,
  Server,
  Layers,
  Brain,
  FileCode,
  Package,
} from 'lucide-react';

interface Technology {
  name: string;
  version?: string;
  category: string;
  confidence: number;
  files?: string[];
}

interface Framework {
  name: string;
  version?: string;
  type: string;
}

interface TechData {
  languages: Array<{ name: string; percentage: number; files: number }>;
  frameworks: Framework[];
  databases: string[];
  infrastructure: string[];
  ai_ml: {
    models: Array<{ name: string; format: string; path: string }>;
    frameworks: string[];
    datasets: string[];
  };
  technologies: Technology[];
}

const categoryIcons: Record<string, typeof Cpu> = {
  language: Code,
  framework: Layers,
  database: Database,
  infrastructure: Server,
  ai_ml: Brain,
  web: Globe,
};

function TechnologyCard({ tech }: { tech: Technology }) {
  const Icon = categoryIcons[tech.category] || Cpu;

  return (
    <div className="flex items-center gap-3 p-3 bg-gray-800/50 rounded-lg">
      <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-gray-700">
        <Icon className="h-5 w-5 text-blue-400" />
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="font-medium text-white">{tech.name}</span>
          {tech.version && (
            <Badge variant="default" className="text-xs">{tech.version}</Badge>
          )}
        </div>
        <p className="text-xs text-gray-500 capitalize">{tech.category}</p>
      </div>
      <div className="text-right">
        <div className={`text-sm font-medium ${
          tech.confidence >= 0.9 ? 'text-green-400' :
          tech.confidence >= 0.7 ? 'text-yellow-400' : 'text-gray-400'
        }`}>
          {Math.round(tech.confidence * 100)}%
        </div>
        <p className="text-xs text-gray-500">confidence</p>
      </div>
    </div>
  );
}

function LanguageBar({ languages }: { languages: Array<{ name: string; percentage: number; files: number }> }) {
  const colors = [
    'bg-blue-500', 'bg-green-500', 'bg-yellow-500', 'bg-purple-500',
    'bg-red-500', 'bg-cyan-500', 'bg-orange-500', 'bg-pink-500',
  ];

  return (
    <div>
      <div className="flex h-4 rounded-full overflow-hidden mb-3">
        {languages.map((lang, i) => (
          <div
            key={lang.name}
            className={`${colors[i % colors.length]}`}
            style={{ width: `${lang.percentage}%` }}
            title={`${lang.name}: ${lang.percentage.toFixed(1)}%`}
          />
        ))}
      </div>
      <div className="flex flex-wrap gap-4">
        {languages.map((lang, i) => (
          <div key={lang.name} className="flex items-center gap-2">
            <div className={`w-3 h-3 rounded-full ${colors[i % colors.length]}`} />
            <span className="text-sm text-white">{lang.name}</span>
            <span className="text-sm text-gray-500">{lang.percentage.toFixed(1)}%</span>
            <span className="text-xs text-gray-600">({lang.files} files)</span>
          </div>
        ))}
      </div>
    </div>
  );
}

function TechnologyContent() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const projectId = searchParams.get('project');

  const { data: projectsData } = useFetch(() => api.projects.list(), []);
  const projects = projectsData?.data || [];

  const { data: techData, loading, error } = useFetch(
    () => projectId ? api.analysis.raw(projectId, 'tech-id') as Promise<any> : Promise.resolve(null),
    [projectId]
  );

  const technology = useMemo(() => {
    if (!techData?.findings) return null;
    const findings = techData.findings;
    return {
      languages: findings.detection?.languages || [],
      frameworks: findings.frameworks?.detected || [],
      databases: findings.infrastructure?.databases || [],
      infrastructure: findings.infrastructure?.platforms || [],
      ai_ml: {
        models: findings.models?.detected || [],
        frameworks: findings.ai_security?.frameworks || [],
        datasets: findings.datasets?.detected || [],
      },
      technologies: findings.detection?.technologies || [],
    } as TechData;
  }, [techData]);

  const handleProjectChange = (newProjectId: string) => {
    router.push(`/technology?project=${encodeURIComponent(newProjectId)}`);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-white flex items-center gap-2">
          <Cpu className="h-6 w-6 text-cyan-500" />
          Technology Stack
        </h1>
        <p className="mt-1 text-gray-400">
          Languages, frameworks, and infrastructure detected in your repositories
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

      {projectId && technology && (
        <>
          {/* Languages */}
          {technology.languages.length > 0 && (
            <Card>
              <CardTitle className="flex items-center gap-2">
                <Code className="h-5 w-5 text-blue-500" />
                Languages
              </CardTitle>
              <CardContent className="mt-4">
                <LanguageBar languages={technology.languages} />
              </CardContent>
            </Card>
          )}

          {/* Frameworks */}
          {technology.frameworks.length > 0 && (
            <Card>
              <CardTitle className="flex items-center gap-2">
                <Layers className="h-5 w-5 text-purple-500" />
                Frameworks & Libraries
              </CardTitle>
              <CardContent className="mt-4">
                <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
                  {technology.frameworks.map((fw) => (
                    <div key={fw.name} className="flex items-center gap-3 p-3 bg-gray-800/50 rounded-lg">
                      <Package className="h-5 w-5 text-purple-400" />
                      <div>
                        <span className="font-medium text-white">{fw.name}</span>
                        {fw.version && (
                          <span className="ml-2 text-sm text-gray-500">v{fw.version}</span>
                        )}
                        <p className="text-xs text-gray-500">{fw.type}</p>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

          {/* AI/ML */}
          {(technology.ai_ml.models.length > 0 || technology.ai_ml.frameworks.length > 0) && (
            <Card>
              <CardTitle className="flex items-center gap-2">
                <Brain className="h-5 w-5 text-pink-500" />
                AI/ML Components
              </CardTitle>
              <CardContent className="mt-4 space-y-4">
                {technology.ai_ml.frameworks.length > 0 && (
                  <div>
                    <h4 className="text-sm font-medium text-gray-400 mb-2">Frameworks</h4>
                    <div className="flex flex-wrap gap-2">
                      {technology.ai_ml.frameworks.map((fw) => (
                        <Badge key={fw} variant="info">{fw}</Badge>
                      ))}
                    </div>
                  </div>
                )}
                {technology.ai_ml.models.length > 0 && (
                  <div>
                    <h4 className="text-sm font-medium text-gray-400 mb-2">Models Detected</h4>
                    <div className="space-y-2">
                      {technology.ai_ml.models.map((model) => (
                        <div key={model.path} className="flex items-center gap-3 p-2 bg-gray-800/50 rounded">
                          <Brain className="h-4 w-4 text-pink-400" />
                          <span className="text-sm text-white">{model.name}</span>
                          <Badge variant="default" className="text-xs">{model.format}</Badge>
                          <span className="text-xs text-gray-500 truncate">{model.path}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          )}

          {/* Infrastructure */}
          {(technology.databases.length > 0 || technology.infrastructure.length > 0) && (
            <Card>
              <CardTitle className="flex items-center gap-2">
                <Server className="h-5 w-5 text-green-500" />
                Infrastructure
              </CardTitle>
              <CardContent className="mt-4 space-y-4">
                {technology.databases.length > 0 && (
                  <div>
                    <h4 className="text-sm font-medium text-gray-400 mb-2">Databases</h4>
                    <div className="flex flex-wrap gap-2">
                      {technology.databases.map((db) => (
                        <Badge key={db} variant="success">{db}</Badge>
                      ))}
                    </div>
                  </div>
                )}
                {technology.infrastructure.length > 0 && (
                  <div>
                    <h4 className="text-sm font-medium text-gray-400 mb-2">Platforms</h4>
                    <div className="flex flex-wrap gap-2">
                      {technology.infrastructure.map((platform) => (
                        <Badge key={platform} variant="info">{platform}</Badge>
                      ))}
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          )}

          {/* All Technologies */}
          {technology.technologies.length > 0 && (
            <Card>
              <CardTitle className="flex items-center gap-2">
                <Cpu className="h-5 w-5 text-cyan-500" />
                All Detected Technologies
              </CardTitle>
              <CardContent className="mt-4">
                <div className="grid gap-3 md:grid-cols-2">
                  {technology.technologies.map((tech) => (
                    <TechnologyCard key={`${tech.name}-${tech.category}`} tech={tech} />
                  ))}
                </div>
              </CardContent>
            </Card>
          )}
        </>
      )}

      {projectId && loading && (
        <Card className="p-8 text-center text-gray-400">Loading technology data...</Card>
      )}

      {projectId && error && (
        <Card className="p-8 text-center text-red-400">
          No technology data available. Run a scan with the tech-id scanner.
        </Card>
      )}

      {!projectId && (
        <Card className="text-center py-12">
          <Cpu className="h-12 w-12 text-gray-600 mx-auto mb-4" />
          <p className="text-gray-400">Select a project to view technology stack analysis</p>
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
