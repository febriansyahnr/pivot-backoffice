package ledgerController

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {
	claim := merchant.MerchantAuthTokenClaims{
		MerchantId: uuid.NewString(),
	}
	payload := ledger_model.UpdateLedgerEntryRequest{
		Usecase:           constant.ReferenceDisbursement,
		Status:            constant.StatusSuccess,
		ReasonDescription: "test",
		ReasonType:        "other",
	}

	testCases := []struct {
		Name             string
		GetRequest       func() []byte
		MockSetup        func(svc *mockSvc.ILedgerService)
		SetHeaders       func(req *http.Request)
		SetUrlParam      bool
		Claim            *merchant.MerchantAuthTokenClaims
		WantErr          bool
		ExpectedCode     int
		ExpectedResponse string
	}{
		{
			Name: "SUCCESS: Update Success Disbursement",
			MockSetup: func(svc *mockSvc.ILedgerService) {
				svc.On("BulkUpdateLedgerEntry", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			Claim:       &claim,
			SetUrlParam: true,
			GetRequest: func() []byte {
				payload := ledger_model.UpdateLedgerEntryRequest{
					ReferenceID:       uuid.New(),
					Usecase:           constant.ReferenceDisbursement,
					Status:            constant.StatusSuccess,
					ReasonDescription: "",
					ReasonType:        "",
				}
				b, _ := json.Marshal(payload)
				return b
			},
			WantErr:          false,
			ExpectedCode:     200,
			ExpectedResponse: `{"code":"00","message":"OK","data":null}`,
		},
		{
			Name: "SUCCESS: Update Fail Disbursement",
			MockSetup: func(svc *mockSvc.ILedgerService) {
				svc.On("BulkUpdateLedgerEntry", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			Claim:       &claim,
			SetUrlParam: true,
			GetRequest: func() []byte {
				payload := ledger_model.UpdateLedgerEntryRequest{
					ReferenceID:       uuid.New(),
					Usecase:           constant.ReferenceDisbursement,
					Status:            constant.StatusFailed,
					ReasonDescription: "Failed description",
					ReasonType:        "Failed Type",
				}
				b, _ := json.Marshal(payload)
				return b
			},
			WantErr:          false,
			ExpectedCode:     200,
			ExpectedResponse: `{"code":"00","message":"OK","data":null}`,
		},
		{
			Name: "SUCCESS: Update Success Wallet Withdrawal",
			MockSetup: func(svc *mockSvc.ILedgerService) {
				svc.On("BulkUpdateLedgerEntry", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			Claim:       &claim,
			SetUrlParam: true,
			GetRequest: func() []byte {
				payload := ledger_model.UpdateLedgerEntryRequest{
					ReferenceID:       uuid.New(),
					Usecase:           constant.ReferenceWallet,
					Status:            constant.StatusSuccess,
					ReasonDescription: "",
					ReasonType:        "",
				}
				b, _ := json.Marshal(payload)
				return b
			},
			WantErr:          false,
			ExpectedCode:     200,
			ExpectedResponse: `{"code":"00","message":"OK","data":null}`,
		},
		{
			Name: "SUCCESS: Update Fail Wallet Withdrawal",
			MockSetup: func(svc *mockSvc.ILedgerService) {
				svc.On("BulkUpdateLedgerEntry", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			Claim:       &claim,
			SetUrlParam: true,
			GetRequest: func() []byte {
				payload := ledger_model.UpdateLedgerEntryRequest{
					ReferenceID:       uuid.New(),
					Usecase:           constant.ReferenceWallet,
					Status:            constant.StatusFailed,
					ReasonDescription: "Failed description",
					ReasonType:        "Failed Type",
				}
				b, _ := json.Marshal(payload)
				return b
			},
			WantErr:          false,
			ExpectedCode:     200,
			ExpectedResponse: `{"code":"00","message":"OK","data":null}`,
		},
		{
			Name: "SUCCESS: With Submerchant ID",
			MockSetup: func(svc *mockSvc.ILedgerService) {
				svc.On("BulkUpdateLedgerEntry", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			SetHeaders: func(req *http.Request) {
				req.Header.Set("X-Merchant-Id", uuid.NewString())
			},
			Claim:       &claim,
			SetUrlParam: true,
			GetRequest: func() []byte {
				b, _ := json.Marshal(payload)
				return b
			},
			WantErr:          false,
			ExpectedCode:     200,
			ExpectedResponse: `{"code":"00","message":"OK","data":null}`,
		},
		{
			Name: "ERROR: No URL param",
			MockSetup: func(svc *mockSvc.ILedgerService) {
			},
			Claim:       &claim,
			SetUrlParam: false,
			GetRequest: func() []byte {
				b, _ := json.Marshal(payload)
				return b
			},
			WantErr:          true,
			ExpectedCode:     400,
			ExpectedResponse: `{"code":"40","message":"invalid UUID length: 0","error":{"type":"API_ERROR","message":"invalid UUID length: 0","recommendation":""},"data":null}`,
		},
		{
			Name: "ERROR: Record Transactions",
			MockSetup: func(svc *mockSvc.ILedgerService) {
				svc.On("BulkUpdateLedgerEntry", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error"))
			},
			Claim:       &claim,
			SetUrlParam: true,
			GetRequest: func() []byte {
				b, _ := json.Marshal(payload)
				return b
			},
			WantErr:          true,
			ExpectedCode:     500,
			ExpectedResponse: `{"code":"99","message":"error","error":{"type":"UNKNOWN","message":"error","recommendation":""},"data":null}`,
		},
		{
			Name: "ERROR: Decode",
			MockSetup: func(svc *mockSvc.ILedgerService) {
			},
			Claim:       &claim,
			SetUrlParam: true,
			GetRequest: func() []byte {
				b, _ := json.Marshal("payload")
				return b
			},
			WantErr:          true,
			ExpectedCode:     400,
			ExpectedResponse: ` {"code":"40","message":"invalid request payload","error":{"type":"API_ERROR","message":"invalid request payload","recommendation":""},"data":null}`,
		},
		{
			Name: "ERROR: Failed validator",
			MockSetup: func(svc *mockSvc.ILedgerService) {
			},
			Claim:       &claim,
			SetUrlParam: true,
			GetRequest: func() []byte {
				b, _ := json.Marshal(ledger_model.CreateNewLedgerEntryRequest{})
				return b
			},
			WantErr:          true,
			ExpectedCode:     400,
			ExpectedResponse: ` {"code":"40","message":"invalid request payload","error":{"type":"API_ERROR","message":"invalid request payload","recommendation":""},"data":null}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()
			ledgerSvc := mockSvc.NewILedgerService(t)

			if tc.MockSetup != nil {
				tc.MockSetup(ledgerSvc)
			}

			baseUrl := "/api/internal/v2/ledger/" + uuid.New().String()
			req := httptest.NewRequest(http.MethodPut, baseUrl, bytes.NewReader(tc.GetRequest()))
			chiRouterCtx := chi.NewRouteContext()
			chiRouterCtx.URLParams.Add("referenceId", uuid.New().String())

			if tc.SetUrlParam {
				ctx = context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx)
				req = req.WithContext(ctx)
			}

			if tc.Claim != nil {
				ctx = context.WithValue(ctx, constant.CtxMerchantInfo, tc.Claim)
				req = req.WithContext(ctx)
			}

			if tc.SetHeaders != nil {
				tc.SetHeaders(req)
			}

			ctrl := New(ledgerSvc)

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(ctrl.Update)
			handler.ServeHTTP(httpRecorder, req)

			if !assert.Equal(t, tc.ExpectedCode, httpRecorder.Code) {
				t.Logf("Response: %s", httpRecorder.Body.String())
			}

			assert.JSONEqf(t, tc.ExpectedResponse, httpRecorder.Body.String(), "want: %v, got: %v", tc.ExpectedResponse, httpRecorder.Body.String())
		})
	}
}
