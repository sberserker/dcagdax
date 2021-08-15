package exchanges

//go:generate mockgen -destination=../mocks/mock_exchange.go -package=mocks github.com/sberserker/dcagdax/exchanges Exchange

import (
	"time"

	"github.com/shopspring/decimal"
)

type CalcLimitOrder func(askPrice decimal.Decimal, fiatAmount decimal.Decimal) (orderPrice decimal.Decimal, orderSize decimal.Decimal)

type Exchange interface {
	GetTickerSymbol(baseCurrency string, quoteCurrency string) string

	GetTicker(productId string) (*Ticker, error)

	GetProduct(productId string) (*Product, error)

	Deposit(currency string, amount float64) (*time.Time, error)

	CreateOrder(productId string, amount float64, orderType OrderTypeType, limitOrderFunc CalcLimitOrder) (*Order, error)

	LastPurchaseTime(ticker string, currency string, since time.Time) (*time.Time, error)

	GetFiatAccount(currency string) (*Account, error)

	GetPendingTransfers(currency string) ([]PendingTransfer, error)
}

type OrderTypeType int32

const (
	Market OrderTypeType = 0
	Limit  OrderTypeType = 1
)

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
