package unifiedPaymentModel

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	fdsCommonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/fdsProcessor/fdsCommon"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/orchestrator"
	paymentCaptureModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/paymentCapture"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/proto/messages/callback"
	splitRoutingPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/splitRoutingPayment"
	typ "github.com/paper-indonesia/pivot-backoffice/pkg/types"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/paper-indonesia/pdk/v2/encrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UnifiedPaymentSessionResponse struct {
	ID string `json:"id"`

	// From Request
	ClientReferenceID          string                                                       `json:"clientReferenceId"`
	Amount                     Amount                                                       `json:"amount"`
	AutoConfirm                bool                                                         `json:"autoConfirm"`
	Mode                       string                                                       `json:"mode"`
	RedirectUrl                RedirectUrl                                                  `json:"redirectUrl"`
	PaymentType                string                                                       `json:"paymentType"`
	PaymentMethod              *PaymentMethod                                               `json:"paymentMethod"`
	StatementDescriptor        string                                                       `json:"statementDescriptor"`
	SplitRoutingConfigurations *[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration `json:"splitRoutingConfigurations,omitempty"`
	SaveForFutureUse           *bool                                                        `json:"saveForFutureUse,omitempty"`
	ShowSavedPayment           *bool                                                        `json:"showSavedPayment,omitempty"`
	ExpirationMode             string                                                       `json:"expirationMode,omitempty"`
	RecurringID                string                                                       `json:"recurringId,omitempty"`
	InitiateFirstAuthorization bool                                                         `json:"initiateFirstAuthorization,omitempty"`

	// Generated
	Status              string                   `json:"status"`
	InvestigationStatus *string                  `json:"investigationStatus,omitempty"`
	CreatedAt           time.Time                `json:"createdAt"`
	UpdatedAt           time.Time                `json:"updatedAt"`
	ExpiryAt            *time.Time               `json:"expiryAt"`
	PaymentUrl          string                   `json:"paymentUrl,omitempty"`
	ShortPaymentUrl     string                   `json:"shortPaymentUrl,omitempty"`
	EncryptionKey       *encrypt.RSAEncryptedKey `json:"encryptionKey,omitempty"` // Generated only for CARD API

	// Charge Detail
	ChargeDetails    []*ChargeResponse `json:"chargeDetails"`
	AutoSplitDetails *AutoSplitDetails `json:"autoSplitDetails,omitempty"`

	// Customer
	CustomerId          *string                      `json:"customerId,omitempty"`
	CustomerInformation *CustomerInformationResponse `json:"customer,omitempty"`

	CancelledAt        *time.Time `json:"cancelledAt,omitempty"`
	CancellationReason string     `json:"cancellationReason,omitempty"`

	Metadata interface{} `json:"metadata,omitempty"`

	FdsRiskAssessment *fdsCommonModel.FdsRiskAssessment `json:"fdskRiskAssessment,omitempty"`

	PaymentMethodOptions PaymentMethodOptions           `json:"-"` // Do not show for payment method options
	CardFundedPayout     *CardFundedPayout              `json:"-"` // Do not show for card-funded payout
	AutoSplitPayment     *AutoSplitPayment              `json:"-"` // Do not show for auto split payment
	StatusHistory        []PaymentStatusHistoryResponse `json:"statusHistory,omitempty"`
	// Recurring Payment, the data is not displayed as it is for internal use only.
	FirstAuthorizationMethod  string                `json:"-"`
	FirstAuthorizationOrderID *string               `json:"-"`
	RecurringBillingCycle     RecurringBillingCycle `json:"-"`

	LastInquiryAt *time.Time `json:"lastInquiryAt,omitempty"`
	// Bypass status page for mode redirect
	BypassStatusPage bool `json:"bypassStatusPage"`
}

type ChargeResponse struct {
	ID                              string                            `json:"id" db:"uuid"`
	PaymentSessionID                string                            `json:"paymentSessionId" db:"payment_session_id"`
	PaymentSessionClientReferenceID string                            `json:"paymentSessionClientReferenceId" db:"payment_session_reference_id"`
	Amount                          Amount                            `json:"amount" db:"amount"`
	StatementDescriptor             string                            `json:"statementDescriptor" db:"statement_descriptor"`
	Status                          string                            `json:"status"`
	AuthorizedAmount                *Amount                           `json:"authorizedAmount" db:"authorized_amount"`
	CapturedAmount                  *Amount                           `json:"capturedAmount" db:"captured_amount"`
	IsCaptured                      bool                              `json:"isCaptured" db:"is_captured"`
	FailureCode                     string                            `json:"failureCode,omitempty" db:"failure_code"`
	FailureMessage                  string                            `json:"failureMessage,omitempty"`
	Recommendation                  string                            `json:"recommendation,omitempty"`
	CreatedAt                       time.Time                         `json:"createdAt" db:"created_at"`
	UpdatedAt                       time.Time                         `json:"updatedAt" db:"updated_at"`
	PaidAt                          *time.Time                        `json:"paidAt" db:"transaction_timestamp"`
	ExpiredAt                       *time.Time                        `json:"expiredAt,omitempty" db:"expired_at"` // only for dashboard
	FdsRiskAssessment               *fdsCommonModel.FdsRiskAssessment `json:"fdsRiskAssessment,omitempty"`

	MerchantID     string             `json:"-" db:"merchant_id"`
	MerchantName   string             `json:"-" db:"merchant_name"`
	AdditionalInfo types.NullJSONText `db:"additional_info" json:"-"`

	NetworkResponseCode string `json:"networkResponseCode,omitempty"`
	*ChargePaymentMethodDetails
	StatusHistory []ChargeStatusHistoryResponse `json:"statusHistory,omitempty"`

	// Optional payment captures (do not show) -> then assign to the last charge
	CaptureHistoriesJSON types.NullJSONText                            `json:"-" db:"capture_histories"`
	CaptureHistories     []*paymentCaptureModel.CaptureHistoryResponse `json:"captureHistories,omitempty" db:"-"`

	LastInquiryAt   *time.Time       `json:"lastInquiryAt,omitempty"`
	VirtualTerminal *VirtualTerminal `json:"virtualTerminal,omitempty"`

	SafeToRetry      *bool  `json:"safeToRetry,omitempty"`
	SettlementStatus string `json:"-" db:"settlement_status"`
}

type AutoSplitDetails struct {
	Status                      string           `json:"status"`
	NumberOfCharges             int              `json:"numberOfCharges"`
	NumberOfSuccessfulCharges   int              `json:"numberOfSuccessfulCharges"`
	NumberOfInProcessCharges    int              `json:"numberOfInProcessCharges"`
	NumberOfFailedCharges       int              `json:"numberOfFailedCharges"`
	TotalSuccessfulChargeAmount *Amount          `json:"totalSuccessfulChargeAmount"`
	TotalFailedChargeAmount     *Amount          `json:"totalFailedChargeAmount"`
	TotalInProcessChargeAmount  *Amount          `json:"totalInProcessChargeAmount"`
	ChargesDetails              []ChargeResponse `json:"chargesDetails"`
}

func BuildAutoSplitDetails(charges []ChargeResponse) *AutoSplitDetails {
	details := &AutoSplitDetails{
		NumberOfCharges: len(charges),
	}

	for _, c := range charges {
		switch c.Status {
		case constant.ChargeStatusSuccess:
			details.NumberOfSuccessfulCharges++
			if details.TotalSuccessfulChargeAmount == nil {
				details.TotalSuccessfulChargeAmount = &Amount{Currency: c.Amount.Currency}
			}
			details.TotalSuccessfulChargeAmount.Value += c.Amount.Value
		case constant.ChargeStatusFailed, constant.ChargeStatusExpired:
			details.NumberOfFailedCharges++
			if details.TotalFailedChargeAmount == nil {
				details.TotalFailedChargeAmount = &Amount{Currency: c.Amount.Currency}
			}
			details.TotalFailedChargeAmount.Value += c.Amount.Value
		default:
			details.NumberOfInProcessCharges++
			if details.TotalInProcessChargeAmount == nil {
				details.TotalInProcessChargeAmount = &Amount{Currency: c.Amount.Currency}
			}
			details.TotalInProcessChargeAmount.Value += c.Amount.Value
		}
	}

	details.Status = mapAutoSplitStatus(details)
	details.ChargesDetails = charges
	return details
}

func mapAutoSplitStatus(d *AutoSplitDetails) string {
	total := d.NumberOfCharges
	if total == 0 {
		return constant.AutoSplitStatusProcessing
	}
	if d.NumberOfSuccessfulCharges == total {
		return constant.AutoSplitStatusSuccess
	}
	if d.NumberOfFailedCharges == total {
		return constant.AutoSplitStatusFailed
	}
	if d.NumberOfSuccessfulCharges > 0 && d.NumberOfFailedCharges > 0 {
		return constant.AutoSplitStatusPartialSuccess
	}
	return constant.AutoSplitStatusProcessing
}

type CaptureHistoryResponse struct {
	ID             string    `json:"captureId"`
	Currency       string    `json:"currency"`
	CapturedAmount float64   `json:"capturedAmount"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
}

// SNAP response structs for callback formatting
type SnapVAAdditionalInfo struct {
	ReferenceID           string `json:"referenceId"`
	Issuer                string `json:"issuer"`
	VirtualAccountTrxType string `json:"virtualAccountTrxType"`
	ExpiredDate           string `json:"expiredDate"`
	VAStatus              string `json:"vaStatus"`
	PaymentStatus         string `json:"paymentStatus"`
	BankReferenceID       string `json:"bankReferenceId"`
}

type SnapQRISAdditionalInfo struct {
	RRN             string `json:"RRN"`
	QrType          string `json:"qrType"`
	QrStatus        string `json:"qrStatus"`
	QrExpiredDate   string `json:"qrExpiredDate"`
	MerchantName    string `json:"merchantName"`
	PaymentStatus   string `json:"paymentStatus"`
	TransactionDate string `json:"transactionDate"`
}

type SnapAmountInfo struct {
	Currency string `json:"currency"`
	Value    string `json:"value"`
}

type SnapVAResponse struct {
	TrxID              string               `json:"trxId"`
	VirtualAccountNo   string               `json:"virtualAccountNo"`
	VirtualAccountName string               `json:"virtualAccountName"`
	PaidAmount         SnapAmountInfo       `json:"paidAmount"`
	TotalAmount        SnapAmountInfo       `json:"totalAmount"`
	TrxDateTime        string               `json:"trxDateTime"`
	AdditionalInfo     SnapVAAdditionalInfo `json:"additionalInfo"`
}

type SnapQRISResponse struct {
	OriginalReferenceNo        string                 `json:"originalReferenceNo"`
	OriginalPartnerReferenceNo string                 `json:"originalPartnerReferenceNo"`
	LatestTransactionStatus    string                 `json:"latestTransactionStatus"`
	TransactionStatusDesc      string                 `json:"transactionStatusDesc"`
	Amount                     SnapAmountInfo         `json:"amount"`
	AdditionalInfo             SnapQRISAdditionalInfo `json:"additionalInfo"`
}

func formatAmountValue(value float64) string {
	if value == float64(int64(value)) {
		return fmt.Sprintf("%.0f", value)
	}
	return fmt.Sprintf("%.2f", value)
}

func (c *ChargeResponse) ToSnapPayment() interface{} {
	if c.ChargePaymentMethodDetails != nil {
		if c.ChargePaymentMethodDetails.VirtualAccount != nil {
			return c.toSnapVAResponse()
		} else if c.ChargePaymentMethodDetails.Qr != nil {
			return c.toSnapQrisResponse()
		}
	}
	if c.VirtualAccount != nil {
		return c.toSnapVAResponse()
	}
	if c.Qr != nil {
		return c.toSnapQrisResponse()
	}
	return c
}

func (c *ChargeResponse) toSnapVAResponse() *SnapVAResponse {
	var vaData *ChargePaymentMethodDetailVirtualAccount
	if c.VirtualAccount != nil {
		vaData = c.VirtualAccount
	} else if c.ChargePaymentMethodDetails != nil && c.ChargePaymentMethodDetails.VirtualAccount != nil {
		vaData = c.ChargePaymentMethodDetails.VirtualAccount
	}

	// Helper function to get transaction date
	trxDateTime := ""
	if c.PaidAt != nil {
		trxDateTime = util.SnapCompatible(*c.PaidAt)
	}

	// Helper function to get VA status
	vaStatus := ""
	if vaData != nil {
		if c.Status == constant.StatusSuccess {
			vaStatus = constant.StatusActive
		} else {
			vaStatus = constant.StatusInactive
		}
	}

	// Initialize response with direct population
	response := &SnapVAResponse{
		TrxID: func() string {
			if c.PaymentSessionID != "" {
				return c.PaymentSessionID
			}
			return c.ID
		}(),
		VirtualAccountNo: func() string {
			if vaData != nil {
				return vaData.VirtualAccountNumber
			}
			return ""
		}(),
		VirtualAccountName: func() string {
			if vaData != nil {
				return vaData.VirtualAccountName
			}
			return ""
		}(),
		PaidAmount: SnapAmountInfo{
			Currency: c.Amount.Currency,
			Value:    formatAmountValue(c.Amount.Value),
		},
		TotalAmount: SnapAmountInfo{
			Currency: c.Amount.Currency,
			Value:    formatAmountValue(c.Amount.Value),
		},
		TrxDateTime: trxDateTime,
		AdditionalInfo: SnapVAAdditionalInfo{
			ReferenceID: c.PaymentSessionClientReferenceID,
			Issuer: func() string {
				if vaData != nil {
					return vaData.Channel
				}
				return ""
			}(),
			VirtualAccountTrxType: func() string {
				if vaData != nil {
					return vaData.VirtualAccountTrxType
				}
				return ""
			}(),
			VAStatus:      vaStatus,
			PaymentStatus: c.Status,
			BankReferenceID: func() string {
				if vaData != nil {
					return vaData.BankReferenceNo
				}
				return ""
			}(),
		},
	}

	return response
}

func (c *ChargeResponse) toSnapQrisResponse() *SnapQRISResponse {
	// Helper function to get QR data
	var qrData *ChargePaymentMethodDetailQr
	if c.Qr != nil {
		qrData = c.Qr
	} else if c.ChargePaymentMethodDetails != nil && c.ChargePaymentMethodDetails.Qr != nil {
		qrData = c.ChargePaymentMethodDetails.Qr
	}

	// Helper function to get transaction date
	transactionDate := ""
	if c.PaidAt != nil {
		transactionDate = util.SnapCompatible(*c.PaidAt)
	}

	// Helper function to get expiry date
	qrExpiredDate := ""
	if qrData != nil && !qrData.ExpiryAt.IsZero() {
		qrExpiredDate = util.SnapCompatible(qrData.ExpiryAt)
	}

	// Initialize response with direct population
	response := &SnapQRISResponse{
		OriginalReferenceNo:        c.PaymentSessionID,
		OriginalPartnerReferenceNo: c.PaymentSessionClientReferenceID,
		LatestTransactionStatus:    "00",
		TransactionStatusDesc:      c.Status,
		Amount: SnapAmountInfo{
			Currency: c.Amount.Currency,
			Value:    formatAmountValue(c.Amount.Value),
		},
		AdditionalInfo: SnapQRISAdditionalInfo{
			RRN: func() string {
				if qrData != nil {
					return qrData.RetrievalReferenceNumber
				}
				return ""
			}(),
			QrType: func() string {
				if qrData != nil {
					return qrData.QrType
				}
				return ""
			}(),
			QrStatus: func() string {
				if qrData != nil {
					return constant.StatusActive
				}
				return ""
			}(),
			QrExpiredDate: qrExpiredDate,
			MerchantName: func() string {
				if qrData != nil {
					return qrData.MerchantName
				}
				return ""
			}(),
			PaymentStatus:   c.Status,
			TransactionDate: transactionDate,
		},
	}

	return response
}

// mapStatusToSnapResponse maps ChargeResponse status to SNAP response codes
func (c *ChargeResponse) mapStatusToSnapResponse() (string, string) {
	switch strings.ToUpper(c.Status) {
	case constant.StatusFailed:
		return "4015200", "Transaction Failed"
	case constant.StatusPending:
		return "2005201", "Transaction Pending"
	case constant.StatusSuccess, "COMPLETED":
		return "2005200", "Successful"
	default:
		return "2005200", "Successful"
	}
}

func (c *ChargeResponse) SetFailureDetail() {
	paymentMethodType := ""
	if c.Card != nil {
		paymentMethodType = "Card"
	} else if c.Qr != nil {
		paymentMethodType = "QR"
	} else if c.VirtualAccount != nil {
		paymentMethodType = "Virtual Account"
	}

	if c.Status == constant.ChargeStatusExpired {
		c.FailureCode = "CHARGE_EXPIRED"
		c.FailureMessage = fmt.Sprintf("%s charge failed due to the transaction time has exceeded the channel expiration time.", paymentMethodType)
		c.Recommendation = "The shopper can try again or use another payment method."
		return
	}

	if c.Card != nil {
		c.Card.ExpMonth = typ.String(c.GetCardExpiryMonth())
		c.Card.ExpYear = typ.String(c.GetCardExpiryYear())

		if c.Status == constant.ChargeStatusFailed {
			issuerAuthorizationCode := ""
			if c.Card.AuthorizationResult != nil {
				issuerAuthorizationCode = c.Card.AuthorizationResult.IssuerAuthorizationCode
			}

			c.NetworkResponseCode = issuerAuthorizationCode
			// Only set failure detail if FDS is not passed
			if c.FdsRiskAssessment != nil && c.FdsRiskAssessment.Status == constant.FDS_STATUS_REJECTED {
				c.FailureCode = "BLOCKED_BY_FDS"
				c.FailureMessage = "Card payment failed. \n\nThe transaction was declined by FDS due to the transactions being categorized as high risk."
				c.Recommendation = "Verify and validate the transaction. The shopper can try again after the transaction has been validated."
				return
			}
			// Only set failure detail if FDS is not passed
			if c.FdsRiskAssessment != nil && c.FdsRiskAssessment.Status != constant.FDS_STATUS_PASSED {
				if c.FdsRiskAssessment.Status == constant.FDS_STATUS_REJECTED {
					c.FailureCode = "BLOCKED_BY_FDS"
					c.FailureMessage = "Card payment failed. \n\nThe transaction was declined by FDS due to the transactions being categorized as high risk."
					c.Recommendation = "Verify and validate the transaction. The shopper can try again after the transaction has been validated."

				} else if c.FdsRiskAssessment.Status == constant.FDS_STATUS_REVIEW {
					c.FailureCode = "REQUIRE_REVIEW"
					c.FailureMessage = "Card payment failed. \n\nThe transaction was deferred by FDS due to the transactions being categorized as suspicious."
					c.Recommendation = "Verify and validate the transaction. Approve or reject the transaction after it has been validated."
				}
				return
			}

			if c.Card.ResponseCode != nil && c.Card.ResponseCode.GatewayCode == constant.CreditCardGatewayCodeAborted {
				c.FailureCode = "CANCELLED_BY_USER"
				c.FailureMessage = "Card payment failed. \n\nThe 3DS attempt was cancelled by the cardholder."
				c.Recommendation = "The shopper can try again or use another payment method."
			} else if c.Card.AuthenticationResult != nil && c.Card.AuthenticationResult.ThreeDsResult != constant.CreditCardAuthenticationSuccess {
				c.FailureCode = "AUTHENTICATION_FAILED"
				c.FailureMessage = "Card authentication failed. \n\nThe 3DS attempt was rejected by the issuer."
				c.Recommendation = "The cardholder should contact their issuer for clarification. The shopper can try again after resolving the issue with their issuer, or use another payment method."
			}

			switch issuerAuthorizationCode {
			case "01", "03", "05", "06", "12", "13", "22", "40", "57", "61", "62", "63", "64", "65", "6P", "70", "82", "92", "93", "100", "109", "110", "115":
				c.FailureCode = "DECLINED_BY_CHANNEL"
				c.FailureMessage = "Card payment failed. \n\nThe transaction was declined by the channel."
				c.Recommendation = "The cardholder should contact their issuer for clarification. The shopper can try again after resolving the issue with their issuer, or use another payment method."
			case "54", "101":
				c.FailureCode = "DECLINED_BY_CHANNEL"
				c.FailureMessage = "Card payment failed. \n\nThe transaction was declined by the issuer due to the card has already expired."
				c.Recommendation = "The shopper should try again with another valid card or use another payment method."
			case "N7":
				c.FailureCode = "DECLINED_BY_CHANNEL"
				c.FailureMessage = "Card payment failed. \n\nThe transaction was declined by the issuer due to the submitted CVV is invalid."
				c.Recommendation = "The shopper should try again with another valid card or use another payment method."
			case "14", "15", "21", "46", "52", "53", "78", "79", "111":
				c.FailureCode = "INVALID_ACCOUNT"
				c.FailureMessage = "Card payment failed. \n\nThe transaction was declined by the issuer due to the card being marked as invalid."
				c.Recommendation = "The shopper should try again with another valid card or use another payment method."
			case "04", "07", "41", "43", "200":
				c.FailureCode = "SUSPECTED_FRAUD"
				c.FailureMessage = "Card payment failed. \n\nThe transaction was declined by the issuer due to the card being marked as stolen or potential fraud."
				c.Recommendation = "The card was reported as lost, the shopper should be validated for authenticity and be referred to their issuer."
			case "34", "59", "83":
				c.FailureCode = "SUSPECTED_FRAUD"
				c.FailureMessage = "Card payment failed. \n\nThe transaction was declined by channel due to the account being blocked or suspected as fraud."
				c.Recommendation = "The channel has declined the transaction due to suspicion of fraud, the shopper should be validated for authenticity and be referred to their issuer."
			case "51", "116", "121":
				c.FailureCode = "INSUFFICIENT_FUND"
				c.FailureMessage = "Card payment failed. \n\nThe transaction was declined by the issuer due to credit limit or balance is not sufficient."
				c.Recommendation = "Insufficient funds in the cardholder's account. The shopper can try again after adding funds to their bank account, or use another payment method."
			case "19", "80", "90", "91", "96", "911":
				c.FailureCode = "CHANNEL_UNAVAILABLE"
				c.FailureMessage = "Card payment failed. \n\nThe transaction failed due to the issuer being unavailable or having a system malfunction."
				c.Recommendation = "The issuing bank cannot be contacted. The shopper should try again or use another payment method."
			}
		}
	}
}

func (c *ChargeResponse) SetCaptureHistoriesFromJSON() {
	if c.CaptureHistoriesJSON.Valid {
		_ = json.Unmarshal(c.CaptureHistoriesJSON.JSONText, &c.CaptureHistories)

		// Sort descending as alternative sql issues.
		sort.Slice(c.CaptureHistories, func(i, j int) bool {
			return c.CaptureHistories[i].CreatedAt.After(c.CaptureHistories[j].CreatedAt)
		})
	}
}

func (c *ChargeResponse) RemoveUnusedResponse() {
	if c.Card != nil {
		c.Card.MIDInfo, c.Card.Device, c.Card.Error = nil, nil, nil
		c.Card.MerchantCategoryCode, c.Card.Description, c.Card.SettlementDate = "", "", ""
		c.Card.ResponseCode, c.Card.ACSURL, c.Card.CardHolderName, c.Card.Fingerprint, c.Card.SaveForFutureUse, c.Card.BankMerchantID, c.Card.CardName = nil, "", "", "", nil, "", ""

		if c.Card.AuthenticationResult != nil {
			// auto split scenario
			c.Card.AuthenticationResult.CallbackTransactionID = ""

			// authentication for external 3ds
			c.Card.AuthenticationResult.TransactionID = ""
			c.Card.AuthenticationResult.TransactionStatus = ""
			c.Card.AuthenticationResult.AuthenticationScheme = ""
			c.Card.AuthenticationResult.AcsTransactionID = ""
			c.Card.AuthenticationResult.AcsReference = ""
			c.Card.AuthenticationResult.AuthenticationTime = nil
		}
	}
	if c.Ewallet != nil {
		c.Ewallet.AppRedirectURL, c.Ewallet.WebRedirectURL, c.Ewallet.ReferenceNo, c.Ewallet.PartnerReferenceNo = "", "", "", ""
	}
	if c.ProcessorReferenceID != "" {
		c.ProcessorReferenceID = ""
	}

	if c.VirtualAccount != nil {
		c.VirtualAccount.BankReferenceNo = ""
	}

	if c.Qr != nil {
		c.Qr.StoreID = ""
	}
}

func (p *UnifiedPaymentSessionResponse) SetPaymentSimulationForStaging(config *config.Config) {
	if config.Environment != constant.EnvironmentStaging || p.PaymentMethod == nil ||
		!slices.Contains([]string{constant.UnifiedPaymentSessionStatusRequireAction, constant.UnifiedStaticPaymentStatusActive}, p.Status) {
		return
	}

	if !slices.Contains([]string{constant.UnifiedPaymentMethodQris, constant.UnifiedPaymentMethodVA}, p.PaymentMethod.Type) {
		return
	}

	// Check if Metadata is nil or not a map, and initialize if needed
	if p.Metadata == nil {
		p.Metadata = map[string]any{}
	}

	if data, ok := p.Metadata.(map[string]any); ok {
		data[constant.PaymentSimulatorKey] = fmt.Sprintf(
			config.MerchantPortalConfig.PaymentSimulationPatternURL,
			base64.StdEncoding.EncodeToString([]byte(p.ID)),
		)
	}
}

func (p *UnifiedPaymentSessionResponse) SetPaymentURLForAPIMode() {
	if p.Mode != constant.UnifiedPaymentModeAPI || p.PaymentMethod == nil || len(p.ChargeDetails) == 0 {
		return
	}

	if (p.PaymentMethod.Type == constant.UnifiedPaymentMethodQris && p.PaymentType == constant.UnifiedPaymentTypeMultiple) ||
		p.IsSubsequentRecurringPayment() {
		return
	}

	if p.PaymentMethod.Type != constant.UnifiedPaymentMethodCard && p.PaymentMethod.Type != constant.UnifiedPaymentMethodEWallet {
		return
	}

	for _, charge := range p.ChargeDetails {
		if charge.Card != nil {
			p.PaymentUrl = charge.Card.ACSURL
			return

		} else if charge.Ewallet != nil {
			// should not replace the url and use existing url due to processing event trigger
			// because dana did not have processing callback. need to revisit the shopee pay too
			if charge.Ewallet.Channel == constant.UnifiedPaymentEWalletDanaAcquirer {
				return
			}

			p.PaymentUrl = charge.Ewallet.WebRedirectURL
			if charge.Ewallet.Channel == constant.UnifiedPaymentEWalletShopeePayAcquirer {
				p.PaymentUrl = charge.Ewallet.AppRedirectURL
			}
			return
		}
	}
}

func (p *UnifiedPaymentSessionResponse) SetEncryptionKeyForCard() {
	if !p.AutoConfirm && p.Mode == constant.UnifiedPaymentModeAPI && p.PaymentMethod != nil && p.PaymentMethod.Type == constant.UnifiedPaymentMethodCard {
		publicKey, privateKey, _ := encrypt.RSAKeyPairWithBase64Fmt(2048)
		p.EncryptionKey = encrypt.NewRSAEncryptedKey(publicKey, privateKey)
	}
}

func (p *UnifiedPaymentSessionResponse) IsSubsequentRecurringPayment() bool {
	return p.RecurringID != "" && !p.InitiateFirstAuthorization
}

func (p *UnifiedPaymentSessionResponse) UnmarshalJSON(data []byte) error {
	type Alias UnifiedPaymentSessionResponse // Prevent recursion
	raw := &struct {
		Metadata json.RawMessage `json:"metadata"` // intercept just this one
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if len(raw.Metadata) == 0 {
		return nil
	}

	var m map[string]any
	if err := json.Unmarshal(raw.Metadata, &m); err != nil {
		return fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	// Use the inner "value" if present and it's a map
	if v, ok := m["value"].(map[string]any); ok {
		p.Metadata = &v
	} else {
		p.Metadata = &m
	}

	return nil
}

func (c *ChargeResponse) GetCardExpiryMonth() string {
	var expMonth string
	if c.Card.ExpMonth != "" {
		if month := c.Card.ExpMonth.String(); month != "" && len(month) == 2 {
			expMonth = month
		}
	}
	return expMonth
}

func (c *ChargeResponse) GetCardExpiryYear() string {
	expYear := ""
	if c.Card.ExpYear != "" {
		if year, _ := c.Card.ExpYear.Int64(); year != 0 {
			expYear = fmt.Sprintf("%v", year)
		}
		if len(expYear) > 2 {
			expYear = expYear[2:]
		}
	}
	return expYear
}

func (c *ChargeResponse) SetAuthorizedAmount(amount *Amount) {
	c.AuthorizedAmount = amount
}

func (p *UnifiedPaymentSessionResponse) IsStaticPayment() bool {
	return p.PaymentType == constant.UnifiedPaymentTypeMultiple
}

func AccountTransactionToChargeResponse(charge *orchestratorModel.AccountTransactionWithUseCase) *ChargeResponse {
	chargeStatus := ""
	statementDescriptor := ""
	chargeMethodDetails := &ChargePaymentMethodDetails{}
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

	chargeResp := &ChargeResponse{
		ID:                              charge.UUID.String(),
		PaymentSessionID:                charge.ReferenceID,
		PaymentSessionClientReferenceID: charge.ClientReferenceID,
		Amount: Amount{
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

	chargeResp.SetFailureDetail()
	chargeResp.RemoveUnusedResponse()

	switch chargeStatus {
	case constant.ChargeStatusSuccess:
		chargeResp.IsCaptured = true
		chargeResp.AuthorizedAmount = &chargeResp.Amount
		chargeResp.CapturedAmount = &chargeResp.Amount
	case constant.ChargeStatusWaitingForCapture:
		chargeResp.CapturedAmount = &chargeResp.Amount
	}

	return chargeResp
}

func (c *ChargeResponse) ToPbChargeResponse() *pb.UnifiedPaymentV2CallbackRequest_ChargeResponse {
	pbCharge := &pb.UnifiedPaymentV2CallbackRequest_ChargeResponse{
		Id:                              c.ID,
		PaymentSessionId:                c.PaymentSessionID,
		PaymentSessionClientReferenceId: c.PaymentSessionClientReferenceID,
		Amount: &pb.UnifiedPaymentV2CallbackRequest_Amount{
			Value:    c.Amount.Value,
			Currency: c.Amount.Currency,
		},
		StatementDescriptor: c.StatementDescriptor,
		Status:              c.Status,
		IsCaptured:          c.IsCaptured,
		FailureCode:         c.FailureCode,
		FailureMessage:      c.FailureMessage,
		Recommendation:      c.Recommendation,
		CreatedAt:           timestamppb.New(c.CreatedAt),
		UpdatedAt:           timestamppb.New(c.UpdatedAt),
		NetworkResponseCode: c.NetworkResponseCode,
	}

	if c.AuthorizedAmount != nil {
		pbCharge.AuthorizedAmount = &pb.UnifiedPaymentV2CallbackRequest_Amount{
			Value:    c.AuthorizedAmount.Value,
			Currency: c.AuthorizedAmount.Currency,
		}
	}

	if c.CapturedAmount != nil {
		pbCharge.CapturedAmount = &pb.UnifiedPaymentV2CallbackRequest_Amount{
			Value:    c.CapturedAmount.Value,
			Currency: c.CapturedAmount.Currency,
		}
	}

	if c.PaidAt != nil {
		pbCharge.PaidAt = timestamppb.New(*c.PaidAt)
	}

	if c.FdsRiskAssessment != nil {
		pbCharge.FdsRiskAssessment = &pb.UnifiedPaymentV2CallbackRequest_FdsRiskAssessment{
			Score:          c.FdsRiskAssessment.Score.InexactFloat64(),
			Level:          c.FdsRiskAssessment.Level,
			Recommendation: c.FdsRiskAssessment.Recommendation,
			Status:         c.FdsRiskAssessment.Status,
			EvaluatedAt:    c.FdsRiskAssessment.EvaluatedAt.Format(time.RFC3339),
		}
	}

	// some payment can be cancelled from require action
	// should handle the method detail carefully due to struct embedding (composition)
	if c.ChargePaymentMethodDetails != nil && c.Ewallet != nil {
		pbCharge.Ewallet = &pb.UnifiedPaymentV2CallbackRequest_PaymentMethodOptionEwallet{
			Channel: c.Ewallet.Channel,
		}
	}

	if c.ChargePaymentMethodDetails != nil && c.Qr != nil {
		pbCharge.Qr = &pb.UnifiedPaymentV2CallbackRequest_ChargePaymentMethodDetailQr{
			Acquirer:                 c.Qr.Acquirer,
			QrContent:                c.Qr.QrContent,
			QrUrl:                    c.Qr.QrUrl,
			QrType:                   c.Qr.QrType,
			RetrievalReferenceNumber: c.Qr.RetrievalReferenceNumber,
			IssuerName:               c.Qr.IssuerName,
			ExpiryAt:                 timestamppb.New(c.Qr.ExpiryAt),
			MerchantName:             c.Qr.MerchantName,
		}
	}

	if c.ChargePaymentMethodDetails != nil && c.VirtualAccount != nil {
		vaProto := &pb.UnifiedPaymentV2CallbackRequest_ChargePaymentMethodDetailVirtualAccount{
			Channel:               c.VirtualAccount.Channel,
			VirtualAccountNumber:  c.VirtualAccount.VirtualAccountNumber,
			VirtualAccountName:    c.VirtualAccount.VirtualAccountName,
			VirtualAccountTrxType: c.VirtualAccount.VirtualAccountTrxType,
			ExpiryAt:              timestamppb.New(c.VirtualAccount.ExpiryAt),
		}
		if c.VirtualAccount.BankReferenceNo != "" {
			vaProto.BankReferenceNo = c.VirtualAccount.BankReferenceNo
		}
		pbCharge.VirtualAccount = vaProto
	}

	if c.ChargePaymentMethodDetails != nil && c.Card != nil {
		pbCharge.Card = &pb.UnifiedPaymentV2CallbackRequest_ChargePaymentMethodDetailCard{
			First6:         c.Card.First6,
			First8:         c.Card.First8,
			Last4:          c.Card.Last4,
			CardHolderName: c.Card.CardHolderName,
			AcsUrl:         c.Card.ACSURL,
			BankMerchantId: c.Card.BankMerchantID,
			BinInformations: &pb.UnifiedPaymentV2CallbackRequest_ChargePaymentMethodDetailBinInformation{
				Type:        c.Card.BinInformations.Type,
				IssuingBank: c.Card.BinInformations.IssuingBank,
				Brand:       c.Card.BinInformations.Brand,
				Country:     c.Card.BinInformations.Country,
			},
			ApprovalCode: c.Card.ApprovalCode,
		}

		// Handle ExpMonth
		if c.Card.ExpMonth != "" {
			pbCharge.Card.ExpMonth = c.Card.ExpMonth.String()
		}

		// Handle ExpYear
		if c.Card.ExpYear != "" {
			pbCharge.Card.ExpYear = c.Card.ExpYear.String()
		}

		if c.Card.AuthenticationResult != nil {
			pbCharge.Card.AuthenticationResult = &pb.UnifiedPaymentV2CallbackRequest_ChargePaymentMethodDetailCardAuthenticationResult{
				ThreeDsVersion: c.Card.AuthenticationResult.ThreeDsVersion,
				ThreeDsResult:  c.Card.AuthenticationResult.ThreeDsResult,
				ThreeDsMethod:  c.Card.AuthenticationResult.ThreeDsMethod,
				EciCode:        c.Card.AuthenticationResult.EciCode,
			}
		}

		if c.Card.AuthorizationResult != nil {
			pbCharge.Card.AuthorizationResult = &pb.UnifiedPaymentV2CallbackRequest_ChargePaymentMethodDetailCardAuthorizationResult{
				AcquirerReferenceNumber:  c.Card.AuthorizationResult.AcquirerReferenceNumber,
				RetrievalReferenceNumber: c.Card.AuthorizationResult.RetrievalReferenceNumber,
				Stan:                     c.Card.AuthorizationResult.Stan,
				AvsResult:                c.Card.AuthorizationResult.AvsResult,
				CvvResult:                c.Card.AuthorizationResult.CvvResult,
				AuthorizedAmount: &pb.UnifiedPaymentV2CallbackRequest_Amount{
					Value:    c.Card.AuthorizationResult.AuthorizedAmount.Value,
					Currency: c.Card.AuthorizationResult.AuthorizedAmount.Currency,
				},
				IssuerAuthorizationCode: c.Card.AuthorizationResult.IssuerAuthorizationCode,
				NetworkTransactionId:    c.Card.AuthorizationResult.NetworkTransactionID,
			}
		}

		if c.Card.ResponseCode != nil {
			pbCharge.Card.ResponseCode = &pb.UnifiedPaymentV2CallbackRequest_ChargePaymentMethodDetailCardResponseCode{
				GatewayCode:           c.Card.ResponseCode.GatewayCode,
				GatewayRecommendation: c.Card.ResponseCode.GatewayRecommendation,
			}
		}
	}

	if len(c.CaptureHistories) > 0 {
		pbCapturedHistories := make([]*pb.UnifiedPaymentV2CallbackRequest_CaptureHistoryResponse, len(c.CaptureHistories))
		for i, captureHistory := range c.CaptureHistories {
			pbCapturedHistories[i] = &pb.UnifiedPaymentV2CallbackRequest_CaptureHistoryResponse{
				Id:             captureHistory.ID,
				Status:         captureHistory.Status,
				Currency:       captureHistory.Currency,
				CapturedAmount: captureHistory.CapturedAmount,
				CreatedAt:      timestamppb.New(captureHistory.CreatedAt),
			}
		}
		pbCharge.CaptureHistories = pbCapturedHistories
	}

	return pbCharge
}

func (c *ChargeResponse) SetCaptureHistories(paymentCaptures []*paymentCaptureModel.PaymentCapture) {
	// Set captureHistories to the chargeResp if any
	if len(paymentCaptures) > 0 {
		capturedHistories := make([]*paymentCaptureModel.CaptureHistoryResponse, len(paymentCaptures))
		for i, capture := range paymentCaptures {
			capturedHistories[i] = &paymentCaptureModel.CaptureHistoryResponse{
				ID:             capture.ID,
				Status:         capture.Status,
				Currency:       capture.Currency,
				CapturedAmount: capture.Amount,
				CreatedAt:      capture.CreatedAt,
			}
		}
		c.CaptureHistories = capturedHistories
	}
}

type CustomerPaymentMethodResponse struct {
	Token          string    `json:"token"`
	PaymentMethod  string    `json:"paymentMethod"`
	PaymentChannel string    `json:"paymentChannel"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`

	Card *CustomerPaymentMethodCardResponse `json:"card,omitempty"`
}

type CustomerPaymentMethodCardResponse struct {
	Fingerprint         string      `json:"fingerprint,omitempty"`
	Network             string      `json:"network"`
	First6              string      `json:"first6"`
	First8              string      `json:"first8"`
	Last4               string      `json:"last4"`
	ExpMonth            interface{} `json:"expMonth"`
	ExpYear             interface{} `json:"expYear"`
	CardHolderFirstName string      `json:"cardHolderFirstName"`
	CardHolderLastName  string      `json:"cardHolderLastName"`
	CardHolderEmail     string      `json:"cardHolderEmail,omitempty"`
	CardHolderPhone     string      `json:"cardHolderPhone,omitempty"`
}

type CustomerInformationResponse struct {
	CustomerID string `json:"-"`
	GivenName  string `json:"givenName" `
	// will have both Surname and Surname
	// for backward compatibility
	Surname              *string                          `json:"surname,omitempty"`
	SureName             string                           `json:"sureName"`
	Email                string                           `json:"email" validate:"required,email"`
	PhoneNumber          *UnifiedPaymentPhoneNumber       `json:"phoneNumber" validate:"omitempty"`
	RefundPreference     *UnifiedPaymentRefundPreference  `json:"refundPreference,omitempty" validate:"omitempty"`
	StoredPaymentMethods []*CustomerPaymentMethodResponse `json:"storedPaymentMethods,omitempty"`
}

// PaymentStatusHistoryResponse represents the status history response for payment sessions
type PaymentStatusHistoryResponse struct {
	Status         string     `json:"status"`
	Label          string     `json:"label"`
	Description    string     `json:"description,omitempty"`
	Recommendation string     `json:"recommendation,omitempty"`
	Timestamp      *time.Time `json:"timestamp,omitempty"`
}

// ChargeStatusHistoryResponse represents the status history response for charges
type ChargeStatusHistoryResponse struct {
	Status         string     `json:"status"`
	Label          string     `json:"label"`
	Description    string     `json:"description,omitempty"`
	Recommendation string     `json:"recommendation,omitempty"`
	Timestamp      *time.Time `json:"timestamp,omitempty"`
}

// RefundStatusHistoryResponse represents the status history response for refunds
type RefundStatusHistoryResponse struct {
	Status         string     `json:"status"`
	Label          string     `json:"label"`
	Description    string     `json:"description,omitempty"`
	Recommendation string     `json:"recommendation,omitempty"`
	Timestamp      *time.Time `json:"timestamp,omitempty"`
}

// RemoveCardTokenizationResponse is a structure for the return value when the card tokenization is successfully removed from the customer
type RemoveCardTokenizationResponse struct {
	CustomerID string `json:"customerId"`
	TokenID    string `json:"tokenId"`
}

type CaptureResponse struct {
	ID                              string    `json:"id"`
	PaymentSessionID                string    `json:"paymentSessionId"`
	PaymentSessionClientReferenceId string    `json:"paymentSessionClientReferenceId"`
	ReleaseRemainingAmount          bool      `json:"releaseRemainingAmount"`
	Amount                          *Amount   `json:"amount"`
	Status                          string    `json:"status"` // SUCCESS, PENDING, FAILED
	CreatedAt                       time.Time `json:"createdAt"`
	UpdatedAt                       time.Time `json:"updatedAt"`
}

type GeneratePaymentTokenResponse struct {
	Token string `json:"token"`
}

type GetBinDetailResponse struct {
	BIN       string `json:"bin"`
	CardType  string `json:"cardType"`
	Principal string `json:"principal"`
	CardLevel string `json:"cardLevel"`
	Issuer    string `json:"issuer"`
	Country   string `json:"country"`
	Currency  string `json:"currency"`
}

type InquiryResult struct {
	Status                 string
	UpdatedStatus          bool
	LastInquiryAt          *time.Time
	ResponseCode           string
	ResponseMessage        string
	Amount                 *Amount
	ProcessorID            string
	ProcessorTransactionID string
	TrxDatetime            *time.Time
	ProcessorReferenceNo   string
}
