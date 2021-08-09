package exchanges

import "time"

type Exchange interface {
	GetTickerSymbol(baseCurrency string, quoteCurrency string) string

	GetTicker(productId string) (*Ticker, error)

	GetProduct(productId string) (Product, error)

	Deposit(currency string, amount float64) (*time.Time, error)

	CreateOrder(productId string, amount float64) (Order, error)

	LastPurchaseTime(ticker string) (*time.Time, error)

	GetFiatAccount(currency string) (*Account, error)

	GetPendingTransfers(currency string) ([]PendingTransfer, error)
}

type Order struct {
	Symbol  string
	OrderID string
}

type Ticker struct {
	Price float64
}

type Product struct {
	QuoteCurrency string
	BaseCurrency  string
	BaseMinSize   float64
}

type Account struct {
	Available float64
}

type PendingTransfer struct {
	Amount float64
}
