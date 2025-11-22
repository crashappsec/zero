<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# GitHub Repository Rename Instructions

## Current Status
- ✅ Local directory renamed: `gibson-powers`
- 🔜 GitHub repository needs rename: `skills-and-prompts-and-rag` → `gibson-powers`

## Option 1: Rename via GitHub Web Interface (Recommended)

1. Go to: https://github.com/crashappsec/skills-and-prompts-and-rag
2. Click **Settings** tab
3. Scroll to **Repository name** section
4. Change name from `skills-and-prompts-and-rag` to `gibson-powers`
5. Click **Rename**
6. GitHub will automatically:
   - Redirect old URLs to new repository
   - Update clone URLs
   - Preserve all issues, PRs, stars, etc.

## Option 2: Rename via GitHub CLI

If you have `gh` CLI authenticated:

```bash
# Authenticate if needed
gh auth login

# Rename repository
gh repo rename gibson-powers --yes

# Verify
gh repo view
```

## After Rename: Update Local Git Remote

Once the GitHub repository is renamed, update your local git remote:

```bash
cd /Users/curphey/Documents/GitHub/gibson-powers

# Check current remote
git remote -v

# Update remote URL (GitHub handles redirects, but best to update)
git remote set-url origin https://github.com/crashappsec/gibson-powers.git

# Verify
git remote -v

# Test connection
git fetch
```

## Verification Checklist

After renaming, verify:

- [ ] Repository accessible at: https://github.com/crashappsec/gibson-powers
- [ ] Old URL redirects properly
- [ ] `git fetch` works with new remote
- [ ] `git push` works with new remote
- [ ] All GitHub features intact (issues, PRs, wiki, etc.)
- [ ] Clone URLs updated in GitHub UI

## Additional Working Directories

These additional working directories were configured and may need remote updates:

```bash
# Update each directory after GitHub rename
cd /Users/curphey/Documents/GitHub/gibson-powers/prompts
git remote set-url origin https://github.com/crashappsec/gibson-powers.git

cd /Users/curphey/Documents/GitHub/gibson-powers/prompts/dora
git remote set-url origin https://github.com/crashappsec/gibson-powers.git

cd /Users/curphey/Documents/GitHub/gibson-powers/prompts/code-ownership
git remote set-url origin https://github.com/crashappsec/gibson-powers.git

cd /Users/curphey/Documents/GitHub/gibson-powers/rag/supply-chain
git remote set-url origin https://github.com/crashappsec/gibson-powers.git
```

## Notes

- GitHub maintains permanent redirects from old repository name
- All links will continue to work via redirects
- Updating remote URLs explicitly is good practice but not strictly required
- Forks, stars, watchers, and all other metadata are preserved
- No data loss occurs during rename

## Quick Command Reference

```bash
# View current repository info
gh repo view

# Check remote URLs
git remote -v

# Update remote
git remote set-url origin https://github.com/crashappsec/gibson-powers.git

# Test connection
git remote show origin

# Fetch from new URL
git fetch

# Push to verify
git push
```
