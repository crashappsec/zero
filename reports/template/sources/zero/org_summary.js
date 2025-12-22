// Org-wide summary aggregating all repositories
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const dataDir = path.join(__dirname, 'data');

function aggregateOrgData() {
  const summary = {
    total_repos: 0,
    total_packages: 0,
    total_vulns: 0,
    critical_vulns: 0,
    high_vulns: 0,
    medium_vulns: 0,
    low_vulns: 0,
    total_secrets: 0,
    repos_with_vulns: 0,
    repos_with_secrets: 0,
    repos_with_critical: 0,
    is_org_mode: false
  };

  try {
    const items = fs.readdirSync(dataDir);
    if (!items.length) return [summary];

    const firstPath = path.join(dataDir, items[0]);
    const stat = fs.statSync(firstPath);

    if (stat.isDirectory()) {
      // Org mode
      summary.is_org_mode = true;

      for (const repoName of items) {
        const repoPath = path.join(dataDir, repoName);
        const repoStat = fs.statSync(repoPath);
        if (!repoStat.isDirectory()) continue;

        summary.total_repos++;
        aggregateRepo(repoPath, summary);
      }
    } else {
      // Single repo mode
      summary.total_repos = 1;
      aggregateRepo(dataDir, summary);
    }
  } catch (e) {
    // Return default summary on error
  }

  return [summary];
}

function aggregateRepo(repoPath, summary) {
  try {
    // Load packages.json
    const packagesPath = path.join(repoPath, 'packages.json');
    if (fs.existsSync(packagesPath)) {
      const data = JSON.parse(fs.readFileSync(packagesPath, 'utf8'));
      const vulns = data.summary?.vulns || {};
      const licenses = data.summary?.licenses || {};

      summary.total_packages += licenses.total_packages || 0;
      summary.total_vulns += vulns.total_vulnerabilities || 0;
      summary.critical_vulns += vulns.critical || 0;
      summary.high_vulns += vulns.high || 0;
      summary.medium_vulns += vulns.medium || 0;
      summary.low_vulns += vulns.low || 0;

      if ((vulns.total_vulnerabilities || 0) > 0) summary.repos_with_vulns++;
      if ((vulns.critical || 0) > 0) summary.repos_with_critical++;
    }

    // Load code-security.json
    const securityPath = path.join(repoPath, 'code-security.json');
    if (fs.existsSync(securityPath)) {
      const data = JSON.parse(fs.readFileSync(securityPath, 'utf8'));
      const secrets = data.summary?.secrets?.total || 0;
      summary.total_secrets += secrets;
      if (secrets > 0) summary.repos_with_secrets++;
    }
  } catch (e) {
    // Skip repo on error
  }
}

export const data = aggregateOrgData();
