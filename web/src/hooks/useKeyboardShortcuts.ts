'use client';

import { useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';

interface ShortcutConfig {
  key: string;
  ctrl?: boolean;
  shift?: boolean;
  meta?: boolean;
  alt?: boolean;
  action: () => void;
  description: string;
}

const shortcuts: ShortcutConfig[] = [];

export function useKeyboardShortcuts() {
  const router = useRouter();

  // Navigation shortcuts
  const navigationShortcuts: ShortcutConfig[] = [
    { key: 'g', description: 'Go to Dashboard', action: () => router.push('/') },
    { key: 'p', description: 'Go to Projects', action: () => router.push('/projects') },
    { key: 's', description: 'Go to Scans', action: () => router.push('/scans') },
    { key: 'r', description: 'Go to Reports', action: () => router.push('/reports') },
    { key: ',', description: 'Go to Settings', action: () => router.push('/settings') },
  ];

  // Action shortcuts
  const actionShortcuts: ShortcutConfig[] = [
    { key: 'n', description: 'New Scan', action: () => router.push('/scans?new=true') },
    { key: '/', description: 'Focus Search', action: () => {
      const searchInput = document.querySelector('[data-search-input]') as HTMLInputElement;
      searchInput?.focus();
    }},
    { key: 'Escape', description: 'Clear/Close', action: () => {
      const searchInput = document.querySelector('[data-search-input]') as HTMLInputElement;
      if (document.activeElement === searchInput) {
        searchInput.blur();
      }
    }},
  ];

  const allShortcuts = [...navigationShortcuts, ...actionShortcuts];

  const handleKeyDown = useCallback((event: KeyboardEvent) => {
    // Ignore if typing in input/textarea
    const target = event.target as HTMLElement;
    if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable) {
      // Only allow Escape in inputs
      if (event.key !== 'Escape') return;
    }

    // Ignore if modifier keys are pressed (except for specific shortcuts)
    if (event.ctrlKey || event.metaKey || event.altKey) return;

    const shortcut = allShortcuts.find(s => s.key === event.key);
    if (shortcut) {
      event.preventDefault();
      shortcut.action();
    }
  }, [allShortcuts]);

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

  return { shortcuts: allShortcuts };
}

// Hook to show keyboard shortcuts help
export function useShortcutsHelp() {
  const { shortcuts } = useKeyboardShortcuts();

  return shortcuts.map(s => ({
    key: s.key === ',' ? 'comma' : s.key,
    description: s.description,
  }));
}
