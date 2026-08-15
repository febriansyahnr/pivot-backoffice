package orchestrator_model

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"

	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
)

type AccountTransaction struct {
	UUID                   uuid.UUID          `db:"uuid" json:"uuid"`
	ReferenceID            string             `db:"reference_id" json:"referenceId"`
	MerchantID             uuid.UUID          `db:"merchant_id" json:"merchantId"`
	AccountID              uuid.UUID          `db:"account_id" json:"accountId"`
	MerchantReferenceID    *string            `db:"merchant_reference_id" json:"merchantReferenceId"`
	Currency               string             `db:"currency" json:"currency"`
	Credit                 float64            `db:"credit" json:"credit"`
	Debit                  float64            `db:"debit" json:"debit"`
	Reference              string             `db:"reference" json:"reference"`
	Type                   string             `db:"type" json:"type"`
	Channel                string             `db:"channel" json:"channel"`
	Status                 string             `db:"status" json:"status"`
	ReasonType             sql.NullString     `db:"reason_type" json:"reasonType"`
	ReasonDescription      sql.NullString     `db:"reason_description" json:"reasonDescription"`
	Remarks                string             `db:"remarks" json:"remarks"`
	Processor              string             `db:"processor_reference" json:"processorReference"`
	ProcessorID            string             `db:"processor_reference_id" json:"processorReferenceId"`
	ProcessorTransactionID string             `db:"processor_transaction_id" json:"processorTransactionId"`
	TransactionTimestamp   time.Time          `db:"transaction_timestamp" json:"transactionTimestamp"`
	AdditionalInfo         types.NullJSONText `db:"additional_info" json:"additionalInfo"`
	CreatedAt              time.Time          `db:"created_at" json:"createdAt"`
	UpdatedAt              time.Time          `db:"updated_at" json:"updatedAt"`
	DeletedAt              sql.NullTime       `db:"deleted_at" json:"deletedAt,omitempty"`
	SettlementAt           sql.NullTime       `db:"settlement_at" json:"settlementAt"`
	SettlementStatus       sql.NullString     `db:"settlement_status" json:"settlementStatus"`
	SettlementModel        sql.NullString     `db:"settlement_model" json:"settlementModel"`
}

type AccountTransactionWithUseCase struct {
	UUID                 uuid.UUID          `db:"uuid" json:"uuid"`
	ReferenceID          string             `db:"reference_id" json:"referenceId"`
	MerchantID           uuid.UUID          `db:"merchant_id" json:"merchantId"`
	AccountID            uuid.UUID          `db:"account_id" json:"accountId"`
	Currency             string             `db:"currency" json:"currency"`
	Credit               float64            `db:"credit" json:"credit"`
	Debit                float64            `db:"debit" json:"debit"`
	Reference            string             `db:"reference" json:"reference"`
	Type                 string             `db:"type" json:"type"`
	Channel              string             `db:"channel" json:"channel"`
	Status               string             `db:"status" json:"status"`
	ReasonType           sql.NullString     `db:"reason_type" json:"reasonType"`
	ReasonDescription    sql.NullString     `db:"reason_description" json:"reasonDescription"`
	Remarks              string             `db:"remarks" json:"remarks"`
	ProcessorReference   string             `db:"processor_reference" json:"processor_reference"`
	AdditionalInfo       types.NullJSONText `db:"additional_info" json:"-"`
	TransactionTimestamp time.Time          `db:"transaction_timestamp" json:"transactionTimestamp"`
	CreatedAt            time.Time          `db:"created_at" json:"createdAt"`
	UpdatedAt            time.Time          `db:"updated_at" json:"updatedAt"`
	SettlementAt         sql.NullTime       `db:"settlement_at" json:"settlementAt"`
	SettlementStatus     sql.NullString     `db:"settlement_status" json:"settlementStatus"`
	SettlementModel      sql.NullString     `db:"settlement_model" json:"settlementModel"`

	// Use case attributes
	SenderName             string     `db:"sender_name" json:"senderName"`
	Fee                    float64    `db:"fee" json:"fee"`
	BeneficiaryAccountNo   string     `db:"beneficiary_account_no" json:"beneficiaryAccountNo"`
	BeneficiaryAccountName string     `db:"beneficiary_account_name" json:"beneficiaryAccountName"`
	BeneficiaryBankName    string     `db:"beneficiary_bank_name" json:"beneficiaryBankName"`
	ClientReferenceID      string     `db:"client_reference_id" json:"clientReferenceId"`
	ApprovedAt             *time.Time `db:"approved_at" json:"approvedAt"`
	ProcessorReferenceId   string     `db:"processor_reference_id" json:"trxRef"`
	ProcessorTransactionId string     `db:"processor_transaction_id" json:"trxId"`
	CreatedBy              string     `db:"created_by" json:"createdBy"`
	ApprovedBy             string     `db:"approved_by" json:"approvedBy"`
	BalanceType            string     `db:"balance_type" json:"balanceType"`
	MerchantReferenceID    string     `db:"merchant_reference_id" json:"merchantReferenceId"`
	BankReference          string     `db:"bank_reference" json:"-"`
	// Internal Data
	AdditionalInfoObj interface{} `db:"-" json:"additionalInfo"`
}

