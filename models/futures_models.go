package models

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type ListenKey struct {
	Key string `json:"listenKey"`
}

type BalanceResponse []Balance

type Balance struct {
	Asset 				string 	`json:"asset"`
	Balance 			float64 `json:"balance,string"`
	AvailableBalance 	float64	`json:"availableBalance,string"`
}

type Vwap struct {
	Symbol 		string 	 `json:"symbol"`
	Vwap 		float64	 `json:"weightedAvgPrice,string"`
	LastPrice 	float64	 `json:"lastPrice,string"`
}

type OrderResponse struct {
	OrderId		int64 	`json:"orderId"`
	Symbol 		string 	`json:"symbol"`
	Side		string 	`json:"side"`
	Price 		float64 `json:"price,string"`
	Quantity 	float64 `json:"origQty,string"`
}

type PriceFilter struct {
	MaxPrice float64 `json:"maxPrice,string"`
	MinPrice float64 `json:"minPrice,string"`
	TickSize float64 `json:"tickSize,string"`
}

type LotFilter struct {
	MaxQuantity float64 `json:"maxQty,string"`
	MinQuantity float64 `json:"minQty,string"`
	StepSize float64 	`json:"stepSize,string"`
}


type ExchangeSymbolInformation struct {
	Symbol string `json:"symbol"`
	Filters  []map[string]interface{} `json:"filters"`
}

type ExchangeInformation struct {
	Symbols []ExchangeSymbolInformation `json:"symbols"`
}

type KlinesFrame struct {
	Open	float64
	High  	float64
	Low 	float64
	Close 	float64
	Volume 	float64
}

type Klines []KlinesFrame

func (k *KlinesFrame) UnmarshalJSON(data []byte) error {
	var v []interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		fmt.Printf("Error whilde decoding %v\n", err)
		return err
	}
	k.Open, _ = strconv.ParseFloat(v[1].(string), 64)
	k.High, _ = strconv.ParseFloat(v[2].(string), 64)
	k.Low, _ = strconv.ParseFloat(v[3].(string), 64)
	k.Close, _ = strconv.ParseFloat(v[4].(string), 64)
	k.Volume, _ = strconv.ParseFloat(v[5].(string), 64)
	return nil
}

type BookData struct {
	Price 	float64
	Quantity float64
}

type OrderBook struct {
	Bids []BookData `json:"bids"`
	Asks []BookData `json:"asks"`
}

func (bd *BookData) UnmarshalJSON(data []byte) error {
	var v []interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		fmt.Printf("Error whilde decoding %v\n", err)
		return err
	}
	bd.Price, _ = strconv.ParseFloat(v[0].(string), 64)
	bd.Quantity, _ = strconv.ParseFloat(v[1].(string), 64)
	return nil
}