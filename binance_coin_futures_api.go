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
	mainNetBaseURLCoin 		= "https://dapi.binance.com"
	testNetBaseURLCoin 		= "https://testnet.binancefuture.com"
	ticker24HrEndPointCoin 	= "/dapi/v1/ticker/24hr"
	orderBookEndPointCoin 	= "/dapi/v1/depth"
	exchangeInformationEndPointCoin 	= "/dapi/v1/exchangeInfo"
	klinesEndpointCoin 	= "/dapi/v1/klines"
)

type BinanceCoinFuturesApi struct {
	Client *http.Client
	BaseUrl string
	PublicKey string
	SecretKey string
	Logger *logrus.Logger
}

func (bcfa *BinanceCoinFuturesApi) SetApiKeys(public, secret string) {
	bcfa.PublicKey = public
	bcfa.SecretKey = secret
}

func (bcfa *BinanceCoinFuturesApi) NewNetClient() {
	// Todo: make sure timeout and handshake won't cause any problems.
	var netTransport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 2 * time.Second,
		}).DialContext,
		DisableKeepAlives: false,
		TLSHandshakeTimeout: 2 * time.Second,
		ForceAttemptHTTP2:     true,
	}
	bcfa.Client = &http.Client{
		Timeout:   0,
		Transport: netTransport,
	}
}

func (bcfa *BinanceCoinFuturesApi) NewNetClientHTTP2() {
	var netTransport = &http2.Transport{}
	bcfa.Client = &http.Client{
		Timeout:   0,
		Transport: netTransport,
	}
}

func(bcfa *BinanceCoinFuturesApi) UseMainNet() {
	bcfa.BaseUrl = mainNetBaseURLCoin
}
func(bcfa *BinanceCoinFuturesApi) UseTestNet() {
	bcfa.BaseUrl = testNetBaseURLCoin
}

// Adds timestamp and creates a signature according to binance rules
func (bcfa BinanceCoinFuturesApi) signParameters(parameters *url.Values) string {
	tonce := strconv.FormatInt(time.Now().UnixNano(), 10)[0:13]
	parameters.Add("timestamp", tonce)
	signature, _ := bcfa.getSha256Signature(parameters.Encode())
	return signature
}

// Signs given parameter values
func (bcfa BinanceCoinFuturesApi) getSha256Signature(parameters string) (string, error) {
	mac := hmac.New(sha256.New, []byte(bcfa.SecretKey))
	_, err := mac.Write([]byte(parameters))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func (bcfa BinanceCoinFuturesApi) parseResponseBody(body io.ReadCloser) ([]byte, error) {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		bcfa.Logger.Println("Something went wrong during reading request body, ", err)
	}
	return data, nil
}

func (bcfa BinanceCoinFuturesApi) doPublicRequest(httpVerb, endPoint string, parameters url.Values) ([]byte, error) {
	fullURL := bcfa.BaseUrl + endPoint
	if parameters != nil {
		fullURL += "?" + parameters.Encode()
	}
	fmt.Println(fullURL)
	request, _ := http.NewRequest(httpVerb, fullURL, nil)
	response, err := bcfa.Client.Do(request)
	if err != nil {
		// Failure to speak HTTP - Connectivity error
		// Non 2XX status does not produce errors
		bcfa.Logger.Error("Connectivity error while Client.Do ", err)
		return nil, err
	}
	// Log rate limit for debug purposes
	// Even if request results in non 2xx status it provides
	// rate limit information
	bcfa.Logger.Println(endPoint + ", rate limit used: ",
		response.Header.Get("X-Mbx-Used-Weight-1m"))

	data, err := bcfa.parseResponseBody(response.Body)
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
		bcfa.Logger.Error(endPoint, err.Error())
		return nil, err
	}
	return data, nil
}

func (bcfa BinanceCoinFuturesApi) doSignedRequest(httpVerb, endPoint string, parameters url.Values) ([]byte, error) {
	signature := bcfa.signParameters(&parameters)
	headers := make(http.Header)
	headers.Add("X-MBX-APIKEY", bcfa.PublicKey)
	fullURL := bcfa.BaseUrl + endPoint + "?" + parameters.Encode() +"&signature=" + signature
	request, _ := http.NewRequest(httpVerb, fullURL, nil)
	request.Header = headers
	response, err := bcfa.Client.Do(request)
	if err != nil {
		// Failure to speak HTTP - Connectivity error
		// Non 2XX status does not produce errors
		bcfa.Logger.Error("Connectivity error while Client.Do ", err)
		return nil, err
	}
	// Log rate limit for debug purposes
	// Even if request results in non 2xx status it provides
	// rate limit information
	bcfa.Logger.Println(endPoint + ", rate limit used: ",
		response.Header.Get("X-Mbx-Used-Weight-1m"))

	data, err := bcfa.parseResponseBody(response.Body)
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
		bcfa.Logger.Error(endPoint, err.Error())
		return nil, err
	}
	return data, nil
}

func(bcfa BinanceCoinFuturesApi) GetExchangeInformation() ([]byte, error) {
	return bcfa.doPublicRequest("GET", exchangeInformationEndPointCoin, nil)
}

// 	Contains weighted average price (vwap)
func (bcfa BinanceCoinFuturesApi) Get24HourTickerPriceChangeStatistics(symbol string) ([]byte, error) {
	parameters := url.Values{}
	parameters.Add("symbol", symbol)
	return bcfa.doPublicRequest("GET", ticker24HrEndPointCoin, parameters)
}

func (bcfa BinanceCoinFuturesApi) GetOrderBook(symbol string, limit int) ([]byte, error)  {
	parameters := url.Values{}
	parameters.Add("symbol", symbol)
	parameters.Add("limit", fmt.Sprintf("%d", limit))
	return bcfa.doPublicRequest("GET", orderBookEndPointCoin, parameters)
}

// limit can be just 1 if only current candle is needed.
func(bcfa BinanceCoinFuturesApi) GetKlines(symbol, interval string, limit int) ([]byte, error) {
	parameters := url.Values{}
	parameters.Add("symbol", symbol)
	parameters.Add("interval", interval)
	parameters.Add("limit", fmt.Sprintf("%d", limit))
	return bcfa.doPublicRequest("GET", klinesEndpointCoin, parameters)
}



// ======================= SIGNED API CALLS ================================

func (bcfa BinanceCoinFuturesApi) GetUserStreamKey() ([]byte, error) {
	// todo change this end point, it was copied from futures api.
	return bcfa.doSignedRequest("POST", listenKeyEndPoint, url.Values{})
}