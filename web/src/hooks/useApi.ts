'use client';

import { useState, useEffect, useCallback } from 'react';
import { api, connectScanWS, streamChat } from '@/lib/api';
import type { Project, ScanJob, StreamChunk, QueueStats, AgentInfo } from '@/lib/types';

// Generic hook for fetching data
export function useFetch<T>(
  fetcher: () => Promise<T>,
  deps: unknown[] = []
): { data: T | null; loading: boolean; error: Error | null; refetch: () => void } {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetch = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await fetcher();
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Unknown error'));
    } finally {
      setLoading(false);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, deps);

  useEffect(() => {
    fetch();
  }, [fetch]);

  return { data, loading, error, refetch: fetch };
}

// Projects hook
export function useProjects() {
  return useFetch(async () => {
    const res = await api.projects.list();
    return res.data;
  }, []);
}

// Single project hook
export function useProject(id: string) {
  return useFetch(() => api.projects.get(id), [id]);
}

// Agents hook
export function useAgents() {
  return useFetch(async () => {
    const res = await api.agents();
    return res.data;
  }, []);
}

// Scan queue stats hook
export function useQueueStats(pollInterval = 5000) {
  const [stats, setStats] = useState<QueueStats | null>(null);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const data = await api.scans.stats();
        setStats(data);
      } catch {
        // Ignore errors
      }
    };

    fetchStats();
    const interval = setInterval(fetchStats, pollInterval);
    return () => clearInterval(interval);
  }, [pollInterval]);

  return stats;
}

// Active scans hook with polling
export function useActiveScans(pollInterval = 3000) {
  const [scans, setScans] = useState<ScanJob[]>([]);

  useEffect(() => {
    const fetchScans = async () => {
      try {
        const res = await api.scans.active();
        setScans(res.data);
      } catch {
        // Ignore errors
      }
    };

    fetchScans();
    const interval = setInterval(fetchScans, pollInterval);
    return () => clearInterval(interval);
  }, [pollInterval]);

  return scans;
}

// Scan progress hook with WebSocket
export function useScanProgress(jobId: string | null) {
  const [job, setJob] = useState<ScanJob | null>(null);
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    if (!jobId) return;

    // Initial fetch
    api.scans.get(jobId).then(setJob).catch(() => {});

    // WebSocket connection
    const ws = connectScanWS(
      jobId,
      (msg) => {
        // Update job based on WebSocket messages
        const data = msg as { type: string; status?: string; progress?: ScanJob['progress'] };
        if (data.type === 'job_status') {
          setJob((prev) => (prev ? { ...prev, status: data.status as ScanJob['status'] } : prev));
        } else if (data.type === 'scanner_progress' && data.progress) {
          setJob((prev) => (prev ? { ...prev, progress: data.progress } : prev));
        } else if (data.type === 'scan_complete') {
          // Refetch to get final state
          api.scans.get(jobId).then(setJob).catch(() => {});
        }
      },
      () => setConnected(false)
    );

    ws.onopen = () => setConnected(true);
    ws.onclose = () => setConnected(false);

    return () => ws.close();
  }, [jobId]);

  return { job, connected };
}

// Chat hook
export function useChat(agentId = 'zero') {
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [messages, setMessages] = useState<{ role: 'user' | 'assistant'; content: string }[]>([]);
  const [isStreaming, setIsStreaming] = useState(false);
  const [streamingContent, setStreamingContent] = useState('');

  const sendMessage = useCallback(
    async (message: string, projectId?: string) => {
      // Add user message
      setMessages((prev) => [...prev, { role: 'user', content: message }]);
      setIsStreaming(true);
      setStreamingContent('');

      return new Promise<void>((resolve, reject) => {
        streamChat(
          message,
          { session_id: sessionId || undefined, agent_id: agentId, project_id: projectId },
          (chunk: StreamChunk) => {
            if (chunk.type === 'start') {
              setSessionId(chunk.session_id);
            } else if (chunk.type === 'delta') {
              setStreamingContent((prev) => prev + (chunk.content || ''));
            } else if (chunk.type === 'done') {
              setMessages((prev) => [...prev, { role: 'assistant', content: chunk.content || '' }]);
              setStreamingContent('');
              setIsStreaming(false);
              resolve();
            } else if (chunk.type === 'error') {
              setIsStreaming(false);
              reject(new Error(chunk.error));
            }
          },
          (error) => {
            setIsStreaming(false);
            reject(error);
          }
        );
      });
    },
    [sessionId, agentId]
  );

  const reset = useCallback(() => {
    setSessionId(null);
    setMessages([]);
    setStreamingContent('');
    setIsStreaming(false);
  }, []);

  return {
    sessionId,
    messages,
    isStreaming,
    streamingContent,
    sendMessage,
    reset,
  };
}
