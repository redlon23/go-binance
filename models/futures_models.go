package models

type ListenKey struct {
	Key string `json:"listenKey"`
}

type Wvap struct {
	Symbol 		string 		`json:"symbol"`
	WVap 		float64		`json:"weightedAvgPrice,string"`
	LastPrice 	float64		`json:"lastPrice,string"`
}


