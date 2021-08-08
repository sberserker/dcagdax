package exchanges

import (
	"errors"
	"os"

	exchange "github.com/sberserker/dcagdax/coinbase"
)

type Coinbase struct {
	client *exchange.Client
}

func NewCoinbase() (*Coinbase, error) {
	secret := os.Getenv("GDAX_SECRET")
	key := os.Getenv("GDAX_KEY")
	passphrase := os.Getenv("GDAX_PASSPHRASE")

	if secret == "" {
		return nil, errors.New("GDAX_SECRET environment variable is required")
	} else {
		os.Setenv("COINBASE_SECRET", secret)
	}
	if key == "" {
		return nil, errors.New("GDAX_KEY environment variable is required")
	} else {
		os.Setenv("COINBASE_KEY", key)
	}
	if passphrase == "" {
		return nil, errors.New("GDAX_PASSPHRASE environment variable is required")
	} else {
		os.Setenv("COINBASE_PASSPHRASE", key)
	}

	return &Coinbase{
		client: exchange.NewClient(secret, key, passphrase),
	}, nil
}

func (c *Coinbase) MinimumPurchaseSize(productId string) (float64, error) {
	return 0, nil
}

func (c *Coinbase) MakePurchase(productId string, amount float64) error {
	return nil
}

func (c *Coinbase) GetTicker(productId string) {

}

func (c *Coinbase) GetProducts() error {
	return nil
}

func (c *Coinbase) ListAccountTransfers(accountId string) {

}

func (c *Coinbase) ListAccountLedger(accountId string) {

}

func (c *Coinbase) CreateOrder(productId string, market string) {

}

func (c *Coinbase) Deposit(currency string, amount float64, bankId string) {

}
