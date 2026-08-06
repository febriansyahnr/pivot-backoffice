package recurringContractModel

import (
	"time"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
)

type CreateRecurringContractResponse struct {
	RecurringID string `json:"recurringId"`
	CustomerID  string `json:"customerId"`
}

type CancelRecurringContractResponse struct {
	RecurringID string `json:"recurringId"`
	Status      string `json:"status"`
}

type GetRecurringByIDRequest struct {
	RecurringID string `json:"-"`
	MerchantID  string `json:"-"`
}

type GetRecurringByIDDashboardResponse struct {
	UUID              string             `json:"uuid"`
	MerchantID        string             `json:"merchantId"`
	CustomerID        string             `json:"customerId"`
	ClientReferenceID string             `json:"clientReferenceId"`
	Plan              Plan               `json:"plan"`
	Trials            []Trial            `json:"trials"`
	Billing           Billing            `json:"billing"`
	Amount            commonModel.Amount `json:"amount"`
	Status            string             `json:"status"`
	UpdatedAt         time.Time          `json:"updatedAt"`
	CreatedAt         time.Time          `json:"createdAt"`
}
