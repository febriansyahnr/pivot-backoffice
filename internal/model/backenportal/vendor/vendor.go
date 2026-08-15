package vendor

import (
	"database/sql"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/shopspring/decimal"
)

type Vendor struct {
	UUID                string          `json:"uuid" db:"uuid"`
	MerchantID          string          `json:"merchantId" db:"merchant_id"`
	Name                string          `json:"name" db:"name"`
	BeneficialOwner     string          `json:"beneficialOwner" db:"beneficial_owner"`
	BusinessCategory    string          `json:"businessCategory" db:"business_category"`
	AvgMonthlyTpvAmount decimal.Decimal `json:"avgMonthlyTpvAmount" db:"avg_monthly_tpv_amount"`
	BankName            string          `json:"bankName" db:"bank_name"`
	BankCode            string          `json:"bankCode" db:"bank_code"`
	AccountNumber       string          `json:"accountNumber" db:"account_number"`
	AccountName         string          `json:"accountName" db:"account_name"`
	Documents           types.JSONText  `json:"documents" db:"documents"`
	Status              string          `json:"status" db:"status"`
	CreatedAt           time.Time       `json:"createdAt" db:"created_at"`
	UpdatedAt           time.Time       `json:"updatedAt" db:"updated_at"`
	DeletedAt           sql.NullTime    `json:"deletedAt" db:"deleted_at"`
}

type VendorQuery struct {
	UUID       uuid.UUID  `json:"uuid"`
	MerchantID uuid.UUID  `json:"merchantId"`
	Name       string     `json:"name"`
	Status     string     `json:"status"`
	Page       int64      `json:"page"`
	PageSize   int64      `json:"pageSize"`
	SortBy     string     `json:"sortBy"`
	Sort       string     `json:"sort"`
	StartDate  *time.Time `json:"startDate"`
	EndDate    *time.Time `json:"endDate"`
}

type CreateVendorRequest struct {
	MerchantID          string                  `form:"merchantId" validate:"required"`
	Name                string                  `form:"name" validate:"required"`
	BeneficialOwner     string                  `form:"beneficialOwner" validate:"required"`
	BusinessCategory    string                  `form:"businessCategory" validate:"required"`
	AvgMonthlyTpvAmount decimal.Decimal         `form:"avgMonthlyTpvAmount" validate:"required"`
	BankName            string                  `form:"bankName" validate:"required"`
	BankCode            string                  `form:"bankCode" validate:"required"`
	AccountNumber       string                  `form:"accountNumber" validate:"required"`
	AccountName         string                  `form:"accountName" validate:"required"`
	Documents           types.JSONText          `form:"documents,omitempty"`
	DocumentFiles       []*multipart.FileHeader `form:"-"`
}

type UpdateVendorRequest struct {
	UUID                string                  `form:"uuid" validate:"required"`
	Name                *string                 `form:"name"`
	BeneficialOwner     *string                 `form:"beneficialOwner"`
	BusinessCategory    *string                 `form:"businessCategory"`
	AvgMonthlyTpvAmount *decimal.Decimal        `form:"avgMonthlyTpvAmount"`
	BankName            *string                 `form:"bankName"`
	BankCode            *string                 `form:"bankCode"`
	AccountNumber       *string                 `form:"accountNumber"`
	AccountName         *string                 `form:"accountName"`
	Documents           *types.JSONText         `form:"documents"`
	DocumentFiles       []*multipart.FileHeader `form:"-"`
}

type VendorResponse struct {
	UUID                string          `json:"uuid"`
	MerchantID          string          `json:"merchantId"`
	Name                string          `json:"name"`
	BeneficialOwner     string          `json:"beneficialOwner"`
	BusinessCategory    string          `json:"businessCategory"`
	AvgMonthlyTpvAmount decimal.Decimal `json:"avgMonthlyTpvAmount"`
	BankName            string          `json:"bankName"`
	BankCode            string          `json:"bankCode"`
	AccountNumber       string          `json:"accountNumber"`
	AccountName         string          `json:"accountName"`
	Documents           types.JSONText  `json:"documents,omitempty"`
	Status              string          `json:"status"`
	CreatedAt           time.Time       `json:"createdAt"`
	UpdatedAt           time.Time       `json:"updatedAt"`
}

