package exchanges

type Exchange interface {
	MinimumPurchaseSize(productId string) (float64, error)

	MakePurchase(productId string, amount float64) error

	GetTicker(productId string)

	GetProducts() error

	ListAccountTransfers(accountId string)

	ListAccountLedger(accountId string)

	CreateOrder(productId string, market string)

	Deposit(currency string, amount float64, bankId string)
}
