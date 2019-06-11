package com.stripe.sample;

import java.nio.file.Paths;

import static spark.Spark.get;
import static spark.Spark.post;
import static spark.Spark.staticFiles;
import static spark.Spark.port;

import com.google.gson.Gson;
import com.google.gson.annotations.SerializedName;

import com.stripe.Stripe;
import com.stripe.model.Customer;
import com.stripe.model.PaymentIntent;
import com.stripe.param.CustomerCreateParams;
import com.stripe.param.PaymentIntentCreateParams;

import io.github.cdimascio.dotenv.Dotenv;

public class Server {
    private static Gson gson = new Gson();

    static class StripeKeyResponse {
        private String publicKey;

        public StripeKeyResponse(String publicKey) {
            this.publicKey = publicKey;
        }
    }

    static class PayRequestBody {
        @SerializedName("items")
        Object[] items;
        @SerializedName("paymentIntentId")
        String paymentIntentId;
        @SerializedName("paymentMethodId")
        String paymentMethodId;
        @SerializedName("currency")
        String currency;
        @SerializedName("isSavingCard")
        Boolean isSavingCard;

        public Object[] getItems() {
            return items;
        }

        public String getPaymentIntentId() {
            return paymentIntentId;
        }

        public String getPaymentMethodId() {
            return paymentMethodId;
        }

        public String getCurrency() {
            return currency;
        }

        public Boolean getIsSavingCard() {
            return isSavingCard;
        }
    }

    static class PayResponseBody {
        private String clientSecret;
        private String paymentIntentId;
        private Boolean requiresAction;
        private String error;

        public PayResponseBody() {

        }

        public void setClientSecret(String clientSecret) {
            this.clientSecret = clientSecret;
        }

        public void setPaymentIntentId(String paymentIntentId) {
            this.paymentIntentId = paymentIntentId;
        }

        public void setRequiresAction(Boolean requiresAction) {
            this.requiresAction = requiresAction;
        }

        public void setError(String error) {
            this.error = error;
        }

    }

    static int calculateOrderAmount(Object[] items) {
        // Replace this constant with a calculation of the order's amount
        // Calculate the order total on the server to prevent
        // users from directly manipulating the amount on the client
        return 1400;
    }

    static PayResponseBody generateResponse(PaymentIntent intent, PayResponseBody response) {
        switch (intent.getStatus()) {
        case "requires_action":
        case "requires_source_action":
            // Card requires authentication
            response.setClientSecret(intent.getClientSecret());
            response.setPaymentIntentId(intent.getId());
            response.setRequiresAction(true);
            break;
        case "requires_payment_method":
        case "requires_source":
            // Card was not properly authenticated, suggest a new payment method
            response.setError("Your card was denied, please provide a new payment method");
            break;
        case "succeeded":
            System.out.println("💰 Payment received!");
            // Payment is complete, authentication not required
            // To cancel the payment after capture you will need to issue a Refund
            // (https://stripe.com/docs/api/refunds)
            response.setClientSecret(intent.getClientSecret());
            break;
        default:
            response.setError("Unrecognized status");
        }
        return response;
    }

    public static void main(String[] args) {
        port(4242);
        String ENV_PATH = "../../../";
        Dotenv dotenv = Dotenv.configure().directory(ENV_PATH).load();

        Stripe.apiKey = dotenv.get("STRIPE_SECRET_KEY");

        staticFiles.externalLocation(
                Paths.get(Paths.get("").toAbsolutePath().toString(), dotenv.get("STATIC_DIR")).normalize().toString());

        get("/stripe-key", (request, response) -> {
            response.type("application/json");
            // Send public key to client
            return gson.toJson(new StripeKeyResponse(dotenv.get("STRIPE_PUBLIC_KEY")));
        });

        post("/pay", (request, response) -> {
            PayRequestBody postBody = gson.fromJson(request.body(), PayRequestBody.class);

            PaymentIntent intent;
            PayResponseBody responseBody = new PayResponseBody();
            try {
                if (postBody.getPaymentIntentId() == null) {
                    int orderAmount = calculateOrderAmount(postBody.getItems());
                    // Create new PaymentIntent for the order
                    PaymentIntentCreateParams.Builder createParams = new PaymentIntentCreateParams.Builder()
                            .setCurrency(postBody.getCurrency()).setAmount(new Long(orderAmount))
                            .setPaymentMethod(postBody.getPaymentMethodId())
                            .setConfirmationMethod(PaymentIntentCreateParams.ConfirmationMethod.MANUAL)
                            .setConfirm(true);
                    
                    if(postBody.getIsSavingCard()) {
                        // Create a Customer to store the PaymentMethod
                        Customer customer = Customer.create(new CustomerCreateParams.Builder().build());
                        createParams.setCustomer(customer.getId());
                        
                        // Set save_payment_method to true to attach the PaymentMethod to the Customer
                        createParams.setSavePaymentMethod(true);

                        // setup_future_usage tells Stripe how you plan on using the saved card
                        // set to "off_session" if you plan on charging the saved card when your customer is not present
                        createParams.setSetupFutureUsage(PaymentIntentCreateParams.SetupFutureUsage.OFF_SESSION);
                    }

                    // Create a PaymentIntent with the order amount and currency
                    intent = PaymentIntent.create(createParams.build());
                } else {
                    // Confirm the PaymentIntent to collect the money
                    intent = PaymentIntent.retrieve(postBody.getPaymentIntentId());
                    intent = intent.confirm();
                }

                responseBody = generateResponse(intent, responseBody);
            } catch (Exception e) {
                // Handle "hard declines" e.g. insufficient funds, expired card, etc
                // See https://stripe.com/docs/declines/codes for more
                responseBody.setError(e.getMessage());
            }

            return gson.toJson(responseBody);
        });
    }
}