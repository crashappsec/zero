# Performance Optimization Guide

## Core Web Vitals

### Largest Contentful Paint (LCP)
Measures loading performance. Target: < 2.5 seconds.

**Optimization strategies:**
- Optimize and compress images (WebP, AVIF)
- Preload critical resources
- Use CDN for static assets
- Remove render-blocking resources
- Server-side render above-the-fold content

```html
<!-- Preload hero image -->
<link rel="preload" as="image" href="/hero.webp" />

<!-- Preload critical font -->
<link rel="preload" as="font" href="/font.woff2" crossorigin />
```

### First Input Delay (FID) / Interaction to Next Paint (INP)
Measures interactivity. Target: < 100ms (FID) / < 200ms (INP).

**Optimization strategies:**
- Code-split and lazy load non-critical JS
- Break up long tasks (> 50ms)
- Use web workers for heavy computation
- Minimize main thread work

### Cumulative Layout Shift (CLS)
Measures visual stability. Target: < 0.1.

**Optimization strategies:**
- Set explicit dimensions on images/videos
- Reserve space for ads/embeds
- Avoid inserting content above existing content
- Use CSS `contain` property

```tsx
// Reserve space for dynamic content
<div style={{ minHeight: '200px' }}>
  {content || <Skeleton />}
</div>

// Image with dimensions
<img src={src} width={800} height={600} alt="..." />
```

## Bundle Optimization

### Code Splitting

```tsx
// Route-based splitting
const Dashboard = lazy(() => import('./pages/Dashboard'));
const Settings = lazy(() => import('./pages/Settings'));

// Component-based splitting
const HeavyChart = lazy(() => import('./components/HeavyChart'));

// With Suspense
<Suspense fallback={<Loading />}>
  <Dashboard />
</Suspense>
```

### Tree Shaking

```tsx
// ❌ Imports entire library
import _ from 'lodash';

// ✅ Imports only what's used
import debounce from 'lodash/debounce';

// ✅ Named imports (with proper library support)
import { debounce } from 'lodash-es';
```

### Dynamic Imports

```tsx
// Load on demand
const handleClick = async () => {
  const { format } = await import('date-fns');
  setFormattedDate(format(date, 'PPP'));
};

// Conditional loading
if (process.env.NODE_ENV === 'development') {
  const devTools = await import('./devTools');
  devTools.init();
}
```

### Bundle Analysis

```bash
# Webpack
npx webpack-bundle-analyzer stats.json

# Vite
npx vite-bundle-visualizer

# Next.js
ANALYZE=true npm run build
```

## React Performance

### Memoization

```tsx
// Memoize component
const ExpensiveList = memo(({ items }) => (
  <ul>
    {items.map(item => <li key={item.id}>{item.name}</li>)}
  </ul>
));

// Memoize callback
const handleClick = useCallback((id) => {
  setSelected(id);
}, []);

// Memoize computed value
const sortedItems = useMemo(() => {
  return [...items].sort((a, b) => a.name.localeCompare(b.name));
}, [items]);
```

### Virtualization

For long lists (> 100 items):

```tsx
import { useVirtualizer } from '@tanstack/react-virtual';

function VirtualList({ items }) {
  const parentRef = useRef(null);

  const virtualizer = useVirtualizer({
    count: items.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 50,
  });

  return (
    <div ref={parentRef} style={{ height: '400px', overflow: 'auto' }}>
      <div style={{ height: `${virtualizer.getTotalSize()}px` }}>
        {virtualizer.getVirtualItems().map((virtualItem) => (
          <div
            key={virtualItem.key}
            style={{
              position: 'absolute',
              top: virtualItem.start,
              height: virtualItem.size,
            }}
          >
            {items[virtualItem.index].name}
          </div>
        ))}
      </div>
    </div>
  );
}
```

### Avoiding Re-renders

```tsx
// ❌ Creates new object every render
<Component style={{ color: 'red' }} />

// ✅ Stable reference
const style = useMemo(() => ({ color: 'red' }), []);
<Component style={style} />

// ❌ Creates new callback every render
<Button onClick={() => handleClick(id)} />

// ✅ Stable callback
const handleButtonClick = useCallback(() => handleClick(id), [id]);
<Button onClick={handleButtonClick} />
```

