package gemini

import (
	"encoding/json"
	"fmt"

	"github.com/claudiocandio/gemini-api/logger"
)

// Symbols
func (api *Api) Symbols() ([]string, error) {

	url := api.url + symbols_URI

	logger.Debug("func Symbols", fmt.Sprintf("url:%s", url))

	var symbols []string

	body, err := api.request("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &symbols); err != nil {
		return nil, err
	}

	logger.Debug("func Symbols: unmarshal",
		fmt.Sprintf("symbols:%v", symbols),
	)

	return symbols, nil
}

// Symbol Details
func (api *Api) SymbolDetails(symbol string) (Symbol, error) {

	url := api.url + symbol_details_URI + "/" + symbol

	logger.Debug("func SymbolDetails", fmt.Sprintf("url:%s", url))

	var s Symbol

	body, err := api.request("GET", url, nil)

	if err != nil {
		return s, err
	}

	if err := json.Unmarshal(body, &s); err != nil {
		return s, err
	}

	logger.Debug("func SymbolDetails: unmarshal",
		fmt.Sprintf("symbols:%v", s),
	)

	return s, nil
}

// TickerV1
func (api *Api) TickerV1(symbol string) (TickerV1, error) {

	url := api.url + ticker_v1_URI + symbol

	logger.Debug("func TickerV1", fmt.Sprintf("url:%s", url))

	var tickerV1 TickerV1

	body, err := api.request("GET", url, nil)
	if err != nil {
		return tickerV1, err
	}

	if err := json.Unmarshal(body, &tickerV1); err != nil {
		return tickerV1, err
	}

	logger.Debug("func TickerV1: unmarshal",
		fmt.Sprintf("tickerV1:%v", tickerV1),
	)

	return tickerV1, nil
}

// TickerV2
func (api *Api) TickerV2(symbol string) (TickerV2, error) {

	url := api.url + ticker_v2_URI + symbol

	logger.Debug("func TickerV2", fmt.Sprintf("url:%s", url))

	var tickerV2 TickerV2

	body, err := api.request("GET", url, nil)
	if err != nil {
		return tickerV2, err
	}

	if err := json.Unmarshal(body, &tickerV2); err != nil {
		return tickerV2, err
	}

	logger.Debug("func TickerV2: unmarshal",
		fmt.Sprintf("tickerV2:%v", tickerV2),
	)

	return tickerV2, nil
}

// Order Book
func (api *Api) OrderBook(symbol string, args Args) (Book, error) {

	url := api.url + book_URI + symbol

	logger.Debug("func OrderBook", fmt.Sprintf("url:%s", url))

	var book Book

	body, err := api.request("GET", url, args)
	if err != nil {
		return book, err
	}

	if err := json.Unmarshal(body, &book); err != nil {
		return book, err
	}

	logger.Debug("func OrderBook: unmarshal",
		fmt.Sprintf("book:%v", book),
	)

	return book, nil
}

// Trades
func (api *Api) Trades(symbol string, args Args) ([]Trade, error) {

	url := api.url + trades_URI + symbol

	logger.Debug("func Trades",
		fmt.Sprintf("url:%v", url),
		fmt.Sprintf("args:%v", args),
	)

	var trade []Trade

	body, err := api.request("GET", url, args)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &trade); err != nil {
		return nil, err
	}
	// adding TimestampmsT
	for i, r := range trade {
		trade[i].TimestampmsT = msToTime(r.Timestampms)
	}

	logger.Debug("func Trades: unmarshal",
		fmt.Sprintf("trade:%v", trade),
	)

	return trade, nil
}

// Current Auction
func (api *Api) CurrentAuction(symbol string) (CurrentAuction, error) {

	url := api.url + auction_URI + symbol

	logger.Debug("func CurrentAuction", fmt.Sprintf("url:%s", url))

	var currentAuction CurrentAuction

	body, err := api.request("GET", url, nil)
	if err != nil {
		return currentAuction, err
	}

	if err := json.Unmarshal(body, &currentAuction); err != nil {
		return currentAuction, err
	}

	// adding TimestampmsT
	if currentAuction.NextAuction > 0 {
		currentAuction.NextAuctionT = msToTime(currentAuction.NextAuction)
	}
	if currentAuction.NextUpdate > 0 {
		currentAuction.NextUpdateT = msToTime(currentAuction.NextUpdate)
	}

	logger.Debug("func CurrentAuction: unmarshal",
		fmt.Sprintf("currentAuction:%v", currentAuction),
	)

	return currentAuction, nil
}

// Auction History
// Args{"since": 50, "limit": 0, "includeIndicative": true}
func (api *Api) AuctionHistory(symbol string, args Args) ([]Auction, error) {

	url := api.url + auction_URI + symbol + "/history"

	logger.Debug("func AuctionHistory",
		fmt.Sprintf("url:%v", url),
		fmt.Sprintf("args:%v", args),
	)

	var auction []Auction

	body, err := api.request("GET", url, args)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &auction); err != nil {
		return nil, err
	}

	// adding TimestampmsT
	for i, au := range auction {
		auction[i].TimestampmsT = msToTime(au.Timestampms)
	}

	logger.Debug("func AuctionHistory: unmarshal",
		fmt.Sprintf("auction:%v", auction),
	)

	return auction, nil
}
