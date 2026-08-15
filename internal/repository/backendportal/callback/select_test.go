package callbackRepository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/callback"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCallbackFindCallbackMasterByMerchantID(t *testing.T) {
	name := "Virtual Account"
	now := time.Now()
	callbackMasterResult := callbackModel.CallbackMaster{
		UUID:        uuid.New(),
		Name:        name,
		Description: "API",
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   sql.NullTime{},
	}

	testCases := []struct {
		name      string
		input     string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  *callbackModel.CallbackMaster
		wantErr   bool
	}{
		{
			name: "SUCCESS: Find Callback Master by Merchant ID",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*callback_model.CallbackMaster"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*callbackModel.CallbackMaster) = callbackMasterResult
				})
			},
			input:    name,
			expected: &callbackMasterResult,
			wantErr:  false,
		},
		{
			name: "ERROR: Callback Master Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*callback_model.CallbackMaster"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)

			},
			input:    name,
			expected: nil,
			wantErr:  false,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*callback_model.CallbackMaster"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))

			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tt.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "callback_masters")
			transaction, err := repo.FindCallbackMasterByName(ctx, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, transaction)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestCallbackFindCallbackByNameAndMerchantID(t *testing.T) {
	merchantID := uuid.New()
	name := "Virtual Account"
	now := time.Now()
	callbackResult := callbackModel.Callback{
		UUID:        merchantID,
		URL:         "https://localhost/v1/payment-callback",
		Description: "API",
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   sql.NullTime{},
	}

	type request struct {
		Name       string
		MerchantID uuid.UUID
	}

	input := request{
		Name:       name,
		MerchantID: merchantID,
	}

	testCases := []struct {
		name      string
		input     request
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  *callbackModel.Callback
		wantErr   bool
	}{
		{
			name: "SUCCESS: Find Callback by Merchant ID And Name",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*callback_model.Callback"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					constant.UuidMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*callbackModel.Callback) = callbackResult
				})
			},
			input:    input,
			expected: &callbackResult,
			wantErr:  false,
		},
		{
			name: "ERROR: Callback Master Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*callback_model.Callback"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					constant.UuidMockType(),
				).Return(sql.ErrNoRows)

			},
			input:    input,
			expected: nil,
			wantErr:  false,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*callback_model.Callback"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					constant.UuidMockType(),
				).Return(errors.New("database error"))

			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tt.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "callbacks")
			transaction, err := repo.FindCallbackByNameAndMerchantID(ctx, tt.input.Name, tt.input.MerchantID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, transaction)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestGetCallbackURLByMerchantId(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)
	ptrSliceOfCallbackURLSetReqMockType := mock.AnythingOfType("*[]callback_model.CallbackURLSettingResp")

	repo := New(db, nil)
	tests := []struct {
		name       string
		masterName string
		setupMocks func()
		wantErr    string
	}{
		{
			name: "ERROR:Some internal GetCallbackURLByMerchantId error",
			setupMocks: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(),
					ptrSliceOfCallbackURLSetReqMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "SUCCESS",
			setupMocks: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(),
					ptrSliceOfCallbackURLSetReqMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(nil)
			},
		},
		{
			name:       "SUCCESS for SNAP",
			masterName: "SNAP",
			setupMocks: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(),
					ptrSliceOfCallbackURLSetReqMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMocks()
			if _, err := repo.GetCallbackURLByMerchantId(context.Background(), uuid.NewString(), test.masterName); test.wantErr == "" {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestGetCallbackAPIKeyByMerchantId(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)
	tests := []struct {
		name       string
		setupMocks func()
		wantErr    string
	}{
		{
			name: "ERROR:Some internal GetCallbackAPIKeyByMerchantId error",
			setupMocks: func() {
				db.On(
					"GetContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "SUCCESS",
			setupMocks: func() {
				db.On(
					"GetContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), constant.StringMockType(),
				).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMocks()

			if _, err := repo.GetCallbackAPIKeyByMerchantId(context.Background(), uuid.NewString()); test.wantErr == "" {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestGetCallbackIdByMerchantAndMasterCallbackId(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)
	tests := []struct {
		name       string
		setupMocks func()
		wantErr    string
	}{
		{
			name: "ERROR:Some internal GetCallbackIdByMerchantAndMasterCallbackId error",
			setupMocks: func() {
				db.On(
					"GetContext", constant.ValueCtxMockType(),
					constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "SUCCESS",
			setupMocks: func() {
				db.On(
					"GetContext", constant.ValueCtxMockType(),
					constant.PtrStringMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMocks()

			if _, err := repo.GetCallbackIdByMerchantAndMasterCallbackId(context.Background(), uuid.NewString(), uuid.NewString()); test.wantErr == "" {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
