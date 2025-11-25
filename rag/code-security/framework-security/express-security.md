# Express.js Security Patterns

## Overview

Express.js is minimal by design, requiring explicit security configurations. Many vulnerabilities arise from missing middleware or improper input handling.

## Common Vulnerabilities

### 1. SQL/NoSQL Injection

#### SQL Injection

```javascript
// VULNERABLE - String concatenation
app.get('/user', (req, res) => {
  const query = `SELECT * FROM users WHERE id = ${req.query.id}`;
  db.query(query);
});

// SECURE - Parameterized queries
app.get('/user', (req, res) => {
  const query = 'SELECT * FROM users WHERE id = ?';
  db.query(query, [req.query.id]);
});
```

#### NoSQL Injection (MongoDB)

```javascript
// VULNERABLE - Direct object from request
app.post('/login', (req, res) => {
  User.findOne({
    username: req.body.username,
    password: req.body.password  // Can inject { $gt: "" }
  });
});

// SECURE - Validate and sanitize
const mongoSanitize = require('express-mongo-sanitize');
app.use(mongoSanitize());

app.post('/login', (req, res) => {
  const username = String(req.body.username);
  const password = String(req.body.password);
  User.findOne({ username, password });
});
```

### 2. Cross-Site Scripting (XSS)

#### Reflected XSS

```javascript
// VULNERABLE - Reflecting user input
app.get('/search', (req, res) => {
  res.send(`<h1>Results for: ${req.query.q}</h1>`);
});

// SECURE - Escape output
const escapeHtml = require('escape-html');
app.get('/search', (req, res) => {
  res.send(`<h1>Results for: ${escapeHtml(req.query.q)}</h1>`);
});

// BEST - Use template engine with auto-escaping
app.set('view engine', 'ejs');  // EJS escapes by default with <%= %>
app.get('/search', (req, res) => {
  res.render('search', { query: req.query.q });
});
```

#### Stored XSS

```javascript
// VULNERABLE - Storing and displaying raw HTML
app.post('/comment', (req, res) => {
  Comment.create({ content: req.body.content });
});

// Later:
app.get('/comments', (req, res) => {
  const comments = await Comment.find();
  res.send(comments.map(c => `<div>${c.content}</div>`).join(''));
});

// SECURE - Sanitize HTML
const createDOMPurify = require('dompurify');
const { JSDOM } = require('jsdom');
const DOMPurify = createDOMPurify(new JSDOM('').window);

app.post('/comment', (req, res) => {
  const clean = DOMPurify.sanitize(req.body.content);
  Comment.create({ content: clean });
});
```

### 3. Command Injection

```javascript
// VULNERABLE - User input in shell command
const { exec } = require('child_process');

app.get('/ping', (req, res) => {
  exec(`ping -c 4 ${req.query.host}`, (err, stdout) => {
    res.send(stdout);
  });
});

// SECURE - Use execFile with array arguments
const { execFile } = require('child_process');

app.get('/ping', (req, res) => {
  // Validate input
  const hostRegex = /^[a-zA-Z0-9.-]+$/;
  if (!hostRegex.test(req.query.host)) {
    return res.status(400).send('Invalid host');
  }

  execFile('ping', ['-c', '4', req.query.host], (err, stdout) => {
    res.send(stdout);
  });
});
```

### 4. Path Traversal

```javascript
// VULNERABLE - Direct file access
app.get('/files/:name', (req, res) => {
  res.sendFile(`/uploads/${req.params.name}`);
  // Attacker: /files/../../../etc/passwd
});

// SECURE - Validate and resolve path
const path = require('path');

app.get('/files/:name', (req, res) => {
  const uploadsDir = path.resolve('/uploads');
  const filePath = path.resolve(uploadsDir, req.params.name);

  // Ensure path is within uploads directory
  if (!filePath.startsWith(uploadsDir)) {
    return res.status(403).send('Forbidden');
  }

  res.sendFile(filePath);
});
```

### 5. Missing Security Headers

