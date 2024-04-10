package coinbasev3

// GetPaymentMethods get payment methods.
func (c *ApiClient) GetPaymentMethods() (PaymentMethods, error) {
	u := "https://api.coinbase.com/api/v3/brokerage/payment_methods"

	var result PaymentMethods
	resp, err := c.client.R().
		SetSuccessResult(&result).Get(u)
	if err != nil {
		return result, err
	}

	if !resp.IsSuccessState() {
		return result, ErrFailedToUnmarshal
	}

	return result, nil
}

// PaymentMethods represents the payment methods.
type PaymentMethods struct {
	PaymentMethods []PaymentMethod `json:"payment_methods"`
}

// PaymentMethod represents the payment method.
type PaymentMethod struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Currency string `json:"currency"`
}
