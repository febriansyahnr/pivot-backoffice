package merchantTopUp

import "time"

type MerchantTopUp struct {
	ID              string     `db:"uuid" json:"id"`
	MerchantID      string     `db:"merchant_id" json:"merchantId"`
	AccountName     string     `db:"account_name" json:"accountName"`
	PaymentMethodID string     `db:"payment_method_id" json:"paymentMethodId"`
	ReferenceNumber string     `db:"reference_number" json:"referenceNumber"`
	CreatedAt       time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedAt       *time.Time `db:"deleted_at" json:"-"`
	Instructions    *string    `db:"instructions" json:"instructions,omitempty"`
}
