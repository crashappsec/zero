# API Design Guide

## REST API Principles

### Resource-Oriented Design

Design APIs around resources (nouns), not actions (verbs).

```
# Good - Resources
GET    /users           # List users
GET    /users/123       # Get user
POST   /users           # Create user
PUT    /users/123       # Replace user
PATCH  /users/123       # Update user
DELETE /users/123       # Delete user

# Bad - Actions
GET    /getUsers
POST   /createUser
POST   /deleteUser
```

### URL Structure

```
/api/v1/{resource}/{id}/{sub-resource}/{sub-id}

Examples:
/api/v1/users
/api/v1/users/123
/api/v1/users/123/orders
/api/v1/users/123/orders/456
```

**Guidelines:**
- Use lowercase with hyphens: `/order-items` not `/orderItems`
- Use plural nouns: `/users` not `/user`
- Limit nesting to 2-3 levels
- Use query params for filtering: `/users?status=active`

### HTTP Methods

| Method | Purpose | Idempotent | Safe |
|--------|---------|------------|------|
| GET | Retrieve | Yes | Yes |
| POST | Create | No | No |
| PUT | Replace | Yes | No |
| PATCH | Update | No* | No |
| DELETE | Remove | Yes | No |

### Status Codes

**Success (2xx):**
- `200 OK` - Successful GET, PUT, PATCH
- `201 Created` - Successful POST with Location header
- `204 No Content` - Successful DELETE

**Client Errors (4xx):**
- `400 Bad Request` - Malformed request
- `401 Unauthorized` - Missing/invalid authentication
- `403 Forbidden` - Authenticated but not authorized
- `404 Not Found` - Resource doesn't exist
- `409 Conflict` - State conflict
- `422 Unprocessable Entity` - Validation error
- `429 Too Many Requests` - Rate limited

**Server Errors (5xx):**
- `500 Internal Server Error` - Unexpected error
- `503 Service Unavailable` - Temporary outage

### Response Format

Use consistent envelope:

```json
{
  "data": {
    "id": "123",
    "type": "user",
    "attributes": {
      "name": "John Doe",
      "email": "john@example.com"
    }
  },
  "meta": {
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

For collections:

```json
{
  "data": [...],
  "meta": {
    "total": 100,
    "page": 1,
    "per_page": 20
  },
  "links": {
    "self": "/api/v1/users?page=1",
    "next": "/api/v1/users?page=2",
    "prev": null
  }
}
```

For errors:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request parameters",
    "details": [
      {
        "field": "email",
        "message": "Invalid email format"
      }
    ]
  }
}
```

### Pagination

**Offset-based** (simple, allows jumping to page):
```
GET /users?limit=20&offset=40
```

**Cursor-based** (consistent with changing data):
```
GET /users?limit=20&cursor=eyJpZCI6MTIzfQ
```

### Filtering & Sorting

```
GET /users?status=active&role=admin           # Simple filter
GET /users?filter[status]=active              # Bracketed filter
GET /users?sort=created_at                    # Ascending
GET /users?sort=-created_at                   # Descending
GET /users?sort=status,-created_at            # Multiple
```

### Versioning

**URL path** (explicit, easy to route):
```
/api/v1/users
/api/v2/users
```

**Header** (cleaner URLs, harder to test):
```
Accept: application/vnd.api+json;version=1
```

## GraphQL API Principles

### Schema Design

```graphql
type User {
  id: ID!
  email: String!
  name: String
  orders(first: Int, after: String): OrderConnection!
  createdAt: DateTime!
}

type OrderConnection {
  edges: [OrderEdge!]!
  pageInfo: PageInfo!
}

type OrderEdge {
  cursor: String!
  node: Order!
}
```

### Query Design

```graphql
type Query {
  # Single resource
  user(id: ID!): User

  # Collection with pagination
  users(
    first: Int
    after: String
    filter: UserFilter
  ): UserConnection!

  # Search
  searchUsers(query: String!): UserConnection!
}

input UserFilter {
  status: UserStatus
  role: Role
  createdAfter: DateTime
}
```

### Mutation Design

```graphql
type Mutation {
  createUser(input: CreateUserInput!): CreateUserPayload!
  updateUser(id: ID!, input: UpdateUserInput!): UpdateUserPayload!
  deleteUser(id: ID!): DeleteUserPayload!
}

input CreateUserInput {
  email: String!
  name: String!
  role: Role
}

type CreateUserPayload {
  user: User
  errors: [Error!]
}
```

### N+1 Prevention

Use DataLoader to batch requests:

```javascript
// Without DataLoader: N+1 queries
const resolvers = {
  Order: {
    user: (order) => db.users.findById(order.userId) // Called N times!
  }
};

// With DataLoader: 1 batched query
const userLoader = new DataLoader(ids => db.users.findByIds(ids));

const resolvers = {
  Order: {
    user: (order) => userLoader.load(order.userId) // Batched!
  }
};
```

## Error Handling

### Error Response Structure

```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "User with ID 123 not found",
    "details": {
      "resource": "user",
      "id": "123"
    },
    "request_id": "req_abc123"
  }
}
```

### Error Codes

Define application-specific error codes:

| Code | HTTP | Description |
|------|------|-------------|
| VALIDATION_ERROR | 422 | Input validation failed |
| RESOURCE_NOT_FOUND | 404 | Resource doesn't exist |
| RESOURCE_EXISTS | 409 | Duplicate resource |
| UNAUTHORIZED | 401 | Auth required |
| FORBIDDEN | 403 | Permission denied |
| RATE_LIMITED | 429 | Too many requests |
| INTERNAL_ERROR | 500 | Unexpected error |

### Validation Errors

Return all validation errors at once:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      {"field": "email", "code": "INVALID_FORMAT", "message": "Invalid email"},
      {"field": "age", "code": "OUT_OF_RANGE", "message": "Must be 18-120"}
    ]
  }
}
```

## Security

### Authentication

```javascript
// JWT in Authorization header
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...

// Validate and extract user
const token = req.headers.authorization?.split(' ')[1];
const user = jwt.verify(token, SECRET);
req.user = user;
```

### Authorization

```javascript
// Role-based
if (!user.roles.includes('admin')) {
  throw new ForbiddenError('Admin access required');
}

// Resource-based
const post = await db.posts.findById(id);
if (post.authorId !== user.id) {
  throw new ForbiddenError('Not the author');
}
```

### Rate Limiting

```javascript
// Token bucket: 100 requests per minute
const limiter = rateLimit({
  windowMs: 60 * 1000,
  max: 100,
  message: { error: { code: 'RATE_LIMITED' } }
});
```

### Input Validation

Always validate:
- Type (string, number, array)
- Format (email, URL, UUID)
- Range (min/max length, value bounds)
- Business rules (unique, exists)

```javascript
const schema = z.object({
  email: z.string().email(),
  age: z.number().min(18).max(120),
  role: z.enum(['user', 'admin'])
});

const data = schema.parse(req.body);
```

## Performance

### Caching

```javascript
// Cache-Control headers
res.set('Cache-Control', 'public, max-age=300'); // 5 min cache
res.set('Cache-Control', 'private, no-cache');    // No caching
res.set('ETag', hash(data));                      // Conditional requests
```

### Compression

```javascript
app.use(compression());
```

### Connection Pooling

```javascript
const pool = new Pool({
  max: 20,
  idleTimeoutMillis: 30000,
  connectionTimeoutMillis: 2000
});
```

### Pagination Limits

```javascript
const limit = Math.min(req.query.limit || 20, 100); // Max 100
```
