package paymentModel

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	commonPb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/common"
	settlementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/settlement"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	splitRoutingPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/splitRoutingPayment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/encryption"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/go/snap"
	snapVa "github.com/paper-indonesia/pdk/go/snap/structs/va"
	"github.com/shopspring/decimal"
)

type PaymentRequest struct {
	ReferenceID    string                         `json:"referenceId"`
	PaymentMethod  string                         `json:"paymentMethod" validate:"required"`
	TotalAmount    Amount                         `json:"totalAmount"`
	Customer       PaymentRequestCustomer         `json:"customer"`
	PaymentItems   *[]PaymentItemRequest          `json:"paymentItems"`
	VirtualAccount *PaymentMetadataVirtualAccount `json:"virtualAccount"`
	Qris           *PaymentMetadataQris           `json:"qris"`
	IsSnap         bool                           `json:"isSnap"`

	UUID              string                    `json:"-"`
	InitiateStatus    string                    `json:"-"`
	ClientRedirectUrl UnifiedPaymentRedirectUrl `json:"-"`
	PaymentUrl        string                    `json:"-"`
	IsUnifiedPayment  bool                      `json:"-"`
	CreatedBy         string                    `json:"-"`
	PaymentType       string                    `json:"-"`

	SplitRoutingConfigurations *[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration `json:"splitRoutingConfigurations,omitempty" validate:"omitempty,dive"`
}

type PaymentUpdateRequest struct {
	PaymentId      string                     `json:"-"`
	MerchantId     string                     `json:"merchantId"`
	TotalAmount    *Amount                    `json:"totalAmount"`
	ExpiredAt      *time.Time                 `json:"expiredAt"`
	AccountEmail   string                     `json:"accountEmail" example:"updateva@email.com"`
	AccountPhone   string                     `json:"accountPhone" example:"081239123"`
	AccountName    string                     `json:"accountName" example:"Tester VA"`
	BillDetails    []snapCoreModel.BillDetail `json:"billDetails"`
	AdditionalInfo map[string]interface{}     `json:"additionalInfo"`
	MaxAmount      *Amount                    `json:"maxAmount"`
	MinAmount      *Amount                    `json:"minAmount"`
}

type Amount struct {
	Value    decimal.Decimal `json:"value" validate:"min=0"`
	Currency string          `json:"currency" validate:"required"`
}

func (d *Amount) ProtoAmount() *commonPb.Amount {
	if d == nil {
		return nil
	}
	return &commonPb.Amount{
		Currency: d.Currency,
		Value:    d.Value.String(),
	}
}

type PaymentItemRequest struct {
	ItemID      string          `json:"itemId" `
	Name        string          `json:"name"`
	Description string          `json:"description" `
	Qty         int             `json:"qty"`
	Amount      Amount          `json:"amount"`
	Metadata    *map[string]any `json:"metadata,omitempty"`
}

type PaymentRequestCustomer struct {
	CustomerID string          `json:"customerId"`
	Name       string          `json:"name"`
	Email      string          `json:"email"`
	Phone      string          `json:"phone"`
	Metadata   *map[string]any `json:"metadata"`
}

type PaymentMetadataVirtualAccount struct {
	Issuer                string     `json:"issuer" validate:"required"`
	VirtualAccountTrxType string     `json:"virtualAccountTrxType" validate:"required"`
	VirtualAccountNumber  string     `json:"virtualAccountNumber"`
	VirtualAccountName    string     `json:"virtualAccountName"`
	MinAmount             *Amount    `json:"minAmount,omitempty"`
	MaxAmount             *Amount    `json:"maxAmount,omitempty"`
	ExpiredDate           *time.Time `json:"expiredDate"`

	// common field of snap
	BillDetails *[]snapVa.BillDetail `json:"billDetails"`
}

type VirtualAccountPaymentNotificationRequest struct {
	Acquirer               string             `json:"acquirer" validate:"required"`
	Number                 string             `json:"vaNumber" validate:"required"`
	Status                 string             `json:"status" validate:"required"`
	PaidAmount             commonModel.Amount `json:"paidAmount"`
	ExpiredAt              *time.Time         `json:"expiredAt"`
	Processor              string             `json:"processor"`
	ProcessorID            string             `json:"processorID"`
	ProcessorTransactionID string             `json:"processorTransactionID"`
	TrxDatetime            time.Time          `json:"trxDatetime"`
	AdditionalData         string             `json:"additionalData"`
	BankReferenceId        string             `json:"bankReferenceId"`
}

type GetPaymentMethodFilterRequest struct {
	MerchantID      string                             `json:"-"`
	Payment         *PaymentDetailForPaymentUIResponse `json:"-"`
	Category        string                             `json:"category"`
	Type            string                             `json:"type"`
	Subtype         string                             `json:"subtype"`
	Acquirer        string                             `json:"acquirer"`
	Status          string                             `json:"status" validate:"oneof=ACTIVE INACTIVE"`
	InstallmentPlan InstallmentPlanFilterRequest
}

type InstallmentPlanFilterRequest struct {
	InstallmentPlanID string
}

type GetActivePaymentMethodRequest = GetPaymentMethodFilterRequest

type UpdatePaymentMethodStatusRequest struct {
	Status *bool `json:"status" validate:"required"`
}

type QrisPaymentNotificationRequest struct {
	Acquirer    string             `json:"acquirer" validate:"required"`
	ReferenceNo string             `json:"referenceNo" validate:"required"`
	Status      string             `json:"status" validate:"required"`
	PaidAmount  commonModel.Amount `json:"paidAmount"`
	ExpiredAt   *time.Time         `json:"expiredAt"`

	Processor              string    `json:"processor"`
	ProcessorID            string    `json:"processorID"`
	ProcessorTransactionID string    `json:"processorTransactionID"`
	TransactionTime        time.Time `json:"transactionTime"`
}

type PaymentMetadataQris struct {
	QrType         string  `json:"qrType" validate:"required" example:"DYNAMIC or STATIC"`
	QrMethodType   string  `json:"qrMethodType" validate:"required" example:"MPM or CPM"`
	SubMerchantId  string  `json:"subMerchantId"`
	Amount         *Amount `json:"amount"`
	ValidityPeriod int     `json:"validityPeriod"`
}

type PaymentSimulationMetadataVA struct {
	Uuid           string    `json:"uuid"`
	Number         string    `json:"number"`
	Acquirer       string    `json:"acquirer"`
	AccountName    string    `json:"accountName"`
	IsSingleUse    bool      `json:"isSingleUse"`
	IsClosedAmount bool      `json:"isClosedAmount"`
	TotalAmount    Amount    `json:"totalAmount"`
	CreatedAt      string    `json:"createdAt"`
	ExpiredAt      time.Time `json:"expiredAt"`
}

type PaymentSnapQrisExpiredCallbackRequest struct {
	*PaymentQrisCallbackRequest
	AdditionalInfo map[string]any `json:"additionalInfo"`
}

type PaymentQrisCallbackRequest struct {
	OriginalReferenceNo        string                            `json:"originalReferenceNo"`
	OriginalPartnerReferenceNo string                            `json:"originalPartnerReferenceNo"`
	LatestTransactionStatus    string                            `json:"latestTransactionStatus"`
	TransactionStatusDesc      string                            `json:"transactionStatusDesc"`
	Amount                     commonModel.Amount                `json:"amount"`
	AdditionalInfo             PaymentQrisCallbackAdditionalInfo `json:"additionalInfo"`
}

type PaymentQrisCallbackAdditionalInfo struct {
	RRN             string `json:"RRN"`
	QrType          string `json:"qrType"`
	QrStatus        string `json:"qrStatus"`
	QrExpiredDate   string `json:"qrExpiredDate"`
	MerchantName    string `json:"merchantName"`
	PaymentStatus   string `json:"paymentStatus"`
	TransactionDate string `json:"transactionDate"`
}

type ProcessPaymentSimulation struct {
	PaidAmount  commonModel.Amount `json:"paidAmount" validate:"required"`
	PhoneNumber string             `json:"phoneNumber,omitempty" validate:"omitempty"`
	OTP         string             `json:"otp,omitempty" validate:"omitempty"`
}

type QueryQrMpmStaticRequest struct {
	ReferenceId  string `json:"referenceId" validate:"required"`
	FromDateTime string `json:"fromDateTime"`
	ToDateTime   string `json:"toDateTime"`
	PageNumber   int    `json:"pageNumber"`
	PageSize     int    `json:"pageSize"`
}

type SnapQueryQrMpmDynamicRequest struct {
	OriginalReferenceNo        string `json:"originalReferenceNo"`
	OriginalPartnerReferenceNo string `json:"originalPartnerReferenceNo"`
	ServiceCode                string `json:"serviceCode"`
}

func (q *QueryQrMpmStaticRequest) Validate() {
	// Validate page size
	if q.PageSize < 20 || q.PageSize > 100 {
		q.PageSize = 20
	}

	// Validate page number
	if q.PageNumber < 1 {
		q.PageNumber = 1
	}
}

type SNAPQueryQrMpmStaticReq struct {
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	FromDateTime       string `json:"fromDateTime"`
	ToDateTime         string `json:"toDateTime"`
	PageSize           int    `json:"pageSize"`
	PageNumber         int    `json:"pageNumber"`
}

type SnapGenerateQrMpmRequest struct {
	PartnerReferenceNo string                `json:"partnerReferenceNo,omitempty"`
	SubMerchantId      string                `json:"subMerchantId,omitempty"`
	Amount             *Amount               `json:"amount,omitempty"`
	ValidityPeriod     string                `json:"validityPeriod,omitempty"`
	AdditionalInfo     *SnapQrAdditionalInfo `json:"additionalInfo,omitempty"`
}

type SnapVACreateData struct {
	snapVa.VACommonField
	VirtualAccountTrxType string               `json:"virtualAccountTrxType"`
	ExpiredDate           string               `json:"expiredDate,omitempty"`
	TotalAmount           *snapVa.Amount       `json:"totalAmount,omitempty"`
	LastUpdateDate        string               `json:"lastUpdateDate,omitempty"`
	PaymentDate           string               `json:"paymentDate,omitempty"`
	AdditionalInfo        SnapVAAdditionalInfo `json:"additionalInfo"`
}

type SnapVaUpdateRequest struct {
	TrxID                 string               `json:"trxId"`
	VirtualAccountEmail   string               `json:"virtualAccountEmail"`
	VirtualAccountPhone   string               `json:"virtualAccountPhone"`
	VirtualAccountName    string               `json:"virtualAccountName"`
	ExpiredDate           string               `json:"expiredDate"`
	TotalAmount           *snapVa.Amount       `json:"totalAmount,omitempty"`
	BillDetails           *[]snapVa.BillDetail `json:"billDetails"`
	AdditionalInfo        SnapVAAdditionalInfo `json:"additionalInfo"`
	VirtualAccountTrxType string               `json:"_"`
}

type SnapVAAdditionalInfo struct {
	ReferenceID   string                  `json:"referenceId"`
	Issuer        string                  `json:"issuer"`
	MinAmount     *snapVa.Amount          `json:"minAmount,omitempty"`
	MaxAmount     *snapVa.Amount          `json:"maxAmount,omitempty"`
	VaStatus      string                  `json:"vaStatus,omitempty"`
	PaymentStatus string                  `json:"paymentStatus,omitempty"`
	Customer      *PaymentRequestCustomer `json:"customer,omitempty"`
	Metadata      map[string]any          `json:"metadata,omitempty"`
}

type PaymentMethodVirtualAccount struct {
	Issuer                string         `json:"issuer"`
	VirtualAccountTrxType string         `json:"virtualAccountTrxType"`
	VirtualAccountName    string         `json:"virtualAccountName"`
	VirtualAccountNumber  string         `json:"virtualAccountNumber"`
	MinAmount             *Amount        `json:"minAmount,omitempty"`
	MaxAmount             *Amount        `json:"maxAmount,omitempty"`
	ExpiredDate           *time.Time     `json:"expiredDate"`
	Metadata              map[string]any `json:"metadata,omitempty"`
}

type SnapVaGetRequest struct {
	TrxID            string `json:"trxId"`
	VirtualAccountNo string `json:"virtualAccountNo"`
}

type FilterPaymentHistoryOption struct {
	MerchantID       string
	ReferenceID      string
	Status           string
	PaymentMethod    string
	SettlementModel  string
	StartDate        time.Time // payment creation date
	EndDate          time.Time // payment creation date
	PaymentStartDate time.Time // payment date from user
	PaymentEndDate   time.Time // payment date from user
	Sort             string    // Date, Amount, Paid Amount,
	SortBy           string
	Page             int
	PerPage          int
}

type GetInvestigationSummaryOption struct {
	MerchantID string
	StartDate  time.Time
	EndDate    time.Time
}

func (r *FilterPaymentHistoryOption) Validate() error {
	if r.PaymentMethod != "" {
		if err := ValidatePaymentMethod(r.PaymentMethod); err != nil {
			return err
		}
	}

	if r.Status != "" {
		statuses := strings.Split(r.Status, ",")
		for _, status := range statuses {
			if err := ValidateUnifiedPaymentStatus(status); err != nil {
				return err
			}
		}
	}

	if r.SortBy == "" {
		r.SortBy = DefaultSortColumn
	}

	if err := ValidatePaymentHistorySortColumn(r.SortBy); err != nil {
		return err
	}

	return nil
}

type PaymentHistoryDetailOption struct {
	PaymentID  string
	MerchantID string
}

type PaymentInsightOption struct {
	MerchantID string
	Status     string
}

type GetPaymentDashboardInsightRequest struct {
	StartDate  time.Time
	EndDate    time.Time
	MerchantId string
}

type DateRange struct {
	StartDate string // Format date only
	EndDate   string // Format date only
}

type QueryPaymentSuccessRateComparisonRequest struct {
	MerchantId   string
	CurrentRange DateRange
	PrevRange    DateRange
}

func (d *SNAPQueryQrMpmStaticReq) ValidateAndSetValue() error {
	if d.PartnerReferenceNo == "" {
		return pkgErrs.New(response.SnapErrRequiredField, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "partnerReferenceNo"))
	}

	// Validate and set FromDateTime and ToDateTime with error handling
	var err error
	d.FromDateTime, err = validateAndSetDateWithError(d.FromDateTime, util.SnapDateFormatLayout, true, "fromDateTime")
	if err != nil {
		return err
	}

	d.ToDateTime, err = validateAndSetDateWithError(d.ToDateTime, util.SnapDateFormatLayout, false, "toDateTime")
	if err != nil {
		return err
	}

	// Check if FromDateTime is not after ToDateTime
	fromDateTimeParsed, _ := time.Parse(util.SnapDateFormatLayout, d.FromDateTime)
	toDateTimeParsed, _ := time.Parse(util.SnapDateFormatLayout, d.ToDateTime)

	if fromDateTimeParsed.After(toDateTimeParsed) {
		return pkgErrs.New(response.SnapErrFieldFormat, fmt.Errorf(constant.InvalidFieldFormatSnapFmt, "fromDateTime"))
	}

	if d.PageSize < 1 {
		d.PageSize = 20
	} else if d.PageSize > 100 {
		d.PageSize = 100
	}

	if d.PageNumber < 1 {
		d.PageNumber = 1
	}
	return nil
}

