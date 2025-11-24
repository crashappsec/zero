# TypeScript

**Category**: languages
**Description**: TypeScript programming language - JavaScript with static typing, developed by Microsoft
**Homepage**: https://www.typescriptlang.org

## Package Detection

### NPM
- `typescript`
- `ts-node`
- `tsx`
- `@types/node`
- `@typescript-eslint/parser`
- `@typescript-eslint/eslint-plugin`

## Configuration Files

- `tsconfig.json`
- `tsconfig.*.json`
- `tsconfig.build.json`
- `tsconfig.node.json`
- `.ts-node`

## File Extensions

- `.ts`
- `.tsx` (React TSX)
- `.mts` (ES modules)
- `.cts` (CommonJS)
- `.d.ts` (type declarations)

## Import Detection

### TypeScript
**Pattern**: `import\s+.*\s+from\s+['"]`
- ES module import
- Example: `import { useState } from 'react';`

**Pattern**: `import\s+type\s+`
- Type-only import
- Example: `import type { User } from './types';`

**Pattern**: `:\s*(string|number|boolean|object|any|void|never)`
- Type annotations
- Example: `function greet(name: string): void {}`

**Pattern**: `interface\s+\w+`
- Interface declaration
- Example: `interface User { name: string; }`

**Pattern**: `type\s+\w+\s*=`
- Type alias
- Example: `type ID = string | number;`

## Environment Variables

- `TS_NODE_PROJECT`
- `TS_NODE_COMPILER_OPTIONS`

## Version Indicators

- TypeScript 5.x (current)
- TypeScript 4.x (widely used)

## Detection Notes

- Look for `.ts` or `.tsx` files
- tsconfig.json is the primary indicator
- Check for @types/* packages in dependencies
- Often used alongside JavaScript in the same project
- `.d.ts` files indicate type definitions

## Detection Confidence

- **File Extension Detection**: 95% (HIGH)
- **tsconfig.json Detection**: 95% (HIGH)
- **Package Detection**: 95% (HIGH)
