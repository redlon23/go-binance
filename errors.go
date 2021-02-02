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

type RequestError struct {
	StatusCode int
	UrlUsed string
	Message string
}

func (re *RequestError) Error() string {
	return fmt.Sprintf("Status Code: %d - url used: %s - Message: %s", re.StatusCode, re.UrlUsed, re.Message)
}
