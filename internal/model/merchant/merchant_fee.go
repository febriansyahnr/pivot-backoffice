package merchant

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/snap/bankTransfer"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
)

var (
	loc, _ = time.LoadLocation(constant.TimeLoc)
)

type MerchantFee struct {
	UUID              string             `redis:"uuid" json:"uuid" db:"uuid"`
	MerchantID        string             `redis:"merchantId" json:"merchantId" db:"merchant_id"`
	Reference         string             `redis:"reference" json:"reference" db:"reference"`
	PaymentMethod     *string            `redis:"paymentMethod" json:"paymentMethod" db:"payment_method"`
	Channel           *string            `redis:"channel" json:"channel" db:"channel"`
	AmountType        string             `redis:"amountType" json:"amountType" db:"amount_type"`
	Amount            float64            `redis:"amount" json:"amount" db:"amount"`
	MaxFeeAmount      *float64           `redis:"maxFeeAmount" json:"maxFeeAmount" db:"max_fee_amount"`
	Percentage        float64            `redis:"percentage" json:"percentage" db:"percentage"`
	ReferenceType     string             `redis:"referenceType" json:"referenceType" db:"reference_type"`
	DeductionType     string             `redis:"deductionType" json:"deductionType" db:"deduction_type"`
	DeductionDay      *int16             `redis:"deductionDay" json:"deductionDay" db:"deduction_day"`
	DeductionLastDate *time.Time         `redis:"deductionLastDate" json:"deductionLastDate" db:"deduction_last_date"`
	TaxType           string             `redis:"taxType" json:"taxType" db:"tax_type"`
	TaxPercentage     float64            `redis:"taxPercentage" json:"taxPercentage" db:"tax_percentage"`
	SettlementModel   *string            `redis:"settlementModel" json:"settlementModel" db:"settlement_model"`
	SettlementMethod  *string            `redis:"settlementMethod" json:"settlementMethod" db:"settlement_method"`
	SettlementConfigs types.NullJSONText `redis:"-" json:"-" db:"settlement_configs"`
	TieringModel      *string            `redis:"-" json:"tieringModel,omitempty" db:"tiering_model"`
	TieringType       *string            `redis:"-" json:"tieringType,omitempty" db:"tiering_type"`
	TieringConfigs    types.NullJSONText `redis:"-" json:"-" db:"tiering_configs"`
	CreatedAt         time.Time          `redis:"-" json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time          `redis:"-" json:"updatedAt" db:"updated_at"`
	DeletedAt         *time.Time         `redis:"-" json:"deletedAt" db:"deleted_at"`

	SettlementConfigsObj SettlementConfig   `redis:"settlementConfigs" json:"settlementConfigs" db:"-"`
	TieringConfigsObj    []FeeTieringConfig `redis:"-" json:"tieringConfigs,omitempty" db:"-"`
}

func (m *MerchantFee) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

func (m *MerchantFee) UnmarshalBinary(buf []byte) error {
	return json.Unmarshal(buf, m)
}

type NewMerchantFeeRequest struct {
	MerchantID       string  `json:"merchantId" validate:"required"`
	Reference        string  `json:"reference" validate:"required"`
	PaymentMethod    string  `json:"paymentMethod" validate:"required_if=Reference PAYMENT,omitempty,oneof=VIRTUAL_ACCOUNT QRIS CREDIT_CARD EWALLET INSTALLMENT VIRTUAL_TERMINAL"`
	Channel          string  `json:"channel" validate:"required_if=PaymentMethod INSTALLMENT,omitempty,uppercase"`
	InstallmentTenor int     `json:"installmentTenor" validate:"required_if=PaymentMethod INSTALLMENT,omitempty,min=1"`
	Amount           float64 `json:"amount" validate:"min=0.0"`
	AmountType       string  `json:"amountType" validate:"required,oneof=AMOUNT PERCENTAGE AMOUNT_PERCENTAGE"`
	MaxFeeAmount     float64 `json:"maxFeeAmount" validate:"required_if=Reference PLATFORM_TRANSACTION,omitempty,min=0"`
	Percentage       float64 `json:"percentage" validate:"min=0,max=100"`
	ReferenceType    string  `json:"referenceType" validate:"required_if=Reference WALLET,omitempty,oneof=TOP_UP TRANSFER BANK_TRANSFER BILL MERCHANT_PAYMENT WALLET_TRANSACTION CHANNEL ACCOUNT"`
	DeductionType    string  `json:"deductionType" validate:"required,oneof=DIRECT AUTOMATED MANUAL"`
	DeductionDay     int16   `json:"deductionDay" validate:"required_if=DeductionType AUTOMATED,required_if=Reference PLATFORM_ACTIVITY,omitempty,min=1,max=31"`
	TaxType          string  `json:"taxType" validate:"required,oneof=NON_PKP INCLUSIVE EXCLUSIVE"`
	TaxPercentage    float64 `json:"taxPercentage" validate:"min=0,max=100"`
	SettlementModel  string  `json:"settlementModel" validate:"omitempty,oneof=DIRECT AGGREGATOR"`
	SettlementMethod string  `json:"settlementMethod" validate:"required_if=Reference PAYMENT_FUNDED_PAYOUT,omitempty,uppercase,oneof=INSTANT STANDARD"`
}

type UpdateMerchantFeeRequest struct {
	ID               string  `json:"id" validate:"required"`
	MerchantID       string  `json:"merchantId" validate:"required"`
	Amount           float64 `json:"amount" validate:"min=0"`
	AmountType       string  `json:"amountType" validate:"required,oneof=AMOUNT PERCENTAGE AMOUNT_PERCENTAGE"`
	MaxFeeAmount     float64 `json:"maxFeeAmount" validate:"omitempty,min=1"`
	Percentage       float64 `json:"percentage" validate:"min=0,max=100"`
	DeductionType    string  `json:"deductionType" validate:"required,oneof=DIRECT AUTOMATED MANUAL"`
	DeductionDay     int16   `json:"deductionDay" validate:"required_if=DeductionType AUTOMATED,omitempty,min=1,max=31"`
	TaxType          string  `json:"taxType" validate:"required,oneof=NON_PKP INCLUSIVE EXCLUSIVE"`
	TaxPercentage    float64 `json:"taxPercentage" validate:"min=0,max=100"`
	SettlementModel  string  `json:"settlementModel" validate:"omitempty,oneof=DIRECT AGGREGATOR"`
	SettlementMethod string  `json:"settlementMethod" validate:"omitempty,oneof=INSTANT STANDARD"`
}

type GetMerchantFeeRequest struct {
	ID               string `json:"id" validate:"required"`
	MerchantID       string `json:"merchantId" validate:"required"`
	AmountType       string `json:"amount_type" validate:"required"`
	Reference        string `json:"reference" validate:"required"`
	ReferenceType    string `json:"referenceType"`
	PaymentMethod    string `json:"paymentMethod"`
	Channel          string `json:"channel"`
	SettlementModel  string `json:"settlementModel"`
	SettlementMethod string `json:"settlementMethod"`
}

type MerchantFeeResponse struct {
	UUID             string    `json:"uuid"`
	MerchantID       string    `json:"merchantId"`
	PaymentMethod    *string   `json:"paymentMethod"`
	Channel          *string   `json:"channel,omitempty"`
	Amount           float64   `json:"amount"`
	AmountType       string    `json:"amountType"`
	MaxFeeAmount     *float64  `json:"maxFeeAmount,omitempty"`
	Percentage       float64   `json:"percentage"`
	Reference        string    `json:"reference"`
	ReferenceType    string    `json:"referenceType"`
	DeductionType    string    `json:"deductionType"`
	DeductionDay     *int16    `json:"deductionDay,omitempty"`
	TaxType          string    `json:"taxType"`
	TaxPercentage    float64   `json:"taxPercentage"`
	SettlementModel  *string   `json:"settlementModel"`
	SettlementMethod *string   `json:"settlementMethod,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type GetSettlementConfigRequest struct {
	MerchantId       string  `json:"merchantId"`
	Reference        string  `json:"reference"`
	Method           *string `json:"method"`
	Channel          *string `json:"channel"`
	SettlementMethod string  `json:"settlementMethod"` // INSTANT / STANDARD
}

