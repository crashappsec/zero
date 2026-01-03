'use client';

import { cn } from '@/lib/utils';

interface BadgeProps {
  children: React.ReactNode;
  variant?: 'default' | 'success' | 'warning' | 'error' | 'info';
  size?: 'sm' | 'md';
  className?: string;
}

const variants = {
  default: 'bg-gray-700 text-gray-300',
  success: 'bg-green-900/50 text-green-400 border-green-700',
  warning: 'bg-yellow-900/50 text-yellow-400 border-yellow-700',
  error: 'bg-red-900/50 text-red-400 border-red-700',
  info: 'bg-blue-900/50 text-blue-400 border-blue-700',
};

const sizes = {
  sm: 'text-xs px-1.5 py-0.5',
  md: 'text-sm px-2 py-1',
};

export function Badge({ children, variant = 'default', size = 'sm', className }: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded border font-medium',
        variants[variant],
        sizes[size],
        className
      )}
    >
      {children}
    </span>
  );
}

export function SeverityBadge({ severity }: { severity: string }) {
  const variant =
    severity === 'critical'
      ? 'error'
      : severity === 'high'
        ? 'warning'
        : severity === 'medium'
          ? 'warning'
          : 'info';

  return (
    <Badge variant={variant} size="sm">
      {severity.toUpperCase()}
    </Badge>
  );
}

export function StatusBadge({ status }: { status: string }) {
  const variant =
    status === 'complete' || status === 'success'
      ? 'success'
      : status === 'failed' || status === 'error'
        ? 'error'
        : status === 'running' || status === 'scanning'
          ? 'info'
          : 'default';

  return (
    <Badge variant={variant} size="sm">
      {status}
    </Badge>
  );
}
