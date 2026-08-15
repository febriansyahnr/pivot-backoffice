package reconciliation

import (
	"time"

	"github.com/shopspring/decimal"
)

type ReconResult struct {
	Transactions []*Transaction
	Status       string
	ErrorReason  string
	VAStatic     *ReconVAStatic
}

func (r *ReconResult) GetFirstLatestDate() (firstDate *time.Time, lastDate *time.Time) {
	if len(r.Transactions) == 0 {
		return nil, nil
	}

	if len(r.Transactions) == 0 {
		return nil, nil
	}
	// find first date
	firstDate = &r.Transactions[0].TransactionDate
	lastDate = &r.Transactions[0].TransactionDate
	// iterate through the transactions to find the first and latest date
	for _, transaction := range r.Transactions {
		if transaction.TransactionDate.Before(*firstDate) {
			firstDate = &transaction.TransactionDate
		}
		if transaction.TransactionDate.After(*lastDate) {
			lastDate = &transaction.TransactionDate
		}
	}
	return firstDate, lastDate
}

func (r *ReconResult) ShouldReconcileVAStatic() bool {
	if r.VAStatic == nil || len(*r.VAStatic) == 0 {
		return false
	}
	return true
}

type ReconVAStaticResult struct {
	Indexes     []int
	UUIDs       []string
	IsValid     bool
	TotalAmount decimal.Decimal
}

type ReconVAStatic map[string]*ReconVAStaticResult

func (r *ReconVAStatic) Add(reference string, index int, amount decimal.Decimal, trx *ReconTransactionModel) {
	if _, ok := (*r)[reference]; !ok {
		(*r)[reference] = &ReconVAStaticResult{
			Indexes:     []int{index},
			UUIDs:       []string{trx.UUID},
			TotalAmount: amount,
		}
	} else {
		va := (*r)[reference]
		va.Indexes = append(va.Indexes, index)
		va.UUIDs = append(va.UUIDs, trx.UUID)
		va.TotalAmount = va.TotalAmount.Add(amount)
	}
}

func (r *ReconVAStatic) Keys() []string {
	keys := make([]string, 0, len(*r))
	for key := range *r {
		keys = append(keys, key)
	}
	return keys
}

func (r *ReconVAStatic) GetIndexes(reference string) []int {
	if _, ok := (*r)[reference]; !ok {
		return []int{}
	}
	return (*r)[reference].Indexes
}
