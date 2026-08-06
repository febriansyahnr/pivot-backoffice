package ledgerController

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	payload := ledger_model.CreateNewLedgerEntryRequest{
		ReferenceID:          uuid.NewString(),
		Usecase:              constant.ReferenceDisbursement,
		TransactionType:      constant.TypeTopUp,
		Channel:              constant.ChannelVirtualAccount,
		Remarks:              "test",
		TransactionTimestamp: time.Now(),
		Amount:               1000,
		Currency:             "IDR",
		TransferType:         constant.TransferTypePayIn,
	}

	validHeader := func(req *http.Request) {
		req.Header.Set(constant.HeaderXMerchantId, uuid.NewString())
	}

	testCases := []struct {
		Name             string
		GetRequest       func() []byte
		MockSetup        func(svc *mockSvc.ILedgerService)
		SetHeaders       func(req *http.Request)
		Claim            *merchant.MerchantAuthTokenClaims
		WantErr          bool
		ExpectedCode     int
		ExpectedResponse string
	}{
		{
			Name: "SUCCESS: Disbursement",
			MockSetup: func(svc *mockSvc.ILedgerService) {
				svc.On("RecordTransaction", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			GetRequest: func() []byte {
				payload := ledger_model.CreateNewLedgerEntryRequest{
					ReferenceID:          uuid.NewString(),
					Usecase:              constant.ReferenceDisbursement,
					TransactionType:      constant.TypeDisbursement,
					Channel:              constant.ChannelBankTransfer,
					Remarks:              "Disbursement payout",
					TransactionTimestamp: time.Now(),
					Amount:               1000,
					Currency:             constant.CurrencyIDR,
					TransferType:         constant.TransferTypePayOut,
					MoneyFlowType:        constant.MoneyFlowIndirect,
					SenderAccountID:      uuid.New(),
					ParentAccountID:      uuid.New(),
				}
				b, _ := json.Marshal(payload)
				return b
			},
			SetHeaders:       validHeader,
			WantErr:          false,
			ExpectedCode:     200,
			ExpectedResponse: `{"code":"00","message":"OK","data":null}`,
		},
		{
			Name: "SUCCESS: Disbursement TopUp",
			MockSetup: func(svc *mockSvc.ILedgerService) {
				svc.On("RecordTransaction", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			GetRequest: func() []byte {
				payload := ledger_model.CreateNewLedgerEntryRequest{
					ReferenceID:          uuid.NewString(),
					Usecase:              constant.ReferenceDisbursement,
					TransactionType:      constant.TypeTopUp,
					Channel:              constant.ChannelVirtualAccount,
					Remarks:              "Disbursement top up balance",
					TransactionTimestamp: time.Now(),
					Amount:               1000,
					Currency:             constant.CurrencyIDR,
					TransferType:         constant.TransferTypePayIn,
					MoneyFlowType:        constant.MoneyFlowIndirect,
					SenderAccountID:      uuid.New(),
					ParentAccountID:      uuid.New(),
				}
				b, _ := json.Marshal(payload)
				return b
			},
			SetHeaders:       validHeader,
			WantErr:          false,
			ExpectedCode:     200,
			ExpectedResponse: `{"code":"00","message":"OK","data":null}`,
		},
		{
			Name: "SUCCESS: Disbursement Top Up Fee",
			MockSetup: func(svc *mockSvc.ILedgerService) {
				svc.On("RecordTransaction", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			GetRequest: func() []byte {
				payload := ledger_model.CreateNewLedgerEntryRequest{
					ReferenceID:          uuid.NewString(),
					Usecase:              constant.ReferenceDisbursement,
					TransactionType:      constant.TypeFee,
					Channel:              constant.ChannelVirtualAccount,
					Remarks:              "Disbursement top up fee",
					TransactionTimestamp: time.Now(),
					Amount:               1000,
					Currency:             constant.CurrencyIDR,
					TransferType:         constant.TransferTypePayIn,
					MoneyFlowType:        constant.MoneyFlowIndirect,
					SenderAccountID:      uuid.New(),
					ParentAccountID:      uuid.New(),
				}
				b, _ := json.Marshal(payload)
				return b
			},
			SetHeaders:       validHeader,
			WantErr:          false,
			ExpectedCode:     200,
			ExpectedResponse: `{"code":"00","message":"OK","data":null}`,
		},
		{
			Name: "SUCCESS: Disbursement Fee",
			MockSetup: func(svc *mockSvc.ILedgerService) {
				svc.On("RecordTransaction", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			GetRequest: func() []byte {
				payload := ledger_model.CreateNewLedgerEntryRequest{
					ReferenceID:          uuid.NewString(),
					Usecase:              constant.ReferenceDisbursement,
					TransactionType:      constant.TypeFee,
					Channel:              constant.ChannelVirtualAccount,
					Remarks:              "Disbursement fee",
					TransactionTimestamp: time.Now(),
					Amount:               1000,
					Currency:             constant.CurrencyIDR,
					TransferType:         constant.TransferTypePayOut,
					MoneyFlowType:        constant.MoneyFlowIndirect,
					SenderAccountID:      uuid.New(),
					ParentAccountID:      uuid.New(),
				}
				b, _ := json.Marshal(payload)
				return b
			},
			SetHeaders:       validHeader,
			WantErr:          false,
			ExpectedCode:     200,
			ExpectedResponse: `{"code":"00","message":"OK","data":null}`,
		},
		{
			Name: "SUCCESS: Payment",
			MockSetup: func(svc *mockSvc.ILedgerService) {
				svc.On("RecordTransaction", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			GetRequest: func() []byte {
				payload := ledger_model.CreateNewLedgerEntryRequest{
					ReferenceID:          uuid.NewString(),
					Usecase:              constant.ReferencePayment,
					TransactionType:      constant.TypePayment,
					Channel:              constant.ChannelVirtualAccount,
					Remarks:              "Payment fee",
					TransactionTimestamp: time.Now(),
					Amount:               1000,
					Currency:             constant.CurrencyIDR,
					TransferType:         constant.TransferTypePayIn,
					MoneyFlowType:        constant.MoneyFlowIndirect,
					RecipientAccountID:   uuid.New(),
					ParentAccountID:      uuid.New(),
				}
				b, _ := json.Marshal(payload)
				return b
			},
			SetHeaders:       validHeader,
			WantErr:          false,
			ExpectedCode:     200,
			ExpectedResponse: `{"code":"00","message":"OK","data":null}`,
		},
		{
			Name: "SUCCESS: Payment Fee",
			MockSetup: func(svc *mockSvc.ILedgerService) {
				svc.On("RecordTransaction", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			GetRequest: func() []byte {
				payload := ledger_model.CreateNewLedgerEntryRequest{
					ReferenceID:          uuid.NewString(),
					Usecase:              constant.ReferencePayment,
					TransactionType:      constant.TypeFee,
					Channel:              constant.ChannelVirtualAccount,
					Remarks:              "Payment fee",
					TransactionTimestamp: time.Now(),
					Amount:               1000,
					Currency:             constant.CurrencyIDR,
					TransferType:         constant.TransferTypePayIn,
					MoneyFlowType:        constant.MoneyFlowIndirect,
					RecipientAccountID:   uuid.New(),
					ParentAccountID:      uuid.New(),
				}
				b, _ := json.Marshal(payload)
				return b
			},
			SetHeaders:       validHeader,
			WantErr:          false,
			ExpectedCode:     200,
			ExpectedResponse: `{"code":"00","message":"OK","data":null}`,
		},
		{
			Name: "SUCCESS: Wallet P2P Transfer",
			MockSetup: func(svc *mockSvc.ILedgerService) {
				svc.On("RecordTransaction", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			GetRequest: func() []byte {
				payload := ledger_model.CreateNewLedgerEntryRequest{
					ReferenceID:          uuid.NewString(),
					Usecase:              constant.ReferenceWallet,
					TransactionType:      constant.TypeWalletTransfer,
					Channel:              "",
					Remarks:              "Transfer to recipient",
					TransactionTimestamp: time.Now(),
					Amount:               1000,
					Currency:             constant.CurrencyIDR,
					TransferType:         constant.TransferTypeP2P,
					MoneyFlowType:        constant.MoneyFlowIndirect,
					SenderAccountID:      uuid.New(),
					ParentAccountID:      uuid.New(),
					RecipientAccountID:   uuid.New(),
				}
				b, _ := json.Marshal(payload)
				return b
			},
			SetHeaders:       validHeader,
			WantErr:          false,
			ExpectedCode:     200,
			ExpectedResponse: `{"code":"00","message":"OK","data":null}`,
		},
		{
			Name: "SUCCESS: Wallet PayIn TopUp",
			MockSetup: func(svc *mockSvc.ILedgerService) {
				svc.On("RecordTransaction", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			GetRequest: func() []byte {
				payload := ledger_model.CreateNewLedgerEntryRequest{
					ReferenceID:          uuid.NewString(),
					Usecase:              constant.ReferenceWallet,
					TransactionType:      constant.TypeWalletTopUp,
					Channel:              "",
					Remarks:              "Top up wallet",
					TransactionTimestamp: time.Now(),
					Amount:               1000,
					Currency:             constant.CurrencyIDR,
					TransferType:         constant.TransferTypePayIn,
					MoneyFlowType:        constant.MoneyFlowIndirect,
					RecipientAccountID:   uuid.New(),
					ParentAccountID:      uuid.New(),
				}
				b, _ := json.Marshal(payload)
				return b
			},
			SetHeaders:       validHeader,
			WantErr:          false,
			ExpectedCode:     200,
			ExpectedResponse: `{"code":"00","message":"OK","data":null}`,
		},
		{
			Name: "SUCCESS: Wallet PayOut Withdrawal",
			MockSetup: func(svc *mockSvc.ILedgerService) {
				svc.On("RecordTransaction", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			GetRequest: func() []byte {
				payload := ledger_model.CreateNewLedgerEntryRequest{
					ReferenceID:          uuid.NewString(),
					Usecase:              constant.ReferenceWallet,
					TransactionType:      constant.TypeWalletWithdrawal,
					Channel:              "",
					Remarks:              "withdraw balance",
					TransactionTimestamp: time.Now(),
					Amount:               1000,
					Currency:             constant.CurrencyIDR,
					TransferType:         constant.TransferTypePayOut,
					MoneyFlowType:        constant.MoneyFlowIndirect,
					SenderAccountID:      uuid.New(),
					ParentAccountID:      uuid.New(),
				}
				b, _ := json.Marshal(payload)
				return b
			},
			SetHeaders:       validHeader,
			WantErr:          false,
			ExpectedCode:     200,
			ExpectedResponse: `{"code":"00","message":"OK","data":null}`,
		},
		{
			Name: "ERROR: Record Transactions",
			MockSetup: func(svc *mockSvc.ILedgerService) {
				svc.On("RecordTransaction", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error"))
			},
			GetRequest: func() []byte {
				b, _ := json.Marshal(payload)
				return b
			},
			SetHeaders:       validHeader,
			WantErr:          true,
			ExpectedCode:     500,
			ExpectedResponse: `{"code":"99","message":"error","error":{"type":"UNKNOWN","message":"error","recommendation":""},"data":null}`,
		},
		{
			Name: "ERROR: Decode",
			MockSetup: func(svc *mockSvc.ILedgerService) {
			},
			GetRequest: func() []byte {
				b, _ := json.Marshal("payload")
				return b
			},
			SetHeaders:       validHeader,
			WantErr:          true,
			ExpectedCode:     400,
			ExpectedResponse: `{"code":"40","message":"json: cannot unmarshal string into Go value of type ledger_model.CreateNewLedgerEntryRequest","error":{"type":"API_ERROR","message":"json: cannot unmarshal string into Go value of type ledger_model.CreateNewLedgerEntryRequest","recommendation":""},"data":null}`,
		},
		{
			Name: "ERROR: Failed validator",
			MockSetup: func(svc *mockSvc.ILedgerService) {
			},
			GetRequest: func() []byte {
				b, _ := json.Marshal(ledger_model.CreateNewLedgerEntryRequest{})
				return b
			},
			SetHeaders:       validHeader,
			WantErr:          true,
			ExpectedCode:     400,
			ExpectedResponse: `{"code":"40","message":"Key: 'CreateNewLedgerEntryRequest.ReferenceID' Error:Field validation for 'ReferenceID' failed on the 'required' tag\nKey: 'CreateNewLedgerEntryRequest.Usecase' Error:Field validation for 'Usecase' failed on the 'required' tag\nKey: 'CreateNewLedgerEntryRequest.TransactionType' Error:Field validation for 'TransactionType' failed on the 'required' tag\nKey: 'CreateNewLedgerEntryRequest.TransactionTimestamp' Error:Field validation for 'TransactionTimestamp' failed on the 'required' tag\nKey: 'CreateNewLedgerEntryRequest.Amount' Error:Field validation for 'Amount' failed on the 'required' tag\nKey: 'CreateNewLedgerEntryRequest.Currency' Error:Field validation for 'Currency' failed on the 'required' tag\nKey: 'CreateNewLedgerEntryRequest.TransferType' Error:Field validation for 'TransferType' failed on the 'required' tag","error":{"type":"API_ERROR","message":"Key: 'CreateNewLedgerEntryRequest.ReferenceID' Error:Field validation for 'ReferenceID' failed on the 'required' tag\nKey: 'CreateNewLedgerEntryRequest.Usecase' Error:Field validation for 'Usecase' failed on the 'required' tag\nKey: 'CreateNewLedgerEntryRequest.TransactionType' Error:Field validation for 'TransactionType' failed on the 'required' tag\nKey: 'CreateNewLedgerEntryRequest.TransactionTimestamp' Error:Field validation for 'TransactionTimestamp' failed on the 'required' tag\nKey: 'CreateNewLedgerEntryRequest.Amount' Error:Field validation for 'Amount' failed on the 'required' tag\nKey: 'CreateNewLedgerEntryRequest.Currency' Error:Field validation for 'Currency' failed on the 'required' tag\nKey: 'CreateNewLedgerEntryRequest.TransferType' Error:Field validation for 'TransferType' failed on the 'required' tag","recommendation":""},"data":null}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()
			ledgerSvc := mockSvc.NewILedgerService(t)

			if tc.MockSetup != nil {
				tc.MockSetup(ledgerSvc)
			}

			baseUrl := "/api/internal/v2/ledger/"
			req := httptest.NewRequest(http.MethodPost, baseUrl, bytes.NewReader(tc.GetRequest()))
			chiRouterCtx := chi.NewRouteContext()

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			req = req.WithContext(ctx)

			if tc.SetHeaders != nil {
				tc.SetHeaders(req)
			}

			ctrl := New(ledgerSvc)

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(ctrl.Create)
			handler.ServeHTTP(httpRecorder, req)

			if !assert.Equal(t, tc.ExpectedCode, httpRecorder.Code) {
				t.Logf("Response => %v", httpRecorder.Body.String())
			}

			assert.JSONEqf(t, tc.ExpectedResponse, httpRecorder.Body.String(), "want: %v, got: %v", tc.ExpectedResponse, httpRecorder.Body.String())
		})
	}
}