func (a *AccountTransactionWithUseCase) ToTransactionHistory() TransactionHistory {
	result := TransactionHistory{
		Id:                     a.UUID.String(),
		MerchantReferenceID:    a.MerchantReferenceID,
		BalanceType:            a.BalanceType,
		Type:                   a.Type,
		Channel:                a.Channel,
		Amount:                 -1 * a.Debit,
		Fee:                    a.Fee,
		BankReference:          a.BankReference,
		Status:                 a.Status,
		CreatedAt:              a.CreatedAt,
		CreatedBy:              a.CreatedBy,
		Remarks:                a.Remarks,
		BeneficiaryBankName:    a.BeneficiaryBankName,
		BeneficiaryAccountNo:   a.BeneficiaryAccountNo,
		BeneficiaryAccountName: a.BeneficiaryAccountName,
	}
	if a.Credit != 0 {
		result.Amount = a.Credit
	}
	if a.ReasonType.Valid {
		result.ReasonType = a.ReasonType.String
	}
	if a.ReasonDescription.Valid {
		result.ReasonDescription = a.ReasonDescription.String
	}
	if a.SettlementAt.Valid {
		result.SettlementAt = &a.SettlementAt.Time
	}
	return result
}

type CreateNewTransactionRequest struct {
	ReferenceID            string      `json:"referenceId"`
	MerchantID             uuid.UUID   `json:"merchantId"`
	AccountID              uuid.UUID   `json:"accountId"`
	MerchantReferenceID    *string     `json:"merchantReferenceId"`
	Currency               string      `json:"currency"`
	Credit                 float64     `json:"credit"`
	Debit                  float64     `json:"debit"`
	Reference              string      `json:"reference"`
	Type                   string      `json:"type"`
	Channel                string      `json:"channel"`
	Status                 string      `json:"status"`
	ReasonType             string      `json:"reasonType"`
	ReasonDescription      string      `json:"reasonDescription"`
	Remarks                string      `json:"remarks"`
	TransactionTimestamp   time.Time   `json:"transactionTimestamp"`
	AdditionalInfo         interface{} `json:"additionalInfo"`
	ProcessorReference     string      `json:"processorReference"`
	ProcessorReferenceID   string      `json:"processorReferenceId"`
	ProcessorTransactionID string      `json:"processorTransactionId"`
	SettlementStatus       string      `json:"settlementStatus"`
	SettlementAt           time.Time   `json:"settlementAt"`
	SettlementModel        string      `json:"settlementModel"`
}

// UUIDGenerator is a function type for generating UUIDs
type UUIDGenerator func() (uuid.UUID, error)

// defaultUUIDGenerator is the default implementation using uuid.New
var defaultUUIDGenerator UUIDGenerator = func() (uuid.UUID, error) {
	return uuid.New(), nil
}

