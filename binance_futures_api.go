package go_binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	// ====== URL & END POINTS ======
	mainNetBaseURL     				= "https://fapi.binance.com"
	testNetBaseURL     				= "https://testnet.binancefuture.com"

	ticker24HrEndPoint 				= "/fapi/v1/ticker/24hr"
	listenKeyEndPoint     			= "/fapi/v1/listenKey"
	orderEndPoint	   				= "/fapi/v1/order"
	exchangeInformationEndPoint 	= "/fapi/v1/exchangeInfo"
	orderBookEndpoint 				= "/fapi/v1/depth"
	klinesEndpoint					= "/fapi/v1/klines"
	futuresAccountBalanceEndpoint 	= "/fapi/v2/balance"
	accountInformationEndpoint 		= "/fapi/v2/account"
	allOpenOrdersEndPoint 			= "/fapi/v1/allOpenOrders"

	// ====== Parameter Types ======
	SideBuy 			= "BUY"
	SideSell 			= "SELL"

	OrderTypeLimit 		= "LIMIT"
	OrderTypeMarket 	= "MARKET"
	GoodTillCancel 		= "GTC"
	GoodTillCrossing 	= "GTX"
	OrderTypeStopMarket = "STOP_MARKET"
	OrderTypeStop		= "STOP"

	DefaultOrderBookLimit = 500
	DefaultKlineLimit     = 500

	KlineInterval5Min 	= "1m"
	KlineInterval15Min 	= "15m"
	KlineInterval30Min 	= "30m"
	KlineInterval1Hour 	= "1h"
	KlineInterval4Hour 	= "4h"
	KlineInterval8Hour 	= "8m"
)



type BinanceFuturesApi struct {
	Client *http.Client
	BaseUrl string
	PublicKey string
	SecretKey string
	Logger *logrus.Logger
}

func (bfa *BinanceFuturesApi) SetApiKeys(public, secret string) {
	bfa.PublicKey = public
	bfa.SecretKey = secret
}

func (bfa *BinanceFuturesApi) NewNetClient() {
	// Todo: make sure timeout and handshake won't cause any problems.
	var netTransport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 2 * time.Second,
		}).DialContext,
		DisableKeepAlives: false,
		TLSHandshakeTimeout: 2 * time.Second,
		ForceAttemptHTTP2:     true,
	}
	bfa.Client = &http.Client{
		Timeout:   0,
		Transport: netTransport,
	}
}

func (bfa *BinanceFuturesApi) NewNetClientHTTP2() {
	var netTransport = &http2.Transport{}
	bfa.Client = &http.Client{
		Timeout:   0,
		Transport: netTransport,
	}
}

func(bfa *BinanceFuturesApi) UseMainNet() {
	bfa.BaseUrl = mainNetBaseURL
}
func(bfa *BinanceFuturesApi) UseTestNet() {
	bfa.BaseUrl = testNetBaseURL
}

// Adds timestamp and creates a signature according to binance rules
func (bfa BinanceFuturesApi) signParameters(parameters *url.Values) string {
	tonce := strconv.FormatInt(time.Now().UnixNano(), 10)[0:13]
	parameters.Add("timestamp", tonce)
	signature, _ := bfa.getSha256Signature(parameters.Encode())
	return signature
}

// Signs given parameter values
func (bfa BinanceFuturesApi) getSha256Signature(parameters string) (string, error) {
	mac := hmac.New(sha256.New, []byte(bfa.SecretKey))
	_, err := mac.Write([]byte(parameters))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func (bfa BinanceFuturesApi) parseResponseBody(body io.ReadCloser) ([]byte, error) {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		bfa.Logger.Println("Something went wrong during reading request body, ", err)
	}
	return data, nil
}

