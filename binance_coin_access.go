package go_binance

import (
"fmt"
"github.com/redlon23/go-binance/models"
"github.com/sirupsen/logrus"
"log"
"os"
)



type BinanceCoinAccess struct {
	Api BinanceCoinFuturesApi
	WebSocket BinanceFuturesWebSocket
}

func (bca *BinanceCoinAccess) SetApiKeys(public, secret string) {
	bca.Api.SetApiKeys(public, secret)
}

func(bca *BinanceCoinAccess) UseMainNet() {
	bca.Api.UseMainNet()
	bca.WebSocket.UseMainNet()
}

func(bca *BinanceCoinAccess) UseTestNet() {
	bca.Api.UseTestNet()
	bca.WebSocket.UseTestNet()
}

func (bca *BinanceCoinAccess) PrepareLoggers() {
	bca.Api.Logger = logrus.New()
	bca.WebSocket.Logger = logrus.New()
	bca.Api.Logger.Formatter = new(logrus.JSONFormatter)
	bca.WebSocket.Logger.Formatter = new(logrus.JSONFormatter)
	// Check log folder and create if it doesn't exists
	CheckLogsFolder()

	apilogs, err := os.OpenFile("logs/binance_api.log", os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		bca.Api.Logger.SetOutput(apilogs)
	} else {
		fmt.Println("Failed to log to file for binance api calls, using default stderr")
	}

	wslogs, err := os.OpenFile("logs/binance_ws.log", os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		bca.WebSocket.Logger.SetOutput(wslogs)
	} else {
		fmt.Println("Failed to log to file for binance websocket, using default stderr")
	}
}


func(bca *BinanceCoinAccess) PrepareAccess() {
	fmt.Println("HTTP/2.0 Client Preparing...")
	bca.Api.NewNetClientHTTP2()
	fmt.Println("Websocket Connection opening...")
	err := bca.WebSocket.OpenWebSocketConnection()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Success!")
}

// Uses http2 client then opens a websocket connection
// Handles listen key generation and enables user stream
func(bca *BinanceCoinAccess) PrepareAccessWithUserStream() {
	fmt.Println("HTTP/2.0 Client Preparing...")
	bca.Api.NewNetClientHTTP2()
	fmt.Println("Websocket Connection (User Stream) opening...")
	data, err := bca.Api.GetUserStreamKey()
	if err != nil {
		log.Fatal(err)
	}
	listenKey := GetKey(data)
	err = bca.WebSocket.OpenWebSocketConnectionWithUserStream(listenKey)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Success !")
}


func (bca *BinanceCoinAccess) TestUserStream(positionChannel, orderChannel, priceChannel chan []byte) {
	_ = bca.WebSocket.SubscribeSymbolTickerStream("BTCUSDT")
	for {
		_, message, err := bca.WebSocket.Connection.ReadMessage()
		if err != nil {
			log.Println("Error occurred while reading message from the connection", err)
		} else {
			fmt.Println(string(message))

			// Unmarshal Meta Information
			meta, err := ParseMetaInformation(message)
			if err != nil {
				log.Println(err)
				log.Println(meta)
			}

			if meta.Event == models.EventSymbolTicker {
				priceChannel <- message
			} else if meta.Event == models.EventOrder {
				// Order: new, canceled, expired
				orderChannel <- message
			} else if meta.Event == models.EventAccount &&
				meta.Reason.MessageType == models.ReasonOrder{
				// Account: balance, position, funding, adjustment (look below)
				// !!! -> Only order related messages are accepted <- !!!
				positionChannel <- message
			}
		}
	}
}