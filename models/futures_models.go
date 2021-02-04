package models

type ListenKey struct {
	Key string `json:"listenKey"`
}

type BalanceResponse []Balance

type Balance struct {
	Asset 				string `json:"asset"`
	Balance 			float64 `json:"balance,string"`
	AvailableBalance 	float64`json:"availableBalance,string"`
}

type Wvap struct {
	Symbol 		string 		`json:"symbol"`
	WVap 		float64		`json:"weightedAvgPrice,string"`
	LastPrice 	float64		`json:"lastPrice,string"`
}


