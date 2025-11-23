# Stripe API Patterns

## API Endpoint Patterns

### REST API Endpoints
- `https://api.stripe.com/v1/*`
- `https://api.stripe.com/v2/*`
- `https://uploads.stripe.com/*`
- `https://files.stripe.com/*`
- `https://connect.stripe.com/*`

### Webhook Endpoints
- Pattern: `/stripe/webhook`
- Pattern: `/webhooks/stripe`
- Pattern: `/api/stripe/webhook`
- Pattern: `/stripe-webhook`

## API Method Signatures

### JavaScript/Node.js
```javascript
stripe.charges.create()
stripe.customers.create()
stripe.paymentIntents.create()
stripe.subscriptions.create()
stripe.invoices.create()
stripe.refunds.create()
stripe.tokens.create()
stripe.sources.create()
stripe.paymentMethods.attach()
stripe.checkout.sessions.create()
stripe.webhooks.constructEvent()
```

### Python
```python
stripe.Charge.create()
stripe.Customer.create()
stripe.PaymentIntent.create()
stripe.Subscription.create()
stripe.Invoice.create()
stripe.Refund.create()
stripe.Token.create()
stripe.Source.create()
stripe.PaymentMethod.attach()
stripe.checkout.Session.create()
stripe.Webhook.construct_event()
```

### Ruby
```ruby
Stripe::Charge.create()
Stripe::Customer.create()
Stripe::PaymentIntent.create()
Stripe::Subscription.create()
Stripe::Invoice.create()
Stripe::Refund.create()
Stripe::Token.create()
Stripe::Source.create()
Stripe::PaymentMethod.attach()
Stripe::Checkout::Session.create()
Stripe::Webhook.construct_event()
```

### PHP
```php
\Stripe\Charge::create()
\Stripe\Customer::create()
\Stripe\PaymentIntent::create()
\Stripe\Subscription::create()
\Stripe\Invoice::create()
\Stripe\Refund::create()
\Stripe\Token::create()
\Stripe\Source::create()
\Stripe\PaymentMethod::attach()
\Stripe\Checkout\Session::create()
\Stripe\Webhook::constructEvent()
```

### Go
```go
stripe.Charge.New()
stripe.Customer.New()
stripe.PaymentIntent.New()
stripe.Subscription.New()
stripe.Invoice.New()
stripe.Refund.New()
stripe.Token.New()
stripe.Source.New()
stripe.PaymentMethod.Attach()
stripe.CheckoutSession.New()
webhook.ConstructEvent()
```

### Java
```java
Charge.create()
Customer.create()
PaymentIntent.create()
Subscription.create()
Invoice.create()
Refund.create()
Token.create()
Source.create()
PaymentMethod.attach()
Session.create()
Webhook.constructEvent()
```

## API Response Patterns

### Success Response Objects
- `charge.id` with prefix `ch_`
- `customer.id` with prefix `cus_`
- `payment_intent.id` with prefix `pi_`
- `subscription.id` with prefix `sub_`
- `invoice.id` with prefix `in_`
- `refund.id` with prefix `re_`
- `token.id` with prefix `tok_`
- `source.id` with prefix `src_`
- `payment_method.id` with prefix `pm_`
- `session.id` with prefix `cs_`

### Error Response Patterns
```javascript
StripeCardError
StripeInvalidRequestError
StripeAPIError
StripeConnectionError
StripeAuthenticationError
StripePermissionError
StripeRateLimitError
```

## SDK Initialization Patterns

### JavaScript/Node.js
```javascript
const stripe = require('stripe')(process.env.STRIPE_SECRET_KEY);
const stripe = new Stripe(apiKey);
Stripe.setApiKey()
```

### Python
```python
stripe.api_key = os.environ.get('STRIPE_SECRET_KEY')
stripe.api_key = 'sk_'
```

### Ruby
```ruby
Stripe.api_key = ENV['STRIPE_SECRET_KEY']
Stripe.api_key = 'sk_'
```

### PHP
```php
\Stripe\Stripe::setApiKey($stripeSecretKey);
\Stripe\Stripe::setApiKey($_ENV['STRIPE_SECRET_KEY']);
```

### Go
```go
stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
stripe.Key = "sk_"
```

## Webhook Event Types

Common event types to search for:
- `charge.succeeded`
- `charge.failed`
- `customer.created`
- `customer.updated`
- `customer.deleted`
- `payment_intent.succeeded`
- `payment_intent.payment_failed`
- `invoice.payment_succeeded`
- `invoice.payment_failed`
- `subscription.created`
- `subscription.updated`
- `subscription.deleted`
- `checkout.session.completed`
- `payment_method.attached`

## Detection Confidence

- **HIGH**: Direct API calls with stripe domain
- **HIGH**: SDK method calls (stripe.charges.create, etc.)
- **MEDIUM**: Webhook endpoint patterns
- **MEDIUM**: Stripe-specific error handling
- **LOW**: Generic payment processing code without clear Stripe markers
