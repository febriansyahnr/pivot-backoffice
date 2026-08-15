package paymentModel

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	snapVa "github.com/paper-indonesia/pdk/go/snap/structs/va"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	constantPayment "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	card "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/creditcard"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/customer"
	fdsCommonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/fdsProcessor/fdsCommon"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/orchestrator"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/proto/messages/callback"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/refund"
	snapCoreModelQr "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/qr"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/virtualAccount"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PaymentResponse struct {
	UUID              string                            `json:"uuid"`
	MerchantID        string                            `json:"merchantId"`
	ReferenceID       string                            `json:"referenceId"`
	Customer          *PaymentRequestCustomer           `json:"customer,omitempty"`
	Status            string                            `json:"status"`
	PaidAmount        *commonModel.Amount               `json:"paidAmount,omitempty"`
	TotalAmount       *Amount                           `json:"totalAmount,omitempty"`
	PaymentMethodId   string                            `json:"paymentMethodId,omitempty"`
	PaymentMethod     string                            `json:"paymentMethod,omitempty"`
	VirtualAccount    *PaymentVirtualAccountResponse    `json:"virtualAccount,omitempty"`
	Qris              *PaymentQrisResponse              `json:"qris,omitempty"`
	PaymentItems      *[]PaymentResponseItem            `json:"paymentItems,omitempty"`
	TransactionDate   *time.Time                        `json:"transactionDate,omitempty"`
	LastUpdateDate    *time.Time                        `json:"lastUpdateDate,omitempty"`
	MerchantName      string                            `json:"merchantName,omitempty"`
	AdditionalInfo    *map[string]any                   `json:"additionalInfo,omitempty"`
	FdsRiskAssessment *fdsCommonModel.FdsRiskAssessment `json:"fdsRiskAssessment,omitempty"`

	CreatedAt        time.Time  `json:"-"`
	ExpiredAt        *time.Time `json:"-"`
	PaymentURL       string     `json:"-"`
	IsUnifiedPayment bool       `json:"-"`
	PaymentType      string     `json:"-"`
}

type PaymentVirtualAccountResponse struct {
	Issuer                string               `json:"issuer"`
	VirtualAccountTrxType string               `json:"virtualAccountTrxType"`
	VirtualAccountNumber  string               `json:"virtualAccountNumber"`
	VirtualAccountName    string               `json:"virtualAccountName"`
	MinAmount             *Amount              `json:"minAmount"`
	MaxAmount             *Amount              `json:"maxAmount"`
	ExpiredDate           *time.Time           `json:"expiredDate"`
	IsSnap                bool                 `json:"isSnap,omitempty"`
	BillDetails           *[]snapVa.BillDetail `json:"billDetails,omitempty"`
	Metadata              map[string]any       `json:"metadata,omitempty"`
}

type PaymentQrisResponse struct {
	SubMerchantId      string                `json:"subMerchantId,omitempty"`
	ReferenceNo        string                `json:"referenceNo"`
	PartnerReferenceNo string                `json:"partnerReferenceNo,omitempty"`
	MerchantName       string                `json:"merchantName"`
	QrContent          string                `json:"qrContent,omitempty"`
	QrUrl              string                `json:"qrUrl,omitempty"`
	QrImage            *string               `json:"qrImage,omitempty"`
	QrType             string                `json:"qrType"`
	QrStatus           string                `json:"qrStatus"`
	QrExpiredDate      string                `json:"qrExpiredDate"`
	ValidityPeriod     *int                  `json:"validityPeriod,omitempty"`
	PaymentStatus      string                `json:"paymentStatus"`
	Amount             *Amount               `json:"amount,omitempty"`
	Acquirer           string                `json:"acquirer,omitempty"`
	TransactionDate    string                `json:"transactionDate"`
	DetailData         *[]QrStaticDetailData `json:"detailData"`

	Metadata map[string]any `json:"metadata,omitempty"`
}

type QrStaticDetailData struct {
	Amount         commonModel.Amount               `json:"amount,omitempty"`
	Status         string                           `json:"status,omitempty"`
	Type           string                           `json:"type,omitempty"`
	AdditionalInfo QrStaticDetailDataAdditionalInfo `json:"additionalInfo,omitempty"`
}

type QrStaticDetailDataAdditionalInfo struct {
	RRN             string `json:"RRN,omitempty"`
	TransactionDate string `json:"transactionDate,omitempty"`
}

type SnapPaymentVirtualAccountResponse struct {
	TrxId               string                       `json:"trxId"`
	VirtualAccountEmail string                       `json:"virtualAccountEmail,omitempty"`
	VirtualAccountPhone string                       `json:"virtualAccountPhone,omitempty"`
	VirtualAccountNo    string                       `json:"virtualAccountNo"`
	VirtualAccountName  string                       `json:"virtualAccountName"`
	PaidAmount          *commonModel.Amount          `json:"paidAmount"`
	TotalAmount         *commonModel.Amount          `json:"totalAmount,omitempty"`
	TrxDateTime         string                       `json:"trxDateTime,omitempty"`
	AdditionalInfo      *SnapVAPaymentAdditionalInfo `json:"additionalInfo,omitempty"`
}

type SnapVAPaymentAdditionalInfo struct {
	Event                 string          `json:"event,omitempty"`
	MerchantId            string          `json:"merchantId,omitempty"`
	ReferenceId           string          `json:"referenceId"`
	Customer              *SnapVACustomer `json:"customer,omitempty"`
	Issuer                string          `json:"issuer"`
	VirtualAccountTrxType string          `json:"virtualAccountTrxType"`
	ExpiredDate           string          `json:"expiredDate,omitempty"`
	MinAmount             *Amount         `json:"minAmount,omitempty"`
	MaxAmount             *Amount         `json:"maxAmount,omitempty"`
	VaStatus              string          `json:"vaStatus"`
	PaymentStatus         string          `json:"paymentStatus"`
	BankReferenceId       string          `json:"bankReferenceId,omitempty"`
}

type SnapVACustomer struct {
	CustomerID string `json:"customerId,omitempty"`
	Name       string `json:"name,omitempty"`
}

