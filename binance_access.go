package go_binance

import (
	"encoding/json"
	"fmt"
	"github.com/redlon23/go-binance/models"
	"github.com/sirupsen/logrus"
	"log"
	"os"
)

func getKey(data []byte) string {
	k := new(models.ListenKey)
	if err := json.Unmarshal(data, &k); err != nil {
		log.Println(err)
		return ""
	}
	return k.Key
}

func parseMetaInformation(wsMessage []byte) (*models.StreamMetaMessage, error) {
	meta := new(models.StreamMetaMessage)
	err := json.Unmarshal(wsMessage, &meta)
	if err != nil {
		log.Println(err)
		return meta, err
	}
	return meta, nil
}

// Checks if log folder exists, if not creates one
func checkLogsFolder() {
	if _, err := os.Stat("./logs"); os.IsNotExist(err) {
		fmt.Println("No logs folder in the current directory, creating one...")
		err := os.Mkdir("logs", 066)
		if err != nil {
			fmt.Println("Failed to create a folder!")
		}
	} else {
		fmt.Println("There is already a logs folder in the current directory!")
	}
}

type BinanceAccess struct {
	Api BinanceFuturesApi
	WebSocket BinanceFuturesWebSocket
}

func (ba *BinanceAccess) SetApiKeys(public, secret string) {
	ba.Api.SetApiKeys(public, secret)
}

func(ba *BinanceAccess) UseMainNet() {
	ba.Api.UseMainNet()
	ba.WebSocket.UseMainNet()
}

func(ba *BinanceAccess) UseTestNet() {
	ba.Api.UseTestNet()
	ba.WebSocket.UseTestNet()
}

func (ba *BinanceAccess) PrepareLoggers() {
	ba.Api.Logger = logrus.New()
	ba.WebSocket.Logger = logrus.New()
	ba.Api.Logger.Formatter = new(logrus.JSONFormatter)
	ba.WebSocket.Logger.Formatter = new(logrus.JSONFormatter)
	// Check log folder and create if it doesn't exists
	checkLogsFolder()

	apilogs, err := os.OpenFile("logs/binance_api.log", os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		ba.Api.Logger.SetOutput(apilogs)
	} else {
		fmt.Println("Failed to log to file for binance api calls, using default stderr")
	}

	wslogs, err := os.OpenFile("logs/binance_ws.log", os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		ba.WebSocket.Logger.SetOutput(wslogs)
	} else {
		fmt.Println("Failed to log to file for binance websocket, using default stderr")
	}
}


func(ba *BinanceAccess) PrepareAccess() {
	fmt.Println("HTTP/2.0 Client Preparing...")
	ba.Api.NewNetClientHTTP2()
	fmt.Println("Websocket Connection opening...")
	err := ba.WebSocket.OpenWebSocketConnection()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Success!")
}

// Uses http2 client then opens a websocket connection
// Handles listen key generation and enables user stream
func(ba *BinanceAccess) PrepareAccessWithUserStream() {
	fmt.Println("HTTP/2.0 Client Preparing...")
	ba.Api.NewNetClientHTTP2()
	fmt.Println("Websocket Connection (User Stream) opening...")
	data, err := ba.Api.GetUserStreamKey()
	if err != nil {
		log.Fatal(err)
	}
	listenKey := getKey(data)
	err = ba.WebSocket.OpenWebSocketConnectionWithUserStream(listenKey)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Success !")
}


func (ba *BinanceAccess) TestUserStream(positionChannel, orderChannel, priceChannel chan []byte) {
	_ = ba.WebSocket.SubscribeSymbolTickerStream("BTCUSDT")
	for {
		_, message, err := ba.WebSocket.Connection.ReadMessage()
		if err != nil {
			log.Println("Error occurred while reading message from the connection", err)
		} else {
			fmt.Println(string(message))

			// Unmarshal Meta Information
			meta, err := parseMetaInformation(message)
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