func validateAndSetDateWithError(dateStr string, layout string, isFromDate bool, fieldName string) (string, error) {
	currentTime := util.ConvertToJakarta(time.Now().UTC())
	if dateStr == "" {
		if isFromDate {
			// Set FromDateTime to start of the current day (00:00)
			startOfDay := time.Date(
				currentTime.Year(),
				currentTime.Month(),
				currentTime.Day(),
				0, 0, 0, 0,
				currentTime.Location())
			return startOfDay.Format(layout), nil
		}
		// Set ToDateTime to the current time
		return currentTime.Format(layout), nil
	}

	// Parse date and check if it complies with SnapDateFormatLayout
	parsedTime, err := time.ParseInLocation(layout, dateStr, currentTime.Location())
	if err != nil || parsedTime.After(time.Now()) {
		// Return error if the date is invalid
		return "", pkgErrs.New(response.SnapErrFieldFormat, fmt.Errorf(constant.InvalidFieldFormatSnapFmt, fieldName))
	}

	return dateStr, nil
}

func (r *SnapQueryQrMpmDynamicRequest) ValidateAndSetValue() error {

	if r.OriginalReferenceNo == "" && r.OriginalPartnerReferenceNo == "" {
		return pkgErrs.New(response.SnapErrRequiredField, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "OriginalPartnerReferenceNo"))
	}

	if r.ServiceCode == "" {
		return pkgErrs.New(response.SnapErrRequiredField, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "serviceCode"))
	}

	if r.ServiceCode != constant.GenerateQrisMPMSnapApiCode {
		return pkgErrs.New(response.SnapErrFieldFormat, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "serviceCode"))
	}

	return nil
}

