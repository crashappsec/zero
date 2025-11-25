# Django Security Patterns

## Overview

Django has strong security defaults but misconfigurations and bypasses can introduce vulnerabilities.

## Common Vulnerabilities

### 1. SQL Injection

#### Raw SQL Queries

```python
# VULNERABLE - String formatting in raw SQL
def get_user(request):
    user_id = request.GET.get('id')
    user = User.objects.raw(f"SELECT * FROM auth_user WHERE id = {user_id}")
    return user

# VULNERABLE - String concatenation
cursor.execute("SELECT * FROM users WHERE name = '" + name + "'")

# SECURE - Parameterized queries
def get_user(request):
    user_id = request.GET.get('id')
    user = User.objects.raw("SELECT * FROM auth_user WHERE id = %s", [user_id])
    return user

# BEST - Use ORM
def get_user(request):
    user_id = request.GET.get('id')
    user = User.objects.get(id=user_id)
    return user
```

#### Extra() and RawSQL

```python
# VULNERABLE - Unvalidated input in extra()
User.objects.extra(where=[f"username = '{username}'"])

# SECURE - Use params
User.objects.extra(where=["username = %s"], params=[username])

# BEST - Use filter()
User.objects.filter(username=username)
```

### 2. Cross-Site Scripting (XSS)

#### Template Auto-escaping

```python
# VULNERABLE - Marking content as safe without sanitization
from django.utils.safestring import mark_safe

def show_content(request):
    content = request.POST.get('content')
    return render(request, 'page.html', {'content': mark_safe(content)})

# VULNERABLE - Using |safe filter on user input
# In template: {{ user_content|safe }}

# SECURE - Let Django escape automatically
def show_content(request):
    content = request.POST.get('content')
    return render(request, 'page.html', {'content': content})
# Template: {{ content }}  # Auto-escaped

# If HTML needed, sanitize first
import bleach

def show_content(request):
    content = request.POST.get('content')
    clean_content = bleach.clean(content, tags=['p', 'b', 'i'])
    return render(request, 'page.html', {'content': mark_safe(clean_content)})
```

### 3. CSRF Vulnerabilities

```python
# VULNERABLE - Exempting CSRF protection
from django.views.decorators.csrf import csrf_exempt

@csrf_exempt  # Don't do this for state-changing operations!
def update_profile(request):
    # ...

# SECURE - Use CSRF token properly
# In views - CSRF is enabled by default
def update_profile(request):
    # CSRF middleware handles protection
    pass

# In templates
<form method="post">
    {% csrf_token %}
    <!-- form fields -->
</form>

# In AJAX
const csrftoken = document.querySelector('[name=csrfmiddlewaretoken]').value;
fetch('/api/update/', {
    method: 'POST',
    headers: {'X-CSRFToken': csrftoken},
    body: data
});
```

### 4. Authentication Issues

#### Hardcoded Secrets

```python
# VULNERABLE - Hardcoded SECRET_KEY
SECRET_KEY = 'django-insecure-abc123def456'

# SECURE - Environment variable
import os
SECRET_KEY = os.environ.get('DJANGO_SECRET_KEY')

# Or use django-environ
import environ
env = environ.Env()
SECRET_KEY = env('SECRET_KEY')
```

#### Weak Password Validation

```python
# VULNERABLE - No password validation
AUTH_PASSWORD_VALIDATORS = []

# SECURE - Strong validation
AUTH_PASSWORD_VALIDATORS = [
    {'NAME': 'django.contrib.auth.password_validation.UserAttributeSimilarityValidator'},
    {'NAME': 'django.contrib.auth.password_validation.MinimumLengthValidator',
     'OPTIONS': {'min_length': 12}},
    {'NAME': 'django.contrib.auth.password_validation.CommonPasswordValidator'},
    {'NAME': 'django.contrib.auth.password_validation.NumericPasswordValidator'},
]
```

### 5. Insecure Direct Object References (IDOR)

