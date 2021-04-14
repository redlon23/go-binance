package go_binance

import (
	"fmt"
	"github.com/redlon23/go-binance/models"
	"log"
)

type BinanceAccess struct {
	Api BinanceFutures
	WebSocket BinanceFutureSocket
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
	// Check log folder and create if it doesn't exists
	CheckLogsFolder()

	ba.Api.PrepareLoggers()
	ba.WebSocket.PrepareLoggers()
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
	listenKey := GetKey(data)
	err = ba.WebSocket.OpenWebSocketConnectionWithUserStream(listenKey)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Success !")
}


func (ba *BinanceAccess) TestUserStream(positionChannel, orderChannel, priceChannel chan []byte) {
	_ = ba.WebSocket.SubscribeSymbolTickerStream("BTCUSDT")
	for {
		_, message, err := ba.WebSocket.ReadFromConnection()
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