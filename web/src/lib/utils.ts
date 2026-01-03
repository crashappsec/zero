import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatDate(date: string | Date): string {
  return new Date(date).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds.toFixed(1)}s`;
  const minutes = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60);
  return `${minutes}m ${secs}s`;
}

export function formatRelativeTime(date: string | Date): string {
  const now = new Date();
  const then = new Date(date);
  const diffMs = now.getTime() - then.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffMins < 1) return 'just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;
  return formatDate(date);
}

export function getSeverityColor(severity: string): string {
  switch (severity.toLowerCase()) {
    case 'critical':
      return 'text-red-600 bg-red-100';
    case 'high':
      return 'text-orange-600 bg-orange-100';
    case 'medium':
      return 'text-yellow-600 bg-yellow-100';
    case 'low':
      return 'text-blue-600 bg-blue-100';
    default:
      return 'text-gray-600 bg-gray-100';
  }
}

export function getStatusColor(status: string): string {
  switch (status.toLowerCase()) {
    case 'complete':
    case 'success':
    case 'fresh':
      return 'text-green-600 bg-green-100';
    case 'running':
    case 'scanning':
    case 'cloning':
      return 'text-blue-600 bg-blue-100';
    case 'queued':
    case 'pending':
    case 'stale':
      return 'text-yellow-600 bg-yellow-100';
    case 'failed':
    case 'error':
    case 'expired':
      return 'text-red-600 bg-red-100';
    case 'canceled':
      return 'text-gray-600 bg-gray-100';
    default:
      return 'text-gray-600 bg-gray-100';
  }
}

export function getFreshnessIndicator(level: string): { color: string; label: string } {
  switch (level) {
    case 'fresh':
      return { color: 'bg-green-500', label: 'Fresh' };
    case 'stale':
      return { color: 'bg-yellow-500', label: 'Stale' };
    case 'very_stale':
      return { color: 'bg-orange-500', label: 'Very Stale' };
    case 'expired':
      return { color: 'bg-red-500', label: 'Expired' };
    default:
      return { color: 'bg-gray-500', label: 'Unknown' };
  }
}
