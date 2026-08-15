package paymentModel

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx/types"
	paymentCaptureModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/paymentCapture"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConst "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/common"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/creditcard"
	fdsCommonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/fdsProcessor/fdsCommon"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/orchestrator"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/proto/messages/callback"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/unifiedPayment"

	"github.com/paper-indonesia/pdk/v2/gcp"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	DefaultSortColumn = "createdAt"
)

type Payment struct {
	UUID                     string           `json:"uuid"`
	ReferenceID              *string          `json:"referenceId"`
	MerchantID               string           `json:"merchantId"`
	CustomerID               string           `json:"customerId"`
	PaymentMethodID          string           `json:"paymentMethodId"`
	ProcessorReferenceNumber *string          `json:"processorReferenceNumber"`
	RecurringContractID      *string          `json:"recurringContractId,omitempty"`
	Currency                 string           `json:"currency"`
	Amount                   decimal.Decimal  `json:"amount"`
	Fee                      *decimal.Decimal `json:"fee"`
	Discount                 *decimal.Decimal `json:"discount"`
	TotalAmount              decimal.Decimal  `json:"totalAmount"`
	Status                   string           `json:"status"`
	Type                     string           `json:"type"`
	Metadata                 *map[string]any  `json:"metadata"`
	PaymentURL               string           `json:"paymentUrl"`
	ReasonType               *string          `json:"reasonType,omitempty"`
	ReasonDescription        *string          `json:"reasonDescription,omitempty"`
	CreatedAt                time.Time        `json:"createdAt"`
	UpdatedAt                time.Time        `json:"updatedAt"`
	DeletedAt                *time.Time       `json:"deletedAt"`
	ExpiredAt                *time.Time       `json:"expiredAt"`
	InvestigationStartedAt   *time.Time       `json:"investigationStartedAt"`

	PaymentMethod          PaymentMethod `json:"-"`
	Processor              string        `json:"-"`
	ProcessorID            string        `json:"-"`
	ProcessorTransactionID string        `json:"-"`
	ReconReferenceNo       string        `json:"-"`
	TrxDatetime            *time.Time    `json:"-"`
	SnapCoreId             *string       `json:"-"` // From metadata.snapCore.uuid
	BankReferenceId        string        `json:"-"`
	CreatedBy              *string       `json:"-"`
	CreatedFrom            string        `json:"-"`

	// Optional payment captures
	PaymentCaptures []*paymentCaptureModel.PaymentCapture `json:"-"`

	// Recurring payment details for internal needs
	RecurringPayment *unifiedPaymentModel.MetadataRecurringPayment `json:"-"`

	// Card-funded payout attribute.
	//
	// Note: Before using this attribute, ensure that the previous process has parsed the metadata and assigned its value to this attribute.
	CardFundedPayout *unifiedPaymentModel.CardFundedPayout `json:"-"`

	// Auto split payment attribute.
	//
	// Note: Before using this attribute, ensure the value has been parsed from metadata.
	AutoSplitPayment *unifiedPaymentModel.AutoSplitPayment `json:"-"`

	InquiryDetail *unifiedPaymentModel.InquiryDetail `json:"-"`
}

func (p Payment) GetGroupPaymentType() string {
	if p.RecurringContractID != nil && *p.RecurringContractID != "" {
		return constant.GroupPaymentTypeRecurringPayment
	}
	if p.AutoSplitPayment != nil {
		return constant.GroupPaymentTypeSplitPayment
	}

	switch p.Type {
	case constant.PaymentTypeCardFundedPayout:
		return constant.GroupPaymentTypeCardFundedPayout

	case constant.PaymentTypeOneDollarAuth:
		return constant.GroupPaymentTypeOneDollarAuth

	case constant.PaymentTypeVirtualTerminal:
		return constant.GroupPaymentTypeVirtualTerminal

	default:
		return constant.GroupPaymentTypePayment
	}
}

type PaymentDTO struct {
	UUID                     string           `json:"uuid" db:"uuid"`
	ReferenceID              *string          `json:"referenceId" db:"reference_id"`
	MerchantID               string           `json:"merchantId" db:"merchant_id"`
	CustomerID               string           `json:"customerId" db:"customer_id"`
	PaymentMethodID          string           `json:"paymentMethodId" db:"payment_method_id"`
	ProcessorReferenceNumber *string          `json:"processorReferenceNumber" db:"processor_reference_number"`
	RecurringContractID      *string          `json:"recurringContractId" db:"recurring_contract_id"`
	Currency                 string           `json:"currency" db:"currency"`
	Amount                   decimal.Decimal  `json:"amount" db:"amount"`
	Fee                      *decimal.Decimal `json:"fee" db:"fee"`
	Discount                 *decimal.Decimal `json:"discount" db:"discount"`
	TotalAmount              decimal.Decimal  `json:"totalAmount" db:"total_amount"`
	Status                   string           `json:"status" db:"status"`
	Type                     string           `json:"type" db:"type"`
	Metadata                 *string          `json:"metadata" db:"metadata"`
	PaymentURL               string           `json:"paymentUrl" db:"payment_url"`
	ReasonType               *string          `json:"reasonType" db:"reason_type"`
	ReasonDescription        *string          `json:"reasonDescription" db:"reason_description"`
	CreatedBy                *string          `json:"createdBy" db:"created_by"`
	CreatedFrom              string           `json:"createdFrom" db:"created_from"`
	CreatedAt                time.Time        `json:"createdAt" db:"created_at"`
	UpdatedAt                time.Time        `json:"updatedAt" db:"updated_at"`
	DeletedAt                *time.Time       `json:"deletedAt" db:"deleted_at"`
	ExpiredAt                *time.Time       `json:"expiredAt" db:"expired_at"`
}

