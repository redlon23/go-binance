package go_binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/redlon23/go-binance/models"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var (
	// ====== URL & END POINTS ======
	baseURL            = "https://fapi.binance.com"
	testNetBaseURL     = "https://testnet.binancefuture.com"
	ticker24HrEndPoint = "/fapi/v1/ticker/24hr"
	ListenEndPoint     = "/fapi/v1/listenKey"
	OrderEndPoint	   = "/fapi/v1/order"

	// ====== Parameter Types ======
	SideBuy 			= "BUY"
	SideSell 			= "SELL"

	OrderTypeLimit 		= "LIMIT"
	OrderTypeMarket 	= "MARKET"
	GoodTillCancel 		= "GTC"
	GoodTillCrossing 	= "GTX"
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
	// Todo: make sure timeout and handshake time out won't cause any problems.
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
	bfa.BaseUrl = baseURL
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
		log.Println("Something went wrong during reading request body, ", err)
	}
	return data, err
}

// 	Contains weighted average price (wvap)
func (bfa BinanceFuturesApi) Get24HourTickerPriceChangeStatistics(symbol string) []byte {
	fullURL := bfa.BaseUrl + ticker24HrEndPoint + "?symbol=" + symbol
	request, err := http.NewRequest("GET", fullURL, nil)

	if err != nil {
		log.Println(err)
	}

	response, err := bfa.Client.Do(request)
	if err != nil {
		panic(err)
	}
	data, err := bfa.parseResponseBody(response.Body)
	if err != nil {
		panic(err)
	}
	err = response.Body.Close()
	if err != nil {
		panic(err)
	}

	return data
}

func (bfa BinanceFuturesApi) GetWvap(symbol string) models.Wvap {
	data := bfa.Get24HourTickerPriceChangeStatistics(symbol)
	wvap := new(models.Wvap)
	err := json.Unmarshal(data, &wvap)
	if err != nil {
		log.Println(err)
	}
	return *wvap
}

func (bfa BinanceFuturesApi) GetUserStreamKey() ([]byte, error) {
	fullUrl := bfa.BaseUrl  + ListenEndPoint + "?"
	parameters := url.Values{}
	signature := bfa.signParameters(&parameters)
	headers := make(http.Header)
	headers.Add("X-MBX-APIKEY", bfa.PublicKey)
	requestUrl := fullUrl + "signature=" + signature
	req, err := http.NewRequest("POST", requestUrl, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header = headers
	response, err := bfa.Client.Do(req)
	if err != nil {
		return nil, err
	}
	data, _ := bfa.parseResponseBody(response.Body)
	return data, nil
}

func (bfa BinanceFuturesApi) KeepAliveUserStream() ([]byte, error) {
	fullUrl := bfa.BaseUrl  + ListenEndPoint + "?"
	parameters := url.Values{}
	signature := bfa.signParameters(&parameters)
	headers := make(http.Header)
	headers.Add("X-MBX-APIKEY", bfa.PublicKey)
	requestUrl := fullUrl + "signature=" + signature
	req, err := http.NewRequest("PUT", requestUrl, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header = headers
	response, err := bfa.Client.Do(req)
	if err != nil {
		return nil, err
	}
	data, _ := bfa.parseResponseBody(response.Body)
	return data, nil
}

func (bfa BinanceFuturesApi) PlaceLimitOrder(symbol, side string, price, qty float64) ([]byte, error) {
	fullUrl := bfa.BaseUrl  + OrderEndPoint + "?"
	parameters := url.Values{}
	parameters.Add("symbol", symbol)
	parameters.Add("side", side)
	parameters.Add("type", OrderTypeLimit)
	parameters.Add("timeInForce", GoodTillCancel)
	// TODO: Price and quantity precision must be dynamic, it is based on each symbol.
	parameters.Add("quantity", fmt.Sprintf("%.3f", qty))
	parameters.Add("price", fmt.Sprintf("%.2f", price))
	signature := bfa.signParameters(&parameters)
	headers := make(http.Header)
	headers.Add("X-MBX-APIKEY", bfa.PublicKey)
	// Include parameters in the request url
	requestUrl := fullUrl + parameters.Encode() +"&signature=" + signature
	req, err := http.NewRequest("POST", requestUrl, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header = headers
	response, err := bfa.Client.Do(req)
	if err != nil {
		return nil, err
	}
	data, _ := bfa.parseResponseBody(response.Body)
	return data, nil
}

func (bfa BinanceFuturesApi) PlacePostOnlyLimitOrder(symbol, side string, price, qty float64) ([]byte, error) {
	fullUrl := bfa.BaseUrl  + OrderEndPoint + "?"
	parameters := url.Values{}
	parameters.Add("symbol", symbol)
	parameters.Add("side", side)
	parameters.Add("type", OrderTypeLimit)
	parameters.Add("timeInForce", GoodTillCrossing)
	// TODO: Price and quantity precision must be dynamic, it is based on each symbol.
	parameters.Add("quantity", fmt.Sprintf("%.3f", qty))
	parameters.Add("price", fmt.Sprintf("%.2f", price))
	signature := bfa.signParameters(&parameters)
	headers := make(http.Header)
	headers.Add("X-MBX-APIKEY", bfa.PublicKey)
	requestUrl := fullUrl + parameters.Encode() +"&signature=" + signature
	req, err := http.NewRequest("POST", requestUrl, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header = headers
	response, err := bfa.Client.Do(req)
	if err != nil {
		return nil, err
	}
	data, _ := bfa.parseResponseBody(response.Body)
	return data, nil
}

func (bfa BinanceFuturesApi) PlaceMarketOrder(symbol, side string, qty float64) ([]byte, error) {
	fullUrl := bfa.BaseUrl  + OrderEndPoint + "?"
	parameters := url.Values{}
	parameters.Add("symbol", symbol)
	parameters.Add("side", side)
	parameters.Add("type", OrderTypeMarket)
	// TODO: Price and quantity precision must be dynamic, it is based on each symbol.
	parameters.Add("quantity", fmt.Sprintf("%.3f", qty))
	signature := bfa.signParameters(&parameters)
	headers := make(http.Header)
	headers.Add("X-MBX-APIKEY", bfa.PublicKey)
	requestUrl := fullUrl + parameters.Encode() +"&signature=" + signature
	req, err := http.NewRequest("POST", requestUrl, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header = headers
	response, err := bfa.Client.Do(req)
	if err != nil {
		return nil, err
	}
	data, _ := bfa.parseResponseBody(response.Body)
	return data, nil
}
