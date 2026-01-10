'use client';

import { cn } from '@/lib/utils';

export type TierLevel = 'elite' | 'good' | 'fair' | 'needs_focus' | 'unknown';

interface BenchmarkTierProps {
  value: number | string;
  label: string;
  tiers: {
    elite: { max: number; label: string };
    good: { max: number; label: string };
    fair: { max: number; label: string };
  };
  unit?: string;
  lowerIsBetter?: boolean;
  className?: string;
}

// Determine tier based on value and tier thresholds
export function getTier(
  value: number,
  tiers: BenchmarkTierProps['tiers'],
  lowerIsBetter: boolean = true
): TierLevel {
  if (lowerIsBetter) {
    if (value <= tiers.elite.max) return 'elite';
    if (value <= tiers.good.max) return 'good';
    if (value <= tiers.fair.max) return 'fair';
    return 'needs_focus';
  } else {
    // Higher is better (invert comparison)
    if (value >= tiers.elite.max) return 'elite';
    if (value >= tiers.good.max) return 'good';
    if (value >= tiers.fair.max) return 'fair';
    return 'needs_focus';
  }
}

const tierColors: Record<TierLevel, { bg: string; text: string; border: string }> = {
  elite: { bg: 'bg-green-500/20', text: 'text-green-500', border: 'border-green-500' },
  good: { bg: 'bg-blue-500/20', text: 'text-blue-500', border: 'border-blue-500' },
  fair: { bg: 'bg-yellow-500/20', text: 'text-yellow-500', border: 'border-yellow-500' },
  needs_focus: { bg: 'bg-red-500/20', text: 'text-red-500', border: 'border-red-500' },
  unknown: { bg: 'bg-gray-500/20', text: 'text-gray-500', border: 'border-gray-500' },
};

const tierLabels: Record<TierLevel, string> = {
  elite: 'Elite',
  good: 'Good',
  fair: 'Fair',
  needs_focus: 'Needs Focus',
  unknown: 'Unknown',
};

export function BenchmarkTier({
  value,
  label,
  tiers,
  unit = '',
  lowerIsBetter = true,
  className,
}: BenchmarkTierProps) {
  const numericValue = typeof value === 'string' ? parseFloat(value) || 0 : value;
  const tier = isNaN(numericValue) ? 'unknown' : getTier(numericValue, tiers, lowerIsBetter);
  const colors = tierColors[tier];

  return (
    <div className={cn('rounded-lg p-4', colors.bg, className)}>
      <div className="flex items-center justify-between mb-2">
        <span className="text-sm text-gray-400">{label}</span>
        <span className={cn('text-xs font-medium px-2 py-0.5 rounded', colors.bg, colors.text)}>
          {tierLabels[tier]}
        </span>
      </div>
      <p className={cn('text-2xl font-bold', colors.text)}>
        {typeof value === 'string' ? value : value.toLocaleString()}
        {unit && <span className="text-sm font-normal text-gray-400 ml-1">{unit}</span>}
      </p>
      <div className="mt-3 flex gap-1">
        {(['elite', 'good', 'fair', 'needs_focus'] as const).map((t) => (
          <div
            key={t}
            className={cn(
              'h-1.5 flex-1 rounded',
              tier === t ? tierColors[t].bg.replace('/20', '') : 'bg-gray-700'
            )}
          />
        ))}
      </div>
      <div className="mt-1 flex justify-between text-xs text-gray-500">
        <span>{tiers.elite.label}</span>
        <span>{tiers.good.label}</span>
        <span>{tiers.fair.label}</span>
        <span>Needs Focus</span>
      </div>
    </div>
  );
}

