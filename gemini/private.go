package gemini

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/claudiocandio/gemini-api/logger"
)

// Past Trades
// Args{"limit_trades": 50, "timestamp": "2021-12-01T15:04:01"}
// limit_trades": 0 -> retrieves all trades
func (api *Api) PastTrades(symbol string, args Args) ([]PastTrade, error) {
	const max_limit_tradesAPI int = 500

	var maxTrades, limit_trades int = 0, max_limit_tradesAPI
	if arg, ok := args["limit_trades"]; ok {
		limit_trades = int(arg.(int))
		maxTrades = limit_trades
		if limit_trades > max_limit_tradesAPI {
			limit_trades = max_limit_tradesAPI
		}
	}

	var timestamp int64 = 0 // default
	if arg, ok := args["timestamp"]; ok {
		// timestamp ms
		timestamp = arg.(time.Time).UnixNano() / 1e6
	}

	logger.Debug("func PastTrades",
		fmt.Sprintf("args:%v", args),
		fmt.Sprintf("maxTrades:%d", maxTrades),
		fmt.Sprintf("limit_trades:%d", limit_trades),
		fmt.Sprintf("timestamp:%v", timestamp),
	)

	url := api.url + past_trades_URI

	var ptrade []PastTrade
	var pastTrade []PastTrade

	readTrades := true
	for readTrades {
		params := map[string]interface{}{
			"request":      past_trades_URI,
			"nonce":        nonce(),
			"symbol":       symbol,
			"limit_trades": limit_trades,
			"timestamp":    timestamp,
		}

		logger.Debug("func PastTrades",
			fmt.Sprintf("url:%v", url),
			fmt.Sprintf("params:%v", params),
		)

		body, err := api.request("POST", url, params)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(body, &ptrade); err != nil {
			return nil, err
		}

		pastTrade = append(pastTrade, ptrade...)

		if len(ptrade) > 0 &&
			(len(pastTrade) < maxTrades || maxTrades == 0) &&
			len(ptrade) == limit_trades {

			timestamp = ptrade[0].Timestamp + 1
			logger.Debug("func PastTrades: next limit_trades page")
		} else {
			readTrades = false
			logger.Debug("func PastTrades: end limit_trades page")
		}
	}

	// adding TimestampmsT
	for i, t := range pastTrade {
		pastTrade[i].TimestampmsT = msToTime(t.Timestampms)
	}

	logger.Debug("func PastTrades: unmarshal",
		fmt.Sprintf("pastTrade:%v", pastTrade),
	)

	return pastTrade, nil
}

// Trade Volume
func (api *Api) TradeVolume() ([][]TradeVolume, error) {

	url := api.url + trade_volume_URI
	params := map[string]interface{}{
		"request": trade_volume_URI,
		"nonce":   nonce(),
	}

	logger.Debug("func TradeVolume",
		fmt.Sprintf("url:%v", url),
		fmt.Sprintf("params:%v", params),
	)

	var tradeVolume [][]TradeVolume

	body, err := api.request("POST", url, params)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &tradeVolume); err != nil {
		return nil, err
	}

	logger.Debug("func TradeVolume: unmarshal",
		fmt.Sprintf("tradeVolume:%v", tradeVolume),
	)

	return tradeVolume, nil
}

// Active Orders
func (api *Api) ActiveOrders() ([]Order, error) {

	url := api.url + active_orders_URI
	params := map[string]interface{}{
		"request": active_orders_URI,
		"nonce":   nonce(),
	}

	logger.Debug("func ActiveOrders",
		fmt.Sprintf("url:%v", url),
		fmt.Sprintf("params:%v", params),
	)

	var order []Order

	body, err := api.request("POST", url, params)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &order); err != nil {
		return nil, err
	}

	logger.Debug("func ActiveOrders: unmarshal",
		fmt.Sprintf("orders:%v", order),
	)

	return order, nil
}

// Order Status
func (api *Api) OrderStatus(orderId string) (Order, error) {

	url := api.url + order_status_URI
	params := map[string]interface{}{
		"request":  order_status_URI,
		"nonce":    nonce(),
		"order_id": orderId,
	}

	logger.Debug("func OrderStatus",
		fmt.Sprintf("url:%v", url),
		fmt.Sprintf("params:%v", params),
	)

	var order Order

	body, err := api.request("POST", url, params)
	if err != nil {
		return order, err
	}

	if err := json.Unmarshal(body, &order); err != nil {
		return order, err
	}

	logger.Debug("func OrderStatus: unmarshal",
		fmt.Sprintf("order:%v", order),
	)

	return order, nil
}

