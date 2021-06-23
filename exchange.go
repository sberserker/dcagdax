package main

import (
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/grishinsana/goftx"
)

type Exchagne interface {
	MinimumPurchaseSize(productId string) (float64, error)

	MakePurchase(productId string, amount float64) error

	GetTicker(productId string)

	GetProducts() error

	ListAccountTransfers(accountId string)

	ListAccountLedger(accountId string)

	CreateOrder(productId string, market string)

	Deposit(currency string, amount float64, bankId string)
}

type Ftx struct {
	client *goftx.Client
}

func NewFtx() (*Ftx, error) {
	key := os.Getenv("FTX_KEY")

	secret := os.Getenv("FTX_SECRET")

	client := goftx.New(
		goftx.WithFTXUS(),
		goftx.WithAuth(key, secret),
		goftx.WithHTTPClient(&http.Client{
			Timeout: 5 * time.Second,
		}),
	)

	_, err := client.Account.GetAccountInformation()

	if err != nil {
		return nil, err
	}

	return &Ftx{client: client}, nil
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

func (f *Ftx) GetAccountFiatValue(currency string) (float64, error) {
	info, err := f.client.Account.GetAccountInformation()

	if err != nil {
		return 0, err
	}

	s := f.client.SubAccounts
	sa, err := s.GetSubaccountBalances(info.Username)

	log.Print(sa)

	if err != nil {
		return 0, err
	}

	return 0, nil
}
