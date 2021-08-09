package gemini

import "time"

const (
	base_URL    = "https://api.gemini.com"
	sandbox_URL = "https://api.sandbox.gemini.com"
	//ws_base_URL    = "wss://api.gemini.com"
	//ws_sandbox_URL = "wss://api.sandbox.gemini.com"

	// public
	symbols_URI        = "/v1/symbols"
	symbol_details_URI = "/v1/symbols/details"
	ticker_v1_URI      = "/v1/pubticker/"
	ticker_v2_URI      = "/v2/ticker/"
	book_URI           = "/v1/book/"
	trades_URI         = "/v1/trades/"
	auction_URI        = "/v1/auction/"

	// authenticated
	past_trades_URI    = "/v1/mytrades"
	trade_volume_URI   = "/v1/tradevolume"
	active_orders_URI  = "/v1/orders"
	order_status_URI   = "/v1/order/status"
	new_order_URI      = "/v1/order/new"
	cancel_order_URI   = "/v1/order/cancel"
	cancel_all_URI     = "/v1/order/cancel/all"
	cancel_session_URI = "/v1/order/cancel/session"
	heartbeat_URI      = "/v1/heartbeat"
	account_URI        = "/v1/account"
	transfers_URI      = "/v1/transfers"

	// fund mgmt
	balances_URI            = "/v1/balances"
	new_deposit_address_URI = "/v1/deposit/"
	deposit_addresses_URI   = "/v1/addresses/"
	withdraw_funds_URI      = "/v1/withdraw/"

	// websockets
	//order_events_URI = "/v1/order/events"
	//market_data_URI  = "/v1/marketdata/"
)

type GenericResponse struct {
	Result string `json:"result"`
}

type Order struct {
	OrderId           string   `json:"order_id"`
	ClientOrderId     string   `json:"client_order_id"`
	Symbol            string   `json:"symbol"`
	Exchange          string   `json:"exchange"`
	Price             float64  `json:"price,string"`
	AvgExecutionPrice float64  `json:"avg_execution_price,string"`
	Side              string   `json:"side"`
	Type              string   `json:"type"`
	Options           []string `json:"options"`
	//	Timestamp         string    `json:"timestamp"`
	Timestampms     int64     `json:"timestampms"`
	TimestampmsT    time.Time `json:"timestampmst,omitempty"`
	IsLive          bool      `json:"is_live"`
	IsCancelled     bool      `json:"is_cancelled"`
	Reason          string    `json:"reason"`
	WasForced       bool      `json:"was_forced"`
	ExecutedAmount  float64   `json:"executed_amount,string"`
	RemainingAmount float64   `json:"remaining_amount,string"`
	OriginalAmount  float64   `json:"original_amount,string"`
	IsHidden        bool      `json:"is_hidden"`
}

type Trade struct {
	Timestamp    int64     `json:"timestamp"`
	Timestampms  int64     `json:"timestampms"`
	TimestampmsT time.Time `json:"timestampmst,omitempty"`
	TradeId      int64     `json:"tid"`
	Price        float64   `json:"price,string"`
	Amount       float64   `json:"amount,string"`
	Exchange     string    `json:"exchange"`
	Type         string    `json:"type"`
	Broken       bool      `json:"broken,omitempty"`
}

type PastTrade struct {
	Price           float64   `json:"price,string"`
	Amount          float64   `json:"amount,string"`
	Timestamp       int64     `json:"timestamp"`
	Timestampms     int64     `json:"timestampms"`
	TimestampmsT    time.Time `json:"timestampmst,omitempty"`
	Type            string    `json:"type"`
	Aggressor       bool      `json:"aggressor"`
	FeeCurrency     string    `json:"fee_currency"`
	FeeAmount       float64   `json:"fee_amount,string"`
	TradeId         int64     `json:"tid"`
	OrderId         string    `json:"order_id"`
	Client_Order_Id string    `json:"client_order_id,omitempty"`
	Exchange        string    `json:"exchange"`
	IsAuctionFill   bool      `json:"is_auction_fill"`
	Break           string    `json:"break,omitempty"`
}

type TickerV1 struct {
	Bid    float64        `json:"bid,string"`
	Ask    float64        `json:"ask,string"`
	Last   float64        `json:"last,string"`
	Volume TickerV1Volume `json:"volume"`
}
type TickerV1Volume struct {
	BTC       float64 `json:",string"`
	ETH       float64 `json:",string"`
	USD       float64 `json:",string"`
	Timestamp int64   `json:"timestamp"`
}

type TickerV2 struct {
	Symbol  string   `json:"symbol"`
	Open    float64  `json:"open,string"`
	High    float64  `json:"high,string"`
	Low     float64  `json:"low,string"`
	Close   float64  `json:"close,string"`
	Changes []string `json:"changes"`
	Bid     float64  `json:"bid,string"`
	Ask     float64  `json:"ask,string"`
}

type TradeVolume struct {
	Symbol            string  `json:"symbol"`
	BaseCurrency      string  `json:"base_currency"`
	NotionalCurrency  string  `json:"notional_currency"`
	DataDate          string  `json:"data_date"`
	TotalVolumeBase   float64 `json:"total_volume_base"`
	MakeBuySellRatio  float64 `json:"maker_buy_sell_ratio"`
	BuyMakerBase      float64 `json:"buy_maker_base"`
	BuyMakerNotional  float64 `json:"buy_maker_notional"`
	BuyMakerCount     float64 `json:"buy_maker_count"`
	SellMakerBase     float64 `json:"sell_maker_base"`
	SellMakerNotional float64 `json:"sell_maker_notional"`
	SellMakerCount    float64 `json:"sell_maker_count"`
	BuyTakerBase      float64 `json:"buy_taker_base"`
	BuyTakerNotional  float64 `json:"buy_taker_notional"`
	BuyTakerCount     float64 `json:"buy_taker_count"`
	SellTakerBase     float64 `json:"sell_taker_base"`
	SellTakerNotional float64 `json:"sell_taker_notional"`
	SellTakerCount    float64 `json:"sell_taker_count"`
}

