package callback_model

import (
	"errors"
	"fmt"
	"testing"
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
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestRegisterCallbackRequestToCallback(t *testing.T) {
	req := RegisterCallbackRequest{
		MerchantID:  uuid.New(),
		Name:        constant.TypePayment,
		BaseURL:     "https://paper.id",
		URL:         "/callback",
		Description: "Request Callback For Payment",
	}

	callback := req.ToCallback(uuid.New())

	if callback.MerchantID != req.MerchantID {
		t.Errorf("expected MerchantID %v, got %v", req.MerchantID, callback.MerchantID)
	}

	if *callback.BaseURL != req.BaseURL {
		t.Errorf("expected BaseURL %v, got %v", req.BaseURL, *callback.BaseURL)
	}

	if callback.URL != req.URL {
		t.Errorf("expected URL %v, got %v", req.URL, callback.URL)
	}

	if callback.Description != req.Description {
		t.Errorf("[RegisterCallback] expected Description %v, got %v", req.Description, callback.Description)
	}
}

func TestRegisterCallbackRequestToCallbackMaster(t *testing.T) {
	req := RegisterCallbackRequest{
		Name:        constant.TypePayment,
		Description: "Register Callback for Payment",
	}

	callbackMaster := req.ToCallbackMaster()

	if callbackMaster.Name != req.Name {
		t.Errorf("expected Name %v, got %v", req.Name, callbackMaster.Name)
	}

	if callbackMaster.Description != req.Description {
		t.Errorf("[ToCallbackMaster] expected Description %v, got %v", req.Description, callbackMaster.Description)
	}
}

func TestRegisterCallbackRequestToResponse(t *testing.T) {
	req := RegisterCallbackRequest{
		MerchantID:  uuid.New(),
		Name:        constant.TypePayment,
		BaseURL:     "https://paper.id/register",
		URL:         "/callback",
		Description: "Register Callback for Virtual Account",
	}

	callbackMaster := &CallbackMaster{
		UUID:        uuid.New(),
		Name:        req.Name,
		Description: req.Description,
	}

	callbackID := uuid.New()

	response := req.ToResponse(callbackMaster, callbackID)

	if response.CallbackMasterID != callbackMaster.UUID {
		t.Errorf("expected CallbackMasterID %v, got %v", callbackMaster.UUID, response.CallbackMasterID)
	}

	if response.CallbackName != callbackMaster.Name {
		t.Errorf("expected CallbackName %v, got %v", callbackMaster.Name, response.CallbackName)
	}

	if response.CallbackID != callbackID {
		t.Errorf("expected CallbackID %v, got %v", callbackID, response.CallbackID)
	}

	if response.BaseURL != req.BaseURL {
		t.Errorf("expected BaseURL %v, got %v", req.BaseURL, response.BaseURL)
	}

	if response.URL != req.URL {
		t.Errorf("expected URL %v, got %v", req.URL, response.URL)
	}

	if response.Description != req.Description {
		t.Errorf("expected Description %v, got %v", req.Description, response.Description)
	}
}

func TestCallbackLogWithMasterMarshalJSON(t *testing.T) {
	baseURL := "http://test"
	event := "PAYOUT.DONE"
	response := "{}"

	callbackLogWithMaster := CallbackLogWithMaster{
		UUID:       uuid.New(),
		CallbackID: uuid.New(),
		Type:       "test",
		BaseURL:    &baseURL,
		URL:        "www.test.com",
		Event:      &event,
		Request:    "{}",
		Response:   &response,
		Status:     "DELIVERED",
		Retry:      0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		MerchantID: uuid.NewString(),
	}

	_, err := callbackLogWithMaster.MarshalJSON()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCallbackLogWithMaster_ToCallbackLog(t *testing.T) {
	baseURL := "http://test"
	event := "PAYOUT.DONE"
	response := "{}"

	callbackLogWithMaster := CallbackLogWithMaster{
		UUID:       uuid.New(),
		CallbackID: uuid.New(),
		Type:       "test",
		BaseURL:    &baseURL,
		URL:        "www.test.com",
		Event:      &event,
		Request:    "{}",
		Response:   &response,
		Status:     "DELIVERED",
		Retry:      0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		MerchantID: uuid.NewString(),
	}

	callbackLog := callbackLogWithMaster.ToCallbackLog()

	if callbackLog.UUID != callbackLogWithMaster.UUID {
		t.Errorf("expected UUID %v, got %v", callbackLogWithMaster.UUID, callbackLog.UUID)
	}

	if callbackLog.CallbackID != callbackLogWithMaster.CallbackID {
		t.Errorf("expected CallbackID %v, got %v", callbackLogWithMaster.CallbackID, callbackLog.CallbackID)
	}

	if callbackLog.Event != callbackLogWithMaster.Event {
		t.Errorf("expected Event %v, got %v", callbackLogWithMaster.Event, callbackLog.Event)
	}

	if callbackLog.Request != callbackLogWithMaster.Request {
		t.Errorf("expected Request %v, got %v", callbackLogWithMaster.Request, callbackLog.Request)
	}

	if callbackLog.Response != callbackLogWithMaster.Response {
		t.Errorf("expected Response %v, got %v", callbackLogWithMaster.Response, callbackLog.Response)
	}

	if callbackLog.Status != callbackLogWithMaster.Status {
		t.Errorf("expected Status %v, got %v", callbackLogWithMaster.Status, callbackLog.Status)
	}

	if callbackLog.Retry != callbackLogWithMaster.Retry {
		t.Errorf("expected Retry %v, got %v", callbackLogWithMaster.Retry, callbackLog.Retry)
	}

	if callbackLog.CreatedAt != callbackLogWithMaster.CreatedAt {
		t.Errorf("expected CreatedAt %v, got %v", callbackLogWithMaster.CreatedAt, callbackLog.CreatedAt)
	}

	if callbackLog.UpdatedAt != callbackLogWithMaster.UpdatedAt {
		t.Errorf("expected UpdatedAt %v, got %v", callbackLogWithMaster.UpdatedAt, callbackLog.UpdatedAt)
	}
}

func TestProcessCallbackRequestBinding(t *testing.T) {
	tests := []struct {
		name          string
		callbackName  string
		callbackEvent string
		request       *anypb.Any
		wantError     error
		wantResult    any
	}{
		{
			name:         "SUCCESS:Disbursement",
			callbackName: constant.CallbackNameDisbursement,
			wantResult:   &disbursementModel.DisbursementDataCallbackRequest{},
		},
		{
			name:         "SUCCESS:Payment SNAP QRIS",
			callbackName: constant.CallbackMasterPaymentSNAPQRIS,
			request: func() *anypb.Any {
				request, _ := anypb.New(&pb.UnifiedPaymentV2CallbackRequest_ChargeResponse{})
				return request
			}(),
			wantResult: &unifiedPaymentModel.ChargeResponse{},
		},
		{
			name:         "SUCCESS:Payment QRIS",
			callbackName: constant.CallbackMasterPaymentSNAPQRIS,
			wantResult:   &paymentModel.PaymentQrisCallbackRequest{},
		},
		{
			name:         "SUCCESS:International payout",
			callbackName: constant.CallbackNameXB,
			wantResult:   &xbModel.GetXbPayoutDetailResponse{},
		},
		{
			name:          "SUCCESS:Payment card paid",
			callbackEvent: constant.CallbackEventPaymentCreditcardPaid,
			wantResult:    &creditcardModel.SendCallbackPaymentNotificationDataRequest{},
		},
		{
			name:          "SUCCESS:Payment VA paid",
			callbackEvent: constant.CallbackEventPaymentVirtualAccountPaid,
			wantResult:    &paymentModel.PaymentResponse{},
		},
		{
			name:          "SUCCESS:Payment VA paid (unified payment)",
			callbackEvent: constant.CallbackEventPaymentVirtualAccountPaid,
			request: func() *anypb.Any {
				request, _ := anypb.New(&pb.UnifiedPaymentV2CallbackRequest_ChargeResponse{})
				return request
			}(),
			wantResult: &unifiedPaymentModel.ChargeResponse{},
		},
		{
			name:          "SUCCESS:Virtual card notification",
			callbackEvent: constant.CallbackEventVirtualCardNotification,
			wantResult:    &virtualCard.CallbackResponse{},
		},
		{
			name:          "SUCCESS:Payment for general unified payment notification v1",
			callbackEvent: fmt.Sprintf(constant.CallbackEventUnifiedPaymentPattern, "PROCESSING"),
			wantResult:    &paymentModel.UnifiedPaymentCallbackRequest{},
		},
		{
			name:          "SUCCESS:Payment for general unified payment notification v2",
			callbackEvent: fmt.Sprintf(constant.CallbackEventUnifiedPaymentPattern, "PROCESSING"),
			request: func() *anypb.Any {
				request, _ := anypb.New(&pb.UnifiedPaymentV2CallbackRequest{})
				return request
			}(),
			wantResult: &unifiedPaymentModel.UnifiedPaymentSessionResponse{},
		},
		{
			name:          "SUCCESS:Charge detail",
			callbackEvent: fmt.Sprintf(constant.CallbackEventUnifiedPaymentChargePattern, "TEST"),
			wantResult:    &unifiedPaymentModel.ChargeResponse{},
		},
		{
			name:         "SUCCESS:Wallet topup",
			callbackName: constant.CallbackNameWalletTopup,
			wantResult:   &walletTopUp.TopUpCallbackRequest{},
		},
		{
			name:          "SUCCESS:Wallet user activation",
			callbackEvent: constant.CallbackEventWalletUserActivation,
			wantResult:    &walletUserActivation.UserActivationCallbackRequest{},
		},
		{
			name:          "SUCCESS:Wallet activate account linkage",
			callbackEvent: constant.CallbackEventWalletActivateAccountLinkage,
			wantResult:    &walletUserActivation.AccountLinkageRequest{},
		},
		{
			name:          "SUCCESS:Wallet transaction",
			callbackEvent: constant.CallbackEventWalletTransaction,
			wantResult:    &walletUserActivation.WalletTransaction{},
		},
		{
			name:         "SUCCESS:Merchant topup",
			callbackName: constant.CallbackNameMerchantTopUp,
			wantResult:   &merchantTopUp.MerchantTopUpCallbackRequest{},
		},
		{
			name:          "SUCCESS:Refund",
			callbackEvent: constant.CallbackEventRefundSuccess,
			wantResult:    &refundModel.RefundResponse{},
		},
		{
			name:         "SUCCESS:Sub account registration",
			callbackName: constant.CallbackNameSubAccountRegistration,
			wantResult:   &merchant.SubAccountRegistrationCallback{},
		},
		{
			name:         "SUCCESS:Withdraw balance",
			callbackName: constant.CallbackNameWithdrawal,
			wantResult:   &withdrawal.WithdrawalStatusCallbackRequest{},
		},
		{
			name:          "SUCCESS:SNAP QRIS expired",
			callbackName:  constant.CallbackMasterPaymentSNAPQRIS,
			callbackEvent: "PAYMENT.EXPIRED",
			wantResult: &paymentModel.PaymentSnapQrisExpiredCallbackRequest{
				AdditionalInfo: map[string]any{},
			},
		},
		{
			name:         "ERROR:Unmapped request",
			callbackName: "XXXX",
			wantResult:   nil,
			wantError:    errors.New("mapping request data not found"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			callback := &ProcessCallbackRequest{
				Name:  test.callbackName,
				Event: test.callbackEvent,
			}
			assert.Equal(t, test.wantError, callback.Bind(test.request))
			assert.Equal(t, test.wantResult, callback.Request)
		})
	}
}

func TestProcessCallbackRequestToWorkflowPreparationResponse(t *testing.T) {

	var (
		callbackId  = "e0f3a06b-7434-4ccd-b6d3-4b1dc2c9977c"
		callbackUrl = "http://localhost/events"              // NOSONAR
		merchantId  = "00000000-0000-0000-0000-000000000000" // NOSONAR
	)

	tests := []struct {
		request    ProcessCallbackRequest
		wantResult WorkflowMerchantCallbackPreparationResponse
	}{
		{
			request: ProcessCallbackRequest{
				Name:  constant.CallbackNameDisbursement,
				Event: constant.CallbackEventPayoutDone,
			},
			wantResult: WorkflowMerchantCallbackPreparationResponse{
				Name:        constant.CallbackNameDisbursement,
				EventName:   constant.CallbackEventPayoutDone,
				MerchantId:  merchantId,
				Request:     "eyJldmVudCI6IlBBWU9VVC5ET05FIiwiZGF0YSI6bnVsbH0=",
				CallbackId:  callbackId,
				CallbackUrl: callbackUrl,
			},
		},
		{
			request: ProcessCallbackRequest{
				Name:    constant.CallbackNamePayment,
				Event:   constant.CallbackEventPaymentQrisMpmPaid,
				IsSnap:  true,
				Request: &paymentModel.UnifiedPaymentResponse{},
			},
			wantResult: WorkflowMerchantCallbackPreparationResponse{
				Name:        constant.CallbackNamePayment,
				EventName:   constant.CallbackEventPaymentQrisMpmPaid,
				MerchantId:  merchantId,
				IsSnap:      true,
				Request:     "eyJ1dWlkIjoiIiwibWVyY2hhbnRJZCI6IiIsInJlZmVyZW5jZUlkIjoiIiwic3RhdHVzIjoiIiwicGFpZEFtb3VudCI6eyJjdXJyZW5jeSI6IiIsInZhbHVlIjoiIn0sImFtb3VudCI6eyJjdXJyZW5jeSI6IiIsInZhbHVlIjoiIn0sInBheW1lbnRUeXBlRGV0YWlsIjp7fX0=",
				CallbackId:  callbackId,
				CallbackUrl: callbackUrl,
			},
		},
		{
			request: ProcessCallbackRequest{
				Name:   constant.CallbackNamePayment,
				Event:  fmt.Sprintf(constant.CallbackEventUnifiedPaymentChargePattern, "SUCCESS"),
				IsSnap: true,
				Request: &unifiedPaymentModel.ChargeResponse{
					ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{},
				},
			},
			wantResult: WorkflowMerchantCallbackPreparationResponse{
				Name:        constant.CallbackNamePayment,
				EventName:   fmt.Sprintf(constant.CallbackEventUnifiedPaymentChargePattern, "SUCCESS"),
				MerchantId:  merchantId,
				IsSnap:      true,
				Request:     "eyJpZCI6IiIsInBheW1lbnRTZXNzaW9uSWQiOiIiLCJwYXltZW50U2Vzc2lvbkNsaWVudFJlZmVyZW5jZUlkIjoiIiwiYW1vdW50Ijp7InZhbHVlIjowLCJjdXJyZW5jeSI6IiJ9LCJzdGF0ZW1lbnREZXNjcmlwdG9yIjoiIiwic3RhdHVzIjoiIiwiYXV0aG9yaXplZEFtb3VudCI6bnVsbCwiY2FwdHVyZWRBbW91bnQiOm51bGwsImlzQ2FwdHVyZWQiOmZhbHNlLCJjcmVhdGVkQXQiOiIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsInVwZGF0ZWRBdCI6IjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwicGFpZEF0IjpudWxsfQ==",
				CallbackId:  callbackId,
				CallbackUrl: callbackUrl,
			},
		},
		{
			request: ProcessCallbackRequest{
				Name:   constant.CallbackNamePayment,
				Event:  constant.CallbackEventPaymentVirtualAccountPaid,
				IsSnap: true,
				Request: &paymentModel.PaymentResponse{
					TransactionDate: util.ValueToPtr(time.Date(2025, 10, 10, 0, 0, 0, 0, time.UTC)),
					VirtualAccount:  &paymentModel.PaymentVirtualAccountResponse{},
				},
			},
			wantResult: WorkflowMerchantCallbackPreparationResponse{
				Name:        constant.CallbackNamePayment,
				EventName:   constant.CallbackEventPaymentVirtualAccountPaid,
				MerchantId:  merchantId,
				IsSnap:      true,
				Request:     "eyJ0cnhJZCI6IiIsInZpcnR1YWxBY2NvdW50Tm8iOiIiLCJ2aXJ0dWFsQWNjb3VudE5hbWUiOiIiLCJwYWlkQW1vdW50IjpudWxsLCJ0cnhEYXRlVGltZSI6IjIwMjUtMTAtMTBUMDc6MDA6MDArMDc6MDAiLCJhZGRpdGlvbmFsSW5mbyI6eyJyZWZlcmVuY2VJZCI6IiIsImlzc3VlciI6IiIsInZpcnR1YWxBY2NvdW50VHJ4VHlwZSI6IiIsInZhU3RhdHVzIjoiQUNUSVZFIiwicGF5bWVudFN0YXR1cyI6IlNVQ0NFU1MifX0=",
				CallbackId:  callbackId,
				CallbackUrl: callbackUrl,
			},
		},
		{
			request: ProcessCallbackRequest{
				Name:   constant.CallbackNameMerchantTopUp,
				Event:  constant.CallbackEventMerchantTopUpSuccess,
				IsSnap: true,
				Request: &merchantTopUp.MerchantTopUpCallbackRequest{
					UUID: "341f378c-85b6-4ef7-80a5-1debdb6baa46",
				},
			},
			wantResult: WorkflowMerchantCallbackPreparationResponse{
				Name:        constant.CallbackNameMerchantTopUp,
				EventName:   constant.CallbackEventMerchantTopUpSuccess,
				MerchantId:  merchantId,
				IsSnap:      true,
				Request:     "eyJ1dWlkIjoiMzQxZjM3OGMtODViNi00ZWY3LTgwYTUtMWRlYmRiNmJhYTQ2IiwibWVyY2hhbnRJZCI6IiIsIm1lcmNoYW50TmFtZSI6IiIsImFjY291bnROYW1lIjoiIiwiYW1vdW50Ijp7ImN1cnJlbmN5IjoiIiwidmFsdWUiOiIifSwiYmFsYW5jZUJlZm9yZSI6eyJjdXJyZW5jeSI6IiIsInZhbHVlIjoiIn0sImJhbGFuY2VBZnRlciI6eyJjdXJyZW5jeSI6IiIsInZhbHVlIjoiIn0sInBheW1lbnRNZXRob2QiOnsidHlwZSI6IiJ9LCJwYXltZW50TWV0aG9kT3B0aW9ucyI6e30sInRyYW5zYWN0aW9uVGltZSI6IjAwMDEtMDEtMDFUMDA6MDA6MDBaIn0=",
				CallbackId:  callbackId,
				CallbackUrl: callbackUrl,
				ReferenceId: util.ValueToPtr("341f378c-85b6-4ef7-80a5-1debdb6baa46"),
			},
		},
	}
	for _, test := range tests {
		result := test.request.ToWorkflowPreparationResponse(callbackId, callbackUrl)

		assert.Equal(t, test.wantResult, result)
	}
}

func TestGetReferenceId(t *testing.T) {
	tests := []struct {
		input           ProcessCallbackRequest
		wantReferenceId *string
	}{
		{
			input: ProcessCallbackRequest{
				Request: nil,
			},
			wantReferenceId: nil,
		},
		{
			input: ProcessCallbackRequest{
				Request: make(chan int, 1),
			},
			wantReferenceId: nil,
		},
		{
			input: ProcessCallbackRequest{
				Request: []map[string]any{{"message": "OK"}},
			},
			wantReferenceId: nil,
		},
		{
			input: ProcessCallbackRequest{
				Event: constant.CallbackEventPayoutDone,
				Request: disbursementModel.DisbursementDataCallbackRequest{
					UUID: "a7821217-2941-423c-a716-b35f490ebb66", // NOSONAR
				},
			},
			wantReferenceId: util.ValueToPtr("a7821217-2941-423c-a716-b35f490ebb66"), // NOSONAR
		},
		{
			input: ProcessCallbackRequest{
				Event: constant.CallbackEventPayoutCancelled,
				Request: disbursementModel.DisbursementDataCallbackRequest{
					UUID: "92316dfd-23c6-4946-a809-d3731bae6917", // NOSONAR
				},
			},
			wantReferenceId: util.ValueToPtr("92316dfd-23c6-4946-a809-d3731bae6917"), // NOSONAR
		},
		{
			input: ProcessCallbackRequest{
				Request: &unifiedPaymentModel.ChargeResponse{},
			},
			wantReferenceId: nil,
		},
		{
			input: ProcessCallbackRequest{
				Request: &xbModel.GetXbPayoutDetailResponse{
					ReferenceId: "70deff43-ad52-432f-91cf-37a622c82dc1", // NOSONAR
				},
			},
			wantReferenceId: util.ValueToPtr("70deff43-ad52-432f-91cf-37a622c82dc1"), // NOSONAR
		},
		{
			input: ProcessCallbackRequest{
				Request: &paymentModel.PaymentResponse{
					ReferenceID: "10656803-f75f-43d2-95f7-c8295f60fcc6", // NOSONAR
				},
			},
			wantReferenceId: util.ValueToPtr("10656803-f75f-43d2-95f7-c8295f60fcc6"), // NOSONAR
		},
		{
			input: ProcessCallbackRequest{
				Request: &walletTopUp.TopUpCallbackRequest{
					UserReferenceId: "c8d8c53d-4ff1-41f7-a676-489fe69b8ece", // NOSONAR
				},
			},
			wantReferenceId: util.ValueToPtr("c8d8c53d-4ff1-41f7-a676-489fe69b8ece"), // NOSONAR
		},
		{
			input: ProcessCallbackRequest{
				Request: &refundModel.RefundResponse{
					ClientReferenceID: "cd3bf52e-2148-4bf5-a845-39791e62af53", // NOSONAR
				},
			},
			wantReferenceId: util.ValueToPtr("cd3bf52e-2148-4bf5-a845-39791e62af53"), // NOSONAR
		},
		{
			input: ProcessCallbackRequest{
				Event: fmt.Sprintf(constant.CallbackEventWithdrawPattern, "SUCCESS"),
				Request: &withdrawal.OpenAPIWithdrawalResponse{
					Withdrawal: withdrawal.OpenAPIWithdrawalDetailResponse{
						ReferenceID: "REF/20251022/0007",
					},
				},
			},
			wantReferenceId: util.ValueToPtr("REF/20251022/0007"), // NOSONAR
		},
		{
			input: ProcessCallbackRequest{
				Event:   fmt.Sprintf(constant.CallbackEventWithdrawPattern, "SUCCESS"),
				Request: &withdrawal.OpenAPIWithdrawalResponse{},
			},
			wantReferenceId: nil, // NOSONAR
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.wantReferenceId, GetReferenceId(test.input.Event, test.input.Request))
	}
}
