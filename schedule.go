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

type syncRequest struct {
	usd         float64
	orderSpread float64
	orderType   exchanges.OrderTypeType
	fee         float64
	every       time.Duration
	until       time.Time
	after       time.Time
	autoFund    bool
	force       bool
	coins       []string
	currency    string
}

type orderDetails struct {
	symbol string
	amount float64
}

type gdaxSchedule struct {
	logger    *zap.SugaredLogger
	exchange  exchanges.Exchange
	debug     bool
	req       syncRequest
	coins     map[string]orderDetails
	sleepFunc func(time.Duration)
}

func newGdaxSchedule(
	exchange exchanges.Exchange,
	l *zap.SugaredLogger,
	debug bool,
	syncRequest syncRequest,
) (*gdaxSchedule, error) {
	schedule := gdaxSchedule{
		logger:   l,
		exchange: exchange,
		debug:    debug,

		req:       syncRequest,
		coins:     map[string]orderDetails{},
		sleepFunc: sleep,
	}

	total := 0

	for _, c := range syncRequest.coins {
		arr := strings.Split(c, ":")
		coin := arr[0]
		percentage, err := strconv.Atoi(arr[1])
		if err != nil {
			return &schedule, err
		}

		total += int(percentage)

		symbol := exchange.GetTickerSymbol(coin, schedule.req.currency)
		minimum, err := schedule.minimumUSDPurchase(symbol)

		if err != nil {
			return nil, err
		}

		if schedule.req.usd == 0.0 {
			schedule.req.usd = minimum + 0.1
		}

		//schedule.usd * percentage / 100
		scheduledForCoin, _ := decimal.NewFromFloat(schedule.req.usd).Mul(decimal.NewFromFloat(float64(percentage))).Div(decimal.NewFromFloat(100)).Truncate(2).Float64()

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

	until := s.req.until
	if until.IsZero() {
		until = time.Now()
	}

	if now.After(until) {
		return errors.New("Deadline has passed, not taking any action")
	}

	if !s.req.after.IsZero() && !now.After(s.req.after) {
		return fmt.Errorf("Configured to start after %s, not taking any action", s.req.after)
	}

	s.logger.Infow("Dollar cost averaging",
		s.req.currency, s.req.usd,
		"every", every,
		"until", until.String(),
	)

	since := now.Add(-*every)

	if s.req.force != true {
		if time, err := s.timeToPurchase(since); err != nil {
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

		if !s.req.autoFund {
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
			s.sleepFunc(waitTime)
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

func (s *gdaxSchedule) timeToPurchase(since time.Time) (bool, error) {
	timeSinceLastPurchase, err := s.timeSinceLastPurchase(since)

	if err != nil {
		return false, err
	}

	if timeSinceLastPurchase == nil {
		return true, nil
	}

	s.logger.Infow(
		"Time since last purchase hours",
		"hours", timeSinceLastPurchase.Hours(),
	)

	if timeSinceLastPurchase.Seconds() < s.req.every.Seconds() {
		// We purchased something recently, so hang tight.
		return false, nil
	}

	return true, nil
}

func (s *gdaxSchedule) additionalUsdNeeded() (float64, error) {
	usdAccount, err := s.exchange.GetFiatAccount(s.req.currency)
	if err != nil {
		return 0, err
	}

	if usdAccount.Available >= s.req.usd {
		return 0, nil
	}

	availableBalance := decimal.NewFromFloat(usdAccount.Available).Truncate(2)

	s.logger.Infow(
		"Avaialable balance",
		"amount", availableBalance.String(),
	)

	//account may have some fraction of cents from previous trading so cut everything after 0.01
	//s.usd - availableBalance
	dollarsNeeded, _ := decimal.NewFromFloat(s.req.usd).Sub(availableBalance).Truncate(2).Float64()

	return dollarsNeeded, nil
}

func (s *gdaxSchedule) pendingTransfers() (float64, error) {
	transfers, err := s.exchange.GetPendingTransfers(s.req.currency)
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

func (s *gdaxSchedule) timeSinceLastPurchase(since time.Time) (*time.Duration, error) {
	coins := make([]string, 0, len(s.coins))
	for k := range s.coins {
		coins = append(coins, k)
	}

	lastPurchaseTime, err := s.exchange.LastPurchaseTime(coins[0], s.req.currency, since) //taking the first coins a marker, make sure to put your main coin first

	if err != nil {
		return nil, err
	}

	if lastPurchaseTime == nil {
		s.logger.Infow(
			"No transactions found since",
			"since", since.Local(),
		)
		return nil, nil
	}

	s.logger.Infow(
		"Last transaction time",
		"time", lastPurchaseTime.Local(),
	)

	timeSinceLastPurchase := time.Now().Sub(*lastPurchaseTime)
	return &timeSinceLastPurchase, nil
}

func (s *gdaxSchedule) makePurchase(productId string, amount float64) error {
	if s.debug {
		return skippedForDebug
	}

	order, err := s.exchange.CreateOrder(productId, amount, s.req.orderType, s.calcLimitOrder)

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

	payoutAt, err := s.exchange.Deposit(s.req.currency, amount)

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

func (s *gdaxSchedule) calcLimitOrder(askPrice decimal.Decimal, fiatAmount decimal.Decimal) (orderPrice decimal.Decimal, orderSize decimal.Decimal) {

	//reduce fiat Amount to include fees %
	//(1-fee)/100 * fiatAmount
	fiatAmount = decimal.NewFromFloat((100 - s.req.fee) / 100).Mul(fiatAmount)

	spread := decimal.NewFromFloat(s.req.orderSpread)

	//calc order price
	//ask * spread / 100 + ask
	orderPrice = askPrice.Mul(spread).Div(decimal.NewFromInt32(100)).Add(askPrice).Truncate(2)

	//order size
	//fiatAmount / orderPrice
	orderSize = fiatAmount.Div(orderPrice).Truncate(8)

	s.logger.Infow(
		"Limit order",
		"size", orderSize.String(),
		"price", orderPrice.String(),
	)

	return orderPrice, orderSize
}

func sleep(waitTime time.Duration) {
	time.Sleep(waitTime)
}
