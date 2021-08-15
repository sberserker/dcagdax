package coinbase

type PaymentMethod struct {
	ID        string `json:"id"`
	CreatedAt Time   `json:"created_at,string"`
	UpdatedAt Time   `json:"updated_at,string"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Currency  string `json:"currency"`
}

type DepositParams struct {
	Amount          float64 `json:"amount,string"`
	Currency        string  `json:"currency"`
	PaymentMethodID string  `json:"payment_method_id"`
}

type DepositResponse struct {
	ID       string  `json:"id"`
	Amount   float64 `json:"amount,string"`
	Currency string  `json:"currency"`
	PayoutAt Time    `json:"payout_at,string"`
}

func (c *Client) ListPaymentMethods(p ...ListHoldsParams) ([]PaymentMethod, error) {
	// paginationParams := PaginationParams{}
	// if len(p) > 0 {
	// 	paginationParams = p[0].Pagination
	// }

	// return NewCursor(c, "GET", fmt.Sprintf("/payment-methods", id),
	// 	&paginationParams)

	paymentMethods := []PaymentMethod{}

	_, err := c.Request("GET", "/payment-methods", nil, &paymentMethods)
	return paymentMethods, err
}

func (c *Client) Deposit(deposit DepositParams) (DepositResponse, error) {
	// paginationParams := PaginationParams{}
	// if len(p) > 0 {
	// 	paginationParams = p[0].Pagination
	// }

	// return NewCursor(c, "GET", fmt.Sprintf("/payment-methods", id),
	// 	&paginationParams)

	reponse := DepositResponse{}

	_, err := c.Request("POST", "/deposits/payment-method", deposit, &reponse)
	return reponse, err
}
