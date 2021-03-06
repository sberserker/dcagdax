package exchanges

import (
	"errors"
	"fmt"
	"os"
	"time"

	exchange "github.com/sberserker/dcagdax/clients/coinbase"
	"github.com/shopspring/decimal"
)

type Coinbase struct {
	client *exchange.Client
}

func NewCoinbase() (*Coinbase, error) {
	secret := os.Getenv("COINBASE_SECRET")
	key := os.Getenv("COINBASE_KEY")
	passphrase := os.Getenv("COINBASE_PASSPHRASE")

	if secret == "" {
		return nil, errors.New("COINBASE_SECRET environment variable is required")
	}

	if key == "" {
		return nil, errors.New("COINBASE_KEY environment variable is required")
	}

	if passphrase == "" {
		return nil, errors.New("COINBASE_PASSPHRASE environment variable is required")
	}

	return &Coinbase{
		client: exchange.NewClient(secret, key, passphrase),
	}, nil
}

func (c *Coinbase) CreateOrder(productId string, amount float64, orderType OrderTypeType, limitOrderFunc CalcLimitOrder) (*Order, error) {

	var orderReq exchange.Order

	if orderType == Limit {
		ticker, err := c.client.GetTicker(productId)
		if err != nil {
			return nil, err
		}

		orderPrice, orderSize := limitOrderFunc(decimal.NewFromFloat(ticker.Ask), decimal.NewFromFloat(amount))
		orderPricef, _ := orderPrice.Float64()
		orderSizef, _ := orderSize.Float64()

		orderReq = exchange.Order{
			ProductId: productId,
			Type:      "limit",
			Side:      "buy",
			Size:      orderSizef,
			Price:     orderPricef,
		}
	} else {
		orderReq = exchange.Order{
			ProductId: productId,
			Type:      "market",
			Side:      "buy",
			Funds:     amount,
		}
	}

	order, err := c.client.CreateOrder(&orderReq)

	if err != nil {
		return nil, err
	}

	return &Order{
		Symbol:  order.ProductId,
		OrderID: order.Id,
	}, nil
}

func (c *Coinbase) GetTickerSymbol(baseCurrency string, quoteCurrency string) string {
	return baseCurrency + "-" + quoteCurrency
}

func (c *Coinbase) GetTicker(productId string) (*Ticker, error) {
	ticker, err := c.client.GetTicker(productId)
	if err != nil {
		return nil, err
	}
	return &Ticker{Price: ticker.Price}, nil
}

func (c *Coinbase) GetProduct(productId string) (*Product, error) {
	product, err := c.client.GetProduct(productId)

	if err != nil {
		return nil, err
	}

	return &Product{
		QuoteCurrency: product.QuoteCurrency,
		BaseCurrency:  product.BaseCurrency,
		BaseMinSize:   product.BaseMinSize,
	}, nil
}

func (c *Coinbase) Deposit(currency string, amount float64) (*time.Time, error) {
	paymentMethods, err := c.client.ListPaymentMethods()

	if err != nil {
		return nil, err
	}

	var bankAccount *exchange.PaymentMethod = nil

	for i := range paymentMethods {
		if paymentMethods[i].Type == "ach_bank_account" {
			bankAccount = &paymentMethods[i]
		}
	}

	if bankAccount == nil {
		return nil, errors.New("No ACH bank account found on this account")
	}

	depositResponse, err := c.client.Deposit(exchange.DepositParams{
		Amount:          amount,
		Currency:        currency,
		PaymentMethodID: bankAccount.ID,
	})

	if err != nil {
		return nil, err
	}

	payoutAt := depositResponse.PayoutAt.Time()
	return &payoutAt, nil
}

func (c *Coinbase) LastPurchaseTime(coin string, currency string, since time.Time) (*time.Time, error) {
	var transactions []exchange.LedgerEntry
	account, err := c.accountFor(coin) //taking the first coins a marker, make sure to put your main coin first
	if err != nil {
		return nil, err
	}
	cursor := c.client.ListAccountLedger(account.Id)

	lastTransactionTime := time.Time{}
	for cursor.HasMore {
		if err := cursor.NextPage(&transactions); err != nil {
			return nil, err
		}

		// Consider trade transactions
		for _, t := range transactions {
			if t.CreatedAt.Time().After(lastTransactionTime) && t.Type == "match" {
				lastTransactionTime = t.CreatedAt.Time()
			}
		}
	}

	return &lastTransactionTime, nil
}

func (c *Coinbase) GetFiatAccount(currency string) (*Account, error) {
	account, err := c.accountFor(currency)
	if err != nil {
		return nil, err
	}

	return &Account{Available: account.Available}, nil
}

func (c *Coinbase) GetPendingTransfers(currency string) ([]PendingTransfer, error) {
	pendingTransfers := []PendingTransfer{}

	account, err := c.accountFor(currency)
	if err != nil {
		return pendingTransfers, err
	}

	// Dang, we don't have enough funds. Let's see if money is on the way.
	var transfers []exchange.Transfer
	cursor := c.client.ListAccountTransfers(account.Id)

	for cursor.HasMore {
		if err := cursor.NextPage(&transfers); err != nil {
			return pendingTransfers, err
		}

		for _, t := range transfers {
			unprocessed := (t.ProcessedAt.Time() == time.Time{})
			notCanceled := (t.CanceledAt.Time() == time.Time{})
			//if it's pending for more than 1 day consider it stuck
			//coinbase sometimes have those issues which support is unable to resolve
			stuck := t.CreatedAt.Time().Before(time.Now().AddDate(0, 0, -1))

			// This transfer is stil pending, so count it.
			if unprocessed && notCanceled && !stuck {
				pendingTransfers = append(pendingTransfers, PendingTransfer{Amount: t.Amount})
			}
		}
	}

	return pendingTransfers, nil
}

func (c *Coinbase) accountFor(currencyCode string) (*exchange.Account, error) {
	accounts, err := c.client.GetAccounts()
	if err != nil {
		return nil, err
	}

	for _, a := range accounts {
		if a.Currency == currencyCode {
			return &a, nil
		}
	}

	return nil, fmt.Errorf("No %s wallet on this account", currencyCode)
}
