'use client';

import { useState, useRef, useEffect } from 'react';
import { Download, ChevronDown } from 'lucide-react';
import { Button } from './Button';

interface ExportButtonProps {
  onExport: (format: 'csv' | 'json') => void;
  disabled?: boolean;
  label?: string;
}

export function ExportButton({ onExport, disabled, label = 'Export' }: ExportButtonProps) {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  return (
    <div className="relative" ref={dropdownRef}>
      <Button
        variant="secondary"
        size="sm"
        onClick={() => setIsOpen(!isOpen)}
        disabled={disabled}
        icon={<Download className="h-4 w-4" />}
      >
        {label}
        <ChevronDown className="h-3 w-3 ml-1" />
      </Button>

      {isOpen && (
        <div className="absolute right-0 mt-1 w-32 rounded-md bg-gray-800 border border-gray-700 shadow-lg z-50">
          <div className="py-1">
            <button
              onClick={() => {
                onExport('csv');
                setIsOpen(false);
              }}
              className="w-full px-4 py-2 text-left text-sm text-gray-300 hover:bg-gray-700 hover:text-white"
            >
              Export CSV
            </button>
            <button
              onClick={() => {
                onExport('json');
                setIsOpen(false);
              }}
              className="w-full px-4 py-2 text-left text-sm text-gray-300 hover:bg-gray-700 hover:text-white"
            >
              Export JSON
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
