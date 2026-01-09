'use client';

import { useState, useEffect, useRef, useCallback } from 'react';
import { Card, CardTitle, CardContent } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { api, connectBanterWS, BanterMessage } from '@/lib/api';
import {
  MessageSquare,
  Sparkles,
  Volume2,
  VolumeX,
  RefreshCw,
  Users,
} from 'lucide-react';

// Agent avatars/colors for visual distinction
const agentColors: Record<string, string> = {
  cereal: 'bg-yellow-500',
  razor: 'bg-red-500',
  plague: 'bg-purple-500',
  joey: 'bg-blue-500',
  acid: 'bg-pink-500',
  flushot: 'bg-green-500',
  nikon: 'bg-indigo-500',
  gill: 'bg-cyan-500',
  hal: 'bg-orange-500',
  blade: 'bg-gray-500',
  phreak: 'bg-teal-500',
  gibson: 'bg-amber-500',
};

const agentEmojis: Record<string, string> = {
  cereal: '🥣',
  razor: '🔪',
  plague: '☣️',
  joey: '🔧',
  acid: '🔥',
  flushot: '💉',
  nikon: '📷',
  gill: '🔐',
  hal: '🤖',
  blade: '📋',
  phreak: '⚖️',
  gibson: '📊',
};

interface AgentBanterProps {
  maxMessages?: number;
  autoConnect?: boolean;
  className?: string;
}

export function AgentBanter({
  maxMessages = 10,
  autoConnect = true,
  className = '',
}: AgentBanterProps) {
  const [messages, setMessages] = useState<BanterMessage[]>([]);
  const [enabled, setEnabled] = useState(false);
  const [connected, setConnected] = useState(false);
  const [loading, setLoading] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Add a new message
  const addMessage = useCallback((msg: BanterMessage) => {
    setMessages((prev) => [...prev, msg].slice(-maxMessages));
  }, [maxMessages]);

  // Connect to WebSocket
  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return;

    wsRef.current = connectBanterWS(
      (msg) => addMessage(msg),
      () => setConnected(false),
      (initialEnabled) => {
        setConnected(true);
        setEnabled(initialEnabled);
      }
    );

    wsRef.current.onclose = () => setConnected(false);
  }, [addMessage]);

  // Disconnect from WebSocket
  const disconnect = useCallback(() => {
    wsRef.current?.close();
    wsRef.current = null;
    setConnected(false);
  }, []);

  // Toggle banter on/off
  const toggleBanter = async () => {
    setLoading(true);
    try {
      await api.banter.toggle(!enabled);
      setEnabled(!enabled);
    } catch {
      // Ignore errors
    }
    setLoading(false);
  };

  // Generate on-demand banter
  const generateBanter = async () => {
    setLoading(true);
    try {
      const msg = await api.banter.generate();
      addMessage(msg);
    } catch {
      // Ignore errors
    }
    setLoading(false);
  };

  // Generate a multi-agent exchange
  const generateExchange = async () => {
    setLoading(true);
    try {
      const result = await api.banter.exchange();
      result.messages.forEach((msg) => addMessage(msg));
    } catch {
      // Ignore errors
    }
    setLoading(false);
  };

  // Auto-connect on mount
  useEffect(() => {
    if (autoConnect) {
      connect();
    }
    return () => disconnect();
  }, [autoConnect, connect, disconnect]);

  // Auto-scroll to bottom
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  return (
    <Card className={className}>
      <CardTitle className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Sparkles className="h-5 w-5 text-yellow-500" />
          Agent Banter
          {connected && (
            <Badge variant="success" className="text-xs">
              Live
            </Badge>
          )}
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={toggleBanter}
            disabled={loading}
            title={enabled ? 'Disable banter' : 'Enable banter'}
          >
            {enabled ? (
              <Volume2 className="h-4 w-4 text-green-500" />
            ) : (
              <VolumeX className="h-4 w-4 text-gray-500" />
            )}
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={generateBanter}
            disabled={loading}
            title="Generate banter"
          >
            <MessageSquare className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={generateExchange}
            disabled={loading}
            title="Generate exchange"
          >
            <Users className="h-4 w-4" />
          </Button>
        </div>
      </CardTitle>
      <CardContent className="mt-4">
        <div className="space-y-3 max-h-80 overflow-y-auto">
          {messages.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              <Sparkles className="h-8 w-8 mx-auto mb-2 opacity-50" />
              <p>No banter yet...</p>
              <p className="text-sm mt-1">
                {enabled
                  ? 'Waiting for agents to chat'
                  : 'Enable banter to see agent conversations'}
              </p>
              <Button
                variant="outline"
                size="sm"
                className="mt-4"
                onClick={generateExchange}
                disabled={loading}
              >
                <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
                Generate Exchange
              </Button>
            </div>
          ) : (
            messages.map((msg) => (
              <BanterMessageItem key={msg.id} message={msg} />
            ))
          )}
          <div ref={messagesEndRef} />
        </div>
      </CardContent>
    </Card>
  );
}

function BanterMessageItem({ message }: { message: BanterMessage }) {
  const color = agentColors[message.agent] || 'bg-gray-500';
  const emoji = agentEmojis[message.agent] || '🤖';

  return (
    <div className="flex items-start gap-3 p-2 rounded-lg bg-gray-800/30 hover:bg-gray-800/50 transition-colors">
      <div
        className={`w-8 h-8 rounded-full ${color} flex items-center justify-center text-white text-sm font-bold shrink-0`}
        title={message.agent_name}
      >
        {emoji}
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="font-medium text-white">{message.agent_name}</span>
          <Badge
            variant={
              message.type === 'pun'
                ? 'warning'
                : message.type === 'quote'
                ? 'info'
                : 'default'
            }
            className="text-xs"
          >
            {message.type}
          </Badge>
          {message.target && (
            <span className="text-xs text-gray-500">
              → {message.target}
            </span>
          )}
        </div>
        <p className="text-sm text-gray-300 mt-1">{message.message}</p>
      </div>
    </div>
  );
}

// Compact version for sidebar/footer
export function AgentBanterCompact({ className = '' }: { className?: string }) {
  const [message, setMessage] = useState<BanterMessage | null>(null);
  const [loading, setLoading] = useState(false);

  const refresh = async () => {
    setLoading(true);
    try {
      const msg = await api.banter.generate();
      setMessage(msg);
    } catch {
      // Ignore errors
    }
    setLoading(false);
  };

  useEffect(() => {
    refresh();
    const interval = setInterval(refresh, 60000); // Refresh every minute
    return () => clearInterval(interval);
  }, []);

  if (!message) return null;

  const emoji = agentEmojis[message.agent] || '🤖';

  return (
    <div
      className={`flex items-center gap-2 p-2 rounded-lg bg-gray-800/30 text-sm ${className}`}
    >
      <span>{emoji}</span>
      <span className="font-medium text-gray-400">{message.agent_name}:</span>
      <span className="text-gray-300 truncate flex-1">{message.message}</span>
      <Button
        variant="ghost"
        size="sm"
        onClick={refresh}
        disabled={loading}
        className="shrink-0"
      >
        <RefreshCw className={`h-3 w-3 ${loading ? 'animate-spin' : ''}`} />
      </Button>
    </div>
  );
}
