package models

const (
	ReasonOrder       = "ORDER"
	ReasonFunding     = "FUNDING_FEE"
	EventAccount      = "ACCOUNT_UPDATE"
	EventOrder 		  = "ORDER_TRADE_UPDATE"
	EventSymbolTicker = "24hrTicker"
)

// Used for sending subscribe and unsubscribe message in websocket
type LiveStream struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	Id     int      `json:"id"`
}

// Identification meta for incoming websocket messages
type StreamMetaMessage struct {
	EventTime int `json:"E"`
	Event string `json:"e"`
	Reason struct {
		MessageType string `json:"m"`
	} `json:"a"`
}

type Position struct {
	Symbol 		string `json:"s"`
	Side 		string
	Quantity 	float64 `json:"pa,string"`
	EntryPrice 	float64 `json:"ep,string"`
}

type StreamAccountUpdate struct {
	UpdateData struct{
		Positions []Position `json:"P"`
	} `json:"a"`
}

type StreamOrderUpdate struct {
	Order struct{
		Symbol 			string `json:"s"`
		Side 			string `json:"S"`
		Type 			string `json:"o"`
		ExecutionStatus string `json:"X"`
		ExecutionType 	string `json:"x"`
		ClientId 		string `json:"c"`
	} `json:"o"`
}

type StreamSymbolTickerUpdate struct {
	Symbol string `json:"s"`
	ChangePercent int `json:"C"`
	Price float64 `json:"c,string"`
	Wvap float64 `json:"w,string"`
}

type LiquidationInformation struct{
	Symbol 				string  `json:"s"`
	Side 				string  `json:"S"`
	Type 				string  `json:"o"`
	OriginalQuantity	float64 `json:"q,string"`
	AccumulatedQuantity float64 `json:"Q,string"`
	ExecutionStatus  	string  `json:"X"`
	ExecutionType 	 	string  `json:"x"`
	ClientId 		 	string  `json:"c"`
}

type LiquidationOrder struct {
	Liquidation LiquidationInformation `json:"o"`
}