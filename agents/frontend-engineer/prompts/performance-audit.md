# Frontend Performance Audit Prompt

## Context
You are performing a performance audit on a frontend React application to identify bottlenecks and optimization opportunities.

## Audit Areas

### 1. Bundle Analysis
Analyze bundle size and composition:
- Total bundle size (JS + CSS)
- Largest chunks and their contents
- Duplicate dependencies
- Unused code (tree-shaking effectiveness)
- Heavy dependencies that could be replaced

### 2. Loading Performance
Analyze initial load:
- Critical rendering path
- Render-blocking resources
- Resource prioritization (preload, prefetch)
- Code splitting effectiveness
- Lazy loading implementation

### 3. Runtime Performance
Analyze React performance:
- Unnecessary re-renders
- Missing memoization
- Expensive computations in render
- Large lists without virtualization
- Memory leaks (event listeners, subscriptions)

### 4. Asset Optimization
Analyze static assets:
- Image formats and sizes
- Font loading strategy
- SVG optimization
- Caching strategy

### 5. Core Web Vitals
Estimate impact on:
- LCP (Largest Contentful Paint)
- FID/INP (First Input Delay / Interaction to Next Paint)
- CLS (Cumulative Layout Shift)

## Analysis Process

1. **Scan package.json** for heavy dependencies
2. **Review webpack/vite config** for optimization settings
3. **Analyze component structure** for re-render issues
4. **Check image handling** for optimization
5. **Review data fetching** for efficiency

## Output Format

```json
{
  "summary": {
    "overall_health": "good|moderate|poor",
    "estimated_lcp": "x.xs",
    "critical_issues": 0,
    "opportunities": 0
  },
  "bundle_analysis": {
    "total_size_kb": 0,
    "largest_dependencies": [],
    "recommendations": []
  },
  "render_performance": {
    "issues": [],
    "recommendations": []
  },
  "asset_optimization": {
    "issues": [],
    "recommendations": []
  },
  "prioritized_actions": [
    {
      "priority": 1,
      "action": "...",
      "impact": "high|medium|low",
      "effort": "low|medium|high",
      "estimated_improvement": "..."
    }
  ]
}
```

## Example Findings

### Bundle Issues
| Issue | Impact | Fix |
|-------|--------|-----|
| moment.js (290KB) | High | Replace with date-fns |
| Full lodash import | Medium | Use lodash-es with tree shaking |
| Large image in bundle | Medium | Move to CDN with lazy loading |

### Render Issues
| Component | Issue | Fix |
|-----------|-------|-----|
| ProductList | Re-renders on every state change | Wrap with React.memo |
| Dashboard | Inline object props | Extract to useMemo |
| SearchResults | 1000+ items without virtualization | Add react-virtual |

### Quick Wins
1. Add `loading="lazy"` to below-fold images
2. Preload hero image
3. Enable gzip/brotli compression
4. Add font-display: swap to custom fonts
