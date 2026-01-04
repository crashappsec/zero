'use client';

import { useState, useRef, useEffect, Suspense } from 'react';
import { useSearchParams } from 'next/navigation';
import { MainLayout } from '@/components/layout/Sidebar';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { useChat, useAgents, useFetch } from '@/hooks/useApi';
import { api } from '@/lib/api';
import type { AgentInfo } from '@/lib/types';
import {
  Send,
  Bot,
  User,
  Trash2,
  ChevronDown,
  Terminal,
  Shield,
  Package,
  Lock,
  Scale,
  Code,
  Server,
  Cpu,
  Wrench,
  Cloud,
  BarChart3,
  Key,
  Brain,
} from 'lucide-react';

const agentIcons: Record<string, React.ReactNode> = {
  zero: <Terminal className="h-5 w-5" />,
  cereal: <Package className="h-5 w-5" />,
  razor: <Shield className="h-5 w-5" />,
  blade: <Scale className="h-5 w-5" />,
  phreak: <Scale className="h-5 w-5" />,
  acid: <Code className="h-5 w-5" />,
  dade: <Server className="h-5 w-5" />,
  nikon: <Cpu className="h-5 w-5" />,
  joey: <Wrench className="h-5 w-5" />,
  plague: <Cloud className="h-5 w-5" />,
  gibson: <BarChart3 className="h-5 w-5" />,
  gill: <Key className="h-5 w-5" />,
  turing: <Brain className="h-5 w-5" />,
};