type SettlementConfig struct {
	Type           string                  `json:"type" validate:"required"`
	CutOff         *SettlementConfigCutOff `json:"cutOff" validate:"omitempty,excluded_if=Type INSTANT"`
	SettlementTime string                  `json:"settlementTime,omitempty" validate:"omitempty,datetime=15:04:05"` // When Type using D+
	IsOnHold       bool                    `json:"isOnHold"`
}

type SettlementConfigCutOff struct {
	Window   SettlementConfigCutOffWindow   `json:"window" validate:"required"`
	Deferral SettlementConfigCutOffDeferral `json:"deferral" validate:"required"`
}

type SettlementConfigCutOffWindow struct {
	StartTime string `json:"startTime" validate:"required,datetime=15:04:05"`
	EndTime   string `json:"endTime" validate:"required,datetime=15:04:05"`
}

func (s *SettlementConfig) ValidateRequest() error {
	if s.Type != constant.SettlementTypeInstant && !util.IsValidSettlementTime(s.Type) {
		return fmt.Errorf("invalid settlement type format")
	}
	if util.IsSettlementTimeDayBased(s.Type) {
		_, err := util.ParseTimeToDatetime(time.Now().In(loc), s.SettlementTime)
		if err != nil {
			return fmt.Errorf("invalid settlement time format")
		}

		if s.CutOff != nil {
			s.CutOff.Deferral.ExecutionTime = s.SettlementTime
		}
	}

	return nil
}

