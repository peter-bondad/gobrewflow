package paymongo

import (
	"context"
	"gobrewflow/internal/services/payments"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestAdapter_CreateCheckout(t *testing.T) {
	if err := godotenv.Load("../../../../.env"); err != nil {
		t.Fatal(err)
	}

	baseURL := os.Getenv("PAYMONGO_BASE_URL")
	secret := os.Getenv("PAYMONGO_TEST_SECRET_KEY")

	if baseURL == "" {
		t.Fatal("PAYMONGO_BASE_URL is required")
	}

	if secret == "" {
		t.Fatal("PAYMONGO_TEST_SECRET_KEY is required")
	}

	client := NewClient(baseURL, secret, nil)
	adapter := NewAdapter(client)

	output, err := adapter.CreateCheckout(
		context.Background(),
		payments.CreateCheckoutInput{
			ReferenceID: "TEST-ORDER-001",
			Amount:      10000,
			Currency:    "PHP",
			Description: "Test Order",
			SuccessURL:  "http://localhost:3000/payment/success",
			CancelURL:   "http://localhost:3000/payment/cancel",
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Checkout ID: %s", output.CheckoutID)
	t.Logf("Checkout URL: %s", output.CheckoutURL)
}
