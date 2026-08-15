package snapCoreModel

import (
	"fmt"
	"strconv"

	"github.com/shopspring/decimal"
)

type Amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

func (a *Amount) Decimal() decimal.Decimal {
	if dec, err := decimal.NewFromString(a.Value); err != nil {
		return decimal.Zero
	} else {
		return dec
	}
}

// String (ISO4217)
func (a *Amount) SnapFormat() error {
	value, err := strconv.ParseFloat(a.Value, 64)
	if err != nil {
		return err
	}

	a.Value = fmt.Sprintf("%.2f", value)
	return nil
}