```python
# VULNERABLE - No ownership check
def view_document(request, doc_id):
    doc = Document.objects.get(id=doc_id)
    return render(request, 'document.html', {'doc': doc})

# SECURE - Verify ownership
def view_document(request, doc_id):
    doc = get_object_or_404(Document, id=doc_id, owner=request.user)
    return render(request, 'document.html', {'doc': doc})

# Or using permissions
from django.contrib.auth.decorators import permission_required

@permission_required('documents.view_document')
def view_document(request, doc_id):
    doc = get_object_or_404(Document, id=doc_id)
    if not request.user.has_perm('documents.view_document', doc):
        raise PermissionDenied
    return render(request, 'document.html', {'doc': doc})
```

### 6. Mass Assignment

```python
# VULNERABLE - All fields from request
def update_user(request):
    user = request.user
    for key, value in request.POST.items():
        setattr(user, key, value)  # Can set is_superuser!
    user.save()

# SECURE - Explicit fields
def update_user(request):
    user = request.user
    user.first_name = request.POST.get('first_name', user.first_name)
    user.last_name = request.POST.get('last_name', user.last_name)
    user.save()

# BEST - Use forms with explicit fields
class UserUpdateForm(forms.ModelForm):
    class Meta:
        model = User
        fields = ['first_name', 'last_name', 'email']  # Whitelist
```

### 7. File Upload Vulnerabilities

```python
# VULNERABLE - No validation
def upload_file(request):
    file = request.FILES['file']
    with open(f'/uploads/{file.name}', 'wb') as f:
        for chunk in file.chunks():
            f.write(chunk)

# SECURE - Validate file type and sanitize name
import os
from django.core.files.storage import FileSystemStorage
from django.utils.text import get_valid_filename
import magic

ALLOWED_TYPES = ['image/jpeg', 'image/png', 'application/pdf']
MAX_SIZE = 10 * 1024 * 1024  # 10MB

def upload_file(request):
    file = request.FILES['file']

    # Check size
    if file.size > MAX_SIZE:
        raise ValidationError("File too large")

    # Check MIME type (not just extension)
    mime = magic.from_buffer(file.read(1024), mime=True)
    file.seek(0)
    if mime not in ALLOWED_TYPES:
        raise ValidationError("Invalid file type")

    # Sanitize filename
    safe_name = get_valid_filename(file.name)

    # Use Django's storage
    fs = FileSystemStorage(location='/uploads/')
    fs.save(safe_name, file)
```

## Security Settings

### Production Settings

```python
# settings/production.py

DEBUG = False
ALLOWED_HOSTS = ['yourdomain.com', 'www.yourdomain.com']

# Security middleware
MIDDLEWARE = [
    'django.middleware.security.SecurityMiddleware',
    # ... other middleware
]

# HTTPS settings
SECURE_SSL_REDIRECT = True
SECURE_PROXY_SSL_HEADER = ('HTTP_X_FORWARDED_PROTO', 'https')
SESSION_COOKIE_SECURE = True
CSRF_COOKIE_SECURE = True

# HSTS
SECURE_HSTS_SECONDS = 31536000
SECURE_HSTS_INCLUDE_SUBDOMAINS = True
SECURE_HSTS_PRELOAD = True

# Other security headers
SECURE_CONTENT_TYPE_NOSNIFF = True
SECURE_BROWSER_XSS_FILTER = True
X_FRAME_OPTIONS = 'DENY'

# Session security
SESSION_COOKIE_HTTPONLY = True
SESSION_COOKIE_AGE = 3600  # 1 hour
SESSION_EXPIRE_AT_BROWSER_CLOSE = True
```

## Django Security Checklist

- [ ] SECRET_KEY from environment
- [ ] DEBUG = False in production
- [ ] ALLOWED_HOSTS configured
- [ ] HTTPS enforced (SECURE_SSL_REDIRECT)
- [ ] Secure cookies enabled
- [ ] CSRF protection active
- [ ] Password validators configured
- [ ] User input properly escaped
- [ ] SQL queries parameterized
- [ ] File uploads validated
- [ ] Permissions checked on views
- [ ] Security middleware enabled
