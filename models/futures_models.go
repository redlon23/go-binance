package models

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
