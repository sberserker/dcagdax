package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	exchange "github.com/blampe/dcagdax/coinbase"
	"go.uber.org/zap"
)

var skippedForDebug = errors.New("Skipping because trades are not enabled")

type gdaxSchedule struct {
	logger *zap.SugaredLogger
	client *exchange.Client
	debug  bool

	usd      float64
	every    time.Duration
	until    time.Time
	autoFund bool
	coins    map[string]float64
	force    bool
}

func newGdaxSchedule(
	c *exchange.Client,
	l *zap.SugaredLogger,
	debug bool,
	autoFund bool,
	usd float64,
	every time.Duration,
	until time.Time,
	coins []string,
	force bool,
) (*gdaxSchedule, error) {
	schedule := gdaxSchedule{
		logger: l,
		client: c,
		debug:  debug,

		usd:      usd,
		every:    every,
		until:    until,
		autoFund: autoFund,
		coins:    map[string]float64{},
		force:    force,
	}

	total := 0

	for _, c := range coins {
		arr := strings.Split(c, ":")
		coin := arr[0]
		percentage, err := strconv.Atoi(arr[1])
		if err != nil {
			return &schedule, err
		}

		total += int(percentage)

		minimum, err := schedule.minimumUSDPurchase(coin)

		if err != nil {
			return nil, err
		}

		if schedule.usd == 0.0 {
			schedule.usd = minimum + 0.1
		}

		scheduledForCoin := roundFloat(schedule.usd*float64(percentage)/100, 8)
		schedule.coins[coin] = scheduledForCoin

		if scheduledForCoin < minimum {
			return nil, fmt.Errorf(
				"Coinbase minimum %s trade amount is $%.02f, but you're trying to purchase $%f",
				coin, minimum, scheduledForCoin,
			)
		}
	}

	if total != 100 {
		return nil, fmt.Errorf("selected percentage must be exactly 100, provided %d", total)
	}

	return &schedule, nil
}

func roundFloat(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return math.Floor(f*shift+.5) / shift
}

// Sync initiates trades & funding with a DCA strategy.
func (s *gdaxSchedule) Sync() error {

	now := time.Now()

	until := s.until
	if until.IsZero() {
		until = time.Now()
	}

	if now.After(until) {
		return errors.New("Deadline has passed, not taking any action")
	}

	s.logger.Infow("Dollar cost averaging",
		"USD", s.usd,
		"every", every,
		"until", until.String(),
	)

	if s.force != true {
		if time, err := s.timeToPurchase(); err != nil {
			return err
		} else if !time {
			return errors.New("Detected a recent purchase, waiting for next purchase window")
		}
	} else {
		c := askForConfirmation("Force method is used proceed?")
		if !c {
			return errors.New("User regected the trade")
		}
	}

	if funded, err := s.sufficientUsdAvailable(); err != nil {
		return err
	} else if !funded {
		needed, err := s.additionalUsdNeeded()
		if err != nil {
			return err
		}

		if needed == 0 {
			return errors.New("Not enough available funds, wait for transfers to settle")
		}

		if needed > 0 {
			s.logger.Infow(
				"Insufficient funds",
				"needed", needed,
			)
			if s.autoFund {
				s.logger.Infow(
					"Creating a transfer request for $%.02f",
					"needed", needed,
				)

				if s.debug != true {
					err := s.makeDeposit(needed)
					if err != nil {
						return err
					}
				} else {
					s.logger.Infow("Deposit skipped for debug")
				}
			}
		}
		return nil
	}

	s.logger.Infow(
		"Placing an order",
		"coins", s.coins,
		"purchaseCurrency", "USD",
		"purchaseAmount", s.usd,
	)

	for coin, amount := range s.coins {
		productId := coin + "-" + "USD"
		if err := s.makePurchase(productId, amount); err != nil {
			s.logger.Warn(err)
		}
	}

	return nil
}

func (s *gdaxSchedule) minimumUSDPurchase(coin string) (float64, error) {
	productId := coin + "-" + "USD"
	ticker, err := s.client.GetTicker(productId)

	if err != nil {
		return 0, err
	}

	products, err := s.client.GetProducts()

	if err != nil {
		return 0, err
	}

	for _, p := range products {
		if p.BaseCurrency == coin {
			return math.Max(p.BaseMinSize*ticker.Price, 1.0), nil
		}
	}

	return 0, errors.New(productId + " not found")
}

