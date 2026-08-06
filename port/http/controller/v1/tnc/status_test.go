package tnc_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	tncModel "github.com/paper-indonesia/pivot-backoffice/internal/model/tnc"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/tnc"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStatus(t *testing.T) {
	claims := &userModel.UserTokenClaims{UUID: "user-1", MerchantId: "merchant-1", Email: "u@merchant.com"}

	tests := []struct {
		name           string
		ctxSetup       func(context.Context) context.Context
		mockSetup      func(*mockService.ITNCService)
		expectedStatus int
	}{
		{
			name: "SUCCESS: returns signing status",
			ctxSetup: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxUserInfoKey, claims)
			},
			mockSetup: func(svc *mockService.ITNCService) {
				svc.On("GetTNCSigningStatus", mock.Anything, "merchant-1").
					Return(&tncModel.TNCSigningStatus{IsSigned: true, ActiveVersion: "1.2.0"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ERROR: no claims in context",
			ctxSetup:       func(ctx context.Context) context.Context { return ctx },
			mockSetup:      func(svc *mockService.ITNCService) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "ERROR: merchant id missing from claims",
			ctxSetup: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxUserInfoKey, &userModel.UserTokenClaims{})
			},
			mockSetup:      func(svc *mockService.ITNCService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "ERROR: service failure",
			ctxSetup: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxUserInfoKey, claims)
			},
			mockSetup: func(svc *mockService.ITNCService) {
				svc.On("GetTNCSigningStatus", mock.Anything, "merchant-1").
					Return(nil, errPkg.New(response.HttpErrDatabase, errors.New("db down")))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mockService.NewITNCService(t)
			tc.mockSetup(mockSvc)

			controller := tnc.New(mockSvc, validator.New())

			req := httptest.NewRequest(http.MethodGet, "/api/v1/tnc/status", nil)
			req = req.WithContext(tc.ctxSetup(req.Context()))
			rr := httptest.NewRecorder()

			controller.Status(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			if tc.expectedStatus == http.StatusOK {
				var resp response.ApiResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
				assert.Equal(t, response.HttpStatusOK, resp.Code)
			}
			mockSvc.AssertExpectations(t)
		})
	}
}
