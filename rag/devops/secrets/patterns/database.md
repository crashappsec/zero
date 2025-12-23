# Database Credentials

## Connection Strings

### PostgreSQL
```
Pattern: postgres(ql)?://[^:]+:[^@]+@[^/]+/[^\s]+
Pattern: postgresql://[^:]+:[^@]+@[^/]+/[^\s]+
Example: postgresql://user:password@localhost:5432/mydb
Severity: critical
```

Components:
- Protocol: `postgresql://` or `postgres://`
- Username: before first `:`
- Password: between `:` and `@`
- Host: after `@`, before `/`
- Database: after `/`

### MySQL
```
Pattern: mysql://[^:]+:[^@]+@[^/]+/[^\s]+
Pattern: mysql\+pymysql://[^:]+:[^@]+@[^/]+/[^\s]+
Example: mysql://root:secret@localhost:3306/mydb
Severity: critical
```

### MongoDB
```
Pattern: mongodb(\+srv)?://[^:]+:[^@]+@[^\s]+
Example: mongodb+srv://user:pass@cluster.mongodb.net/db
Severity: critical
```

MongoDB Atlas uses `mongodb+srv://` protocol.

### Redis
```
Pattern: redis://[^:]*:[^@]+@[^/]+
Pattern: rediss://[^:]*:[^@]+@[^/]+ (TLS)
Example: redis://:password@localhost:6379
Severity: high
```

Redis URLs may omit username (empty before first `:`).

### SQL Server
```
Pattern: Server=[^;]+;.*Password=[^;]+
Pattern: Data Source=[^;]+;.*Password=[^;]+
Example: Server=myserver;Database=mydb;User Id=sa;Password=secret;
Severity: critical
```

ADO.NET connection string format.

### Oracle
```
Pattern: [^/]+/[^@]+@[^\s]+
Context: Oracle TNS connections
Pattern: jdbc:oracle:[^:]+:@[^:]+:[^:]+:[^\s]+
Severity: critical
```

---

## Environment Variables

### Common Patterns
```
DATABASE_URL=
DB_PASSWORD=
DB_PASS=
POSTGRES_PASSWORD=
MYSQL_PASSWORD=
MYSQL_ROOT_PASSWORD=
MONGO_PASSWORD=
REDIS_PASSWORD=
```

Severity: critical when values are present.

---

## Configuration Files

### Django (settings.py)
```python
DATABASES = {
    'default': {
        'PASSWORD': 'exposed_password',
    }
}
```

### Rails (database.yml)
```yaml
production:
  password: exposed_password
```

### Node.js (common patterns)
```javascript
password: process.env.DB_PASSWORD || 'fallback_password'
```

The fallback is the vulnerability.

### PHP (config files)
```php
$db_password = 'exposed_password';
define('DB_PASSWORD', 'exposed_password');
```

---

## ORM Configuration

### Prisma
```
DATABASE_URL="postgresql://user:password@host:5432/db"
```

### TypeORM
```typescript
{
  password: "exposed_password"
}
```

### Sequelize
```javascript
new Sequelize('database', 'username', 'password', {})
```

### SQLAlchemy
```python
engine = create_engine('postgresql://user:pass@host/db')
```

---

## Cloud Database Services

### AWS RDS
```
Pattern: [a-z]+-[a-z0-9]+\.[a-z0-9]+\.[a-z]+-[a-z]+-[0-9]\.rds\.amazonaws\.com
Context: Combined with credentials
Severity: informational (hostname only), critical (with credentials)
```

### Supabase
```
Pattern: db\.[a-z]+\.supabase\.co
Pattern: postgresql://postgres:[^@]+@db\.[a-z]+\.supabase\.co
Severity: critical
```

### PlanetScale
```
Pattern: aws\.connect\.psdb\.cloud
Pattern: mysql://[^:]+:[^@]+@[^/]+\.psdb\.cloud
Severity: critical
```

### MongoDB Atlas
```
Pattern: \.mongodb\.net
Pattern: mongodb\+srv://[^:]+:[^@]+@[^/]+\.mongodb\.net
Severity: critical
```

### Neon
```
Pattern: [a-z]+-[a-z]+-[0-9]+\.neon\.tech
Pattern: postgresql://[^:]+:[^@]+@[^/]+\.neon\.tech
Severity: critical
```

---

## Detection Notes

### High-Risk Files
- `.env`, `.env.local`, `.env.production`
- `config/database.yml`
- `settings.py`, `local_settings.py`
- `docker-compose.yml`
- `application.properties`
- `appsettings.json`

### False Positives
- `localhost` connections in development
- Example files (`.env.example`)
- Placeholder values (`your_password_here`)
- Test database URLs

### Password Patterns in Connection Strings
```
Pattern: ://[^:]+:([^@]+)@
```
Captures password between `:` and `@`.

### Security Considerations
- Connection strings expose host, potentially allowing network mapping
- Credentials should use secret management
- Rotate credentials if exposed in git history
