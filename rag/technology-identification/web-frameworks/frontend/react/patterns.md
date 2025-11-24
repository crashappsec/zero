# React

**Category**: web-frameworks/frontend
**Description**: A JavaScript library for building user interfaces
**Homepage**: https://www.npmjs.com/package/react

## Package Detection

### NPM
*Core React packages*

- `react`
- `react-dom`

### YARN
*Core React packages via Yarn*

- `react`
- `react-dom`

### PNPM
*Core React packages via pnpm*

- `react`
- `react-dom`

### Related Packages
- `react-router`
- `react-router-dom`
- `react-redux`
- `react-query`
- `@tanstack/react-query`
- `react-hook-form`
- `react-native`
- `next`
- `gatsby`

## Import Detection

### Javascript
File extensions: .js, .jsx, .mjs

**Pattern**: `import\s+React\s+from\s+['"]react['"]`
- Default React import (ES6)
- Example: `import React from 'react';`

**Pattern**: `import\s+\{[^}]+\}\s+from\s+['"]react['"]`
- Named imports from React
- Example: `import { useState, useEffect } from 'react';`

**Pattern**: `import\s+\*\s+as\s+React\s+from\s+['"]react['"]`
- Namespace import
- Example: `import * as React from 'react';`

**Pattern**: `const\s+React\s*=\s*require\(['"]react['"]\)`
- CommonJS require
- Example: `const React = require('react');`

**Pattern**: `from\s+['"]react-dom['"]`
- React DOM imports
- Example: `import ReactDOM from 'react-dom';`

### Typescript
File extensions: .ts, .tsx

**Pattern**: `import\s+React\s+from\s+['"]react['"]`
- Default React import
- Example: `import React from 'react';`

**Pattern**: `import\s+type\s+\{[^}]+\}\s+from\s+['"]react['"]`
- Type-only imports
- Example: `import type { FC, ReactNode } from 'react';`

**Pattern**: `import\s+\{[^}]+\}\s+from\s+['"]react['"]`
- Named imports
- Example: `import { useState, useEffect } from 'react';`

## Environment Variables

*Create React App environment variables*

- Prefix: `REACT_APP_*`
*Next.js public environment variables*

- Prefix: `NEXT_PUBLIC_*`
*Gatsby environment variables*

- Prefix: `GATSBY_*`
*Vite environment variables*

- Prefix: `VITE_*`

## Configuration Files

- `.babelrc`
- `.babelrc.json`
- `.babelrc.js`
- `babel.config.js`
- `babel.config.json`
- `webpack.config.js`
- `webpack.config.ts`
- `tsconfig.json`
- `vite.config.js`
- `vite.config.ts`
- `.eslintrc`
- `.eslintrc.json`
- `.eslintrc.js`
- `eslint.config.js`

## Detection Notes

- Presence of both 'react' and 'react-dom' is strong signal
- Related packages alone may indicate React usage
- Check for JSX file extensions (.jsx, .tsx)

## Detection Confidence

- **Package Detection**: 95% (HIGH)
- **Import Detection**: 85% (HIGH)
- **Environment Variable Detection**: 60% (MEDIUM)
