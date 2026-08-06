package customerRepository

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDelete(t *testing.T) {
	uuid := "foobar"
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Delete customer",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failure Delete customer",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(false, fmt.Errorf("Delete error"))

			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			err := repo.Delete(ctx, uuid, "")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestRemovePaymentMethodFromCustomerByIDAndTokenID(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	tokenId := "382f8821-e06f-4f64-97b1-b2eba32ee406"

	tests := []struct {
		name      string
		tokenId   string
		setupMock func()
		wantError error
	}{
		{
			name:      "ERROR:Token not found", // NOSONAR
			tokenId:   "",
			setupMock: func() { /* Empty Function */ },
			wantError: constant.ErrDataNotFound,
		},
		{
			name:    "ERROR:Some error", // NOSONAR
			tokenId: tokenId,
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(false, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name:    "SUCCESS", // NOSONAR
			tokenId: tokenId,
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(true, nil)
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			paymentMethods := []*unifiedPaymentModel.CustomerPaymentMethodResponse{}
			rawPaymentMethods := []byte(`[{"token":"382f8821-e06f-4f64-97b1-b2eba32ee406","status":"ACTIVE"},{"token":"2856b483-e33d-4afa-aa19-496580002634","status":"ACTIVE"}]`)

			assert.NoError(t, json.Unmarshal(rawPaymentMethods, &paymentMethods))
			assert.Equal(
				t, test.wantError, repo.RemovePaymentMethodFromCustomerByIDAndTokenID(t.Context(), uuid.NewString(), test.tokenId, paymentMethods),
			)
		})
	}
}
