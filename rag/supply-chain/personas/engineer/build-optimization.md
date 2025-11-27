<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Build Optimization and Bundle Analysis

## Understanding Bundle Size Impact

### Why Bundle Size Matters

```
User Impact:
├── Initial Load Time: +100KB = ~1s on 3G
├── Time to Interactive: Blocked by JS parsing
├── Mobile Data Costs: Real money in many markets
└── SEO Rankings: Core Web Vitals factor
```

**Budget Guidelines:**
| App Type | Initial JS Budget | Total Budget |
|----------|-------------------|--------------|
| Landing page | <50KB | <200KB |
| Web app | <150KB | <500KB |
| Complex SPA | <300KB | <1MB |

## Tree Shaking Optimization

### What is Tree Shaking?

Tree shaking removes unused code from your bundle:

```javascript
// lodash - BAD (imports entire library ~70KB)
import _ from 'lodash';
_.get(obj, 'path');

// lodash - GOOD (imports only what's used ~2KB)
import get from 'lodash/get';
get(obj, 'path');

// lodash-es - BEST (tree-shakeable ES modules)
import { get } from 'lodash-es';
```

### Enabling Tree Shaking

**Webpack:**
```javascript
// webpack.config.js
module.exports = {
  mode: 'production',  // Enables tree shaking
  optimization: {
    usedExports: true,
    sideEffects: true,
  }
};
```

**Package.json sideEffects:**
```json
{
  "sideEffects": false,
  // Or specify files with side effects
  "sideEffects": ["*.css", "*.scss"]
}
```

### Common Tree Shaking Blockers

1. **CommonJS imports:**
   ```javascript
   // NOT tree-shakeable
   const { get } = require('lodash');

   // Tree-shakeable
   import { get } from 'lodash-es';
   ```

2. **Barrel exports with side effects:**
   ```javascript
   // index.js barrel - may block tree shaking
   export * from './moduleA';
   export * from './moduleB';
   ```

3. **Dynamic imports:**
   ```javascript
   // NOT tree-shakeable
   const method = 'get';
   import('lodash')[method]();
   ```

## Analyzing Bundle Content

### Webpack Bundle Analyzer

```bash
# Install
npm install --save-dev webpack-bundle-analyzer

# Add to webpack config
const BundleAnalyzerPlugin = require('webpack-bundle-analyzer').BundleAnalyzerPlugin;

module.exports = {
  plugins: [
    new BundleAnalyzerPlugin()
  ]
};

# Or analyze existing stats
npx webpack --profile --json > stats.json
npx webpack-bundle-analyzer stats.json
```

### Source Map Explorer

```bash
# Install
npm install --save-dev source-map-explorer

# Analyze
npx source-map-explorer dist/main.js
```

### Bundle Size Tracking

```bash
# Install bundlesize
npm install --save-dev bundlesize

# package.json configuration
{
  "bundlesize": [
    {
      "path": "./dist/main.js",
      "maxSize": "150 kB"
    },
    {
      "path": "./dist/vendor.js",
      "maxSize": "250 kB"
    }
  ]
}

# Run check
npx bundlesize
```

## Code Splitting Strategies

### Route-Based Splitting

```javascript
// React with lazy loading
import { lazy, Suspense } from 'react';

const Dashboard = lazy(() => import('./Dashboard'));
const Settings = lazy(() => import('./Settings'));

function App() {
  return (
    <Suspense fallback={<Loading />}>
      <Routes>
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/settings" element={<Settings />} />
      </Routes>
    </Suspense>
  );
}
```

### Vendor Chunking

```javascript
// webpack.config.js
optimization: {
  splitChunks: {
    chunks: 'all',
    cacheGroups: {
      vendor: {
        test: /[\\/]node_modules[\\/]/,
        name: 'vendors',
        chunks: 'all',
      },
      // Separate large libraries
      react: {
        test: /[\\/]node_modules[\\/](react|react-dom)[\\/]/,
        name: 'react',
        chunks: 'all',
      },
    },
  },
}
```

### Dynamic Imports for Features

```javascript
// Load heavy library only when needed
async function generatePDF() {
  const { jsPDF } = await import('jspdf');
  const doc = new jsPDF();
  // Generate PDF...
}

// Load polyfills conditionally
if (!window.IntersectionObserver) {
  await import('intersection-observer');
}
```

## Dependency Optimization

### Identifying Heavy Dependencies

```bash
# Using bundlephobia CLI
npx bundlephobia moment lodash axios

# Output:
# moment: 72.1kB (minified), 20.2kB (gzipped)
# lodash: 71.5kB (minified), 25.2kB (gzipped)
# axios: 13.7kB (minified), 4.6kB (gzipped)
```

