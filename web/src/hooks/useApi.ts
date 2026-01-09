'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { api, connectScanWS, streamChat } from '@/lib/api';
import type { Repo, Project, ScanJob, StreamChunk, QueueStats, AgentInfo, ToolCallInfo } from '@/lib/types';

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

// Repos hooks (renamed from Projects)
export function useRepos() {
  return useFetch(async () => {
    const res = await api.repos.list();
    return res.data;
  }, []);
}

export function useRepo(id: string) {
  return useFetch(() => api.repos.get(id), [id]);
}

// Backwards compatibility: Projects hooks
export function useProjects() {
  return useFetch(async () => {
    const res = await api.repos.list();
    return res.data;
  }, []);
}

export function useProject(id: string) {
  return useFetch(() => api.repos.get(id), [id]);
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

// Chat stage type for UX feedback
export type ChatStage = 'idle' | 'sending' | 'thinking' | 'tool_running' | 'responding' | 'delegating';

// Delegation info for sub-agent progress
export interface DelegationInfo {
  agentName: string;
  event: string;
  toolCalls: ToolCallInfo[];
}

// Chat hook with tool call tracking and elapsed time
export function useChat(agentId = 'zero') {
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [messages, setMessages] = useState<{ role: 'user' | 'assistant'; content: string; toolCalls?: ToolCallInfo[] }[]>([]);
  const [isStreaming, setIsStreaming] = useState(false);
  const [streamingContent, setStreamingContent] = useState('');
  const [activeToolCalls, setActiveToolCalls] = useState<ToolCallInfo[]>([]);
  const [stage, setStage] = useState<ChatStage>('idle');
  const [elapsedTime, setElapsedTime] = useState(0);
  const [startTime, setStartTime] = useState<number | null>(null);
  const [delegation, setDelegation] = useState<DelegationInfo | null>(null);

  // Store cleanup function in a ref to avoid stale closures
  const cleanupRef = useRef<(() => void) | null>(null);
  const toolCallCounterRef = useRef(0);
  const sessionIdRef = useRef<string | null>(null);
  const timerRef = useRef<NodeJS.Timeout | null>(null);

  // Keep sessionIdRef in sync
  useEffect(() => {
    sessionIdRef.current = sessionId;
  }, [sessionId]);

  // Elapsed time timer
  useEffect(() => {
    if (startTime) {
      timerRef.current = setInterval(() => {
        setElapsedTime(Math.floor((Date.now() - startTime) / 1000));
      }, 100);
    } else {
      if (timerRef.current) {
        clearInterval(timerRef.current);
        timerRef.current = null;
      }
      setElapsedTime(0);
    }
    return () => {
      if (timerRef.current) {
        clearInterval(timerRef.current);
      }
    };
  }, [startTime]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (cleanupRef.current) {
        cleanupRef.current();
      }
      if (timerRef.current) {
        clearInterval(timerRef.current);
      }
    };
  }, []);

  const sendMessage = useCallback(
    async (message: string, projectId?: string) => {
      // Cancel any existing stream
      if (cleanupRef.current) {
        cleanupRef.current();
        cleanupRef.current = null;
      }

      // Add user message immediately (optimistic UI)
      setMessages((prev) => [...prev, { role: 'user', content: message }]);
      setIsStreaming(true);
      setStreamingContent('');
      setActiveToolCalls([]);
      setStage('sending');
      setStartTime(Date.now());

      let currentToolCalls: ToolCallInfo[] = [];
      let isCancelled = false;

      return new Promise<void>((resolve, reject) => {
        const cleanup = streamChat(
          message,
          { session_id: sessionIdRef.current || undefined, agent_id: agentId, project_id: projectId },
          (chunk: StreamChunk) => {
            if (isCancelled) return;

            if (chunk.type === 'start') {
              setSessionId(chunk.session_id);
              setStage('thinking');
            } else if (chunk.type === 'delta') {
              setStage('responding');
              setStreamingContent((prev) => prev + (chunk.content || ''));
            } else if (chunk.type === 'tool_call') {
              setStage('tool_running');
              // Add new tool call
              const newToolCall: ToolCallInfo = {
                id: `tool-${Date.now()}-${toolCallCounterRef.current++}`,
                name: chunk.tool_name || 'unknown',
                input: chunk.tool_input || {},
                status: 'running',
                startTime: Date.now(),
              };
              currentToolCalls = [...currentToolCalls, newToolCall];
              setActiveToolCalls([...currentToolCalls]);
            } else if (chunk.type === 'tool_result') {
              // Mark last tool call as complete
              if (currentToolCalls.length > 0) {
                const lastIdx = currentToolCalls.length - 1;
                currentToolCalls[lastIdx] = {
                  ...currentToolCalls[lastIdx],
                  status: chunk.is_error ? 'error' : 'complete',
                  endTime: Date.now(),
                };
                setActiveToolCalls([...currentToolCalls]);
              }
            } else if (chunk.type === 'delegation') {
              // Handle sub-agent delegation events
              const agentName = chunk.delegated_agent || 'Agent';
              const event = chunk.delegated_event || 'working';

              if (event === 'start') {
                setStage('delegating');
                setDelegation({ agentName, event, toolCalls: [] });
              } else if (event === 'done') {
                setDelegation(null);
                setStage('thinking'); // Back to parent agent
              } else if (event === 'tool_call') {
                // Track sub-agent tool calls
                setDelegation((prev) => {
                  if (!prev) return { agentName, event, toolCalls: [] };
                  const newToolCall: ToolCallInfo = {
                    id: `delegate-tool-${Date.now()}-${toolCallCounterRef.current++}`,
                    name: chunk.tool_name || 'unknown',
                    input: chunk.tool_input || {},
                    status: 'running',
                    startTime: Date.now(),
                  };
                  return { ...prev, event: 'tool_call', toolCalls: [...prev.toolCalls, newToolCall] };
                });
              } else if (event === 'tool_result') {
                // Mark sub-agent tool as complete
                setDelegation((prev) => {
                  if (!prev || prev.toolCalls.length === 0) return prev;
                  const lastIdx = prev.toolCalls.length - 1;
                  const updatedCalls = [...prev.toolCalls];
                  updatedCalls[lastIdx] = {
                    ...updatedCalls[lastIdx],
                    status: chunk.is_error ? 'error' : 'complete',
                    endTime: Date.now(),
                  };
                  return { ...prev, event: 'tool_result', toolCalls: updatedCalls };
                });
              } else if (event === 'text') {
                // Sub-agent is responding
                setDelegation((prev) => prev ? { ...prev, event: 'responding' } : null);
              }
            } else if (chunk.type === 'done') {
              // Save message with tool calls
              setMessages((prev) => [
                ...prev,
                {
                  role: 'assistant',
                  content: chunk.content || '',
                  toolCalls: currentToolCalls.length > 0 ? currentToolCalls : undefined,
                },
              ]);
              setStreamingContent('');
              setActiveToolCalls([]);
              setIsStreaming(false);
              setStage('idle');
              setStartTime(null);
              cleanupRef.current = null;
              resolve();
            } else if (chunk.type === 'error') {
              setIsStreaming(false);
              setActiveToolCalls([]);
              setStage('idle');
              setStartTime(null);
              cleanupRef.current = null;
              reject(new Error(chunk.error));
            }
          },
          (error) => {
            if (isCancelled) return;
            setIsStreaming(false);
            setActiveToolCalls([]);
            setStage('idle');
            setStartTime(null);
            cleanupRef.current = null;
            reject(error);
          }
        );

        // Store cleanup function
        cleanupRef.current = () => {
          isCancelled = true;
          cleanup();
        };
      });
    },
    [agentId]
  );

  const reset = useCallback(() => {
    // Cancel any ongoing stream
    if (cleanupRef.current) {
      cleanupRef.current();
      cleanupRef.current = null;
    }
    setSessionId(null);
    setMessages([]);
    setStreamingContent('');
    setActiveToolCalls([]);
    setIsStreaming(false);
    setStage('idle');
    setStartTime(null);
    setDelegation(null);
  }, []);

  return {
    sessionId,
    messages,
    isStreaming,
    streamingContent,
    activeToolCalls,
    stage,
    elapsedTime,
    delegation,
    sendMessage,
    reset,
  };
}
