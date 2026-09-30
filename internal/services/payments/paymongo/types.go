package paymongo

// PayMongo Request

type createCheckoutRequest struct {
	Data createCheckoutData `json:"data"`
}

type createCheckoutData struct {
	Attributes createCheckoutAttributes `json:"attributes"`
}

type createCheckoutAttributes struct {
	ReferenceNumber    string             `json:"reference_number"`
	LineItems          []checkoutLineItem `json:"line_items"`
	PaymentMethodTypes []string           `json:"payment_method_types"`
	SuccessURL         string             `json:"success_url"`
	CancelURL          string             `json:"cancel_url"`
}

type checkoutLineItem struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

// PayMongo Response

type createCheckoutResponse struct {
	Data checkoutData `json:"data"`
}

type checkoutData struct {
	ID         string             `json:"id"`
	Type       string             `json:"type"`
	Attributes checkoutAttributes `json:"attributes"`
}

type checkoutAttributes struct {
	CheckoutURL string `json:"checkout_url"`
}