type CurrentAuction struct {
	ClosedUntil                  int64     `json:"closed_until_ms,omitempty"`
	LastAuctionEid               int64     `json:"last_auction_eid,omitempty"`
	LastAuctionPrice             float64   `json:"last_auction_price,string,omitempty"`
	LastAuctionQuantity          float64   `json:"last_auction_quantity,string,omitempty"`
	LastHighestBidPrice          float64   `json:"last_highest_bid_price,string,omitempty"`
	LastLowestAskPrice           float64   `json:"last_lowest_ask_price,string,omitempty"`
	LastCollarPrice              float64   `json:"last_collar_price,string,omitempty"`
	MostRecentIndicativePrice    float64   `json:"most_recent_indicative_price,string,omitempty"`
	MostRecentIndicativeQuantity float64   `json:"most_recent_indicative_quantity,string,omitempty"`
	MostRecentHighestBidPrice    float64   `json:"most_recent_highest_bid_price,string,omitempty"`
	MostRecentLowestAskPrice     float64   `json:"most_recent_lowest_ask_price,string,omitempty"`
	MostRecentCollarPrice        float64   `json:"most_recent_collar_price,string,omitempty"`
	NextUpdate                   int64     `json:"next_update_ms,omitempty"`
	NextUpdateT                  time.Time `json:"next_update_mst,omitempty"`
	NextAuction                  int64     `json:"next_auction_ms"`
	NextAuctionT                 time.Time `json:"next_auction_mst,omitempty"`
}

type Auction struct {
	Timestampms     int64     `json:"timestampms"`
	TimestampmsT    time.Time `json:"timestampmst,omitempty"`
	AuctionId       int64     `json:"auction_id"`
	Eid             int64     `json:"eid"`
	EventType       string    `json:"event_type"`
	AuctionResult   string    `json:"auction_result"`
	AuctionPrice    float64   `json:"auction_price,string"`
	AuctionQuantity float64   `json:"auction_quantity,string"`
	HighestBidPrice float64   `json:"highest_bid_price,string"`
	LowestAskPrice  float64   `json:"lowest_ask_price,string"`
	CollarPrice     float64   `json:"collar_price,string"`
}

type CancelResult struct {
	Result  string              `json:"result"`
	Details CancelResultDetails `json:"details"`
}

type CancelResultDetails struct {
	CancelledOrders []float64 `json:"cancelledOrders"`
	CancelRejects   []float64 `json:"cancelRejects"`
}

type FundBalance struct {
	Currency               string  `json:"currency"`
	Amount                 float64 `json:"amount,string"`
	Available              float64 `json:"available,string"`
	AvailableForWithdrawal float64 `json:"availableForWithdrawal,string"`
	Type                   string  `json:"type"`
}

type AccountDetail struct {
	Account             Account `json:"account"`
	Users               []Users `json:"users"`
	Memo_Reference_Code string  `json:"memo_reference_code"`
}

type Account struct {
	AccountName string    `json:"accountname"`
	ShortName   string    `json:"shortname"`
	Type        string    `json:"type"`
	Created     int64     `json:"created,string"`
	CreatedT    time.Time `json:"createdt,omitempty"`
}

type Users struct {
	Name        string    `json:"name"`
	LastSignIn  time.Time `json:"lastsignin"`
	Status      string    `json:"status"`
	CountryCode string    `json:"countrycode"`
	IsVerified  bool      `json:"isverified"`
}

type NewDepositAddress struct {
	Request string `json:"request"`
	Address string `json:"address"`
	Label   string `json:"label"`
}

type DepositAddresses struct {
	Address      string    `json:"address"`
	Timestamp    int64     `json:"timestamp"`
	TimestampmsT time.Time `json:"timestampmst,omitempty"`
	Label        string    `json:"label,omitempty"`
}

type WithdrawFundsResult struct {
	Address      string `json:"address"`
	Amount       string `json:"amount"`
	TxHash       string `json:"txHash"`
	WithdrawalID string `json:"withdrawalID,omitempty"`
	Message      string `json:"message,omitempty"`
}

type Book struct {
	Bids BookEntries `json:"bids"`
	Asks BookEntries `json:"asks"`
}

type BookEntries []BookEntry

type BookEntry struct {
	Price  float64 `json:"price,string"`
	Amount float64 `json:"amount,string"`
}

type Transfer struct {
	Type         string    `json:"type"`
	Status       string    `json:"status"`
	Timestampms  int64     `json:"timestampms"`
	TimestampmsT time.Time `json:"timestampmst,omitempty"`
	Eid          int64     `json:"eid"`
	AdvancedEid  int64     `json:"advanceEid"`
	Currency     string    `json:"currency"`
	Amount       float64   `json:"amount,string"`
	Method       string    `json:"method,omitempty"`
	TxHash       string    `json:"txHash,omitempty"`
	OutputIdx    float64   `json:"outputIdx,omitempty"`
	Destination  string    `json:"destination,omitempty"`
	Purpose      string    `json:"purpose,omitempty"`
}

type Symbol struct {
	Type           string  `json:"symbol"`
	BaseCurrency   string  `json:"base_currency"`
	QuoteCurrency  string  `json:"quote_currency"`
	TickSize       float64 `json:"tick_size"`
	QuoteIncrement float64 `json:"quote_increment"`
	MinOrderSize   float64 `json:"min_order_size,string"`
	Status         string  `json:"status"`
}