// GetDefaultUUIDGenerator returns the current UUID generator function.
// This is primarily used for testing.
func GetDefaultUUIDGenerator() UUIDGenerator {
	return defaultUUIDGenerator
}

// SetDefaultUUIDGenerator sets the UUID generator function.
// This is primarily used for testing.
func SetDefaultUUIDGenerator(generator UUIDGenerator) {
	defaultUUIDGenerator = generator
}

func NewAccountTransaction(request *CreateNewTransactionRequest) (*AccountTransaction, error) {
	var (
		additionalInfo []byte
		err            error
	)

	if request.Credit < 0 && request.Debit < 0 {
		return nil, errors.New("credit and debit cannot be negative")
	}

	if err = ValidateUseCase(request.Reference); err != nil {
		return nil, err
	}

	if err = validateStatus(request.Status); err != nil {
		return nil, err
	}

	if request.TransactionTimestamp.IsZero() {
		return nil, constant.ErrInvalidTransactionTimestamp
	}

	if request.AdditionalInfo != nil {
		additionalInfo, err = json.Marshal(request.AdditionalInfo)
		if err != nil {
			return nil, constant.ErrInvalidAdditionalInfo
		}
	}

	id, err := defaultUUIDGenerator()
	if err != nil {
		return nil, err
	}

	accountTrx := &AccountTransaction{
		UUID:                 id,
		ReferenceID:          request.ReferenceID,
		MerchantID:           request.MerchantID,
		AccountID:            request.AccountID,
		MerchantReferenceID:  request.MerchantReferenceID,
		Currency:             request.Currency,
		Credit:               request.Credit,
		Debit:                request.Debit,
		Type:                 request.Type,
		Channel:              request.Channel,
		Status:               request.Status,
		Remarks:              request.Remarks,
		TransactionTimestamp: request.TransactionTimestamp,
		Reference:            request.Reference,
		ReasonType:           sql.NullString{String: request.ReasonType, Valid: request.ReasonType != ""},
		ReasonDescription:    sql.NullString{String: request.ReasonDescription, Valid: request.ReasonDescription != ""},
		CreatedAt:            time.Now().UTC(),
		UpdatedAt:            time.Now().UTC(),
		AdditionalInfo:       types.NullJSONText{JSONText: additionalInfo, Valid: additionalInfo != nil},
		SettlementStatus:     sql.NullString{String: request.SettlementStatus, Valid: request.SettlementStatus != ""},
		SettlementAt:         sql.NullTime{Time: request.SettlementAt, Valid: !request.SettlementAt.IsZero()},
		SettlementModel:      sql.NullString{String: request.SettlementModel, Valid: request.SettlementModel != ""},
	}

	if request.ProcessorReference != "" {
		accountTrx.Processor = request.ProcessorReference
	}

	if request.ProcessorReferenceID != "" {
		accountTrx.ProcessorID = request.ProcessorReferenceID
	}

	if request.ProcessorTransactionID != "" {
		accountTrx.ProcessorTransactionID = request.ProcessorTransactionID
	}

	return accountTrx, nil
}

func ValidateUseCase(usecase string) error {
	switch strings.ToUpper(usecase) {
	case constant.ReferenceDisbursement, constant.ReferencePayment, constant.ReferenceWallet, constant.ReferencePlatform:
		return nil
	default:
		return constant.ErrInvalidUsecase
	}
}

func validateType(usecaseType string) error {
	switch strings.ToUpper(usecaseType) {
	case "", constant.TypePayment, constant.TypeDisbursement, constant.TypeTopUp,
		constant.TypeManualAdjust, constant.TypeFee, constant.TypeAccountInquiryFee,
		constant.TypeWalletTopUp, constant.TypeWalletTransfer, constant.TypeWalletWithdrawal, constant.TypeWalletBillPayment:
		return nil
	default:
		return constant.ErrInvalidType
	}
}