type PaymentWithPaymentMethodDTO struct {
	UUID                     string           `json:"uuid" db:"uuid"`
	ReferenceID              *string          `json:"referenceId" db:"reference_id"`
	MerchantID               string           `json:"merchantId" db:"merchant_id"`
	CustomerID               string           `json:"customerId" db:"customer_id"`
	PaymentMethodID          string           `json:"paymentMethodId" db:"payment_method_id"`
	ProcessorReferenceNumber *string          `json:"processorReferenceNumber" db:"processor_reference_number"`
	RecurringContractID      *string          `json:"recurringContractId,omitempty" db:"recurring_contract_id"`
	Currency                 string           `json:"currency" db:"currency"`
	Amount                   decimal.Decimal  `json:"amount" db:"amount"`
	Fee                      *decimal.Decimal `json:"fee" db:"fee"`
	Discount                 *decimal.Decimal `json:"discount" db:"discount"`
	TotalAmount              decimal.Decimal  `json:"totalAmount" db:"total_amount"`
	Status                   string           `json:"status" db:"status"`
	Type                     string           `json:"type" db:"type"`
	Metadata                 *string          `json:"metadata" db:"metadata"`
	PaymentURL               string           `json:"paymentUrl" db:"payment_url"`
	ReasonType               *string          `json:"reasonType" db:"reason_type"`
	ReasonDescription        *string          `json:"reasonDescription" db:"reason_description"`
	CreatedFrom              string           `json:"createdFrom" db:"created_from"`
	CreatedBy                *string          `json:"createdBy" db:"created_by"`
	CreatedAt                time.Time        `json:"createdAt" db:"created_at"`
	UpdatedAt                time.Time        `json:"updatedAt" db:"updated_at"`
	DeletedAt                sql.NullTime     `json:"deletedAt" db:"deleted_at"`
	ExpiredAt                sql.NullTime     `json:"expiredAt" db:"expired_at"`
	InvestigationStartedAt   sql.NullTime     `json:"investigationStartedAt" db:"investigation_started_at"`

	PaymentMethodType     sql.NullString `json:"paymentMethodType" db:"payment_method_type"`
	PaymentMethodName     sql.NullString `json:"paymentMethodName" db:"payment_method_name"`
	PaymentMethodAcquirer sql.NullString `json:"paymentMethodAcquirer" db:"payment_method_acquirer"`
	PaymentMethodBankName sql.NullString `json:"paymentMethodBankName" db:"payment_method_bank_name"`
	PaymentMethodLogo     sql.NullString `json:"paymentMethodLogo" db:"payment_method_logo"`
	PaymentSnapCoreId     *string        `json:"-" db:"payment_snap_core_id"`

	// Optional payment captures - stored as raw JSON from database
	PaymentCapturesRaw types.NullJSONText `json:"-" db:"payment_captures"`
}

// PaymentReceiptDTO contains all data needed for generating payment receipt
// Fetched via single JOIN query to reduce database round-trips
type PaymentReceiptDTO struct {
	UUID          string          `db:"uuid"`
	ReferenceID   *string         `db:"reference_id"`
	MerchantID    string          `db:"merchant_id"`
	TotalAmount   decimal.Decimal `db:"total_amount"`
	Status        string          `db:"status"`
	CreatedAt     time.Time       `db:"created_at"`
	MerchantName  sql.NullString  `db:"merchant_name"`
	PaymentMethod sql.NullString  `db:"payment_method_type"`
}

type ExpiringPayment struct {
	UUID         string    `json:"uuid" db:"uuid"`
	MerchantID   string    `json:"merchantId" db:"merchant_id"`
	ExpiredAt    time.Time `json:"expiredAt" db:"expired_at"`
	ChargeStatus string    `json:"chargeStatus" db:"charge_status"`
	Note         string    `json:"note,omitempty"`
}

func ValidatePaymentStatus(status string) error {
	switch strings.ToUpper(status) {
	case paymentConst.PAYMENT_STATUS_PENDING, paymentConst.PAYMENT_STATUS_VOID, paymentConst.PAYMENT_STATUS_SUCCESS:
		return nil
	default:
		return constant.ErrInvalidPaymentStatus
	}
}

func ValidateUnifiedPaymentStatus(status string) error {
	switch strings.ToUpper(status) {
	case paymentConst.PAYMENT_STATUS_PENDING, paymentConst.PAYMENT_STATUS_VOID, paymentConst.PAYMENT_STATUS_SUCCESS,
		constant.UnifiedPaymentSessionStatusRequirePaymentMethod,
		constant.UnifiedPaymentSessionStatusRequireConfirmation,
		constant.UnifiedPaymentSessionStatusRequireAction,
		constant.UnifiedPaymentSessionStatusProcessing,
		constant.UnifiedPaymentSessionStatusCancelled,
		constant.UnifiedPaymentSessionStatusExpired,
		constant.UnifiedPaymentSessionStatusPaid,
		constant.UnifiedStaticPaymentStatusActive,
		constant.UnifiedStaticPaymentStatusInactive,
		constant.PaymentStatusRefunded:
		return nil
	default:
		return constant.ErrInvalidPaymentStatus
	}
}

func ValidatePaymentHistorySortColumn(column string) error {
	switch column {
	case "createdAt", "amount", "amountPaid", "paymentDate":
		return nil
	default:
		return constant.ErrInvalidPaymentHistorySortColumn
	}
}

func (a *Payment) PaymentFromDTO(dto *PaymentDTO) {
	a.UUID = dto.UUID
	a.ReferenceID = dto.ReferenceID
	a.MerchantID = dto.MerchantID
	a.CustomerID = dto.CustomerID
	a.PaymentMethodID = dto.PaymentMethodID
	a.ProcessorReferenceNumber = dto.ProcessorReferenceNumber
	a.Currency = dto.Currency
	a.Amount = dto.Amount
	a.Fee = dto.Fee
	a.Discount = dto.Discount
	a.TotalAmount = dto.TotalAmount
	a.Status = dto.Status
	a.Type = dto.Type
	a.PaymentURL = dto.PaymentURL
	a.ReasonType = dto.ReasonType
	a.ReasonDescription = dto.ReasonDescription
	a.CreatedAt = dto.CreatedAt
	a.UpdatedAt = dto.UpdatedAt
	a.DeletedAt = dto.DeletedAt
	a.ExpiredAt = dto.ExpiredAt

	if dto.Metadata != nil {
		var metadata map[string]interface{}
		json.Unmarshal([]byte(*dto.Metadata), &metadata)
		a.Metadata = &metadata
	}
}

func (a *Payment) ToDTO() *PaymentDTO {
	var metadata string
	if a.Metadata != nil {
		metadataJson, _ := json.Marshal(a.Metadata)
		metadata = string(metadataJson)
	}

	return &PaymentDTO{
		UUID:                     a.UUID,
		MerchantID:               a.MerchantID,
		ReferenceID:              a.ReferenceID,
		ProcessorReferenceNumber: a.ProcessorReferenceNumber,
		CustomerID:               a.CustomerID,
		PaymentMethodID:          a.PaymentMethodID,
		Currency:                 a.Currency,
		Amount:                   a.Amount,
		Fee:                      a.Fee,
		Discount:                 a.Discount,
		TotalAmount:              a.TotalAmount,
		Status:                   a.Status,
		Type:                     a.Type,
		Metadata:                 &metadata,
		PaymentURL:               a.PaymentURL,
		ReasonType:               a.ReasonType,
		ReasonDescription:        a.ReasonDescription,
		CreatedAt:                a.CreatedAt,
		UpdatedAt:                a.UpdatedAt,
		DeletedAt:                a.DeletedAt,
		ExpiredAt:                a.ExpiredAt,
	}
}

