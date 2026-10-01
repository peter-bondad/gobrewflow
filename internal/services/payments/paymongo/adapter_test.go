package paymongo

import (
	"context"
	"gobrewflow/internal/services/payments"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestAdapter_CreateCheckout(t *testing.T) {
	// 1. Load environment variables from .env.
	if err := godotenv.Load("../../../../.env"); err != nil {
		t.Fatal(err)
	}

	// 2. Get the PayMongo configuration.
	baseURL := os.Getenv("PAYMONGO_BASE_URL")
	secret := os.Getenv("PAYMONGO_TEST_SECRET_KEY")
	webhookSecret := os.Getenv("PAYMONGO_WEBHOOK_SECRET")

	// 3. Make sure the required environment variables exist.
	if baseURL == "" {
		t.Fatal("PAYMONGO_BASE_URL is required")
	}

	if secret == "" {
		t.Fatal("PAYMONGO_TEST_SECRET_KEY is required")
	}

	if webhookSecret == "" {
		t.Fatal("PAYMONGO_WEBHOOK_SECRET is required")
	}

	// 4. Create the PayMongo client and adapter.
	client := NewClient(baseURL, secret, nil)
	adapter := NewAdapter(client, webhookSecret)

	// 5. Create a test checkout session.
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

	// 6. Fail the test if creating the checkout fails.
	if err != nil {
		t.Fatal(err)
	}

	// 7. Print the result so we can test the checkout manually.
	t.Logf("Checkout ID: %s", output.CheckoutID)
	t.Logf("Checkout URL: %s", output.CheckoutURL)
}
