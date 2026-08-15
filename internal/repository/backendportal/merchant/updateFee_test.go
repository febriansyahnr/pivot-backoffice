package merchant

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMerchantRepository_UpdateMerchantFee(t *testing.T) {
	now := time.Now()

	dataUpdate := &merchantModel.MerchantFee{
		UUID:       "uuid-uuid-uuid",
		MerchantID: "merchant-id",
		Amount:     1000,
		Reference:  "DISBURSEMENT",
		UpdatedAt:  now,
		CreatedAt:  now,
	}

	testCases := []struct {
		desc      string
		input     *merchantModel.MerchantFee
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			desc:  "SUCCESS: update merchant_fees",
			input: dataUpdate,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					constant.PtrMerchantFeeMockType(),
				).Return(true, nil)
			},
		},
		{
			desc:  "ERROR: update merchant_fees",
			input: dataUpdate,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					constant.PtrMerchantFeeMockType(),
				).Return(false, errors.New("update error"))
			},
			wantErr: true,
		},
		{
			desc: "ERROR: No Rows Affected",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					constant.PtrMerchantFeeMockType(),
				).Return(false, nil)
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			err := repo.UpdateMerchantFee(context.Background(), tc.input)

			if (err != nil) != tc.wantErr {
				t.Errorf("MerchantRepository.CreateMerchantFee() got errpr: %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestUpdateMerchantFeeLastDeductionDate(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

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
					"ExecContext", constant.ValueCtxMockType(), constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest, // NOSONAR
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"ExecContext", constant.ValueCtxMockType(), constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantErr, repo.UpdateMerchantFeeLastDeductionDate(context.Background(), uuid.NewString(), "", time.Now().UTC()))
		})
	}
}

func TestUpdateFeeTieringConfig(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	tests := []struct {
		name      string
		request   *merchantModel.FeeTieringRequest
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Exec context",
			setupMock: func() {
				db.On(
					"ExecContext", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
					constant.StringMockType(), constant.JSONTextMockType(), constant.TimeMockType(), constant.StringMockType(),
				).Once().Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:No row affected",
			setupMock: func() {
				db.On(
					"ExecContext", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
					constant.StringMockType(), constant.JSONTextMockType(), constant.TimeMockType(), constant.StringMockType(),
				).Once().Return(false, nil)
			},
			wantErr: constant.ErrDataNotFound,
		},
		{
			name: "SUCCESS:No applied fee",
			setupMock: func() {
				db.On(
					"ExecContext", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
					constant.StringMockType(), constant.JSONTextMockType(), constant.TimeMockType(), constant.StringMockType(),
				).Return(true, nil)
			},
		},
		{
			name: "ERROR:Named execute context",
			request: &merchantModel.FeeTieringRequest{
				AppliedFee: &merchantModel.FeeTieringConfig{},
			},
			setupMock: func() {
				db.On(
					"NamedExecContext", constant.ValueCtxMockType(), constant.StringMockType(), mock.Anything,
				).Once().Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS",
			request: &merchantModel.FeeTieringRequest{
				AppliedFee: &merchantModel.FeeTieringConfig{},
			},
			setupMock: func() {
				db.On(
					"NamedExecContext", constant.ValueCtxMockType(), constant.StringMockType(), mock.Anything,
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()
			if test.request == nil {
				test.request = &merchantModel.FeeTieringRequest{}
			}
			assert.Equal(t, test.wantErr, repo.UpdateFeeTieringConfig(context.Background(), test.request))
		})
	}
}

func TestAppliedFeeFromTiers(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Named execute context",
			setupMock: func() {
				db.On(
					"NamedExecContext", constant.ValueCtxMockType(), constant.StringMockType(), mock.Anything,
				).Once().Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"NamedExecContext", constant.ValueCtxMockType(), constant.StringMockType(), mock.Anything,
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()
			assert.Equal(t, test.wantErr, repo.AppliedFeeFromTiers(context.Background(), "123", &merchantModel.FeeTieringConfig{}))
		})
	}
}
