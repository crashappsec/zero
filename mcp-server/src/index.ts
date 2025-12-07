#!/usr/bin/env node
/**
 * Zero MCP Server
 * Data layer for Zero repository analysis system
 *
 * Copyright (c) 2025 Crash Override Inc.
 * https://crashoverride.com
 * SPDX-License-Identifier: GPL-3.0
 */

import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
  ListResourcesRequestSchema,
  ReadResourceRequestSchema,
} from "@modelcontextprotocol/sdk/types.js";
import { z } from "zod";
import * as fs from "fs/promises";
import * as path from "path";
import { glob } from "glob";
import os from "os";

// Configuration
const PROJECTS_DIR = process.env.PHANTOM_HOME || path.join(os.homedir(), ".phantom", "projects");

// Types for analysis data
interface Project {
  id: string;
  owner: string;
  repo: string;
  path: string;
  analysisPath: string;
  repoPath: string;
  availableAnalyses: string[];
}

interface MalcontentFinding {
  path: string;
  riskLevel: string;
  riskScore: number;
  behaviors: Array<{
    description: string;
    matchStrings: string[];
    riskLevel: string;
    riskScore: number;
    ruleUrl: string;
    id: string;
    ruleName: string;
  }>;
}

interface VulnerabilityFinding {
  id: string;
  package: string;
  severity: string;
  title: string;
  fixedIn?: string;
  url?: string;
}

// Helper functions
async function getProjects(): Promise<Project[]> {
  const projects: Project[] = [];

  try {
    const orgs = await fs.readdir(PROJECTS_DIR);

    for (const org of orgs) {
      const orgPath = path.join(PROJECTS_DIR, org);
      const orgStat = await fs.stat(orgPath);
      if (!orgStat.isDirectory()) continue;

      const repos = await fs.readdir(orgPath);
      for (const repo of repos) {
        const repoPath = path.join(orgPath, repo);
        const repoStat = await fs.stat(repoPath);
        if (!repoStat.isDirectory()) continue;

        const analysisPath = path.join(repoPath, "analysis");
        const sourceRepoPath = path.join(repoPath, "repo");

        // Check what analyses are available
        const availableAnalyses: string[] = [];
        try {
          const analysisFiles = await fs.readdir(analysisPath);
          for (const file of analysisFiles) {
            if (file.endsWith(".json")) {
              availableAnalyses.push(file.replace(".json", ""));
            }
          }
        } catch {
          // No analysis directory yet
        }

        projects.push({
          id: `${org}/${repo}`,
          owner: org,
          repo,
          path: repoPath,
          analysisPath,
          repoPath: sourceRepoPath,
          availableAnalyses,
        });
      }
    }
  } catch (error) {
    // Projects directory doesn't exist yet
  }

  return projects;
}

async function readAnalysisFile(projectId: string, analysisType: string): Promise<unknown | null> {
  const [owner, repo] = projectId.split("/");
  const filePath = path.join(PROJECTS_DIR, owner, repo, "analysis", `${analysisType}.json`);

  try {
    const content = await fs.readFile(filePath, "utf-8");
    return JSON.parse(content);
  } catch {
    return null;
  }
}

async function getMalcontentFindings(projectId: string, minRiskLevel?: string): Promise<MalcontentFinding[]> {
  const data = await readAnalysisFile(projectId, "malcontent") as { Files?: Record<string, unknown> } | null;
  if (!data || !data.Files) return [];

  const riskLevels = ["LOW", "MEDIUM", "HIGH", "CRITICAL"];
  const minIndex = minRiskLevel ? riskLevels.indexOf(minRiskLevel.toUpperCase()) : 0;

  const findings: MalcontentFinding[] = [];

  for (const [filePath, fileData] of Object.entries(data.Files)) {
    const file = fileData as {
      Path: string;
      RiskLevel: string;
      RiskScore: number;
      Behaviors?: Array<{
        Description: string;
        MatchStrings: string[];
        RiskLevel: string;
        RiskScore: number;
        RuleURL: string;
        ID: string;
        RuleName: string;
      }>;
    };

    const fileRiskIndex = riskLevels.indexOf(file.RiskLevel);
    if (fileRiskIndex >= minIndex) {
      findings.push({
        path: file.Path,
        riskLevel: file.RiskLevel,
        riskScore: file.RiskScore,
        behaviors: (file.Behaviors || []).map(b => ({
          description: b.Description,
          matchStrings: b.MatchStrings,
          riskLevel: b.RiskLevel,
          riskScore: b.RiskScore,
          ruleUrl: b.RuleURL,
          id: b.ID,
          ruleName: b.RuleName,
        })),
      });
    }
  }

  // Sort by risk score descending
  findings.sort((a, b) => b.riskScore - a.riskScore);

  return findings;
}

