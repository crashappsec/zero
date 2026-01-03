'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { cn } from '@/lib/utils';
import {
  LayoutDashboard,
  FolderKanban,
  Scan,
  MessageSquare,
  Terminal,
  AlertTriangle,
  Package,
  Key,
  Settings,
  Users,
  Server,
  Cpu,
  BarChart3,
  Sparkles,
} from 'lucide-react';

const navigation = [
  { name: 'Dashboard', href: '/', icon: LayoutDashboard },
  { name: 'Projects', href: '/projects', icon: FolderKanban },
  { name: 'Scans', href: '/scans', icon: Scan },
  { name: 'Dependencies', href: '/dependencies', icon: Package },
  { name: 'Vulnerabilities', href: '/vulnerabilities', icon: AlertTriangle },
  { name: 'Secrets', href: '/secrets', icon: Key },
  { name: 'Code Ownership', href: '/ownership', icon: Users },
  { name: 'Technology', href: '/technology', icon: Cpu },
  { name: 'DevOps', href: '/devops', icon: Server },
  { name: 'Code Quality', href: '/quality', icon: BarChart3 },
  { name: 'Developer Experience', href: '/devx', icon: Sparkles },
  { name: 'Agent Chat', href: '/chat', icon: MessageSquare },
  { name: 'Settings', href: '/settings', icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="hidden lg:fixed lg:inset-y-0 lg:left-0 lg:z-50 lg:flex lg:w-64 lg:flex-col border-r border-gray-800 bg-gray-900">
      {/* Logo */}
      <div className="flex h-16 items-center gap-3 border-b border-gray-800 px-6">
        <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-green-600">
          <Terminal className="h-5 w-5 text-white" />
        </div>
        <div>
          <h1 className="text-lg font-bold text-white">Zero</h1>
          <p className="text-xs text-gray-500">Developer Intelligence</p>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 space-y-1 px-3 py-4">
        {navigation.map((item) => {
          const isActive = pathname === item.href || pathname.startsWith(item.href + '/');
          return (
            <Link
              key={item.name}
              href={item.href}
              className={cn(
                'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                isActive
                  ? 'bg-green-600/10 text-green-500'
                  : 'text-gray-400 hover:bg-gray-800 hover:text-white'
              )}
            >
              <item.icon className="h-5 w-5" />
              {item.name}
            </Link>
          );
        })}
      </nav>

      {/* Footer */}
      <div className="border-t border-gray-800 p-4">
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
