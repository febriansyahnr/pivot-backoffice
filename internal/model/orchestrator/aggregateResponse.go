package orchestrator_model

type AggregateResponse struct {
	CountOfCredit   int     `json:"countOfCredit" db:"count_of_credit"`
	CountOfDebit    int     `json:"countOfDebit" db:"count_of_debit"`
	SumOfCredit     float64 `json:"sumOfCredit" db:"sum_of_credit"`
	SumOfDebit      float64 `json:"sumOfDebit" db:"sum_of_debit"`
	SumOfPendCredit float64 `json:"-" db:"sum_of_pend_credit"`
	SumOfPendDebit  float64 `json:"-" db:"sum_of_pend_debit"`
}

type BulkAggregateResponse struct {
	MerchantID    string  `json:"merchantId" db:"merchant_id"`
	AccountID     string  `json:"accountId" db:"account_id"`
	CountOfCredit int     `json:"countOfCredit" db:"count_of_credit"`
	CountOfDebit  int     `json:"countOfDebit" db:"count_of_debit"`
	SumOfCredit   float64 `json:"sumOfCredit" db:"sum_of_credit"`
	SumOfDebit    float64 `json:"sumOfDebit" db:"sum_of_debit"`
}
