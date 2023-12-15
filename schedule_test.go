package main

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sberserker/dcagdax/exchanges"
	"github.com/sberserker/dcagdax/mocks"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func loggerStub(t *testing.T) *zap.Logger {
	return zaptest.NewLogger(t, zaptest.WrapOptions(zap.Hooks(func(e zapcore.Entry) error {
		return nil
	})))
}

func TestCalcLimitOrder(t *testing.T) {
	type test struct {
		askPrice      float64
		fiatAmout     float64
		fee           float64
		spread        float64
		outOrderPrise float64
		outOrderSize  float64
	}

	tests := []test{
		{askPrice: 1, fiatAmout: 10, fee: 1, spread: 1, outOrderPrise: 1.01, outOrderSize: 9.80198019},
	}

	for _, tc := range tests {
		s := gdaxSchedule{}
		s.req = syncRequest{fee: tc.fee, orderSpread: tc.spread}
		s.logger = loggerStub(t).Sugar()

		orderPrice, orderSize := s.calcLimitOrder(decimal.NewFromFloat(tc.askPrice), decimal.NewFromFloat(tc.fiatAmout))

		orderPricef, _ := orderPrice.Float64()
		orderSizef, _ := orderSize.Float64()

		assert.Equal(t, tc.outOrderPrise, orderPricef)
		assert.Equal(t, tc.outOrderSize, orderSizef)
	}
}

func TestSyncWhenNotAGoodTime(t *testing.T) {
	type test struct {
		until   time.Time
		after   time.Time
		err     string
		message string
	}

	afterTomorrow := time.Now().AddDate(0, 0, 1)
	afterTomrrowMesage := fmt.Sprintf("Configured to start after %v, not taking any action", afterTomorrow)

	tests := []test{
		{until: time.Time{}, after: afterTomorrow, err: afterTomrrowMesage, message: "not yet time to run"},
		{until: time.Now().AddDate(0, 0, -1), after: time.Time{}, err: "Deadline has passed, not taking any action", message: "time to run passed"},
	}

	for _, tc := range tests {
		s := gdaxSchedule{}
		s.logger = loggerStub(t).Sugar()
		s.req = syncRequest{until: tc.until, after: tc.after}

		err := s.Sync()

		assert.Equal(t, tc.err, err.Error(), tc.message)
	}
}

func TestWhenRecentPurchase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockExchange(ctrl)

	s := gdaxSchedule{}
	s.logger = loggerStub(t).Sugar()
	s.req = syncRequest{every: 24 * time.Hour, currency: "USD"} // setup run every 24 hrs
	s.markerCoin = "BTC"
	s.exchange = m

	t.Run("when recent purchase", func(t *testing.T) {
		//but last run was 12 hours ago
		lastPurchaseTime := time.Now().Add(-12 * time.Hour)
		m.EXPECT().LastPurchaseTime("BTC", "USD", gomock.Any()).Return(&lastPurchaseTime, nil)

		err := s.Sync()

		assert.Equal(t, "Detected a recent purchase, waiting for next purchase window", err.Error())
	})

	t.Run("when recent purchase falsed", func(t *testing.T) {
		//but last run was 12 hours ago
		lastPurchaseTime := time.Now().Add(-12 * time.Hour)
		m.EXPECT().LastPurchaseTime("BTC", "USD", gomock.Any()).Return(&lastPurchaseTime, errors.New("some error"))

		err := s.Sync()

		assert.Equal(t, "some error", err.Error())
	})
}
func TestTimeToPurchase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockExchange(ctrl)

	s := gdaxSchedule{}
	s.logger = loggerStub(t).Sugar()
	s.req = syncRequest{every: 24 * time.Hour, currency: "USD"} // setup run every 24 hrs
	s.markerCoin = "BTC"
	s.exchange = m

	t.Run("when recent purchase", func(t *testing.T) {
		lastPurchaseTime := time.Now().Add(-12 * time.Hour) //last purchase time 12 hrs ago
		m.EXPECT().LastPurchaseTime("BTC", "USD", gomock.Any()).Return(&lastPurchaseTime, nil)

		result, err := s.timeToPurchase(time.Now().Add(-24 * time.Hour))

		assert.False(t, result)
		assert.Nil(t, err)
	})

	t.Run("when no recent purchase", func(t *testing.T) {
		lastPurchaseTime := time.Now().Add(-48 * time.Hour) //last purchase time 2 days ago
		m.EXPECT().LastPurchaseTime("BTC", "USD", gomock.Any()).Return(&lastPurchaseTime, nil)

		result, err := s.timeToPurchase(time.Now().Add(-24 * time.Hour))

		assert.True(t, result)
		assert.Nil(t, err)
	})
}

