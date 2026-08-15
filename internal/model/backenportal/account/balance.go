package account_model

import "github.com/google/uuid"

type GetBulkBalanceRequest struct {
	MerchantIDs []uuid.UUID
	Usecase     string
}

type AvailableBalanceResponse struct {
	Balance  float64
	Currency string
}
