# API Review Prompt

## Context
You are reviewing backend API code for design quality, performance, security, and maintainability.

## Review Checklist

### API Design
- [ ] RESTful conventions followed (nouns, HTTP methods)
- [ ] Consistent response envelope
- [ ] Appropriate status codes
- [ ] Pagination for collections
- [ ] Proper error responses with codes

### Security
- [ ] Authentication required for protected routes
- [ ] Authorization checks (ownership, roles)
- [ ] Input validation and sanitization
- [ ] Rate limiting in place
- [ ] No sensitive data in logs/errors

### Performance
- [ ] Database queries optimized (no N+1)
- [ ] Appropriate indexes exist
- [ ] Pagination limits enforced
- [ ] Caching where appropriate
- [ ] Connection pooling configured

### Error Handling
- [ ] All errors caught and handled
- [ ] Consistent error format
- [ ] No stack traces in production responses
- [ ] Logging for debugging

### Testing
- [ ] Unit tests for business logic
- [ ] Integration tests for API endpoints
- [ ] Edge cases covered
- [ ] Error scenarios tested

## Output Format

```markdown
## API Review: [Service/Endpoint Name]

### Design Issues

1. **[Issue Type]** (severity: high|medium|low)
   - Location: `path/to/file.ts:line`
   - Issue: Description
   - Recommendation: Specific fix

### Security Concerns

1. **[Vulnerability Type]**
   - Location: `path/to/file.ts:line`
   - Risk: Potential impact
   - Fix: How to remediate

### Performance Recommendations

1. **[Optimization]**
   - Current: What's happening now
   - Impact: Performance cost
   - Suggestion: Improvement

### Positive Observations

- Good patterns worth noting

### Summary

| Category | Issues |
|----------|--------|
| Design | X |
| Security | X |
| Performance | X |
| Total | X |
```

## Example Output

```markdown
## API Review: User Orders API

### Design Issues

1. **Verb in URL** (severity: medium)
   - Location: `routes/orders.ts:15`
   - Issue: Using POST /api/getOrders instead of GET /api/orders
   - Recommendation: Change to `GET /api/orders` with query params

2. **Inconsistent Response Format** (severity: low)
   - Location: `controllers/orders.ts:45`
   - Issue: Success returns `{orders: [...]}` but error returns `{message: ""}`
   - Recommendation: Use consistent envelope `{data, error, meta}`

### Security Concerns

1. **Missing Authorization Check** (severity: high)
   - Location: `controllers/orders.ts:23`
   - Risk: Any authenticated user can view any order
   - Fix: Add ownership check `if (order.userId !== req.user.id)`

2. **SQL Injection Risk** (severity: critical)
   - Location: `services/orders.ts:67`
   - Risk: Direct string interpolation in query
   - Fix: Use parameterized query

### Performance Recommendations

1. **N+1 Query Problem**
   - Current: Fetching user for each order in loop
   - Impact: 100 orders = 101 queries
   - Suggestion: Use JOIN or batch fetch users

### Positive Observations

- Good use of TypeScript types for request/response
- Proper async/await error handling
- Clear separation of routes, controllers, services
```
