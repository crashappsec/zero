# Docker Compose Patterns

## File Names

### Standard Compose Files
- `docker-compose.yml`
- `docker-compose.yaml`
- `compose.yml`
- `compose.yaml`

### Environment-Specific Files
- `docker-compose.override.yml`
- `docker-compose.dev.yml`
- `docker-compose.prod.yml`
- `docker-compose.test.yml`
- `docker-compose.local.yml`

## Compose File Versions

### Version 3 (Recommended for Docker Compose v2)
```yaml
version: '3'
version: '3.8'
version: '3.9'
```

### Version 2 (Legacy)
```yaml
version: '2'
version: '2.4'
```

### No Version (Compose Spec)
```yaml
# Modern Compose Specification (no version needed)
services:
  app:
    image: myapp
```

## Service Definition Patterns

### Basic Service
```yaml
services:
  web:
    image: nginx:alpine
    ports:
      - "80:80"

  app:
    build: .
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
```

### Build Configuration
```yaml
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
      args:
        - BUILD_DATE=${BUILD_DATE}
        - VERSION=${VERSION}
      target: production
      cache_from:
        - myapp:cache
```

### Image Patterns
```yaml
services:
  db:
    image: postgres:15
  redis:
    image: redis:7-alpine
  app:
    image: myorg/myapp:latest
  web:
    image: nginx:1.25-alpine
```

### Port Mappings
```yaml
services:
  web:
    ports:
      - "8080:80"           # host:container
      - "443:443"
      - "127.0.0.1:8080:80" # bind to specific IP
      - "3000-3005:3000-3005" # port range
    expose:
      - "8080"              # only to other services
```

### Environment Variables
```yaml
services:
  app:
    environment:
      NODE_ENV: production
      DATABASE_URL: postgres://db:5432/mydb
      API_KEY: ${API_KEY}
    env_file:
      - .env
      - .env.local
```

### Volumes
```yaml
services:
  app:
    volumes:
      - ./src:/app/src       # bind mount
      - node_modules:/app/node_modules # named volume
      - /var/log              # anonymous volume
      - type: bind
        source: ./config
        target: /app/config
        read_only: true

volumes:
  node_modules:
  db_data:
    driver: local
```

### Networks
```yaml
services:
  web:
    networks:
      - frontend
      - backend
  db:
    networks:
      - backend

networks:
  frontend:
    driver: bridge
  backend:
    driver: bridge
    internal: true
```

### Dependencies
```yaml
services:
  web:
    depends_on:
      - db
      - redis

  app:
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_started
```

### Health Checks
```yaml
services:
  db:
    image: postgres:15
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 30s
```

### Command and Entrypoint
```yaml
services:
  app:
    command: npm start
    # or
    command: ["npm", "start"]
    # or
    entrypoint: /app/entrypoint.sh
    # or
    entrypoint: ["python", "app.py"]
```

### Working Directory
```yaml
services:
  app:
    working_dir: /app
```

### User
```yaml
services:
  app:
    user: "1000:1000"
    # or
    user: node
```

### Restart Policy
```yaml
services:
  app:
    restart: always
    # or: no, on-failure, unless-stopped
```

## Common Service Stacks

### LAMP Stack
```yaml
services:
  web:
    image: php:8.2-apache
    ports:
      - "80:80"
    volumes:
      - ./src:/var/www/html

  db:
    image: mysql:8
    environment:
      MYSQL_ROOT_PASSWORD: secret
      MYSQL_DATABASE: mydb
    volumes:
      - db_data:/var/lib/mysql

volumes:
  db_data:
```

### Node.js + PostgreSQL
```yaml
services:
  app:
    build: .
    ports:
      - "3000:3000"
    environment:
      DATABASE_URL: postgres://db:5432/mydb
    depends_on:
      - db

  db:
    image: postgres:15
    environment:
      POSTGRES_DB: mydb
      POSTGRES_PASSWORD: secret
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:
```

### Microservices
```yaml
services:
  api:
    build: ./api
    ports:
      - "8000:8000"

  web:
    build: ./web
    ports:
      - "3000:3000"
    depends_on:
      - api

  db:
    image: postgres:15
    volumes:
      - db_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
    depends_on:
      - web
      - api

volumes:
  db_data:
```