func (r *SnapGenerateQrMpmRequest) ValidateAndSetValue() error {
	fmt.Printf("SnapGenerateQrMpmRequest: %+v\n", r.AdditionalInfo.QrType)
	if r.AdditionalInfo == nil {
		return pkgErrs.New(response.SnapErrFieldFormat, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "additionalInfo"))
	}

	switch r.AdditionalInfo.QrType {
	case constant.QrTypeStatic:
		if r.Amount != nil { // return error when fill amount
			return pkgErrs.New(response.SnapErrFieldFormat, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "amount"))
		}
		if r.ValidityPeriod != "" { // return error when fill validity period
			return pkgErrs.New(response.SnapErrFieldFormat, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "validityPeriod"))
		}
	case constant.QrTypeDynamic:
		if r.PartnerReferenceNo == "" { // return error when partner reference no
			return pkgErrs.New(response.SnapErrRequiredField, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "partnerReferenceNo"))
		}

		if r.Amount == nil { // return error when empty amount
			return pkgErrs.New(response.SnapErrRequiredField, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "amount"))
		}

		if r.Amount != nil && r.Amount.Value.Cmp(decimal.NewFromInt(constant.SnapQrisTypeDynamicMinAmount)) < 0 || r.Amount.Value.Cmp(decimal.NewFromInt(constant.SnapQrisTypeDynamicMaxAmount)) > 0 { // if amount exist but amount is invalid, return error
			return pkgErrs.New(response.SnapErrInvalidAmount, errors.New(constant.InvalidAmountSnapErrMsg))
		}

		if r.Amount != nil && r.Amount.Currency != "IDR" { // if amount exist but currency is invalid, return error
			return pkgErrs.New(response.SnapErrFieldFormat, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "currency"))
		}

		// Convert validity period from string to int and return error when invalid validity period
		validityPeriodInt, err := strconv.Atoi(r.ValidityPeriod)
		if (r.ValidityPeriod != "" && err != nil) || validityPeriodInt < 0 || validityPeriodInt > constant.QrisDynamicValidityPeriodMax {
			return pkgErrs.New(response.SnapErrFieldFormat, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "validityPeriod"))
		}

		// if validity period empty then set default validity period
		if r.ValidityPeriod == "" || validityPeriodInt == 0 {
			r.ValidityPeriod = "3600"
		}

	default:
		return nil
	}

	return nil
}

