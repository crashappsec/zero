// List all repositories in the organization
// This file discovers repos from the data directory structure
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const dataDir = path.join(__dirname, 'data');

function discoverRepos() {
  const repos = [];

  try {
    // Check if we're in org mode (data dir contains repo subdirectories)
    // or repo mode (data dir contains JSON files directly)
    const items = fs.readdirSync(dataDir);

    // Check first item to determine mode
    const firstItem = items[0];
    if (!firstItem) return repos;

    const firstPath = path.join(dataDir, firstItem);
    const stat = fs.statSync(firstPath);

    if (stat.isDirectory()) {
      // Org mode: each subdirectory is a repo
      for (const repoName of items) {
        const repoPath = path.join(dataDir, repoName);
        const repoStat = fs.statSync(repoPath);
        if (!repoStat.isDirectory()) continue;

        const repo = loadRepoSummary(repoName, repoPath);
        if (repo) repos.push(repo);
      }
    } else {
      // Repo mode: data dir contains JSON files directly
      // Extract repo name from metadata or use 'current'
      const repo = loadRepoSummary('current', dataDir);
      if (repo) repos.push(repo);
    }
  } catch (e) {
    // Return empty if no repos found
  }

  return repos;
}

function loadRepoSummary(repoName, repoPath) {
  try {
    // Try to load packages.json for summary data
    const packagesPath = path.join(repoPath, 'packages.json');
    const sbomPath = path.join(repoPath, 'sbom.json');
    const securityPath = path.join(repoPath, 'code-security.json');

    let packages = 0, vulns = 0, secrets = 0, critical = 0, high = 0;

    if (fs.existsSync(packagesPath)) {
      const data = JSON.parse(fs.readFileSync(packagesPath, 'utf8'));
      packages = data.summary?.licenses?.total_packages || 0;
      vulns = data.summary?.vulns?.total_vulnerabilities || 0;
      critical = data.summary?.vulns?.critical || 0;
      high = data.summary?.vulns?.high || 0;
    }

    if (fs.existsSync(securityPath)) {
      const data = JSON.parse(fs.readFileSync(securityPath, 'utf8'));
      secrets = data.summary?.secrets?.total || 0;
    }

    return {
      name: repoName,
      path: repoPath,
      packages,
      vulns,
      secrets,
      critical,
      high,
      hasData: packages > 0 || vulns > 0 || secrets > 0
    };
  } catch (e) {
    return null;
  }
}

// Export repos list
export const data = discoverRepos();
