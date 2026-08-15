package merchant_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal/merchant"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	mysqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetBillingFees(t *testing.T) {

	log := loggerMock.NewILogger(t)
	db := mysqlExtMock.NewIMySqlExt(t)

	repo := New(db, log)

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *merchant.BillingFeeResponse
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(assert.AnError)
				log.On(
					"Error", mock.Anything, "Failed while recap merchant billing fees", mock.Anything,
				).Once().Return()
			},
			wantErr: assert.AnError,
		},
		{
			name: "SUCCESS:Data not found", // NOSONAR
			setupMock: func() {
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(sql.ErrNoRows)
			},
			wantResult: &merchant.BillingFeeResponse{Details: map[string][]merchant.BillingFeeDetailResponse{}},
		},
		{
			name: "SUCCESS:Data found", // NOSONAR
			setupMock: func() {
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*[]merchant.BillingFeeDetailResponse)) = []merchant.BillingFeeDetailResponse{
						{
							Type:           constant.ReferencePayment,
							Method:         constant.ChannelVirtualAccount,
							Channel:        "PERMATA",
							Total:          1,
							TrxAmount:      10_000,
							FeeType:        constant.MerchantFeeAmountType,
							FeeAmount:      1_000,
							TotalFeeAmount: 1_000,
						},
						{
							Type:           constant.ReferencePayment,
							Method:         constant.ChannelQris,
							Channel:        "BNC",
							Total:          4,
							TrxAmount:      400_000,
							FeeType:        constant.MerchantFeePercentageType,
							FeePercentage:  0.5,
							TotalFeeAmount: 2_000,
						},
						{
							Type:           constant.ReferencePayment,
							Method:         constant.ChannelCreditCard,
							Channel:        "LOCAL_VISA",
							Total:          2,
							TrxAmount:      1_000_000,
							FeeType:        constant.MerchantFeePercentageType,
							FeePercentage:  0.7,
							TotalFeeAmount: 14_000,
						},
						{
							Type:           constant.ReferenceDisbursement,
							Channel:        "BCA",
							Total:          2,
							TrxAmount:      1_000_000,
							FeeType:        constant.MerchantFeeAmountType,
							FeeAmount:      4_000,
							TotalFeeAmount: 8_000,
						},
						{
							Type:           constant.ReferenceAccountInquiry,
							Total:          2,
							FeeType:        constant.MerchantFeeAmountType,
							FeeAmount:      1_000,
							TotalFeeAmount: 2_000,
						},
						{
							Type:           constant.ReferencePlatformActivity,
							Total:          1,
							FeeType:        constant.MerchantFeeAmountType,
							FeeAmount:      10_000,
							TotalFeeAmount: 10_000,
						},
						{
							Type:           constant.ReferencePlatformTransaction,
							Total:          10,
							TrxAmount:      1_000_000,
							FeeType:        constant.MerchantFeePercentageType,
							FeePercentage:  0.1,
							TotalFeeAmount: 1_000,
						},
						{
							Type:           "",
							Total:          1,
							FeeType:        constant.MerchantFeeAmountType,
							TotalFeeAmount: 1_000,
						},
					}
				}).Return(nil)
			},
			wantResult: &merchant.BillingFeeResponse{
				Total:          23,
				TotalFeeAmount: 39_000,
				Details: map[string][]merchant.BillingFeeDetailResponse{
					"payments": {
						{
							Type:           constant.ReferencePayment,
							Method:         constant.ChannelVirtualAccount,
							Channel:        "PERMATA",
							Total:          1,
							TrxAmount:      10_000,
							FeeType:        constant.MerchantFeeAmountType,
							FeeAmount:      1_000,
							TotalFeeAmount: 1_000,
						},
						{
							Type:           constant.ReferencePayment,
							Method:         constant.ChannelQris,
							Channel:        "BNC",
							Total:          4,
							TrxAmount:      400_000,
							FeeType:        constant.MerchantFeePercentageType,
							FeePercentage:  0.5,
							TotalFeeAmount: 2_000,
						},
						{
							Type:           constant.ReferencePayment,
							Method:         constant.ChannelCreditCard,
							Channel:        "LOCAL_VISA",
							Total:          2,
							TrxAmount:      1_000_000,
							FeeType:        constant.MerchantFeePercentageType,
							FeePercentage:  0.7,
							TotalFeeAmount: 14_000,
						},
					},
					"payouts": {
						{
							Type:           constant.ReferenceDisbursement,
							Channel:        "BCA",
							Total:          2,
							TrxAmount:      1_000_000,
							FeeType:        constant.MerchantFeeAmountType,
							FeeAmount:      4_000,
							TotalFeeAmount: 8_000,
						},
					},
					"accountInquiry": {
						{
							Type:           constant.ReferenceAccountInquiry,
							Total:          2,
							FeeType:        constant.MerchantFeeAmountType,
							FeeAmount:      1_000,
							TotalFeeAmount: 2_000,
						},
					},
					"platformActivity": {
						{
							Type:           constant.ReferencePlatformActivity,
							Total:          1,
							FeeType:        constant.MerchantFeeAmountType,
							FeeAmount:      10_000,
							TotalFeeAmount: 10_000,
						},
					},
					"platformTransaction": {
						{
							Type:           constant.ReferencePlatformTransaction,
							Total:          10,
							TrxAmount:      1_000_000,
							FeeType:        constant.MerchantFeePercentageType,
							FeePercentage:  0.1,
							TotalFeeAmount: 1_000,
						},
					},
					"others": {
						{
							Type:           "",
							Total:          1,
							FeeType:        constant.MerchantFeeAmountType,
							TotalFeeAmount: 1_000,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetBillingFees(context.Background(), merchant.BillingFeeRequest{BillingDateRangeRequest: &merchant.BillingDateRangeRequest{}})
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestPayBillingFees(t *testing.T) {
	log := loggerMock.NewILogger(t)
	db := mysqlExtMock.NewIMySqlExt(t)

	repo := New(db, log)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(false, assert.AnError)
				log.On(
					"Error", mock.Anything, "Failed while execute query for pay billing fees", logger.Error(assert.AnError),
				).Once().Return()
			},
			wantErr: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()

			assert.Equal(t, test.wantErr, repo.PayBillingFees(context.Background(), merchant.PayBillingFeeRequest{BillingDateRangeRequest: &merchant.BillingDateRangeRequest{}}))
		})
	}
}