func (r *SnapVACreateData) ValidateAndSetValue() (minAmount, maxAmount *Amount, totalAmount Amount, paymentItems *[]PaymentItemRequest, expiredDate *time.Time, err error) {
	// Validate virtualAccountTrxType
	if r.VirtualAccountTrxType == "" {
		return nil, nil, Amount{}, nil, nil, pkgErrs.New(response.SnapErrRequiredField, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "virtualAccountTrxType"))
	}

	validTrxTypes := map[string]bool{"CLOSED_DYNAMIC": true, "OPEN_STATIC": true, "CLOSED_STATIC": true}
	if !validTrxTypes[r.VirtualAccountTrxType] {
		return nil, nil, Amount{}, nil, nil, pkgErrs.New(response.SnapErrInvalidVA, errors.New(constant.InvalidBillVirtualAccountErrMsg))
	}

	// Validate virtualAccountNo for OPEN_STATIC and CLOSED_STATIC
	if r.VirtualAccountTrxType == "OPEN_STATIC" || r.VirtualAccountTrxType == "CLOSED_STATIC" {
		if r.VirtualAccountNo == "" {
			return nil, nil, Amount{}, nil, nil, pkgErrs.New(response.SnapErrRequiredField, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "virtualAccountNo"))
		}
		match, _ := regexp.MatchString(`^[0-9]{4,8}$`, r.VirtualAccountNo)
		if !match {
			return nil, nil, Amount{}, nil, nil, pkgErrs.New(response.SnapErrFieldFormat, fmt.Errorf(constant.InvalidFieldFormatSnapFmt, "virtualAccountNo"))
		}
	}

	// Validate virtualAccountName
	if r.VirtualAccountName == "" {
		return nil, nil, Amount{}, nil, nil, pkgErrs.New(response.SnapErrRequiredField, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "virtualAccountName"))
	}

	// Validate expiredDate for CLOSED_DYNAMIC
	if r.VirtualAccountTrxType == "CLOSED_DYNAMIC" && r.ExpiredDate == "" {
		return nil, nil, Amount{}, nil, nil, pkgErrs.New(response.SnapErrRequiredField, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "expiredDate"))
	}

	// Validate totalAmount for CLOSED_DYNAMIC and CLOSED_STATIC
	if r.VirtualAccountTrxType == "CLOSED_DYNAMIC" || r.VirtualAccountTrxType == "CLOSED_STATIC" {
		if r.TotalAmount == nil {
			return nil, nil, Amount{}, nil, nil, pkgErrs.New(response.SnapErrRequiredField, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "totalAmount"))
		}
		if err = validateAmountField(r.TotalAmount, "totalAmount"); err != nil {
			return nil, nil, Amount{}, nil, nil, err
		}
	}

	// Validate additionalInfo
	if r.AdditionalInfo.ReferenceID == "" {
		return nil, nil, Amount{}, nil, nil, pkgErrs.New(response.SnapErrRequiredField, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "additionalInfo.referenceId"))
	}

	if r.AdditionalInfo.Issuer == "" {
		return nil, nil, Amount{}, nil, nil, pkgErrs.New(response.SnapErrRequiredField, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "additionalInfo.issuer"))
	}

	var defaultCustomer PaymentRequestCustomer
	if r.AdditionalInfo.Customer == nil {
		r.AdditionalInfo.Customer = &defaultCustomer
	}

	// Validate and clear minAmount and maxAmount for OPEN_STATIC
	if r.VirtualAccountTrxType != "OPEN_STATIC" {
		r.AdditionalInfo.MinAmount = nil
		r.AdditionalInfo.MaxAmount = nil
	}

	if minAmount, maxAmount, err = validateAndSetAmounts(r.AdditionalInfo.MinAmount, r.AdditionalInfo.MaxAmount, "additionalInfo.minAmount", "additionalInfo.maxAmount"); err != nil {
		return nil, nil, Amount{}, nil, nil, err
	}

	// Parse expiredDate if present
	if expiredDate, err = parseExpiredDate(r.ExpiredDate); err != nil {
		return nil, nil, Amount{}, nil, nil, err
	}

	// Set paymentItems and totalAmount
	if r.TotalAmount != nil {
		totalAmount = Amount{
			Value:    r.TotalAmount.Decimal(),
			Currency: r.TotalAmount.Currency,
		}
		paymentItems = &[]PaymentItemRequest{
			{Amount: totalAmount},
		}
	} else if r.BillDetails == nil {
		totalAmount = Amount{
			Value:    decimal.NewFromInt(0),
			Currency: "IDR",
		}
	}

	return minAmount, maxAmount, totalAmount, paymentItems, expiredDate, nil
}

