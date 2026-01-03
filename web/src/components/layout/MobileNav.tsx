'use client';

import { useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  Menu,
  X,
  Home,
  FolderOpen,
  Scan,
  AlertTriangle,
  Key,
  Package,
  Settings,
  MessageSquare,
  Users,
  Server,
  Cpu,
  BarChart3,
  Sparkles,
} from 'lucide-react';

const navItems = [
  { href: '/', label: 'Dashboard', icon: Home },
  { href: '/projects', label: 'Projects', icon: FolderOpen },
  { href: '/scans', label: 'Scans', icon: Scan },
  { href: '/dependencies', label: 'Dependencies', icon: Package },
  { href: '/vulnerabilities', label: 'Vulnerabilities', icon: AlertTriangle },
  { href: '/secrets', label: 'Secrets', icon: Key },
  { href: '/ownership', label: 'Code Ownership', icon: Users },
  { href: '/technology', label: 'Technology', icon: Cpu },
  { href: '/devops', label: 'DevOps', icon: Server },
  { href: '/quality', label: 'Code Quality', icon: BarChart3 },
  { href: '/devx', label: 'Developer Experience', icon: Sparkles },
  { href: '/chat', label: 'Chat', icon: MessageSquare },
  { href: '/settings', label: 'Settings', icon: Settings },
];

export function MobileNav() {
  const [isOpen, setIsOpen] = useState(false);
  const pathname = usePathname();

  return (
    <div className="lg:hidden">
      {/* Mobile header */}
      <div className="fixed top-0 left-0 right-0 z-50 flex items-center justify-between border-b border-gray-700 bg-gray-900 px-4 py-3">
        <Link href="/" className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-green-600">
            <span className="text-lg font-bold text-white">Z</span>
          </div>
          <span className="font-semibold text-white">Zero</span>
        </Link>
        <button
          onClick={() => setIsOpen(!isOpen)}
          className="rounded-md p-2 text-gray-400 hover:bg-gray-800 hover:text-white"
          aria-label="Toggle menu"
        >
          {isOpen ? <X className="h-6 w-6" /> : <Menu className="h-6 w-6" />}
        </button>
      </div>

      {/* Mobile menu overlay */}
      {isOpen && (
        <div className="fixed inset-0 z-40 bg-black/50" onClick={() => setIsOpen(false)} />
      )}

      {/* Mobile menu */}
      <div
        className={`fixed top-0 right-0 z-50 h-full w-64 transform bg-gray-900 border-l border-gray-700 transition-transform duration-200 ease-in-out ${
          isOpen ? 'translate-x-0' : 'translate-x-full'
        }`}
      >
        <div className="flex items-center justify-between border-b border-gray-700 px-4 py-3">
          <span className="font-semibold text-white">Menu</span>
          <button
            onClick={() => setIsOpen(false)}
            className="rounded-md p-2 text-gray-400 hover:bg-gray-800 hover:text-white"
            aria-label="Close menu"
          >
            <X className="h-5 w-5" />
          </button>
        </div>
        <nav className="p-4 space-y-1">
          {navItems.map((item) => {
            const isActive = pathname === item.href || (item.href !== '/' && pathname.startsWith(item.href));
            const Icon = item.icon;
            return (
              <Link
                key={item.href}
                href={item.href}
                onClick={() => setIsOpen(false)}
                className={`flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors ${
                  isActive
                    ? 'bg-gray-800 text-white'
                    : 'text-gray-400 hover:bg-gray-800 hover:text-white'
                }`}
              >
                <Icon className="h-5 w-5" />
                {item.label}
              </Link>
            );
          })}
        </nav>
      </div>

      {/* Spacer for fixed header */}
      <div className="h-14" />
    </div>
  );
}
