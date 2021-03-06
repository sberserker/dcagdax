package exchanges

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/grishinsana/goftx"
	"github.com/grishinsana/goftx/models"
	"github.com/shopspring/decimal"
)

type Ftx struct {
	client *goftx.Client
}

func NewFtx(ftxUS bool) (*Ftx, error) {
	key := os.Getenv("FTX_KEY")
	secret := os.Getenv("FTX_SECRET")

	if secret == "" {
		return nil, errors.New("FTX_KEY environment variable is required")
	}
	if key == "" {
		return nil, errors.New("FTX_SECRET environment variable is required")
	}

	var client *goftx.Client

	if ftxUS {
		client = goftx.New(
			goftx.WithFTXUS(),
			goftx.WithAuth(key, secret),
			goftx.WithHTTPClient(&http.Client{
				Timeout: 5 * time.Second,
			}),
		)
	} else {
		client = goftx.New(
			goftx.WithAuth(key, secret),
			goftx.WithHTTPClient(&http.Client{
				Timeout: 5 * time.Second,
			}),
		)
	}

	return &Ftx{client: client}, nil
}

func (f *Ftx) GetTickerSymbol(baseCurrency string, quoteCurrency string) string {
	return baseCurrency + "/" + quoteCurrency
}

func (f *Ftx) GetTicker(productId string) (*Ticker, error) {
	m, err := f.client.Markets.GetMarketByName(productId)

	if err != nil {
		return nil, err
	}

	price, _ := m.Last.Float64()

	return &Ticker{Price: price}, nil
}

func (f *Ftx) GetProduct(productId string) (*Product, error) {
	m, err := f.client.Markets.GetMarketByName(productId)

	if err != nil {
		return nil, err
	}

	minZise, _ := m.MinProvideSize.Float64()

	return &Product{
		QuoteCurrency: m.QuoteCurrency,
		BaseCurrency:  m.BaseCurrency,
		BaseMinSize:   minZise,
	}, nil
}

func (f *Ftx) Deposit(currency string, amount float64) (*time.Time, error) {
	return nil, errors.New("ftx exchange bank deposit is not supported by exchange api")
}

func (f *Ftx) CreateOrder(productId string, amount float64, orderType OrderTypeType, limitOrderFunc CalcLimitOrder) (*Order, error) {

	if orderType == Market {
		return nil, errors.New("ftx market oder type is size based and is not supported use limit order type instead")
	}

	m, err := f.client.Markets.GetMarketByName(productId)

	if err != nil {
		return nil, err
	}

	orderPrice, orderSize := limitOrderFunc(m.Ask, decimal.NewFromFloat(amount))
	clientOrderID := uuid.New().String()

	p := models.PlaceOrderPayload{
		Market:   productId,
		Type:     models.LimitOrder,
		Side:     "buy",
		Size:     orderSize,
		Price:    orderPrice,
		ClientID: &clientOrderID,
	}

	order, err := f.client.PlaceOrder(&p)
	if err != nil {
		return nil, nil
	}

	return &Order{
		OrderID: strconv.FormatInt(order.ID, 10),
	}, nil
}

func (f *Ftx) LastPurchaseTime(ticker string, currency string, since time.Time) (*time.Time, error) {
	product := f.GetTickerSymbol(ticker, currency)
	t := since.Unix()

	p := models.GetFillsParams{Market: &product, StartTime: &t}

	fills, err := f.client.GetFills(&p)
	if err != nil {
		return nil, err
	}

	if len(fills) == 0 {
		return nil, nil
	}

	for _, f := range fills {
		if f.Side == "buy" {
			return &f.Time.Time, nil
		}
	}

	return nil, nil
}

func (f *Ftx) GetFiatAccount(currency string) (*Account, error) {
	balances, err := f.client.GetBalances()
	if err != nil {
		return nil, err
	}

	for _, b := range balances {
		if b.Coin == currency {
			avaialbe, _ := b.Free.Float64()
			return &Account{
				Available: avaialbe,
			}, nil
		}
	}

	return nil, fmt.Errorf("Cannot find %s account", currency)
}

func (f *Ftx) GetPendingTransfers(currency string) ([]PendingTransfer, error) {
	return []PendingTransfer{}, nil
}

func (f *Ftx) MinimumPurchaseSize(productId string) (float64, error) {
	m, err := f.client.Markets.GetMarketByName(productId)

	if err != nil {
		return 0, err
	}

	minSize, _ := m.MinProvideSize.Float64()
	ask, _ := m.Ask.Float64()

	//calculate min order in currency value or return $1
	minOrder := minSize * ask

	return math.Max(minOrder, 1.0), nil
}
