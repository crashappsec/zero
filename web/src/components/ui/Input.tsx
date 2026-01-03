'use client';

import { cn } from '@/lib/utils';
import { forwardRef } from 'react';

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  icon?: React.ReactNode;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, icon, ...props }, ref) => {
    return (
      <div className="relative">
        {icon && (
          <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">{icon}</div>
        )}
        <input
          ref={ref}
          className={cn(
            'w-full rounded-md border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100',
            'placeholder:text-gray-500',
            'focus:border-green-500 focus:outline-none focus:ring-1 focus:ring-green-500',
            'disabled:opacity-50 disabled:cursor-not-allowed',
            icon && 'pl-10',
            className
          )}
          {...props}
        />
      </div>
    );
  }
);

Input.displayName = 'Input';