async function getVulnerabilities(projectId: string, severity?: string): Promise<VulnerabilityFinding[]> {
  const data = await readAnalysisFile(projectId, "vulnerabilities") as { vulnerabilities?: Array<{
    id: string;
    package: string;
    severity: string;
    title: string;
    fixed_in?: string;
    url?: string;
  }> } | null;

  if (!data || !data.vulnerabilities) return [];

  let vulns = data.vulnerabilities.map(v => ({
    id: v.id,
    package: v.package,
    severity: v.severity,
    title: v.title,
    fixedIn: v.fixed_in,
    url: v.url,
  }));

  if (severity) {
    vulns = vulns.filter(v => v.severity.toUpperCase() === severity.toUpperCase());
  }

  return vulns;
}

async function getTechnologies(projectId: string): Promise<unknown> {
  return await readAnalysisFile(projectId, "technology") ||
         await readAnalysisFile(projectId, "tech-discovery");
}

async function getPackageHealth(projectId: string): Promise<unknown> {
  return await readAnalysisFile(projectId, "package-health");
}

async function getLicenses(projectId: string): Promise<unknown> {
  return await readAnalysisFile(projectId, "licenses");
}

async function getCodeSecurity(projectId: string): Promise<unknown> {
  return await readAnalysisFile(projectId, "code-security") ||
         await readAnalysisFile(projectId, "semgrep");
}

// Create the MCP server
const server = new Server(
  {
    name: "zero",
    version: "1.0.0",
  },
  {
    capabilities: {
      tools: {},
      resources: {},
    },
  }
);

// Tool definitions
const tools = [
  {
    name: "list_projects",
    description: "List all hydrated projects with their available analyses. Returns project IDs in owner/repo format.",
    inputSchema: {
      type: "object" as const,
      properties: {
        owner: {
          type: "string",
          description: "Optional: filter by organization/owner name",
        },
      },
    },
  },
  {
    name: "get_project_summary",
    description: "Get a summary of a project including available analyses and basic stats",
    inputSchema: {
      type: "object" as const,
      properties: {
        project: {
          type: "string",
          description: "Project ID in owner/repo format (e.g., 'expressjs/express')",
        },
      },
      required: ["project"],
    },
  },
  {
    name: "get_malcontent",
    description: "Get malcontent (malware/suspicious behavior) findings for a project. Use this when investigating supply chain risks or suspicious code patterns.",
    inputSchema: {
      type: "object" as const,
      properties: {
        project: {
          type: "string",
          description: "Project ID in owner/repo format",
        },
        min_risk: {
          type: "string",
          enum: ["LOW", "MEDIUM", "HIGH", "CRITICAL"],
          description: "Minimum risk level to include (default: all)",
        },
        limit: {
          type: "number",
          description: "Maximum number of findings to return (default: 20)",
        },
      },
      required: ["project"],
    },
  },
  {
    name: "get_vulnerabilities",
    description: "Get known vulnerabilities (CVEs) for a project's dependencies",
    inputSchema: {
      type: "object" as const,
      properties: {
        project: {
          type: "string",
          description: "Project ID in owner/repo format",
        },
        severity: {
          type: "string",
          enum: ["LOW", "MEDIUM", "HIGH", "CRITICAL"],
          description: "Filter by severity level",
        },
      },
      required: ["project"],
    },
  },
  {
    name: "get_technologies",
    description: "Get detected technologies, frameworks, and libraries used in a project",
    inputSchema: {
      type: "object" as const,
      properties: {
        project: {
          type: "string",
          description: "Project ID in owner/repo format",
        },
      },
      required: ["project"],
    },
  },
  {
    name: "get_package_health",
    description: "Get package health scores and dependency analysis for a project",
    inputSchema: {
      type: "object" as const,
      properties: {
        project: {
          type: "string",
          description: "Project ID in owner/repo format",
        },
      },
      required: ["project"],
    },
  },
  {
    name: "get_licenses",
    description: "Get license information for a project's dependencies",
    inputSchema: {
      type: "object" as const,
      properties: {
        project: {
          type: "string",
          description: "Project ID in owner/repo format",
        },
      },
      required: ["project"],
    },
  },
  {
    name: "get_code_security",
    description: "Get code security findings from static analysis (Semgrep)",
    inputSchema: {
      type: "object" as const,
      properties: {
        project: {
          type: "string",
          description: "Project ID in owner/repo format",
        },
      },
      required: ["project"],
    },
  },
  {
    name: "search_findings",
    description: "Search across all findings for a pattern or keyword",
    inputSchema: {
      type: "object" as const,
      properties: {
        query: {
          type: "string",
          description: "Search query (searches in paths, descriptions, package names)",
        },
        project: {
          type: "string",
          description: "Optional: limit search to a specific project",
        },
        type: {
          type: "string",
          enum: ["malcontent", "vulnerabilities", "code-security", "all"],
          description: "Type of findings to search (default: all)",
        },
      },
      required: ["query"],
    },
  },
  {
    name: "get_analysis_raw",
    description: "Get raw analysis JSON for any analysis type. Use this when you need the complete data structure.",
    inputSchema: {
      type: "object" as const,
      properties: {
        project: {
          type: "string",
          description: "Project ID in owner/repo format",
        },
        analysis_type: {
          type: "string",
          description: "Analysis type (e.g., 'malcontent', 'vulnerabilities', 'technology', 'package-health', 'licenses', 'code-security')",
        },
      },
      required: ["project", "analysis_type"],
    },
  },
];