// New Order
func (api *Api) NewOrder(symbol, clientOrderId string, amount, price float64, side string, options []string) (Order, error) {

	url := api.url + new_order_URI
	params := map[string]interface{}{
		"request":         new_order_URI,
		"nonce":           nonce(),
		"client_order_id": clientOrderId,
		"symbol":          symbol,
		"amount":          strconv.FormatFloat(amount, 'f', -1, 64),
		"price":           strconv.FormatFloat(price, 'f', -1, 64),
		"side":            side,
		"type":            "exchange limit",
	}

	if options != nil {
		params["options"] = options
	}

	logger.Debug("func NewOrder",
		fmt.Sprintf("url:%v", url),
		fmt.Sprintf("params:%v", params),
	)

	var order Order

	body, err := api.request("POST", url, params)
	if err != nil {
		return order, err
	}

	if err := json.Unmarshal(body, &order); err != nil {
		return order, err
	}

	order.TimestampmsT = msToTime(order.Timestampms)

	logger.Debug("func NewOrder: unmarshal",
		fmt.Sprintf("order:%v", order),
	)

	return order, nil
}

// Cancel Order
func (api *Api) CancelOrder(orderId string) (Order, error) {

	url := api.url + cancel_order_URI
	params := map[string]interface{}{
		"request":  cancel_order_URI,
		"nonce":    nonce(),
		"order_id": orderId,
	}

	logger.Debug("func CancelOrder",
		fmt.Sprintf("url:%v", url),
		fmt.Sprintf("params:%v", params),
	)

	var order Order

	body, err := api.request("POST", url, params)
	if err != nil {
		return order, err
	}

	if err := json.Unmarshal(body, &order); err != nil {
		return order, err
	}

	logger.Debug("func CancelOrder: unmarshal",
		fmt.Sprintf("order:%v", order),
	)

	return order, nil
}

// Cancel All
// This will cancel all outstanding orders created by all sessions owned
// by this account, including interactive orders placed through the UI.
// Note that this cancels orders that were not placed using this API key.
func (api *Api) CancelAll() (CancelResult, error) {

	url := api.url + cancel_all_URI
	params := map[string]interface{}{
		"request": cancel_all_URI,
		"nonce":   nonce(),
	}

	logger.Debug("func CancelAll",
		fmt.Sprintf("url:%v", url),
		fmt.Sprintf("params:%v", params),
	)

	var cancelResult CancelResult

	body, err := api.request("POST", url, params)
	if err != nil {
		return cancelResult, err
	}

	if err := json.Unmarshal(body, &cancelResult); err != nil {
		return cancelResult, err
	}

	logger.Debug("func CancelAll: unmarshal",
		fmt.Sprintf("cancelResult:%v", cancelResult),
	)

	return cancelResult, nil
}

// This will cancel all orders opened by this session.
// This will have the same effect as heartbeat expiration if "Require Heartbeat" is selected for the session.
func (api *Api) CancelSession() (CancelResult, error) {

	url := api.url + cancel_session_URI
	params := map[string]interface{}{
		"request": cancel_session_URI,
		"nonce":   nonce(),
	}

	logger.Debug("func CancelSession",
		fmt.Sprintf("url:%v", url),
		fmt.Sprintf("params:%v", params),
	)

	var cancelResult CancelResult

	body, err := api.request("POST", url, params)
	if err != nil {
		return cancelResult, err
	}

	if err := json.Unmarshal(body, &cancelResult); err != nil {
		return cancelResult, err
	}

	logger.Debug("func CancelSession: unmarshal",
		fmt.Sprintf("cancelResult:%v", cancelResult),
	)

	return cancelResult, nil
}

// Heartbeat
// This will prevent a session from timing out and canceling orders if the
// require heartbeat flag has been set. Note that this is only required if
// no other private API requests have been made. The arrival of any message
// resets the heartbeat timer.
func (api *Api) Heartbeat() (GenericResponse, error) {

	url := api.url + heartbeat_URI
	params := map[string]interface{}{
		"request": heartbeat_URI,
		"nonce":   nonce(),
	}

	logger.Debug("func Heartbeat",
		fmt.Sprintf("url:%v", url),
		fmt.Sprintf("params:%v", params),
	)

	var genericResponse GenericResponse

	body, err := api.request("POST", url, params)
	if err != nil {
		return genericResponse, err
	}

	if err := json.Unmarshal(body, &genericResponse); err != nil {
		return genericResponse, err
	}

	logger.Debug("func Heartbeat: unmarshal",
		fmt.Sprintf("genericResponse:%v", genericResponse),
	)

	return genericResponse, nil
}

// Balances
func (api *Api) Balances() ([]FundBalance, error) {

	url := api.url + balances_URI
	params := map[string]interface{}{
		"request": balances_URI,
		"nonce":   nonce(),
	}

	logger.Debug("func Balances",
		fmt.Sprintf("url:%v", url),
		fmt.Sprintf("params:%v", params),
	)

	var fundBalance []FundBalance

	body, err := api.request("POST", url, params)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &fundBalance); err != nil {
		return nil, err
	}

	logger.Debug("func Balances: unmarshal",
		fmt.Sprintf("fundBalance:%v", fundBalance),
	)

	return fundBalance, nil
}

