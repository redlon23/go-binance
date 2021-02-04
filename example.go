package go_binance

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/redlon23/go-binance/models"
	"log"
)

// pass as many channel as you subscribe to different streams
// in this example, assume you just use ticker and liquidation
// you should also pass same channels to your handlers to process them.
func channelControl(conn *websocket.Conn, tickerChan, liquidationChan chan []byte) {
	defer func() {
		connectionError := conn.Close()
		if connectionError != nil {
			log.Fatal(connectionError)
		}
	}()
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			continue
		}
		meta := new(models.StreamMetaMessage)
		err = json.Unmarshal(msg, &meta)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if meta.Event == models.EventSymbolTicker {
			tickerChan <- msg
		} else if meta.Event == models.EventLiquidation {
			liquidationChan <- msg
		} else {
			fmt.Println(string(msg))
		}
	}
}