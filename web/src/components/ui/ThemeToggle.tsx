'use client';

import { Moon, Sun, Monitor } from 'lucide-react';
import { useTheme } from '@/hooks/useTheme';

export function ThemeToggle() {
  const { theme, setTheme } = useTheme();

  return (
    <div className="flex items-center gap-1 rounded-lg border border-gray-700 p-1">
      <button
        onClick={() => setTheme('light')}
        className={`rounded-md p-1.5 transition-colors ${
          theme === 'light'
            ? 'bg-gray-700 text-white'
            : 'text-gray-400 hover:text-white hover:bg-gray-800'
        }`}
        title="Light mode"
        aria-label="Light mode"
      >
        <Sun className="h-4 w-4" />
      </button>
      <button
        onClick={() => setTheme('dark')}
        className={`rounded-md p-1.5 transition-colors ${
          theme === 'dark'
            ? 'bg-gray-700 text-white'
            : 'text-gray-400 hover:text-white hover:bg-gray-800'
        }`}
        title="Dark mode"
        aria-label="Dark mode"
      >
        <Moon className="h-4 w-4" />
      </button>
      <button
        onClick={() => setTheme('system')}
        className={`rounded-md p-1.5 transition-colors ${
          theme === 'system'
            ? 'bg-gray-700 text-white'
            : 'text-gray-400 hover:text-white hover:bg-gray-800'
        }`}
        title="System preference"
        aria-label="System preference"
      >
        <Monitor className="h-4 w-4" />
      </button>
    </div>
  );
}
