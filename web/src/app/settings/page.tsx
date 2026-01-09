'use client';

import { useState, useEffect } from 'react';
import { MainLayout } from '@/components/layout/Sidebar';
import { Card, CardTitle, CardContent } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { useToast } from '@/components/ui/Toast';
import { useFetch } from '@/hooks/useApi';
import { api } from '@/lib/api';
import type { ScannerInfo, ProfileInfo } from '@/lib/types';
import { ThemeToggle } from '@/components/ui/ThemeToggle';
import {
  Settings,
  Server,
  Shield,
  Clock,
  RefreshCw,
  CheckCircle,
  XCircle,
  AlertTriangle,
  Cpu,
  HardDrive,
  Zap,
  Palette,
  Keyboard,
} from 'lucide-react';

function HealthStatus() {
  const { data: health, loading, error, refetch } = useFetch(
    () => api.health(),
    []
  );

  return (
    <Card>
      <CardTitle className="flex items-center gap-2">
        <Server className="h-5 w-5 text-blue-500" />
        API Server Status
      </CardTitle>
      <CardContent className="mt-4">
        {loading ? (
          <div className="flex items-center gap-2 text-gray-400">
            <RefreshCw className="h-4 w-4 animate-spin" />
            Checking...
          </div>
        ) : error ? (
          <div className="flex items-center gap-2 text-red-400">
            <XCircle className="h-5 w-5" />
            <span>API Server Offline</span>
          </div>
        ) : health ? (
          <div className="space-y-4">
            <div className="flex items-center gap-2 text-green-400">
              <CheckCircle className="h-5 w-5" />
              <span className="font-medium">Connected</span>
            </div>
            <div className="grid gap-2 text-sm">
              <div className="flex justify-between">
                <span className="text-gray-400">Status</span>
                <Badge variant="success">{health.status}</Badge>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-400">Version</span>
                <span className="text-white">{health.version}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-400">Last Check</span>
                <span className="text-white">
                  {new Date(health.timestamp).toLocaleTimeString()}
                </span>
              </div>
            </div>
            <Button variant="outline" size="sm" onClick={refetch}>
              <RefreshCw className="h-4 w-4 mr-2" />
              Refresh
            </Button>
          </div>
        ) : null}
      </CardContent>
    </Card>
  );
}

