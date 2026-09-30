package paymongo

import (
	"context"
	"encoding/json"
	"gobrewflow/internal/services/payments"
	"net/http"
)

var _ payments.PaymentGateway = (*Adapter)(nil)

type Adapter struct {
	client *Client
}

func NewAdapter(client *Client) *Adapter {
	return &Adapter{
		client: client,
	}
}

var _ payments.PaymentGateway = (*Adapter)(nil)

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

	responseBody, err := a.client.do(
		ctx,
		http.MethodPost,
		"/v2/checkout_sessions",
		body,
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
