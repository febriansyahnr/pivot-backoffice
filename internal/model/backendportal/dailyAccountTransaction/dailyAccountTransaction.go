package dailyAccountTransactionModel

import "time"

// DailyAccountTransaction represents the schema of the daily_account_transactions table
type DailyAccountTransaction struct {
	ID           string    `json:"id" db:"id"`                      // Primary key
	AccountID    string    `json:"accountId" db:"account_id"`       // Account identifier
	Date         time.Time `json:"date" db:"date"`                  // Date of the transaction
	Timezone     string    `json:"timezone" db:"timezone"`          // Timezone information
	BegBalance   float64   `json:"begBalance" db:"beg_balance"`     // Beginning balance
	DebitTrx     int       `json:"debitTrx" db:"debit_trx"`         // Number of debit transactions
	DebitAmount  float64   `json:"debitAmount" db:"debit_amount"`   // Total debit amount
	CreditTrx    int       `json:"creditTrx" db:"credit_trx"`       // Number of credit transactions
	CreditAmount float64   `json:"creditAmount" db:"credit_amount"` // Total credit amount
	EODBalance   float64   `json:"eodBalance" db:"eod_balance"`     // End of day balance
	CreatedAt    time.Time `json:"created_at" db:"created_at"`      // Record creation timestamp
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`      // Record update timestamp
}
