package util

import (
	"math"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func ConvertFloatToCurrency(number float64) string {
	if number == 0 {
		return "0"
	}

	p := message.NewPrinter(language.Indonesian)
	formattedNumber := p.Sprintf("%.0f", number)

	return formattedNumber
}

func HasDecimal(value float64) bool {
	return value != math.Floor(value)
}
