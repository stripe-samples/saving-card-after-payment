package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"fmt"

	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/customer"
	"github.com/stripe/stripe-go/v71/paymentintent"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("godotenv.Load: %v", err)
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	http.Handle("/", http.FileServer(http.Dir(os.Getenv("STATIC_DIR"))))
	http.HandleFunc("/stripe-key", handleStripeKey)
	http.HandleFunc("/pay", handlePay)

	addr := "localhost:4242"
	log.Printf("Listening on %s ...", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleStripeKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, struct {
		PublishableKey string `json:"publicKey"`
	}{
		PublishableKey: os.Getenv("STRIPE_PUBLISHABLE_KEY"),
	})
}

// PayItemParams represents a single item passed from the client.
// In practice, the ID of the PayItemParams object would be some
// ID or reference to an internal product that you can use to
// determine the price. You need to implement calculateOrderAmount
// or a similar function to actually calculate the amount here
// on the server. That way, the user cannot modify the amount that
// is charged by changing the client.
type PayItemParams struct {
	ID string `json:"id"`
}

// PayRequestParams represents the structure of the request from
// the client.
type PayRequestParams struct {
	Currency        string          `json:"currency"`
	SaveCard        bool            `json:"isSavingCard"`
	Items           []PayItemParams `json:"items"`
	PaymentMethodID string          `json:"paymentMethodId"`

	// paymentIntentId will only be included on subsequent calls, not the initial call.
	PaymentIntentID string `json:"paymentIntentId"`
}

// PayResponse represents the response object generated for handlePay
type PayResponse struct {
	RequiresAction  bool   `json:"requiresAction"`
	PaymentIntentID string `json:"paymentIntentId"`
	ClientSecret    string `json:"clientSecret"`
	Error           string `json:"error"`
}

func calculateOrderAmount(items []PayItemParams) (amount int64) {
	amount = 1400
	return
}

func generateResponse(intent *stripe.PaymentIntent) (resp PayResponse) {
	switch status := intent.Status; status {
	case "requires_action", "requires_source_action":
		resp.RequiresAction = true
		resp.PaymentIntentID = intent.ID
		resp.ClientSecret = intent.ClientSecret
	case "requires_payment_method", "requires_source":
		resp.Error = "Your card was denied, please provide a new payment method"
	case "succeeded":
		fmt.Println("💰 Payment received!")
		resp.ClientSecret = intent.ClientSecret
	}
	return
}

func handlePay(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	// Decode the incoming request
	req := PayRequestParams{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewDecoder.Decode: %v", err)
		return
	}

	// calculate the order amount, you'll need to modify `calculateOrderAmount`
	// to match your business logic.
	orderAmount := calculateOrderAmount(req.Items)

	// If this is a second call later to confirm (not the initial call to create
	// the payment intent), then we want to confirm the payment intent and return.
	if req.PaymentIntentID != "" {
		intent, err := paymentintent.Confirm(req.PaymentIntentID, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("paymentintent.Confirm: %v", err)
			return
		}
		resp := generateResponse(intent)
		writeJSON(w, resp)
		return
	}

	paymentIntentParams := &stripe.PaymentIntentParams{
		Amount:             stripe.Int64(orderAmount),
		Currency:           stripe.String(req.Currency),
		PaymentMethod:      stripe.String(req.PaymentMethodID),
		ConfirmationMethod: stripe.String(string(stripe.PaymentIntentConfirmationMethodManual)),
		Confirm:            stripe.Bool(true),
	}

	if req.SaveCard {
		// Create a Customer to store the PaymentMethod
		customerParams := &stripe.CustomerParams{}
		c, err := customer.New(customerParams)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("customer.New: %v", err)
			return
		}
		paymentIntentParams.Customer = stripe.String(c.ID)

		// SetupFutureUsage saves the card and tells Stripe how you plan to use it later
		// set to PaymentIntentSetupFutureUsageOffSession if you plan on charging
		// the saved card when the customer is not present
		paymentIntentParams.SetupFutureUsage = stripe.String(string(stripe.PaymentIntentSetupFutureUsageOffSession))
	}

	pi, err := paymentintent.New(paymentIntentParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("paymentintent.New: %v", err)
		return
	}

	resp := generateResponse(pi)
	writeJSON(w, resp)
	return
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewEncoder.Encode: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := io.Copy(w, &buf); err != nil {
		log.Printf("io.Copy: %v", err)
		return
	}
}