function ScannersConfig() {
  const { data: scannersData, loading } = useFetch(
    () => api.scanners.list(),
    []
  );
  const scanners = scannersData?.data || [];

  return (
    <Card>
      <CardTitle className="flex items-center gap-2">
        <Shield className="h-5 w-5 text-green-500" />
        Available Scanners
      </CardTitle>
      <CardContent className="mt-4">
        {loading ? (
          <div className="text-gray-400">Loading scanners...</div>
        ) : scanners.length === 0 ? (
          <div className="text-gray-400">No scanners available</div>
        ) : (
          <div className="space-y-3">
            {scanners.map((scanner) => (
              <div
                key={scanner.name}
                className="flex items-start justify-between p-3 bg-gray-800/50 rounded-lg"
              >
                <div>
                  <p className="font-medium text-white">{scanner.name}</p>
                  <p className="text-sm text-gray-400">{scanner.description}</p>
                  {scanner.features && scanner.features.length > 0 && (
                    <div className="flex flex-wrap gap-1 mt-2">
                      {scanner.features.map((feature) => (
                        <Badge key={feature} variant="default" className="text-xs">
                          {feature}
                        </Badge>
                      ))}
                    </div>
                  )}
                </div>
                <CheckCircle className="h-5 w-5 text-green-500 shrink-0" />
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function ProfilesConfig() {
  const { data: profilesData, loading } = useFetch(
    () => api.profiles.list(),
    []
  );
  const profiles = profilesData?.data || [];

  return (
    <Card>
      <CardTitle className="flex items-center gap-2">
        <Zap className="h-5 w-5 text-yellow-500" />
        Scan Profiles
      </CardTitle>
      <CardContent className="mt-4">
        {loading ? (
          <div className="text-gray-400">Loading profiles...</div>
        ) : profiles.length === 0 ? (
          <div className="text-gray-400">No profiles available</div>
        ) : (
          <div className="space-y-3">
            {profiles.map((profile) => (
              <div
                key={profile.name}
                className="p-3 bg-gray-800/50 rounded-lg"
              >
                <div className="flex items-center justify-between mb-2">
                  <p className="font-medium text-white">{profile.name}</p>
                  {profile.estimated_time && (
                    <span className="text-xs text-gray-500 flex items-center gap-1">
                      <Clock className="h-3 w-3" />
                      {profile.estimated_time}
                    </span>
                  )}
                </div>
                <p className="text-sm text-gray-400 mb-2">{profile.description}</p>
                {profile.scanners && profile.scanners.length > 0 && (
                  <div className="flex flex-wrap gap-1">
                    {profile.scanners.map((scanner) => (
                      <Badge key={scanner} variant="info" className="text-xs">
                        {scanner}
                      </Badge>
                    ))}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function EnvironmentInfo() {
  const [envStatus, setEnvStatus] = useState<{
    anthropic_key: boolean;
    github_token: boolean;
  }>({ anthropic_key: false, github_token: false });

  useEffect(() => {
    // Check if API keys are configured (via health endpoint or config)
    api.config()
      .then((config: any) => {
        setEnvStatus({
          anthropic_key: config?.has_anthropic_key || false,
          github_token: config?.has_github_token || false,
        });
      })
      .catch(() => {
        // Ignore errors
      });
  }, []);

  return (
    <Card>
      <CardTitle className="flex items-center gap-2">
        <HardDrive className="h-5 w-5 text-purple-500" />
        Environment
      </CardTitle>
      <CardContent className="mt-4">
        <div className="space-y-3">
          <div className="flex items-center justify-between p-3 bg-gray-800/50 rounded-lg">
            <div className="flex items-center gap-2">
              <span className="text-white">ANTHROPIC_API_KEY</span>
            </div>
            {envStatus.anthropic_key ? (
              <Badge variant="success">Configured</Badge>
            ) : (
              <Badge variant="warning">Not Set</Badge>
            )}
          </div>
          <div className="flex items-center justify-between p-3 bg-gray-800/50 rounded-lg">
            <div className="flex items-center gap-2">
              <span className="text-white">GITHUB_TOKEN</span>
            </div>
            {envStatus.github_token ? (
              <Badge variant="success">Configured</Badge>
            ) : (
              <Badge variant="warning">Not Set</Badge>
            )}
          </div>
        </div>
        <p className="mt-4 text-xs text-gray-500">
          Environment variables are required for agent chat and GitHub repository access.
          Set these in your shell or .env file before starting the API server.
        </p>
      </CardContent>
    </Card>
  );
}

export default function SettingsPage() {
  return (
    <MainLayout>
      <div className="space-y-6">
        {/* Header */}
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            <Settings className="h-6 w-6 text-gray-400" />
            Settings
          </h1>
          <p className="mt-1 text-gray-400">
            View system configuration and status
          </p>
        </div>

        {/* Preferences */}
        <Card>
          <CardTitle className="flex items-center gap-2">
            <Palette className="h-5 w-5 text-pink-500" />
            Preferences
          </CardTitle>
          <CardContent className="mt-4">
            <div className="space-y-4">
              <div className="flex items-center justify-between p-3 bg-gray-800/50 rounded-lg">
                <div>
                  <p className="font-medium text-white">Theme</p>
                  <p className="text-sm text-gray-400">Choose your preferred color scheme</p>
                </div>
                <ThemeToggle />
              </div>
              <div className="flex items-center justify-between p-3 bg-gray-800/50 rounded-lg">
                <div>
                  <p className="font-medium text-white">Keyboard Shortcuts</p>
                  <p className="text-sm text-gray-400">Quick navigation with keyboard</p>
                </div>
                <div className="flex items-center gap-2">
                  <kbd className="rounded bg-gray-900 px-2 py-1 text-xs font-mono text-gray-400 border border-gray-700">
                    ?
                  </kbd>
                  <span className="text-sm text-gray-500">to view</span>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        <div className="grid gap-6 lg:grid-cols-2">
          <HealthStatus />
          <EnvironmentInfo />
        </div>

        <div className="grid gap-6 lg:grid-cols-2">
          <ScannersConfig />
          <ProfilesConfig />
        </div>

        {/* About */}
        <Card>
          <CardTitle>About Zero</CardTitle>
          <CardContent className="mt-4">
            <p className="text-gray-400">
              Zero is a developer intelligence platform with specialist AI agents for
              comprehensive repository analysis. It provides insights into code ownership,
              technology stack, dependencies, DevOps practices, developer experience,
              and code quality.
            </p>
            <p className="mt-4 text-sm text-gray-500">
              Named after characters from the movie Hackers (1995) - &quot;Hack the planet!&quot;
            </p>
          </CardContent>
        </Card>
      </div>
    </MainLayout>
  );
}
