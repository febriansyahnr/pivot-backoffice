package controller_validator

import "github.com/go-playground/validator/v10"

const (
	USD = "USD"
	IDR = "IDR"
)

// IsSupportedCurrency returns true if the currency is supported
func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, IDR:
		return true
	}
	return false
}

var CurrencyValidator validator.Func = func(fieldLevel validator.FieldLevel) bool {
	if currency, ok := fieldLevel.Field().Interface().(string); ok {
		return IsSupportedCurrency(currency)
	}
	return false
}
