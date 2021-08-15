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
	s.req = syncRequest{every: 24 * time.Hour} // setup run every 24 hrs
	s.coins = map[string]orderDetails{"BTC": {}}
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

func TestSyncWhenSuccessful(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockExchange(ctrl)

	s := gdaxSchedule{}
	s.logger = loggerStub(t).Sugar()
	s.req = syncRequest{every: 24 * time.Hour, orderType: exchanges.Market, autoFund: true, currency: "USD", usd: 50} // setup run every 24 hrs
	s.coins = map[string]orderDetails{"BTC": {symbol: "btcusd", amount: 50}}
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
