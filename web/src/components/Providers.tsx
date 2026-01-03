'use client';

import { ReactNode } from 'react';
import { ToastProvider } from '@/components/ui/Toast';
import { KeyboardShortcutsHelp } from '@/components/KeyboardShortcutsHelp';
import { useKeyboardShortcuts } from '@/hooks/useKeyboardShortcuts';

function KeyboardShortcutsProvider({ children }: { children: ReactNode }) {
  useKeyboardShortcuts();
  return <>{children}</>;
}

export function Providers({ children }: { children: ReactNode }) {
  return (
    <ToastProvider>
      <KeyboardShortcutsProvider>
        {children}
        <KeyboardShortcutsHelp />
      </KeyboardShortcutsProvider>
    </ToastProvider>
  );
}