func (bfa BinanceFuturesApi) doPublicRequest(httpVerb, endPoint string, parameters url.Values) ([]byte, error) {
	fullURL := bfa.BaseUrl + endPoint
	if parameters != nil {
		fullURL += "?" + parameters.Encode()
	}
	request, _ := http.NewRequest(httpVerb, fullURL, nil)
	response, err := bfa.Client.Do(request)
	if err != nil {
		// Failure to speak HTTP - Connectivity error
		// Non 2XX status does not produce errors
		bfa.Logger.Error("Connectivity error while Client.Do ", err)
		return nil, err
	}
	// Log rate limit for debug purposes
	// Even if request results in non 2xx status it provides
	// rate limit information
	bfa.Logger.Println(endPoint + ", rate limit used: ",
		response.Header.Get("X-Mbx-Used-Weight-1m"))

	data, err := bfa.parseResponseBody(response.Body)
	if err != nil {
		return nil, err
	}

	// Non 2xx status does not return error
	if response.StatusCode != 200 {
		bem := new(BinanceErrorMessage)
		err = json.Unmarshal(data, &bem)
		if err != nil {
			bem = nil
		}
		err = &RequestError{
			StatusCode: response.StatusCode,
			UrlUsed: 	fullURL,
			Message:   	*bem,
		}
		bfa.Logger.Error(endPoint, err.Error())
		return nil, err
	}
	return data, nil
}

func (bfa BinanceFuturesApi) doSignedRequest(httpVerb, endPoint string, parameters url.Values) ([]byte, error) {
	signature := bfa.signParameters(&parameters)
	headers := make(http.Header)
	headers.Add("X-MBX-APIKEY", bfa.PublicKey)
	fullURL := bfa.BaseUrl + endPoint + "?" + parameters.Encode() +"&signature=" + signature
	request, _ := http.NewRequest(httpVerb, fullURL, nil)
	request.Header = headers
	response, err := bfa.Client.Do(request)
	if err != nil {
		// Failure to speak HTTP - Connectivity error
		// Non 2XX status does not produce errors
		bfa.Logger.Error("Connectivity error while Client.Do ", err)
		return nil, err
	}
	// Log rate limit for debug purposes
	// Even if request results in non 2xx status it provides
	// rate limit information
	bfa.Logger.Println(endPoint + ", rate limit used: ",
		response.Header.Get("X-Mbx-Used-Weight-1m"))

	data, err := bfa.parseResponseBody(response.Body)
	if err != nil {
		return nil, err
	}

	// Non 2xx status does not return error
	if response.StatusCode != 200 {
		bem := new(BinanceErrorMessage)
		err = json.Unmarshal(data, &bem)
		if err != nil {
			bem = nil
		}
		err = &RequestError{
			StatusCode: response.StatusCode,
			UrlUsed: 	fullURL,
			Message:   	*bem,
		}
		bfa.Logger.Error(endPoint, err.Error())
		return nil, err
	}
	return data, nil
}

// ======================= PUBLIC API CALLS ================================

// 	Contains weighted average price (vwap)
func (bfa BinanceFuturesApi) Get24HourTickerPriceChangeStatistics(symbol string) ([]byte, error) {
	parameters := url.Values{}
	parameters.Add("symbol", symbol)
	return bfa.doPublicRequest("GET", ticker24HrEndPoint, parameters)
}

func (bfa BinanceFuturesApi) GetOrderBook(symbol string, limit int) ([]byte, error)  {
	parameters := url.Values{}
	parameters.Add("symbol", symbol)
	parameters.Add("limit", fmt.Sprintf("%d", limit))
	return bfa.doPublicRequest("GET", orderBookEndpoint, parameters)
}

func(bfa BinanceFuturesApi) GetExchangeInformation() ([]byte, error) {
	return bfa.doPublicRequest("GET", exchangeInformationEndPoint, nil)
}
// limit can be just 1 if only current candle is needed.
func(bfa BinanceFuturesApi) GetKlines(symbol, interval string, limit int) ([]byte, error) {
	parameters := url.Values{}
	parameters.Add("symbol", symbol)
	parameters.Add("interval", interval)
	parameters.Add("limit", fmt.Sprintf("%d", limit))
	return bfa.doPublicRequest("GET", klinesEndpoint, parameters)
}

// ======================= SIGNED API CALLS ================================

func (bfa BinanceFuturesApi) GetUserStreamKey() ([]byte, error) {
	return bfa.doSignedRequest("POST", listenKeyEndPoint, url.Values{})
}

// Keepalive a user data stream to prevent a time out. User data streams will close after 60 minutes.
// It's recommended to send a ping about every 60 minutes.
// returns no information, it is completely fine to ignore the byte slice
func (bfa BinanceFuturesApi) UpdateKeepAliveUserStream() ([]byte, error) {
	return  bfa.doSignedRequest("PUT", listenKeyEndPoint, url.Values{})
}

