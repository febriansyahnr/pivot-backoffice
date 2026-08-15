package paymentCaptureModel

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type PaymentCapture struct {
	ID                     string    `db:"id" json:"id"`
	PaymentID              string    `db:"payment_id" json:"paymentId"`
	ProcessorReferenceID   *string   `db:"processor_reference_id" json:"processorReferenceId"`
	Status                 string    `db:"status" json:"status"`
	ReleaseRemainingAmount bool      `db:"release_remaining_amount" json:"releaseRemainingAmount"`
	Currency               string    `db:"currency" json:"currency"`
	Amount                 float64   `db:"amount" json:"amount"`
	CreatedAt              time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt              time.Time `db:"updated_at" json:"updatedAt"`
}

func (p *PaymentCapture) UnmarshalJSON(data []byte) error {
	// 1. Define a shadow type with the same fields
	type Alias PaymentCapture

	// 2. Create a temp struct with the alias + raw field
	aux := struct {
		ReleaseRemainingAmount json.RawMessage `json:"releaseRemainingAmount"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	// 3. Unmarshal normally into aux (fills all fields automatically)
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 4. Custom parsing for ONLY releaseRemainingAmount
	if len(aux.ReleaseRemainingAmount) > 0 {
		releaseRemainingAmount, err := strconv.ParseBool(string(aux.ReleaseRemainingAmount))
		if err != nil {
			return fmt.Errorf("invalid releaseRemainingAmount: %s", aux.ReleaseRemainingAmount)
		}

		p.ReleaseRemainingAmount = releaseRemainingAmount
	}

	return nil
}
