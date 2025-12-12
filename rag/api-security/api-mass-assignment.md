# API Mass Assignment

Detection patterns for mass assignment and parameter pollution vulnerabilities.

## OWASP API Security Top 10

- **API3:2023** - Broken Object Property Level Authorization
- **API6:2023** - Unrestricted Access to Sensitive Business Flows

## Patterns

### Direct Request Body Assignment

CATEGORY: api-mass-assignment
SEVERITY: high
CONFIDENCE: 90
CWE: CWE-915
OWASP: API3:2023

Direct body to create/update (JavaScript/TypeScript):
```
PATTERN: \.(create|update|updateOne|findOneAndUpdate|save)\s*\(\s*req\.body\s*\)
LANGUAGES: javascript, typescript
```

Spread operator with request body:
```
PATTERN: \{\s*\.\.\.req\.body\s*[,}]
LANGUAGES: javascript, typescript
```

Object.assign with request body:
```
PATTERN: Object\.assign\s*\([^,]+,\s*req\.body
LANGUAGES: javascript, typescript
```

### Python Mass Assignment

CATEGORY: api-mass-assignment
SEVERITY: high
CONFIDENCE: 90
CWE: CWE-915
OWASP: API3:2023

SQLAlchemy update with request data:
```
PATTERN: \.(update|filter)\s*\([^)]*\)\.(update|values)\s*\(\s*\*\*request\.(json|form)
LANGUAGES: python
```

Django update with request data:
```
PATTERN: \.(create|update|update_or_create)\s*\(\s*\*\*request\.(POST|data|json)
LANGUAGES: python
```

Pydantic model from request without validation:
```
PATTERN: \w+Model\s*\(\s*\*\*request\.(json|data)\s*\)
LANGUAGES: python
```

### Mongoose Mass Assignment

CATEGORY: api-mass-assignment
SEVERITY: high
CONFIDENCE: 90
CWE: CWE-915
OWASP: API3:2023

Mongoose update without field restriction:
```
PATTERN: (findByIdAndUpdate|findOneAndUpdate|updateOne|updateMany)\s*\([^,]+,\s*req\.body
LANGUAGES: javascript, typescript
```

New model from request body:
```
PATTERN: new\s+\w+Model\s*\(\s*req\.body\s*\)
LANGUAGES: javascript, typescript
```

### Sequelize Mass Assignment

CATEGORY: api-mass-assignment
SEVERITY: high
CONFIDENCE: 90
CWE: CWE-915
OWASP: API3:2023

Sequelize create/update from body:
```
PATTERN: \w+\.(create|update|bulkCreate)\s*\(\s*req\.body
LANGUAGES: javascript, typescript
```

### Prisma Mass Assignment

CATEGORY: api-mass-assignment
SEVERITY: high
CONFIDENCE: 90
CWE: CWE-915
OWASP: API3:2023

Prisma create/update from body:
```
PATTERN: prisma\.\w+\.(create|update|upsert)\s*\(\s*\{\s*data\s*:\s*req\.body
LANGUAGES: javascript, typescript
```

### Ruby on Rails Mass Assignment

CATEGORY: api-mass-assignment
SEVERITY: high
CONFIDENCE: 90
CWE: CWE-915
OWASP: API3:2023

Unpermitted params to create/update:
```
PATTERN: \.(create|update|new)\s*\(\s*params(?!\.(permit|require))
LANGUAGES: ruby
```

### Sensitive Field Modification

CATEGORY: api-mass-assignment
SEVERITY: critical
CONFIDENCE: 85
CWE: CWE-915
OWASP: API3:2023

Role/permission fields in request:
```
PATTERN: (role|isAdmin|is_admin|admin|permission|permissions|privilege)\s*[=:]\s*req\.(body|params|query)
LANGUAGES: javascript, typescript, python
```

Account status modification:
```
PATTERN: (status|active|verified|approved|enabled)\s*[=:]\s*req\.(body|params|query)
LANGUAGES: javascript, typescript, python
```

### HTTP Parameter Pollution

CATEGORY: api-mass-assignment
SEVERITY: medium
CONFIDENCE: 80
CWE: CWE-235
OWASP: API3:2023

Array parameters without validation:
```
PATTERN: req\.(query|params)\[['"][^'"]+['"]\]\s*(?!\.filter|\.map|\.every|\.some)
LANGUAGES: javascript, typescript
```

### GraphQL Mass Assignment

CATEGORY: api-mass-assignment
SEVERITY: high
CONFIDENCE: 85
CWE: CWE-915
OWASP: API3:2023

GraphQL input directly to database:
```
PATTERN: (args|input)\s*=>\s*\w+\.(create|update)\s*\(\s*(args|input)
LANGUAGES: javascript, typescript
```

### Prototype Pollution

CATEGORY: api-mass-assignment
SEVERITY: critical
CONFIDENCE: 90
CWE: CWE-1321
OWASP: API3:2023

Object merge with user input:
```
PATTERN: (merge|extend|assign|defaults)\s*\([^,]+,\s*req\.(body|params|query)
LANGUAGES: javascript, typescript
```

Lodash merge vulnerability:
```
PATTERN: _\.(merge|defaultsDeep|set)\s*\([^,]+,\s*req\.(body|params|query)
LANGUAGES: javascript, typescript
```

### Field Blacklist vs Whitelist

CATEGORY: api-mass-assignment
SEVERITY: medium
CONFIDENCE: 75
CWE: CWE-915
OWASP: API3:2023

Blacklist approach (weaker):
```
PATTERN: delete\s+req\.body\.(password|role|admin|isAdmin)
LANGUAGES: javascript, typescript
```

### Missing Input Validation Schema

CATEGORY: api-mass-assignment
SEVERITY: medium
CONFIDENCE: 70
CWE: CWE-20
OWASP: API3:2023

Route without validation middleware:
```
PATTERN: app\.(post|put|patch)\s*\([^)]+,\s*(?!.*validate|validator|schema|joi|yup|zod).*\(req
LANGUAGES: javascript, typescript
```

## Remediation Examples

### Safe Patterns

Allow-list specific fields (JavaScript):
```javascript
// SAFE: Only allow specific fields
const { name, email, bio } = req.body;
await User.update({ name, email, bio }, { where: { id: req.params.id } });
```

Use DTOs/Validation (TypeScript):
```typescript
// SAFE: Use class-validator/class-transformer
@Post()
async create(@Body() createUserDto: CreateUserDto) {
  return this.userService.create(createUserDto);
}
```

Mongoose strict schemas:
```javascript
// SAFE: Enable strict mode
const userSchema = new Schema({ name: String }, { strict: true });
```

## References

- [OWASP Mass Assignment](https://cheatsheetseries.owasp.org/cheatsheets/Mass_Assignment_Cheat_Sheet.html)
- [CWE-915: Improperly Controlled Modification](https://cwe.mitre.org/data/definitions/915.html)
- [CWE-1321: Prototype Pollution](https://cwe.mitre.org/data/definitions/1321.html)