func (s *gdaxSchedule) timeToPurchase() (bool, error) {
	timeSinceLastPurchase, err := s.timeSinceLastPurchase()

	if err != nil {
		return false, err
	}

	if timeSinceLastPurchase.Seconds() < s.every.Seconds() {
		// We purchased something recently, so hang tight.
		return false, nil
	}

	return true, nil
}

func (s *gdaxSchedule) sufficientUsdAvailable() (bool, error) {
	usdAccount, err := s.accountFor("USD")

	if err != nil {
		return false, err
	}

	return (usdAccount.Available >= s.usd), nil
}

func (s *gdaxSchedule) additionalUsdNeeded() (float64, error) {
	if funded, err := s.sufficientUsdAvailable(); err != nil {
		return 0, err
	} else if funded {
		return 0, nil
	}

	usdAccount, err := s.accountFor("USD")
	if err != nil {
		return 0, nil
	}

	dollarsNeeded := s.usd - usdAccount.Available
	if dollarsNeeded < 0 {
		return 0, errors.New("Invalid account balance")
	}

	// Dang, we don't have enough funds. Let's see if money is on the way.
	var transfers []exchange.Transfer
	cursor := s.client.ListAccountTransfers(usdAccount.Id)

	dollarsInbound := 0.0

	for cursor.HasMore {
		if err := cursor.NextPage(&transfers); err != nil {
			return 0, err
		}

		for _, t := range transfers {
			unprocessed := (t.ProcessedAt.Time() == time.Time{})
			notCanceled := (t.CanceledAt.Time() == time.Time{})

			// This transfer is stil pending, so count it.
			if unprocessed && notCanceled {
				s.logger.Infow(
					"Deposit is in progress",
					"amount", t.Amount,
				)
				dollarsInbound += t.Amount
			}
		}
	}

	// If our incoming transfers don't cover our purchase need then we'll need
	// to cover that with an additional deposit.
	return math.Max(dollarsNeeded-dollarsInbound, 0), nil
}

func (s *gdaxSchedule) timeSinceLastPurchase() (time.Duration, error) {
	coins := make([]string, 0, len(s.coins))
	for k := range s.coins {
		coins = append(coins, k)
	}

	var transactions []exchange.LedgerEntry
	account, err := s.accountFor(coins[0]) //taking the first coins a marker, make sure to put your main coin first
	if err != nil {
		return 0, err
	}
	cursor := s.client.ListAccountLedger(account.Id)

	lastTransactionTime := time.Time{}
	now := time.Now()

	for cursor.HasMore {
		if err := cursor.NextPage(&transactions); err != nil {
			return 0, err

		}

		// Consider trade transactions
		for _, t := range transactions {
			if t.CreatedAt.Time().After(lastTransactionTime) && t.Type == "match" {
				lastTransactionTime = t.CreatedAt.Time()
			}
		}
	}

	return now.Sub(lastTransactionTime), nil
}

func (s *gdaxSchedule) makePurchase(productId string, amount float64) error {
	if s.debug {
		return skippedForDebug
	}

	order, err := s.client.CreateOrder(
		&exchange.Order{
			ProductId: productId,
			Type:      "market",
			Side:      "buy",
			Funds:     amount, // Coinbase has a limit of 8 decimal places.
		},
	)

	if err != nil {
		return err
	}

	s.logger.Infow(
		"Placed order",
		"productId", productId,
		"orderId", order.Id,
	)

	return nil
}

func (s *gdaxSchedule) makeDeposit(amount float64) error {
	paymentMethods, err := s.client.ListPaymentMethods()

	if err != nil {
		return err
	}

	var bankAccount *exchange.PaymentMethod = nil

	for _, p := range paymentMethods {
		if p.Type == "ach_bank_account" {
			bankAccount = &p
		}
	}

	if bankAccount == nil {
		return errors.New("No ACH bank account found on this account")
	}

	depositResponse, err := s.client.Deposit(exchange.DepositParams{
		Amount:          amount,
		Currency:        "USD",
		PaymentMethodID: bankAccount.ID,
	})

	if err != nil {
		return err
	}

	s.logger.Infow(
		"Deposit initiated successfully",
		"payout", depositResponse.PayoutAt,
	)

	return nil
}

func (s *gdaxSchedule) accountFor(currencyCode string) (*exchange.Account, error) {
	accounts, err := s.client.GetAccounts()
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

func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}
