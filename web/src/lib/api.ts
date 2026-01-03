import type {
  Project,
  ScannerInfo,
  AgentInfo,
  ProfileInfo,
  ScanJob,
  QueueStats,
  ChatSession,
  AnalysisSummary,
  Vulnerability,
  Secret,
  Dependency,
  ListResponse,
  HealthResponse,
  StreamChunk,
} from './types';

const API_BASE = '/api';

async function fetchJSON<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(error.error || error.details || 'Request failed');
  }

  return res.json();
}

// System endpoints
export const api = {
  // Health & config
  health: () => fetchJSON<HealthResponse>('/health'),
  config: () => fetchJSON<Record<string, unknown>>('/config'),

  // Lists
  scanners: () => fetchJSON<ListResponse<ScannerInfo>>('/scanners'),
  agents: () => fetchJSON<ListResponse<AgentInfo>>('/agents'),
  profiles: () => fetchJSON<ListResponse<ProfileInfo>>('/profiles'),

  // Projects
  projects: {
    list: () => fetchJSON<ListResponse<Project>>('/projects'),
    get: (id: string) => fetchJSON<Project>(`/projects/${encodeURIComponent(id)}`),
    delete: (id: string) => fetchJSON<void>(`/projects/${encodeURIComponent(id)}`, { method: 'DELETE' }),
    freshness: (id: string) => fetchJSON<{ freshness: string }>(`/projects/${encodeURIComponent(id)}/freshness`),
  },

  // Analysis
  analysis: {
    summary: (projectId: string) =>
      fetchJSON<AnalysisSummary>(`/analysis/${encodeURIComponent(projectId)}/summary`),
    vulnerabilities: (projectId: string) =>
      fetchJSON<ListResponse<Vulnerability>>(`/analysis/${encodeURIComponent(projectId)}/vulnerabilities`),
    secrets: (projectId: string) =>
      fetchJSON<ListResponse<Secret>>(`/analysis/${encodeURIComponent(projectId)}/secrets`),
    dependencies: (projectId: string) =>
      fetchJSON<ListResponse<Dependency>>(`/analysis/${encodeURIComponent(projectId)}/dependencies`),
    raw: (projectId: string, type: string) =>
      fetchJSON<unknown>(`/projects/${encodeURIComponent(projectId)}/analysis/${type}`),
  },

  // Scans
  scans: {
    start: (target: string, profile = 'standard', options?: { force?: boolean; depth?: number }) =>
      fetchJSON<{ job_id: string; ws_endpoint: string }>('/scans', {
        method: 'POST',
        body: JSON.stringify({ target, profile, ...options }),
      }),
    get: (jobId: string) => fetchJSON<ScanJob>(`/scans/${jobId}`),
    cancel: (jobId: string) => fetchJSON<void>(`/scans/${jobId}`, { method: 'DELETE' }),
    active: () => fetchJSON<ListResponse<ScanJob>>('/scans/active'),
    history: () => fetchJSON<ListResponse<ScanJob>>('/scans/history'),
    stats: () => fetchJSON<QueueStats>('/scans/stats'),
  },

  // Chat
  chat: {
    send: (message: string, options?: { session_id?: string; agent_id?: string; project_id?: string }) =>
      fetchJSON<{ session_id: string; agent_id: string; response: string }>('/chat', {
        method: 'POST',
        body: JSON.stringify({ message, ...options }),
      }),
    sessions: () => fetchJSON<ListResponse<ChatSession>>('/chat/sessions'),
    session: (id: string) => fetchJSON<ChatSession>(`/chat/sessions/${id}`),
    deleteSession: (id: string) => fetchJSON<void>(`/chat/sessions/${id}`, { method: 'DELETE' }),
  },
};

// Streaming chat via SSE
export function streamChat(
  message: string,
  options: { session_id?: string; agent_id?: string; project_id?: string },
  onChunk: (chunk: StreamChunk) => void,
  onError: (error: Error) => void
): () => void {
  const controller = new AbortController();

  fetch(`${API_BASE}/chat/stream`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ message, ...options }),
    signal: controller.signal,
  })
    .then(async (res) => {
      if (!res.ok) {
        throw new Error('Stream request failed');
      }

      const reader = res.body?.getReader();
      if (!reader) throw new Error('No response body');

      const decoder = new TextDecoder();
      let buffer = '';

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            try {
              const chunk = JSON.parse(line.slice(6)) as StreamChunk;
              onChunk(chunk);
            } catch {
              // Ignore parse errors
            }
          }
        }
      }
    })
    .catch((err) => {
      if (err.name !== 'AbortError') {
        onError(err);
      }
    });

  return () => controller.abort();
}

// WebSocket for scan progress
export function connectScanWS(
  jobId: string,
  onMessage: (msg: unknown) => void,
  onError: (error: Event) => void
): WebSocket {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const ws = new WebSocket(`${protocol}//${window.location.host}/ws/scan/${jobId}`);

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);
      onMessage(data);
    } catch {
      // Ignore parse errors
    }
  };

  ws.onerror = onError;

  return ws;
}

// WebSocket for agent chat
export function connectAgentWS(
  options: { session?: string; agent?: string },
  onMessage: (chunk: StreamChunk) => void,
  onError: (error: Event) => void
): WebSocket {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const params = new URLSearchParams();
  if (options.session) params.set('session', options.session);
  if (options.agent) params.set('agent', options.agent);

  const ws = new WebSocket(`${protocol}//${window.location.host}/ws/agent?${params}`);

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data) as StreamChunk;
      onMessage(data);
    } catch {
      // Ignore parse errors
    }
  };

  ws.onerror = onError;

  return ws;
}