func (a *Payment) PaymentFromPaymentWithPaymentMethodDTO(dto *PaymentWithPaymentMethodDTO) {
	a.UUID = dto.UUID
	a.MerchantID = dto.MerchantID
	a.ReferenceID = dto.ReferenceID
	a.CustomerID = dto.CustomerID
	a.PaymentMethodID = dto.PaymentMethodID
	a.ProcessorReferenceNumber = dto.ProcessorReferenceNumber
	a.RecurringContractID = dto.RecurringContractID
	a.Currency = dto.Currency
	a.Amount = dto.Amount
	a.Fee = dto.Fee
	a.Discount = dto.Discount
	a.TotalAmount = dto.TotalAmount
	a.Status = dto.Status
	a.Type = dto.Type
	a.CreatedAt = dto.CreatedAt
	a.UpdatedAt = dto.UpdatedAt
	a.PaymentURL = dto.PaymentURL
	a.ReasonType = dto.ReasonType
	a.ReasonDescription = dto.ReasonDescription
	a.CreatedBy = dto.CreatedBy
	a.CreatedFrom = dto.CreatedFrom

	// Parse PaymentCaptures from raw JSON
	if dto.PaymentCapturesRaw.Valid {
		var paymentCaptures []*paymentCaptureModel.PaymentCapture
		if err := json.Unmarshal(dto.PaymentCapturesRaw.JSONText, &paymentCaptures); err == nil {
			// Sort descending as alternative sql issues.
			sort.Slice(paymentCaptures, func(i, j int) bool {
				return paymentCaptures[i].CreatedAt.After(paymentCaptures[j].CreatedAt)
			})

			a.PaymentCaptures = paymentCaptures
		}
	}

	if dto.DeletedAt.Valid {
		a.DeletedAt = &dto.DeletedAt.Time
	} else {
		a.DeletedAt = nil
	}

	if dto.ExpiredAt.Valid {
		a.ExpiredAt = &dto.ExpiredAt.Time
	} else {
		a.ExpiredAt = nil
	}

	if dto.InvestigationStartedAt.Valid {
		a.InvestigationStartedAt = &dto.InvestigationStartedAt.Time
	} else {
		a.InvestigationStartedAt = nil
	}

	if dto.Metadata != nil {
		var metadata map[string]interface{}
		json.Unmarshal([]byte(*dto.Metadata), &metadata)
		a.Metadata = &metadata
	}
	if dto.PaymentSnapCoreId != nil {
		a.SnapCoreId = dto.PaymentSnapCoreId
	}

	a.PaymentMethod = PaymentMethod{
		Type:     dto.PaymentMethodType.String,
		Name:     dto.PaymentMethodName.String,
		Acquirer: dto.PaymentMethodAcquirer.String,
		BankName: &dto.PaymentMethodBankName.String,
		Logo:     &dto.PaymentMethodLogo.String,
	}
}

func (a *Payment) UpdatePaymentFromCreditcardNotification(dto *creditcardModel.PaymentNotificationDataRequest) error {
	if dto.PaymentURL != "" {
		a.PaymentURL = dto.PaymentURL
	}

	a.Amount = dto.Amount
	a.Currency = dto.Currency
	a.ProcessorReferenceNumber = &dto.AcquirerTransactionID
	a.UpdatedAt = dto.Updated

	switch dto.PaymentStatus {
	case constant.CreditCardProcessorStatusSuccess:
		a.Status = constant.CreditCardStatusSuccess
	case constant.CreditCardProcessorStatusVoid:
		a.Status = constant.CreditCardStatusVoid
	case constant.CreditCardProcessorStatusFailed,
		constant.CreditCardProcessorStatusBlocked,
		constant.CreditCardProcessorStatusExpired:
		a.Status = constant.CreditCardStatusFailed
	case constant.CreditCardProcessorStatusRefunded:
		a.Status = constant.CreditCardStatusRefunded // To be discuss with product
	default:
		a.Status = constant.StatusPending
	}

	return nil
}

func (a *Payment) ToPaymentCreditCardCallbackRequest() (*pb.PaymentCreditCardCallbackRequest, error) {
	metadata, err := a.GetCreditcardMetadata()
	if err != nil {
		return nil, err
	}

	paymentStatus, err := a.GetCreditcardPaymentStatus()
	if err != nil {
		return nil, err
	}

	return &pb.PaymentCreditCardCallbackRequest{
		Created:        a.CreatedAt.Format(time.RFC3339),
		PaymentId:      a.UUID,
		MerchantId:     a.MerchantID,
		BankMerchantId: metadata.BankMerchantID,
		ReferenceId:    *a.ReferenceID,
		Amount:         a.Amount.String(),
		Currency:       a.Currency,
		PaymentUrl:     a.PaymentURL,
		CardData:       metadata.CardData.ToPaymentCreditCardDataRequest(),
		PaymentStatus:  paymentStatus,
		Updated:        a.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (a *Payment) GetTransactionStatus() string {
	switch a.Status {
	case constant.CreditCardStatusSuccess, constant.CreditCardStatusPAID:
		return constant.StatusSuccess
	case constant.CreditCardStatusVoid,
		constant.CreditCardStatusExpired,
		constant.CreditCardStatusFailed,
		constant.CreditCardStatusRefunded,
		constant.CreditCardStatusBlocked:
		return constant.StatusFailed
	default:
		return constant.StatusPending
	}
}

func (a *Payment) GetCreditcardMetadata() (*creditcardModel.CreditcardMetadata, error) {
	jsonData, err := json.Marshal(a.Metadata)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %s", err.Error())
	}

	var creditCardMetadata creditcardModel.CreditcardMetadata
	if err = json.Unmarshal(jsonData, &creditCardMetadata); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %s", err.Error())
	}

	return &creditCardMetadata, nil
}

func (a *Payment) GetCreditcardPaymentStatus() (string, error) {
	metadata, err := a.GetCreditcardMetadata()
	if err != nil {
		return "", err
	}

	var paymentStatus string
	switch metadata.ProcessorStatus {
	case constant.CreditCardProcessorStatusSuccess:
		paymentStatus = constant.CreditCardStatusPAID
	case constant.CreditCardProcessorStatusFailed:
		paymentStatus = constant.CreditCardStatusFailed
	case constant.CreditCardProcessorStatusBlocked:
		paymentStatus = constant.CreditCardStatusBlocked
	case constant.CreditCardProcessorStatusRefunded:
		paymentStatus = constant.CreditCardStatusRefunded
	case constant.CreditCardProcessorStatusExpired:
		paymentStatus = constant.CreditCardStatusExpired
	case constant.CreditCardProcessorStatusVoid:
		paymentStatus = constant.CreditCardStatusVoid
	default:
		paymentStatus = constant.CreditCardStatusWaitingForPayment
	}

	// Force processor status to using unified payment status
	if a.Status == constant.CreditCardStatusSuccess {
		paymentStatus = constant.CreditCardStatusPAID

	} else if a.Type != constant.TypeVirtualTerminal && a.Status == constant.UnifiedPaymentSessionStatusProcessing {
		paymentStatus = constant.CreditCardStatusProcessing
	}

	return paymentStatus, nil
}

