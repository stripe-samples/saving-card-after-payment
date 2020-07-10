package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/customer"
	"github.com/stripe/stripe-go/v71/paymentintent"
	"github.com/stripe/stripe-go/v71/webhook"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("godotenv.Load: %v", err)
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	http.Handle("/", http.FileServer(http.Dir(os.Getenv("STATIC_DIR"))))
	http.HandleFunc("/create-payment-intent", handleCreatePaymentIntent)
	http.HandleFunc("/webhook", handleWebhook)

	addr := "localhost:4242"
	log.Printf("Listening on %s ...", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
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
	Currency string          `json:"currency"`
	Items    []PayItemParams `json:"items"`
}

func calculateOrderAmount(items []PayItemParams) (amount int64) {
	amount = 1400
	return
}

func handleCreatePaymentIntent(w http.ResponseWriter, r *http.Request) {
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

	customerParams := &stripe.CustomerParams{}
	c, err := customer.New(customerParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("customer.New: %v", err)
		return
	}

	paymentIntentParams := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(orderAmount),
		Currency: stripe.String(req.Currency),
		Customer: stripe.String(c.ID),
	}

	pi, err := paymentintent.New(paymentIntentParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("paymentintent.New: %v", err)
		return
	}

	writeJSON(w, struct {
		PublicKey    string `json:"publicKey"`
		ClientSecret string `json:"clientSecret"`
		ID           string `json:"id"`
	}{
		PublicKey:    os.Getenv("STRIPE_PUBLISHABLE_KEY"),
		ClientSecret: pi.ClientSecret,
		ID:           pi.ID,
	})
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("ioutil.ReadAll: %v", err)
		return
	}

	event, err := webhook.ConstructEvent(b, r.Header.Get("Stripe-Signature"), os.Getenv("STRIPE_WEBHOOK_SECRET"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("webhook.ConstructEvent: %v", err)
		return
	}

	if event.Type == "payment_method.attached" {
		log.Printf("❗ PaymentMethod successfully attached to Customer")
		return
	}

	if event.Type == "payment_intent.succeeded" {
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if string(paymentIntent.SetupFutureUsage) == "" {
			log.Printf("❗ Customer did not want to save the card.")
		}

		log.Printf("💰 Payment received!")
		return
	}

	if event.Type == "payment_intent.payment_failed" {
		log.Printf("❌ Payment failed.")
		return
	}

	writeJSON(w, struct {
		Status string `json:"status"`
	}{
		Status: "success",
	})
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