func (r *SnapVaUpdateRequest) ValidateAndSetValue() (minAmount *Amount, maxAmount *Amount, totalAmount Amount, expiredDate *time.Time, err error) {

	// Validate VirtualAccountName
	if r.VirtualAccountName == "" {
		return nil, nil, Amount{}, nil, pkgErrs.New(response.SnapErrRequiredField, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "virtualAccountName"))
	}

	// Validate TrxID
	if r.TrxID == "" {
		return nil, nil, Amount{}, nil, pkgErrs.New(response.SnapErrRequiredField, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "trxId"))
	}

	// Validate TotalAmount and Currency for certain VA types
	if r.VirtualAccountTrxType == "CLOSED_DYNAMIC" || r.VirtualAccountTrxType == "CLOSED_STATIC" {
		fmt.Printf("r.TotalAmount: %+v\n", r.TotalAmount)
		if r.TotalAmount == nil {

			return nil, nil, Amount{}, nil, pkgErrs.New(response.SnapErrRequiredField, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "totalAmount"))
		}

		if !validateAmountFormat(r.TotalAmount.Value) {
			return nil, nil, Amount{}, nil, pkgErrs.New(response.SnapErrInvalidAmount, errors.New(constant.InvalidAmountSnapErrMsg))
		}

		if r.TotalAmount.Currency != "IDR" {
			return nil, nil, Amount{}, nil, pkgErrs.New(response.SnapErrFieldFormat, fmt.Errorf(constant.InvalidFieldFormatSnapFmt, "totalAmount.currency"))
		}
	}

	// Validate ExpiredDate for CLOSED_DYNAMIC VA type
	if r.VirtualAccountTrxType == "CLOSED_DYNAMIC" && r.ExpiredDate == "" {
		return nil, nil, Amount{}, nil, pkgErrs.New(response.SnapErrRequiredField, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "expiredDate"))
	}

	if minAmount, maxAmount, err = validateAndSetAmounts(r.AdditionalInfo.MinAmount, r.AdditionalInfo.MaxAmount, "additionalInfo.minAmount", "additionalInfo.maxAmount"); err != nil {
		return nil, nil, Amount{}, nil, err
	}

	// Parse ExpiredDate if present
	if expiredDate, err = parseExpiredDate(r.ExpiredDate); err != nil {
		return nil, nil, Amount{}, nil, err
	}

	// Set TotalAmount and PaymentItems
	if r.TotalAmount != nil {
		totalAmount = Amount{
			Value:    r.TotalAmount.Decimal(),
			Currency: r.TotalAmount.Currency,
		}
	} else {
		totalAmount = Amount{
			Value:    decimal.NewFromInt(0),
			Currency: "IDR",
		}
	}

	return minAmount, maxAmount, totalAmount, expiredDate, nil
}

func parseExpiredDate(expiredDateStr string) (*time.Time, error) {
	if expiredDateStr == "" {
		return nil, nil
	}

	var parsedDate time.Time
	var err error
	parsedDate, err = time.Parse(time.RFC3339, expiredDateStr)
	if err != nil || parsedDate.IsZero() {
		parsedDate, err = time.Parse(snap.SnapDateFormatLayout, expiredDateStr)
		if err != nil {
			return nil, pkgErrs.New(response.SnapErrFieldFormat, fmt.Errorf(constant.InvalidFieldFormatSnapFmt, "expiredDate"))
		}
	}
	return &parsedDate, nil
}

