# Frontend Code Review Prompt

## Context
You are reviewing frontend React/TypeScript code for quality, performance, accessibility, and maintainability.

## Review Checklist

### Component Architecture
- [ ] Single responsibility - component does one thing
- [ ] Appropriate size (< 300 lines generally)
- [ ] Clear prop interface
- [ ] Proper TypeScript types
- [ ] No prop drilling (> 3 levels)

### React Patterns
- [ ] Hooks follow Rules of Hooks
- [ ] Dependencies arrays are correct
- [ ] No unnecessary re-renders
- [ ] Appropriate memoization (not excessive)
- [ ] Error boundaries for error-prone sections

### Performance
- [ ] No inline objects/arrays in JSX (unless trivial)
- [ ] Large lists are virtualized
- [ ] Images optimized and lazy-loaded
- [ ] Code splitting for large components
- [ ] No expensive computations in render

### Accessibility
- [ ] Interactive elements are keyboard accessible
- [ ] Images have alt text
- [ ] Form inputs have labels
- [ ] Color contrast is sufficient
- [ ] ARIA attributes used correctly

### Testing
- [ ] Tests exist for critical functionality
- [ ] Tests follow Testing Library best practices
- [ ] Tests are not testing implementation details
- [ ] Edge cases are covered

### Security
- [ ] No dangerouslySetInnerHTML without sanitization
- [ ] User input is validated
- [ ] No sensitive data in client state
- [ ] API calls use proper authentication

## Output Format

Provide findings in priority order:

### Critical Issues
Issues that must be fixed before merge.

### Recommendations
Improvements that would enhance code quality.

### Positive Observations
Good patterns worth noting.

## Example Output

```markdown
## Code Review: UserProfile Component

### Critical Issues

1. **XSS Vulnerability** (security)
   - Location: `UserProfile.tsx:45`
   - Issue: `dangerouslySetInnerHTML` used with unsanitized user bio
   - Fix: Use DOMPurify or remove dangerous HTML rendering

2. **Missing Error Boundary** (reliability)
   - Location: `UserProfile.tsx`
   - Issue: Component fetches data but has no error handling UI
   - Fix: Wrap with ErrorBoundary or add error state rendering

### Recommendations

1. **Performance**: Consider memoizing `formatUserData` function (line 23)
2. **Accessibility**: Add `aria-label` to icon-only edit button (line 67)
3. **Types**: Replace `any` type for `userData` prop with proper interface

### Positive Observations

- Good use of custom hook `useUserData` for data fetching
- Proper loading state handling
- Clean component composition
```
