# JavaScript

**Category**: languages
**Description**: JavaScript programming language - the language of the web, used for frontend, backend (Node.js), and mobile development
**Homepage**: https://developer.mozilla.org/en-US/docs/Web/JavaScript

## Package Detection

### NPM
- `node`
- `npm`
- `yarn`
- `pnpm`

## Configuration Files

- `package.json`
- `package-lock.json`
- `yarn.lock`
- `pnpm-lock.yaml`
- `.npmrc`
- `.yarnrc`
- `.yarnrc.yml`
- `.nvmrc`
- `.node-version`
- `jsconfig.json`
- `.eslintrc`
- `.eslintrc.js`
- `.eslintrc.json`
- `.prettierrc`
- `.babelrc`
- `babel.config.js`
- `webpack.config.js`
- `rollup.config.js`
- `vite.config.js`
- `esbuild.config.js`

## File Extensions

- `.js`
- `.mjs` (ES modules)
- `.cjs` (CommonJS)
- `.jsx` (React JSX)

## Import Detection

### JavaScript
**Pattern**: `import\s+.*\s+from\s+['"]`
- ES module import
- Example: `import React from 'react';`

**Pattern**: `require\(['"]`
- CommonJS require
- Example: `const fs = require('fs');`

**Pattern**: `export\s+(default\s+)?(function|class|const|let|var)`
- ES module export
- Example: `export default function App() {}`

## Environment Variables

- `NODE_ENV`
- `NODE_PATH`
- `NPM_TOKEN`
- `NODE_OPTIONS`

## Version Indicators

- Node.js 22 (current)
- Node.js 20 LTS (active LTS)
- Node.js 18 LTS (maintenance)
- ES2024 (ECMAScript 2024)

## Detection Notes

- Look for `.js` files in repository
- package.json is the primary indicator
- Check for node_modules directory (usually gitignored)
- Look for bundler configs (webpack, rollup, vite)
- Modern projects may use ES modules (.mjs) or type: "module"

## Detection Confidence

- **File Extension Detection**: 95% (HIGH)
- **Configuration File Detection**: 95% (HIGH)
- **Package Detection**: 90% (HIGH)