func validateChannel(channelType string) error {
	switch strings.ToUpper(channelType) {
	case "", constant.ChannelBalance, constant.ChannelVirtualAccount, constant.ChannelCreditCard,
		constant.ChannelBankTransfer, constant.ChannelManualTransfer, constant.ChannelPPOB,
		constant.ChannelBalanceAdjustment:
		return nil
	default:
		return constant.ErrInvalidChannel
	}
}

func validateStatus(status string) error {
	switch strings.ToUpper(status) {
	case constant.StatusSuccess, constant.StatusFailed, constant.StatusPending:
		return nil
	default:
		return constant.ErrInvalidStatus
	}
}

// GetCreditcardMetadataFromAdditionalInfo returns the creditcard metadata from AccountTransactionWithUseCase
func (a *AccountTransactionWithUseCase) GetCreditcardMetadataFromAdditionalInfo() (*creditcardModel.CreditcardMetadata, error) {
	var (
		cardBrand       string
		cardCountryCode string
	)

	// Check if AdditionalInfo is valid
	if !a.AdditionalInfo.Valid || len(a.AdditionalInfo.JSONText) == 0 {
		return nil, errors.New("additional info is empty")
	}

	// Parse the JSON directly into a map
	var additionalData map[string]interface{}
	if err := json.Unmarshal(a.AdditionalInfo.JSONText, &additionalData); err != nil {
		return nil, err
	}

	methodDetail, hasMethodDetail := additionalData["methodDetail"].(map[string]interface{})
	if !hasMethodDetail {
		return nil, errors.New("methodDetail is empty")
	}

	// Create CreditcardMetadata object
	metadata := &creditcardModel.CreditcardMetadata{
		ProcessorStatus: getString(additionalData, "chargeStatus"),
	}

	cardData, hasCardData := methodDetail["card"].(map[string]interface{})
	if !hasCardData {
		// Try alternative key "cardData"
		cardData, hasCardData = methodDetail["cardData"].(map[string]interface{})
		if !hasCardData {
			return nil, errors.New("card data not found in methodDetail")
		}
	}

	last4, hasLast4 := cardData["last4"].(string)
	if !hasLast4 {
		last4, hasLast4 = cardData["last4Digit"].(string)
		if !hasLast4 {
			return nil, errors.New("last4 not found in cardData")
		}
	}

	first8, hasFirst8 := cardData["first8"].(string)
	if !hasFirst8 {
		first8, hasFirst8 = cardData["first8Digit"].(string)
		if !hasFirst8 {
			return nil, errors.New("first8 not found in cardData")
		}
	}

	cardBrand, hasCardBrand := cardData["brand"].(string)
	if !hasCardBrand {
		cardBrand, hasCardBrand = cardData["cardBrand"].(string)
		if !hasCardBrand {
			cardBrand = ""
		}
	}

	cardCountryCode, hasCardCountryCode := cardData["countryCode"].(string)
	if !hasCardCountryCode {
		cardCountryCode = ""
	}

	// Extract card data
	metadata.CardData = &creditcardModel.CardDataRequest{
		Last4Digit:  last4,
		First8Digit: first8,
		Fingerprint: getString(cardData, "fingerprint"),
		CardBrand:   cardBrand,
		CountryCode: cardCountryCode,
	}

	// Extract BIN information
	if binInfo, ok := cardData["binInformations"].(map[string]interface{}); ok {
		metadata.CardData.CardType = getString(binInfo, "type")
		metadata.CardData.CardBrand = getString(binInfo, "brand")
		metadata.CardData.CardIssuing = getString(binInfo, "issuingBank")
		metadata.CardData.CountryCode = getString(binInfo, "country")
	}

	if cardBrand == "" && metadata.CardData.CardBrand == "" {
		return nil, errors.New("cardBrand not found in cardData")
	}

	if cardCountryCode == "" && metadata.CardData.CountryCode == "" {
		return nil, errors.New("cardCountryCode not found in cardData")
	}

	if cardBrand != "" {
		metadata.CardData.CardBrand = cardBrand
	}

	if cardCountryCode != "" {
		metadata.CardData.CountryCode = cardCountryCode
	}

	// Extract authorization data
	if authData, ok := cardData["authorizationResult"].(map[string]interface{}); ok && len(authData) > 0 {
		authBytes, err := json.Marshal(authData)
		if err == nil && len(authBytes) > 0 {
			metadata.AuthorizationData = &creditcardModel.PaymentNotificationAuthorizationDataRequest{}
			_ = json.Unmarshal(authBytes, metadata.AuthorizationData)
		}
	}

	// Extract authentication data
	if authData, ok := cardData["authenticationResult"].(map[string]interface{}); ok && len(authData) > 0 {
		authBytes, err := json.Marshal(authData)
		if err == nil && len(authBytes) > 0 {
			metadata.AuthenticationData = &creditcardModel.PaymentNotificationAuthenticationDataRequest{}
			_ = json.Unmarshal(authBytes, metadata.AuthenticationData)
		}
	}

	// Extract fee details
	if feeDetail, ok := additionalData["feeDetail"].(map[string]interface{}); ok && len(feeDetail) > 0 {
		feeBytes, err := json.Marshal(feeDetail)
		if err == nil && len(feeBytes) > 0 {
			metadata.FeeDetail = &feeModel.FeeMetadataObject{}
			_ = json.Unmarshal(feeBytes, metadata.FeeDetail)
		}
	}

	return metadata, nil
}