// Handle tool listing
server.setRequestHandler(ListToolsRequestSchema, async () => {
  return { tools };
});

// Handle tool calls
server.setRequestHandler(CallToolRequestSchema, async (request) => {
  const { name, arguments: args } = request.params;

  try {
    switch (name) {
      case "list_projects": {
        const projects = await getProjects();
        const filtered = args?.owner
          ? projects.filter(p => p.owner === args.owner)
          : projects;

        return {
          content: [{
            type: "text",
            text: JSON.stringify(filtered.map(p => ({
              id: p.id,
              availableAnalyses: p.availableAnalyses,
            })), null, 2),
          }],
        };
      }

      case "get_project_summary": {
        const project = args?.project as string;
        const projects = await getProjects();
        const found = projects.find(p => p.id === project);

        if (!found) {
          return {
            content: [{
              type: "text",
              text: `Project '${project}' not found. Use list_projects to see available projects.`,
            }],
            isError: true,
          };
        }

        // Get quick stats from available analyses
        const stats: Record<string, unknown> = {
          project: found.id,
          availableAnalyses: found.availableAnalyses,
        };

        if (found.availableAnalyses.includes("malcontent")) {
          const findings = await getMalcontentFindings(project);
          stats.malcontent = {
            totalFiles: findings.length,
            byRisk: {
              critical: findings.filter(f => f.riskLevel === "CRITICAL").length,
              high: findings.filter(f => f.riskLevel === "HIGH").length,
              medium: findings.filter(f => f.riskLevel === "MEDIUM").length,
              low: findings.filter(f => f.riskLevel === "LOW").length,
            },
          };
        }

        if (found.availableAnalyses.includes("vulnerabilities")) {
          const vulns = await getVulnerabilities(project);
          stats.vulnerabilities = {
            total: vulns.length,
            bySeverity: {
              critical: vulns.filter(v => v.severity === "CRITICAL").length,
              high: vulns.filter(v => v.severity === "HIGH").length,
              medium: vulns.filter(v => v.severity === "MEDIUM").length,
              low: vulns.filter(v => v.severity === "LOW").length,
            },
          };
        }

        return {
          content: [{
            type: "text",
            text: JSON.stringify(stats, null, 2),
          }],
        };
      }

      case "get_malcontent": {
        const project = args?.project as string;
        const minRisk = args?.min_risk as string | undefined;
        const limit = (args?.limit as number) || 20;

        const findings = await getMalcontentFindings(project, minRisk);

        return {
          content: [{
            type: "text",
            text: JSON.stringify({
              project,
              totalFindings: findings.length,
              findings: findings.slice(0, limit),
            }, null, 2),
          }],
        };
      }

      case "get_vulnerabilities": {
        const project = args?.project as string;
        const severity = args?.severity as string | undefined;

        const vulns = await getVulnerabilities(project, severity);

        return {
          content: [{
            type: "text",
            text: JSON.stringify({
              project,
              total: vulns.length,
              vulnerabilities: vulns,
            }, null, 2),
          }],
        };
      }

      case "get_technologies": {
        const project = args?.project as string;
        const tech = await getTechnologies(project);

        return {
          content: [{
            type: "text",
            text: JSON.stringify({
              project,
              technologies: tech,
            }, null, 2),
          }],
        };
      }

      case "get_package_health": {
        const project = args?.project as string;
        const health = await getPackageHealth(project);

        return {
          content: [{
            type: "text",
            text: JSON.stringify({
              project,
              health,
            }, null, 2),
          }],
        };
      }

      case "get_licenses": {
        const project = args?.project as string;
        const licenses = await getLicenses(project);

        return {
          content: [{
            type: "text",
            text: JSON.stringify({
              project,
              licenses,
            }, null, 2),
          }],
        };
      }

      case "get_code_security": {
        const project = args?.project as string;
        const security = await getCodeSecurity(project);

        return {
          content: [{
            type: "text",
            text: JSON.stringify({
              project,
              findings: security,
            }, null, 2),
          }],
        };
      }

      case "search_findings": {
        const query = (args?.query as string).toLowerCase();
        const projectFilter = args?.project as string | undefined;
        const typeFilter = (args?.type as string) || "all";

        const projects = await getProjects();
        const filtered = projectFilter
          ? projects.filter(p => p.id === projectFilter)
          : projects;

        const results: Array<{
          project: string;
          type: string;
          finding: unknown;
        }> = [];

        for (const project of filtered) {
          // Search malcontent
          if (typeFilter === "all" || typeFilter === "malcontent") {
            const findings = await getMalcontentFindings(project.id);
            for (const finding of findings) {
              const searchText = JSON.stringify(finding).toLowerCase();
              if (searchText.includes(query)) {
                results.push({
                  project: project.id,
                  type: "malcontent",
                  finding,
                });
              }
            }
          }

          // Search vulnerabilities
          if (typeFilter === "all" || typeFilter === "vulnerabilities") {
            const vulns = await getVulnerabilities(project.id);
            for (const vuln of vulns) {
              const searchText = JSON.stringify(vuln).toLowerCase();
              if (searchText.includes(query)) {
                results.push({
                  project: project.id,
                  type: "vulnerability",
                  finding: vuln,
                });
              }
            }
          }
        }

        return {
          content: [{
            type: "text",
            text: JSON.stringify({
              query,
              totalResults: results.length,
              results: results.slice(0, 50),
            }, null, 2),
          }],
        };
      }

      case "get_analysis_raw": {
        const project = args?.project as string;
        const analysisType = args?.analysis_type as string;

        const data = await readAnalysisFile(project, analysisType);

        if (!data) {
          return {
            content: [{
              type: "text",
              text: `No ${analysisType} analysis found for project '${project}'`,
            }],
            isError: true,
          };
        }

        return {
          content: [{
            type: "text",
            text: JSON.stringify(data, null, 2),
          }],
        };
      }

      default:
        return {
          content: [{
            type: "text",
            text: `Unknown tool: ${name}`,
          }],
          isError: true,
        };
    }
  } catch (error) {
    return {
      content: [{
        type: "text",
        text: `Error: ${error instanceof Error ? error.message : String(error)}`,
      }],
      isError: true,
    };
  }
});