func validateAndSetAmounts(minAmountInfo, maxAmountInfo *snapVa.Amount, minAmountField, maxAmountField string) (minAmount *Amount, maxAmount *Amount, err error) {
	if minAmountInfo != nil {
		if err = validateAmountField(minAmountInfo, minAmountField); err != nil {
			return nil, nil, err
		}
		minAmount = &Amount{
			Value:    minAmountInfo.Decimal(),
			Currency: minAmountInfo.Currency,
		}
	}

	if maxAmountInfo != nil {
		if err = validateAmountField(maxAmountInfo, maxAmountField); err != nil {
			return nil, nil, err
		}
		maxAmount = &Amount{
			Value:    maxAmountInfo.Decimal(),
			Currency: maxAmountInfo.Currency,
		}
	}

	return minAmount, maxAmount, nil
}

func validateAmountFormat(amount string) bool {
	match, _ := regexp.MatchString(`^\d+(\.\d{0,2})?$`, amount)
	return match
}

func validateAmountField(amount *snapVa.Amount, fieldName string) error {
	if amount.Value == "" {
		return pkgErrs.New(response.SnapErrRequiredField, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, fieldName+".value"))
	}

	if !validateAmountFormat(amount.Value) {
		return pkgErrs.New(response.SnapErrInvalidAmount, errors.New(constant.InvalidAmountSnapErrMsg))
	}

	if amount.Currency != "IDR" {
		return pkgErrs.New(response.SnapErrFieldFormat, fmt.Errorf(constant.InvalidFieldFormatSnapFmt, fieldName+".currency"))
	}

	return nil
}

func (r *SnapVaGetRequest) ValidateAndSetValue() error {
	if r.TrxID == "" {
		return pkgErrs.New(response.SnapErrRequiredField, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "trxId"))
	}

	if r.VirtualAccountNo == "" {
		return pkgErrs.New(response.SnapErrRequiredField, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, "virtualAccountNo"))
	}

	return nil
}

type PostCreateFeeTransactionRequest struct {
	SettlementTransactionMetadata         *settlementModel.AccountTransactionMetadataObject
	FeeTransactionID, LinkedTransactionID uuid.UUID
	Status, Channel, Currency             string
	TransactionAmount                     float64
	SettlementStatus                      *string
	SettlementAt                          *time.Time
	SettlementModel                       *string
}

type CreateUnifiedPaymentRequest struct {
	PaymentID  string `json:"-"`
	MerchantID string `json:"-"`
	PaymentURL string `json:"-"`
	CreatedBy  string `json:"-"`

	ClientReferenceID string                    `json:"clientReferenceId" validate:"required"`
	Amount            Amount                    `json:"amount" validate:"required"`
	PaymentMethod     string                    `json:"paymentMethod" validate:"required,oneof=VIRTUAL_ACCOUNT QRIS CREDIT_CARD"`
	RedirectUrl       UnifiedPaymentRedirectUrl `json:"redirectUrl"`
	Customer          PaymentRequestCustomer    `json:"customer" validate:"required"`
	ExpiryAt          time.Time                 `json:"expiryAt" validate:"required"`

	PaymentMethodOptions       *UnifiedPaymentMethodOption                                  `json:"paymentMethodOptions" validate:"required_unless=PaymentMethod QRIS"`
	PaymentItems               *[]PaymentItemRequest                                        `json:"paymentItems"`
	SplitRoutingConfigurations *[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration `json:"splitRoutingConfigurations,omitempty" validate:"omitempty,dive"`
}

type UnifiedPaymentMethodOption struct {
	VirtualAccount *UnifiedPaymentMethodOptionVirtualAccount `json:"virtualAccount,omitempty"`
	Card           *UnifiedPaymentMethodOptionCard           `json:"card,omitempty"`
}

type UnifiedPaymentMethodOptionVirtualAccount struct {
	Issuer string `json:"issuer" validate:"required"`
}

type UnifiedPaymentMethodOptionCard struct {
	AuthenticationMethod string `json:"authenticationMethod" validate:"required"`
	BankMerchantID       string `json:"bankMerchantId"`
}

type UnifiedPaymentRedirectUrl struct {
	SuccessUrl string `json:"successUrl"`
	FailedUrl  string `json:"failedUrl"`
	ExpiredUrl string `json:"expiredUrl,omitempty"`
}

func (p *CreateUnifiedPaymentRequest) ToCreateUnifiedPaymentSessionRequest() *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest {
	request := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
		ClientReferenceID: p.ClientReferenceID,
		Amount: unifiedPaymentModel.Amount{
			Currency: p.Amount.Currency,
			Value:    p.Amount.Value.InexactFloat64(),
		},
		AutoConfirm:         true,
		StatementDescriptor: "",
		ExpiryAt:            p.ExpiryAt,
		Mode:                constant.UnifiedPaymentModeRedirect,
		RedirectUrl: unifiedPaymentModel.RedirectUrl{
			SuccessReturnUrl:    p.RedirectUrl.SuccessUrl,
			FailureReturnUrl:    p.RedirectUrl.FailedUrl,
			ExpirationReturnUrl: "",
		},
		PaymentMethod: &unifiedPaymentModel.PaymentMethod{
			Type: constant.MapToUnifiedPaymentMethod(p.PaymentMethod),
		},
		SplitRoutingConfigurations: p.SplitRoutingConfigurations,
		CustomerInformation: &unifiedPaymentModel.CustomerInformation{
			GivenName: p.Customer.Name,
			Surname:   &p.Customer.Name,
			SureName:  p.Customer.Name,
			Email:     p.Customer.Email,
		},
		MerchantID:        p.MerchantID,
		CreatedBy:         p.CreatedBy,
		IsMigratingFromV1: true,
	}

	if p.PaymentMethodOptions != nil {
		if p.PaymentMethodOptions.VirtualAccount != nil {
			request.PaymentMethodOptions.VirtualAccount = &unifiedPaymentModel.PaymentMethodOptionVirtualAccount{
				Channel: p.PaymentMethodOptions.VirtualAccount.Issuer,
			}
		}

		if p.PaymentMethodOptions.Card != nil {
			request.PaymentMethodOptions.Card = &unifiedPaymentModel.PaymentMethodOptionCard{}
		}
	}

	return request
}

