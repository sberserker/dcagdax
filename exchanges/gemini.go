package exchanges

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/sberserker/dcagdax/clients/gemini"
)

type Gemini struct {
	client *gemini.Api
}

func NewGemini() (*Gemini, error) {
	key := os.Getenv("GEMINI_KEY")
	secret := os.Getenv("GEMINI_SECRET")

	if key == "" {
		return nil, errors.New("GEMINI_API_KEY environment variable is required")
	}

	if secret == "" {
		return nil, errors.New("GEMINI_API_SECRET environment variable is required")
	}

	api := gemini.New(
		true, // if this is false, it will use Gemini Sandox site: <https://api.sandbox.gemini.com>
		// if this is true,  it will use Gemini Production site: <https://api.gemini.com>
		key,
		secret,
	)

	return &Gemini{
		client: api,
	}, nil
}

func (g *Gemini) GetTickerSymbol(baseCurrency string, quoteCurrency string) string {
	return baseCurrency + quoteCurrency
}

func (g *Gemini) GetTicker(productId string) (*Ticker, error) {
	ticker, err := g.client.TickerV2(productId)
	if err != nil {
		return nil, err
	}

	return &Ticker{
		Price: ticker.Bid,
	}, nil
}

func (g *Gemini) GetProduct(productId string) (Product, error) {
	symbol, err := g.client.SymbolDetails(productId)
	if err != nil {
		return Product{}, err
	}

	return Product{
		QuoteCurrency: symbol.QuoteCurrency,
		BaseCurrency:  symbol.BaseCurrency,
		BaseMinSize:   symbol.MinOrderSize,
	}, nil
}

func (g *Gemini) Deposit(currency string, amount float64) (*time.Time, error) {
	return nil, errors.New("gemini exchange bank deposit is not implemented")
}

func (g *Gemini) CreateOrder(productId string, amount float64) (Order, error) {
	//gemini doesn't support market order type
	//set limit order with high enough price to get filled

	ticker, err := g.client.TickerV2(productId)
	if err != nil {
		return Order{}, err
	}

	//add 2% to make order executed
	price, _ := decimal.NewFromFloat(ticker.Ask + ticker.Ask*2/100).Truncate(2).Float64()
	orderAmt, _ := decimal.NewFromFloat(amount / price).Truncate(8).Float64()

	clientOrderID := uuid.New().String()

	_, err = g.client.NewOrder(productId, clientOrderID, orderAmt, price, "Buy", nil)
	if err != nil {
		return Order{}, err
	}

	return Order{
		Symbol:  productId,
		OrderID: clientOrderID,
	}, nil
}

func (g *Gemini) LastPurchaseTime(ticker string) (*time.Time, error) {
	//past trades history for a given symbol
	//they go in opposite order
	args := gemini.Args{}
	args["limit_trades"] = 10

	lastTransactionTime := time.Time{}

	trades, err := g.client.PastTrades(ticker, args)
	if err != nil {
		return nil, err
	}

	if len(trades) == 0 {
		return &lastTransactionTime, nil
	}

	lastTransactionTime = time.Unix(trades[0].Timestamp, 0)

	return &lastTransactionTime, nil
}

func (g *Gemini) GetFiatAccount(currency string) (*Account, error) {
	balances, err := g.client.Balances()
	if err != nil {
		return nil, err
	}

	var fiatBalance *gemini.FundBalance

	for _, t := range balances {
		if t.Currency == currency {
			fiatBalance = &t
			break
		}
	}

	if fiatBalance == nil {
		return nil, fmt.Errorf("Cannot find %s balance", currency)
	}

	return &Account{Available: fiatBalance.Available}, nil
}

//this is not something gemini can profide
func (g *Gemini) GetPendingTransfers(currency string) ([]PendingTransfer, error) {
	return []PendingTransfer{}, nil
}
