package paymongo

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestAdapter_VerifyWebhookSignature(t *testing.T) {
	secret := "whsec_test_secret"
	payload := []byte(`{"event_type":"payment.paid"}`)
	timestamp := "1750000000"

	// PayMongo signs: timestamp + "." + raw body.
	signaturePayload := timestamp + "." + string(payload)

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signaturePayload))

	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	adapter := &Adapter{
		webhookSecret: secret,
	}

	t.Run("valid signature", func(t *testing.T) {
		signature := "t=" + timestamp + ",te=" + expectedSignature

		err := adapter.VerifyWebhookSignature(payload, signature)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("invalid signature", func(t *testing.T) {
		signature := "t=" + timestamp + ",te=invalid_signature"

		err := adapter.VerifyWebhookSignature(payload, signature)

		if !errors.Is(err, ErrInvalidWebhookSignature) {
			t.Fatalf(
				"expected ErrInvalidWebhookSignature, got %v",
				err,
			)
		}
	})

	t.Run("missing timestamp", func(t *testing.T) {
		signature := "te=" + expectedSignature

		err := adapter.VerifyWebhookSignature(payload, signature)

		if !errors.Is(err, ErrInvalidWebhookSignature) {
			t.Fatalf(
				"expected ErrInvalidWebhookSignature, got %v",
				err,
			)
		}
	})

	t.Run("missing test signature", func(t *testing.T) {
		signature := "t=" + timestamp

		err := adapter.VerifyWebhookSignature(payload, signature)

		if !errors.Is(err, ErrInvalidWebhookSignature) {
			t.Fatalf(
				"expected ErrInvalidWebhookSignature, got %v",
				err,
			)
		}
	})

	t.Run("modified payload", func(t *testing.T) {
		signature := "t=" + timestamp + ",te=" + expectedSignature

		modifiedPayload := []byte(`{"event_type":"payment.failed"}`)

		err := adapter.VerifyWebhookSignature(
			modifiedPayload,
			signature,
		)

		if !errors.Is(err, ErrInvalidWebhookSignature) {
			t.Fatalf(
				"expected ErrInvalidWebhookSignature, got %v",
				err,
			)
		}
	})
}

func TestAdapter_ParseWebhook(t *testing.T) {
	paymentID := uuid.NewString()

	payload := []byte(`{
		"data": {
			"type": "checkout_session.payment.paid",
			"data": {
				"id": "cs_test_123",
				"type": "checkout_session",
				"attributes": {
					"reference_number": "ORDER-001",
					"payments": [
						{
							"id": "` + paymentID + `",
							"attributes": {
								"amount": 10000,
								"currency": "PHP",
								"status": "paid",
								"source": {
									"type": "gcash"
								}
							}
						}
					]
				}
			}
		}
	}`)

	adapter := &Adapter{}

	result, err := adapter.ParseWebhook(payload)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.CheckoutID != "cs_test_123" {
		t.Fatalf(
			"expected checkout ID cs_test_123, got %s",
			result.CheckoutID,
		)
	}

	if result.PaymentID == nil {
		t.Fatal("expected payment ID")
	}

	if *result.PaymentID != paymentID {
		t.Fatalf(
			"expected payment ID %s, got %s",
			paymentID,
			*result.PaymentID,
		)
	}

	if result.ReferenceNumber != "ORDER-001" {
		t.Fatalf(
			"expected reference number ORDER-001, got %s",
			result.ReferenceNumber,
		)
	}

	if result.Amount != 10000 {
		t.Fatalf(
			"expected amount 10000, got %d",
			result.Amount,
		)
	}

	if result.Currency != "PHP" {
		t.Fatalf(
			"expected currency PHP, got %s",
			result.Currency,
		)
	}

	if result.Status != "PAID" {
		t.Fatalf(
			"expected status PAID, got %s",
			result.Status,
		)
	}

	if result.PaymentMethod == nil {
		t.Fatal("expected payment method")
	}

	if *result.PaymentMethod != "GCASH" {
		t.Fatalf(
			"expected payment method GCASH, got %s",
			*result.PaymentMethod,
		)
	}
}

func TestAdapter_ParseWebhook_Errors(t *testing.T) {
	t.Run("invalid JSON", func(t *testing.T) {
		adapter := &Adapter{}

		_, err := adapter.ParseWebhook(
			[]byte(`{"invalid"`),
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("unsupported event", func(t *testing.T) {
		adapter := &Adapter{}

		payload := []byte(`{
			"data": {
				"type": "checkout_session.payment.failed"
			}
		}`)

		_, err := adapter.ParseWebhook(payload)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("missing checkout session ID", func(t *testing.T) {
		adapter := &Adapter{}

		payload := []byte(`{
			"data": {
				"type": "checkout_session.payment.paid",
				"data": {
					"id": "",
					"attributes": {
						"payments": [
							{
								"id": "pay_test_123",
								"attributes": {
									"amount": 10000,
									"currency": "PHP",
									"status": "paid"
								}
							}
						]
					}
				}
			}
		}`)

		_, err := adapter.ParseWebhook(payload)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("missing payment", func(t *testing.T) {
		adapter := &Adapter{}

		payload := []byte(`{
			"data": {
				"type": "checkout_session.payment.paid",
				"data": {
					"id": "cs_test_123",
					"attributes": {
						"payments": []
					}
				}
			}
		}`)

		_, err := adapter.ParseWebhook(payload)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
