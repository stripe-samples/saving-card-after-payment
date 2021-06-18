# Saving a card after a payment

You can save a card to a customer to reuse for a later payment. Stripe ensures the card is properly authenticated before storing to reduce the risk that the cardholder will have to re-authenticate.

To save a card to a customer you need to use the `setup_future_usage` parameter.

When saving a card you will want to consider how you intend to reuse the card for future payments.

**👩‍💻 On-session reuse -** Charging the card when your customer is in your application or website, e.g:

- An e-commerce store that lets existing customers pay with a saved card.

**🖥️ Off-session reuse -** Charging the card when the user is no longer in your app or on your website, e.g:

- A monthly subscription that charges the card on the first of the month.
- A hotel that charges a deposit before the trip and the full amount after the trip.

Setting `setup_future_usage` to "off_session" will optimize for future off-session payments, while "on_session" will optimize for future on-session usage. If you plan on reusing the card for both on and off-session usage, set `setup_future_usage` to "off_session".

**Demo**

See the sample [live](https://c45nv.sse.codesandbox.io/) or [fork](https://codesandbox.io/s/saving-card-after-a-payment-c45nv) on CodeSandbox.

The demo is running in test mode -- use `4242424242424242` as a test card number with any CVC + future expiration date.

Use the `4000000000003220` test card number to trigger a 3D Secure challenge flow.

Read more about testing on Stripe at https://stripe.com/docs/testing.

<img src="./saving-card-after-payment.gif" alt="A checkout form with a checkbox to let you save a payment method" align="center">

There are two implementations depending on whether you want to use webhooks for any post-payment process:

- **[/using-webhooks](/using-webhooks)** Confirms the payment on the client and requires using webhooks or other async event handlers for any post-payment logic (e.g. sending email receipts, fulfilling orders).
- **[/without-webhooks](/without-webhooks)** Confirms the payment on the server and allows you to run any post-payment logic right after.

This sample shows:

<!-- prettier-ignore -->
|     | Using webhooks | Without webhooks
:--- | :---: | :---:
💳 **Collecting card and cardholder details.** Both integrations use [Stripe Elements](https://stripe.com/docs/stripe-js) to build a custom checkout form. | ✅  | ✅ |
🙅 **Handling card authentication requests and declines.** Attempts to charge a card can fail if the bank declines the purchase or requests additional authentication.  | ✅  | ✅ |
💁 **Saving cards to reuse later.** Both integrations show how to save a card to a Customer for later use. | ✅ | ✅ |
🏦 **Easily scalable to other payment methods.** Webhooks enable easy adoption of other asynchroneous payment methods like direct debits and push-based payment flows. | ✅ | ❌ |

## How to run locally

This sample includes 6 server implementations in Go, Node, Ruby, Python, Java, and PHP for the two integration types: [using-webhooks](/using-webhooks) and [without-webhooks](/without-webhooks).

Follow the steps below to run locally.

**1. Clone and configure the sample**

The Stripe CLI is the fastest way to clone and configure a sample to run locally.

**Using the Stripe CLI**

If you haven't already installed the CLI, follow the [installation steps](https://github.com/stripe/stripe-cli#installation) in the project README. The CLI is useful for cloning samples and locally testing webhooks and Stripe integrations.

In your terminal shell, run the Stripe CLI command to clone the sample:

```
stripe samples create saving-card-after-payment
```

The CLI will walk you through picking your integration type, server and client languages, and configuring your .env config file with your Stripe API keys.

**Installing and cloning manually**

If you do not want to use the Stripe CLI, you can manually clone and configure the sample yourself:

```
git clone https://github.com/stripe-samples/saving-card-after-payment
```

Copy the .env.example file into a file named .env in the folder of the server you want to use. For example:

```
cp .env.example using-webhooks/server/node/.env
```

You will need a Stripe account in order to run the demo. Once you set up your account, go to the Stripe [developer dashboard](https://stripe.com/docs/development/quickstart#api-keys) to find your API keys.

```
STRIPE_PUBLISHABLE_KEY=<replace-with-your-publishable-key>
STRIPE_SECRET_KEY=<replace-with-your-secret-key>
```

`STATIC_DIR` tells the server where to the client files are located and does not need to be modified unless you move the server files.

**2. Follow the server instructions on how to run:**

Pick the server language you want and follow the instructions in the server folder README on how to run.

For example, if you want to run the Node server in `using-webhooks`:

```
cd using-webhooks/server/node # there's a README in this folder with instructions
npm install
npm start
```

**4. [Optional] Run a webhook locally:**

If you want to test the `using-webhooks` integration with a local webhook on your machine, you can use the Stripe CLI to easily spin one up.

First [install the CLI](https://stripe.com/docs/stripe-cli) and [link your Stripe account](https://stripe.com/docs/stripe-cli#link-account).

```
stripe listen --forward-to localhost:4242/webhook
```

The CLI will print a webhook secret key to the console. Set `STRIPE_WEBHOOK_SECRET` to this value in your .env file.

You should see events logged in the console where the CLI is running.

When you are ready to create a live webhook endpoint, follow our guide in the docs on [configuring a webhook endpoint in the dashboard](https://stripe.com/docs/webhooks/setup#configure-webhook-settings).

## FAQ

Q: Why did you pick these frameworks?

A: We chose the most minimal framework to convey the key Stripe calls and concepts you need to understand. These demos are meant as an educational tool that helps you roadmap how to integrate Stripe within your own system independent of the framework.

## Get support
If you found a bug or want to suggest a new [feature/use case/sample], please [file an issue](../../issues).

If you have questions, comments, or need help with code, we're here to help:
- on [Discord](https://stripe.com/go/developer-chat)
- on Twitter at [@StripeDev](https://twitter.com/StripeDev)
- on Stack Overflow at the [stripe-payments](https://stackoverflow.com/tags/stripe-payments/info) tag
- by [email](mailto:support+github@stripe.com)

Sign up to [stay updated with developer news](https://go.stripe.global/dev-digest).

## Author(s)

[@adreyfus-stripe](https://twitter.com/adrind)
