'use client';

import { useState, Suspense, useEffect, useRef } from 'react';
import { useSearchParams } from 'next/navigation';
import { MainLayout } from '@/components/layout/Sidebar';
import { Card, CardTitle, CardContent } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Badge, StatusBadge } from '@/components/ui/Badge';
import { useToast } from '@/components/ui/Toast';
import { useActiveScans, useQueueStats, useFetch } from '@/hooks/useApi';
import { useNotifications } from '@/hooks/useNotifications';
import { api } from '@/lib/api';
import { formatRelativeTime, formatDuration } from '@/lib/utils';
import type { ScanJob, ProfileInfo } from '@/lib/types';
import {
  Scan,
  Play,
  X,
  Clock,
  CheckCircle,
  AlertTriangle,
  Loader2,
  GitBranch,
  Bell,
  BellOff,
} from 'lucide-react';

function NewScanForm({ onStart, initialTarget = '' }: { onStart: (job: { job_id: string }) => void; initialTarget?: string }) {
  const [target, setTarget] = useState(initialTarget);
  const [profile, setProfile] = useState('all-quick');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showDropdown, setShowDropdown] = useState(false);
  const [filterText, setFilterText] = useState('');
  const dropdownRef = useRef<HTMLDivElement>(null);
  const toast = useToast();

  const { data: profilesData } = useFetch(() => api.profiles.list(), []);
  const { data: projectsData } = useFetch(() => api.projects.list(), []);
  const profiles = profilesData?.data || [];
  const projects = projectsData?.data || [];

  // Filter projects based on input
  const filteredProjects = filterText
    ? projects.filter(p =>
        p.id.toLowerCase().includes(filterText.toLowerCase()) ||
        p.name.toLowerCase().includes(filterText.toLowerCase())
      )
    : projects;

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setShowDropdown(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!target.trim()) return;

    setLoading(true);
    setError(null);

    try {
      const result = await api.scans.start(target.trim(), profile);
      onStart(result);
      setTarget('');
      setFilterText('');
      toast.success('Scan started', `Scanning ${target.trim()}`);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to start scan';
      setError(message);
      toast.error('Failed to start scan', message);
    } finally {
      setLoading(false);
    }
  };

  const handleSelectProject = (projectId: string) => {
    setTarget(projectId);
    setFilterText('');
    setShowDropdown(false);
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setTarget(value);
    setFilterText(value);
    setShowDropdown(true);
  };

  return (
    <Card>
      <CardTitle className="flex items-center gap-2">
        <Play className="h-5 w-5 text-green-500" />
        New Scan
      </CardTitle>
      <CardContent className="mt-4">
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="relative" ref={dropdownRef}>
            <label className="block text-sm font-medium text-gray-300 mb-1">
              Repository
            </label>
            <div className="relative">
              <Input
                placeholder="owner/repo or org name"
                value={target}
                onChange={handleInputChange}
                onFocus={() => setShowDropdown(true)}
                icon={<GitBranch className="h-4 w-4" />}
              />
              {projects.length > 0 && (
                <button
                  type="button"
                  onClick={() => setShowDropdown(!showDropdown)}
                  className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-300"
                >
                  <svg className={`h-4 w-4 transition-transform ${showDropdown ? 'rotate-180' : ''}`} fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                  </svg>
                </button>
              )}
            </div>

            {/* Dropdown with hydrated repos */}
            {showDropdown && projects.length > 0 && (
              <div className="absolute z-50 mt-1 w-full rounded-md border border-gray-700 bg-gray-800 shadow-lg max-h-64 overflow-y-auto">
                {filteredProjects.length > 0 ? (
                  <>
                    <div className="px-3 py-2 text-xs font-medium text-gray-500 border-b border-gray-700">
                      Hydrated Repositories ({filteredProjects.length})
                    </div>
                    {Object.entries(
                      filteredProjects.reduce((acc, project) => {
                        const org = project.owner || 'unknown';
                        if (!acc[org]) acc[org] = [];
                        acc[org].push(project);
                        return acc;
                      }, {} as Record<string, typeof projects>)
                    ).map(([org, orgProjects]) => (
                      <div key={org}>
                        <div className="px-3 py-1.5 text-xs font-semibold text-gray-400 bg-gray-900/50">
                          {org}
                        </div>
                        {orgProjects.map((project) => (
                          <button
                            key={project.id}
                            type="button"
                            onClick={() => handleSelectProject(project.id)}
                            className="w-full px-3 py-2 text-left text-sm text-gray-200 hover:bg-gray-700/50 flex items-center justify-between"
                          >
                            <span className="flex items-center gap-2">
                              <GitBranch className="h-3.5 w-3.5 text-gray-500" />
                              {project.name}
                            </span>
                            {project.freshness && (
                              <span className={`text-xs ${
                                project.freshness.level === 'fresh' ? 'text-green-500' :
                                project.freshness.level === 'stale' ? 'text-yellow-500' :
                                'text-red-500'
                              }`}>
                                {project.freshness.age_string}
                              </span>
                            )}
                          </button>
                        ))}
                      </div>
                    ))}
                  </>
                ) : filterText ? (
                  <div className="px-3 py-4 text-center text-sm text-gray-500">
                    No matching repos. Press Enter to scan &ldquo;{filterText}&rdquo;
                  </div>
                ) : null}
              </div>
            )}

            <p className="mt-1 text-xs text-gray-500">
              {projects.length > 0
                ? 'Select from hydrated repos or enter a new one'
                : 'Enter owner/repo (e.g., expressjs/express) or org name'}
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1">
              Profile
            </label>
            <select
              value={profile}
              onChange={(e) => setProfile(e.target.value)}
              className="w-full rounded-md border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 focus:border-green-500 focus:outline-none focus:ring-1 focus:ring-green-500"
            >
              {profiles.length > 0 ? (
                profiles.map((p) => (
                  <option key={p.name} value={p.name}>
                    {p.name} - {p.description}
                  </option>
                ))
              ) : (
                <>
                  <option value="all-quick">all-quick - Fast analysis across all dimensions</option>
                  <option value="all-complete">all-complete - Complete analysis with all features</option>
                  <option value="security-focused">security-focused - Deep security analysis</option>
                </>
              )}
            </select>
          </div>

          {error && (
            <p className="text-sm text-red-400">{error}</p>
          )}

          <Button type="submit" loading={loading} className="w-full">
            Start Scan
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

function ScanJobCard({ job, onCancel }: { job: ScanJob; onCancel?: () => void }) {
  const isActive = job.status === 'queued' || job.status === 'cloning' || job.status === 'scanning';
  const isFailed = job.status === 'failed';
  const isComplete = job.status === 'complete';

  return (
    <Card className={isActive ? 'border-blue-700' : isFailed ? 'border-red-700' : ''}>
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <div className="relative">
            {isActive ? (
              <Loader2 className="h-5 w-5 text-blue-500 animate-spin" />
            ) : isComplete ? (
              <CheckCircle className="h-5 w-5 text-green-500" />
            ) : isFailed ? (
              <AlertTriangle className="h-5 w-5 text-red-500" />
            ) : (
              <Scan className="h-5 w-5 text-gray-500" />
            )}
          </div>
          <div>
            <p className="font-medium text-white">{job.target}</p>
            <div className="flex items-center gap-2 mt-1">
              <StatusBadge status={job.status} />
              <Badge variant="default">{job.profile}</Badge>
            </div>
          </div>
        </div>
        {isActive && onCancel && (
          <Button variant="ghost" size="sm" onClick={onCancel}>
            <X className="h-4 w-4" />
          </Button>
        )}
      </div>

      {/* Progress */}
      {isActive && job.progress && (
        <div className="mt-4">
          <div className="flex items-center justify-between text-sm text-gray-400 mb-1">
            <span>{job.progress.phase}</span>
            <span>
              {job.progress.scanners_complete}/{job.progress.scanners_total} scanners
            </span>
          </div>
          <div className="h-2 rounded-full bg-gray-700 overflow-hidden">
            <div
              className="h-full bg-blue-500 transition-all duration-300 relative scan-progress"
              style={{
                width: `${
                  job.progress.scanners_total
                    ? (job.progress.scanners_complete / job.progress.scanners_total) * 100
                    : 0
                }%`,
              }}
            />
          </div>
          {job.progress.current_scanner && (
            <p className="mt-1 text-xs text-gray-500">
              Running: {job.progress.current_scanner}
            </p>
          )}
        </div>
      )}

      {/* Metadata */}
      <div className="mt-4 flex items-center gap-4 text-sm text-gray-500">
        <span className="flex items-center gap-1">
          <Clock className="h-4 w-4" />
          {formatRelativeTime(job.started_at)}
        </span>
        {job.duration_seconds && (
          <span>Duration: {formatDuration(job.duration_seconds)}</span>
        )}
      </div>

      {/* Error */}
      {job.error && (
        <p className="mt-2 text-sm text-red-400">{job.error}</p>
      )}

      {/* Results */}
      {isComplete && job.project_ids && job.project_ids.length > 0 && (
        <div className="mt-4">
          <p className="text-sm text-gray-400 mb-2">Scanned projects:</p>
          <div className="flex flex-wrap gap-2">
            {job.project_ids.map((id) => (
              <a
                key={id}
                href={`/projects/${encodeURIComponent(id)}`}
                className="text-sm text-green-400 hover:text-green-300"
              >
                {id}
              </a>
            ))}
          </div>
        </div>
      )}
    </Card>
  );
}

function ScanFormWithParams({ onStart }: { onStart: (job: { job_id: string }) => void }) {
  const searchParams = useSearchParams();
  const target = searchParams.get('target') || '';
  return <NewScanForm onStart={onStart} initialTarget={target} />;
}

function ScansPageContent() {
  const toast = useToast();
  const notifications = useNotifications();
  const activeScans = useActiveScans(2000);
  const stats = useQueueStats();
  const { data: historyData, refetch: refetchHistory } = useFetch(
    () => api.scans.history(),
    []
  );

  // Track previous scan statuses to detect changes
  const prevScansRef = useRef<Map<string, string>>(new Map());

  const history = historyData?.data || [];

  // Detect scan status changes and show toast + browser notifications
  useEffect(() => {
    const prevScans = prevScansRef.current;

    activeScans.forEach((scan) => {
      const prevStatus = prevScans.get(scan.job_id);

      if (prevStatus && prevStatus !== scan.status) {
        if (scan.status === 'complete') {
          toast.success('Scan complete', `${scan.target} finished successfully`);
          notifications.notifyScanComplete(scan.target, true, scan.project_ids);
          refetchHistory();
        } else if (scan.status === 'failed') {
          toast.error('Scan failed', scan.error || `${scan.target} failed`);
          notifications.notifyScanComplete(scan.target, false);
          refetchHistory();
        } else if (scan.status === 'canceled') {
          toast.warning('Scan canceled', `${scan.target} was canceled`);
          refetchHistory();
        }
      }

      prevScans.set(scan.job_id, scan.status);
    });
  }, [activeScans, toast, notifications, refetchHistory]);

  const handleScanStart = (job: { job_id: string }) => {
    // The active scans will auto-update via polling
  };

  const handleCancel = async (jobId: string) => {
    try {
      await api.scans.cancel(jobId);
      toast.info('Canceling scan...');
    } catch {
      toast.error('Failed to cancel scan');
    }
  };

  const handleEnableNotifications = async () => {
    const granted = await notifications.requestPermission();
    if (granted) {
      toast.success('Notifications enabled', 'You will receive browser notifications for scan events');
    } else {
      toast.warning('Notifications blocked', 'Please enable notifications in your browser settings');
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Scans</h1>
          <p className="mt-1 text-gray-400">
            Manage repository scans and view history
          </p>
        </div>
        {notifications.supported && (
          <Button
            variant={notifications.permission === 'granted' ? 'ghost' : 'outline'}
            size="sm"
            onClick={handleEnableNotifications}
            icon={notifications.permission === 'granted' ? <Bell className="h-4 w-4" /> : <BellOff className="h-4 w-4" />}
          >
            {notifications.permission === 'granted' ? 'Notifications On' : 'Enable Notifications'}
          </Button>
        )}
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4">
        <Card className="text-center">
          <p className="text-2xl font-bold text-blue-500">{stats?.running_jobs ?? 0}</p>
          <p className="text-sm text-gray-400">Running</p>
        </Card>
        <Card className="text-center">
          <p className="text-2xl font-bold text-yellow-500">{stats?.queued_jobs ?? 0}</p>
          <p className="text-sm text-gray-400">Queued</p>
        </Card>
        <Card className="text-center">
          <p className="text-2xl font-bold text-green-500">{stats?.completed_jobs ?? 0}</p>
          <p className="text-sm text-gray-400">Completed</p>
        </Card>
        <Card className="text-center">
          <p className="text-2xl font-bold text-red-500">{stats?.failed_jobs ?? 0}</p>
          <p className="text-sm text-gray-400">Failed</p>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* New Scan Form */}
        <div>
          <Suspense fallback={<Card className="animate-pulse h-64"><div /></Card>}>
            <ScanFormWithParams onStart={handleScanStart} />
          </Suspense>
          </div>

          {/* Active & History */}
          <div className="lg:col-span-2 space-y-6">
            {/* Active Scans */}
            {activeScans.length > 0 && (
              <section>
                <h2 className="text-lg font-semibold text-white mb-4">Active Scans</h2>
                <div className="space-y-4">
                  {activeScans.map((scan) => (
                    <ScanJobCard
                      key={scan.job_id}
                      job={scan}
                      onCancel={() => handleCancel(scan.job_id)}
                    />
                  ))}
                </div>
              </section>
            )}

            {/* History */}
            <section>
              <h2 className="text-lg font-semibold text-white mb-4">Recent History</h2>
              {history.length > 0 ? (
                <div className="space-y-4">
                  {history.map((scan) => (
                    <ScanJobCard key={scan.job_id} job={scan} />
                  ))}
                </div>
              ) : (
                <Card className="text-center py-8">
                  <p className="text-gray-400">No scan history yet</p>
                </Card>
              )}
            </section>
          </div>
        </div>
      </div>
  );
}

export default function ScansPage() {
  return (
    <MainLayout>
      <ScansPageContent />
    </MainLayout>
  );
}
