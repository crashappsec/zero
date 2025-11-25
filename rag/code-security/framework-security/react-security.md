# React Security Patterns

## Overview

React applications face unique security challenges related to XSS, state management, and third-party dependencies.

## Common Vulnerabilities

### 1. Cross-Site Scripting (XSS)

#### Dangerous: dangerouslySetInnerHTML

```jsx
// VULNERABLE - XSS via dangerouslySetInnerHTML
function Comment({ html }) {
  return <div dangerouslySetInnerHTML={{ __html: html }} />;
}

// SECURE - Use DOMPurify for sanitization
import DOMPurify from 'dompurify';

function Comment({ html }) {
  const sanitized = DOMPurify.sanitize(html);
  return <div dangerouslySetInnerHTML={{ __html: sanitized }} />;
}

// BEST - Avoid dangerouslySetInnerHTML entirely
function Comment({ text }) {
  return <div>{text}</div>;  // React escapes by default
}
```

#### URL Injection via href

```jsx
// VULNERABLE - javascript: protocol XSS
function Link({ url, text }) {
  return <a href={url}>{text}</a>;
}

// SECURE - Validate URL protocol
function Link({ url, text }) {
  const isValidUrl = (u) => {
    try {
      const parsed = new URL(u);
      return ['http:', 'https:'].includes(parsed.protocol);
    } catch {
      return false;
    }
  };

  return <a href={isValidUrl(url) ? url : '#'}>{text}</a>;
}
```

### 2. State Exposure

#### Sensitive Data in State

```jsx
// VULNERABLE - Password in Redux store (visible in DevTools)
const authSlice = createSlice({
  name: 'auth',
  initialState: {
    user: null,
    password: '', // Don't store passwords!
    token: null
  }
});

// SECURE - Only store tokens, never passwords
const authSlice = createSlice({
  name: 'auth',
  initialState: {
    user: null,
    token: null,
    isAuthenticated: false
  }
});
```

### 3. Insecure Direct Object References

```jsx
// VULNERABLE - User ID from URL without authorization check
function UserProfile() {
  const { userId } = useParams();
  const { data } = useQuery(['user', userId], () => fetchUser(userId));
  return <Profile user={data} />;
}

// SECURE - Backend should verify authorization
// Frontend should also check current user
function UserProfile() {
  const { userId } = useParams();
  const { currentUser } = useAuth();
  const { data } = useQuery(['user', userId], () => fetchUser(userId));

  // Frontend check (backend must also verify!)
  if (data && data.id !== currentUser.id && !currentUser.isAdmin) {
    return <Unauthorized />;
  }

  return <Profile user={data} />;
}
```

### 4. Client-Side Authorization

```jsx
// VULNERABLE - Authorization only on frontend
function AdminPanel() {
  const { user } = useAuth();
  if (!user.isAdmin) return <Unauthorized />;

  return <AdminDashboard />;  // Data still fetched from API
}

// SECURE - Backend enforces authorization
// Frontend is just UX, not security
function AdminPanel() {
  const { user } = useAuth();
  const { data, error } = useQuery('admin-data', fetchAdminData);

  if (error?.status === 403) return <Unauthorized />;
  if (!user.isAdmin) return <Unauthorized />;

  return <AdminDashboard data={data} />;
}
```

### 5. Insecure Token Storage

```jsx
// VULNERABLE - Token in localStorage (XSS accessible)
localStorage.setItem('token', response.token);

// BETTER - HttpOnly cookie (server-side)
// The token should be set via Set-Cookie header with HttpOnly flag
// React just makes authenticated requests, doesn't handle token storage

// If localStorage is required, consider:
// 1. Short-lived access tokens
// 2. Token refresh mechanism
// 3. Proper CSP headers to mitigate XSS
```

## Security Best Practices

### Content Security Policy

```jsx
// In your server or meta tag
<meta
  httpEquiv="Content-Security-Policy"
  content="default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline';"
/>
```

### Form Handling

```jsx
// SECURE - CSRF protection with tokens
function LoginForm() {
  const [csrfToken, setCsrfToken] = useState('');

  useEffect(() => {
    // Fetch CSRF token from server
    fetch('/api/csrf-token')
      .then(r => r.json())
      .then(d => setCsrfToken(d.token));
  }, []);

  const handleSubmit = (e) => {
    e.preventDefault();
    fetch('/api/login', {
      method: 'POST',
      headers: {
        'X-CSRF-Token': csrfToken,
        'Content-Type': 'application/json'
      },
      credentials: 'include',
      body: JSON.stringify(formData)
    });
  };
}
```

### Environment Variables

```jsx
// VULNERABLE - Exposing sensitive keys
const API_KEY = process.env.REACT_APP_SECRET_API_KEY;
// All REACT_APP_ vars are bundled into client code!

// SECURE - Only expose public keys
const PUBLIC_KEY = process.env.REACT_APP_PUBLIC_KEY;
// Keep secrets server-side only
```

## Dependency Security

### Common Vulnerable Patterns

1. **Outdated React versions** - Update regularly
2. **Vulnerable dependencies** - Run `npm audit` regularly
3. **Typosquatting** - Verify package names carefully
4. **Malicious packages** - Use lockfiles, verify maintainers

### Security Headers

Ensure your server sets:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security: max-age=31536000`
- `Content-Security-Policy: ...`
