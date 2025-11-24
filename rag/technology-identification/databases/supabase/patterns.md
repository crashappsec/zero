# Supabase

**Category**: databases
**Description**: Supabase - open source Firebase alternative
**Homepage**: https://supabase.com

## Package Detection

### NPM
*Supabase JavaScript SDKs*

- `@supabase/supabase-js`
- `@supabase/auth-helpers-nextjs`
- `@supabase/auth-helpers-react`
- `@supabase/ssr`

### PYPI
*Supabase Python SDK*

- `supabase`

### RUBYGEMS
*Supabase Ruby SDK*

- `supabase`

## Import Detection

### Javascript

**Pattern**: `from\s+['"]@supabase/supabase-js['"]`
- Type: esm_import

**Pattern**: `from\s+['"]@supabase/ssr['"]`
- Type: esm_import

**Pattern**: `from\s+['"]@supabase/auth-helpers`
- Type: esm_import

### Python

**Pattern**: `from\s+supabase`
- Type: python_import

## Environment Variables

*Supabase project URL*

*Supabase anonymous key*

*Supabase service role key*

*Next.js public Supabase URL*

*Next.js public anon key*


## Detection Confidence

- **Package Detection**: 95% (HIGH)
- **Import Detection**: 90% (HIGH)
- **Environment Variable Detection**: 85% (MEDIUM)
- **API Endpoint Detection**: 80% (MEDIUM)
