# Stripe Import Patterns

## Package Names

### NPM (JavaScript/Node.js)
- `stripe` (official SDK)
- `@stripe/stripe-js` (Stripe.js for browsers)
- `@stripe/react-stripe-js` (React components)
- `stripe-node` (legacy)

### Python (PyPI)
- `stripe` (official SDK)
- `django-stripe` (Django integration)
- `dj-stripe` (Django Stripe integration)
- `flask-stripe` (Flask integration)

### Ruby (RubyGems)
- `stripe` (official SDK)
- `stripe-ruby-mock` (testing)

### PHP (Composer)
- `stripe/stripe-php` (official SDK)
- `cartalyst/stripe-laravel` (Laravel integration)
- `omnipay/stripe` (Omnipay driver)

### Go (Go modules)
- `github.com/stripe/stripe-go/v*`
- `github.com/stripe/stripe-go`

### Java (Maven)
- `com.stripe:stripe-java`

### .NET (NuGet)
- `Stripe.net`

### Rust (Cargo)
- `async-stripe`
- `stripe-rust`

## Import Statements

### JavaScript/Node.js
```javascript
import Stripe from 'stripe';
import { loadStripe } from '@stripe/stripe-js';
import { Elements, CardElement } from '@stripe/react-stripe-js';
const stripe = require('stripe');
const Stripe = require('stripe');
```

### TypeScript
```typescript
import Stripe from 'stripe';
import type { Stripe as StripeType } from 'stripe';
import { loadStripe, StripeElementsOptions } from '@stripe/stripe-js';
```

### Python
```python
import stripe
from stripe import Customer, Charge, PaymentIntent
from stripe.error import StripeError
import stripe.webhook
```

### Ruby
```ruby
require 'stripe'
require 'stripe/webhook'
```

### PHP
```php
require_once('vendor/autoload.php');
use Stripe\Stripe;
use Stripe\Customer;
use Stripe\Charge;
use Stripe\PaymentIntent;
use Stripe\Webhook;
```

### Go
```go
import (
    "github.com/stripe/stripe-go/v78"
    "github.com/stripe/stripe-go/v78/charge"
    "github.com/stripe/stripe-go/v78/customer"
    "github.com/stripe/stripe-go/v78/paymentintent"
    "github.com/stripe/stripe-go/v78/webhook"
)
```

### Java
```java
import com.stripe.Stripe;
import com.stripe.model.Charge;
import com.stripe.model.Customer;
import com.stripe.model.PaymentIntent;
import com.stripe.net.Webhook;
```

### C#/.NET
```csharp
using Stripe;
using Stripe.Checkout;
```

### Rust
```rust
use async_stripe::*;
use stripe::*;
```

## Package Manager Files

### package.json (Node.js)
```json
"dependencies": {
  "stripe": "^*",
  "@stripe/stripe-js": "^*",
  "@stripe/react-stripe-js": "^*"
}
```

### requirements.txt (Python)
```
stripe==*
stripe>=*
```

### Pipfile (Python)
```
[packages]
stripe = "*"
```

### Gemfile (Ruby)
```ruby
gem 'stripe'
```

### composer.json (PHP)
```json
"require": {
  "stripe/stripe-php": "^*"
}
```

### go.mod (Go)
```
require github.com/stripe/stripe-go/v78 v*
```

### pom.xml (Java Maven)
```xml
<dependency>
    <groupId>com.stripe</groupId>
    <artifactId>stripe-java</artifactId>
    <version>*</version>
</dependency>
```

### Cargo.toml (Rust)
```toml
[dependencies]
async-stripe = "*"
```

### *.csproj (.NET)
```xml
<PackageReference Include="Stripe.net" Version="*" />
```

## HTML/Browser Script Tags

```html
<script src="https://js.stripe.com/v3/"></script>
<script src="https://checkout.stripe.com/checkout.js"></script>
```

## Detection Confidence

- **HIGH**: Package dependency in package manager files
- **HIGH**: Import statements in source code
- **MEDIUM**: Script tags in HTML
- **LOW**: Generic payment-related imports without Stripe specifics
