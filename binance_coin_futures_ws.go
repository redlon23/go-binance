package go_binance

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/redlon23/go-binance/models"
	"github.com/sirupsen/logrus"
	"log"
	"strings"
)

const (
	baseNetWSURLCoin        	= "wss://dstream.binance.com/ws/"
	testNetWSURLCoin			= "wss://dstream.binancefuture.com/ws/"
	liquidationStreamName 	= "forceOrder"
	bookTickerSteamName 	= "bookTicker"
	symbolTickerName 		= "ticker"
)


type BinanceCoinFuturesWebSocket struct {
	Connection *websocket.Conn
	BaseUrl string
	SubscribeIdCounter int
	Logger *logrus.Logger
}

func (bcfws *BinanceCoinFuturesWebSocket) UseMainNet() {
	bcfws.BaseUrl = baseNetWSURLCoin
}

func (bcfws *BinanceCoinFuturesWebSocket) UseTestNet() {
	bcfws.BaseUrl = testNetWSURLCoin
}

func (bcfws *BinanceCoinFuturesWebSocket) IncrementSubscribeIdCounter() {
	bcfws.SubscribeIdCounter++
}

// Starts a websocket connection with default ping handler.
func (bcfws *BinanceCoinFuturesWebSocket) OpenWebSocketConnection() error  {
	connection, _, err := websocket.DefaultDialer.Dial(baseNetWSURL, nil)
	if err != nil {
		log.Println("Default Dialer, had an error during initial dial, ", err)
		return err
	}
	// -> The websocket server will send a ping frame every 5 minutes.
	//    If the websocket server does not receive a pong frame back from the connection
	//    within a 15 minute period, the connection will be disconnected.

	// -> Connection comes with default ping handler which sends pong in response to ping.
	//    Ping handler will be used internally by gorilla/websocket.

	// -> Passing nil to SetPingHandler will set default handler on the connection.
	connection.SetPingHandler(nil)
	bcfws.Connection = connection
	return nil
}

// Starts a websocket connection with default ping handler.
func (bcfws *BinanceCoinFuturesWebSocket) OpenWebSocketConnectionWithUserStream(listenKey string) error  {
	connection, _, err := websocket.DefaultDialer.Dial(bcfws.BaseUrl + listenKey, nil)
	if err != nil {
		bcfws.Logger.Println("Default Dialer, had an error during initial dial, ", err)
		return err
	}
	// -> The websocket server will send a ping frame every 5 minutes.
	//    If the websocket server does not receive a pong frame back from the connection
	//    within a 15 minute period, the connection will be disconnected.

	// -> Connection comes with default ping handler which sends pong in response to ping.
	//    Ping handler will be used internally by gorilla/websocket.

	// -> Passing nil to SetPingHandler will set default handler on the connection.
	connection.SetPingHandler(nil)
	bcfws.Connection = connection
	return nil
}

// Subscribes to given symbol-stream type over provided connection
func (bcfws BinanceCoinFuturesWebSocket) SubscribeToStream(symbol, streamType string) error {
	parameter := fmt.Sprintf("%s@%s", strings.ToLower(symbol), streamType)
	subscribeMap := models.LiveStream{Method: "SUBSCRIBE", Params: []string{parameter}, Id: bcfws.SubscribeIdCounter}
	bcfws.IncrementSubscribeIdCounter()
	err := bcfws.Connection.WriteJSON(subscribeMap)
	if err != nil {
		bcfws.Logger.Println("Error has occurred while sending subscribe message through connection", err)
		return err
	}
	return nil
}

func (bcfws BinanceCoinFuturesWebSocket) SubscribeLiquidationStream(symbol string) error {
	return bcfws.SubscribeToStream(symbol, liquidationStreamName)
}

func (bcfws BinanceCoinFuturesWebSocket) SubscribeBookTickerStream(symbol string) error {
	return bcfws.SubscribeToStream(symbol, bookTickerSteamName)
}

func (bcfws BinanceCoinFuturesWebSocket) SubscribeSymbolTickerStream(symbol string) error {
	return bcfws.SubscribeToStream(symbol, symbolTickerName)
}