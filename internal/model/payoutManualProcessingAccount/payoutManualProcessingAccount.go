package payoutManualProcessingAccount

import (
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
)

type PayoutManualProcessingAccount struct {
	UUID          string    `json:"uuid" db:"uuid"`
	MerchantID    string    `json:"merchantId" db:"merchant_id"`
	MerchantName  string    `json:"merchantName" db:"merchant_name"`
	BankCode      string    `json:"bankCode" db:"bank_code"`
	AccountNumber string    `json:"accountNumber" db:"account_number"`
	Status        string    `json:"status" db:"status"`
	UpdatedBy     string    `json:"updatedBy" db:"updated_by"`
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
}

type PayoutManualProcessingAccountQuery struct {
	MerchantID    uuid.UUID `json:"merchantId"`
	BankCode      string    `json:"bankCode"`
	AccountNumber string    `json:"accountNumber"`
	Status        string    `json:"status"`
	Page          int64     `json:"page"`
	PageSize      int64     `json:"pageSize"`
	SortBy        string    `json:"sortBy"`
	Sort          string    `json:"sort"`
}

type CreatePayoutManualProcessingAccountRequest struct {
	MerchantID    string `json:"merchantId" form:"merchantId" validate:"required"`
	BankCode      string `json:"bankCode" form:"bankCode" validate:"required"`
	AccountNumber string `json:"accountNumber" form:"accountNumber" validate:"required"`
	UpdatedBy     string `json:"updatedBy" form:"updatedBy" validate:"required"`
}

type UpdatePayoutManualProcessingAccountRequest struct {
	UUID      string  `json:"uuid" form:"uuid" validate:"required"`
	Status    *string `json:"status" form:"status"`
	UpdatedBy string  `json:"updatedBy" form:"updatedBy" validate:"required"`
}

type PayoutManualProcessingAccountResponse struct {
	UUID          string    `json:"uuid"`
	MerchantID    string    `json:"merchantId"`
	MerchantName  string    `json:"merchantName"`
	BankCode      string    `json:"bankCode"`
	AccountNumber string    `json:"accountNumber"`
	Status        string    `json:"status"`
	UpdatedBy     string    `json:"updatedBy"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (q *PayoutManualProcessingAccountQuery) BuildCondition() (string, []any) {
	conditions := []string{}
	args := []any{}

	if q.MerchantID != uuid.Nil {
		conditions = append(conditions, "a.merchant_id = ?")
		args = append(args, q.MerchantID.String())
	}

	if q.BankCode != "" {
		conditions = append(conditions, "a.bank_code = ?")
		args = append(args, q.BankCode)
	}

	if q.AccountNumber != "" {
		conditions = append(conditions, "a.account_number = ?")
		args = append(args, q.AccountNumber)
	}

	if q.Status != "" {
		conditions = append(conditions, "a.status = ?")
		args = append(args, q.Status)
	}

	return strings.Join(conditions, " AND "), args
}

func (q *PayoutManualProcessingAccountQuery) BuildOrderBy() string {
	allowedSortBy := map[string]string{
		"bankCode":      "bank_code",
		"accountNumber": "account_number",
		"status":        "status",
		"uuid":          "uuid",
	}

	column, ok := allowedSortBy[q.SortBy]
	if !ok {
		return "bank_code ASC"
	}

	sort := strings.ToUpper(q.Sort)
	if sort != "ASC" && sort != "DESC" {
		sort = "ASC"
	}

	return column + " " + sort
}

func (c *CreatePayoutManualProcessingAccountRequest) ToPayoutManualProcessingAccount() *PayoutManualProcessingAccount {
	return &PayoutManualProcessingAccount{
		UUID:          util.GenerateUUID().String(),
		MerchantID:    c.MerchantID,
		BankCode:      c.BankCode,
		AccountNumber: c.AccountNumber,
		Status:        constant.StatusActive,
		UpdatedBy:     c.UpdatedBy,
	}
}

func (a *PayoutManualProcessingAccount) Update(req *UpdatePayoutManualProcessingAccountRequest) {
	if req.Status != nil {
		a.Status = *req.Status
	}
	a.UpdatedBy = req.UpdatedBy
}

func (a *PayoutManualProcessingAccount) ToResponse() *PayoutManualProcessingAccountResponse {
	return &PayoutManualProcessingAccountResponse{
		UUID:          a.UUID,
		MerchantID:    a.MerchantID,
		MerchantName:  a.MerchantName,
		BankCode:      a.BankCode,
		AccountNumber: a.AccountNumber,
		Status:        a.Status,
		UpdatedBy:     a.UpdatedBy,
		UpdatedAt:     a.UpdatedAt,
	}
}
