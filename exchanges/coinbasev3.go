package exchanges

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	exchange "github.com/sberserker/dcagdax/clients/coinbase"
	"github.com/sberserker/dcagdax/clients/coinbasev3"
	"github.com/shopspring/decimal"
)

type CoinbaseV3 struct {
	client3  *coinbasev3.ApiClient
	client   *exchange.Client
	accounts map[string]*account
}

type account struct {
	Id        string  `json:"id"`
	Hold      float64 `json:"hold,string"`
	Available float64 `json:"available,string"`
	Currency  string  `json:"currency"`
}

func NewCoinbaseV3() (*CoinbaseV3, error) {
	secret := os.Getenv("COINBASE_SECRET")
	key := os.Getenv("COINBASE_KEY")

	if secret == "" {
		return nil, errors.New("COINBASE_SECRET environment variable is required")
	}

	if key == "" {
		return nil, errors.New("COINBASE_KEY environment variable is required")
	}

	client := exchange.NewClient(secret, key, "")
	client.BaseURL = "https://api.coinbase.com/v2"
	client3 := coinbasev3.NewApiClient(key, secret)

	return &CoinbaseV3{
		accounts: map[string]*account{},
		client3:  client3,
		client:   client,
	}, nil
}

func (c *CoinbaseV3) CreateOrder(productId string, amount float64, orderType OrderTypeType, limitOrderFunc CalcLimitOrder) (*Order, error) {

	var orderReq coinbasev3.CreateOrderRequest

	if orderType == Limit {
		trades, err := c.client3.GetMarketTrades(productId, 10)
		if err != nil {
			return nil, err
		}

		bestAsk, err := decimal.NewFromString(trades.BestAsk)
		if err != nil {
			return nil, err
		}
		orderPrice, orderSize := limitOrderFunc(bestAsk, decimal.NewFromFloat(amount))

		orderReq = coinbasev3.CreateOrderRequest{
			ClientOrderID: uuid.NewString(),
			ProductID:     productId,
			Side:          coinbasev3.OrderSideBuy,
			OrderConfiguration: coinbasev3.OrderConfiguration{
				LimitLimitGtc: &coinbasev3.LimitLimitGtc{
					BaseSize:   orderPrice.String(),
					LimitPrice: orderSize.String(),
				},
			},
		}
	} else {
		orderReq = coinbasev3.CreateOrderRequest{
			ClientOrderID: uuid.NewString(),
			ProductID:     productId,
			Side:          coinbasev3.OrderSideBuy,
			OrderConfiguration: coinbasev3.OrderConfiguration{
				MarketMarketIoc: &coinbasev3.MarketMarketIoc{
					QuoteSize: decimal.NewFromFloat(amount).StringFixedBank(2),
				},
			},
		}
	}

	order, err := c.client3.CreateOrder(orderReq)

	if err != nil {
		return nil, err
	}

	if !order.Success {
		return nil, errors.New(fmt.Sprintf("order failed with %s, %s", order.FailureReason, order.ErrorResponse.Message))
	}

	return &Order{
		Symbol:  order.SuccessResponse.ProductId,
		OrderID: order.OrderId,
	}, nil
}

func (c *CoinbaseV3) GetTickerSymbol(baseCurrency string, quoteCurrency string) string {
	return baseCurrency + "-" + quoteCurrency
}

func (c *CoinbaseV3) GetTicker(productId string) (*Ticker, error) {
	ticker, err := c.client3.GetMarketTrades(productId, 10)
	if err != nil {
		return nil, err
	}

	bestAsk, err := strconv.ParseFloat(ticker.BestAsk, 64)
	if err != nil {
		return nil, err
	}

	return &Ticker{Price: bestAsk}, nil
}

func (c *CoinbaseV3) GetProduct(productId string) (*Product, error) {
	product, err := c.client3.GetProduct(productId)

	if err != nil {
		return nil, err
	}

	price, err := strconv.ParseFloat(product.BaseMinSize, 64)
	if err != nil {
		return nil, err
	}

	return &Product{
		QuoteCurrency: product.QuoteCurrencyId,
		BaseCurrency:  product.BaseCurrencyId,
		BaseMinSize:   price,
	}, nil
}

func (c *CoinbaseV3) Deposit(currency string, amount float64) (*time.Time, error) {
	account, err := c.accountFor(currency) //taking the first coins a marker, make sure to put your main coin first
	if err != nil {
		return nil, err
	}

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

	depositResponse, err := c.client.Deposit(account.Id, exchange.DepositParams{
		Amount:          amount,
		Currency:        currency,
		PaymentMethodID: bankAccount.ID,
	})

	if err != nil {
		return nil, err
	}

	payoutAt := depositResponse.Data.PayoutAt
	return &payoutAt, nil
}

func (c *CoinbaseV3) LastPurchaseTime(coin string, currency string, since time.Time) (*time.Time, error) {

	orders, err := c.client3.GetListOrders(coinbasev3.ListOrdersQuery{
		ProductId:   c.GetTickerSymbol(coin, currency),
		StartDate:   since.Format("2006-01-02T15:04:05.999999999Z07:00"),
		OrderStatus: []string{"FILLED"},
	})

	if err != nil {
		return nil, err
	}

	if len(orders.Orders) > 0 {
		return &orders.Orders[0].CreatedTime, nil
	}

	return &time.Time{}, nil
}

func (c *CoinbaseV3) GetFiatAccount(currency string) (*Account, error) {

	account, err := c.accountFor(currency)
	if err != nil {
		return nil, err
	}

	return &Account{Available: account.Available}, nil
}

func (c *CoinbaseV3) GetPendingTransfers(currency string) ([]PendingTransfer, error) {
	pendingTransfers := []PendingTransfer{}
	// // Dang, we don't have enough funds. Let's see if money is on the way.
	// var transfers []exchange.Transfer
	// cursor := c.client.ListAccountTransfers(account.Id)

	// for cursor.HasMore {
	// 	if err := cursor.NextPage(&transfers); err != nil {
	// 		return pendingTransfers, err
	// 	}

	// 	for _, t := range transfers {
	// 		unprocessed := (t.ProcessedAt.Time() == time.Time{})
	// 		notCanceled := (t.CanceledAt.Time() == time.Time{})
	// 		//if it's pending for more than 1 day consider it stuck
	// 		//coinbase sometimes have those issues which support is unable to resolve
	// 		stuck := t.CreatedAt.Time().Before(time.Now().AddDate(0, 0, -1))

	// 		// This transfer is stil pending, so count it.
	// 		if unprocessed && notCanceled && !stuck {
	// 			pendingTransfers = append(pendingTransfers, PendingTransfer{Amount: t.Amount})
	// 		}
	// 	}
	// }
	return pendingTransfers, nil
}

func (c *CoinbaseV3) accountFor(currencyCode string) (*account, error) {

	// cache accounts
	if a, found := c.accounts[currencyCode]; found {
		return a, nil
	}

	accounts, err := c.client3.ListAccounts(100, "")
	if err != nil {
		return nil, err
	}

	for _, a := range accounts.Accounts {
		available, err := strconv.ParseFloat(a.AvailableBalance.Value, 64)
		if err != nil {
			return nil, err
		}

		hold, err := strconv.ParseFloat(a.Hold.Value, 64)
		if err != nil {
			return nil, err
		}

		if a.Currency == currencyCode {
			acct := &account{
				Id:        a.Uuid,
				Currency:  a.Currency,
				Available: available,
				Hold:      hold,
			}

			c.accounts[currencyCode] = acct
			return acct, nil
		}
	}

	return nil, fmt.Errorf("No %s wallet on this account", currencyCode)
}
