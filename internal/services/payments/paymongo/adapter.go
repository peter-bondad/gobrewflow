package paymongo

import (
	"context"
	"encoding/json"
	"gobrewflow/internal/services/payments"
	"gobrewflow/shared"
)

var _ payments.PaymentGateway = (*Adapter)(nil)

type Adapter struct {
	client        *Client
	webhookSecret string
}

func NewAdapter(
	client *Client,
	webhookSecret string,
) *Adapter {
	return &Adapter{
		client:        client,
		webhookSecret: webhookSecret,
	}
}

func (a *Adapter) CreateCheckout(
	ctx context.Context,
	input payments.CreateCheckoutInput,
) (*payments.CreateCheckoutOutput, error) {
	req := createCheckoutRequest{
		Data: createCheckoutData{
			Attributes: createCheckoutAttributes{
				ReferenceNumber: input.ReferenceID,
				LineItems: []checkoutLineItem{
					{
						Name:     input.Description,
						Quantity: 1,
						Amount:   input.Amount,
						Currency: input.Currency,
					},
				},
				PaymentMethodTypes: []string{
					"gcash",
				},
				SuccessURL: input.SuccessURL,
				CancelURL:  input.CancelURL,
			},
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	idempotencyKey := shared.GenerateIdempotencyKey()

	responseBody, err := a.client.DoCreateCheckout(
		ctx,
		body,
		idempotencyKey,
	)
	if err != nil {
		return nil, err
	}

	var res createCheckoutResponse

	if err := json.Unmarshal(responseBody, &res); err != nil {
		return nil, err

	}

	return &payments.CreateCheckoutOutput{
		CheckoutID:  res.Data.ID,
		CheckoutURL: res.Data.Attributes.CheckoutURL,
	}, nil

}