type PaymentResponseItem struct {
	ItemID      string `json:"itemId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Amount      Amount `json:"amount"`
	Qty         int    `json:"qty"`
}

type PaymentInsightResponse struct {
	TotalBalance      *commonModel.Amount `json:"totalBalance"`
	TodayTotalSuccess *PaymentInsightItem `json:"todayTotalSuccess"`
	TodayTotalPending *PaymentInsightItem `json:"todayTotalPending"`
	TodayTotalVoid    *PaymentInsightItem `json:"todayTotalVoid"`
}

type PaymentHistoryDetailResponse struct {
	UUID                   string                                             `json:"uuid"`
	MerchantID             string                                             `json:"merchantId"`
	CustomerID             string                                             `json:"customerId"`
	ReferenceID            string                                             `json:"referenceId"`
	RecurringID            string                                             `json:"recurringId,omitempty"`
	PaymentMethod          string                                             `json:"paymentMethod" example:"VIRTUAL_ACCOUNT, QRIS, CREDIT_CARD"`
	PaymentMethodCategory  string                                             `json:"paymentMethodType" example:"close dynamic, open static, closed static, static, dynamic, debit, credit"`
	ProcessorRefNumber     string                                             `json:"processorReferenceNumber"`
	TypeDetail             PaymentTypeDetail                                  `json:"paymentTypeDetail"`
	Channel                string                                             `json:"paymentChannel"`
	BankReferenceID        string                                             `json:"bankReferenceId"`
	Amount                 commonModel.Amount                                 `json:"amount"`
	AmountPaid             commonModel.Amount                                 `json:"amountPaid"`
	Status                 string                                             `json:"status"`
	PaymentLink            string                                             `json:"paymentLink,omitempty"`
	CreatedAt              time.Time                                          `json:"createdAt"`
	UpdatedAt              time.Time                                          `json:"updatedAt"`
	ExpiredAt              *time.Time                                         `json:"expiredAt"`
	InvestigationStartedAt *time.Time                                         `json:"investigationStartedAt"`
	CancelledAt            *time.Time                                         `json:"cancelledAt"`
	CancellationReason     string                                             `json:"cancellationReason,omitempty"`
	TotalSplitAmount       commonModel.Amount                                 `json:"totalSplitAmount"`
	Fee                    commonModel.Amount                                 `json:"fee"`
	SettledAmount          commonModel.Amount                                 `json:"settledAmount"`
	SplitRoutingConfigs    []SplitRoutingConfiguration                        `json:"splitRoutingConfigurations,omitempty"`
	SettlementModel        string                                             `json:"settlementModel"`
	FdsRiskAssessment      *fdsCommonModel.FdsRiskAssessment                  `json:"fdsRiskAssessment,omitempty"`
	OrderInfo              *unifiedPaymentModel.PaymentOrder                  `json:"orderInfo,omitempty"`
	CustomerInfo           *unifiedPaymentModel.CustomerInformation           `json:"customerInfo,omitempty"`
	Charges                []*unifiedPaymentModel.ChargeResponse              `json:"charges,omitempty"`
	PaymentType            string                                             `json:"paymentType"`
	Metadata               *map[string]any                                    `json:"metadata,omitempty"`
	RefundDetails          []refundModel.RefundResponse                       `json:"refundDetails,omitempty"`
	StatusHistory          []unifiedPaymentModel.PaymentStatusHistoryResponse `json:"statusHistory,omitempty"`
}

type PaymentTypeDetail struct {
	VirtualAccountName   *string `json:"virtualAccountName,omitempty"`
	VirtualAccountNumber *string `json:"virtualAccountNumber,omitempty"`
	QRISMerchantName     *string `json:"qrisMerchantName,omitempty"`
	QRISURL              *string `json:"qrisUrl,omitempty"`
	QrContent            *string `json:"qrContent,omitempty"`
	CardIssuer           *string `json:"cardIssuer,omitempty"`
	CardNumber           *string `json:"cardNumber,omitempty"`
	CardName             string  `json:"cardName,omitempty"`
	CardExpiryMonth      string  `json:"expiryMonth,omitempty"`
	CardExpiryYear       string  `json:"expiryYear,omitempty"`
	CardBrand            string  `json:"-"`

	EWalletChannel        string `json:"ewalletChannel,omitempty"`
	EWalletAppRedirectURL string `json:"ewalletAppRedirectUrl,omitempty"`
	EWalletWebRedirectURL string `json:"ewalletWebRedirectUrl,omitempty"`
}

type PaymentInsightItem struct {
	Total       int                `json:"total"`
	TotalAmount commonModel.Amount `json:"totalAmount"`
}

type SplitRoutingConfiguration struct {
	Type             string  `json:"type"`
	Remarks          string  `json:"remarks"`
	Currency         string  `json:"currency"`
	MerchantID       string  `json:"merchantId"`
	TransferID       string  `json:"transferId"`
	FixedAmount      float64 `json:"fixedAmount"`
	PercentageAmount float64 `json:"percentageAmount"`
	Status           string  `json:"status"`
	Beneficiary      string  `json:"beneficiary"`
	FinalAmount      float64 `json:"finalAmount"`
}

func (p *PaymentResponseItem) ToSnapVaBill() *snapVa.BillDetail {
	return &snapVa.BillDetail{
		BillNo:   p.ItemID,
		BillName: p.Name,
		BillDescription: &snapVa.Description{
			English:   p.Description,
			Indonesia: p.Description,
		},
		BillAmount: snapVa.Amount{
			Value:    p.Amount.Value.StringFixed(2),
			Currency: p.Amount.Currency,
		},
	}
}

func (p *PaymentResponse) ToPaymentResponse(
	payment *PaymentDTO,
	snapCoreResp *snapCoreModel.CreateVirtualAccountResponseData,
	paymentRequest *PaymentRequest,
	customer *customerModel.Customer,
	merchantID string,
) {
	p.UUID = payment.UUID
	p.MerchantID = merchantID
	p.ReferenceID = *payment.ReferenceID
	p.Status = payment.Status
	p.TotalAmount = &Amount{
		Value:    decimal.RequireFromString(snapCoreResp.TotalAmount.Value),
		Currency: paymentRequest.TotalAmount.Currency,
	}

	phone := ""
	if customer != nil && customer.PhoneNumber != "" {
		phone = customer.PhoneNumber
	}

	if customer != nil {
		p.Customer = &PaymentRequestCustomer{
			CustomerID: customer.UUID,
			Name:       customerModel.FirstNameAndLastNameToFullName(customer.FirstName, customer.LastName),
			Email:      customer.Email,
			Phone:      phone,
		}
	}
	p.PaymentMethod = paymentRequest.PaymentMethod
	p.PaymentMethodId = payment.PaymentMethodID
	p.TransactionDate = &payment.CreatedAt // TODO: pick from client request later
}

func (p *PaymentResponse) ToVirtualAccountResponse(
	paymentRequest *PaymentRequest,
	snapCoreResp *snapCoreModel.CreateVirtualAccountResponseData,
) {
	var (
		minAmount *Amount
		maxAmount *Amount
	)

	if snapCoreResp.MinAmount.Value != "" {
		minAmount = &Amount{
			Value:    decimal.RequireFromString(snapCoreResp.MinAmount.Value),
			Currency: snapCoreResp.MinAmount.Currency,
		}
	}

	if snapCoreResp.MaxAmount.Value != "" {
		maxAmount = &Amount{
			Value:    decimal.RequireFromString(snapCoreResp.MaxAmount.Value),
			Currency: snapCoreResp.MaxAmount.Currency,
		}
	}

	p.VirtualAccount = &PaymentVirtualAccountResponse{
		Issuer:                snapCoreResp.Acquirer,
		VirtualAccountTrxType: paymentRequest.VirtualAccount.VirtualAccountTrxType,
		VirtualAccountNumber:  snapCoreResp.VirtualAccountNo,
		VirtualAccountName:    snapCoreResp.AccountName,
		ExpiredDate:           &snapCoreResp.ExpiredAt,
		MinAmount:             minAmount,
		MaxAmount:             maxAmount,
	}
}

func (p *PaymentResponse) ToQrisResponse(
	payment *PaymentDTO,
	snapCoreResp *snapCoreModelQr.GenerateQrMpmResponseData,
	paymentRequest *PaymentRequest,
) {
	var amount *Amount
	if paymentRequest.Qris.QrType == constant.QrTypeDynamic {
		amount = &Amount{
			Value:    decimal.RequireFromString(snapCoreResp.Amount.Value),
			Currency: snapCoreResp.Amount.Currency,
		}
	}

	qrStatus := constant.QrStatusActive
	if payment.Status != constantPayment.PAYMENT_STATUS_PENDING {
		qrStatus = constant.QrStatusInactive
	}

	p.UUID = payment.UUID
	p.MerchantID = payment.MerchantID
	p.ReferenceID = *payment.ReferenceID
	p.Status = payment.Status
	p.PaymentMethodId = payment.PaymentMethodID
	p.PaymentMethod = paymentRequest.PaymentMethod
	p.TotalAmount = amount
	p.CreatedAt = payment.CreatedAt
	p.PaymentURL = payment.PaymentURL
	p.ExpiredAt = payment.ExpiredAt
	p.Qris = &PaymentQrisResponse{
		ReferenceNo:        snapCoreResp.ReferenceNo,
		PartnerReferenceNo: snapCoreResp.PartnerReferenceNo,
		MerchantName:       snapCoreResp.MerchantName,
		QrContent:          snapCoreResp.QrContent,
		QrUrl:              snapCoreResp.QrUrl,
		QrImage:            snapCoreResp.QrImage,
		QrType:             paymentRequest.Qris.QrType,
		QrStatus:           qrStatus,
		ValidityPeriod:     snapCoreResp.ValidityPeriod,
		PaymentStatus:      payment.Status,
		SubMerchantId:      paymentRequest.Qris.SubMerchantId,
		Amount:             amount,
		Acquirer:           snapCoreResp.Acquirer,
		TransactionDate:    util.SnapCompatible(payment.UpdatedAt),
	}

	if paymentRequest.Qris.QrType == constant.QrTypeStatic {
		p.Qris.ValidityPeriod = nil
		p.Qris.QrExpiredDate = ""
	}

	if paymentRequest.Qris.QrType == constant.QrTypeDynamic && payment != nil && payment.ExpiredAt != nil {
		p.Qris.QrExpiredDate = util.SnapCompatible(*payment.ExpiredAt)
	}

	if payment.Status == constantPayment.PAYMENT_STATUS_PENDING {
		p.Qris.TransactionDate = ""
	}
}

func (p *PaymentResponse) ToQueryQrMpmStaticResponse(
	payment *Payment,
	snapCoreResp *snapCoreModelQr.GenerateQrMpmResponseData,
	snapCoreQueryResp *snapCoreModelQr.QueryQrMpmStaticResponseData,
) {
	var (
		snapCoreGenerateResp snapCoreModelQr.GenerateQrMpmResponseData
		detailData           []QrStaticDetailData
	)
	// Parse snapCore metadata from data creation
	metadata := *payment.Metadata
	metadataByte, _ := json.Marshal(metadata["snapCore"])
	json.Unmarshal(metadataByte, &snapCoreGenerateResp)

	// Build detail data
	for _, detail := range snapCoreQueryResp.DetailData {
		detailData = append(detailData, QrStaticDetailData{
			Amount: commonModel.Amount{
				Value:    detail.Amount.Value,
				Currency: detail.Amount.Currency,
			},
			Status: detail.Status,
			Type:   detail.Type,
			AdditionalInfo: QrStaticDetailDataAdditionalInfo{
				RRN:             detail.AdditionalInfo.Rrn,
				TransactionDate: detail.DateTime,
			},
		})
	}

	qrStatus := constant.QrStatusActive
	if payment.Status != constantPayment.PAYMENT_STATUS_PENDING {
		qrStatus = constant.QrStatusInactive
	}

	p.UUID = payment.UUID
	p.MerchantID = payment.MerchantID
	p.ReferenceID = *payment.ReferenceID
	p.Status = payment.Status
	p.Qris = &PaymentQrisResponse{
		ReferenceNo:        snapCoreResp.ReferenceNo,
		PartnerReferenceNo: snapCoreResp.PartnerReferenceNo,
		MerchantName:       snapCoreGenerateResp.MerchantName,
		QrContent:          snapCoreGenerateResp.QrContent,
		QrUrl:              snapCoreGenerateResp.QrUrl,
		QrImage:            snapCoreGenerateResp.QrImage,
		QrType:             constant.QrTypeStatic,
		QrStatus:           qrStatus,
		DetailData:         &detailData,
	}
}

func (p *PaymentResponse) ToSnapVAPaymentResponse() SnapPaymentVirtualAccountResponse {
	var (
		totalAmount *commonModel.Amount
		expiredDate string
	)

	trxDateTime := util.SnapCompatible(time.Now())
	if p.TransactionDate != nil && !p.TransactionDate.IsZero() {
		trxDateTime = util.SnapCompatible(*p.TransactionDate)
	}

	if p.TotalAmount != nil && !p.TotalAmount.Value.IsZero() {
		totalAmount = &commonModel.Amount{
			Value:    p.TotalAmount.Value.StringFixed(2),
			Currency: p.TotalAmount.Currency,
		}
	}

	if p.VirtualAccount.ExpiredDate != nil && !p.VirtualAccount.ExpiredDate.IsZero() {
		expiredDate = util.SnapCompatible(*p.VirtualAccount.ExpiredDate)
	}

	vaStatus := "ACTIVE"
	if p.VirtualAccount.VirtualAccountTrxType == constantPayment.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC {
		vaStatus = "INACTIVE"
	}

	bankReferenceId := ""
	if p.AdditionalInfo != nil && (*p.AdditionalInfo)["bankReferenceId"] != nil {
		bankReferenceId = (*p.AdditionalInfo)["bankReferenceId"].(string)
	}

	return SnapPaymentVirtualAccountResponse{
		TrxId:              p.UUID,
		VirtualAccountNo:   p.VirtualAccount.VirtualAccountNumber,
		VirtualAccountName: p.VirtualAccount.VirtualAccountName,
		PaidAmount:         p.PaidAmount,
		TotalAmount:        totalAmount,
		TrxDateTime:        trxDateTime,
		AdditionalInfo: &SnapVAPaymentAdditionalInfo{
			ReferenceId:           p.ReferenceID,
			Issuer:                strings.ToUpper(p.VirtualAccount.Issuer),
			VirtualAccountTrxType: p.VirtualAccount.VirtualAccountTrxType,
			ExpiredDate:           expiredDate,
			MinAmount:             p.VirtualAccount.MinAmount,
			MaxAmount:             p.VirtualAccount.MaxAmount,
			VaStatus:              vaStatus,
			PaymentStatus:         constantPayment.PAYMENT_STATUS_SUCCESS,
			BankReferenceId:       bankReferenceId,
		},
	}
}

func (item PaymentItemRequest) ToSnapCoreBillDetail() snapCoreModel.BillDetail {
	itemQty := item.Qty
	itemAmount := item.Amount
	itemTotalAmount := itemAmount.Value.Mul(decimal.NewFromInt(int64(itemQty)))
	itemDescription := item.Description

	snapCoreBillDetailRequest := snapCoreModel.BillDetail{
		BillName: item.Name,
		BillAmount: snapCoreModel.Amount{
			Value:    itemTotalAmount.StringFixed(2),
			Currency: item.Amount.Currency,
		},
		AdditionalInfo: item.Metadata,
	}

	if itemDescription != "" {
		snapCoreBillDetailRequest.BillDescription = snapCoreModel.Description{
			English:   itemDescription,
			Indonesia: itemDescription,
		}
	}

	return snapCoreBillDetailRequest
}

type PaymentMethodResponseGroup struct {
	VirtualAccount []*PaymentMethodResponse `json:"virtualAccount"`
	BankTransfer   []*PaymentMethodResponse `json:"bankTransfer"`
	CreditCard     []*PaymentMethodResponse `json:"creditCard"`
	QRIS           []*PaymentMethodResponse `json:"qris"`
}

type PaymentMethodResponse struct {
	UUID        string    `json:"uuid" example:"0d1efc99-8991-41db-acaa-899b60aed8a1"`
	Type        string    `json:"type" example:"Virtual Account"`
	Category    string    `json:"category" example:"disbursement_top_up"`
	Name        string    `json:"name" example:"VA Permata"`
	Description *string   `json:"description" example:"Virtual Account Permata"`
	Logo        *string   `json:"logo" example:"https://sandbox.bca.co.id/api/images/logo-bca.png"`
	Acquirer    string    `json:"acquirer" example:"Permata"`
	BankName    *string   `json:"bankName" example:"Permata Bank"`
	CreatedAt   time.Time `json:"createdAt" example:"2021-08-10T00:00:00+07:00"`
	UpdatedAt   time.Time `json:"updatedAt" example:"2021-08-10T00:00:00+07:00"`
}

// ToResponse convert PaymentMethod to PaymentMethodResponse
func (pm *PaymentMethod) ToResponse() *PaymentMethodResponse {
	return &PaymentMethodResponse{
		UUID:        pm.UUID,
		Type:        pm.Type,
		Category:    pm.Category,
		Name:        pm.Name,
		Description: pm.Description,
		Logo:        pm.Logo,
		Acquirer:    pm.Acquirer,
		BankName:    pm.BankName,
		CreatedAt:   pm.CreatedAt,
		UpdatedAt:   pm.UpdatedAt,
	}
}

type SnapQueryQrMpmStaticResponse struct {
	ResponseCode       string                `json:"responseCode"`
	ResponseMessage    string                `json:"responseMessage"`
	ReferenceNo        string                `json:"referenceNo,omitempty"`
	PartnerReferenceNo string                `json:"partnerReferenceNo,omitempty"`
	DetailData         *[]QrStaticDetailData `json:"detailData,omitempty"`
	AdditionalInfo     *SnapQrAdditionalInfo `json:"additionalInfo,omitempty"`
}

type SnapQueryQrMpmDynamicResponse struct {
	ResponseCode               string                       `json:"responseCode"`
	ResponseMessage            string                       `json:"responseMessage"`
	OriginalReferenceNo        string                       `json:"originalReferenceNo,omitempty"`
	OriginalPartnerReferenceNo string                       `json:"originalPartnerReferenceNo,omitempty"`
	ServiceCode                string                       `json:"serviceCode,omitempty"`
	LatestTransactionStatus    string                       `json:"latestTransactionStatus,omitempty"`
	TransactionStatusDesc      string                       `json:"transactionStatusDesc,omitempty"`
	Amount                     *commonModel.Amount          `json:"amount,omitempty"`
	AdditionalInfo             *SnapQrDynamicAdditionalInfo `json:"additionalInfo,omitempty"`
}

type SnapGenerateQrMpmResponse struct {
	ResponseCode       string                `json:"responseCode"`
	ResponseMessage    string                `json:"responseMessage"`
	ReferenceNo        string                `json:"referenceNo,omitempty"`
	PartnerReferenceNo string                `json:"partnerReferenceNo,omitempty"`
	MerchantName       string                `json:"merchantName,omitempty"`
	QrContent          string                `json:"qrContent,omitempty"`
	QrUrl              string                `json:"qrUrl,omitempty"`
	QrImage            *string               `json:"qrImage,omitempty"`
	AdditionalInfo     *SnapQrAdditionalInfo `json:"additionalInfo,omitempty"`
}

type SnapVaResponse struct {
	ResponseCode       string           `json:"responseCode"`
	ResponseMessage    string           `json:"responseMessage"`
	VirtualAccountData SnapVACreateData `json:"virtualAccountData,omitempty"`
}

type PaymentHistoryResponse struct {
	UUID               string             `json:"uuid"`
	ReferenceID        string             `json:"referenceId"`
	Method             string             `json:"method"`
	Channel            string             `json:"channel"`
	ProcessorRefNumber string             `json:"processorReferenceNumber"`
	Amount             commonModel.Amount `json:"amount"`
	AmountPaid         commonModel.Amount `json:"amountPaid"`
	Status             string             `json:"status"`
	CreatedAt          time.Time          `json:"createdAt"`
}

func (r *PaymentResponse) ToOpenApiSnapVAPaymentResponse() SnapVACreateData {
	totalAmount := func() *snapVa.Amount {
		if r.TotalAmount.Value.IsZero() {
			return nil
		}
		return &snapVa.Amount{
			Value:    r.TotalAmount.Value.StringFixed(2),
			Currency: r.TotalAmount.Currency,
		}
	}()

	minAmount := func() *snapVa.Amount {
		if r.VirtualAccount.MinAmount == nil {
			return nil
		}
		return &snapVa.Amount{
			Value:    r.VirtualAccount.MinAmount.Value.StringFixed(2),
			Currency: r.VirtualAccount.MinAmount.Currency,
		}
	}()

	maxAmount := func() *snapVa.Amount {
		if r.VirtualAccount.MaxAmount == nil {
			return nil
		}
		return &snapVa.Amount{
			Value:    r.VirtualAccount.MaxAmount.Value.StringFixed(2),
			Currency: r.VirtualAccount.MaxAmount.Currency,
		}
	}()

	expiredDate := func() string {
		if r.VirtualAccount.ExpiredDate == nil || r.VirtualAccount.ExpiredDate.IsZero() {
			return ""
		}
		return util.SnapCompatible(*r.VirtualAccount.ExpiredDate)
	}()

	lastUpdateDate := func() string {
		if r.LastUpdateDate == nil || r.LastUpdateDate.IsZero() {
			return ""
		}
		return util.SnapCompatible(*r.LastUpdateDate)
	}()

	paymentDate := func() string {
		if r.LastUpdateDate == nil || r.Status != "SUCCESS" {
			return ""
		}
		return util.SnapCompatible(*r.LastUpdateDate)
	}()

	vaStatus := "ACTIVE"
	if r.Status != "PENDING" {
		vaStatus = "INACTIVE"
	}

	if r.Status != "PENDING" && r.Status != "SUCCESS" {
		r.Status = "FAILED"
	}

	var defaultCustomer PaymentRequestCustomer
	if r.Customer == nil {
		r.Customer = &defaultCustomer
	}
	vaData := SnapVACreateData{
		VACommonField: snapVa.VACommonField{
			TrxId:              r.UUID,
			VirtualAccountNo:   r.VirtualAccount.VirtualAccountNumber,
			VirtualAccountName: r.VirtualAccount.VirtualAccountName,
			CustomerNo:         r.Customer.CustomerID,
			BillDetails:        r.VirtualAccount.BillDetails,
		},
		VirtualAccountTrxType: r.VirtualAccount.VirtualAccountTrxType,
		ExpiredDate:           expiredDate,
		TotalAmount:           totalAmount,
		LastUpdateDate:        lastUpdateDate,
		PaymentDate:           paymentDate,
		AdditionalInfo: SnapVAAdditionalInfo{
			ReferenceID:   r.ReferenceID,
			Issuer:        strings.ToUpper(r.VirtualAccount.Issuer),
			MinAmount:     minAmount,
			MaxAmount:     maxAmount,
			VaStatus:      vaStatus,
			PaymentStatus: r.Status,
			Metadata:      r.VirtualAccount.Metadata,
		},
	}

	return vaData
}

type SnapQrAdditionalInfo struct {
	QrType          string         `json:"qrType"`
	QrStatus        string         `json:"qrStatus,omitempty"`
	QrExpiredDate   string         `json:"qrExpiredDate"`
	PaymentStatus   string         `json:"paymentStatus,omitempty"`
	RRN             string         `json:"RRN,omitempty"`
	QrContent       string         `json:"qrContent,omitempty"`
	QrUrl           string         `json:"qrUrl,omitempty"`
	QrImage         *string        `json:"qrImage,omitempty"`
	MerchantName    string         `json:"merchantName,omitempty"`
	TransactionDate string         `json:"transactionDate,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

type SnapQrDynamicAdditionalInfo struct {
	QrType          string         `json:"qrType"`
	QrStatus        string         `json:"qrStatus,omitempty"`
	QrExpiredDate   string         `json:"qrExpiredDate"`
	PaymentStatus   string         `json:"paymentStatus,omitempty"`
	RRN             string         `json:"RRN,omitempty"`
	QrContent       string         `json:"qrContent,omitempty"`
	QrUrl           string         `json:"qrUrl,omitempty"`
	QrImage         *string        `json:"qrImage,omitempty"`
	MerchantName    string         `json:"merchantName,omitempty"`
	TransactionDate string         `json:"transactionDate"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

type SnapQrisNotificationCallbackResponse struct {
	ResponseCode    string                `json:"responseCode"`
	ResponseMessage string                `json:"responseMessage"`
	AdditionalInfo  *SnapQrAdditionalInfo `json:"additionalInfo"`
}

type SnapVACallbackResponse struct {
	ResponseCode       string              `json:"responseCode"`
	ResponseMessage    string              `json:"responseMessage"`
	VirtualAccountData *VirtualAccountData `json:"virtualAccountData"`
}

type VirtualAccountData struct {
	VirtualAccountNo   string `json:"virtualAccountNo"`
	VirtualAccountName string `json:"virtualAccountName"`
}
type UnifiedPaymentResponse struct {
	UUID              string                  `json:"uuid"`
	MerchantID        string                  `json:"merchantId"`
	ReferenceID       string                  `json:"referenceId"`
	Customer          *PaymentRequestCustomer `json:"customer,omitempty"`
	Status            string                  `json:"status"`
	PaidAmount        commonModel.Amount      `json:"paidAmount,omitempty"`
	Amount            commonModel.Amount      `json:"amount,omitempty"`
	PaymentMethod     string                  `json:"paymentMethod,omitempty"`
	PaymentMethodType string                  `json:"paymentMethodType,omitempty"`
	TypeDetail        PaymentTypeDetail       `json:"paymentTypeDetail"`
	PaymentItems      *[]PaymentResponseItem  `json:"paymentItems,omitempty"`
	LastUpdateDate    *time.Time              `json:"lastUpdateDate,omitempty"`
	ExpiryAt          *time.Time              `json:"expiryAt,omitempty"`
}

func (u *UnifiedPaymentResponse) ToSnapPayment() interface{} {
	if u.PaymentMethod == constant.ChannelVirtualAccount && (u.TypeDetail.VirtualAccountName != nil || u.TypeDetail.VirtualAccountNumber != nil) {
		return u.toSnapVAResponse()
	} else if u.PaymentMethod == constant.ChannelQris && u.TypeDetail.QrContent != nil {
		return u.toSnapQrisResponse()
	}

	if u.TypeDetail.VirtualAccountName != nil || u.TypeDetail.VirtualAccountNumber != nil {
		return u.toSnapVAResponse()
	} else if u.TypeDetail.QrContent != nil || (u.PaymentMethod == constant.ChannelQris) {
		return u.toSnapQrisResponse()
	}

	return u
}

func (u *UnifiedPaymentResponse) toSnapVAResponse() SnapVACallbackResponse {
	vaNumber := ""
	vaName := ""

	if u.TypeDetail.VirtualAccountNumber != nil {
		vaNumber = *u.TypeDetail.VirtualAccountNumber
	}
	if u.TypeDetail.VirtualAccountName != nil {
		vaName = *u.TypeDetail.VirtualAccountName
	}

	responseCode := "2005200"
	responseMessage := "Successful"

	switch strings.ToUpper(u.Status) {
	case "FAILED":
		responseCode = "4015200"
		responseMessage = "Transaction Failed"
	case "PENDING":
		responseCode = "2005201"
		responseMessage = "Transaction Pending"
	case "SUCCESS", "COMPLETED":
		responseCode = "2005200"
		responseMessage = "Successful"
	}

	return SnapVACallbackResponse{
		ResponseCode:    responseCode,
		ResponseMessage: responseMessage,
		VirtualAccountData: &VirtualAccountData{
			VirtualAccountNo:   vaNumber,
			VirtualAccountName: vaName,
		},
	}
}

func (u *UnifiedPaymentResponse) toSnapQrisResponse() SnapQrisNotificationCallbackResponse {
	responseCode := "2005200"
	responseMessage := "Successful"

	switch strings.ToUpper(u.Status) {
	case constant.StatusFailed:
		responseCode = "4015200"
		responseMessage = "Transaction Failed"
	case constant.StatusPending:
		responseCode = "2005201"
		responseMessage = "Transaction Pending"
	case constant.StatusSuccess, "COMPLETED":
		responseCode = "2005200"
		responseMessage = "Successful"
	}

	return SnapQrisNotificationCallbackResponse{
		ResponseCode:    responseCode,
		ResponseMessage: responseMessage,
		AdditionalInfo: &SnapQrAdditionalInfo{
			QrStatus:      constant.StatusActive,
			PaymentStatus: u.Status,
		},
	}
}

type CreateUnifiedPaymentResponse struct {
	ID                string    `json:"id"`
	ClientReferenceID string    `json:"clientReferenceId"`
	Amount            Amount    `json:"amount"`
	PaymentMethod     string    `json:"paymentMethod"`
	ExpiryAt          time.Time `json:"expiryAt"`
	PaymentLink       string    `json:"paymentLink"`
}

func MapUnifiedPaymentV2ToV1Response(v2Response *unifiedPaymentModel.UnifiedPaymentSessionResponse) *CreateUnifiedPaymentResponse {
	v1Response := &CreateUnifiedPaymentResponse{
		ID:                v2Response.ID,
		ClientReferenceID: v2Response.ClientReferenceID,
		Amount: Amount{
			Currency: v2Response.Amount.Currency,
			Value:    decimal.NewFromFloat(v2Response.Amount.Value),
		},
		PaymentMethod: constant.MapUnifiedPaymentMethod(v2Response.PaymentMethod.Type),
		PaymentLink:   v2Response.PaymentUrl,
	}

	if v2Response.ExpiryAt != nil {
		v1Response.ExpiryAt = *v2Response.ExpiryAt
	}

	return v1Response
}

type UpdateUnifiedPaymentResponse struct {
	ID                string                  `json:"id"`
	ClientReferenceID string                  `json:"clientReferenceId"`
	Amount            Amount                  `json:"amount"`
	PaymentMethod     string                  `json:"paymentMethod"`
	Customer          *PaymentRequestCustomer `json:"customer,omitempty"`
	ExpiryAt          time.Time               `json:"expiryAt"`
	PaymentLink       string                  `json:"paymentLink,omitempty"`
}

type PaymentDetailForPaymentUIResponse struct {
	UUID               string                             `json:"uuid"`
	MerchantID         string                             `json:"merchantId"`
	DerivedMerchantID  string                             `json:"derivedMerchantId,omitempty"`
	CustomerID         string                             `json:"customerId"`
	ReferenceID        string                             `json:"referenceId"`
	Merchant           MerchantDetail                     `json:"merchant"`
	PaymentMethod      PaymentMethodDetail                `json:"paymentMethod"`
	TypeDetail         PaymentTypeDetail                  `json:"paymentTypeDetail"`
	ProcessorRefNumber string                             `json:"processorReferenceNumber"`
	Channel            string                             `json:"paymentChannel"`
	BankReferenceID    string                             `json:"bankReferenceId"`
	Amount             commonModel.Amount                 `json:"amount"`
	AmountPaid         commonModel.Amount                 `json:"amountPaid"`
	Status             string                             `json:"status"`
	TransactionID      string                             `json:"transactionId"`
	CreatedAt          time.Time                          `json:"createdAt"`
	UpdatedAt          time.Time                          `json:"updatedAt"`
	ExpiredAt          *time.Time                         `json:"expiredAt"`
	PaidAt             *time.Time                         `json:"paidAt"`
	FdsRiskAssessment  *fdsCommonModel.FdsRiskAssessment  `json:"fdsRiskAssessment,omitempty"`
	RedirectUrl        UnifiedPaymentRedirectUrl          `json:"redirectUrl"`
	InfoBanner         *InfoBanner                        `json:"infoBanner,omitempty"`
	ChargeStatus       string                             `json:"chargeStatus,omitempty"`
	ExpirationMode     string                             `json:"expirationMode,omitempty"`
	CreatedFrom        string                             `json:"createdFrom,omitempty"`
	BypassStatusPage   bool                               `json:"bypassStatusPage"`
	Mode               string                             `json:"mode,omitempty"`
	PaymentType        string                             `json:"paymentType,omitempty"`
	Metadata           *PaymentDetailForPaymentUIMetadata `json:"metadata,omitempty"`
	InquiryDetail      *InquiryDetailResponse             `json:"inquiryDetail,omitempty"`
}

type InquiryDetailResponse struct {
	HasFinalStatus bool `json:"hasFinalStatus"`
}

type InfoBanner struct {
	Message string `json:"message"`
	Type    string `json:"type"` // "info", "warning", "success"
}

type MerchantDetail struct {
	Name string `json:"name"`
	Logo string `json:"logo"`
}

type PaymentMethodDetail struct {
	Name     string  `json:"name"`
	Logo     *string `json:"logo"`
	Method   string  `json:"method"`
	Category string  `json:"category"`
}

type PaymentDownloadHistoryResponse struct {
	URL string `json:"url"`
}

type CardFundedPayoutMetadata struct {
	VendorName        string    `json:"vendorName"`
	PayoutID          string    `json:"payoutId"`
	ReferenceID       string    `json:"referenceId"`
	BankAccountNumber string    `json:"bankAccountNumber"`
	BankAccountName   string    `json:"bankAccountName"`
	BankName          string    `json:"bankName"`
	Remarks           string    `json:"remarks"`
	Amount            string    `json:"amount"`
	Fee               string    `json:"fee"`
	TotalAmount       string    `json:"totalAmount"`
	CreatedAt         time.Time `json:"createdAt"`
}

type PaymentDetailForPaymentUIMetadata struct {
	UseCase          string                    `json:"useCase,omitempty"`
	CardFundedPayout *CardFundedPayoutMetadata `json:"cardFundedPayout,omitempty"`
}

func (r *PaymentResponse) ToPbUnifiedPaymentCallbackRequest(paymentEntity *Payment) *pb.UnifiedPaymentCallbackRequest {
	var expiredAt *timestamppb.Timestamp
	if r.ExpiredAt != nil {
		expiredAt = timestamppb.New(*r.ExpiredAt)
	}

	callbackRequest := &pb.UnifiedPaymentCallbackRequest{
		Id:                   r.UUID,
		ClientReferenceId:    r.ReferenceID,
		Amount:               r.TotalAmount.ProtoAmount(),
		PaymentMethod:        r.PaymentMethod,
		ExpiredAt:            expiredAt,
		CreatedAt:            timestamppb.New(r.CreatedAt),
		PaymentLink:          r.PaymentURL,
		PaymentMethodOptions: &pb.UnifiedPaymentCallbackRequest_PaymentMethodOptionsRequest{},
	}

	if r.Customer != nil {
		callbackRequest.Customer = &pb.UnifiedPaymentCallbackRequest_PaymentCustomerRequest{
			CustomerId: r.Customer.CustomerID,
			Name:       r.Customer.Name,
			Email:      r.Customer.Email,
			Phone:      r.Customer.Phone,
		}

		if r.Customer.Metadata != nil {
			protoStruct, _ := structpb.NewStruct(*r.Customer.Metadata)
			pbAny, _ := anypb.New(protoStruct)
			customerMetadataWrapper, _ := anypb.New(pbAny)

			callbackRequest.Customer.Metadata = customerMetadataWrapper
		}
	}

	if r.PaymentItems != nil {
		paymentItems := make([]*pb.UnifiedPaymentCallbackRequest_PaymentItemRequest, len(*r.PaymentItems))
		for i, item := range *r.PaymentItems {
			paymentItems[i] = &pb.UnifiedPaymentCallbackRequest_PaymentItemRequest{
				ItemId:      item.ItemID,
				Name:        item.Name,
				Description: item.Description,
				Amount:      item.Amount.ProtoAmount(),
				Qty:         float64(item.Qty),
			}
		}
		callbackRequest.PaymentItems = paymentItems
	}

	switch r.PaymentMethod {
	case constantPayment.PAYMENT_METHOD_VIRTUAL_ACCOUNT:
		callbackRequest.PaymentMethodOptions.VirtualAccount = &pb.UnifiedPaymentCallbackRequest_PaymentMethodOptionsVirtualAccountRequest{
			Issuer:                r.VirtualAccount.Issuer,
			VirtualAccountTrxType: r.VirtualAccount.VirtualAccountTrxType,
			VirtualAccountNumber:  r.VirtualAccount.VirtualAccountNumber,
			VirtualAccountName:    r.VirtualAccount.VirtualAccountName,
			MinAmount:             r.VirtualAccount.MinAmount.ProtoAmount(),
		}
	case constantPayment.PAYMENT_METHOD_QRIS:
		callbackRequest.PaymentMethodOptions.Qris = &pb.UnifiedPaymentCallbackRequest_PaymentMethodOptionsQrisRequest{
			ReferenceNo:        r.Qris.ReferenceNo,
			PartnerReferenceNo: r.Qris.PartnerReferenceNo,
			MerchantName:       r.Qris.MerchantName,
			QrContent:          r.Qris.QrContent,
			QrUrl:              r.Qris.QrUrl,
			QrType:             r.Qris.QrType,
		}
	case constantPayment.PAYMENT_METHOD_CREDIT_CARD:
		ccCallbackFromPaymentRequest, err := paymentEntity.ToPaymentCreditCardCallbackRequest()
		if err != nil {
			break
		}

		callbackRequest.PaymentMethodOptions.Card = &pb.UnifiedPaymentCallbackRequest_PaymentMethodOptionsCreditCardRequest{
			MerchantId:     ccCallbackFromPaymentRequest.MerchantId,
			BankMerchantId: ccCallbackFromPaymentRequest.BankMerchantId,
			CardData: &pb.UnifiedPaymentCallbackRequest_PaymentCreditCardData{
				CardType:    ccCallbackFromPaymentRequest.CardData.CardType,
				CardBrand:   ccCallbackFromPaymentRequest.CardData.CardBrand,
				CardIssuing: ccCallbackFromPaymentRequest.CardData.CardIssuing,
				CountryCode: ccCallbackFromPaymentRequest.CardData.CountryCode,
				Fingerprint: ccCallbackFromPaymentRequest.CardData.Fingerprint,
			},
			PaymentStatus: ccCallbackFromPaymentRequest.PaymentStatus,
		}
	}

	return callbackRequest
}

// Custom unmarshal logic
func (p *PaymentResponse) UnmarshalJSON(data []byte) error {
	type Alias PaymentResponse // Prevent recursion
	raw := &struct {
		AdditionalInfo json.RawMessage `json:"additionalInfo"` // intercept just this one
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if len(raw.AdditionalInfo) == 0 {
		return nil
	}

	var m map[string]any
	if err := json.Unmarshal(raw.AdditionalInfo, &m); err != nil {
		return fmt.Errorf("failed to unmarshal additionalInfo: %w", err)
	}

	// Use the inner "value" if present and it's a map
	if v, ok := m["value"].(map[string]any); ok {
		p.AdditionalInfo = &v
	} else {
		p.AdditionalInfo = &m
	}

	return nil
}

// LoadPaymentV2CustomerOrderInformation populates the OrderInfo and CustomerInfo fields
// of the PaymentHistoryDetailResponse object using data from the provided payment and customer objects.
//
// For OrderInfo, it extracts and converts the "orderInformation" from payment metadata.
// For CustomerInfo, it builds a complete customer profile including personal details,
// phone number, refund preferences, and stored payment methods from the customer object.
//
// If either payment or customer is nil, the corresponding fields will not be populated.
func (p *PaymentHistoryDetailResponse) LoadPaymentV2CustomerOrderInformation(payment *Payment, customer *customerModel.Customer) {
	if payment != nil && payment.Metadata != nil {
		orderInfo, _ := util.ConvertToStruct[*unifiedPaymentModel.PaymentOrder]((*payment.Metadata)["orderInformation"])
		p.OrderInfo = orderInfo
	}

	if customer != nil {
		refundReference, _ := util.ConvertToStruct[*unifiedPaymentModel.UnifiedPaymentRefundPreference](customer.Metadata["refundPreference"])
		paymentMethods, _ := util.ConvertToStruct[[]*unifiedPaymentModel.CustomerPaymentMethod](customer.Metadata["paymentMethods"])

		p.CustomerInfo = &unifiedPaymentModel.CustomerInformation{
			CustomerID: customer.UUID,
			GivenName:  customer.FirstName,
			Surname:    &customer.LastName,
			SureName:   customer.LastName,
			Email:      customer.Email,
			PhoneNumber: &unifiedPaymentModel.UnifiedPaymentPhoneNumber{
				CountryCode: customer.PhoneCountryCode,
				Number:      customer.PhoneNumber,
			},
			RefundPreference:     refundReference,
			StoredPaymentMethods: paymentMethods,
		}
	}
}

func (p *PaymentResponse) BuildQRISDataFromPaymentV2(payment *Payment, charge *orchestratorModel.AccountTransactionWithUseCase) {
	var unifiedPaymentMetadata unifiedPaymentModel.MetadataUnifiedPayment
	metadataB, _ := json.Marshal(payment.Metadata)
	_ = json.Unmarshal(metadataB, &unifiedPaymentMetadata)

	chargeMethodDetails := &unifiedPaymentModel.ChargePaymentMethodDetails{}
	_ = json.Unmarshal(charge.AdditionalInfo.JSONText, &struct {
		MethodDetail interface{} `json:"methodDetail"`
	}{
		MethodDetail: chargeMethodDetails,
	})

	qrStatus := constant.QrStatusActive
	if payment.Status != constant.UnifiedPaymentSessionStatusRequireAction {
		qrStatus = constant.QrStatusInactive
	}

	if chargeMethodDetails.Qr == nil {
		return
	}

	paymentReferenceID := ""
	if payment.ReferenceID != nil {
		paymentReferenceID = *payment.ReferenceID
	}

	p.Qris = &PaymentQrisResponse{
		ReferenceNo:        chargeMethodDetails.Qr.RetrievalReferenceNumber,
		PartnerReferenceNo: paymentReferenceID,
		MerchantName:       chargeMethodDetails.Qr.MerchantName,
		QrContent:          chargeMethodDetails.Qr.QrContent,
		QrUrl:              chargeMethodDetails.Qr.QrUrl,
		QrImage:            util.ValueToPtr(""),
		QrStatus:           qrStatus,
		PaymentStatus:      constant.MapUnifiedPaymentStatusToV1(payment.Status),
		Amount: &Amount{
			Currency: charge.Currency,
			Value:    decimal.NewFromFloat(charge.Credit),
		},
		QrType:          constant.QrTypeDynamic,
		TransactionDate: util.SnapCompatible(charge.TransactionTimestamp),
	}

	if payment.ExpiredAt != nil {
		p.Qris.QrExpiredDate = util.SnapCompatible(*payment.ExpiredAt)
	}
}

func (p *PaymentResponse) BuildQVADataFromPaymentV2(payment *Payment, charge *orchestratorModel.AccountTransactionWithUseCase) {
	var unifiedPaymentMetadata unifiedPaymentModel.MetadataUnifiedPayment
	metadataB, _ := json.Marshal(payment.Metadata)
	_ = json.Unmarshal(metadataB, &unifiedPaymentMetadata)

	chargeMethodDetails := &unifiedPaymentModel.ChargePaymentMethodDetails{}
	_ = json.Unmarshal(charge.AdditionalInfo.JSONText, &struct {
		MethodDetail interface{} `json:"methodDetail"`
	}{
		MethodDetail: chargeMethodDetails,
	})

	if chargeMethodDetails.VirtualAccount == nil {
		return
	}

	p.VirtualAccount = &PaymentVirtualAccountResponse{
		Issuer:                chargeMethodDetails.VirtualAccount.Channel,
		VirtualAccountTrxType: constantPayment.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC,
		VirtualAccountNumber:  chargeMethodDetails.VirtualAccount.VirtualAccountNumber,
		VirtualAccountName:    chargeMethodDetails.VirtualAccount.VirtualAccountName,
		ExpiredDate:           &chargeMethodDetails.VirtualAccount.ExpiryAt,
	}
}

func (p *Payment) BuildCardDataFromPaymentV2(charge *orchestratorModel.AccountTransactionWithUseCase) {
	var unifiedPaymentMetadata unifiedPaymentModel.MetadataUnifiedPayment
	metadataB, _ := json.Marshal(p.Metadata)
	_ = json.Unmarshal(metadataB, &unifiedPaymentMetadata)

	chargeMethodDetails := &unifiedPaymentModel.ChargePaymentMethodDetails{}
	_ = json.Unmarshal(charge.AdditionalInfo.JSONText, &struct {
		MethodDetail interface{} `json:"methodDetail"`
	}{
		MethodDetail: chargeMethodDetails,
	})

	if chargeMethodDetails.Card == nil {
		return
	}

	chargeCardDetail := chargeMethodDetails.Card
	paymentCCMetadata := card.CreditcardMetadata{
		BankMerchantID: chargeCardDetail.BankMerchantID,
		CardData: &card.CardDataRequest{
			First8Digit: chargeCardDetail.First8,
			Last4Digit:  chargeCardDetail.First6,
			CardType:    chargeCardDetail.BinInformations.Type,
			CardBrand:   chargeCardDetail.BinInformations.Brand,
			CardIssuing: chargeCardDetail.BinInformations.IssuingBank,
			CountryCode: chargeCardDetail.BinInformations.Country,
			Fingerprint: chargeCardDetail.Fingerprint,
		},
	}

	if p.Metadata == nil {
		p.Metadata = &map[string]any{}
	}

	_ = util.MergeStructToMap(p.Metadata, paymentCCMetadata)
}

type VARangeResponse struct {
	UUID       string       `json:"uuid"`
	BankCode   string       `json:"bankCode"`
	BankName   string       `json:"bankName"`
	BankLogo   string       `json:"bankLogo"`
	CloseRange *VARangeData `json:"closeRange,omitempty"`
	OpenRange  *VARangeData `json:"openRange,omitempty"`
	Status     string       `json:"status"`
	CreatedAt  time.Time    `json:"createdAt"`
	UpdatedAt  time.Time    `json:"updatedAt"`
	MID        string       `json:"mid"`
}

type VARangeData struct {
	BinPrefix string `json:"binPrefix"`
	Start     string `json:"start"`
	End       string `json:"end"`
}

func ParsePaymentMethodsToVARangeResponseList(paymentMethods []*PaymentMethodWithPivot, merchantMID string) []VARangeResponse {
	respList := make([]VARangeResponse, len(paymentMethods))
	for i, paymentMethod := range paymentMethods {
		bankName, bankLogo := "", ""
		if paymentMethod.BankName != nil {
			bankName = *paymentMethod.BankName
		}
		if paymentMethod.Logo != nil {
			bankLogo = *paymentMethod.Logo
		}
		status := "INACTIVE"
		if paymentMethod.IsActive && paymentMethod.ActivationStatus == constant.PaymentMethodActivationStatusApproved {
			status = "ACTIVE"
		}

		resp := VARangeResponse{
			UUID:      paymentMethod.UUID,
			BankCode:  paymentMethod.Acquirer,
			BankName:  bankName,
			BankLogo:  bankLogo,
			Status:    status,
			CreatedAt: paymentMethod.CreatedAt,
			UpdatedAt: paymentMethod.UpdatedAt,
			MID:       merchantMID,
		}

		if paymentMethod.MerchantConfigObj != nil && paymentMethod.MerchantConfigObj.PartnerConfig != nil &&
			paymentMethod.MerchantConfigObj.PartnerConfig.VirtualAccount != nil {
			vaConfigs := paymentMethod.MerchantConfigObj.PartnerConfig.VirtualAccount

			for _, vaConfig := range vaConfigs.Items {
				if vaConfig.Type == constantPayment.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC {
					resp.OpenRange = &VARangeData{
						BinPrefix: vaConfig.BINPrefix,
						Start:     vaConfig.StartRange,
						End:       vaConfig.EndRange,
					}
					continue
				}

				resp.CloseRange = &VARangeData{
					BinPrefix: vaConfig.BINPrefix,
					Start:     vaConfig.StartRange,
					End:       vaConfig.EndRange,
				}
			}
		}

		respList[i] = resp
	}

	return respList
}

type InvestigationSummaryItem struct {
	TotalAmount string `json:"totalAmount"`
	Currency    string `json:"currency"`
}

type InvestigationSummaryResponse struct {
	OnInvestigation InvestigationSummaryItem `json:"onInvestigation"`
	Success         InvestigationSummaryItem `json:"success"`
	Failed          InvestigationSummaryItem `json:"failed"`
}

type GetEncryptionKeyResponse struct {
	EncryptionKey string `json:"encryptionKey"`
}

type VCCTerminalBatchChargeResponse struct {
	BatchID       string           `json:"batchId"`
	SuccessCount  int64            `json:"successCount"`
	SuccessTotal  float64          `json:"successTotal"`
	FailedCount   int64            `json:"failedCount"`
	FailedTotal   float64          `json:"failedTotal"`
	FailedCharges []BookingPayload `json:"failedCharges"`
}

// GetPaymentReceiptResponse - Returns PDF bytes directly
type GetPaymentReceiptResponse struct {
	Filename string `json:"-"`
	PDF      []byte `json:"-"` // Internal: PDF bytes for HTTP response
}

// PaymentReceiptData - Data for payment receipt PDF template
type PaymentReceiptData struct {
	TransactionDate string
	ReferenceID     string
	MerchantName    string
	Amount          string
	PaymentID       string
	PaymentMethod   string
	ImageHeader     string
	ImageBackground string
}
type AutoSplitPaymentSummary struct {
	ParentPaymentID             string  `db:"parent_payment_id"`
	NumberOfCharges             int     `db:"total_charge"`
	NumberOfSuccessfulCharges   int     `db:"total_success_charge"`     // PAID
	NumberOfInProcessCharges    int     `db:"total_in_progress_charge"` // PROCESSING
	NumberOfFailedCharges       int     `db:"total_failed_charge"`      // CANCELLED
	NumberOfExpiredCharges      int     `db:"total_expired_charge"`     // EXPIRED
	TotalSuccessfulChargeAmount float64 `db:"total_success_charge_amount"`
	TotalInProgressChargeAmount float64 `db:"total_in_progress_charge_amount"`
	TotalFailedChargeAmount     float64 `db:"total_failed_charge_amount"`
}

func (m AutoSplitPaymentSummary) IsComplete() bool {
	if m.NumberOfCharges == 0 {
		return false
	}

	return m.NumberOfCharges == (m.NumberOfExpiredCharges + m.NumberOfFailedCharges + m.NumberOfSuccessfulCharges)
}
