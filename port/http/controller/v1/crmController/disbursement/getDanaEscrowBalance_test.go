package crmDisbursementController

import (
	"net/http"
	"net/http/httptest"
	"testing"

	routingProcessorModelEscrowBalance "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/escrowBalance"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDanaEscrowBalance(t *testing.T) {
	testCases := []struct {
		name           string
		setupMock      func(req *http.Request, svc *serviceMocks.IRoutingProcessorService)
		expectedStatus int
	}{
		{
			name:           "success",
			expectedStatus: http.StatusOK,
			setupMock: func(req *http.Request, svc *serviceMocks.IRoutingProcessorService) {
				svc.On("GetDanaEscrowBalance", mock.Anything, mock.AnythingOfType("string")).Return(&routingProcessorModelEscrowBalance.EscrowBalanceResponse{
					ResponseCode:    "2001600",
					ResponseMessage: "Successful",
					BalanceAmount:   1000000,
				}, nil)
			},
		},
		{
			name:           "error internal server error",
			expectedStatus: http.StatusInternalServerError,
			setupMock: func(req *http.Request, svc *serviceMocks.IRoutingProcessorService) {
				svc.On("GetDanaEscrowBalance", mock.Anything, mock.AnythingOfType("string")).Return(&routingProcessorModelEscrowBalance.EscrowBalanceResponse{
					ResponseCode:    "2001600",
					ResponseMessage: "Successful",
					BalanceAmount:   1000000,
				}, assert.AnError)
			},
		},
		{
			name:           "error data not found",
			expectedStatus: http.StatusNotFound,
			setupMock: func(req *http.Request, svc *serviceMocks.IRoutingProcessorService) {
				svc.On("GetDanaEscrowBalance", mock.Anything, mock.AnythingOfType("string")).Return(nil, nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			disbursementSvc := serviceMocks.NewIDisbursementService(t)
			routingProcessorSvc := serviceMocks.NewIRoutingProcessorService(t)
			req, err := http.NewRequest("GET", "/crm/v1/balances/dana-balance", nil)
			if err != nil {
				t.Fatal(err)
			}

			tc.setupMock(req, routingProcessorSvc)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctrl := &handler{
					disbursementSvc:     disbursementSvc,
					routingProcessorSvc: routingProcessorSvc,
				}

				ctrl.GetDanaEscrowBalance(w, r)
			})

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}