// Pre-configured benchmarks from LinearB 2026 report
export const LINEARB_BENCHMARKS = {
  cycleTime: {
    elite: { max: 25, label: '<25h' },
    good: { max: 72, label: '25-72h' },
    fair: { max: 161, label: '73-161h' },
  },
  deployTime: {
    elite: { max: 16, label: '<16h' },
    good: { max: 106, label: '16-106h' },
    fair: { max: 277, label: '107-277h' },
  },
  pickupTime: {
    elite: { max: 1, label: '<1h' },
    good: { max: 4, label: '1-4h' },
    fair: { max: 16, label: '5-16h' },
  },
  reviewTime: {
    elite: { max: 3, label: '<3h' },
    good: { max: 14, label: '3-14h' },
    fair: { max: 24, label: '15-24h' },
  },
  mergeTime: {
    elite: { max: 1, label: '<1h' },
    good: { max: 3, label: '1-3h' },
    fair: { max: 16, label: '4-16h' },
  },
  changeFailureRate: {
    elite: { max: 1, label: '<1%' },
    good: { max: 4, label: '1-4%' },
    fair: { max: 17, label: '5-17%' },
  },
  prSize: {
    elite: { max: 100, label: '<100' },
    good: { max: 155, label: '100-155' },
    fair: { max: 228, label: '156-228' },
  },
  reworkRate: {
    elite: { max: 3, label: '<3%' },
    good: { max: 5, label: '3-5%' },
    fair: { max: 8, label: '6-8%' },
  },
  refactorRate: {
    elite: { max: 11, label: '<11%' },
    good: { max: 16, label: '11-16%' },
    fair: { max: 22, label: '17-22%' },
  },
  // Higher is better metrics
  mergeFrequency: {
    elite: { max: 2.0, label: '>2.0' },
    good: { max: 1.2, label: '1.2-2.0' },
    fair: { max: 0.66, label: '0.66-1.2' },
  },
  deployFrequency: {
    elite: { max: 1.2, label: '>1.2' },
    good: { max: 0.5, label: '0.5-1.2' },
    fair: { max: 0.2, label: '0.2-0.5' },
  },
};

// Security benchmarks (Zero-specific)
export const SECURITY_BENCHMARKS = {
  criticalVulns: {
    elite: { max: 0, label: '0' },
    good: { max: 3, label: '1-3' },
    fair: { max: 10, label: '4-10' },
  },
  highVulns: {
    elite: { max: 5, label: '<5' },
    good: { max: 15, label: '5-15' },
    fair: { max: 50, label: '16-50' },
  },
  secretsExposed: {
    elite: { max: 0, label: '0' },
    good: { max: 0, label: '0' },
    fair: { max: 5, label: '1-5' },
  },
  weakCrypto: {
    elite: { max: 0, label: '0' },
    good: { max: 3, label: '1-3' },
    fair: { max: 10, label: '4-10' },
  },
};

// Supply chain benchmarks (Zero-specific)
export const SUPPLY_CHAIN_BENCHMARKS = {
  packageHealth: {
    elite: { max: 85, label: '>85%' },
    good: { max: 70, label: '70-85%' },
    fair: { max: 50, label: '50-70%' },
  },
  vulnerableDeps: {
    elite: { max: 3, label: '<3%' },
    good: { max: 10, label: '3-10%' },
    fair: { max: 25, label: '10-25%' },
  },
  licenseViolations: {
    elite: { max: 0, label: '0' },
    good: { max: 2, label: '0-2' },
    fair: { max: 10, label: '3-10' },
  },
  kevVulns: {
    elite: { max: 0, label: '0' },
    good: { max: 0, label: '0' },
    fair: { max: 2, label: '1-2' },
  },
};

// Quality benchmarks (Zero-specific)
export const QUALITY_BENCHMARKS = {
  testCoverage: {
    elite: { max: 80, label: '>80%' },
    good: { max: 60, label: '60-80%' },
    fair: { max: 40, label: '40-60%' },
  },
  complexity: {
    elite: { max: 5, label: '<5' },
    good: { max: 10, label: '5-10' },
    fair: { max: 20, label: '10-20' },
  },
  techDebtScore: {
    elite: { max: 80, label: '>80' },
    good: { max: 60, label: '60-80' },
    fair: { max: 40, label: '40-60' },
  },
  docCoverage: {
    elite: { max: 80, label: '>80%' },
    good: { max: 50, label: '50-80%' },
    fair: { max: 25, label: '25-50%' },
  },
};

// Team benchmarks (Zero-specific)
export const TEAM_BENCHMARKS = {
  busFactor: {
    elite: { max: 5, label: '>5' },
    good: { max: 3, label: '3-5' },
    fair: { max: 2, label: '2-3' },
  },
  ownershipCoverage: {
    elite: { max: 90, label: '>90%' },
    good: { max: 70, label: '70-90%' },
    fair: { max: 50, label: '50-70%' },
  },
  orphanedFiles: {
    elite: { max: 5, label: '<5%' },
    good: { max: 15, label: '5-15%' },
    fair: { max: 30, label: '15-30%' },
  },
  contributorCount: {
    elite: { max: 10, label: '>10' },
    good: { max: 5, label: '5-10' },
    fair: { max: 2, label: '2-5' },
  },
};
