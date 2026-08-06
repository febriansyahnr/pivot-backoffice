package callback_model

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchantTopUp"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/callback"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/virtualCard"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/walletTopUp"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/walletUserActivation"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/pkg/callback/model"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type RegisterCallbackRequest struct {
	MerchantID  uuid.UUID `json:"merchantId"`               // From bearer token
	Name        string    `json:"name" validate:"required"` // Virtual Account, Disbursement, Bulk Disbursement
	BaseURL     string    `json:"baseUrl"`                  // For SNAP VA Callback
	URL         string    `json:"url" validate:"required"`
	Description string    `json:"description" validate:"required"`
}

func (req RegisterCallbackRequest) ToCallbackMaster() *CallbackMaster {
	return &CallbackMaster{
		UUID:        uuid.New(),
		Name:        req.Name,
		Description: req.Description,
	}
}

func (req RegisterCallbackRequest) ToCallback(callbackMasterUUID uuid.UUID) *Callback {
	return &Callback{
		UUID:             uuid.New(),
		CallbackMasterID: callbackMasterUUID,
		MerchantID:       req.MerchantID,
		BaseURL:          &req.BaseURL,
		URL:              req.URL,
		Description:      req.Description,
	}
}

func (req RegisterCallbackRequest) ToResponse(
	calbackMaster *CallbackMaster,
	callbackID uuid.UUID,
) *RegisterCallbackResponse {
	return &RegisterCallbackResponse{
		CallbackMasterID: calbackMaster.UUID,
		CallbackName:     calbackMaster.Name,
		CallbackID:       callbackID,
		BaseURL:          req.BaseURL,
		URL:              req.URL,
		Description:      req.Description,
	}
}

type ProcessCallbackRequest struct {
	Name       string    `json:"name"`
	Event      string    `json:"event"`
	MerchantID uuid.UUID `json:"merchantId"`
	Request    any       `json:"request"`
	IsSnap     bool      `json:"isSnap"`
}

