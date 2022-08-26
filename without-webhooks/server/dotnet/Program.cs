using System.Text.Json;
using Microsoft.Extensions.Options;
using Stripe;

DotNetEnv.Env.Load();
StripeConfiguration.ApiKey = Environment.GetEnvironmentVariable("STRIPE_SECRET_KEY");

StripeConfiguration.AppInfo = new AppInfo
{
    Name = "https://github.com/stripe-samples/saving-card-after-payment",
    Url = "https://github.com/stripe-samples",
    Version = "0.1.0",
};

StripeConfiguration.ApiKey = Environment.GetEnvironmentVariable("STRIPE_SECRET_KEY");

var builder = WebApplication.CreateBuilder(new WebApplicationOptions
{
    Args = args,
    WebRootPath = Environment.GetEnvironmentVariable("STATIC_DIR")
});

builder.Services.Configure<StripeOptions>(options =>
{
    options.PublishableKey = Environment.GetEnvironmentVariable("STRIPE_PUBLISHABLE_KEY");
    options.SecretKey = Environment.GetEnvironmentVariable("STRIPE_SECRET_KEY");
    options.WebhookSecret = Environment.GetEnvironmentVariable("STRIPE_WEBHOOK_SECRET");
});

var app = builder.Build();

if (app.Environment.IsDevelopment())
{
    app.UseDeveloperExceptionPage();
}

app.UseDefaultFiles();
app.UseStaticFiles();

app.MapGet("/stripe-key", (IOptions<StripeOptions> stripeOptions) =>
     Results.Ok(new { publicKey = stripeOptions.Value.PublishableKey })
);


app.MapPost("/pay", async (HttpRequest request) =>
{
    var json = await new StreamReader(request.Body).ReadToEndAsync();
    PaymentIntent intent = null;
    using var jdocument = JsonDocument.Parse(json);

    var paymentIntentSvc = new PaymentIntentService();
    if (jdocument.RootElement.TryGetProperty("paymentIntentId", out JsonElement paymentIntentEle))
    {
        intent = await paymentIntentSvc.ConfirmAsync(paymentIntentEle.GetString());
    }
    else
    {
        var piOptions = new PaymentIntentCreateOptions
        {
            Amount = 1400,
            Currency = jdocument.RootElement.GetProperty("currency").GetString(),
            PaymentMethod = jdocument.RootElement.GetProperty("paymentMethodId").GetString(),
            ConfirmationMethod = "manual",
            Confirm = true
        };
        
        if (jdocument.RootElement.TryGetProperty("isSavingCard", out JsonElement isSavingCardEle))
        {
            // Create a Customer to store the PaymentMethod for reuse
            var customerSvc = new CustomerService();
            var customer = await customerSvc.CreateAsync(new());

            piOptions.Customer = customer.Id;

            // setup_future_usage saves the card and tells Stripe how you plan to use it later
            // set to 'off_session' if you plan on charging the saved card when the customer is not present
            piOptions.SetupFutureUsage = "off_session";
        }

        intent = await paymentIntentSvc.CreateAsync(piOptions);
    }

    return GenerateResponse(intent);
});

IResult GenerateResponse(PaymentIntent intent)
{
    var status = intent.Status;
    if (status == "requires_action" || status == "requires_source_action")
    {
        // Card requires authentication
        return Results.Json(new
        {
            requiresAction = true,
            paymentIntentId = intent.Id,
            clientSecret = intent.ClientSecret
        });
    }
    else if (status == "requires_payment_method" || status == "requires_source")
    {
        // Card was not properly authenticated, suggest a new payment method
        return Results.Json(new { error = "Your card was denied, please provide a new payment method" });
    }
    else if (status == "succeeded")
    {
        // Payment is complete, authentication not required
        // To cancel the payment after capture you will need to issue a Refund (https://stripe.com/docs/api/refunds)
        return Results.Json(new { clientSecret = intent.ClientSecret });
    }
    return Results.Ok();
}

app.Run();