// returns no information, it is completely fine to ignore the byte slice
func (bfa BinanceFuturesApi) DeleteUserStream() ([]byte, error) {
	return  bfa.doSignedRequest("DELETE", listenKeyEndPoint, url.Values{})
}

func (bfa BinanceFuturesApi) PlaceLimitOrder(symbol, side string, price, qty float64, reduceOnly bool) ([]byte, error) {
	parameters := url.Values{}
	parameters.Add("symbol", symbol)
	parameters.Add("side", side)
	parameters.Add("type", OrderTypeLimit)
	parameters.Add("timeInForce", GoodTillCancel)
	parameters.Add("reduceOnly",  strconv.FormatBool(reduceOnly))
	parameters.Add("quantity", strconv.FormatFloat(qty, 'f', -1, 64))
	parameters.Add("price", strconv.FormatFloat(price, 'f', -1, 64))
	return bfa.doSignedRequest("POST", orderEndPoint, parameters)
}

func (bfa BinanceFuturesApi) PlacePostOnlyLimitOrder(symbol, side string, price, qty float64, reduceOnly bool) ([]byte, error) {
	parameters := url.Values{}
	parameters.Add("symbol", symbol)
	parameters.Add("side", side)
	parameters.Add("type", OrderTypeLimit)
	parameters.Add("timeInForce", GoodTillCrossing)
	parameters.Add("reduceOnly",  strconv.FormatBool(reduceOnly))
	parameters.Add("quantity", strconv.FormatFloat(qty, 'f', -1, 64))
	parameters.Add("price", strconv.FormatFloat(price, 'f', -1, 64))
	return bfa.doSignedRequest("POST", orderEndPoint, parameters)
}

func (bfa BinanceFuturesApi) PlaceMarketOrder(symbol, side string, qty float64, reduceOnly bool) ([]byte, error) {
	parameters := url.Values{}
	parameters.Add("symbol", symbol)
	parameters.Add("side", side)
	parameters.Add("type", OrderTypeMarket)
	parameters.Add("reduceOnly",  strconv.FormatBool(reduceOnly))
	parameters.Add("quantity", strconv.FormatFloat(qty, 'f', -1, 64))
	return bfa.doSignedRequest("POST", orderEndPoint, parameters)
}

// Generally used for trailing profit orders.
// Binance only allows one stop market order to be active
// after initial order any secondary orders will replace the first one.
// In order to use this method as take profit tool
// use the same side as your position side.
func (bfa BinanceFuturesApi) PlaceStopMarketOrder(symbol, side string, stopPrice, qty float64) ([]byte, error) {
	parameters := url.Values{}
	parameters.Add("symbol", symbol)
	parameters.Add("side", side)
	parameters.Add("type", OrderTypeStopMarket)
	parameters.Add("quantity", strconv.FormatFloat(qty, 'f', -1, 64))
	parameters.Add("stopPrice", strconv.FormatFloat(stopPrice, 'f', -1, 64))
	return bfa.doSignedRequest("POST", orderEndPoint, parameters)
}

func (bfa BinanceFuturesApi) CancelSingleOrder(symbol, origClientOrderId string, orderId int64) ([]byte, error) {
	parameters := url.Values{}
	parameters.Add("symbol", symbol)
	parameters.Add("orderId", strconv.FormatInt(orderId, 10))
	parameters.Add("origClientOrderId", origClientOrderId)
	return bfa.doSignedRequest("DELETE", orderEndPoint, parameters)
}

func (bfa BinanceFuturesApi) CancelAllOrders(symbol string) ([]byte, error) {
	parameters := url.Values{}
	parameters.Add("symbol", symbol)
	return bfa.doSignedRequest("DELETE", allOpenOrdersEndPoint, parameters)
}

func (bfa BinanceFuturesApi) GetAccountBalance() ([]byte, error) {
	return bfa.doSignedRequest("GET", futuresAccountBalanceEndpoint, url.Values{})
}

func (bfa BinanceFuturesApi) GetAccountInformation() ([]byte, error) {
	return bfa.doSignedRequest("GET", accountInformationEndpoint, url.Values{})
}
