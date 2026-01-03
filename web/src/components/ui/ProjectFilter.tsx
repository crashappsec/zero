'use client';

import { useState, useRef, useEffect } from 'react';
import { ChevronDown, X, Filter } from 'lucide-react';
import { cn } from '@/lib/utils';

interface Project {
  id: string;
  name?: string;
}

interface ProjectFilterProps {
  projects: Project[];
  selectedProjects: string[];
  onChange: (projectIds: string[]) => void;
  className?: string;
}

export function ProjectFilter({
  projects,
  selectedProjects,
  onChange,
  className,
}: ProjectFilterProps) {
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

  const toggleProject = (projectId: string) => {
    if (selectedProjects.includes(projectId)) {
      onChange(selectedProjects.filter(id => id !== projectId));
    } else {
      onChange([...selectedProjects, projectId]);
    }
  };

  const selectAll = () => {
    onChange([]);
  };

  const isFiltered = selectedProjects.length > 0;
  const displayText = isFiltered
    ? selectedProjects.length === 1
      ? selectedProjects[0]
      : `${selectedProjects.length} projects`
    : 'All Projects';

  return (
    <div ref={dropdownRef} className={cn('relative', className)}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className={cn(
          'flex items-center gap-2 px-3 py-1.5 text-sm rounded-md border transition-colors',
          'focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2 focus:ring-offset-gray-900',
          isFiltered
            ? 'bg-green-600/20 border-green-600 text-green-400'
            : 'bg-gray-800 border-gray-700 text-gray-300 hover:border-gray-600'
        )}
      >
        <Filter className="h-4 w-4" />
        <span>{displayText}</span>
        <ChevronDown className={cn('h-4 w-4 transition-transform', isOpen && 'rotate-180')} />
      </button>

      {isFiltered && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            selectAll();
          }}
          className="ml-2 p-1 text-gray-400 hover:text-white transition-colors"
          title="Clear filter"
        >
          <X className="h-4 w-4" />
        </button>
      )}

      {isOpen && (
        <div className="absolute top-full left-0 mt-1 w-64 max-h-80 overflow-auto z-50 bg-gray-800 border border-gray-700 rounded-md shadow-lg">
          <button
            onClick={selectAll}
            className={cn(
              'w-full px-3 py-2 text-left text-sm hover:bg-gray-700 transition-colors border-b border-gray-700',
              !isFiltered && 'text-green-400 font-medium'
            )}
          >
            All Projects
          </button>
          {projects.map((project) => (
            <button
              key={project.id}
              onClick={() => toggleProject(project.id)}
              className={cn(
                'w-full px-3 py-2 text-left text-sm hover:bg-gray-700 transition-colors flex items-center justify-between',
                selectedProjects.includes(project.id) && 'text-green-400'
              )}
            >
              <span className="truncate">{project.id}</span>
              {selectedProjects.includes(project.id) && (
                <span className="text-green-500">✓</span>
              )}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