// Account
func (api *Api) AccountDetail() (AccountDetail, error) {

	url := api.url + account_URI
	params := map[string]interface{}{
		"request": account_URI,
		"nonce":   nonce(),
	}

	logger.Debug("func AccountDetail",
		fmt.Sprintf("url:%v", url),
		fmt.Sprintf("params:%v", params),
	)

	var accountDetail AccountDetail

	body, err := api.request("POST", url, params)
	if err != nil {
		return accountDetail, err
	}

	if err := json.Unmarshal(body, &accountDetail); err != nil {
		return accountDetail, err
	}

	accountDetail.Account.CreatedT = msToTime(accountDetail.Account.Created)

	logger.Debug("func AccountDetail: unmarshal",
		fmt.Sprintf("accountDetail:%v", accountDetail),
	)

	return accountDetail, nil
}

// New Deposit Address
// currency can be bitcoin, ethereum, bitcoincash, litecoin, zcash, or filecoin
func (api *Api) NewDepositAddress(currency, label string) (NewDepositAddress, error) {

	path := new_deposit_address_URI + currency + "/newAddress"
	url := api.url + path
	params := map[string]interface{}{
		"request": path,
		"nonce":   nonce(),
		"label":   label,
	}

	logger.Debug("func NewDepositAddress",
		fmt.Sprintf("url:%v", url),
		fmt.Sprintf("params:%v", params),
	)

	var newDepositAddress NewDepositAddress

	body, err := api.request("POST", url, params)
	if err != nil {
		return newDepositAddress, err
	}

	if err := json.Unmarshal(body, &newDepositAddress); err != nil {
		return newDepositAddress, err
	}

	logger.Debug("func NewDepositAddress: unmarshal",
		fmt.Sprintf("newDepositAddress:%v", newDepositAddress),
	)

	return newDepositAddress, nil
}

// Get Deposit Addresseses
// currency can be bitcoin, ethereum, bitcoincash, litecoin, zcash, filecoin
func (api *Api) DepositAddresses(currency string) ([]DepositAddresses, error) {

	path := deposit_addresses_URI + currency
	url := api.url + path
	params := map[string]interface{}{
		"request": path,
		"nonce":   nonce(),
	}

	logger.Debug("func DepositAddresses",
		fmt.Sprintf("url:%v", url),
		fmt.Sprintf("params:%v", params),
	)

	var depositAddresses []DepositAddresses

	body, err := api.request("POST", url, params)
	if err != nil {
		return depositAddresses, err
	}

	if err := json.Unmarshal(body, &depositAddresses); err != nil {
		return depositAddresses, err
	}

	// adding TimestampmsT
	for i, address := range depositAddresses {
		depositAddresses[i].TimestampmsT = msToTime(address.Timestamp)
	}

	logger.Debug("func DepositAddresses: unmarshal",
		fmt.Sprintf("depositAddresses:%v", depositAddresses),
	)

	return depositAddresses, nil
}

// Withdraw Crypto Funds
// currency can be btc or eth
func (api *Api) WithdrawFunds(currency, address string, amount float64) (WithdrawFundsResult, error) {

	path := withdraw_funds_URI + currency
	url := api.url + path
	amountstr := fmt.Sprintf("%f", amount)
	params := map[string]interface{}{
		"request": path,
		"nonce":   nonce(),
		"address": address,
		"amount":  amountstr,
	}

	logger.Debug("func WithdrawFunds",
		fmt.Sprintf("url:%v", url),
		fmt.Sprintf("params:%v", params),
	)

	var withdrawFundsResult WithdrawFundsResult

	body, err := api.request("POST", url, params)
	if err != nil {
		return withdrawFundsResult, err
	}

	if err := json.Unmarshal(body, &withdrawFundsResult); err != nil {
		return withdrawFundsResult, err
	}

	logger.Debug("func WithdrawFunds: unmarshal",
		fmt.Sprintf("withdrawFundsResult:%v", withdrawFundsResult),
	)

	return withdrawFundsResult, nil
}

// Args{"timestamp": "2021-12-01T15:04:01", "limit_transfers": 20,"show_completed_deposit_advances": false}
func (api *Api) Transfers(args Args) ([]Transfer, error) {

	url := api.url + transfers_URI
	args["request"] = transfers_URI
	args["nonce"] = nonce()

	logger.Debug("func Transfers",
		fmt.Sprintf("url:%v", url),
		fmt.Sprintf("args:%v", args),
	)

	var transfer []Transfer

	body, err := api.request("POST", url, args)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &transfer); err != nil {
		return nil, err
	}

	// adding TimestampmsT
	for i, tr := range transfer {
		transfer[i].TimestampmsT = msToTime(tr.Timestampms)
	}

	logger.Debug("func Transfers: unmarshal",
		fmt.Sprintf("transfer:%v", transfer),
	)

	return transfer, nil
}
