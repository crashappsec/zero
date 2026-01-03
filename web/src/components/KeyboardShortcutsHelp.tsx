'use client';

import { useState, useEffect } from 'react';
import { X, Keyboard } from 'lucide-react';

interface Shortcut {
  key: string;
  description: string;
}

const shortcuts: Shortcut[] = [
  { key: 'g', description: 'Go to Dashboard' },
  { key: 'p', description: 'Go to Projects' },
  { key: 's', description: 'Go to Scans' },
  { key: 'r', description: 'Go to Reports' },
  { key: ',', description: 'Go to Settings' },
  { key: 'n', description: 'New Scan' },
  { key: '/', description: 'Focus Search' },
  { key: 'Esc', description: 'Clear/Close' },
  { key: '?', description: 'Show Shortcuts' },
];

export function KeyboardShortcutsHelp() {
  const [isOpen, setIsOpen] = useState(false);

  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      // Ignore if typing in input
      const target = event.target as HTMLElement;
      if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA') return;

      if (event.key === '?' && !event.ctrlKey && !event.metaKey) {
        event.preventDefault();
        setIsOpen(prev => !prev);
      }

      if (event.key === 'Escape') {
        setIsOpen(false);
      }
    }

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="relative w-full max-w-md rounded-lg bg-gray-900 border border-gray-700 shadow-xl">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-gray-700 px-6 py-4">
          <div className="flex items-center gap-2">
            <Keyboard className="h-5 w-5 text-blue-500" />
            <h2 className="text-lg font-semibold text-white">Keyboard Shortcuts</h2>
          </div>
          <button
            onClick={() => setIsOpen(false)}
            className="rounded-md p-1 text-gray-400 hover:bg-gray-800 hover:text-white"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Content */}
        <div className="px-6 py-4">
          <div className="space-y-2">
            {shortcuts.map((shortcut) => (
              <div
                key={shortcut.key}
                className="flex items-center justify-between py-1"
              >
                <span className="text-gray-300">{shortcut.description}</span>
                <kbd className="rounded bg-gray-800 px-2 py-1 text-xs font-mono text-gray-400 border border-gray-700">
                  {shortcut.key}
                </kbd>
              </div>
            ))}
          </div>
        </div>

        {/* Footer */}
        <div className="border-t border-gray-700 px-6 py-3">
          <p className="text-xs text-gray-500 text-center">
            Press <kbd className="rounded bg-gray-800 px-1 py-0.5 text-xs font-mono border border-gray-700">?</kbd> to toggle this help
          </p>
        </div>
      </div>
    </div>
  );
}
