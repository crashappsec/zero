'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { cn } from '@/lib/utils';
import {
  LayoutDashboard,
  Terminal,
  Settings,
  GitFork,
  Bot,
  Package,
  Shield,
  Key,
  BarChart3,
  Server,
  Cpu,
  Users,
  Sparkles,
  ChevronDown,
  Scan,
} from 'lucide-react';
import { useState } from 'react';

interface NavItem {
  name: string;
  href: string;
  icon: React.ComponentType<{ className?: string }>;
}

interface NavGroup {
  title: string;
  items: NavItem[];
  defaultOpen?: boolean;
}

// Scanner navigation - aligned with existing pages
const scannerNavigation: NavItem[] = [
  { name: 'Dependencies', href: '/dependencies', icon: Package },
  { name: 'Vulnerabilities', href: '/vulnerabilities', icon: Shield },
  { name: 'Secrets', href: '/secrets', icon: Key },
  { name: 'Code Quality', href: '/quality', icon: BarChart3 },
  { name: 'DevOps', href: '/devops', icon: Server },
  { name: 'Technology', href: '/technology', icon: Cpu },
  { name: 'Code Ownership', href: '/ownership', icon: Users },
  { name: 'Developer Experience', href: '/devx', icon: Sparkles },
];

// Management navigation
const managementNavigation: NavItem[] = [
  { name: 'Repos', href: '/repos', icon: GitFork },
  { name: 'Scans', href: '/scans', icon: Scan },
  { name: 'Chat', href: '/chat', icon: Bot },
  { name: 'Settings', href: '/settings', icon: Settings },
];

const navigationGroups: NavGroup[] = [
  {
    title: 'Overview',
    items: [{ name: 'Dashboard', href: '/', icon: LayoutDashboard }],
    defaultOpen: true,
  },
  {
    title: 'Scanners',
    items: scannerNavigation,
    defaultOpen: true,
  },
  {
    title: 'Management',
    items: managementNavigation,
    defaultOpen: true,
  },
];

function NavSection({ group }: { group: NavGroup }) {
  const pathname = usePathname();
  const [isOpen, setIsOpen] = useState(group.defaultOpen ?? true);

  // Check if any item in this group is active
  const hasActiveItem = group.items.some(
    (item) => pathname === item.href || pathname.startsWith(item.href + '/')
  );

  return (
    <div className="mb-4">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex w-full items-center justify-between px-3 py-1.5 text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300"
      >
        {group.title}
        <ChevronDown
          className={cn(
            'h-4 w-4 transition-transform',
            isOpen ? 'rotate-0' : '-rotate-90'
          )}
        />
      </button>
      {isOpen && (
        <div className="mt-1 space-y-0.5">
          {group.items.map((item) => {
            const isActive =
              pathname === item.href || pathname.startsWith(item.href + '/');
            return (
              <Link
                key={item.name}
                href={item.href}
                className={cn(
                  'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                  isActive
                    ? 'bg-green-600/10 text-green-600 dark:text-green-500'
                    : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-gray-800 dark:hover:text-white'
                )}
              >
                <item.icon className="h-4 w-4" />
                {item.name}
              </Link>
            );
          })}
        </div>
      )}
    </div>
  );
}

export function Sidebar() {
  return (
    <aside className="hidden lg:fixed lg:inset-y-0 lg:left-0 lg:z-50 lg:flex lg:w-64 lg:flex-col border-r border-gray-200 bg-white dark:border-gray-800 dark:bg-gray-900">
      {/* Logo */}
      <div className="flex h-16 items-center gap-3 border-b border-gray-200 px-6 dark:border-gray-800">
        <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-green-600">
          <Terminal className="h-5 w-5 text-white" />
        </div>
        <div>
          <h1 className="text-lg font-bold text-gray-900 dark:text-white">Zero</h1>
          <p className="text-xs text-gray-500">Developer Intelligence</p>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto px-3 py-4">
        {navigationGroups.map((group) => (
          <NavSection key={group.title} group={group} />
        ))}
      </nav>

      {/* Footer */}
      <div className="border-t border-gray-200 p-4 dark:border-gray-800">
        <div className="flex items-center gap-3 text-sm text-gray-500">
          <Terminal className="h-4 w-4" />
          <span>v0.1.0-experimental</span>
        </div>
      </div>
    </aside>
  );
}

import { MobileNav } from './MobileNav';

export function MainLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen">
      <MobileNav />
      <Sidebar />
      <main className="lg:pl-64">
        <div className="container mx-auto max-w-7xl p-4 lg:p-6">{children}</div>
      </main>
    </div>
  );
}
