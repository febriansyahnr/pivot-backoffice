package merchantTopUp

import "time"

// TopUpTransactionListRequest represents the request for querying top-up transactions
type TopUpTransactionListRequest struct {
	MerchantId    string
	StartDate     time.Time
	EndDate       time.Time
	Status        string // SUCCESS, PENDING, FAILED
	TransactionID string // UUID from account_transactions
	ReferenceID   string // UUID from merchant_top_up_references
	SortOrder     string // asc or desc
	Page          int64
	PerPage       int64
}

// TopUpTransactionResponse represents a top-up transaction with payment method details
type TopUpTransactionResponse struct {
	UUID                string    `json:"id" db:"uuid"`
	ReferenceID         string    `json:"trxId" db:"reference_id"`
	MerchantReferenceID string    `json:"merchantReferenceId" db:"merchant_reference_id"`
	Type                string    `json:"type" db:"type"`
	Channel             string    `json:"channel" db:"channel"`
	Date                time.Time `json:"date" db:"date"`
	Amount              float64   `json:"amount" db:"amount"`
	Status              string    `json:"status" db:"status"`
	BalanceType         string    `json:"balanceType" db:"balance_type"`
}