func (a *Payment) ToOpenAPICreditcardGetPaymentByUUID() (*creditcardModel.OpenAPIGetCardPaymentByIdResponse, error) {
	var referenceID string
	if a.ReferenceID != nil {
		referenceID = *a.ReferenceID
	}

	metadata, err := a.GetCreditcardMetadata()
	if err != nil {
		err = fmt.Errorf(constant.ErrWhenGetCreditcardMetaData, err)
		return nil, err
	}

	paymentStatus, err := a.GetCreditcardPaymentStatus()
	if err != nil {
		err = fmt.Errorf(constant.ErrWhenCreditcardGetPaymentStatus, err)
		return nil, err
	}

	return &creditcardModel.OpenAPIGetCardPaymentByIdResponse{
		UUID:           a.UUID,
		MerchantID:     a.MerchantID,
		BankMerchantID: metadata.BankMerchantID,
		ReferenceID:    referenceID,
		PaymentStatus:  paymentStatus,
		Amount:         a.Amount,
		Currency:       a.Currency,
		PaymentURL:     a.PaymentURL,
		CardData:       metadata.CardData.ToSendCallbackCardDataRequest(),
		Created:        a.CreatedAt.Format(time.RFC3339),
		Updated:        a.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (a *Payment) ToInternalCreditcardGetPaymentByUUID() (*creditcardModel.InternalGetCardPaymentByIdResponse, error) {
	var referenceID string
	if a.ReferenceID != nil {
		referenceID = *a.ReferenceID
	}

	metadata, err := a.GetCreditcardMetadata()
	if err != nil {
		err = fmt.Errorf(constant.ErrWhenGetCreditcardMetaData, err)
		return nil, err
	}

	metadata.CardConfig = &creditcardModel.CreditCardConfig{}
	metadata.ProcessingConfig = &creditcardModel.CreditCardProcessingConfig{}
	var unifiedPaymentMetadata unifiedPaymentModel.MetadataUnifiedPayment
	metadataB, _ := json.Marshal(a.Metadata)
	_ = json.Unmarshal(metadataB, &unifiedPaymentMetadata)
	if unifiedPaymentMetadata.SaveForFutureUse != nil {
		metadata.CardConfig.SavedFutureUse = *unifiedPaymentMetadata.SaveForFutureUse
	}
	if unifiedPaymentMetadata.ShowSavedPayment != nil {
		metadata.CardConfig.ShowSavedPayment = *unifiedPaymentMetadata.ShowSavedPayment
	}
	if unifiedPaymentMetadata.PaymentMethodOptions.Card != nil {
		metadata.ThreeDsMethod = unifiedPaymentMetadata.PaymentMethodOptions.Card.ThreeDsMethod
		metadata.CaptureMethod = unifiedPaymentMetadata.PaymentMethodOptions.Card.CaptureMethod

		if unifiedPaymentMetadata.PaymentMethodOptions.Card.ProcessingConfig != nil {
			metadata.ProcessingConfig.BankMerchantId = unifiedPaymentMetadata.PaymentMethodOptions.Card.ProcessingConfig.BankMerchantId
			metadata.ProcessingConfig.MerchantIdTag = unifiedPaymentMetadata.PaymentMethodOptions.Card.ProcessingConfig.MerchantIdTag
		}
	}
	if autoSplit := unifiedPaymentMetadata.AutoSplitPayment; autoSplit != nil {
		a.AutoSplitPayment = autoSplit
		metadata.AutoSplitPayment = autoSplit.ToCardAutoSplitPayment()
	}
	paymentStatus, err := a.GetCreditcardPaymentStatus()
	if err != nil {
		err = fmt.Errorf(constant.ErrWhenCreditcardGetPaymentStatus, err)
		return nil, err
	}

	return &creditcardModel.InternalGetCardPaymentByIdResponse{
		UUID:           a.UUID,
		MerchantID:     a.MerchantID,
		BankMerchantID: metadata.BankMerchantID,
		ReferenceID:    referenceID,
		RecurringID:    a.RecurringContractID,
		PaymentType:    a.Type,
		PaymentStatus:  paymentStatus,
		Amount:         a.Amount,
		Fee:            a.Fee,
		Discount:       a.Discount,
		TotalAmount:    a.TotalAmount,
		Currency:       a.Currency,
		PaymentURL:     a.PaymentURL,
		Metadata:       metadata,
		ExpirationMode: unifiedPaymentMetadata.ExpirationMode,
		ExpiredAt:      util.ValueOfPtr(a.ExpiredAt),
		Created:        a.CreatedAt.Format(time.RFC3339),
		Updated:        a.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (a *Payment) ToUnifiedPaymentResponse() *unifiedPaymentModel.UnifiedPaymentSessionResponse {
	clientReferenceID := ""
	if a.ReferenceID != nil {
		clientReferenceID = *a.ReferenceID
	}
	amount, _ := a.Amount.Float64()

	var unifiedPaymentMetadata unifiedPaymentModel.MetadataUnifiedPayment
	metadataB, _ := json.Marshal(a.Metadata)
	_ = json.Unmarshal(metadataB, &unifiedPaymentMetadata)
	a.AutoSplitPayment = unifiedPaymentMetadata.AutoSplitPayment

	clientRedirectUrl := unifiedPaymentModel.RedirectUrl{}
	if unifiedPaymentMetadata.ClientRedirectUrl != nil {
		clientRedirectUrl = *unifiedPaymentMetadata.ClientRedirectUrl
	}

	expiredAt := time.Time{}
	if a.ExpiredAt != nil {
		expiredAt = *a.ExpiredAt
	}

	if unifiedPaymentMetadata.EncryptedEncryptionKey != nil {
		// GSM client is globally only set when needed such as in HTTP service, when not set then the process does not need data decryption.
		if gsmClient := gcp.GetGlobalSecretManagerClient(); gsmClient != nil {

			gcpConfig, secret := config.GetGCPConfig(), commonModel.EncryptionSecret{}

			_ = gsmClient.GetSecretValueJSON(
				context.Background(), gcpConfig.ProjectId, gcpConfig.SecretManager.EncryptionSecretName, unifiedPaymentMetadata.EncryptedEncryptionKey.GetSecretVersion(), &secret,
			)
			_ = unifiedPaymentMetadata.EncryptedEncryptionKey.ParseToResponse(secret.Payment.KeyEncryptionKey)

		} else {
			// If the decryption process is not required, the data will be removed (from the response) to prevent the disclosure of the encrypted encryption key.
			unifiedPaymentMetadata.EncryptedEncryptionKey = nil
		}
	}

	resp := &unifiedPaymentModel.UnifiedPaymentSessionResponse{
		ID:                a.UUID,
		ClientReferenceID: clientReferenceID,
		Amount: unifiedPaymentModel.Amount{
			Currency: a.Currency,
			Value:    amount,
		},
		AutoConfirm:                unifiedPaymentMetadata.AutoConfirm,
		Mode:                       unifiedPaymentMetadata.Mode,
		RedirectUrl:                clientRedirectUrl,
		PaymentType:                a.Type,
		PaymentMethod:              unifiedPaymentMetadata.PaymentMethod,
		PaymentMethodOptions:       unifiedPaymentMetadata.PaymentMethodOptions,
		StatementDescriptor:        unifiedPaymentMetadata.StatementDescriptor,
		SplitRoutingConfigurations: unifiedPaymentMetadata.SplitRoutingConfigurations,
		SaveForFutureUse:           unifiedPaymentMetadata.SaveForFutureUse,
		ShowSavedPayment:           unifiedPaymentMetadata.ShowSavedPayment,
		ExpirationMode:             unifiedPaymentMetadata.ExpirationMode,
		Status:                     a.Status,
		CreatedAt:                  a.CreatedAt,
		UpdatedAt:                  a.UpdatedAt,
		PaymentUrl:                 a.PaymentURL,
		Metadata:                   unifiedPaymentMetadata.ClientMetadata,
		EncryptionKey:              unifiedPaymentMetadata.EncryptedEncryptionKey,
		RecurringID:                util.ValueOfPtr(a.RecurringContractID),
		AutoSplitPayment:           unifiedPaymentMetadata.AutoSplitPayment,
	}

	if a.ExpiredAt != nil {
		resp.ExpiryAt = &expiredAt
	}

	if a.ReasonType != nil {
		resp.InvestigationStatus = a.ReasonType
	}

	// exclude existing payment url clearance for ewallet dana
	if unifiedPaymentMetadata.Mode == constant.UnifiedPaymentModeAPI &&
		!(a.PaymentMethod.Type == constant.UnifiedPaymentMethodEWallet && strings.EqualFold(a.PaymentMethod.Acquirer, constant.UnifiedPaymentEWalletDanaAcquirer)) {
		resp.PaymentUrl = ""
	}

	if unifiedPaymentMetadata.CanceledAt != nil {
		resp.CancelledAt = unifiedPaymentMetadata.CanceledAt
	}
	if unifiedPaymentMetadata.CancellationReason != "" {
		resp.CancellationReason = unifiedPaymentMetadata.CancellationReason
	}
	if unifiedPaymentMetadata.MethodDetail != nil {
		if unifiedPaymentMetadata.MethodDetail.Qr != nil {
			resp.PaymentMethod.QrPaymentMethodDetail = unifiedPaymentMetadata.MethodDetail.Qr
		}

		if unifiedPaymentMetadata.MethodDetail.VirtualAccount != nil {
			resp.PaymentMethod.VAPaymentMethodDetail = unifiedPaymentMetadata.MethodDetail.VirtualAccount
		}
	}
	if slices.Contains([]string{
		constant.UnifiedPaymentSessionStatusRequireConfirmation,
		constant.UnifiedPaymentSessionStatusRequirePaymentMethod,
	}, resp.Status) {
		resp.PaymentUrl = ""
	}
	if unifiedPaymentMetadata.RecurringPayment != nil {
		resp.InitiateFirstAuthorization = unifiedPaymentMetadata.RecurringPayment.InitiateFirstAuthorization
		resp.FirstAuthorizationMethod = unifiedPaymentMetadata.RecurringPayment.FirstAuthorizationMethod
		resp.FirstAuthorizationOrderID = unifiedPaymentMetadata.RecurringPayment.FirstAuthorizationOrderID
		resp.RecurringBillingCycle = unifiedPaymentMetadata.RecurringPayment.BillingCycle
	}
	if cardFunded := unifiedPaymentMetadata.CardFundedPayout; cardFunded != nil {
		resp.CardFundedPayout = &unifiedPaymentModel.CardFundedPayout{
			FirstPaymentID:   cardFunded.FirstPaymentID,
			SettlementMethod: cardFunded.SettlementMethod,
			Sequence:         cardFunded.Sequence,
			FeeAmount:        util.ValueOfPtr(a.Fee).InexactFloat64(),
			FeeConfig:        util.ValueOfPtr(unifiedPaymentMetadata.FeeDetail),
			CardID:           cardFunded.CardID,
		}
	}

	return resp
}

func (a *Payment) ToUnifiedPaymentAndChargeResponse(charge *orchestrator_model.AccountTransactionWithUseCase) *unifiedPaymentModel.UnifiedPaymentSessionResponse {
	resp := a.ToUnifiedPaymentResponse()
	a.AutoSplitPayment = resp.AutoSplitPayment
	if charge == nil {
		return resp
	}

	chargeStatus := ""
	statementDescriptor := ""
	chargeMethodDetails := &unifiedPaymentModel.ChargePaymentMethodDetails{}
	_ = json.Unmarshal(charge.AdditionalInfo.JSONText, &struct {
		ChargeStatus        *string     `json:"chargeStatus"`
		StatementDescriptor *string     `json:"statementDescriptor"`
		MethodDetail        interface{} `json:"methodDetail"`
	}{
		MethodDetail:        chargeMethodDetails,
		ChargeStatus:        &chargeStatus,
		StatementDescriptor: &statementDescriptor,
	})

	// Extract FDS risk assessment from additional_info
	var fdsRiskAssessment *fdsCommonModel.FdsRiskAssessment
	if charge.AdditionalInfo.Valid {
		var additionalInfo map[string]interface{}
		if err := json.Unmarshal(charge.AdditionalInfo.JSONText, &additionalInfo); err == nil {
			if fdsData, exists := additionalInfo["fdsRiskAssessment"]; exists {
				// Convert the fdsData to JSON and then unmarshal to FdsRiskAssessment
				if fdsBytes, err := json.Marshal(fdsData); err == nil {
					var fdsAssessment fdsCommonModel.FdsRiskAssessment
					if err := json.Unmarshal(fdsBytes, &fdsAssessment); err == nil {
						fdsRiskAssessment = &fdsAssessment
					}
				}
			}
		}
	}

	chargeResp := &unifiedPaymentModel.ChargeResponse{
		ID:                              charge.UUID.String(),
		PaymentSessionID:                resp.ID,
		PaymentSessionClientReferenceID: resp.ClientReferenceID,
		Amount: unifiedPaymentModel.Amount{
			Currency: charge.Currency,
			Value:    charge.Credit,
		},
		StatementDescriptor:        statementDescriptor,
		Status:                     chargeStatus,
		AuthorizedAmount:           nil,
		CapturedAmount:             nil,
		IsCaptured:                 false,
		CreatedAt:                  charge.CreatedAt,
		UpdatedAt:                  charge.UpdatedAt,
		PaidAt:                     nil,
		FdsRiskAssessment:          fdsRiskAssessment,
		ChargePaymentMethodDetails: chargeMethodDetails,
	}

	if chargeStatus == constant.ChargeStatusSuccess {
		chargeResp.PaidAt = &charge.TransactionTimestamp
	}

	if a.IsFinalStatus() &&
		a.GetAutoSplitTotalSuccessAmount() != nil {
		chargeResp.Amount.Value = a.GetAutoSplitTotalSuccessAmount().ToDecimal().InexactFloat64()
	}

	chargeResp.SetFailureDetail()
	switch chargeStatus {
	case constant.ChargeStatusSuccess:
		chargeResp.IsCaptured = true
		chargeResp.AuthorizedAmount = &resp.Amount
		chargeResp.CapturedAmount = &chargeResp.Amount
	case constant.ChargeStatusWaitingForCapture:
		chargeResp.CapturedAmount = &chargeResp.Amount
		chargeResp.AuthorizedAmount = &resp.Amount
	}

	// Set captureHistories to the chargeResp if any
	chargeResp.SetCaptureHistories(a.PaymentCaptures)
	chargeResp.SettlementStatus = charge.SettlementStatus.String

	resp.ChargeDetails = append(resp.ChargeDetails, chargeResp)

	resp.SetPaymentURLForAPIMode()
	for _, chargeRespTemp := range resp.ChargeDetails {
		chargeRespTemp.RemoveUnusedResponse()
	}

	return resp
}

func (a *Payment) ToPbUnifiedPaymentV2CallbackRequest(charge *orchestrator_model.AccountTransactionWithUseCase, customer *unifiedPaymentModel.CustomerInformationResponse) *pb.UnifiedPaymentV2CallbackRequest {
	var expiredAt *timestamppb.Timestamp
	if a.ExpiredAt != nil {
		expiredAt = timestamppb.New(*a.ExpiredAt)
	}

	referenceId := ""
	if a.ReferenceID != nil {
		referenceId = *a.ReferenceID
	}

	var unifiedPaymentMetadata unifiedPaymentModel.MetadataUnifiedPayment
	metadataB, _ := json.Marshal(a.Metadata)
	_ = json.Unmarshal(metadataB, &unifiedPaymentMetadata)

	clientRedirectUrl := unifiedPaymentModel.RedirectUrl{}
	if unifiedPaymentMetadata.ClientRedirectUrl != nil {
		clientRedirectUrl = *unifiedPaymentMetadata.ClientRedirectUrl
	}

	callbackRequest := &pb.UnifiedPaymentV2CallbackRequest{
		Id:                a.UUID,
		ClientReferenceId: referenceId,
		Amount: &pb.UnifiedPaymentV2CallbackRequest_Amount{
			Currency: a.Currency,
			Value:    a.Amount.InexactFloat64(),
		},
		AutoConfirm: unifiedPaymentMetadata.AutoConfirm,
		Mode:        unifiedPaymentMetadata.Mode,
		PaymentMethod: &pb.UnifiedPaymentV2CallbackRequest_PaymentMethod{
			Type: unifiedPaymentMetadata.PaymentMethod.Type,
		},
		RedirectUrl: &pb.UnifiedPaymentV2CallbackRequest_RedirectUrl{
			SuccessReturnUrl:    clientRedirectUrl.SuccessReturnUrl,
			FailureReturnUrl:    clientRedirectUrl.FailureReturnUrl,
			ExpirationReturnUrl: clientRedirectUrl.ExpirationReturnUrl,
		},
		Status:               a.Status,
		ExpiryAt:             expiredAt,
		CreatedAt:            timestamppb.New(a.CreatedAt),
		UpdatedAt:            timestamppb.New(a.UpdatedAt),
		PaymentUrl:           a.PaymentURL,
		StatementDescriptor:  unifiedPaymentMetadata.StatementDescriptor,
		PaymentMethodOptions: &pb.UnifiedPaymentV2CallbackRequest_PaymentMethodOptions{},
		Metadata:             &anypb.Any{},
	}

	if unifiedPaymentMetadata.CanceledAt != nil {
		callbackRequest.CancelledAt = timestamppb.New(*unifiedPaymentMetadata.CanceledAt)
	}
	if unifiedPaymentMetadata.CancellationReason != "" {
		callbackRequest.CancellationReason = unifiedPaymentMetadata.CancellationReason
	}
	if unifiedPaymentMetadata.SaveForFutureUse != nil {
		callbackRequest.SaveForFutureUse = unifiedPaymentMetadata.SaveForFutureUse
	}
	if unifiedPaymentMetadata.ShowSavedPayment != nil {
		callbackRequest.ShowSavedPayment = unifiedPaymentMetadata.ShowSavedPayment
	}
	if a.ReasonType != nil {
		callbackRequest.InvestigationStatus = a.ReasonType
	}
	if customer != nil {
		customerInfo := &pb.CustomerInformation{
			CustomerId: customer.CustomerID,
			GivenName:  customer.GivenName,
			SureName:   customer.SureName,
			Email:      customer.Email,
		}

		// only set Surname if it's not nil
		if customer.Surname != nil {
			customerInfo.Surname = customer.Surname
		}
		callbackRequest.Customer = customerInfo

		if customer.PhoneNumber != nil {
			callbackRequest.Customer.PhoneNumber = &pb.UnifiedPaymentPhoneNumber{
				Number:      customer.PhoneNumber.Number,
				CountryCode: customer.PhoneNumber.CountryCode,
			}
		}

		if customer.RefundPreference != nil {
			callbackRequest.Customer.RefundPreference = &pb.UnifiedPaymentRefundPreference{
				Method: customer.RefundPreference.Method,
			}

			if customer.RefundPreference.TransferDestination != nil {
				callbackRequest.Customer.RefundPreference.TransferDestination = &pb.RefundTransferDestination{
					ChannelCode: customer.RefundPreference.TransferDestination.ChannelCode,
				}
			}
			if customer.RefundPreference.TransferDestination != nil && customer.RefundPreference.TransferDestination.ChannelInformation != nil {
				callbackRequest.Customer.RefundPreference.TransferDestination.ChannelInformation = &pb.RefundChannelInformation{
					AccountNumber: customer.RefundPreference.TransferDestination.ChannelInformation.AccountNumber,
					AccountName:   customer.RefundPreference.TransferDestination.ChannelInformation.AccountName,
				}
			}
		}

		var storedPaymentMethods []*pb.CustomerPaymentMethod
		for _, method := range customer.StoredPaymentMethods {
			customerPaymentMethod := &pb.CustomerPaymentMethod{
				Token:          method.Token,
				PaymentMethod:  method.PaymentMethod,
				PaymentChannel: method.PaymentChannel,
				Status:         method.Status,
				CreatedAt:      timestamppb.New(method.CreatedAt),
			}

			if method.Card != nil {
				customerPaymentMethod.Card = &pb.CustomerPaymentMethodCard{
					Network:             method.Card.Network,
					First6:              method.Card.First6,
					First8:              method.Card.First8,
					Last4:               method.Card.Last4,
					ExpMonth:            method.Card.ExpMonth.(string),
					ExpYear:             method.Card.ExpYear.(string),
					CardHolderFirstName: method.Card.CardHolderFirstName,
					CardHolderLastName:  method.Card.CardHolderLastName,
					CardHolderEmail:     method.Card.CardHolderEmail,
					CardHolderPhone:     method.Card.CardHolderPhone,
				}
			}

			storedPaymentMethods = append(storedPaymentMethods, customerPaymentMethod)
		}
		callbackRequest.Customer.StoredPaymentMethods = storedPaymentMethods
	}

	if unifiedPaymentMetadata.AutoSplitPayment != nil && unifiedPaymentMetadata.AutoSplitPayment.Summary != nil {
		result := unifiedPaymentMetadata.AutoSplitPayment.Summary.ToAutoSplitDetail()
		chargeResults := []*pb.UnifiedPaymentV2CallbackRequest_ChargeResponse{}
		callbackRequest.AutoSplitDetails = &pb.AutoSplitDetails{
			Status: result.Status,
			TotalSuccessfulChargeAmount: &pb.UnifiedPaymentV2CallbackRequest_Amount{
				Value:    result.TotalSuccessfulChargeAmount.Value,
				Currency: result.TotalSuccessfulChargeAmount.Currency,
			},
			TotalFailedChargeAmount: &pb.UnifiedPaymentV2CallbackRequest_Amount{
				Value:    result.TotalFailedChargeAmount.Value,
				Currency: result.TotalFailedChargeAmount.Currency,
			},
			TotalInProcessChargeAmount: &pb.UnifiedPaymentV2CallbackRequest_Amount{
				Value:    result.TotalInProcessChargeAmount.Value,
				Currency: result.TotalInProcessChargeAmount.Currency,
			},
			NumberOfCharges:           int32(result.NumberOfCharges),
			NumberOfSuccessfulCharges: int32(result.NumberOfSuccessfulCharges),
			NumberOfInProcessCharges:  int32(result.NumberOfInProcessCharges),
			NumberOfFailedCharges:     int32(result.NumberOfFailedCharges),
		}

		for _, charge := range result.ChargesDetails {
			chargeResults = append(chargeResults, charge.ToPbChargeResponse())
		}

		callbackRequest.AutoSplitDetails.ChargesDetails = chargeResults

		// override the credit amount following the total succeeded auto split amout
		charge.Credit = result.TotalSuccessfulChargeAmount.Value
	}

	switch a.PaymentMethod.Type {
	case paymentConst.PAYMENT_METHOD_VIRTUAL_ACCOUNT:
		if unifiedPaymentMetadata.PaymentMethodOptions.VirtualAccount != nil {
			callbackRequest.PaymentMethodOptions.VirtualAccount = &pb.UnifiedPaymentV2CallbackRequest_PaymentMethodOptionVirtualAccount{
				Channel:              unifiedPaymentMetadata.PaymentMethodOptions.VirtualAccount.Channel,
				VirtualAccountNumber: unifiedPaymentMetadata.PaymentMethodOptions.VirtualAccount.VirtualAccountNumber,
				VirtualAccountName:   unifiedPaymentMetadata.PaymentMethodOptions.VirtualAccount.VirtualAccountName,
			}

			if unifiedPaymentMetadata.PaymentMethodOptions.VirtualAccount.ExpiryAt != nil {
				callbackRequest.PaymentMethodOptions.VirtualAccount.ExpiryAt = timestamppb.New(*unifiedPaymentMetadata.PaymentMethodOptions.VirtualAccount.ExpiryAt)
			}
		}

	case paymentConst.PAYMENT_METHOD_QRIS:
		if unifiedPaymentMetadata.PaymentMethodOptions.QR != nil {
			callbackRequest.PaymentMethodOptions.Qr = &pb.UnifiedPaymentV2CallbackRequest_PaymentMethodOptionQR{}

			if unifiedPaymentMetadata.PaymentMethodOptions.QR.ExpiryAt != nil {
				callbackRequest.PaymentMethodOptions.Qr.ExpiryAt = timestamppb.New(*unifiedPaymentMetadata.PaymentMethodOptions.QR.ExpiryAt)
			}
		}

	case paymentConst.PAYMENT_METHOD_CREDIT_CARD:
		if unifiedPaymentMetadata.PaymentMethodOptions.Card != nil {
			callbackRequest.PaymentMethodOptions.Card = &pb.UnifiedPaymentV2CallbackRequest_PaymentMethodOptionCard{
				CaptureMethod: unifiedPaymentMetadata.PaymentMethodOptions.Card.CaptureMethod,
				ThreeDsMethod: unifiedPaymentMetadata.PaymentMethodOptions.Card.ThreeDsMethod,
			}

			if unifiedPaymentMetadata.PaymentMethodOptions.Card.ProcessingConfig != nil {
				callbackRequest.PaymentMethodOptions.Card.ProcessingConfig = &pb.UnifiedPaymentV2CallbackRequest_PaymentMethodOptionCardProcessingConfig{
					BankMerchantId: unifiedPaymentMetadata.PaymentMethodOptions.Card.ProcessingConfig.BankMerchantId,
					MerchantIdTag:  unifiedPaymentMetadata.PaymentMethodOptions.Card.ProcessingConfig.MerchantIdTag,
				}
			}
			if unifiedPaymentMetadata.PaymentMethodOptions.Card.Installment != nil {
				callbackRequest.PaymentMethodOptions.Card.Installment = &pb.UnifiedPaymentV2CallbackRequest_PaymentMethodOptionCardInstallment{
					Enabled: unifiedPaymentMetadata.PaymentMethodOptions.Card.Installment.Enabled,
					// AvailablePlans TODO: Implement this
					// Plan TODO: Implement this
				}
			}
		}
	case paymentConst.PAYMENT_METHOD_EWALLET:
		if unifiedPaymentMetadata.PaymentMethodOptions.Ewallet != nil {
			callbackRequest.PaymentMethodOptions.Ewallet = &pb.UnifiedPaymentV2CallbackRequest_PaymentMethodOptionEwallet{
				Channel: unifiedPaymentMetadata.PaymentMethodOptions.Ewallet.Channel,
			}
		}
	}

	if unifiedPaymentMetadata.ClientMetadata != nil {
		clientMetadata, err := parseMetadata(unifiedPaymentMetadata.ClientMetadata)
		if err != nil {
			clientMetadata = map[string]interface{}{}
		}

		structPB, _ := structpb.NewStruct(clientMetadata)
		callbackRequest.Metadata, _ = anypb.New(structPB)
	}

	if charge == nil {
		return callbackRequest
	}

	chargeMethodDetails := &unifiedPaymentModel.ChargePaymentMethodDetails{}
	_ = json.Unmarshal(charge.AdditionalInfo.JSONText, &struct {
		MethodDetail interface{} `json:"methodDetail"`
	}{
		MethodDetail: chargeMethodDetails,
	})
	if callbackRequest.Mode == constant.UnifiedPaymentModeAPI {
		if chargeMethodDetails.Card != nil {
			callbackRequest.PaymentUrl = chargeMethodDetails.Card.ACSURL
		} else if chargeMethodDetails.Ewallet != nil {
			callbackRequest.PaymentUrl = chargeMethodDetails.Ewallet.WebRedirectURL
			if chargeMethodDetails.Ewallet.Channel == constant.UnifiedPaymentEWalletShopeePayAcquirer {
				callbackRequest.PaymentUrl = chargeMethodDetails.Ewallet.AppRedirectURL
			}
		}
	}

	chargeResponse := unifiedPaymentModel.AccountTransactionToChargeResponse(charge)
	if slices.Contains([]string{constant.ChargeStatusSuccess, constant.ChargeStatusWaitingForCapture}, chargeResponse.Status) {
		chargeResponse.SetAuthorizedAmount(&unifiedPaymentModel.Amount{
			Currency: charge.Currency,
			Value:    a.Amount.InexactFloat64(),
		})
	}

	// Set captureHistories to the chargeResp if any
	chargeResponse.SetCaptureHistories(a.PaymentCaptures)

	pbChargeResponse := chargeResponse.ToPbChargeResponse()

	callbackRequest.ChargeDetails = append(callbackRequest.ChargeDetails, pbChargeResponse)

	return callbackRequest
}

func (a *Payment) GetOnBehalfParentID() string {
	if a.Metadata == nil {
		return ""
	}

	// need to handle the metadata struct wisely
	if onBehalf, ok := (*a.Metadata)["onBehalf"].(map[string]any); ok {
		parentMerchantId, _ := onBehalf["parentMerchantId"].(string)
		return parentMerchantId
	}

	return ""
}

func (p *Payment) IsFeeExempt() bool {
	if p.Type == constant.UnifiedPaymentOneDollarAuthorization {
		return true
	}
	if p.RecurringPayment == nil {
		return false
	}

	isZeroAmountRecurringPayment := !p.RecurringPayment.InitiateFirstAuthorization && p.Amount.IsZero()
	isFirstAuthMethodOneDollarRecurringPayment := p.RecurringPayment.InitiateFirstAuthorization && p.RecurringPayment.FirstAuthorizationMethod == constant.RecurringContractAuthMethodOneDollar

	return isFirstAuthMethodOneDollarRecurringPayment || isZeroAmountRecurringPayment
}

type UpdatePaymentMetadataRequest struct {
	FeeDetail          any
	FeeOnBehalf        any
	SummaryTransaction any
	FingerprintID      any
	IsSnap             any
}

type InvestigationPoPMetadata struct {
	Bucket        string `json:"bucket"`
	Path          string `json:"path"`
	MerchantNotes string `json:"merchantNotes"`
}

type UpdatePaymentForInvestigationRequest struct {
	PaymentID             string
	MerchantID            string
	ReasonType            string
	StartedAt             time.Time
	InvestigationMetadata InvestigationPoPMetadata
}

func parseMetadata(metadata interface{}) (map[string]interface{}, error) {
	metadataMap := make(map[string]interface{})
	if metadata == nil {
		return metadataMap, nil
	}

	bytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(bytes, &metadataMap)
	return metadataMap, nil
}

func (p *Payment) ToUnifiedPaymentMetadata() *unifiedPaymentModel.MetadataUnifiedPayment {
	if p == nil {
		return nil
	}

	metadata := unifiedPaymentModel.MetadataUnifiedPayment{}

	// convert to json
	metadataStr, err := json.Marshal(p.Metadata)
	if err != nil {
		return nil
	}

	if err := json.Unmarshal(metadataStr, &metadata); err != nil {
		return nil
	}

	return &metadata
}

type UpdatePaymentStatusWithReasonRequest struct {
	Status            string
	ReasonType        *string
	ReasonDescription *string
}

func (p *Payment) GetOneDollarAuthorizationUseCase() string {
	if p.Type != constant.UnifiedPaymentOneDollarAuthorization {
		return ""
	}

	unifiedPaymentMetadata := p.ToUnifiedPaymentMetadata()
	if unifiedPaymentMetadata != nil && unifiedPaymentMetadata.OneDollarAuthorization != nil {
		return unifiedPaymentMetadata.OneDollarAuthorization.UseCase
	}

	return ""
}

// Ensure metadata parsing has been performed and the value has been set in the AutoSplitPayment attribute.
func (p *Payment) IsAutoSplitPaymentAuth() bool {
	return p.AutoSplitPayment != nil &&
		p.AutoSplitPayment.TransactionType == constant.AutoSplitPaymentTypeAuthentication
}

func (p *Payment) IsAutoSplitPaymentFirstPayment() bool {
	return p.AutoSplitPayment != nil &&
		p.AutoSplitPayment.TransactionType == constant.AutoSplitPaymentTypeFirstPayment
}

func (p *Payment) IsAutoSplitSubPayments() bool {
	return p.AutoSplitPayment != nil &&
		util.Contains([]string{
			constant.AutoSplitPaymentTypeFirstPayment,
			constant.AutoSplitPaymentTypeSubsequentPayment,
		}, p.AutoSplitPayment.TransactionType)
}

func (p *Payment) GetAutoSplitTotalSuccessAmount() *commonModel.Amount {
	if p.AutoSplitPayment == nil {
		if p.Metadata == nil {
			return nil
		}

		if autoSplitMetadata, ok := (*p.Metadata)["autoSplitPayment"].(*unifiedPaymentModel.AutoSplitPayment); ok && autoSplitMetadata != nil {
			p.AutoSplitPayment = autoSplitMetadata
		}
	}

	if !p.IsAutoSplitPaymentAuth() ||
		p.AutoSplitPayment.Summary == nil {
		return nil
	}

	return &p.AutoSplitPayment.Summary.TotalSuccessfulChargeAmount
}

// IsFinalStatus will check is payment has authentication type or not
// parent payment always has this type
func (p *Payment) IsFinalStatus() bool {
	switch p.Status {
	case constant.UnifiedPaymentSessionStatusPaid, constant.UnifiedPaymentSessionStatusCancelled, constant.UnifiedPaymentSessionStatusExpired, paymentConst.PAYMENT_STATUS_SUCCESS, paymentConst.PaymentStatusFailed:
		return true
	default:
		return false
	}
}

// SetStatusByAutoSplitStatus sets the payment status based on the auto-split payment status.
// Successful or partially successful auto-split statuses map to UnifiedPaymentSessionStatusPaid,
// while cancelled or failed statuses map to UnifiedPaymentSessionStatusCancelled.
func (p *Payment) SetStatusByAutoSplitStatus(status string) {
	switch status {
	case constant.AutoSplitPaymentStatusSuccess, constant.AutoSplitPaymentStatusPartialSuccess:
		p.Status = constant.UnifiedPaymentSessionStatusPaid
	case constant.AutoSplitPaymentStatusCancelled, constant.AutoSplitPaymentStatusFailed:
		p.Status = constant.UnifiedPaymentSessionStatusCancelled
	}
}

// GetLedgerStatus maps the payment status to a ledger-compatible status.
// It returns constant.StatusSuccess for paid/success statuses,
// constant.StatusFailed for cancelled/failed/expired statuses,
// and constant.StatusPending for all other statuses.
func (p *Payment) GetLedgerStatus() string {
	switch p.Status {
	case constant.UnifiedPaymentSessionStatusPaid, constant.StatusSuccess:
		return constant.StatusSuccess
	case constant.UnifiedPaymentSessionStatusCancelled, constant.StatusFailed, constant.UnifiedPaymentSessionStatusExpired:
		return constant.StatusFailed
	}

	return constant.StatusPending
}
