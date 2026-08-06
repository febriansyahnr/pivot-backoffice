package vccsettlement

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/vccSettlement"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestProcessSettlementTransactionInquiry(t *testing.T) {
	validRequestJSON := []byte(`{
		"rcnId": "rcn-123",
		"merchantId": "merchant-456",
		"recordType": "A",
		"billingCycle": "202502",
		"postingDate": "20250201",
		"partnerReferenceNo": "partner-ref-789"
	}`)

	invalidJSON := []byte(`{"invalid json`)

	testCases := []struct {
		name      string
		body      []byte
		routeKey  string
		mockSetup func(service *serviceMocks.IVccSettlementService)
		wantErr   bool
		assertFn  func(t *testing.T, err error, service *serviceMocks.IVccSettlementService)
	}{
		{
			name:     "Success - valid request",
			body:     validRequestJSON,
			routeKey: "",
			mockSetup: func(service *serviceMocks.IVccSettlementService) {
				service.On("ProcessRcnTransactionInquiry", mock.Anything, mock.MatchedBy(func(req *vccSettlement.VccTransactionInquiryRequest) bool {
					return req.RcnId == "rcn-123" &&
						req.MerchantId == "merchant-456" &&
						req.RecordType == "A" &&
						req.BillingCycle == "202502" &&
						req.PostingDate == "20250201" &&
						req.PartnerReferenceNo == "partner-ref-789"
				})).Return(nil).Once()
			},
			wantErr: false,
			assertFn: func(t *testing.T, err error, service *serviceMocks.IVccSettlementService) {
				service.AssertExpectations(t)
			},
		},
		{
			name:     "Failure - invalid JSON",
			body:     invalidJSON,
			routeKey: "",
			mockSetup: func(service *serviceMocks.IVccSettlementService) {
				// Service should not be called
			},
			wantErr: true,
			assertFn: func(t *testing.T, err error, service *serviceMocks.IVccSettlementService) {
				assert.Error(t, err)
				// Service should not be called
				service.AssertNotCalled(t, "ProcessRcnTransactionInquiry", mock.Anything, mock.Anything)
			},
		},
		{
			name:     "Failure - service returns lock acquisition error",
			body:     validRequestJSON,
			routeKey: "",
			mockSetup: func(service *serviceMocks.IVccSettlementService) {
				service.On("ProcessRcnTransactionInquiry", mock.Anything, mock.AnythingOfType("*vccSettlement.VccTransactionInquiryRequest")).
					Return(constant.ErrAcquireTransactionInquiryLock).Once()
			},
			wantErr: true,
			assertFn: func(t *testing.T, err error, service *serviceMocks.IVccSettlementService) {
				assert.Equal(t, constant.ErrAcquireTransactionInquiryLock, err)
				service.AssertExpectations(t)
			},
		},
		{
			name:     "Failure - service returns processing error",
			body:     validRequestJSON,
			routeKey: "",
			mockSetup: func(service *serviceMocks.IVccSettlementService) {
				service.On("ProcessRcnTransactionInquiry", mock.Anything, mock.AnythingOfType("*vccSettlement.VccTransactionInquiryRequest")).
					Return(constant.ErrProcessSettlementTransactionInquiry).Once()
			},
			wantErr: true,
			assertFn: func(t *testing.T, err error, service *serviceMocks.IVccSettlementService) {
				assert.Equal(t, constant.ErrProcessSettlementTransactionInquiry, err)
				service.AssertExpectations(t)
			},
		},
		{
			name:     "Failure - service returns generic error",
			body:     validRequestJSON,
			routeKey: "",
			mockSetup: func(service *serviceMocks.IVccSettlementService) {
				service.On("ProcessRcnTransactionInquiry", mock.Anything, mock.AnythingOfType("*vccSettlement.VccTransactionInquiryRequest")).
					Return(errors.New("unexpected service error")).Once()
			},
			wantErr: true,
			assertFn: func(t *testing.T, err error, service *serviceMocks.IVccSettlementService) {
				assert.Contains(t, err.Error(), "unexpected service error")
				service.AssertExpectations(t)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockService := serviceMocks.NewIVccSettlementService(t)

			tc.mockSetup(mockService)

			// Create handler
			h := New(mockLogger, mockService)

			// Execute
			err := h.ProcessSettlementTransactionInquiry(context.Background(), tc.body, tc.routeKey)

			// Assert
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tc.assertFn != nil {
				tc.assertFn(t, err, mockService)
			}
		})
	}
}
