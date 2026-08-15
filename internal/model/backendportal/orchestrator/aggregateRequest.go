package orchestrator_model

import (
	"time"

	"github.com/google/uuid"
)

type GetAggregateRequest struct {
	MerchantID uuid.UUID  `json:"merchantId"`
	AccountID  uuid.UUID  `json:"accountId"`
	AccountIDs []string   `json:"accountIds"`
	Statuses   []string   `json:"statuses"`
	StartAt    *time.Time `json:"startAt"`
	EndAt      *time.Time `json:"endAt"`
	// Include fee with deduction type manual and automated (pending status)
	IncludeFeeIndirectDeduction bool `json:"withoutFeeIndirectDeduction"`
	// This attribute is used internally to retrieve successful transactions that are pending settlement.
	PendingSettlementBalance bool `json:"-"`
}

type BulkGetAggregateRequest struct {
	AccountClauses              []AccountsAggregationClause
	Statuses                    []string
	IncludeFeeIndirectDeduction bool
}

type AccountsAggregationClause struct {
	MerchantID string
	AccountID  string
	StartAt    *time.Time
	EndAt      *time.Time
}

type GetWalletTotalBalanceRequest struct {
	MerchantID         string
	Status             []string
	IncludeIndirectFee bool
}

type GetSummaryTransactionByReferenceRequest struct {
	MerchantID       uuid.UUID
	ReferenceType    string
	ReferenceID      string
	Status           string
	SettlementStatus string
}

type GetPastDueSettlementTransactionsRequest struct {
	ReferenceID string
	Datetime    time.Time
}
