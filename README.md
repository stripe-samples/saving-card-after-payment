# Saving a card after a payment

You can save a card to a customer to reuse for a later payment. Stripe ensures the card is properly authenticated before storing to reduce the risk that the cardholder will have to re-authenticate.

To save a card to a customer you need to use `setup_future_usage` and `save_payment_method` parameters.

When saving a card you will want to consider how you intend to reuse the card for future payments.

**👩‍💻 On-session reuse -** Charging the card when your customer is in your application or website, e.g:

* An e-commerce store that lets existing customers pay with a saved card.

**🖥️ Off-session reuse -** Charging the card when the user is no longer in your app or on your website, e.g: 

* A monthly subscription that charges the card on the first of the month.
* A hotel that charges a deposit before the trip and the full amount after the trip.

Setting `setup_future_usage` to "off_session" will optimize for future off-session payments, while "on_session" will optimize for future on-session usage. If you plan on reusing the card for both on and off-session usage, set `setup_future_usage` to "off_session".

**Demo**

See the sample [live](https://c45nv.sse.codesandbox.io/) or [fork](https://codesandbox.io/s/saving-card-after-a-payment-c45nv) on CodeSandbox.

The demo is running in test mode -- use `4242424242424242` as a test card number with any CVC + future expiration date.

Use the `4000000000003220` test card number to trigger a 3D Secure challenge flow.

Read more about testing on Stripe at https://stripe.com/docs/testing.

<img src="./saving-card-after-payment.gif" alt="A checkout form with a checkbox to let you save a payment method" align="center">

There are two implementations depending on whether you want to use webhooks for any post-payment process: 
* **[/using-webhooks](/using-webhooks)** Confirms the payment on the client and requires using webhooks or other async event handlers for any post-payment logic (e.g. sending email receipts, fulfilling orders). 
* **[/without-webhooks](/without-webhooks)** Confirms the payment on the server and allows you to run any post-payment logic right after.

This sample shows:
<!-- prettier-ignore -->
|     | Using webhooks | Without webhooks
:--- | :---: | :---:
💳 **Collecting card and cardholder details.** Both integrations use [Stripe Elements](https://stripe.com/docs/stripe-js) to build a custom checkout form. | ✅  | ✅ |
🙅 **Handling card authentication requests and declines.** Attempts to charge a card can fail if the bank declines the purchase or requests additional authentication.  | ✅  | ✅ |
💁 **Saving cards to reuse later.** Both integrations show how to save a card to a Customer for later use. | ✅ | ✅ |
🏦 **Easily scalable to other payment methods.** Webhooks enable easy adoption of other asynchroneous payment methods like direct debits and push-based payment flows. | ✅ | ❌ |


## How to run locally
This sample includes 5 server implementations in Node, Ruby, Python, Java, and PHP for the two integration types: [using-webhooks](/using-webhooks) and [without-webhooks](/without-webhooks). 

If you want to run the sample locally copy the .env.example file to your own .env file: 

```
cp .env.example .env
```

Then follow the instructions in the server directory to run.

You will need a Stripe account with its own set of [API keys](https://stripe.com/docs/development#api-keys).


## FAQ
Q: Why did you pick these frameworks?

A: We chose the most minimal framework to convey the key Stripe calls and concepts you need to understand. These demos are meant as an educational tool that helps you roadmap how to integrate Stripe within your own system independent of the framework.

Q: Can you show me how to build X?

A: We are always looking for new sample ideas, please email dev-samples@stripe.com with your suggestion!

## Author(s)
[@adreyfus-stripe](https://twitter.com/adrind)