func (s *SettlementConfig) GetSettlementDay() int {
	day := 0
	if s.Type == constant.SettlementTypeInstant {
		return day
	}

	if strings.HasPrefix(s.Type, constant.SettlementTimeTransactionBasedPrefix) {
		day, _ = strconv.Atoi(strings.Replace(s.Type, constant.SettlementTimeTransactionBasedPrefix, "", 1))
	} else if strings.HasPrefix(s.Type, constant.SettlementTimeDayBasedPrefix) {
		day, _ = strconv.Atoi(strings.Replace(s.Type, constant.SettlementTimeDayBasedPrefix, "", 1))
	}

	return day
}

func (s *SettlementConfig) GetSettlementTime(processTime time.Time) (time.Time, error) {
	if processTime.IsZero() {
		processTime = time.Now()
	}
	processTime = processTime.In(loc)

	if s.Type == constant.SettlementTypeInstant {
		return time.Time{}, nil
	}

	var (
		isCutOff         = false
		days             = 0
		executionTimeStr = ""
	)
	if s.CutOff != nil {
		var err error
		isCutOff, err = s.CutOff.Window.IsCutOffTime(processTime)
		if err != nil {
			return time.Time{}, err
		}
		executionTimeStr = s.CutOff.Deferral.ExecutionTime
	}

	days, _ = strconv.Atoi(strings.Replace(s.Type, constant.SettlementTimeTransactionBasedPrefix, "", 1))
	if strings.Contains(s.Type, constant.SettlementTimeDayBasedPrefix) {
		days, _ = strconv.Atoi(strings.Replace(s.Type, constant.SettlementTimeDayBasedPrefix, "", 1))
		executionTimeStr = s.SettlementTime
	}

	if isCutOff {
		days += s.CutOff.Deferral.OffsetDays
	}

	// Time using WIB time zone
	var settlementTime = processTime.AddDate(0, 0, days)
	if executionTime, err := util.ParseTimeToDatetime(settlementTime, executionTimeStr); err == nil {
		settlementTime = executionTime
	}

	return settlementTime, nil
}

