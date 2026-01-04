'use client';

import { useState, useEffect } from 'react';
import { Card, CardTitle, CardContent } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { formatRelativeTime } from '@/lib/utils';
import {
  Activity,
  CheckCircle,
  XCircle,
  Loader2,
  Play,
  AlertTriangle,
  Clock,
} from 'lucide-react';

export interface ActivityEvent {
  id: string;
  type: 'scan_started' | 'scan_complete' | 'scan_failed' | 'scanner_progress';
  target: string;
  message: string;
  timestamp: string;
  metadata?: {
    scanner?: string;
    status?: string;
    projectIds?: string[];
    error?: string;
  };
}

const eventIcons: Record<string, typeof CheckCircle> = {
  scan_started: Play,
  scan_complete: CheckCircle,
  scan_failed: XCircle,
  scanner_progress: Loader2,
};

const eventColors: Record<string, string> = {
  scan_started: 'text-blue-500',
  scan_complete: 'text-green-500',
  scan_failed: 'text-red-500',
  scanner_progress: 'text-yellow-500',
};

function ActivityItem({ event }: { event: ActivityEvent }) {
  const Icon = eventIcons[event.type] || Activity;
  const color = eventColors[event.type] || 'text-gray-500';

  return (
    <div className="flex items-start gap-3 py-3 border-b border-gray-700 last:border-0">
      <div className={`mt-0.5 ${color}`}>
        <Icon className={`h-4 w-4 ${event.type === 'scanner_progress' ? 'animate-spin' : ''}`} />
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="font-medium text-white truncate">{event.target}</span>
          {event.metadata?.scanner && (
            <Badge variant="default" className="text-xs">
              {event.metadata.scanner}
            </Badge>
          )}
        </div>
        <p className="text-sm text-gray-400">{event.message}</p>
        {event.metadata?.error && (
          <p className="text-sm text-red-400 mt-1">{event.metadata.error}</p>
        )}
      </div>
      <span className="text-xs text-gray-500 whitespace-nowrap">
        {formatRelativeTime(event.timestamp)}
      </span>
    </div>
  );
}

interface ActivityFeedProps {
  events: ActivityEvent[];
  maxItems?: number;
  title?: string;
  emptyMessage?: string;
}

export function ActivityFeed({
  events,
  maxItems = 10,
  title = 'Recent Activity',
  emptyMessage = 'No recent activity',
}: ActivityFeedProps) {
  const displayEvents = events.slice(0, maxItems);

  return (
    <Card>
      <CardTitle className="flex items-center gap-2">
        <Activity className="h-5 w-5 text-purple-500" />
        {title}
      </CardTitle>
      <CardContent className="mt-4">
        {displayEvents.length === 0 ? (
          <div className="text-center py-8">
            <Clock className="h-8 w-8 text-gray-600 mx-auto mb-2" />
            <p className="text-gray-400">{emptyMessage}</p>
          </div>
        ) : (
          <div className="divide-y divide-gray-700">
            {displayEvents.map((event) => (
              <ActivityItem key={event.id} event={event} />
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

// Hook to manage activity events from various sources
export function useActivityFeed() {
  const [events, setEvents] = useState<ActivityEvent[]>([]);

  const addEvent = (event: Omit<ActivityEvent, 'id' | 'timestamp'>) => {
    const newEvent: ActivityEvent = {
      ...event,
      id: Math.random().toString(36).slice(2),
      timestamp: new Date().toISOString(),
    };
    setEvents((prev) => [newEvent, ...prev].slice(0, 100));
  };

  const clearEvents = () => setEvents([]);

  return { events, addEvent, clearEvents };
}