## Image Optimization

### Modern Formats

```tsx
<picture>
  <source srcSet="/image.avif" type="image/avif" />
  <source srcSet="/image.webp" type="image/webp" />
  <img src="/image.jpg" alt="..." />
</picture>
```

### Responsive Images

```tsx
<img
  srcSet="/image-320.jpg 320w,
          /image-640.jpg 640w,
          /image-1280.jpg 1280w"
  sizes="(max-width: 320px) 280px,
         (max-width: 640px) 600px,
         1200px"
  src="/image-1280.jpg"
  alt="..."
/>
```

### Lazy Loading

```tsx
// Native lazy loading
<img src="/image.jpg" loading="lazy" alt="..." />

// Intersection Observer for custom behavior
const [isVisible, ref] = useIntersectionObserver();

{isVisible && <img src="/heavy-image.jpg" alt="..." />}
```

## Network Optimization

### Resource Hints

```html
<!-- DNS prefetch for external domains -->
<link rel="dns-prefetch" href="https://api.example.com" />

<!-- Preconnect for critical third parties -->
<link rel="preconnect" href="https://fonts.googleapis.com" />

<!-- Prefetch likely next pages -->
<link rel="prefetch" href="/likely-next-page.js" />

<!-- Preload critical resources -->
<link rel="preload" href="/critical.css" as="style" />
```

### Caching Strategy

```javascript
// Service worker caching example
// Cache static assets with long TTL
// Cache API responses with short TTL + revalidation

// HTTP headers
Cache-Control: public, max-age=31536000, immutable  // Static assets
Cache-Control: private, max-age=0, must-revalidate  // HTML
Cache-Control: public, max-age=300, stale-while-revalidate=86400  // API
```

### Data Fetching

```tsx
// Parallel fetching
const [users, posts] = await Promise.all([
  fetchUsers(),
  fetchPosts()
]);

// React Query with stale-while-revalidate
const { data } = useQuery({
  queryKey: ['users'],
  queryFn: fetchUsers,
  staleTime: 5 * 60 * 1000,  // 5 minutes
});
```

## CSS Performance

### Critical CSS

Extract and inline critical above-the-fold CSS:

```html
<head>
  <style>
    /* Critical CSS inlined */
    .hero { ... }
  </style>
  <link rel="preload" href="/full.css" as="style" onload="this.onload=null;this.rel='stylesheet'">
</head>
```

### Reduce CSS Bundle

```css
/* Use CSS containment */
.card {
  contain: layout style paint;
}

/* Avoid expensive selectors */
/* ❌ */ .nav ul li a { }
/* ✅ */ .nav-link { }

/* Prefer transform/opacity for animations */
/* ❌ */ .animate { left: 0; transition: left 0.3s; }
/* ✅ */ .animate { transform: translateX(0); transition: transform 0.3s; }
```

## Monitoring

### Performance API

```tsx
// Measure component render time
useEffect(() => {
  performance.mark('component-mount-start');
  return () => {
    performance.mark('component-mount-end');
    performance.measure(
      'component-mount',
      'component-mount-start',
      'component-mount-end'
    );
  };
}, []);
```

### Web Vitals

```tsx
import { getCLS, getFID, getLCP } from 'web-vitals';

getCLS(console.log);
getFID(console.log);
getLCP(console.log);
```

### React Profiler

```tsx
<Profiler id="Navigation" onRender={onRenderCallback}>
  <Navigation />
</Profiler>

function onRenderCallback(
  id,
  phase,
  actualDuration,
  baseDuration,
  startTime,
  commitTime
) {
  // Log or send to analytics
}
```

## Performance Budget

| Metric | Budget |
|--------|--------|
| Total JS | < 300KB |
| Total CSS | < 100KB |
| LCP | < 2.5s |
| FID/INP | < 100ms |
| CLS | < 0.1 |
| Time to Interactive | < 3.8s |

Tools: Lighthouse CI, bundlesize, size-limit
