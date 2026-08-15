package merchant

import (
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMerchantRepositoryUpdate(t *testing.T) {
	now := time.Now()

	expectedMerchant := &merchantModel.Merchant{
		UUID:      "uuid-uuid-uuid",
		Name:      "test",
		Logo:      "test",
		CreatedAt: now,
	}

	// Define the test cases
	testCases := []struct {
		name      string
		inputUser *merchantModel.Merchant
		result    error
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: update merchant",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
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
					constant.StringMockType(),
				).Return(true, nil)
			},
			wantErr:   false,
			inputUser: expectedMerchant,
			result:    nil,
		},
		{
			name: "FAILED: Error when updating blocked status in user account",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
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
					constant.StringMockType(),
				).Return(false, errors.New("error database"))
			},
			wantErr:   true,
			inputUser: nil,
			result:    errors.New("error database"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			updatedMerchant := &merchantModel.Merchant{
				UUID:      "uuid-uuid-uuid",
				Name:      "test",
				Logo:      "test",
				CreatedAt: now,
			}

			repo := New(mockMysql, mockLogger)

			err := repo.Update(t.Context(), updatedMerchant)
			if (err != nil) != tc.wantErr {
				t.Errorf("UserRepository.UpdateBlocked() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestUpdateCallbackApiKey(t *testing.T) {
	merchantID := uuid.NewString()
	callbackApiKey := "api-key"

	// Define the test cases
	testCases := []struct {
		name      string
		result    error
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: update callback api key",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(true, nil)
			},
			wantErr: false,
			result:  nil,
		},
		{
			name: "FAILED: Error when updating callback api key of merchant",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(false, errors.New("error database"))
			},
			wantErr: true,
			result:  errors.New("error database"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)

			err := repo.UpdateCallbackApiKey(t.Context(), merchantID, callbackApiKey, uint(1))
			if (err != nil) != tc.wantErr {
				t.Errorf("UserRepository.UpdateBlocked() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestUpdateStatusByID(t *testing.T) {
	merchantID := uuid.NewString()

	// Define the test cases
	testCases := []struct {
		name      string
		result    error
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: update status",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(true, nil)
			},
			wantErr: false,
			result:  nil,
		},
		{
			name: "FAILED: Error when updating status of merchant",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
			result:  errors.New("some error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)

			err := repo.UpdateStatusByID(t.Context(), constant.MerchantStatusActive, "", merchantID)
			if (err != nil) != tc.wantErr {
				t.Errorf("UserRepository.UpdateStatusByID() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestUpdateThirdPartyScreeningData(t *testing.T) {
	merchantID := uuid.NewString()
	screeningData := types.NullJSONText{
		JSONText: []byte(`{"John Doe": {"code": "SUCCESS", "transactionId": "test123"}}`),
		Valid:    true,
	}

	// Define the test cases
	testCases := []struct {
		name      string
		result    error
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: update third party screening data",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("types.NullJSONText"),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(true, nil)
			},
			wantErr: false,
			result:  nil,
		},
		{
			name: "FAILED: Error when updating third party screening data of merchant",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("types.NullJSONText"),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
			result:  errors.New("some error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)

			err := repo.UpdateThirdPartyScreeningData(t.Context(), merchantID, screeningData)
			if (err != nil) != tc.wantErr {
				t.Errorf("MerchantRepository.UpdateThirdPartyScreeningData() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestMigrateMerchantSecretsToEncryption(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	tests := []struct {
		name      string
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On("NamedExecContext", mock.Anything, mock.Anything, mock.Anything).Once().Return(false, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On("NamedExecContext", mock.Anything, mock.Anything, mock.Anything).Once().Return(true, nil)
			},
			wantError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantError, repo.MigrateMerchantSecretsToEncryption(t.Context(), merchantModel.MigrateMerchantSecretsToEncryption{}))
		})
	}
}