### Lightweight Alternatives

| Heavy Package | Size | Alternative | Size |
|--------------|------|-------------|------|
| moment | 72KB | date-fns | 13KB* |
| moment | 72KB | dayjs | 2KB |
| lodash | 72KB | lodash-es | 0KB* |
| axios | 14KB | ky | 3KB |
| uuid | 12KB | nanoid | 1KB |
| classnames | 2KB | clsx | 0.5KB |
| numeral | 18KB | Intl.NumberFormat | 0KB |

*With tree shaking

### Replacing Moment.js

```javascript
// moment.js (72KB)
moment().format('YYYY-MM-DD');
moment().add(1, 'days');

// dayjs (2KB) - mostly compatible API
import dayjs from 'dayjs';
dayjs().format('YYYY-MM-DD');
dayjs().add(1, 'day');

// date-fns (tree-shakeable)
import { format, addDays } from 'date-fns';
format(new Date(), 'yyyy-MM-dd');
addDays(new Date(), 1);

// Native (0KB) - for simple cases
new Date().toISOString().split('T')[0];
```

## Build Performance

### Caching Strategies

**Webpack:**
```javascript
module.exports = {
  cache: {
    type: 'filesystem',
    buildDependencies: {
      config: [__filename],
    },
  },
};
```

**Output hashing for cache busting:**
```javascript
output: {
  filename: '[name].[contenthash].js',
  chunkFilename: '[name].[contenthash].chunk.js',
}
```

### Parallel Processing

```javascript
// thread-loader for expensive loaders
module: {
  rules: [
    {
      test: /\.js$/,
      use: [
        'thread-loader',
        'babel-loader',
      ],
    },
  ],
}
```

### Build Time Analysis

```bash
# Webpack build speed
npx speed-measure-webpack-plugin

# Output timing per plugin/loader
SMP  ⏱
General output time took 5.23 secs

 SMP  ⏱  Plugins
    TerserPlugin took 2.14 secs

 SMP  ⏱  Loaders
    babel-loader took 1.92 secs
    css-loader took 0.84 secs
```

## Production Optimizations

### Compression

```javascript
// webpack.config.js
const CompressionPlugin = require('compression-webpack-plugin');

plugins: [
  new CompressionPlugin({
    algorithm: 'gzip',
    test: /\.(js|css|html|svg)$/,
    threshold: 8192,  // Only compress > 8KB
    minRatio: 0.8,
  }),
];
```

### Minification

```javascript
optimization: {
  minimize: true,
  minimizer: [
    new TerserPlugin({
      terserOptions: {
        compress: {
          drop_console: true,  // Remove console.logs
        },
      },
    }),
    new CssMinimizerPlugin(),
  ],
}
```

### Environment-Specific Builds

```javascript
// Remove dev-only code in production
new webpack.DefinePlugin({
  'process.env.NODE_ENV': JSON.stringify('production'),
  __DEV__: false,
});

// In code
if (__DEV__) {
  // This entire block is removed in production
  enableDevTools();
}
```

## Monitoring Bundle Size in CI

### GitHub Actions Example

```yaml
name: Bundle Size Check

on: [pull_request]

jobs:
  bundle-size:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install dependencies
        run: npm ci

      - name: Build
        run: npm run build

      - name: Check bundle size
        uses: preactjs/compressed-size-action@v2
        with:
          repo-token: "${{ secrets.GITHUB_TOKEN }}"
          pattern: "./dist/**/*.js"
```

### Size Limit Configuration

```json
// package.json
{
  "size-limit": [
    {
      "path": "dist/index.js",
      "limit": "50 KB"
    },
    {
      "path": "dist/vendor.js",
      "limit": "200 KB"
    }
  ],
  "scripts": {
    "size": "size-limit",
    "size:ci": "size-limit --ci"
  }
}
```

## Quick Reference

### Bundle Size Checklist

- [ ] Enable production mode in bundler
- [ ] Configure tree shaking (ES modules)
- [ ] Implement code splitting (routes, features)
- [ ] Analyze bundle with visualization tool
- [ ] Replace heavy dependencies with lighter alternatives
- [ ] Set up CI size checks with budgets
- [ ] Enable compression (gzip/brotli)
- [ ] Configure proper cache headers

### Size Impact Rules of Thumb

```
1KB gzipped = ~3-4KB minified = ~10-15KB source
1KB increase = ~10ms on 3G network
100KB JS = ~100ms parse time on mobile
```