type UpdateUnifiedPaymentRequest struct {
	MerchantID string `json:"-"`

	ClientReferenceID string                 `json:"clientReferenceId" validate:"required"`
	Amount            *Amount                `json:"amount"`
	PaymentMethod     string                 `json:"paymentMethod" validate:"required,oneof=VIRTUAL_ACCOUNT QRIS CREDIT_CARD"`
	ExpiryAt          *time.Time             `json:"expiryAt" validate:"required"`
	Customer          PaymentRequestCustomer `json:"customer"`
	// For Change Payment Method
	PaymentMethodOptions *UnifiedPaymentMethodOption `json:"paymentMethodOptions" validate:"-"` // Manual validation
}

var loc, _ = time.LoadLocation(constant.TimeLoc)

type PaymentDownloadHistoryRequest struct {
	MerchantId       string   `json:"-" validate:"required,uuid"`
	ReferenceId      string   `json:"referenceId" validate:"-"`
	Status           []string `json:"status"`
	PaymentMethod    string   `json:"paymentMethod" validate:"omitempty,oneof=QRIS VIRTUAL_ACCOUNT CREDIT_CARD"`
	StartDate        string   `json:"startDate" validate:"required,datetime=2006-01-02"`
	EndDate          string   `json:"endDate" validate:"required,datetime=2006-01-02"`
	Sort             string   `json:"sort" validate:"omitempty,oneof=ASC DESC"` // Date, Amount, Paid Amount,
	SortBy           string   `json:"sortBy" validate:"omitempty,oneof=createdAt amount amountPaid"`
	PaymentStartDate string   `json:"paymentStartDate" validate:"omitempty,datetime=2006-01-02T15:04:05.999Z"`
	PaymentEndDate   string   `json:"paymentEndDate" validate:"omitempty,datetime=2006-01-02T15:04:05.999Z"`
}

// Tech Debt:
// If the date parameters on the payment history page have been adjusted. Then this section also needs to be adjusted.
func (d *PaymentDownloadHistoryRequest) ToFilterPaymentHistoryOption() FilterPaymentHistoryOption {
	filter := FilterPaymentHistoryOption{
		MerchantID:    d.MerchantId,
		ReferenceID:   d.ReferenceId,
		Status:        strings.Join(d.Status, ","),
		PaymentMethod: d.PaymentMethod,
		Sort:          d.Sort,
		SortBy:        d.SortBy,
		Page:          1, PerPage: 1_048_576,
	}

	// NOTE:
	// In the development process, the query condition uses CONVERT_TZ from +00:00 to +07:00.
	// This would be tech debt if the db query had been adjusted.
	filter.StartDate, _ = time.ParseInLocation(time.DateTime, d.StartDate+" 00:00:00", loc)
	filter.EndDate, _ = time.ParseInLocation(time.DateTime, d.EndDate+" 23:59:59", loc)
	filter.PaymentStartDate, _ = time.ParseInLocation(time.DateTime, d.PaymentStartDate, loc)
	filter.PaymentEndDate, _ = time.ParseInLocation(time.DateTime, d.PaymentEndDate, loc)

	// Set Default Value
	if filter.Sort == "" {
		filter.Sort = "ASC"
	}
	if filter.SortBy == "" {
		filter.SortBy = "createdAt"
	}
	return filter
}

// Tech Debt:
// If the date parameters on the payment history page have been adjusted. Then this section also needs to be adjusted.
func (d *PaymentDownloadHistoryRequest) HashFilterKey(endDate time.Time) string {

	r := d.ToFilterPaymentHistoryOption()

	sha := sha256.New()
	_, _ = sha.Write([]byte(fmt.Sprintf(
		"%s:%v:%v:%s:%s:%s:%s", r.MerchantID, r.StartDate.In(time.UTC), endDate, r.Status, r.PaymentMethod, r.Sort, r.SortBy,
	)))

	hash := sha.Sum(nil)
	return hex.EncodeToString(hash)
}

type UnifiedPaymentCallbackRequest struct {
	Id                   string                                      `json:"id,omitempty"`
	ClientReferenceId    string                                      `json:"clientReferenceId"`
	Amount               *commonModel.Amount                         `json:"amount"`
	PaymentMethod        string                                      `json:"paymentMethod"`
	CreatedAt            time.Time                                   `json:"createdAt"`
	ExpiredAt            *time.Time                                  `json:"expiredAt,omitempty"`
	PaymentLink          string                                      `json:"paymentLink"`
	PaymentMethodOptions *UnifiedPaymentCallbackPaymentMethodOptions `json:"paymentMethodOptions"`
	Customer             *PaymentRequestCustomer                     `json:"customer,omitempty"`
	PaymentItems         []*PaymentItemRequest                       `json:"paymentItems,omitempty"`
}

