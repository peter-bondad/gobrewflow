package paymongo

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gobrewflow/internal/services/payments"
)

var ErrInvalidWebhookSignature = errors.New("invalid webhook signature")

// webhookEvent represents the PayMongo webhook payload we receive.
//
// We only define the fields BrewFlow actually needs instead of modeling
// the entire PayMongo webhook response.
type webhookEvent struct {
	Data struct {
		Attributes struct {
			Type string `json:"type"`

			Data struct {
				ID         string `json:"id"`
				Type       string `json:"type"`
				Attributes struct {
					ReferenceNumber string `json:"reference_number"`

					Payments []struct {
						ID         string `json:"id"`
						Attributes struct {
							Amount   int64  `json:"amount"`
							Currency string `json:"currency"`
							Status   string `json:"status"`

							Source struct {
								Type string `json:"type"`
							} `json:"source"`
						} `json:"attributes"`
					} `json:"payments"`
				} `json:"attributes"`
			} `json:"data"`
		} `json:"attributes"`
	} `json:"data"`
}

// VerifyWebhookSignature verifies that the webhook request was sent
// by PayMongo using the webhook secret configured for BrewFlow.
//
// PayMongo signs the timestamp and raw request body using HMAC-SHA256.
// We calculate the same signature and compare it with PayMongo's signature.
func (a *Adapter) VerifyWebhookSignature(
	payload []byte,
	signature string,
) error {
	// 1. Split the PayMongo signature into its individual parts.
	parts := strings.Split(signature, ",")

	var timestamp string
	var testSignature string

	// 2. Extract the timestamp and test/live signatures.
	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}

		switch key {
		case "t":
			timestamp = value
		case "te":
			testSignature = value
		}
	}

	// 3. A timestamp is required to build the signed payload.
	if timestamp == "" {
		return ErrInvalidWebhookSignature
	}

	// 4. BrewFlow currently uses PayMongo test mode,
	// so we verify against the test-mode signature.
	providedSignature := testSignature

	if providedSignature == "" {
		return ErrInvalidWebhookSignature
	}

	// 5. Build the exact payload PayMongo signed:
	//    timestamp + "." + raw request body.
	signaturePayload := timestamp + "." + string(payload)

	// 6. Generate our own HMAC-SHA256 signature using
	//    the PayMongo webhook secret.
	mac := hmac.New(sha256.New, []byte(a.webhookSecret))
	_, _ = mac.Write([]byte(signaturePayload))

	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// 7. Compare PayMongo's signature with our calculated signature.
	if !hmac.Equal(
		[]byte(expectedSignature),
		[]byte(providedSignature),
	) {
		return ErrInvalidWebhookSignature
	}

	// 8. The signatures match, so the webhook can be trusted.
	return nil
}

// ParseWebhook validates and extracts the payment information
// that BrewFlow needs from a PayMongo webhook payload.
func (a *Adapter) ParseWebhook(
	payload []byte,
) (*payments.WebhookResult, error) {
	// 1. Decode the JSON payload sent by PayMongo.
	var event webhookEvent

	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("decode webhook: %w", err)
	}

	// 2. Make sure this is the payment-paid event
	//    that BrewFlow currently handles.
	if event.Data.Attributes.Type != "checkout_session.payment.paid" {
		return nil, fmt.Errorf(
			"unsupported webhook event: %s",
			event.Data.Attributes.Type,
		)
	}

	// 3. Get the checkout session data from the event.
	checkout := event.Data.Attributes.Data

	// 4. A checkout session ID is required
	//    to identify the PayMongo checkout.
	if checkout.ID == "" {
		return nil, errors.New("missing checkout session ID")
	}

	// 5. Make sure PayMongo included the payment information.
	if len(checkout.Attributes.Payments) == 0 {
		return nil, errors.New("missing payment")
	}

	// 6. Get the payment from the checkout session.
	payment := checkout.Attributes.Payments[0]

	// 7. Convert PayMongo's payment status into
	//    BrewFlow's internal payment status.
	status := payments.Status(strings.ToUpper(payment.Attributes.Status))

	// 8. Convert PayMongo's source type into
	//    BrewFlow's internal payment method.
	var paymentMethod *payments.PaymentMethod

	if payment.Attributes.Source.Type != "" {
		method := payments.PaymentMethod(
			strings.ToUpper(payment.Attributes.Source.Type),
		)

		paymentMethod = &method
	}

	// 9. Return only the payment information
	//    that the payment service needs.
	return &payments.WebhookResult{
		CheckoutID:      checkout.ID,
		PaymentID:       &payment.ID,
		ReferenceNumber: checkout.Attributes.ReferenceNumber,
		Amount:          payment.Attributes.Amount,
		Currency:        payment.Attributes.Currency,
		Status:          status,
		PaymentMethod:   paymentMethod,
	}, nil
}
