package tnc_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
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

func TestHistory(t *testing.T) {
	claims := &userModel.UserTokenClaims{UUID: "user-1", MerchantId: "merchant-1", Email: "u@merchant.com"}

	withClaims := func(ctx context.Context) context.Context {
		return context.WithValue(ctx, constant.CtxUserInfoKey, claims)
	}

	tests := []struct {
		name           string
		target         string
		ctxSetup       func(context.Context) context.Context
		mockSetup      func(*mockService.ITNCService)
		expectedStatus int
	}{
		{
			name:     "SUCCESS: default pagination",
			target:   "/api/v1/tnc/history",
			ctxSetup: withClaims,
			mockSetup: func(svc *mockService.ITNCService) {
				svc.On("GetSigningHistory", mock.Anything, mock.AnythingOfType("*tnc.SigningHistoryQuery")).
					Return(&commonModel.PaginationResponse{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "SUCCESS: with version and pagination params",
			target:   "/api/v1/tnc/history?version=1.0.0&page=2&perPage=5",
			ctxSetup: withClaims,
			mockSetup: func(svc *mockService.ITNCService) {
				svc.On("GetSigningHistory", mock.Anything, mock.AnythingOfType("*tnc.SigningHistoryQuery")).
					Return(&commonModel.PaginationResponse{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ERROR: no claims in context",
			target:         "/api/v1/tnc/history",
			ctxSetup:       func(ctx context.Context) context.Context { return ctx },
			mockSetup:      func(svc *mockService.ITNCService) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "ERROR: merchant id missing from claims",
			target: "/api/v1/tnc/history",
			ctxSetup: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, constant.CtxUserInfoKey, &userModel.UserTokenClaims{})
			},
			mockSetup:      func(svc *mockService.ITNCService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "ERROR: invalid page (zero)",
			target:         "/api/v1/tnc/history?page=0",
			ctxSetup:       withClaims,
			mockSetup:      func(svc *mockService.ITNCService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "ERROR: invalid page (non numeric)",
			target:         "/api/v1/tnc/history?page=abc",
			ctxSetup:       withClaims,
			mockSetup:      func(svc *mockService.ITNCService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "ERROR: invalid perPage",
			target:         "/api/v1/tnc/history?perPage=0",
			ctxSetup:       withClaims,
			mockSetup:      func(svc *mockService.ITNCService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "ERROR: service failure",
			target:   "/api/v1/tnc/history",
			ctxSetup: withClaims,
			mockSetup: func(svc *mockService.ITNCService) {
				svc.On("GetSigningHistory", mock.Anything, mock.AnythingOfType("*tnc.SigningHistoryQuery")).
					Return(nil, errPkg.New(response.HttpErrInternal, errors.New("list failed")))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mockService.NewITNCService(t)
			tc.mockSetup(mockSvc)

			controller := tnc.New(mockSvc, validator.New())

			req := httptest.NewRequest(http.MethodGet, tc.target, nil)
			req = req.WithContext(tc.ctxSetup(req.Context()))
			rr := httptest.NewRecorder()

			controller.History(rr, req)

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
