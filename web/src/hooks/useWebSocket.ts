'use client';

import { useEffect, useRef, useState, useCallback } from 'react';

type WebSocketStatus = 'connecting' | 'connected' | 'disconnected' | 'error';

interface UseWebSocketOptions {
  onMessage?: (data: unknown) => void;
  onError?: (error: Event) => void;
  onOpen?: () => void;
  onClose?: () => void;
  reconnect?: boolean;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
}

export function useWebSocket(url: string | null, options: UseWebSocketOptions = {}) {
  const {
    onMessage,
    onError,
    onOpen,
    onClose,
    reconnect = true,
    reconnectInterval = 3000,
    maxReconnectAttempts = 5,
  } = options;

  const [status, setStatus] = useState<WebSocketStatus>('disconnected');
  const [lastMessage, setLastMessage] = useState<unknown>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const connect = useCallback(() => {
    if (!url) return;

    try {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = url.startsWith('ws') ? url : `${protocol}//${window.location.host}${url}`;

      setStatus('connecting');
      const ws = new WebSocket(wsUrl);

      ws.onopen = () => {
        setStatus('connected');
        reconnectAttemptsRef.current = 0;
        onOpen?.();
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          setLastMessage(data);
          onMessage?.(data);
        } catch {
          // Handle non-JSON messages
          setLastMessage(event.data);
          onMessage?.(event.data);
        }
      };

      ws.onerror = (error) => {
        setStatus('error');
        onError?.(error);
      };

      ws.onclose = () => {
        setStatus('disconnected');
        onClose?.();

        // Attempt reconnect
        if (reconnect && reconnectAttemptsRef.current < maxReconnectAttempts) {
          reconnectAttemptsRef.current++;
          reconnectTimeoutRef.current = setTimeout(() => {
            connect();
          }, reconnectInterval);
        }
      };

      wsRef.current = ws;
    } catch (error) {
      setStatus('error');
    }
  }, [url, onMessage, onError, onOpen, onClose, reconnect, reconnectInterval, maxReconnectAttempts]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    setStatus('disconnected');
  }, []);

  const send = useCallback((data: unknown) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(typeof data === 'string' ? data : JSON.stringify(data));
    }
  }, []);

  useEffect(() => {
    if (url) {
      connect();
    }

    return () => {
      disconnect();
    };
  }, [url, connect, disconnect]);

  return {
    status,
    lastMessage,
    send,
    connect,
    disconnect,
  };
}

// Hook for scan-specific WebSocket
export function useScanWebSocket(jobId: string | null, onUpdate?: (data: ScanUpdate) => void) {
  const [scanData, setScanData] = useState<ScanUpdate | null>(null);

  const { status, lastMessage } = useWebSocket(
    jobId ? `/ws/scan/${jobId}` : null,
    {
      onMessage: (data) => {
        const update = data as ScanUpdate;
        setScanData(update);
        onUpdate?.(update);
      },
    }
  );

  return { status, scanData, lastMessage };
}

// Types for scan updates
export interface ScanUpdate {
  type: 'job_status' | 'scanner_progress' | 'scan_complete';
  job_id: string;
  status?: string;
  scanner?: string;
  summary?: string;
  duration?: number;
  success?: boolean;
  project_ids?: string[];
  error?: string;
}

// Hook for global scan events (watches all scans)
export function useGlobalScanEvents(onEvent?: (event: ScanUpdate) => void) {
  const [events, setEvents] = useState<ScanUpdate[]>([]);

  const { status } = useWebSocket('/ws/events', {
    onMessage: (data) => {
      const event = data as ScanUpdate;
      setEvents((prev) => [event, ...prev].slice(0, 50)); // Keep last 50 events
      onEvent?.(event);
    },
  });

  const clearEvents = useCallback(() => {
    setEvents([]);
  }, []);

  return { status, events, clearEvents };
}