type UnifiedPaymentCallbackPaymentMethodOptions struct {
	VirtualAccount interface{} `json:"virtualAccount,omitempty"`
	Card           interface{} `json:"card,omitempty"`
	Qris           interface{} `json:"qris,omitempty"`
}

type GetListFilterRequest struct {
	UUID           string     `json:"uuid"`
	MerchantID     string     `json:"merchantId"`
	ReferenceID    string     `json:"referenceId"`
	PaymentMethod  string     `json:"paymentMethod"`
	Status         string     `json:"status"`
	StartCreatedAt *time.Time `json:"startCreatedAt"`
	EndCreatedAt   *time.Time `json:"endCreatedAt"`

	Sort    string `json:"sort"`
	SortBy  string `json:"sortBy"`
	Page    int    `json:"page"`
	PerPage int    `json:"perPage"`

	IncludeCardFundedPaymentSession bool
}

type PostCreateLedgerRequest struct {
	Status  string             `json:"status"`
	Channel string             `json:"channel"`
	Amount  commonModel.Amount `json:"amount"`

	// Optional
	ChargeID     string `json:"-"`
	ChargeStatus string `json:"-"`
}

type CRMRetryNotificationRequest struct {
	ID            string `json:"-"`
	BankReference string `json:"bankReference"`
	// if true, it will force the payment to succeed, and bypass check status in snap core
	ForceSuccess bool `json:"forceSuccess"`
}

type CRMStaticVARetryNotificationRequest struct {
	VANumber string             `json:"vaNumber" validate:"required"`
	Amount   commonModel.Amount `json:"amount" validate:"required"`
}

type VCCTerminalChargeRequest struct {
	MerchantID       string                     `json:"-" validate:"required,uuid"`
	UserID           string                     `json:"-" validate:"required,uuid"`
	EncryptedRequest *encryption.DataEncryption `json:"-" validate:"required"`
}

type GetVCCTerminalListFilterRequest struct {
	ChargeID        string    `json:"chargeID"`
	ReferenceID     string    `json:"referenceID"`
	MerchantID      string    `json:"merchantId"`
	Status          string    `json:"status"`
	ChargeStartDate time.Time `json:"chargeStartDate"`
	ChargeEndDate   time.Time `json:"chargeEndDate"`

	Sort    string `json:"sort"`
	SortBy  string `json:"sortBy"`
	Page    int    `json:"page"`
	PerPage int    `json:"perPage"`
	Limit   int    `json:"limit"`
}

// GetPaymentReceiptCRMRequest - CRM endpoint request for payment receipt
type GetPaymentReceiptCRMRequest struct {
	ReferenceID string `json:"referenceId" validate:"required"`
	MerchantID  string `json:"merchantId" validate:"required,uuid"`
}

// GetPaymentReceiptRequest - Internal service request
type GetPaymentReceiptRequest struct {
	PaymentID   string // Optional: lookup by payment ID
	ReferenceID string // Lookup by reference ID + merchant ID
	MerchantID  string
}

type GetSubPaymentsRequest struct {
	ReferenceID string `json:"referenceId"`
	MerchantID  string `json:"merchantId"`
	Status      string `json:"status"`
}

type GetAutoSplitPaymentSummaryRequest struct {
	ReferenceID              string
	MerchantID               string
	MaxDateCreation          int
	ExcludeParentCalculation bool // used to avoid miss calculation and invalid final status during finalization
}

type ProcessSplitPaymentRequest struct {
	ParentPaymentID   string
	MerchantID        string
	FingerprintID     string
	ThreeDSCallbackID string
	Summary           *AutoSplitPaymentSummary
	MethodDetail      *unifiedPaymentModel.ChargePaymentMethodDetails
}

func (m AutoSplitPaymentSummary) GetFinalStatus() string {
	autoSplitStatus := constant.AutoSplitPaymentStatusProcessing

	if m.NumberOfCharges == 0 {
		return autoSplitStatus
	}

	switch {
	case m.NumberOfCharges == m.NumberOfFailedCharges:
		autoSplitStatus = constant.AutoSplitPaymentStatusFailed
	case m.NumberOfCharges == m.NumberOfExpiredCharges:
		autoSplitStatus = constant.AutoSplitPaymentStatusCancelled
	case m.NumberOfCharges == m.NumberOfSuccessfulCharges:
		autoSplitStatus = constant.AutoSplitPaymentStatusSuccess
	case m.NumberOfCharges == (m.NumberOfSuccessfulCharges + m.NumberOfFailedCharges + m.NumberOfExpiredCharges):
		autoSplitStatus = constant.AutoSplitPaymentStatusPartialSuccess
	}

	return autoSplitStatus
}

func (m *AutoSplitPaymentSummary) UpdateRecordByParentStatus(status string) {
	switch status {
	case constant.UnifiedPaymentSessionStatusPaid:
		m.NumberOfSuccessfulCharges++
	case constant.UnifiedPaymentSessionStatusCancelled:
		m.NumberOfFailedCharges++
	case constant.UnifiedPaymentSessionStatusExpired:
		m.NumberOfExpiredCharges++
	case constant.UnifiedPaymentSessionStatusRequireAction, constant.UnifiedPaymentSessionStatusProcessing:
		m.NumberOfInProcessCharges++
	}
}

type GetActivePaymentByProcessorReferenceNumberRequest struct {
	ProcessorReferenceNumber string
	Amount                   decimal.Decimal
}
