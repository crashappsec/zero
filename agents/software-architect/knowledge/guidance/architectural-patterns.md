# Architectural Patterns Guide

## Choosing an Architecture

### Decision Framework

```
┌─────────────────────────────────────────────────────────────────┐
│                 Architecture Decision Tree                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Team size < 10 AND domain simple?                              │
│  ├── Yes → Monolith                                             │
│  └── No ↓                                                       │
│                                                                 │
│  Need independent scaling/deployment per component?              │
│  ├── Yes → Microservices                                        │
│  └── No ↓                                                       │
│                                                                 │
│  Variable load with long idle periods?                          │
│  ├── Yes → Serverless                                           │
│  └── No ↓                                                       │
│                                                                 │
│  Need team boundaries without distributed complexity?            │
│  ├── Yes → Modular Monolith                                     │
│  └── No → Evaluate trade-offs case by case                      │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Architecture Patterns

### Monolith

A single deployable unit containing all application functionality.

**When to choose:**
- Early stage, exploring product-market fit
- Small team (< 10 developers)
- Simple domain
- Need for fast iteration

**Structure:**
```
app/
├── src/
│   ├── controllers/
│   ├── services/
│   ├── models/
│   └── utils/
├── package.json
└── Dockerfile
```

**Key practices:**
- Keep code modular even in monolith
- Use interfaces between components
- Plan for eventual extraction

---

### Modular Monolith

Monolith with enforced module boundaries, preparing for potential microservices.

**When to choose:**
- Growing team needing ownership boundaries
- Complex domain but want monolith simplicity
- Preparing for eventual microservices

**Structure:**
```
app/
├── modules/
│   ├── users/
│   │   ├── api/
│   │   ├── domain/
│   │   ├── infrastructure/
│   │   └── module.ts
│   ├── orders/
│   └── payments/
├── shared/
└── package.json
```

**Key practices:**
- Each module owns its data (schema/tables)
- Modules communicate via defined interfaces
- No direct database access across modules
- Consider module-per-team ownership

---

### Microservices

Independently deployable services organized around business capabilities.

**When to choose:**
- Large team (> 20 developers)
- Clear bounded contexts
- Different scaling needs per component
- Need technology flexibility

**Structure:**
```
services/
├── user-service/
│   ├── src/
│   ├── package.json
│   └── Dockerfile
├── order-service/
├── payment-service/
└── docker-compose.yml
```

**Key practices:**
- One database per service
- Async communication preferred
- API contracts (OpenAPI, protobuf)
- Correlation IDs for tracing
- Circuit breakers for resilience

**Anti-patterns to avoid:**
- Distributed monolith (tight coupling)
- Shared databases
- Synchronous chains

---

### Serverless

Functions executed in response to events, managed by cloud provider.

**When to choose:**
- Event-driven workloads
- Variable traffic with idle periods
- Quick to market priority
- Cost optimization for low-traffic

**Structure:**
```
functions/
├── createUser/
│   └── handler.ts
├── processOrder/
│   └── handler.ts
├── serverless.yml
└── package.json
```

**Key practices:**
- Keep functions focused (single purpose)
- Minimize cold start (small bundles, warm-up)
- Use managed services (DynamoDB, SQS)
- Handle idempotency

**Limitations:**
- Execution time limits (15 min max)
- Cold start latency
- Vendor lock-in

---

### Event-Driven Architecture

Services communicate through events, enabling loose coupling and async processing.

**When to choose:**
- Async workflows
- Audit/compliance requirements
- High-throughput processing
- System integration

**Structure:**
```
                    ┌─────────────┐
                    │ Event Bus   │
                    │ (Kafka/SQS) │
                    └──────┬──────┘
           ┌───────────────┼───────────────┐
           │               │               │
     ┌─────▼─────┐   ┌─────▼─────┐   ┌─────▼─────┐
     │  Orders   │   │ Inventory │   │  Notify   │
     │  Service  │   │  Service  │   │  Service  │
     └───────────┘   └───────────┘   └───────────┘
```

**Key practices:**
- Design events carefully (schema evolution)
- Handle duplicate events (idempotency)
- Plan for eventual consistency
- Use dead letter queues
- Implement event replay

---

## Integration Patterns

### API Gateway

Single entry point for all client requests.

```
Client → API Gateway → Service A
                    → Service B
                    → Service C
```

**Responsibilities:**
- Routing
- Authentication
- Rate limiting
- Request transformation
- Response aggregation

### Service Mesh

Infrastructure layer handling service-to-service communication.

**Provides:**
- mTLS between services
- Load balancing
- Circuit breaking
- Observability

**Tools:** Istio, Linkerd, Consul Connect

### Backend for Frontend (BFF)

Dedicated backend per frontend type.

```
Web App → Web BFF ─┐
                   ├→ Services
Mobile → Mobile BFF┘
```

**Benefits:**
- Optimized for each client
- Frontend team ownership
- Decouples frontend from services

---

## Data Architecture Patterns

### Database per Service

Each service owns its data completely.

**Benefits:**
- Loose coupling
- Independent scaling
- Technology freedom

**Challenges:**
- Data consistency
- Cross-service queries
- Data duplication

### Event Sourcing

Store events, not current state.

```
EventStore: [UserCreated, NameChanged, EmailChanged]
                         ↓
                  Current State: {name: "John", email: "john@example.com"}
```

**Benefits:**
- Complete audit trail
- Time travel queries
- Event replay for recovery

### CQRS (Command Query Responsibility Segregation)

Separate read and write models.

```
Commands → Write Model → Event Store
Queries  → Read Model  → Read Database (denormalized)
```

**Benefits:**
- Optimize read and write independently
- Scale reads separately
- Complex queries on read side

---

## Migration Strategies

### Strangler Fig Pattern

Gradually replace legacy system with new components.

```
Phase 1: Route some traffic to new service
         ┌─────────────┐
         │   Router    │
         └──────┬──────┘
                │
       ┌────────┴────────┐
       ▼                 ▼
   Legacy (90%)     New (10%)

Phase N: Complete migration
         All traffic → New Service
```

### Branch by Abstraction

1. Create abstraction over component to replace
2. Implement new version behind abstraction
3. Gradually switch traffic
4. Remove old implementation

---

## Quality Attributes

### Scalability
- **Horizontal**: Add more instances
- **Vertical**: Add more resources to instance
- **Data**: Sharding, partitioning, read replicas

### Reliability
- **Fault tolerance**: Survive component failures
- **Redundancy**: Eliminate single points of failure
- **Graceful degradation**: Partial functionality over complete failure

### Security
- **Defense in depth**: Multiple security layers
- **Least privilege**: Minimum required access
- **Zero trust**: Verify everything

### Maintainability
- **Modularity**: Clear boundaries
- **Testability**: Easy to test in isolation
- **Observability**: Logs, metrics, traces
