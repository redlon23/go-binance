package go_binance

type BinanceFutures interface {
	SetApiKeys(public, secret string)
	NewNetClient()
	NewNetClientHTTP2()
	UseMainNet()
	UseTestNet()
	Get24HourTickerPriceChangeStatistics(symbol string) ([]byte, error)
	GetOrderBook(symbol string, limit int) ([]byte, error)
	GetExchangeInformation() ([]byte, error)
	GetKlines(symbol, interval string, limit int) ([]byte, error)
	GetUserStreamKey() ([]byte, error)
	UpdateKeepAliveUserStream() ([]byte, error)
	DeleteUserStream() ([]byte, error)
	PlaceLimitOrder(symbol, side string, price, qty float64, reduceOnly bool) ([]byte, error)
	PlacePostOnlyLimitOrder(symbol, side string, price, qty float64, reduceOnly bool) ([]byte, error)	//PlaceMarketOrder(symbol, side string, qty float64, reduceOnly bool) ([]byte, error)
	PlaceMarketOrder(symbol, side string, qty float64, reduceOnly bool) ([]byte, error)
	PlaceStopMarketOrder(symbol, side string, stopPrice, qty float64) ([]byte, error)
	CancelSingleOrder(symbol, origClientOrderId string, orderId int64) ([]byte, error)
	CancelAllOrders(symbol string) ([]byte, error)
	GetAccountBalance() ([]byte, error)
	GetAccountInformation() ([]byte, error)
	PrepareLoggers()
}

type BinanceFutureSocket interface {
	UseMainNet()
	UseTestNet()
	IncrementSubscribeIdCounter()
	OpenWebSocketConnection() error
	OpenWebSocketConnectionWithUserStream(listenKey string) error
	SubscribeToStream(symbol, streamType string) error
	SubscribeLiquidationStream(symbol string) error
	SubscribeBookTickerStream(symbol string) error
	SubscribeSymbolTickerStream(symbol string) error
	PrepareLoggers()

	// ReadFromConnection important part
	ReadFromConnection() (messageType int, p []byte, err error)
}