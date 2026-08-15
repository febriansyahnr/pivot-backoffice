package merchant

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	pdkLogMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMerchantRepositoryFindMerchantByID(t *testing.T) {
	merchant := &merchantModel.Merchant{
		UUID:      "uuid-uuid-uuid",
		Name:      "test",
		Logo:      "test",
		CreatedAt: time.Now(),
	}

	testCases := []struct {
		name      string
		id        string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  *merchantModel.Merchant
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get Merchant By ID",
			id:   "uuid-uuid-uuid",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*merchant.Merchant"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					merchantPtr := args.Get(1).(*merchantModel.Merchant)
					*merchantPtr = *merchant
				})
			},
			expected: merchant,
			wantErr:  false,
		},
		{
			name: "ERROR: Merchant Not Found",
			id:   "uuid-uuid-uuid",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*merchant.Merchant"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)
			},
			expected: nil,
			wantErr:  false,
		},
		{
			name: "ERROR: Database Error",
			id:   "uuid-uuid-uuid",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*merchant.Merchant"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "merchants")
			merchantRes, err := repo.FindMerchantByID(ctx, tc.id)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, merchantRes)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}

func TestFindMerchantForQrRegistration(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	ptrSliceBODMockType := mock.AnythingOfType("*[]merchant.QrisBOD")
	ptrQrisMerchantMockType := mock.AnythingOfType("*merchant.QrisMerchant")
	ptrSliceQrisDocumentMockType := mock.AnythingOfType("*[]merchant.QrisDocument")

	qrisMerchant := &merchantModel.QrisMerchant{
		Documents: []merchantModel.QrisDocument{
			{
				LocationRaw: []byte("{}"),
			},
		},
		BoardOfDirectors: []merchantModel.QrisBOD{
			{
				Position:        constant.PositionCommissioner,
				IdentityFileRaw: []byte("{}"),
			},
			{
				Position:        constant.PositionDirector,
				IdentityFileRaw: []byte("{}"),
			},
		},
		AddressRaw: []byte("{}"),
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    string
		wantResult *merchantModel.QrisMerchant
	}{
		{
			name: "ERROR:Fetch merchant and registration",
			setupMock: func() {
				db.On(
					"GetContext", constant.ValueCtxMockType(), ptrQrisMerchantMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(fmt.Errorf(constant.InternalErrorFmt, "001"))
			},
			wantErr: fmt.Sprintf(constant.InternalErrorFmt, "001"),
		},
		{
			name: "ERROR:Merchant not found",
			setupMock: func() {
				db.On(
					"GetContext", constant.ValueCtxMockType(), ptrQrisMerchantMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "ERROR:Unmarshal address",
			setupMock: func() {
				db.On(
					"GetContext", constant.ValueCtxMockType(), ptrQrisMerchantMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*merchantModel.QrisMerchant) = merchantModel.QrisMerchant{
						AddressRaw: []byte("A"),
					}
				}).Return(nil)
			},
			wantErr: "unmarshal: invalid character 'A' looking for beginning of value",
		},
		{
			name: "ERROR:Fetch merchant documents",
			setupMock: func() {
				db.On(
					"GetContext", constant.ValueCtxMockType(), ptrQrisMerchantMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*merchantModel.QrisMerchant) = *qrisMerchant
				}).Return(nil)

				db.On(
					"SelectContext", constant.ValueCtxMockType(), ptrSliceQrisDocumentMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: "merchant document: some error",
		},
		{
			name: "ERROR:Fetch data board of directors",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), ptrSliceQrisDocumentMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*[]merchantModel.QrisDocument) = []merchantModel.QrisDocument{
						{LocationRaw: []byte("{}")},
					}
				}).Return(nil)

				db.On(
					"SelectContext", constant.ValueCtxMockType(), ptrSliceBODMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: "merchant bod: some error",
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), ptrSliceBODMockType, constant.StringMockType(), constant.StringMockType(), constant.StringMockType(),
				).Run(func(args mock.Arguments) {
					qrisMerchant.BODCount = 1
					qrisMerchant.BOCCount = 1
					*args.Get(1).(*[]merchantModel.QrisBOD) = []merchantModel.QrisBOD{
						{Position: constant.PositionCommissioner, IdentityFileRaw: []byte("{}")},
						{Position: constant.PositionDirector, IdentityFileRaw: []byte("{}")},
					}
				}).Return(nil)
			},
			wantResult: qrisMerchant,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			resp, err := repo.FindMerchantForQrRegistration(context.Background(), uuid.NewString(), "BANK")
			if test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, resp)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestGetListMerchantFeeThatUseTiers(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	resultMockType := mock.AnythingOfType("*[]merchant.MerchantFeeThatUseTier")

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult map[string][]merchantModel.MerchantFeeThatUseTier
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), resultMockType, constant.StringMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Data not found",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), resultMockType, constant.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), resultMockType, constant.StringMockType(),
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*[]merchantModel.MerchantFeeThatUseTier)) = []merchantModel.MerchantFeeThatUseTier{
						{MerchantId: "123", Reference: constant.ReferenceDisbursement, RawTieringConfigs: []byte(`{}`)}, {MerchantId: "124", Reference: constant.ReferencePlatformActivity, RawTieringConfigs: []byte(`{}`)},
					}
				}).Return(nil)
			},
			wantResult: map[string][]merchantModel.MerchantFeeThatUseTier{"123": {{MerchantId: "123", Reference: constant.ReferenceDisbursement, RawTieringConfigs: []byte(`{}`)}}, "124_" + constant.ReferencePlatformActivity: {{MerchantId: "124", Reference: constant.ReferencePlatformActivity, RawTieringConfigs: []byte(`{}`)}}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetListMerchantFeeThatUseTiers(context.Background())
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestGetSubMerchantIdListByParentId(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)
	result := []string{"1", "2", "3"}

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult []string
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Data not found",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), constant.StringMockType(),
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*[]string)) = result
				}).Return(nil)
			},
			wantResult: result,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetSubMerchantIdListByParentId(context.Background(), "123456")
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestGetAllActiveMerchantIDs(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)
	result := []string{uuid.NewString(), uuid.NewString(), uuid.NewString()}

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult []string
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Data not found",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), constant.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), constant.StringMockType(),
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*[]string)) = result
				}).Return(nil)
			},
			wantResult: result,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetAllActiveMerchantIDs(context.Background())
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestListUnencryptedMerchantSecretsForMigration(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)
	logger := pdkLogMock.NewILogger(t)

	repo := New(db, logger)

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult []merchantModel.UnencryptedMerchantSecretsForMigration
	}{
		{
			name: "SUCCESS:Data not found", // NOSONAR
			setupMock: func() {
				db.On("SelectContext", mock.Anything, mock.Anything, mock.Anything).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On("SelectContext", mock.Anything, mock.Anything, mock.Anything).Once().Return(assert.AnError)
				logger.On("Error", mock.Anything, "failed while retrieving unencrypted merchant secret list", mock.Anything).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On("SelectContext", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantResult: []merchantModel.UnencryptedMerchantSecretsForMigration{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.ListUnencryptedMerchantSecretsForMigration(t.Context())
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