// Helper function to safely get string values from a map
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

type TransactionHistory struct {
	CreatedAt              time.Time  `db:"created_at"`
	UpdatedAt              time.Time  `db:"updated_at"`
	Type                   string     `db:"type"`
	Channel                string     `db:"channel"`
	CreatedBy              string     `db:"created_by"`
	Id                     string     `db:"id"`
	LinkedId               string     `db:"linked_id"`
	ReferenceId            string     `db:"reference_id"`
	Amount                 float64    `db:"amount"`
	BankReference          string     `db:"bank_reference"`
	Status                 string     `db:"status"`
	ReasonType             string     `db:"reason_type"`
	ReasonDescription      string     `db:"reason_description"`
	Remarks                string     `db:"remarks"`
	BeneficiaryBankName    string     `db:"beneficiary_bank_name"`
	BeneficiaryAccountNo   string     `db:"beneficiary_account_no"`
	BeneficiaryAccountName string     `db:"beneficiary_account_name"`
	ApprovalStatus         string     `db:"approval_status"`
	ApprovedAt             *time.Time `db:"approved_at"`
	ApprovedBy             string     `db:"approved_by"`
	BalanceType            string     `db:"balance_type"`
	MerchantReferenceID    string     `db:"merchant_reference_id"`
	SettlementAt           *time.Time `db:"settlement_at"`
	Fee                    float64    `db:"fee"`
}

type TransactionActivity struct {
	MerchantID string `json:"merchantId" db:"merchant_id"`
	Period     string `json:"period" db:"period"`
	Total      uint64 `json:"total" db:"total"`
}

type AccumulateTransactionFees struct {
	AccountName       string         `db:"account_name"`
	TotalRows         uint64         `db:"total_rows"`
	TotalFees         float64        `db:"total_fees"`
	TotalTaxes        float64        `db:"total_taxes"`
	RawTransactionIds types.JSONText `db:"transaction_ids"`
	TransactionIds    []string       `db:"-"`
}

type CalculatingMerchantTPVSummary struct {
	Type       string  `db:"type"`
	Channel    string  `db:"channel"`
	Additional *string `db:"additional"`
	Frequency  float64 `db:"frequency"`
	Volume     float64 `db:"volume"`
}

type UpdateTransactionWithPendingStatus struct {
	Channel         string
	Metadata        []byte
	UpdatedAt       time.Time
	Processor       string
	ProcessorID     string
	SettlementModel string
}

type UpdateSettlementDetailRequest struct {
	EstimateSettlementAt *time.Time
}
