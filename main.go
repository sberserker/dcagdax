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
	coins = kingpin.Flag(
		"coin",
		"Which coin you want to buy: BTC, LTC, BCH.",
	).Strings()

	exchangeType = kingpin.Flag(
		"exchange",
		"Exchange coinbase, gemini, ftx. Default: coinbase",
	).Default("coinbase").String()

	every = registerGenerousDuration(kingpin.Flag(
		"every",
		"How often to make purchases, e.g. 1h, 7d, 3w.",
	).Required())

	usd = kingpin.Flag(
		"usd",
		"How much USD to spend on each purchase. If unspecified, the minimum purchase amount allowed will be used.",
	).Float()

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
)

func main() {
	kingpin.Version("0.1.0")
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

	schedule, err := newGdaxSchedule(
		exchange,
		logger,
		!*makeTrades,
		*autoFund,
		*usd,
		*every,
		*until,
		*after,
		*coins,
		*force,
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
		exchange, err = exchanges.NewCoinbase()
	case "gemini":
		exchange, err = exchanges.NewGemini()
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