func TestNewSchedule(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockExchange(ctrl)

	t.Run("when success", func(t *testing.T) {
		req := syncRequest{every: 24 * time.Hour, orderType: exchanges.Market, autoFund: true, currency: "USD", usd: 50, coins: []string{"BTC:50", "ETH:50"}} // setup run every 24 hrs

		m.EXPECT().GetTickerSymbol("BTC", "USD").Return("BTC:USD")
		m.EXPECT().GetProduct("BTC:USD").Return(&exchanges.Product{BaseMinSize: 0.001}, nil)
		m.EXPECT().GetTicker("BTC:USD").Return(&exchanges.Ticker{Price: 1000}, nil)

		m.EXPECT().GetTickerSymbol("ETH", "USD").Return("ETH:USD")
		m.EXPECT().GetProduct("ETH:USD").Return(&exchanges.Product{BaseMinSize: 0.5}, nil)
		m.EXPECT().GetTicker("ETH:USD").Return(&exchanges.Ticker{Price: 10}, nil)

		s, err := newGdaxSchedule(m, loggerStub(t).Sugar(), false, req)

		assert.Nil(t, err)
		assert.NotNil(t, s)
		assert.Equal(t, 25.0, s.coins["BTC"].amount)
		assert.Equal(t, "BTC:USD", s.coins["BTC"].symbol)
		assert.Equal(t, 25.0, s.coins["ETH"].amount)
		assert.Equal(t, "ETH:USD", s.coins["ETH"].symbol)
	})

	t.Run("when unbalanced coins request", func(t *testing.T) {
		req := syncRequest{every: 24 * time.Hour, orderType: exchanges.Market, autoFund: true, currency: "USD", usd: 50, coins: []string{"BTC:50", "ETH:49"}} // setup run every 24 hrs

		m.EXPECT().GetTickerSymbol("BTC", "USD").Return("BTC:USD")
		m.EXPECT().GetProduct("BTC:USD").Return(&exchanges.Product{BaseMinSize: 0.001}, nil)
		m.EXPECT().GetTicker("BTC:USD").Return(&exchanges.Ticker{Price: 1000}, nil)

		m.EXPECT().GetTickerSymbol("ETH", "USD").Return("ETH:USD")
		m.EXPECT().GetProduct("ETH:USD").Return(&exchanges.Product{BaseMinSize: 0.5}, nil)
		m.EXPECT().GetTicker("ETH:USD").Return(&exchanges.Ticker{Price: 10}, nil)

		s, err := newGdaxSchedule(m, loggerStub(t).Sugar(), false, req)

		assert.NotNil(t, err)
		assert.Nil(t, s)
		assert.Equal(t, "Total percentages must be exactly 100, provided 99", err.Error())
	})
}

func TestNewScheduleWhenBelowMinSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockExchange(ctrl)
	req := syncRequest{every: 24 * time.Hour, orderType: exchanges.Market, autoFund: true, currency: "USD", usd: 50, coins: []string{"BTC:50"}} // setup run every 24 hrs

	m.EXPECT().GetTickerSymbol("BTC", "USD").Return("BTC:USD")
	m.EXPECT().GetProduct("BTC:USD").Return(&exchanges.Product{BaseMinSize: 0.01}, nil)
	m.EXPECT().GetTicker("BTC:USD").Return(&exchanges.Ticker{Price: 10000}, nil)

	s, err := newGdaxSchedule(m, loggerStub(t).Sugar(), false, req)

	assert.Nil(t, s)
	assert.NotNil(t, err)
	assert.Equal(t, "Coinbase minimum BTC trade amount is $100.00, but you're trying to purchase $25.00", err.Error())
}