func (q *VendorQuery) BuildCondition() (string, []any) {
	conditions := []string{}
	args := []any{}

	if q.UUID != uuid.Nil {
		conditions = append(conditions, "uuid = ?")
		args = append(args, q.UUID.String())
	}

	if q.MerchantID != uuid.Nil {
		conditions = append(conditions, "merchant_id = ?")
		args = append(args, q.MerchantID.String())
	}

	if q.Name != "" {
		conditions = append(conditions, "name LIKE ?")
		args = append(args, "%"+q.Name+"%")
	}

	if q.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, q.Status)
	}

	if q.StartDate != nil {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, *q.StartDate)
	}

	if q.EndDate != nil {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, *q.EndDate)
	}

	// Always exclude soft deleted records
	conditions = append(conditions, "deleted_at IS NULL")

	return strings.Join(conditions, " AND "), args
}

func (q *VendorQuery) BuildOrderBy() string {
	allowedSortBy := map[string]string{
		"createdAt": "created_at",
	}

	column, ok := allowedSortBy[q.SortBy]
	if !ok {
		return "created_at DESC"
	}

	sort := strings.ToUpper(q.Sort)
	if sort != "ASC" && sort != "DESC" {
		sort = "DESC"
	}

	return column + " " + sort
}

func (c *CreateVendorRequest) ToVendor() *Vendor {
	now := time.Now().UTC()
	return &Vendor{
		UUID:                util.GenerateUUID().String(),
		MerchantID:          c.MerchantID,
		Name:                c.Name,
		BeneficialOwner:     c.BeneficialOwner,
		BusinessCategory:    c.BusinessCategory,
		AvgMonthlyTpvAmount: c.AvgMonthlyTpvAmount,
		BankName:            c.BankName,
		BankCode:            c.BankCode,
		AccountNumber:       c.AccountNumber,
		AccountName:         c.AccountName,
		Documents:           c.Documents,
		Status:              constant.CardFundedPayoutStatusActive,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

func (v *Vendor) Update(req *UpdateVendorRequest) {
	if req.Name != nil {
		v.Name = *req.Name
	}
	if req.BeneficialOwner != nil {
		v.BeneficialOwner = *req.BeneficialOwner
	}
	if req.BusinessCategory != nil {
		v.BusinessCategory = *req.BusinessCategory
	}
	if req.AvgMonthlyTpvAmount != nil {
		v.AvgMonthlyTpvAmount = *req.AvgMonthlyTpvAmount
	}
	if req.BankName != nil {
		v.BankName = *req.BankName
	}
	if req.BankCode != nil {
		v.BankCode = *req.BankCode
	}
	if req.AccountNumber != nil {
		v.AccountNumber = *req.AccountNumber
	}
	if req.AccountName != nil {
		v.AccountName = *req.AccountName
	}
	if req.Documents != nil {
		v.Documents = *req.Documents
	}

	v.UpdatedAt = time.Now().UTC()
}

func (v *Vendor) ToResponse() *VendorResponse {
	return &VendorResponse{
		UUID:                v.UUID,
		MerchantID:          v.MerchantID,
		Name:                v.Name,
		BeneficialOwner:     v.BeneficialOwner,
		BusinessCategory:    v.BusinessCategory,
		AvgMonthlyTpvAmount: v.AvgMonthlyTpvAmount,
		BankName:            v.BankName,
		BankCode:            v.BankCode,
		AccountNumber:       v.AccountNumber,
		AccountName:         v.AccountName,
		Documents:           v.Documents,
		Status:              v.Status,
		CreatedAt:           v.CreatedAt,
		UpdatedAt:           v.UpdatedAt,
	}
}
