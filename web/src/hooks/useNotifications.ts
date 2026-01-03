'use client';

import { useEffect, useState, useCallback } from 'react';

type NotificationPermission = 'default' | 'granted' | 'denied';

export function useNotifications() {
  const [permission, setPermission] = useState<NotificationPermission>('default');
  const [supported, setSupported] = useState(false);

  useEffect(() => {
    if (typeof window !== 'undefined' && 'Notification' in window) {
      setSupported(true);
      setPermission(Notification.permission);
    }
  }, []);

  const requestPermission = useCallback(async () => {
    if (!supported) return false;

    try {
      const result = await Notification.requestPermission();
      setPermission(result);
      return result === 'granted';
    } catch {
      return false;
    }
  }, [supported]);

  const notify = useCallback(
    (title: string, options?: NotificationOptions) => {
      if (!supported || permission !== 'granted') return null;

      try {
        const notification = new Notification(title, {
          icon: '/favicon.ico',
          badge: '/favicon.ico',
          ...options,
        });

        // Auto-close after 5 seconds
        setTimeout(() => notification.close(), 5000);

        return notification;
      } catch {
        return null;
      }
    },
    [supported, permission]
  );

  const notifyScanComplete = useCallback(
    (target: string, success: boolean, projectIds?: string[]) => {
      if (success) {
        notify(`Scan Complete: ${target}`, {
          body: projectIds?.length
            ? `Scanned ${projectIds.length} project(s)`
            : 'Scan completed successfully',
          tag: `scan-${target}`,
        });
      } else {
        notify(`Scan Failed: ${target}`, {
          body: 'The scan encountered an error',
          tag: `scan-${target}`,
        });
      }
    },
    [notify]
  );

  const notifyScanStarted = useCallback(
    (target: string) => {
      notify(`Scan Started: ${target}`, {
        body: 'Your scan is now running',
        tag: `scan-${target}`,
      });
    },
    [notify]
  );

  return {
    supported,
    permission,
    requestPermission,
    notify,
    notifyScanComplete,
    notifyScanStarted,
  };
}
