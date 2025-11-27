# Architecture Review Prompt

## Context
You are reviewing system architecture for quality attributes, patterns, and trade-offs.

## Review Areas

### 1. Architecture Style Assessment
- What architecture style is used?
- Is it appropriate for the team size and requirements?
- Are boundaries clearly defined?

### 2. Quality Attributes
- **Scalability**: Can the system handle growth?
- **Reliability**: How does it handle failures?
- **Security**: Is auth/authz properly designed?
- **Maintainability**: Can it evolve over time?
- **Performance**: Are there bottlenecks?

### 3. Component Analysis
- Service boundaries (too fine, too coarse?)
- Data ownership (shared databases?)
- Communication patterns (sync vs async)
- Dependencies (circular? tight coupling?)

### 4. Trade-off Analysis
- What trade-offs were made?
- Are they documented?
- Are they appropriate?

## Output Format

```markdown
## Architecture Review: [System Name]

### Executive Summary
High-level assessment in 2-3 sentences.

### Architecture Overview
- **Style**: [Monolith/Microservices/Serverless/etc.]
- **Components**: [List main components]
- **Communication**: [HTTP/gRPC/Events/etc.]
- **Data Strategy**: [Database per service/Shared/etc.]

### Quality Attribute Assessment

| Attribute | Rating | Notes |
|-----------|--------|-------|
| Scalability | Good/Moderate/Poor | ... |
| Reliability | Good/Moderate/Poor | ... |
| Security | Good/Moderate/Poor | ... |
| Maintainability | Good/Moderate/Poor | ... |

### Concerns

1. **[Concern Name]** (severity: high|medium|low)
   - **Issue**: Description
   - **Impact**: What could go wrong
   - **Recommendation**: Suggested fix

### Trade-off Analysis

| Decision | Trade-off | Assessment |
|----------|-----------|------------|
| Decision 1 | Gained X, sacrificed Y | Appropriate/Questionable |

### Recommendations

1. **Priority 1**: ...
2. **Priority 2**: ...

### Architecture Decision Records (ADRs) Needed

If decisions aren't documented, recommend ADRs:

1. ADR-001: [Decision title]
   - Context: Why this decision matters
   - Decision: What was decided
   - Consequences: Trade-offs
```

## Example Output

```markdown
## Architecture Review: E-Commerce Platform

### Executive Summary
The system uses a microservices architecture that's generally well-designed but has concerning coupling between the Order and Inventory services. The authentication design is solid using OAuth2/OIDC, but authorization is inconsistent across services.

### Architecture Overview
- **Style**: Microservices (12 services)
- **Components**: User, Product, Order, Inventory, Payment, Notification, Gateway
- **Communication**: REST for sync, Kafka for async
- **Data Strategy**: Database per service (PostgreSQL + MongoDB)

### Quality Attribute Assessment

| Attribute | Rating | Notes |
|-----------|--------|-------|
| Scalability | Good | Each service scales independently |
| Reliability | Moderate | Missing circuit breakers |
| Security | Good | OAuth2/OIDC, service mesh mTLS |
| Maintainability | Moderate | Some services too coupled |

### Concerns

1. **Distributed Monolith Risk** (severity: high)
   - **Issue**: Order service makes sync calls to Inventory for every operation
   - **Impact**: If Inventory is down, Order is down
   - **Recommendation**: Use events for inventory checks, cache for reads

2. **Missing Resilience Patterns** (severity: medium)
   - **Issue**: No circuit breakers on external service calls
   - **Impact**: Cascade failures possible
   - **Recommendation**: Add circuit breakers (Hystrix/resilience4j)

3. **Inconsistent Authorization** (severity: medium)
   - **Issue**: Some services check roles, others check ownership, some both
   - **Impact**: Security gaps, confusion
   - **Recommendation**: Centralize authz policy (OPA or service mesh)

### Trade-off Analysis

| Decision | Trade-off | Assessment |
|----------|-----------|------------|
| Microservices | Flexibility vs complexity | Appropriate for team size (40 devs) |
| Kafka for events | Reliability vs latency | Good for eventual consistency use cases |
| MongoDB for Product | Flexibility vs consistency | Questionable - product data is relational |

### Recommendations

1. **Priority 1**: Decouple Order-Inventory with events
2. **Priority 2**: Add circuit breakers to all external calls
3. **Priority 3**: Centralize authorization policy

### Architecture Decision Records (ADRs) Needed

1. ADR-001: Service Communication Strategy
   - When to use sync vs async communication

2. ADR-002: Authorization Strategy
   - Standard approach to authorization across services
```