### Django + PostgreSQL + Redis
```yaml
services:
  web:
    build: .
    command: python manage.py runserver 0.0.0.0:8000
    volumes:
      - .:/code
    ports:
      - "8000:8000"
    depends_on:
      - db
      - redis
    environment:
      - DATABASE_URL=postgres://postgres:postgres@db:5432/postgres
      - REDIS_URL=redis://redis:6379/0

  db:
    image: postgres:15
    volumes:
      - postgres_data:/var/lib/postgresql/data
    environment:
      - POSTGRES_PASSWORD=postgres

  redis:
    image: redis:7-alpine

volumes:
  postgres_data:
```

### WordPress
```yaml
services:
  wordpress:
    image: wordpress:latest
    ports:
      - "8080:80"
    environment:
      WORDPRESS_DB_HOST: db
      WORDPRESS_DB_USER: wordpress
      WORDPRESS_DB_PASSWORD: wordpress
      WORDPRESS_DB_NAME: wordpress
    volumes:
      - wordpress:/var/www/html
    depends_on:
      - db

  db:
    image: mysql:8
    environment:
      MYSQL_DATABASE: wordpress
      MYSQL_USER: wordpress
      MYSQL_PASSWORD: wordpress
      MYSQL_ROOT_PASSWORD: rootpassword
    volumes:
      - db:/var/lib/mysql

volumes:
  wordpress:
  db:
```

## Advanced Patterns

### Secrets
```yaml
services:
  app:
    secrets:
      - db_password
      - api_key

secrets:
  db_password:
    file: ./secrets/db_password.txt
  api_key:
    external: true
```

### Configs
```yaml
services:
  app:
    configs:
      - source: app_config
        target: /app/config.yml

configs:
  app_config:
    file: ./config/app.yml
```

### Extends
```yaml
services:
  web:
    extends:
      file: common.yml
      service: base
    ports:
      - "80:80"
```

### Profiles
```yaml
services:
  app:
    image: myapp
    # always runs

  debug:
    image: myapp-debug
    profiles:
      - debug

  test:
    image: myapp-test
    profiles:
      - test
```

### Resource Limits
```yaml
services:
  app:
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 256M
```

### Logging
```yaml
services:
  app:
    logging:
      driver: json-file
      options:
        max-size: "10m"
        max-file: "3"
```

### Labels
```yaml
services:
  app:
    labels:
      com.example.description: "My application"
      com.example.version: "1.0"
```

## Docker Compose Commands

### Common Commands
```bash
docker-compose up
docker-compose up -d
docker-compose down
docker-compose down -v
docker-compose build
docker-compose logs
docker-compose ps
docker-compose exec app bash
docker-compose restart
docker-compose pull
docker-compose config
```

### Docker Compose v2 (Plugin)
```bash
docker compose up
docker compose down
docker compose build
docker compose logs
docker compose ps
```

## Detection Patterns

### File Content Patterns
```yaml
# Key indicators
version:
services:
volumes:
networks:
secrets:
configs:
```

### Common Service Names
- web, app, api, backend, frontend
- db, database, postgres, mysql, mongodb
- redis, cache, memcached
- nginx, apache, caddy
- worker, queue, celery
- test, dev, prod

### Common Image Patterns
- `postgres`, `mysql`, `mongodb`, `redis`
- `nginx`, `apache`, `caddy`
- `node`, `python`, `golang`, `openjdk`
- `wordpress`, `drupal`, `ghost`

## Security Considerations

### Secrets in Environment Variables (Anti-pattern)
```yaml
# BAD - Hardcoded secrets
environment:
  DB_PASSWORD: mysecretpassword
  API_KEY: hardcoded-key
```

### Better Approach
```yaml
# GOOD - Use secrets or env files
environment:
  DB_PASSWORD: ${DB_PASSWORD}
env_file:
  - .env.secret
```

### Privileged Mode (Caution)
```yaml
services:
  app:
    privileged: true  # Security risk
    cap_add:          # More granular
      - NET_ADMIN
```

## Detection Confidence

- **HIGH**: File named docker-compose.yml or compose.yml
- **HIGH**: Contains "version:" and "services:" keys
- **HIGH**: Contains common Docker Compose service definitions
- **MEDIUM**: Contains docker-compose in filename
- **LOW**: YAML file with service-like structure but no Docker markers