func (t SettlementConfigCutOffWindow) IsCutOffTime(date time.Time) (bool, error) {
	startTime, err := util.ParseTimeToDatetime(date, t.StartTime)
	if err != nil {
		return false, fmt.Errorf("parsing start time: %s", err)
	}

	endTime, err := util.ParseTimeToDatetime(date, t.EndTime)
	if err != nil {
		return false, fmt.Errorf("parsing end time: %s", err)
	}

	if !startTime.Before(endTime) {
		// Check previous day's start time with current day's end time, and current day's start time with next day's end time.
		for _, add := range [][]int{{-1, 0}, {0, 1}} {
			t1 := startTime.AddDate(0, 0, add[0])
			t2 := endTime.AddDate(0, 0, add[1])

			if date.After(t1.Add(-time.Millisecond)) && date.Before(t2.Add(time.Millisecond)) {
				return true, nil
			}
		}
		return false, nil
	}
	return date.After(startTime.Add(-time.Millisecond)) && date.Before(endTime.Add(time.Millisecond)), nil
}

type SettlementConfigCutOffDeferral struct {
	OffsetDays    int    `json:"offsetDays" validate:"required,min=1"`
	ExecutionTime string `json:"executionTime" validate:"required,datetime=15:04:05"`
}

type OnBehalfFeeConfig struct {
	Id            string     `json:"id" db:"id"`
	MerchantId    string     `json:"merchantId" db:"merchant_id"`
	Type          string     `json:"type" db:"type"`
	SubMerchantId *string    `json:"subMerchantId" db:"sub_merchant_id"`
	Reference     string     `json:"reference" db:"reference"`
	ReferenceType string     `json:"referenceType" db:"reference_type"`
	PaymentMethod *string    `json:"paymentMethod" db:"payment_method"`
	AmountType    string     `json:"amountType" db:"amount_type"`
	Amount        float64    `json:"amount" db:"amount"`
	Percentage    float64    `json:"percentage" db:"percentage"`
	CreatedAt     time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time  `json:"updatedAt" db:"updated_at"`
	DeletedAt     *time.Time `json:"deletedAt" db:"deleted_at"`
}

type TransactionFeeOnBehalf struct {
	Reference  string  `json:"-" db:"reference"`
	Type       string  `json:"-" db:"type"`
	AmountType string  `json:"-" db:"amount_type"`
	Amount     float64 `json:"-" db:"amount"`
	Percentage float64 `json:"-" db:"percentage"`
}

type FeeTieringRequest struct {
	FeeId       string             `json:"-" validate:"required,uuid"`
	MerchantId  string             `json:"merchantId" validate:"required,uuid"`
	Model       string             `json:"tieringModel" validate:"omitempty,oneof=MONTHLY_ASSESSED LADDER"`
	Type        string             `json:"type" validate:"required,oneof=TPV FREQUENCY"`
	AppliedTier int                `json:"appliedTier" validate:"omitempty,min=1,max=5"`
	Configs     []FeeTieringConfig `json:"configs" validate:"required,min=1,max=5,dive,required"`
	AppliedFee  *FeeTieringConfig  `json:"-" validate:"-"`
}

type FeeTieringResponse struct {
	MerchantId    string             `json:"merchantId"`
	Reference     string             `json:"reference"`
	PaymentMethod *string            `json:"paymentMethod"`
	DeductionType string             `json:"deductionType"`
	Model         string             `json:"tieringModel"`
	Type          string             `json:"type"`
	AppliedTier   int                `json:"appliedTier,omitempty"`
	Configs       []FeeTieringConfig `json:"configs"`
}

type FeeTieringConfig struct {
	Tier          int      `json:"tier" validate:"required,min=1,max=5"`
	Min           float64  `json:"min" validate:"min=0,ltfield=Max"`
	Max           float64  `json:"max" validate:"required,min=1"`
	AmountType    string   `json:"amountType" validate:"required,oneof=AMOUNT PERCENTAGE AMOUNT_PERCENTAGE"`
	Amount        float64  `json:"amount" validate:"min=0"`
	Percentage    float64  `json:"percentage" validate:"min=0,max=100"`
	MaxFeeAmount  *float64 `json:"maxFeeAmount" validate:"omitempty,min=0"`
	TaxType       string   `json:"taxType" validate:"required,oneof=NON_PKP INCLUSIVE EXCLUSIVE"`
	TaxPercentage float64  `json:"taxPercentage" validate:"min=0,max=100"`
}

