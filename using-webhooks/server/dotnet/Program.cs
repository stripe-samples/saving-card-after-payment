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

app.MapPost("/create-payment-intent", async (HttpRequest request, IOptions<StripeOptions> stripeOptions) =>
{
    var json = await new StreamReader(request.Body).ReadToEndAsync();
    using var jdocument = JsonDocument.Parse(json);

    var customerSvc = new CustomerService();
    var customer = await customerSvc.CreateAsync(new());

    var paymentIntentSvc = new PaymentIntentService();
    var piOptions = new PaymentIntentCreateOptions
    {
        Amount = 1400,
        Currency = jdocument.RootElement.GetProperty("currency").GetString(),
        Customer = customer.Id
    };

    var paymentIntent = await paymentIntentSvc.CreateAsync(piOptions);
    return Results.Ok(new
    {
        publicKey = stripeOptions.Value.PublishableKey,
        clientSecret = paymentIntent.ClientSecret,
        id = paymentIntent.Id
    });
});

app.MapPost("/webhook", async (HttpRequest request, IOptions<StripeOptions> options) =>
{
    var json = await new StreamReader(request.Body).ReadToEndAsync();
    Event stripeEvent;
    try
    {
        stripeEvent = EventUtility.ConstructEvent(
            json,
            request.Headers["Stripe-Signature"],
             options.Value.WebhookSecret, throwOnApiVersionMismatch: false
        );

        if (stripeEvent.Type == "payment_method.attached")
        {
            app.Logger.LogInformation("❗ PaymentMethod successfully attached to Customer");
        }
        else if (stripeEvent.Type == "payment_intent.succeeded")
        {
            var intentData = stripeEvent.Data.Object as PaymentIntent;
            if (string.IsNullOrEmpty(intentData.SetupFutureUsage))
                app.Logger.LogInformation("❗ Customer did not want to save the card.");

            app.Logger.LogInformation("💰 Payment received!");
            // Fulfill any orders, e-mail receipts, etc            
        }
        else if (stripeEvent.Type == "payment_intent.payment_failed")
        {
            app.Logger.LogError("❌ Payment failed.");
        }
    }
    catch (StripeException ex)
    {
        app.Logger.LogError(ex, ex.Message);
        return Results.BadRequest();
    }

    return Results.Ok(new { status = "success" });
});

app.Run();
