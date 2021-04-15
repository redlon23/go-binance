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
	baseNetWSURLCoin        	= "wss://dstream.binance.com/ws/"
	testNetWSURLCoin			= "wss://dstream.binancefuture.com/ws/"
)

type BinanceFuturesCoinWebSocket struct {
	Connection *websocket.Conn
	BaseUrl string
	SubscribeIdCounter int
	Logger *logrus.Logger
}

func (bfcws *BinanceFuturesCoinWebSocket) PrepareLoggers()  {
	bfcws.Logger = logrus.New()
	bfcws.Logger.Formatter = new(logrus.JSONFormatter)


	wslogs, err := os.OpenFile("logs/binance_ws.log", os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		bfcws.Logger.SetOutput(wslogs)
	} else {
		fmt.Println("Failed to log to file for binance websocket, using default stderr")
	}
}

func (bfcws *BinanceFuturesCoinWebSocket) UseMainNet() {
	bfcws.BaseUrl = baseNetWSURLCoin
}

func (bfcws *BinanceFuturesCoinWebSocket) UseTestNet() {
	bfcws.BaseUrl = testNetWSURLCoin
}

func (bfcws *BinanceFuturesCoinWebSocket) IncrementSubscribeIdCounter() {
	bfcws.SubscribeIdCounter++
}

// Starts a websocket connection with default ping handler.
func (bfcws *BinanceFuturesCoinWebSocket) OpenWebSocketConnection() error  {
	connection, _, err := websocket.DefaultDialer.Dial(bfcws.BaseUrl, nil)
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
	bfcws.Connection = connection
	return nil
}


// Starts a websocket connection with default ping handler.
func (bfcws *BinanceFuturesCoinWebSocket) OpenWebSocketConnectionWithUserStream(listenKey string) error  {
	connection, _, err := websocket.DefaultDialer.Dial(bfcws.BaseUrl + listenKey, nil)
	if err != nil {
		bfcws.Logger.Println("Default Dialer, had an error during initial dial, ", err)
		return err
	}
	// -> The websocket server will send a ping frame every 5 minutes.
	//    If the websocket server does not receive a pong frame back from the connection
	//    within a 15 minute period, the connection will be disconnected.

	// -> Connection comes with default ping handler which sends pong in response to ping.
	//    Ping handler will be used internally by gorilla/websocket.

	// -> Passing nil to SetPingHandler will set default handler on the connection.
	connection.SetPingHandler(nil)
	bfcws.Connection = connection
	return nil
}

// Subscribes to given symbol-stream type over provided connection
func (bfcws BinanceFuturesCoinWebSocket) SubscribeToStream(symbol, streamType string) error {
	parameter := fmt.Sprintf("%s@%s", strings.ToLower(symbol), streamType)
	subscribeMap := models.LiveStream{Method: "SUBSCRIBE", Params: []string{parameter}, Id: bfcws.SubscribeIdCounter}
	bfcws.IncrementSubscribeIdCounter()
	err := bfcws.Connection.WriteJSON(subscribeMap)
	if err != nil {
		bfcws.Logger.Println("Error has occurred while sending subscribe message through connection", err)
		return err
	}
	return nil
}

func (bfcws BinanceFuturesCoinWebSocket) SubscribeLiquidationStream(symbol string) error {
	return bfcws.SubscribeToStream(symbol, liquidationStreamName)
}

func (bfcws BinanceFuturesCoinWebSocket) SubscribeBookTickerStream(symbol string) error {
	return bfcws.SubscribeToStream(symbol, bookTickerSteamName)
}

func (bfcws BinanceFuturesCoinWebSocket) SubscribeSymbolTickerStream(symbol string) error {
	return bfcws.SubscribeToStream(symbol, symbolTickerName)
}

func (bfcws BinanceFuturesCoinWebSocket) ReadFromConnection() (messageType int, p []byte, err error) {
	return bfcws.Connection.ReadMessage()
}

func (bfcws *BinanceFuturesCoinWebSocket) CloseConnection() error  {
	return bfcws.Connection.Close()
}