func (tier *FeeTieringConfig) Validate(fee MerchantFee) error {
	fee.AmountType = tier.AmountType
	fee.Amount = tier.Amount
	fee.Percentage = tier.Percentage
	fee.MaxFeeAmount = tier.MaxFeeAmount
	fee.TaxType = tier.TaxType
	fee.TaxPercentage = tier.TaxPercentage

	return fee.validate()
}

func (m *MerchantFee) UpdateMerchantFee(req *UpdateMerchantFeeRequest) (*MerchantFee, error) {
	m.Amount = req.Amount
	m.AmountType = req.AmountType
	m.UpdatedAt = time.Now().UTC()
	m.Percentage = req.Percentage
	m.DeductionType = req.DeductionType
	m.TaxType = req.TaxType
	m.TaxPercentage = req.TaxPercentage
	m.DeductionDay = nil
	m.MaxFeeAmount = nil

	if req.MaxFeeAmount > 0 {
		m.MaxFeeAmount = &req.MaxFeeAmount
	}

	if req.DeductionDay > 0 {
		m.DeductionDay = &req.DeductionDay
	}

	if req.SettlementModel != "" {
		m.SettlementModel = &req.SettlementModel
	}

	if err := m.validate(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *MerchantFee) ToResponse() *MerchantFeeResponse {
	resp := &MerchantFeeResponse{
		UUID:             m.UUID,
		MerchantID:       m.MerchantID,
		PaymentMethod:    m.PaymentMethod,
		Channel:          m.Channel,
		Amount:           m.Amount,
		AmountType:       m.AmountType,
		Percentage:       m.Percentage,
		Reference:        m.Reference,
		ReferenceType:    m.ReferenceType,
		DeductionType:    m.DeductionType,
		TaxType:          m.TaxType,
		TaxPercentage:    m.TaxPercentage,
		SettlementModel:  m.SettlementModel,
		SettlementMethod: m.SettlementMethod,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
	if m.MaxFeeAmount != nil {
		resp.MaxFeeAmount = m.MaxFeeAmount
	}
	if m.DeductionDay != nil {
		resp.DeductionDay = m.DeductionDay
	}
	return resp
}

func NewMerchantFee(req *NewMerchantFeeRequest) (*MerchantFee, error) {
	merchant := &MerchantFee{
		UUID:          uuid.New().String(),
		MerchantID:    req.MerchantID,
		Amount:        req.Amount,
		AmountType:    req.AmountType,
		Percentage:    req.Percentage,
		DeductionType: req.DeductionType,
		TaxType:       req.TaxType,
		TaxPercentage: req.TaxPercentage,
		Reference:     req.Reference,
		ReferenceType: req.ReferenceType,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	if req.PaymentMethod != "" {
		merchant.PaymentMethod = &req.PaymentMethod
	}
	if req.Channel != "" {
		merchant.Channel = &req.Channel
	}
	if req.MaxFeeAmount > 0 {
		merchant.MaxFeeAmount = &req.MaxFeeAmount
	}
	if req.DeductionDay > 0 {
		merchant.DeductionDay = &req.DeductionDay
	}
	if req.PaymentMethod == paymentConstant.PAYMENT_METHOD_INSTALLMENT {
		installmentChannel := strings.ToUpper(fmt.Sprintf(constant.MerchantFeeInstallmentChannelFormat, req.Channel, req.InstallmentTenor))
		merchant.Channel = &installmentChannel
	}
	if req.SettlementModel != "" {
		merchant.SettlementModel = &req.SettlementModel
	}
	if req.SettlementMethod != "" {
		merchant.SettlementMethod = &req.SettlementMethod
	}

	if err := merchant.validate(); err != nil {
		return nil, err
	}
	return merchant, nil
}

func (m *MerchantFee) validate() error {
	if m.Amount < 0.0 {
		return constant.ErrNegativeValue
	}

	if m.AmountType == constant.MerchantFeeAmountType && m.Percentage > 0.00 {
		return pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("percentage must be zero if fee type is AMOUNT"))

	} else if m.AmountType == constant.MerchantFeePercentageType && m.Amount > 0.00 {
		return pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("amount must be zero if fee type is PERCENTAGE"))

	} else if (m.Reference != constant.ReferencePlatformTransaction && m.ReferenceType != constant.WalletTrxType) && m.MaxFeeAmount != nil && *m.MaxFeeAmount > 0 {
		return pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("not allowed to have maximum fee for requested reference & referenceType"))

	} else if m.Reference == constant.ReferencePlatformTransaction && (m.MaxFeeAmount == nil || *m.MaxFeeAmount == 0) {
		return pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("maximum fee amount is required field"))

	} else if m.Reference == constant.ReferencePlatformTransaction && (m.AmountType != constant.MerchantFeePercentageType || m.Amount > 0) {
		return pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("platform transaction fee type must use PERCENTAGE type"))

	} else if m.Reference != constant.ReferencePlatformActivity && m.DeductionType != constant.MerchantFeeDeductionTypeAutomated && m.DeductionDay != nil && *m.DeductionDay > 0 {
		return pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("deduction day only for automated deduction type or platform activity"))

	} else if m.Reference == constant.ReferencePlatformActivity && (m.AmountType != constant.MerchantFeeAmountType || m.Percentage > 0) {
		return pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("platform activity fee type must use AMOUNT type"))

	} else if m.Reference == constant.ReferencePlatformActivity && m.DeductionType == constant.MerchantFeeDeductionTypeDirect {
		return pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("for this reference must use indirect deduction"))

	} else if m.Channel != nil && !slices.Contains([]string{constant.ReferencePayment, constant.ReferenceDisbursement, constant.ReferenceDisbursementVA, constant.ReferenceXB, constant.ReferenceTopUp, constant.ReferencePaymentFundedPayout}, m.Reference) {
		return pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("channel attribute is only applicable for the PAYMENT and PAYOUT reference"))
	}

	if slices.Contains([]string{constant.ReferencePayment, constant.ReferencePaymentFundedPayout}, m.Reference) && util.ValueOfPtr(m.PaymentMethod) == constant.ChannelCreditCard && m.Channel != nil {
		valid := slices.ContainsFunc(config.GetCreditCardReferences().CardBrands, func(brand string) bool {
			return *m.Channel == "LOCAL_"+brand || *m.Channel == "FOREIGN_"+brand
		})
		if !valid {
			return pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("invalid credit card channel"))
		}
	}

	if slices.Contains([]string{constant.ReferenceDisbursement, constant.ReferenceDisbursementVA}, m.Reference) && m.Channel != nil && bankTransfer.NewBankDB().FindByChannelCode(*m.Channel) == nil {
		return pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("invalid disbursement channel code"))
	}

	switch m.Reference {
	case constant.ReferenceDisbursement, constant.ReferenceDisbursementVA, constant.ReferenceAccountInquiry, constant.ReferencePayment,
		constant.ReferencePlatformActivity, constant.ReferencePlatformTransfer, constant.ReferencePlatformTransaction,
		constant.ReferenceWallet, constant.ReferenceRefund, constant.ReferenceXB,
		constant.ReferenceTopUp, constant.ReferencePaymentFundedPayout:

		return nil
	default:

		return pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrInvalidRequestPayload)
	}
}

type MerchantFeeXBQuery struct {
	MerchantID string
	Reference  string
	Channel    string
}

type XbFeeConfigResponse struct {
	Local *MerchantFeeResponse `json:"local"`
	Swift *MerchantFeeResponse `json:"swift"`
}
