package merchant_test

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/merchant"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetBillingFees(t *testing.T) {
	repo := repoMocks.NewIMerchantRepository(t)
	service := New(repo, nil, nil, nil, nil, nil)

	merchantId := "main-merchant-123"
	subMerchantId := "sub-merchant-456"

	mainMerchant := &merchant.Merchant{
		UUID: merchantId,
		Name: "Main Merchant",
	}

	subMerchant := &merchant.Merchant{
		UUID: subMerchantId,
		Name: "Sub Merchant",
	}

	allBillingFees := &merchant.BillingFeeResponse{
		Total:          3,
		TotalFeeAmount: 3000,
		Details: map[string][]merchant.BillingFeeDetailResponse{
			"payments": {
				{
					Type:           "PAYMENT",
					Method:         "CREDIT_CARD",
					Channel:        "VISA",
					MerchantId:     merchantId,
					Total:          1,
					TrxAmount:      1000000,
					FeeType:        "PERCENTAGE",
					FeePercentage:  0.8,
					TotalFeeAmount: 800,
				},
				{
					Type:           "PAYMENT",
					Method:         "QRIS",
					Channel:        "QRIS_DANA",
					MerchantId:     subMerchantId,
					Total:          1,
					TrxAmount:      500000,
					FeeType:        "PERCENTAGE",
					FeePercentage:  0.7,
					TotalFeeAmount: 350,
				},
			},
			"payouts": {
				{
					Type:           "DISBURSEMENT",
					Channel:        "BCA",
					MerchantId:     subMerchantId,
					Total:          1,
					TrxAmount:      200000,
					FeeType:        "AMOUNT",
					FeeAmount:      1250,
					TotalFeeAmount: 1250,
				},
			},
		},
	}

	expectedSubMerchants := []merchant.SubMerchantBillingResponse{
		{
			SubMerchantId:   subMerchantId,
			SubMerchantName: "Sub Merchant",
			Total:           2,
			TotalFeeAmount:  1600,
			Details: map[string][]merchant.BillingFeeDetailResponse{
				"payments": {
					{
						Type:           "PAYMENT",
						Method:         "QRIS",
						Channel:        "QRIS_DANA",
						MerchantId:     subMerchantId,
						Total:          1,
						TrxAmount:      500000,
						FeeType:        "PERCENTAGE",
						FeePercentage:  0.7,
						TotalFeeAmount: 350,
					},
				},
				"payouts": {
					{
						Type:           "DISBURSEMENT",
						Channel:        "BCA",
						MerchantId:     subMerchantId,
						Total:          1,
						TrxAmount:      200000,
						FeeType:        "AMOUNT",
						FeeAmount:      1250,
						TotalFeeAmount: 1250,
					},
				},
			},
		},
	}

	tests := []struct {
		name       string
		request    merchant.BillingFeeRequest
		setupMock  func()
		wantErr    error
		wantResult *merchant.BillingFeeResponse
	}{
		{
			name: "ERROR:Merchant not found",
			request: merchant.BillingFeeRequest{
				MerchantId: merchantId,
			},
			setupMock: func() {
				repo.On("FindMerchantByID", mock.Anything, merchantId).Once().Return(nil, assert.AnError)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, assert.AnError),
		},
		{
			name: "ERROR:Merchant does not exist",
			request: merchant.BillingFeeRequest{
				MerchantId: merchantId,
			},
			setupMock: func() {
				repo.On("FindMerchantByID", mock.Anything, merchantId).Once().Return(nil, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrNotFound, nil),
		},
		{
			name: "ERROR:GetBillingFees repository error",
			request: merchant.BillingFeeRequest{
				MerchantId: merchantId,
			},
			setupMock: func() {
				repo.On("FindMerchantByID", mock.Anything, merchantId).Once().Return(mainMerchant, nil)
				repo.On("GetBillingFees", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "SUCCESS:Only main merchant fees",
			request: merchant.BillingFeeRequest{
				MerchantId: merchantId,
			},
			setupMock: func() {
				mainOnlyFees := &merchant.BillingFeeResponse{
					Total:          1,
					TotalFeeAmount: 800,
					Details: map[string][]merchant.BillingFeeDetailResponse{
						"payments": {
							{
								Type:           "PAYMENT",
								MerchantId:     merchantId,
								Total:          1,
								TotalFeeAmount: 800,
							},
						},
					},
				}
				repo.On("FindMerchantByID", mock.Anything, merchantId).Once().Return(mainMerchant, nil)
				repo.On("GetBillingFees", mock.Anything, mock.Anything).Once().Return(mainOnlyFees, nil)
			},
			wantResult: &merchant.BillingFeeResponse{
				MerchantId:     merchantId,
				MerchantName:   "Main Merchant",
				Total:          1,
				TotalFeeAmount: 800,
				Details: map[string][]merchant.BillingFeeDetailResponse{
					"payments": {
						{
							Type:           "PAYMENT",
							MerchantId:     merchantId,
							Total:          1,
							TotalFeeAmount: 800,
						},
					},
				},
				SubMerchants: nil,
			},
		},
		{
			name: "SUCCESS:Main merchant and sub-merchant fees",
			request: merchant.BillingFeeRequest{
				MerchantId: merchantId,
			},
			setupMock: func() {
				repo.On("FindMerchantByID", mock.Anything, merchantId).Once().Return(mainMerchant, nil)
				repo.On("GetBillingFees", mock.Anything, mock.Anything).Once().Return(allBillingFees, nil)
				repo.On("FindMerchantByID", mock.Anything, subMerchantId).Once().Return(subMerchant, nil)
			},
			wantResult: &merchant.BillingFeeResponse{
				MerchantId:     merchantId,
				MerchantName:   "Main Merchant",
				Total:          3,
				TotalFeeAmount: 2400,
				Details: map[string][]merchant.BillingFeeDetailResponse{
					"payments": {
						{
							Type:           "PAYMENT",
							Method:         "CREDIT_CARD",
							Channel:        "VISA",
							MerchantId:     merchantId,
							Total:          1,
							TrxAmount:      1000000,
							FeeType:        "PERCENTAGE",
							FeePercentage:  0.8,
							TotalFeeAmount: 800,
						},
					},
				},
				SubMerchants: expectedSubMerchants,
			},
		},
		{
			name: "SUCCESS:Sub-merchant merchant lookup fails but continues",
			request: merchant.BillingFeeRequest{
				MerchantId: merchantId,
			},
			setupMock: func() {
				repo.On("FindMerchantByID", mock.Anything, merchantId).Once().Return(mainMerchant, nil)
				repo.On("GetBillingFees", mock.Anything, mock.Anything).Once().Return(allBillingFees, nil)
				repo.On("FindMerchantByID", mock.Anything, subMerchantId).Once().Return(nil, assert.AnError)
			},
			wantResult: &merchant.BillingFeeResponse{
				MerchantId:     merchantId,
				MerchantName:   "Main Merchant",
				Total:          1,
				TotalFeeAmount: 800,
				Details: map[string][]merchant.BillingFeeDetailResponse{
					"payments": {
						{
							Type:           "PAYMENT",
							Method:         "CREDIT_CARD",
							Channel:        "VISA",
							MerchantId:     merchantId,
							Total:          1,
							TrxAmount:      1000000,
							FeeType:        "PERCENTAGE",
							FeePercentage:  0.8,
							TotalFeeAmount: 800,
						},
					},
				},
				SubMerchants: []merchant.SubMerchantBillingResponse{},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.GetBillingFees(context.Background(), test.request)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestPayBillingFees(t *testing.T) {
	repo := repoMocks.NewIMerchantRepository(t)

	service := New(repo, nil, nil, nil, nil, nil)

	billingFees := &merchant.BillingFeeResponse{
		Total:          1,
		TotalFeeAmount: 1_000,
		Details: map[string][]merchant.BillingFeeDetailResponse{
			"accountInquiry": {
				{
					Type:           "ACCOUNT_INQUIRY",
					Total:          1,
					FeeType:        "AMOUNT",
					FeeAmount:      1_000,
					TotalFeeAmount: 1_000,
				},
			},
		},
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *merchant.BillingFeeResponse
	}{
		{
			name: "ERROR:Get billing fees",
			setupMock: func() {
				repo.On("GetBillingFees", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, assert.AnError),
		},
		{
			name: "SUCCESS:Billing fees not found",
			setupMock: func() {
				repo.On("GetBillingFees", mock.Anything, mock.Anything).Once().Return(&merchant.BillingFeeResponse{}, nil)
			},
			wantResult: &merchant.BillingFeeResponse{},
		},
		{
			name: "ERROR:Pay billing fees",
			setupMock: func() {
				repo.On("GetBillingFees", mock.Anything, mock.Anything).Return(billingFees, nil)
				repo.On("PayBillingFees", mock.Anything, mock.Anything).Once().Return(assert.AnError)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, assert.AnError),
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				repo.On("PayBillingFees", mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantResult: billingFees,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()

			result, err := service.PayBillingFees(context.Background(), merchant.PayBillingFeeRequest{})
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
