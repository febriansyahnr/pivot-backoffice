package merchant_test

import (
	"context"
	"database/sql"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/merchant"
	mySqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidateCreateFeeConfigOnBehalf(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil)

	ptrIntMockType := mock.AnythingOfType("*int")

	tests := []struct {
		name       string
		request    *merchant.CreateFeeConfigOnBehalfRequest
		setupMock  func()
		wantErr    error
		wantResult bool
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			request: &merchant.CreateFeeConfigOnBehalfRequest{
				Type:      c.FeeOnBehalfTypeDefault,
				Reference: c.ReferencePayment,
			},
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrIntMockType, c.StringMockType(),
					c.StringMockType(), c.StringMockType(), c.StringMockType(), mock.Anything, c.StringMockType(), c.StringMockType(), mock.Anything,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS:Direct", // NOSONAR
			request: &merchant.CreateFeeConfigOnBehalfRequest{
				MerchantId:    "123456",
				Type:          c.FeeOnBehalfTypeDirect,
				SubMerchantId: util.ValueToPtr("654321"),
				Reference:     c.ReferencePayment,
			},
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrIntMockType, c.StringMockType(),
					c.StringMockType(), mock.Anything, c.StringMockType(), c.StringMockType(), mock.Anything,
					c.StringMockType(), mock.Anything, c.StringMockType(), c.StringMockType(), mock.Anything,
				).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*int) = 4
				}).Return(nil)
			},
			wantResult: true,
		},
		{
			name: "SUCCESS:Default",
			request: &merchant.CreateFeeConfigOnBehalfRequest{
				Type:      c.FeeOnBehalfTypeDefault,
				Reference: c.ReferencePayment,
			},
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrIntMockType, c.StringMockType(),
					c.StringMockType(), c.StringMockType(), c.StringMockType(), mock.Anything, c.StringMockType(), c.StringMockType(), mock.Anything,
				).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*int) = 3
				}).Return(nil)
			},
			wantResult: true,
		},
		{
			name: "SUCCESS:Default with ReferenceType", // New test case to cover lines 52-55
			request: &merchant.CreateFeeConfigOnBehalfRequest{
				MerchantId:    "12345",
				Type:          c.FeeOnBehalfTypeDefault,
				Reference:     c.ReferencePayment,
				ReferenceType: "SOME_REF_TYPE",
				PaymentMethod: util.ValueToPtr("QRIS"),
			},
			setupMock: func() {
				// Use mock.Anything for all string parameters to be more flexible
				db.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*int) = 3
				}).Return(nil)
			},
			wantResult: true,
		},
		{
			name: "SUCCESS:All",
			request: &merchant.CreateFeeConfigOnBehalfRequest{
				Type:      c.FeeOnBehalfTypeAll,
				Reference: c.ReferencePayment,
			},
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), ptrIntMockType, c.StringMockType(),
					c.StringMockType(), c.StringMockType(), c.StringMockType(), mock.Anything, c.StringMockType(), c.StringMockType(), mock.Anything,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*int) = 3
				}).Return(nil)
			},
			wantResult: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.ValidateCreateFeeConfigOnBehalf(context.Background(), test.request)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestCreateFeeConfigOnBehalf(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil)

	dataMockType := mock.AnythingOfType("*merchant.OnBehalfFeeConfig")

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), dataMockType,
				).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"NamedExecContext", c.ValueCtxMockType(), c.StringMockType(), dataMockType,
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantErr, repo.CreateFeeConfigOnBehalf(context.Background(), &merchant.OnBehalfFeeConfig{}))
		})
	}
}

func TestGetFeeConfigOnBehalf(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil)

	result := []merchant.FeeConfigOnBehalfResponse{
		{
			Id:         "5050588d-b7f2-4e2e-a842-9a5f28b246a2",
			Type:       c.FeeOnBehalfTypeDefault,
			AmountType: c.MerchantFeeAmountType,
			Amount:     2_000,
		},
	}
	resultMockType := mock.AnythingOfType("*[]merchant.FeeConfigOnBehalfResponse")

	tests := []struct {
		name          string
		setupMock     func()
		wantErr       error
		wantResult    []merchant.FeeConfigOnBehalfResponse
		referenceType string
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(), mock.Anything,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS:Data not found",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(), mock.Anything,
				).Once().Return(sql.ErrNoRows)
			},
			wantResult: []merchant.FeeConfigOnBehalfResponse{},
		},
		{
			name: "SUCCESS:Data found",
			setupMock: func() {
				db.On(
					"SelectContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(), mock.Anything,
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*[]merchant.FeeConfigOnBehalfResponse)) = result
				}).Return(nil)
			},
			wantResult: result,
		},
		{
			name:          "SUCCESS:Data found with reference type", // New test case to cover lines 111-114
			referenceType: "QRIS_TYPE",
			setupMock: func() {
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*[]merchant.FeeConfigOnBehalfResponse)) = result
				}).Return(nil)
			},
			wantResult: result,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			request := &merchant.GetFeeConfigOnBehalfRequest{
				Reference:     c.ReferencePayment,
				PaymentMethod: util.ValueToPtr(c.ChannelVirtualAccount),
				ReferenceType: test.referenceType,
			}
			result, err := repo.GetFeeConfigOnBehalf(context.Background(), request)

			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestUpdateFeeConfigOnBehalf(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"ExecContext", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.Float64MockType(), c.Float64MockType(), c.TimeMockType(), c.StringMockType(),
				).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Data not found",
			setupMock: func() {
				db.On(
					"ExecContext", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.Float64MockType(), c.Float64MockType(), c.TimeMockType(), c.StringMockType(),
				).Once().Return(false, nil)
			},
			wantErr: c.ErrNoRowsAffected,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"ExecContext", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.Float64MockType(), c.Float64MockType(), c.TimeMockType(), c.StringMockType(),
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantErr, repo.UpdateFeeConfigOnBehalf(context.Background(), "12345", &merchant.UpdateFeeConfigOnBehalfRequest{}))
		})
	}
}

func TestGetTransactionFeeOnBehalf(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil)
	resultMockType := mock.AnythingOfType("*merchant.TransactionFeeOnBehalf")

	args := []interface{}{
		c.ValueCtxMockType(), resultMockType, c.StringMockType(),
	}
	for range 7 {
		args = append(args, c.StringMockType())
	}

	tests := []struct {
		name          string
		setupMock     func()
		wantErr       error
		wantResult    *merchant.TransactionFeeOnBehalf
		referenceType string
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On("GetContext", args...).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS:Data not found", // NOSONAR
			setupMock: func() {
				db.On("GetContext", args...).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS:Data found", // NOSONAR
			setupMock: func() {
				db.On("GetContext", args...).Return(nil)
			},
			wantResult: &merchant.TransactionFeeOnBehalf{},
		},
		{
			name:          "SUCCESS:With reference type", // New test case to cover lines 165-168
			referenceType: "SOME_TYPE",
			setupMock: func() {
				// We need more arguments for this case because reference_type is included in the query
				argsWithReferenceType := []interface{}{
					c.ValueCtxMockType(), resultMockType, c.StringMockType(),
				}
				for range 8 { // One more argument for the reference type
					argsWithReferenceType = append(argsWithReferenceType, c.StringMockType())
				}
				db.On("GetContext", argsWithReferenceType...).Return(nil)
			},
			wantResult: &merchant.TransactionFeeOnBehalf{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetTransactionFeeOnBehalf(context.Background(), "123", "321", c.ReferencePayment, "QRIS", test.referenceType)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
