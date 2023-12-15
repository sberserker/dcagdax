package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/sberserker/dcagdax/exchanges"
)

var (
	exchangeType = kingpin.Flag(
		"exchange",
		"Exchange coinbase, gemini, ftx, ftxus. Default: coinbase",
	).Default("coinbase").String()

	coins = kingpin.Flag(
		"coin",
		"Which coin you want to buy: BTC, LTC, BCH.",
	).Strings()

	every = registerGenerousDuration(kingpin.Flag(
		"every",
		"How often to make purchases, e.g. 1h, 7d, 3w.",
	).Required())

	usd = kingpin.Flag(
		"usd",
		"How much USD to spend on each purchase. If unspecified, the minimum purchase amount allowed will be used.",
	).Float()

	currency = kingpin.Flag(
		"currency",
		"USD, EUR etc",
	).Default("USD").String()

	after = registerDate(kingpin.Flag(
		"after",
		"Start executing trades after this date, e.g. 2017-12-31.",
	))

	until = registerDate(kingpin.Flag(
		"until",
		"Stop executing trades after this date, e.g. 2017-12-31.",
	))

	makeTrades = kingpin.Flag(
		"trade",
		"Actually execute trades.",
	).Bool()

	autoFund = kingpin.Flag(
		"autofund",
		"Automatically initiate ACH deposits.",
	).Bool()

	force = kingpin.Flag(
		"force",
		"Execute trade regardless of the window. Use with caution every run will execute the trade",
	).Bool()

	orderType = kingpin.Flag(
		"type",
		"Order type market, limit. Default: market",
	).Default("market").String()

	orderSpread = kingpin.Flag(
		"spread",
		"Percentage to add above ask price to get limit order executed. Default: 1.0",
	).Default("1.0").Float()

	fee = kingpin.Flag(
		"fee",
		"Fee level to exclude from limit order amount. Default: 0.5",
	).Default("0.5").Float()
)

func main() {
	kingpin.Version("0.1.1")
	kingpin.Parse()

	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	l, _ := config.Build()
	logger := l.Sugar()
	defer logger.Sync()

	exchange, err := initExchange(*exchangeType)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	oType := exchanges.Market
	switch *orderType {
	case "market":
		oType = exchanges.Market
	case "limit":
		oType = exchanges.Limit
	default:
		logger.Warn("unsupported order type " + *orderType)
		os.Exit(1)
	}

	req := syncRequest{
		autoFund:    *autoFund,
		usd:         *usd,
		orderType:   oType,
		orderSpread: *orderSpread,
		fee:         *fee,
		every:       *every,
		until:       *until,
		after:       *after,
		coins:       *coins,
		force:       *force,
		currency:    *currency,
	}

	schedule, err := newGdaxSchedule(
		exchange,
		logger,
		!*makeTrades,
		req,
	)

	if err != nil {
		logger.Warn(err.Error())
		os.Exit(1)
	}

	if err := schedule.Sync(); err != nil {
		logger.Warn(err.Error())
	}
}

func initExchange(exType string) (exchange exchanges.Exchange, err error) {
	switch exType {
	case "coinbase":
		exchange, err = exchanges.NewCoinbaseV3()
	case "gemini":
		exchange, err = exchanges.NewGemini()
	case "ftxus":
		exchange, err = exchanges.NewFtx(true)
	case "ftx":
		exchange, err = exchanges.NewFtx(false)
	default:
		return nil, fmt.Errorf("unsupported exchange %s", exType)
	}
	return exchange, err
}

type generousDuration time.Duration

func registerGenerousDuration(s kingpin.Settings) (target *time.Duration) {
	target = new(time.Duration)
	s.SetValue((*generousDuration)(target))
	return
}

func (d *generousDuration) Set(value string) error {
	durationRegex := regexp.MustCompile(`^(?P<value>\d+)(?P<unit>[hdw])$`)

	if !durationRegex.MatchString(value) {
		return fmt.Errorf("--every misformatted")
	}

	matches := durationRegex.FindStringSubmatch(value)

	hours, _ := strconv.ParseInt(matches[1], 10, 64)
	unit := matches[2]

	switch unit {
	case "d":
		hours *= 24
	case "w":
		hours *= 24 * 7
	}

	duration := time.Duration(hours * int64(time.Hour))

	*d = (generousDuration)(duration)

	return nil
}

func (d *generousDuration) String() string {
	return (*time.Duration)(d).String()
}

type date time.Time

func (d *date) Set(value string) error {
	t, err := time.Parse("2006-01-02", value)

	if err != nil {
		return err
	}

	*d = (date)(t)

	return nil
}

func (d *date) String() string {
	return (*time.Time)(d).String()
}

func registerDate(s kingpin.Settings) (target *time.Time) {
	target = &time.Time{}
	s.SetValue((*date)(target))
	return target
}