func (dst *ProcessCallbackRequest) Bind(request *anypb.Any) error {
	protoJsonBytes, err := protojson.Marshal(request)
	if err != nil {
		return fmt.Errorf("protojson marshal: %v", err)
	}

	if dst.Name == constant.CallbackNameDisbursement {
		dst.Request = &disbursementModel.DisbursementDataCallbackRequest{}

	} else if dst.Name == constant.CallbackMasterPaymentSNAPQRIS {
		if strings.Contains(request.GetTypeUrl(), "ChargeResponse") {
			dst.Request = &unifiedPaymentModel.ChargeResponse{}

		} else if dst.Event == "PAYMENT.EXPIRED" {
			dst.Request = &paymentModel.PaymentSnapQrisExpiredCallbackRequest{
				AdditionalInfo: map[string]any{},
			}
		} else {
			dst.Request = &paymentModel.PaymentQrisCallbackRequest{}
		}
	} else if dst.Name == constant.CallbackNameXB {
		dst.Request = &xbModel.GetXbPayoutDetailResponse{}

	} else if dst.Event == constant.CallbackEventPaymentCreditcardPaid {
		dst.Request = &creditcardModel.SendCallbackPaymentNotificationDataRequest{}

	} else if dst.Event == constant.CallbackEventPaymentVirtualAccountPaid {
		if strings.Contains(request.GetTypeUrl(), "ChargeResponse") {
			dst.Request = &unifiedPaymentModel.ChargeResponse{}

		} else {
			dst.Request = &paymentModel.PaymentResponse{}
		}
	} else if dst.Event == constant.CallbackEventVirtualCardNotification ||
		dst.Event == constant.CallbackEventVirtualCardVisaNotification ||
		dst.Event == constant.CallbackEventPhysicalCardVisaNotification {
		dst.Request = &virtualCard.CallbackResponse{}

	} else if util.IsPatternMatch(constant.UnifiedPaymentCallbackEventPattern, dst.Event) {
		dst.Request = &paymentModel.UnifiedPaymentCallbackRequest{}

		// Check if trx is unified payment v2
		var req pb.UnifiedPaymentV2CallbackRequest
		if err := anypb.UnmarshalTo(request, &req, proto.UnmarshalOptions{}); err == nil {
			dst.Request = &unifiedPaymentModel.UnifiedPaymentSessionResponse{}
		}

	} else if util.IsPatternMatch(constant.UnifiedPaymentChargeCallbackEventPattern, dst.Event) {
		dst.Request = &unifiedPaymentModel.ChargeResponse{}

	} else if dst.Name == constant.CallbackNameWalletTopup {
		dst.Request = &walletTopUp.TopUpCallbackRequest{}

	} else if dst.Event == constant.CallbackEventWalletUserActivation || dst.Event == constant.CallbackEventWalletUserActivationKYC {
		dst.Request = &walletUserActivation.UserActivationCallbackRequest{}

	} else if dst.Event == constant.CallbackEventWalletActivateAccountLinkage {
		dst.Request = &walletUserActivation.AccountLinkageRequest{}

	} else if dst.Event == constant.CallbackEventWalletTransaction {
		dst.Request = &walletUserActivation.WalletTransaction{}

	} else if dst.Name == constant.CallbackNameMerchantTopUp {
		dst.Request = &merchantTopUp.MerchantTopUpCallbackRequest{}

	} else if util.Contains([]string{
		constant.CallbackEventRefundPending,
		constant.CallbackEventRefundSuccess,
		constant.CallbackEventRefundFailed,
		constant.CallbackEventRefundCancelled,
		constant.CallbackEventRefundWaitingBankTransfer,
	}, dst.Event) {
		dst.Request = &refundModel.RefundResponse{}

	} else if dst.Name == constant.CallbackNameSubAccountRegistration {
		dst.Request = &merchant.SubAccountRegistrationCallback{}

	} else if dst.Name == constant.CallbackNameWithdrawal {
		dst.Request = &withdrawal.WithdrawalStatusCallbackRequest{}

	} else if dst.Event == constant.CallbackEventWalletSNAPQrisMPM {
		dst.Request = &walletUserActivation.SnapQrisMpmCallbackRequest{}

	} else if dst.Event == constant.CallbackEventWalletSNAPDirectDebit {
		dst.Request = &walletUserActivation.SnapDirectDebitCallbackRequest{}

	} else {
		return errors.New("mapping request data not found")
	}

	if err := json.Unmarshal(protoJsonBytes, dst.Request); err != nil {
		return fmt.Errorf("json unmarshal: %v", err)
	}

	return nil
}

func (p ProcessCallbackRequest) ToWorkflowPreparationResponse(callbackId, callbackUrl string) WorkflowMerchantCallbackPreparationResponse {

	var payload []byte

	if p.IsSnap {
		switch val := p.Request.(type) {
		case *paymentModel.UnifiedPaymentResponse:
			p.Request = val.ToSnapPayment()

		case *unifiedPaymentModel.ChargeResponse:
			p.Request = val.ToSnapPayment()

		default:
			if constant.IsPaymentVA(p.Event) {

				// Re-mapping callback request data
				var paymentResponse paymentModel.PaymentResponse
				raw, _ := json.Marshal(val)
				_ = json.Unmarshal(raw, &paymentResponse)

				// Convert to new model
				p.Request = paymentResponse.ToSnapVAPaymentResponse()
			}
		}
		payload, _ = json.Marshal(p.Request)

	} else {
		// Non-SNAP Request Payload
		payload, _ = json.Marshal(callbackModel.CallbackPayloadRequest{
			Event: p.Event,
			Data:  p.Request,
		})
	}
	return WorkflowMerchantCallbackPreparationResponse{
		Name:        p.Name,
		EventName:   p.Event,
		MerchantId:  p.MerchantID.String(),
		Request:     base64.StdEncoding.EncodeToString(payload),
		IsSnap:      p.IsSnap,
		ReferenceId: GetReferenceId(p.Event, p.Request),
		CallbackId:  callbackId,
		CallbackUrl: callbackUrl,
	}
}

