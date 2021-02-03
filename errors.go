package go_binance

import "fmt"

const (
	TimestampWrong = -1021
	SignatureWrong = -1022
	ParameterValueWrong = -1102
	PrecisionWrong = -1111
	SymbolWrong = -1121
	ApiKeyWrong = -2014
	GreaterThanMaxQuantity = -4005
)

type BinanceErrorMessage struct {
	Code int `json:"code"`
	Message string `json:"msg"`
}

func (bem BinanceErrorMessage) ErrorMessage() string {
	return fmt.Sprintf("Code-> %d, Reason-> %s", bem.Code, bem.Message)
}

type RequestError struct {
	StatusCode int
	UrlUsed string
	Message BinanceErrorMessage
}

func (re *RequestError) Error() string {
	return fmt.Sprintf("Status Code: %d - url used: %s - Code: %d Reason: %s", re.StatusCode,
		re.UrlUsed, re.Message.Code, re.Message.Message)
}
