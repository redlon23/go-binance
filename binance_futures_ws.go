package go_binance

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/redlon23/go-binance/models"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"strings"
)

const (
	baseNetWSURL        	= "wss://fstream.binance.com/ws/"
	testNetWSURL			= "wss://stream.binancefuture.com/ws/"
	liquidationStreamName 	= "forceOrder"
	bookTickerSteamName 	= "bookTicker"
	symbolTickerName 		= "ticker"
)

type BinanceFuturesWebSocket struct {
	Connection *websocket.Conn
	BaseUrl string
	SubscribeIdCounter int
	Logger *logrus.Logger
}

func (bfws *BinanceFuturesWebSocket) PrepareLoggers()  {
	bfws.Logger = logrus.New()
	bfws.Logger.Formatter = new(logrus.JSONFormatter)


	wslogs, err := os.OpenFile("logs/binance_ws.log", os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		bfws.Logger.SetOutput(wslogs)
	} else {
		fmt.Println("Failed to log to file for binance websocket, using default stderr")
	}
}

func (bfws *BinanceFuturesWebSocket) UseMainNet() {
	bfws.BaseUrl = baseNetWSURL
}

func (bfws *BinanceFuturesWebSocket) UseTestNet() {
	bfws.BaseUrl = testNetWSURL
}

func (bfws *BinanceFuturesWebSocket) IncrementSubscribeIdCounter() {
	bfws.SubscribeIdCounter++
}

// Starts a websocket connection with default ping handler.
func (bfws *BinanceFuturesWebSocket) OpenWebSocketConnection() error  {
	connection, _, err := websocket.DefaultDialer.Dial(bfws.BaseUrl, nil)
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
	bfws.Connection = connection
	return nil
}


// Starts a websocket connection with default ping handler.
func (bfws *BinanceFuturesWebSocket) OpenWebSocketConnectionWithUserStream(listenKey string) error  {
	connection, _, err := websocket.DefaultDialer.Dial(bfws.BaseUrl + listenKey, nil)
	if err != nil {
		bfws.Logger.Println("Default Dialer, had an error during initial dial, ", err)
		return err
	}
	// -> The websocket server will send a ping frame every 5 minutes.
	//    If the websocket server does not receive a pong frame back from the connection
	//    within a 15 minute period, the connection will be disconnected.

	// -> Connection comes with default ping handler which sends pong in response to ping.
	//    Ping handler will be used internally by gorilla/websocket.

	// -> Passing nil to SetPingHandler will set default handler on the connection.
	connection.SetPingHandler(nil)
	bfws.Connection = connection
	return nil
}

// Subscribes to given symbol-stream type over provided connection
func (bfws BinanceFuturesWebSocket) SubscribeToStream(symbol, streamType string) error {
	parameter := fmt.Sprintf("%s@%s", strings.ToLower(symbol), streamType)
	subscribeMap := models.LiveStream{Method: "SUBSCRIBE", Params: []string{parameter}, Id: bfws.SubscribeIdCounter}
	bfws.IncrementSubscribeIdCounter()
	err := bfws.Connection.WriteJSON(subscribeMap)
	if err != nil {
		bfws.Logger.Println("Error has occurred while sending subscribe message through connection", err)
		return err
	}
	return nil
}

func (bfws BinanceFuturesWebSocket) SubscribeLiquidationStream(symbol string) error {
	return bfws.SubscribeToStream(symbol, liquidationStreamName)
}

func (bfws BinanceFuturesWebSocket) SubscribeBookTickerStream(symbol string) error {
	return bfws.SubscribeToStream(symbol, bookTickerSteamName)
}

func (bfws BinanceFuturesWebSocket) SubscribeSymbolTickerStream(symbol string) error {
	return bfws.SubscribeToStream(symbol, symbolTickerName)
}

func (bfws BinanceFuturesWebSocket) ReadFromConnection() (messageType int, p []byte, err error) {
	return bfws.Connection.ReadMessage()
}

func (bfws *BinanceFuturesWebSocket) CloseConnection() error  {
	return bfws.Connection.Close()
}