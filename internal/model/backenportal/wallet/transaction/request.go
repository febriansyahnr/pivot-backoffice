package walletTransactionModel

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type MerchantTransactionHistoryListReq struct {
	MerchantId   string    `json:"merchantId" validate:"required,uuid"`
	StartDateReq string    `json:"startDate" validate:"required_with=EndDate,iso_8601_datetime"`
	EndDateReq   string    `json:"endDate" validate:"required_with=StartDate,iso_8601_datetime,gtecsfield=StartDateReq"`
	Type         string    `json:"type" validate:"-"`
	Status       string    `json:"status" validate:"omitempty,oneof=PENDING SUCCESS FAILED"`
	Id           string    `json:"id" validate:"-"`
	ReferenceId  string    `json:"referenceId" validate:"-"` // Merchant Reference ID
	Page         int64     `json:"page" validate:"omitempty,min=1"`
	PerPage      int64     `json:"perPage" validate:"omitempty,min=1"`
	Sort         string    `json:"sort" validate:"required,oneof=-date date"`
	StartDate    time.Time `json:"-" validate:"-"`
	EndDate      time.Time `json:"-" validate:"-"`
}

func (r *MerchantTransactionHistoryListReq) HashFilter(timezone string) string {
	endDate := r.EndDate
	if time.Now().UTC().Before(r.EndDate) {
		endDate = time.Now().UTC()
	}

	buf := bytes.NewBufferString(
		r.MerchantId + "|" + r.StartDate.Format(time.DateTime) + "|" + endDate.Format(time.DateTime),
	)

	if r.Type != "" {
		_, _ = buf.WriteString("|" + r.Type)
	}
	if r.Status != "" {
		_, _ = buf.WriteString("|" + r.Status)
	}
	if r.Id != "" {
		_, _ = buf.WriteString("|" + r.Id)
	}
	_, _ = buf.WriteString("|" + r.Sort + "|" + timezone)

	hash := sha256.Sum256(buf.Bytes())
	return hex.EncodeToString(hash[:])
}