function AgentSelector({
  agents,
  selected,
  onSelect,
}: {
  agents: AgentInfo[];
  selected: string;
  onSelect: (id: string) => void;
}) {
  const [open, setOpen] = useState(false);
  const selectedAgent = agents.find((a) => a.id === selected) || agents[0];

  return (
    <div className="relative">
      <button
        onClick={() => setOpen(!open)}
        className="flex items-center gap-2 rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-left hover:border-gray-600 transition-colors"
      >
        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-green-600/20 text-green-500">
          {agentIcons[selectedAgent?.id] || <Bot className="h-5 w-5" />}
        </div>
        <div className="flex-1">
          <p className="font-medium text-white">{selectedAgent?.name}</p>
          <p className="text-xs text-gray-500">{selectedAgent?.persona}</p>
        </div>
        <ChevronDown className="h-4 w-4 text-gray-400" />
      </button>

      {open && (
        <>
          <div className="fixed inset-0 z-10" onClick={() => setOpen(false)} />
          <div className="absolute left-0 top-full z-20 mt-1 w-72 rounded-lg border border-gray-700 bg-gray-800 shadow-xl max-h-96 overflow-y-auto">
            {agents.map((agent) => (
              <button
                key={agent.id}
                onClick={() => {
                  onSelect(agent.id);
                  setOpen(false);
                }}
                className={`w-full flex items-center gap-3 px-3 py-2 text-left hover:bg-gray-700 transition-colors ${
                  agent.id === selected ? 'bg-gray-700' : ''
                }`}
              >
                <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-green-600/20 text-green-500">
                  {agentIcons[agent.id] || <Bot className="h-5 w-5" />}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="font-medium text-white">{agent.name}</p>
                  <p className="text-xs text-gray-500 truncate">{agent.description}</p>
                </div>
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  );
}

function ChatMessage({
  role,
  content,
  agentName,
}: {
  role: 'user' | 'assistant';
  content: string;
  agentName?: string;
}) {
  const isUser = role === 'user';

  return (
    <div className={`flex gap-3 ${isUser ? 'flex-row-reverse' : ''}`}>
      <div
        className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-lg ${
          isUser ? 'bg-blue-600' : 'bg-green-600'
        }`}
      >
        {isUser ? <User className="h-4 w-4 text-white" /> : <Bot className="h-4 w-4 text-white" />}
      </div>
      <div
        className={`max-w-[80%] rounded-lg border px-4 py-3 ${
          isUser
            ? 'bg-blue-900/30 border-blue-700'
            : 'bg-gray-800/50 border-gray-700'
        }`}
      >
        {!isUser && agentName && (
          <p className="text-xs text-green-500 mb-1">{agentName}</p>
        )}
        <div className="text-sm text-gray-200 whitespace-pre-wrap">{content}</div>
      </div>
    </div>
  );
}

function ChatPageContent({ projectId }: { projectId?: string }) {
  const [agentId, setAgentId] = useState('zero');
  const [input, setInput] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const { data: agentsData } = useAgents();
  const agents = agentsData || [];

  const {
    messages,
    isStreaming,
    streamingContent,
    sendMessage,
    reset,
  } = useChat(agentId);

  const selectedAgent = agents.find((a) => a.id === agentId);

  // Auto-scroll to bottom
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, streamingContent]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim() || isStreaming) return;

    const message = input.trim();
    setInput('');

    try {
      await sendMessage(message, projectId);
    } catch (err) {
      console.error('Chat error:', err);
    }
  };

  const handleAgentChange = (newAgentId: string) => {
    if (newAgentId !== agentId) {
      setAgentId(newAgentId);
      reset(); // Clear conversation when switching agents
    }
  };

  return (
      <div className="flex h-[calc(100vh-6rem)] flex-col">
        {/* Header */}
        <div className="flex items-center justify-between pb-4">
          <div className="flex items-center gap-4">
            <AgentSelector
              agents={agents}
              selected={agentId}
              onSelect={handleAgentChange}
            />
            {projectId && (
              <Badge variant="info">
                Context: {projectId}
              </Badge>
            )}
          </div>
          {messages.length > 0 && (
            <Button variant="ghost" size="sm" onClick={reset} icon={<Trash2 className="h-4 w-4" />}>
              Clear
            </Button>
          )}
        </div>

        {/* Chat Area */}
        <Card className="flex-1 flex flex-col overflow-hidden">
          {/* Messages */}
          <div className="flex-1 overflow-y-auto p-4 space-y-4">
            {messages.length === 0 && !isStreaming ? (
              <div className="h-full flex flex-col items-center justify-center text-center">
                <div className="flex h-16 w-16 items-center justify-center rounded-full bg-green-600/20 mb-4">
                  {agentIcons[agentId] || <Bot className="h-8 w-8 text-green-500" />}
                </div>
                <h3 className="text-lg font-medium text-white">
                  Chat with {selectedAgent?.name || 'Zero'}
                </h3>
                <p className="mt-1 text-gray-400 max-w-md">
                  {selectedAgent?.description || 'Ask about security analysis, vulnerabilities, or get help with your projects.'}
                </p>
                {projectId && (
                  <p className="mt-4 text-sm text-gray-500">
                    Project context: <span className="text-green-400">{projectId}</span>
                  </p>
                )}
                <div className="mt-6 grid gap-2 text-sm">
                  <button
                    onClick={() => setInput('What security issues should I focus on?')}
                    className="rounded-lg border border-gray-700 px-4 py-2 text-gray-400 hover:border-gray-600 hover:text-white transition-colors"
                  >
                    What security issues should I focus on?
                  </button>
                  <button
                    onClick={() => setInput('Summarize the critical vulnerabilities')}
                    className="rounded-lg border border-gray-700 px-4 py-2 text-gray-400 hover:border-gray-600 hover:text-white transition-colors"
                  >
                    Summarize the critical vulnerabilities
                  </button>
                  <button
                    onClick={() => setInput('Check for supply chain risks')}
                    className="rounded-lg border border-gray-700 px-4 py-2 text-gray-400 hover:border-gray-600 hover:text-white transition-colors"
                  >
                    Check for supply chain risks
                  </button>
                </div>
              </div>
            ) : (
              <>
                {messages.map((msg, i) => (
                  <ChatMessage
                    key={i}
                    role={msg.role}
                    content={msg.content}
                    agentName={msg.role === 'assistant' ? selectedAgent?.name : undefined}
                  />
                ))}
                {isStreaming && streamingContent && (
                  <ChatMessage
                    role="assistant"
                    content={streamingContent}
                    agentName={selectedAgent?.name}
                  />
                )}
                {isStreaming && !streamingContent && (
                  <div className="flex gap-3">
                    <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-green-600">
                      <Bot className="h-4 w-4 text-white animate-pulse" />
                    </div>
                    <div className="flex items-center gap-1 text-gray-400">
                      <span className="animate-bounce">.</span>
                      <span className="animate-bounce" style={{ animationDelay: '0.1s' }}>.</span>
                      <span className="animate-bounce" style={{ animationDelay: '0.2s' }}>.</span>
                    </div>
                  </div>
                )}
                <div ref={messagesEndRef} />
              </>
            )}
          </div>

          {/* Input */}
          <form onSubmit={handleSubmit} className="border-t border-gray-700 p-4">
            <div className="flex gap-2">
              <input
                type="text"
                value={input}
                onChange={(e) => setInput(e.target.value)}
                placeholder={`Ask ${selectedAgent?.name || 'Zero'} a question...`}
                disabled={isStreaming}
                className="flex-1 rounded-lg border border-gray-700 bg-gray-800 px-4 py-2 text-white placeholder:text-gray-500 focus:border-green-500 focus:outline-none focus:ring-1 focus:ring-green-500 disabled:opacity-50"
              />
              <Button type="submit" disabled={isStreaming || !input.trim()}>
                <Send className="h-4 w-4" />
              </Button>
            </div>
            {!api ? (
              <p className="mt-2 text-xs text-yellow-500">
                Note: ANTHROPIC_API_KEY required for chat functionality
              </p>
            ) : null}
          </form>
        </Card>
      </div>
  );
}

function ChatWithProjectId() {
  const searchParams = useSearchParams();
  const projectId = searchParams.get('project') || undefined;
  return <ChatPageContent projectId={projectId} />;
}

export default function ChatPage() {
  return (
    <MainLayout>
      <Suspense fallback={<div className="animate-pulse h-96 bg-gray-800 rounded-lg" />}>
        <ChatWithProjectId />
      </Suspense>
    </MainLayout>
  );
}
