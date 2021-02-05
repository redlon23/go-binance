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
	WVap 		float64	 `json:"weightedAvgPrice,string"`
	LastPrice 	float64	 `json:"lastPrice,string"`
}

type OrderResponse struct {
	OrderId		int64 	`json:"orderId"`
	Symbol 		string 	`json:"symbol"`
	Side		string 	`json:"side"`
	Price 		float64 `json:"price,string"`
	Quantity 	float64 `json:"origQty,string"`
}
