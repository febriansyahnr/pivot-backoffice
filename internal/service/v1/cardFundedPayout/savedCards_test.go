package cardFundedPayoutService_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/cardFundedPayout"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateSavedCard(t *testing.T) {
	cfg := &config.Config{
		UnifiedPaymentConfig: config.UnifiedPaymentConfig{
			CardConfig: &config.UnifiedPaymentCardConfig{
				MaxExpiryDuration:     1,
				MaxExpiryDurationUnit: "hour",
			},
		},
		MerchantPortalConfig: config.MerchantPortalConfig{
			CardFundedPayoutURL: "https://merchant-portal.test/card-funded-payout",
		},
	}

	customerSvc := serviceMock.NewICustomerService(t)
	unifiedPaymentSvc := serviceMock.NewIUnifiedPaymentService(t)
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	svc := New(cfg, log,
		WithCustomerService(customerSvc),
		WithUnifiedPaymentService(unifiedPaymentSvc),
	)

	validRequest := &cardFundedPayoutModel.CreateSavedCardRequest{
		ReferenceID: "ref-123456",
		MerchantID:  "merchant-123",
		CreatedBy:   "user-123",
	}

	testCases := []struct {
		name      string
		request   *cardFundedPayoutModel.CreateSavedCardRequest
		setupMock func()
		wantErr   bool
	}{
		{
			name:    "ERROR: Create unified payment customer failed",
			request: validRequest,
			setupMock: func() {
				customerSvc.On("CreateUnfiedPaymentCustomer", mock.Anything, mock.Anything).
					Return(nil, errors.New("customer service error")).Once()
			},
			wantErr: true,
		},
		{
			name:    "ERROR: Create session failed",
			request: validRequest,
			setupMock: func() {
				customerSvc.On("CreateUnfiedPaymentCustomer", mock.Anything, mock.Anything).
					Return(&customerModel.GeneralCustomerResponse{
						UUID: "customer-uuid-123",
					}, nil).Once()

				unifiedPaymentSvc.On("CreateSession", mock.Anything, mock.Anything).
					Return(nil, errors.New("unified payment service error")).Once()
			},
			wantErr: true,
		},
		{
			name:    "SUCCESS: Create saved card",
			request: validRequest,
			setupMock: func() {
				customerSvc.On("CreateUnfiedPaymentCustomer", mock.Anything, mock.Anything).
					Return(&customerModel.GeneralCustomerResponse{
						UUID: "customer-uuid-123",
					}, nil).Once()

				unifiedPaymentSvc.On("CreateSession", mock.Anything, mock.Anything).
					Return(&unifiedPaymentModel.UnifiedPaymentSessionResponse{
						ID:              "session-id-123",
						ShortPaymentUrl: "https://payment.url/session",
					}, nil).Once()
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			result, err := svc.CreateSavedCard(context.Background(), tc.request)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.request.ReferenceID, result.ReferenceID)
				assert.NotEmpty(t, result.PaymentUrl)
			}

			customerSvc.AssertExpectations(t)
			unifiedPaymentSvc.AssertExpectations(t)
		})
	}
}

func TestGetSavedCardList(t *testing.T) {
	cfg := &config.Config{}
	customerSvc := serviceMock.NewICustomerService(t)
	unifiedPaymentSvc := serviceMock.NewIUnifiedPaymentService(t)
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	svc := New(cfg, log,
		WithCustomerService(customerSvc),
		WithUnifiedPaymentService(unifiedPaymentSvc),
	)

	now := time.Now()
	validFilter := &cardFundedPayoutModel.FilterGetSavedCardList{
		MerchantID: "merchant-123",
		Page:       1,
		PerPage:    10,
		Sort:       "DESC",
		SortBy:     "createdAt",
	}

	testCases := []struct {
		name      string
		filter    *cardFundedPayoutModel.FilterGetSavedCardList
		setupMock func()
		wantErr   bool
	}{
		{
			name:   "ERROR: Get card funded payout saved card list failed",
			filter: validFilter,
			setupMock: func() {
				customerSvc.On("GetCardFundedPayoutSavedCardList", mock.Anything, mock.Anything).
					Return(nil, errors.New("database error")).Once()
			},
			wantErr: true,
		},
		{
			name:   "SUCCESS: Get saved card list with empty result",
			filter: validFilter,
			setupMock: func() {
				customerSvc.On("GetCardFundedPayoutSavedCardList", mock.Anything, mock.Anything).
					Return(&commonModel.PaginationResponse{
						Data: []cardFundedPayoutModel.GetSavedCardResponse{},
						Meta: commonModel.Meta{
							Page:       1,
							PerPage:    10,
							TotalItems: 0,
							TotalPages: 0,
						},
					}, nil).Once()
			},
			wantErr: false,
		},
		{
			name:   "SUCCESS: Get saved card list with data",
			filter: validFilter,
			setupMock: func() {
				customerSvc.On("GetCardFundedPayoutSavedCardList", mock.Anything, mock.Anything).
					Return(&commonModel.PaginationResponse{
						Data: []cardFundedPayoutModel.GetSavedCardResponse{
							{
								ID:             "customer-123",
								CardName:       "VISA",
								PaymentChannel: "CHANNEL",
								IssuingBank:    "BANK",
								Last4:          "1234",
								ExpiryMonth:    "12",
								ExpiryYear:     "2025",
							},
						},
						Meta: commonModel.Meta{
							Page:       1,
							PerPage:    10,
							TotalItems: 1,
							TotalPages: 1,
						},
					}, nil).Once()
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get saved card list with date filter",
			filter: &cardFundedPayoutModel.FilterGetSavedCardList{
				MerchantID:     "merchant-123",
				Page:           1,
				PerPage:        1000,
				Sort:           "ASC",
				SortBy:         "updatedAt",
				StartCreatedAt: ptrTime(now.Add(-24 * time.Hour)),
				EndCreatedAt:   ptrTime(now),
			},
			setupMock: func() {
				customerSvc.On("GetCardFundedPayoutSavedCardList", mock.Anything, mock.Anything).
					Return(&commonModel.PaginationResponse{
						Data: []cardFundedPayoutModel.GetSavedCardResponse{},
						Meta: commonModel.Meta{
							Page:       1,
							PerPage:    1000,
							TotalItems: 0,
							TotalPages: 0,
						},
					}, nil).Once()
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			result, err := svc.GetSavedCardList(context.Background(), tc.filter)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotNil(t, result.Data)
				assert.NotNil(t, result.Meta)
			}

			customerSvc.AssertExpectations(t)
		})
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