func GetReferenceId(event string, request any) *string {
	if request == nil {
		return nil
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil
	}

	var data map[string]any

	if err = json.Unmarshal(requestBytes, &data); err != nil {
		return nil
	}

	// Helper function
	getRef := func(m map[string]any, key string) *string {
		if val, ok := m[key]; ok {
			if str, ok := val.(string); ok && str != "" {
				return &str
			}
		}
		return nil
	}

	// Payout events, use uuid as referenceId
	if strings.HasPrefix(event, "PAYOUT.") || strings.HasPrefix(event, "MERCHANT-TOP-UP.") {
		if ref := getRef(data, "uuid"); ref != nil {
			return ref
		}
		if ref := getRef(data, "batchId"); ref != nil {
			return ref
		}
	}

	// Merchant withdraw event
	if strings.HasPrefix(event, "WITHDRAW.") {
		if wd, ok := data["withdrawal"].(map[string]any); ok {
			if ref, ok := wd["referenceId"].(string); ok && ref != "" {
				return &ref
			}
		}
	}

	// Keys to check in order of priority
	keys := []string{
		"clientReferenceId",               // payments/refunds
		"paymentSessionClientReferenceId", // unified payment charge responses
		"originalPartnerReferenceNo",      // payment requests
		"referenceId",                     // disbursements, payouts, withdrawals
		"user_reference_id",               // wallet top-ups
		"merchantReferenceId",             // orchestrator transactions
		"partnerReferenceNo",              // wallet account linkage
	}

	// Check keys in priority
	for _, k := range keys {
		if ref := getRef(data, k); ref != nil {
			return ref
		}
	}

	// Special case: nested additionalInfo.referenceId
	if addInfo, ok := data["additionalInfo"].(map[string]any); ok {
		if ref := getRef(addInfo, "referenceId"); ref != nil {
			return ref
		}
	}
	return nil
}

type GetListCallbackFilterRequest struct {
	MerchantID *string `json:"merchantId"`
}

type CallbackURLSettingReq struct {
	Info       *http.Request `validate:"-"`
	MerchantID string        `validate:"required,uuid"`
	UserID     string        `validate:"required,uuid"`
	MasterName string        `validate:"-"`
}

type GetListCallbackLogFilterRequest struct {
	MerchantID     string     `json:"-"`
	Type           string     `json:"type"`
	Event          string     `json:"event"`
	StartUpdatedAt *time.Time `json:"startCreatedAt"`
	EndUpdatedAt   *time.Time `json:"endCreatedAt"`
	Status         string     `json:"status"`
	Keyword        string     `json:"keyword"` // Searches both UUID and reference_id
}

type TestAndSaveCallbackURLReq struct {
	Name             string                               `json:"name" validate:"required"`
	URL              string                               `json:"url" validate:"required,url"`
	Payload          callbackModel.CallbackPayloadRequest `json:"payload" validate:"required"`
	CallbackMasterID string                               `json:"-" validate:"required,uuid"`
	MerchantID       string                               `json:"-" validate:"required,uuid"`
	UserID           string                               `json:"-" validate:"required,uuid"`
	Info             *http.Request                        `json:"-" validate:"-"`
}

type SendMerchantCallbackRequest struct {
	EventName   string          `json:"eventName"`
	MerchantId  string          `json:"merchantId"`
	IsSnap      bool            `json:"isSnap"`
	Payload     string          `json:"payload"` // Base64 encoding request body
	RawPayload  json.RawMessage `json:"-"`
	CallbackId  string          `json:"callbackId"`
	CallbackUrl string          `json:"callbackUrl"`
}

type ResendCallbackRequest struct {
	MerchantID        string `json:"merchantId" validate:"required,uuid"`
	Type              string `json:"type" validate:"required,oneof=PAYMENT DISBURSEMENT"`
	ClientReferenceID string `json:"clientReferenceId" validate:"required_without=ReferenceID"`
	ReferenceID       string `json:"referenceId" validate:"omitempty,uuid"`
}

type ResendCallbackResponse struct {
	Message           string `json:"message"`
	Type              string `json:"type"`
	ClientReferenceID string `json:"clientReferenceId,omitempty"`
	ReferenceID       string `json:"referenceId,omitempty"`
}
