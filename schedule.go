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

	"go.uber.org/zap"

	"github.com/sberserker/dcagdax/exchanges"
	"github.com/shopspring/decimal"
)

var skippedForDebug = errors.New("Skipping because trades are not enabled")

const (
	Currency = "USD"
)

type orderDetails struct {
	symbol string
	amount float64
}

type gdaxSchedule struct {
	logger   *zap.SugaredLogger
	exchange exchanges.Exchange
	debug    bool

	usd      float64
	every    time.Duration
	until    time.Time
	after    time.Time
	autoFund bool
	coins    map[string]orderDetails
	force    bool
}

func newGdaxSchedule(
	exchange exchanges.Exchange,
	l *zap.SugaredLogger,
	debug bool,
	autoFund bool,
	usd float64,
	every time.Duration,
	until time.Time,
	after time.Time,
	coins []string,
	force bool,
) (*gdaxSchedule, error) {
	schedule := gdaxSchedule{
		logger:   l,
		exchange: exchange,
		debug:    debug,

		usd:      usd,
		every:    every,
		until:    until,
		after:    after,
		autoFund: autoFund,
		coins:    map[string]orderDetails{},
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

		symbol := exchange.GetTickerSymbol(coin, Currency)
		minimum, err := schedule.minimumUSDPurchase(symbol)

		if err != nil {
			return nil, err
		}

		if schedule.usd == 0.0 {
			schedule.usd = minimum + 0.1
		}

		//schedule.usd * percentage / 100
		scheduledForCoin, _ := decimal.NewFromFloat(usd).Mul(decimal.NewFromFloat(float64(percentage))).Div(decimal.NewFromFloat(100)).Truncate(2).Float64()

		order := orderDetails{
			symbol: symbol,
			amount: scheduledForCoin,
		}

		schedule.coins[coin] = order

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

	if !s.after.IsZero() && !now.After(s.after) {
		return fmt.Errorf("Configured to start after %s, not taking any action", after)
	}

	s.logger.Infow("Dollar cost averaging",
		Currency, s.usd,
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
			return errors.New("User rejected the trade")
		}
	}

	needed, err := s.additionalUsdNeeded()
	if err != nil {
		return err
	}

	//check if there are pending transfers
	//typically pending transfers means something is stuck, need to wait to settle or resolve the issue
	if needed > 0 {
		pending, err := s.pendingTransfers()
		if err != nil {
			return err
		}

		if pending > 0 {
			return errors.New("Wait for transfers to settle")
		}

		s.logger.Infow(
			"Insufficient funds",
			"needed", needed,
		)

		if !s.autoFund {
			return errors.New("No sufficient amount for trade and autofund is disabled. Deposit money to proceed")
		}

		payoutAt, err := s.fund(needed)
		if err != nil {
			return err
		}

		//calculate when money will available to buy and sleep
		waitTime := payoutAt.Add(1 * time.Minute).Sub(time.Now())
		if waitTime < 2*time.Minute {
			s.logger.Infow(
				"Sleeping for",
				"minutes", waitTime.Minutes(),
			)
			time.Sleep(waitTime)
		} else {
			s.logger.Infow(
				"Deposit money will ba available in. Exiting now",
				"minutes", waitTime.Minutes(),
			)
			return nil
		}
	}

	for _, order := range s.coins {
		s.logger.Infow(
			"Placing an order",
			"productId", order.symbol,
			"amount", order.amount,
		)

		if err := s.makePurchase(order.symbol, order.amount); err != nil {
			s.logger.Warn(err)
		}
	}

	return nil
}

func (s *gdaxSchedule) fund(needed float64) (*time.Time, error) {
	s.logger.Infow(
		"Creating a transfer request for $%.02f",
		"needed", needed,
	)

	if s.debug {
		s.logger.Infow("Deposit skipped for debug")
		now := time.Now()
		return &now, nil
	}

	payoutAt, err := s.makeDeposit(needed)
	if err != nil {
		return nil, err
	}

	return payoutAt, nil
}

func (s *gdaxSchedule) minimumUSDPurchase(productId string) (float64, error) {
	product, err := s.exchange.GetProduct(productId)
	if err != nil {
		return 0, err
	}

	ticker, err := s.exchange.GetTicker(productId)

	if err != nil {
		return 0, err
	}

	return math.Max(product.BaseMinSize*ticker.Price, 1.0), nil
}

func (s *gdaxSchedule) timeToPurchase() (bool, error) {
	timeSinceLastPurchase, err := s.timeSinceLastPurchase()

	if err != nil {
		return false, err
	}

	s.logger.Infow(
		"Time since last purchase hours",
		"hours", timeSinceLastPurchase.Hours(),
	)

	if timeSinceLastPurchase.Seconds() < s.every.Seconds() {
		// We purchased something recently, so hang tight.
		return false, nil
	}

	return true, nil
}

func (s *gdaxSchedule) additionalUsdNeeded() (float64, error) {
	usdAccount, err := s.exchange.GetFiatAccount(Currency)
	if err != nil {
		return 0, err
	}

	if usdAccount.Available >= s.usd {
		return 0, nil
	}

	availableBalance := decimal.NewFromFloat(usdAccount.Available).Truncate(2)

	s.logger.Infow(
		"Avaialable balance",
		"amount", availableBalance.String(),
	)

	//account may have some fraction of cents from previous trading so cut everything after 0.01
	//s.usd - availableBalance
	dollarsNeeded, _ := decimal.NewFromFloat(s.usd).Sub(availableBalance).Truncate(2).Float64()

	return dollarsNeeded, nil
}

func (s *gdaxSchedule) pendingTransfers() (float64, error) {
	transfers, err := s.exchange.GetPendingTransfers(Currency)
	if err != nil {
		return 0, err
	}
	if transfers == nil {
		return 0, nil
	}

	dollarsInbound := 0.0

	for _, t := range transfers {
		s.logger.Infow(
			"Deposit is in progress",
			"amount", t.Amount,
		)
		dollarsInbound += t.Amount
	}

	return dollarsInbound, nil
}

func (s *gdaxSchedule) timeSinceLastPurchase() (time.Duration, error) {
	coins := make([]string, 0, len(s.coins))
	for k := range s.coins {
		coins = append(coins, k)
	}

	lastPurchaseTime, err := s.exchange.LastPurchaseTime(coins[0]) //taking the first coins a marker, make sure to put your main coin first

	if err != nil {
		return 0, err
	}

	s.logger.Infow(
		"Last transaction time",
		"time", lastPurchaseTime.Local(),
	)

	return time.Now().Sub(*lastPurchaseTime), nil
}

func (s *gdaxSchedule) makePurchase(productId string, amount float64) error {
	if s.debug {
		return skippedForDebug
	}

	order, err := s.exchange.CreateOrder(productId, amount)

	if err != nil {
		return err
	}

	s.logger.Infow(
		"Placed order",
		"orderId", order.OrderID,
	)

	return nil
}

func (s *gdaxSchedule) makeDeposit(amount float64) (*time.Time, error) {

	payoutAt, err := s.exchange.Deposit(Currency, amount)

	if err != nil {
		return nil, err
	}

	s.logger.Infow(
		"Deposit initiated successfully",
		"payout", payoutAt,
	)

	return payoutAt, nil
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