func TestSyncWhenSuccessful(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockExchange(ctrl)

	s := gdaxSchedule{}
	s.logger = loggerStub(t).Sugar()
	s.req = syncRequest{every: 24 * time.Hour, orderType: exchanges.Market, autoFund: true, currency: "USD", usd: 50} // setup run every 24 hrs
	s.coins = map[string]orderDetails{"BTC": {symbol: "btcusd", amount: 50}}
	s.markerCoin = "BTC"
	s.sleepFunc = func(d time.Duration) {}
	s.exchange = m

	now := time.Now()
	result := exchanges.Order{OrderID: "1"}

	m.EXPECT().LastPurchaseTime("BTC", "USD", gomock.Any()).Return(nil, nil)
	m.EXPECT().GetFiatAccount("USD").Return(&exchanges.Account{Available: 25}, nil)
	m.EXPECT().GetPendingTransfers("USD").Return([]exchanges.PendingTransfer{}, nil)
	m.EXPECT().Deposit("USD", 25.0).Return(&now, nil)
	m.EXPECT().CreateOrder("btcusd", 50.0, exchanges.Market, gomock.Any()).Return(&result, nil)

	err := s.Sync()

	assert.Nil(t, err)
}

func TestSyncWhenNotSufficientBalanceAndAutoFundIsOff(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockExchange(ctrl)

	s := gdaxSchedule{}
	s.logger = loggerStub(t).Sugar()
	s.req = syncRequest{every: 24 * time.Hour, orderType: exchanges.Market, autoFund: false, currency: "USD", usd: 50} // setup run every 24 hrs
	s.coins = map[string]orderDetails{"BTC": {symbol: "btcusd", amount: 50}}
	s.markerCoin = "BTC"
	s.sleepFunc = func(d time.Duration) {}
	s.exchange = m

	m.EXPECT().LastPurchaseTime("BTC", "USD", gomock.Any()).Return(nil, nil)
	m.EXPECT().GetFiatAccount("USD").Return(&exchanges.Account{Available: 25}, nil)
	m.EXPECT().GetPendingTransfers("USD").Return([]exchanges.PendingTransfer{}, nil)

	err := s.Sync()

	assert.NotNil(t, err)
	assert.Equal(t, "No sufficient amount for trade and autofund is disabled. Deposit money to proceed", err.Error())
}

func TestSyncShouldAskForConfirmationWhenForceIsOn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockExchange(ctrl)

	s := gdaxSchedule{}
	s.logger = loggerStub(t).Sugar()
	s.req = syncRequest{every: 24 * time.Hour, orderType: exchanges.Market, autoFund: false, currency: "USD", usd: 50, force: true} // setup run every 24 hrs
	s.coins = map[string]orderDetails{"BTC": {symbol: "btcusd", amount: 50}}
	s.sleepFunc = func(d time.Duration) {}
	s.exchange = m

	t.Run("when rejected", func(t *testing.T) {
		s.confirmFunc = func(s string) bool {
			return false
		}

		err := s.Sync()

		assert.NotNil(t, err)
		assert.Equal(t, "User rejected the trade", err.Error())
	})

	t.Run("when approved", func(t *testing.T) {
		s.confirmFunc = func(s string) bool {
			return true
		}

		result := exchanges.Order{OrderID: "1"}

		m.EXPECT().GetFiatAccount("USD").Return(&exchanges.Account{Available: 50}, nil)
		m.EXPECT().CreateOrder("btcusd", 50.0, exchanges.Market, gomock.Any()).Return(&result, nil)

		err := s.Sync()

		assert.Nil(t, err)
	})
}

func TestSyncWhenDebugIsOn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockExchange(ctrl)

	s := gdaxSchedule{}
	s.logger = loggerStub(t).Sugar()
	s.req = syncRequest{every: 24 * time.Hour, orderType: exchanges.Market, autoFund: true, currency: "USD", usd: 50} // setup run every 24 hrs
	s.debug = true
	s.markerCoin = "BTC"
	s.coins = map[string]orderDetails{"BTC": {symbol: "btcusd", amount: 50}}
	s.sleepFunc = func(d time.Duration) {}
	s.exchange = m

	m.EXPECT().LastPurchaseTime("BTC", "USD", gomock.Any()).Return(nil, nil)
	m.EXPECT().GetFiatAccount("USD").Return(&exchanges.Account{Available: 25}, nil)
	m.EXPECT().GetPendingTransfers("USD").Return([]exchanges.PendingTransfer{}, nil)

	err := s.Sync()

	assert.Nil(t, err)
}