```javascript
// VULNERABLE - No security headers
const app = express();
app.listen(3000);

// SECURE - Use helmet
const helmet = require('helmet');

const app = express();
app.use(helmet());  // Sets many security headers

// Or configure individually
app.use(helmet({
  contentSecurityPolicy: {
    directives: {
      defaultSrc: ["'self'"],
      scriptSrc: ["'self'"],
      styleSrc: ["'self'", "'unsafe-inline'"],
      imgSrc: ["'self'", "data:", "https:"],
    },
  },
  hsts: {
    maxAge: 31536000,
    includeSubDomains: true,
    preload: true
  }
}));
```

### 6. CSRF Vulnerabilities

```javascript
// VULNERABLE - No CSRF protection
app.post('/transfer', (req, res) => {
  transferMoney(req.body.to, req.body.amount);
});

// SECURE - Use csurf middleware
const csrf = require('csurf');
const csrfProtection = csrf({ cookie: true });

app.get('/transfer', csrfProtection, (req, res) => {
  res.render('transfer', { csrfToken: req.csrfToken() });
});

app.post('/transfer', csrfProtection, (req, res) => {
  // Token validated automatically
  transferMoney(req.body.to, req.body.amount);
});
```

### 7. Session Security

```javascript
// VULNERABLE - Insecure session config
app.use(session({
  secret: 'keyboard cat',  // Weak secret
  cookie: {}  // Missing secure options
}));

// SECURE - Proper session configuration
app.use(session({
  secret: process.env.SESSION_SECRET,  // Strong, from env
  name: 'sessionId',  // Change default name
  resave: false,
  saveUninitialized: false,
  cookie: {
    secure: true,        // HTTPS only
    httpOnly: true,      // No JS access
    sameSite: 'strict',  // CSRF protection
    maxAge: 3600000      // 1 hour
  }
}));
```

### 8. Rate Limiting

```javascript
// VULNERABLE - No rate limiting
app.post('/login', (req, res) => {
  // Brute force possible
});

// SECURE - Add rate limiting
const rateLimit = require('express-rate-limit');

const loginLimiter = rateLimit({
  windowMs: 15 * 60 * 1000,  // 15 minutes
  max: 5,  // 5 attempts
  message: 'Too many login attempts'
});

app.post('/login', loginLimiter, (req, res) => {
  // Protected from brute force
});
```

### 9. Information Disclosure

```javascript
// VULNERABLE - Exposing stack traces
app.use((err, req, res, next) => {
  res.status(500).json({ error: err.stack });
});

// SECURE - Generic error in production
app.use((err, req, res, next) => {
  console.error(err);  // Log internally

  if (process.env.NODE_ENV === 'production') {
    res.status(500).json({ error: 'Internal server error' });
  } else {
    res.status(500).json({ error: err.message, stack: err.stack });
  }
});

// Also disable X-Powered-By
app.disable('x-powered-by');
// Or use helmet which does this automatically
```

## Security Middleware Stack

```javascript
const express = require('express');
const helmet = require('helmet');
const rateLimit = require('express-rate-limit');
const mongoSanitize = require('express-mongo-sanitize');
const xss = require('xss-clean');
const hpp = require('hpp');
const cors = require('cors');

const app = express();

// Security headers
app.use(helmet());

// CORS
app.use(cors({
  origin: 'https://yourdomain.com',
  credentials: true
}));

// Rate limiting
app.use(rateLimit({
  windowMs: 15 * 60 * 1000,
  max: 100
}));

// Body parsing with size limit
app.use(express.json({ limit: '10kb' }));
app.use(express.urlencoded({ extended: true, limit: '10kb' }));

// Data sanitization
app.use(mongoSanitize());  // NoSQL injection
app.use(xss());            // XSS

// Prevent parameter pollution
app.use(hpp());

// Trust proxy (if behind reverse proxy)
app.set('trust proxy', 1);
```

## Express Security Checklist

- [ ] helmet() middleware enabled
- [ ] Rate limiting configured
- [ ] CORS properly restricted
- [ ] CSRF protection for forms
- [ ] Input validation on all routes
- [ ] Parameterized database queries
- [ ] Secure session configuration
- [ ] HTTPS enforced
- [ ] Error messages sanitized
- [ ] File upload validation
- [ ] Dependencies audited (`npm audit`)
