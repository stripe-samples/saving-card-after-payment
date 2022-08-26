# Saving a card after a payment

An [.NET](https://dotnet.microsoft.com/download/dotnet) implementation.

## Requirements

* [.NET 6](https://get.dot.net/) or above
* [Configured .env file](../../../README.md)

## How to run

1. Confirm `.env` configuration

Ensure the API keys are configured in `.env` in this directory. It should include the following keys:

```yaml
# Stripe API keys - see https://stripe.com/docs/development/quickstart#api-keys
STRIPE_PUBLISHABLE_KEY=pk_test...
STRIPE_SECRET_KEY=sk_test...

# Required to verify signatures in the webhook handler.
# See README on how to use the Stripe CLI to test webhooks
STRIPE_WEBHOOK_SECRET=whsec_...

# Path to front-end implementation. Note: PHP has it's own front end implementation.
STATIC_DIR=../../client
```

2. Run the application

```
dotnet run 
```

3. Go to `localhost:4242` to see the demo.