// Handle resource listing (for browsing projects as resources)
server.setRequestHandler(ListResourcesRequestSchema, async () => {
  const projects = await getProjects();

  return {
    resources: projects.map(p => ({
      uri: `gibson://projects/${p.id}`,
      name: p.id,
      description: `Analysis data for ${p.id}`,
      mimeType: "application/json",
    })),
  };
});

// Handle resource reading
server.setRequestHandler(ReadResourceRequestSchema, async (request) => {
  const uri = request.params.uri;
  const match = uri.match(/^gibson:\/\/projects\/(.+)$/);

  if (!match) {
    throw new Error(`Invalid resource URI: ${uri}`);
  }

  const projectId = match[1];
  const projects = await getProjects();
  const project = projects.find(p => p.id === projectId);

  if (!project) {
    throw new Error(`Project not found: ${projectId}`);
  }

  // Return summary of all available analyses
  const summary: Record<string, unknown> = {
    project: project.id,
    availableAnalyses: project.availableAnalyses,
  };

  for (const analysisType of project.availableAnalyses) {
    summary[analysisType] = await readAnalysisFile(projectId, analysisType);
  }

  return {
    contents: [{
      uri,
      mimeType: "application/json",
      text: JSON.stringify(summary, null, 2),
    }],
  };
});

// Start the server
async function main() {
  const transport = new StdioServerTransport();
  await server.connect(transport);
  console.error("Zero MCP Server running on stdio");
}

main().catch((error) => {
  console.error("Fatal error:", error);
  process.exit(1);
});
