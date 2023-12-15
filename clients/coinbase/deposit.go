package coinbase

import (
	"fmt"
	"time"
)

type ListPaymentMethod struct {
	Data []PaymentMethod `json:"data"`
}

type PaymentMethod struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at,string"`
	UpdatedAt time.Time `json:"updated_at,string"`
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	Currency  string    `json:"currency"`
}

type DepositParams struct {
	Amount          float64 `json:"amount,string"`
	Currency        string  `json:"currency"`
	PaymentMethodID string  `json:"payment_method"`
}

type DepositResponse struct {
	Data Deposit `json:"data"`
}

func (c *Client) ListPaymentMethods() ([]PaymentMethod, error) {
	paymentMethods := ListPaymentMethod{}

	_, err := c.Request("GET", "/payment-methods", nil, &paymentMethods)
	return paymentMethods.Data, err
}

func (c *Client) Deposit(accountId string, deposit DepositParams) (DepositResponse, error) {
	response := DepositResponse{}

	_, err := c.Request("POST", fmt.Sprintf("/accounts/%s/deposits", accountId), deposit, &response)
	return response, err
}

type ListDeposits struct {
	Data []Deposit `json:"data"`
}

type Deposit struct {
	Amount    Amount    `json:"amount"`
	Subtotal  Amount    `json:"subtotal"`
	Fee       Amount    `json:"fee"`
	CreatedAt time.Time `json:"created_at,string"`
	UpdatedAt time.Time `json:"updated_at,string"`
	PayoutAt  time.Time `json:"payout_at,string"`
}

type Amount struct {
	Amount   float64 `json:"amount,string"`
	Currency string  `json:"currency"`
}

func (c *Client) ListDeposits(id string) ([]Deposit, error) {
	var response ListDeposits
	_, err := c.Request("GET", fmt.Sprintf("/accounts/%s/deposits", id), nil, &response)

	return response.Data, err
}
