package paymongo

import (
	"context"
	"gobrewflow/internal/config"
	"gobrewflow/internal/services/payments"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func TestAdapter_CreateCheckout_Integration(t *testing.T) {
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

	retryCfg := config.RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   500 * time.Millisecond,
		MaxDelay:    2 * time.Second,
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	// 4. Create the PayMongo client and adapter.
	client := NewClient(baseURL, secret, nil, retryCfg, logger)
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

func TestClient_DoCreateCheckout_RetriesThenSucceeds(t *testing.T) {
	var attempts int
	var idempotencyKeys []string

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++

			idempotencyKeys = append(
				idempotencyKeys,
				r.Header.Get("Idempotency-Key"),
			)

			if attempts == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"data":{"id":"cs_test"}}`))
		}),
	)
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := NewClient(
		server.URL,
		"test-secret",
		nil,
		config.RetryConfig{
			MaxAttempts: 3,
			BaseDelay:   time.Millisecond,
			MaxDelay:    2 * time.Millisecond,
		},
		logger,
	)

	key := "test-idempotency-key"

	body, err := client.DoCreateCheckout(
		context.Background(),
		[]byte(`{}`),
		key,
	)
	if err != nil {
		t.Fatal(err)
	}

	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}

	if string(body) != `{"data":{"id":"cs_test"}}` {
		t.Fatalf("unexpected response: %s", body)
	}

	if len(idempotencyKeys) != 2 {
		t.Fatalf("expected 2 idempotency keys, got %d", len(idempotencyKeys))
	}

	if idempotencyKeys[0] != key {
		t.Fatalf("expected first idempotency key %q, got %q", key, idempotencyKeys[0])
	}

	if idempotencyKeys[1] != key {
		t.Fatalf("expected second idempotency key %q, got %q", key, idempotencyKeys[1])
	}
}

func TestClient_DoCreateCheckout_DoesNotRetry400(t *testing.T) {
	var attempts int

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			w.WriteHeader(http.StatusBadRequest)
		}),
	)
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := NewClient(
		server.URL,
		"test-secret",
		nil,
		config.RetryConfig{
			MaxAttempts: 3,
			BaseDelay:   time.Millisecond,
			MaxDelay:    time.Millisecond,
		},
		logger,
	)

	_, err := client.DoCreateCheckout(
		context.Background(),
		[]byte(`{}`),
		"test-key",
	)

	if err == nil {
		t.Fatal("expected error")
	}

	if attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}
}
