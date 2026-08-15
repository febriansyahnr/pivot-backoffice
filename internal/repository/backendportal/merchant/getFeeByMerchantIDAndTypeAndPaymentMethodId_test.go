package merchant_test

import (
	"context"
	"database/sql"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/fee"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal/merchant"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeterminePaymentFeeByMerchantIdMethodAndChannel(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	merchantId := "e4df0972-60ea-4ef0-8afb-b79a06379d72"
	paymentMethod := c.ChannelVirtualAccount
	paymentChannel := "PERMATA" // NOSOANAR
	settlementModel := c.PaymentMethodChannelTypeAggregator

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *merchant.MerchantFee
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything,
					merchantId, c.ReferencePayment, paymentMethod, paymentChannel, settlementModel,
					merchantId, c.ReferencePayment, paymentMethod, settlementModel,
					merchantId, c.ReferencePayment, paymentMethod, paymentChannel,
					merchantId, c.ReferencePayment, paymentMethod,
				).Once().Return(assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "SUCCESS:Data not found",
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything,
					merchantId, c.ReferencePayment, paymentMethod, paymentChannel, settlementModel,
					merchantId, c.ReferencePayment, paymentMethod, settlementModel,
					merchantId, c.ReferencePayment, paymentMethod, paymentChannel,
					merchantId, c.ReferencePayment, paymentMethod,
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS:Data found",
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything,
					merchantId, c.ReferencePayment, paymentMethod, paymentChannel, settlementModel,
					merchantId, c.ReferencePayment, paymentMethod, settlementModel,
					merchantId, c.ReferencePayment, paymentMethod, paymentChannel,
					merchantId, c.ReferencePayment, paymentMethod,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*merchant.MerchantFee) = merchant.MerchantFee{}
				}).Return(nil)
			},
			wantResult: &merchant.MerchantFee{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.DeterminePaymentFeeByMerchantIdMethodAndChannel(context.Background(), &feeModel.GetFeeRequest{
				MerchantID:      merchantId,
				PaymentMethod:   paymentMethod,
				Channel:         paymentChannel,
				SettlementModel: settlementModel,
			})
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestDeterminePayoutFeeByMerchantIdAndChannel(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	merchantId := "e4df0972-60ea-4ef0-8afb-b79a06379d72" // NOSONAR
	payoutChannel := "PERMATA"                           // NOSOANAR
	reference := c.ReferenceDisbursement                 // NOSONAR

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *merchant.MerchantFee
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantId, reference, payoutChannel, merchantId, reference,
				).Once().Return(assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "SUCCESS:Data not found",
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantId, reference, payoutChannel, merchantId, reference,
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS:Data found",
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantId, reference, payoutChannel, merchantId, reference,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*merchant.MerchantFee) = merchant.MerchantFee{}
				}).Return(nil)
			},
			wantResult: &merchant.MerchantFee{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.DeterminePayoutFeeByMerchantIdAndChannel(context.Background(), merchantId, payoutChannel, reference)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestDetermineRefundFeeByMerchantIdAndReferenceType(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	merchantId := "e4df0972-60ea-4ef0-8afb-b79a06379d72"
	referenceType := "CHANNEL" // NOSOANAR

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *merchant.MerchantFee
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantId, c.ReferenceRefund, referenceType,
				).Once().Return(assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "SUCCESS:Data not found",
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantId, c.ReferenceRefund, referenceType,
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS:Data found",
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantId, c.ReferenceRefund, referenceType,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*merchant.MerchantFee) = merchant.MerchantFee{}
				}).Return(nil)
			},
			wantResult: &merchant.MerchantFee{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.DetermineRefundFeeByMerchantIdAndReferenceType(context.Background(), merchantId, referenceType)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestDetermineTopupFeeByMerchantIdMethodAndChannel(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	merchantId := "e4df0972-60ea-4ef0-8afb-b79a06379d72"
	paymentMethod := c.ChannelVirtualAccount
	paymentChannel := "PERMATA" // NOSOANAR

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *merchant.MerchantFee
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything,
					merchantId, c.ReferenceTopUp, paymentMethod, paymentChannel,
					merchantId, c.ReferenceTopUp, paymentMethod,
				).Once().Return(assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "SUCCESS:Data not found",
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything,
					merchantId, c.ReferenceTopUp, paymentMethod, paymentChannel,
					merchantId, c.ReferenceTopUp, paymentMethod,
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS:Data found",
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything,
					merchantId, c.ReferenceTopUp, paymentMethod, paymentChannel,
					merchantId, c.ReferenceTopUp, paymentMethod,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*merchant.MerchantFee) = merchant.MerchantFee{}
				}).Return(nil)
			},
			wantResult: &merchant.MerchantFee{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.DetermineTopupFeeByMerchantIdMethodAndChannel(context.Background(), merchantId, paymentMethod, paymentChannel)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestDeterminePaymentFundedPayoutFeeByMerchantIdMethodAndChannel(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	merchantId := "e4df0972-60ea-4ef0-8afb-b79a06379d72"
	paymentMethod := c.ChannelCreditCard
	paymentChannel := "VISA"
	settlementMethod := c.SettlementMethodInstant

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *merchant.MerchantFee
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything,
					merchantId, c.ReferencePaymentFundedPayout, paymentMethod, settlementMethod, paymentChannel,
					merchantId, c.ReferencePaymentFundedPayout, paymentMethod, settlementMethod,
				).Once().Return(assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "SUCCESS:Data not found",
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything,
					merchantId, c.ReferencePaymentFundedPayout, paymentMethod, settlementMethod, paymentChannel,
					merchantId, c.ReferencePaymentFundedPayout, paymentMethod, settlementMethod,
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS:Data found with channel",
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything,
					merchantId, c.ReferencePaymentFundedPayout, paymentMethod, settlementMethod, paymentChannel,
					merchantId, c.ReferencePaymentFundedPayout, paymentMethod, settlementMethod,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*merchant.MerchantFee) = merchant.MerchantFee{
						Amount:        1500,
						AmountType:    c.MerchantFeeAmountType,
						Percentage:    1.5,
						TaxType:       c.MerchantTaxTypeNonPKP,
						DeductionType: c.MerchantFeeDeductionTypeDirect,
					}
				}).Return(nil)
			},
			wantResult: &merchant.MerchantFee{
				Amount:        1500,
				AmountType:    c.MerchantFeeAmountType,
				Percentage:    1.5,
				TaxType:       c.MerchantTaxTypeNonPKP,
				DeductionType: c.MerchantFeeDeductionTypeDirect,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.DeterminePaymentFundedPayoutFeeByMerchantIdMethodAndChannel(context.Background(), merchantId, paymentMethod, paymentChannel, settlementMethod)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
