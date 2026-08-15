package snapCoreModel

import (
	"time"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
)

type GenerateQrMpmRequest struct {
	PartnerReferenceNo string                 `json:"partnerReferenceNo"`
	Amount             commonModel.Amount     `json:"amount"`
	QrType             string                 `json:"qrType"`
	MerchantID         string                 `json:"merchantId"`
	SubMerchantID      string                 `json:"subMerchantId"`
	StoreID            string                 `json:"storeId"`
	TerminalID         string                 `json:"terminalID"`
	ValidityPeriod     int                    `json:"validityPeriod"`
	AdditionalInfo     map[string]interface{} `json:"additionalInfo"`
	TipType            string                 `json:"tipType"`
	Acquirer           string                 `json:"acquirer"`
}

type QueryQrMpmStaticRequest struct {
	PartnerReferenceNo string `json:"partnerReferenceNo" validate:"required"`
	FromDateTime       string `json:"fromDateTime"`
	ToDateTime         string `json:"toDateTime"`
	PageNumber         int    `json:"pageNumber"`
	PageSize           int    `json:"pageSize"`
}

func (q *QueryQrMpmStaticRequest) Validate() {
	// Validate dates
	q.FromDateTime = validateAndSetDate(q.FromDateTime, util.SnapDateFormatLayout, -3)
	q.ToDateTime = validateAndSetDate(q.ToDateTime, util.SnapDateFormatLayout, 0)

	// Validate page size
	if q.PageSize < 20 || q.PageSize > 100 {
		q.PageSize = 20
	}

	// Validate page number
	if q.PageNumber < 1 {
		q.PageNumber = 1
	}
}

func validateAndSetDate(dateStr string, layout string, months int) string {
	if dateStr == "" {
		return getDefaultDate(months)
	}

	parsedDate, err := time.Parse(layout, dateStr)
	if err != nil || parsedDate.After(time.Now()) {
		return getDefaultDate(months)
	}
	return dateStr
}

func getDefaultDate(months int) string {
	return util.ConvertToJakarta(time.Now().AddDate(0, months, 0)).Format(util.SnapDateFormatLayout)
}

type InquiryStatusQrMpmRequest struct {
	QrisUUID    string `json:"qrisUUID"`
	SkipPublish bool   `json:"skipPublish"`
}
