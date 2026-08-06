package adjustment

import (
	"database/sql"
	"time"
)

type ManualAdjustmentHistory struct {
	UUID            string       `db:"uuid"`
	MerchantID      string       `db:"merchant_id"`
	TransactionDate time.Time    `db:"transaction_date"`
	BankRefID       string       `db:"bank_reference_id"`
	BankAccount     string       `db:"bank_account"`
	Type            string       `db:"type"`
	Action          string       `db:"action"`
	Currency        string       `db:"currency"`
	Amount          float64      `db:"amount"`
	ReferenceID     string       `db:"reference_id"`
	ProofOfTransfer string       `db:"proof_of_transfer"`
	Notes           string       `db:"notes"`
	CreatedBy       string       `db:"created_by"`
	CreatedAt       time.Time    `db:"created_at"`
	UpdatedAt       time.Time    `db:"updated_at"`
	DeletedAt       sql.NullTime `db:"deleted_at"`
}

type BankAccount struct {
	Name      string `json:"name"`
	AccNumber string `json:"account_number"`
}
