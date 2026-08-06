package reconciliation

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/reconciliation"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
)

func TestReconciliationRepositoryGetByUUID(t *testing.T) {
	testCases := []struct {
		name      string
		uuid      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		want      *reconciliation.Reconciliation
		wantErr   error
	}{
		{
			name: "SUCCESS: Get reconciliation by UUID",
			uuid: uuid.NewString(),
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				expectedRecon := &reconciliation.Reconciliation{
					UUID:           uuid.NewString(),
					FilePath:       "test/path.csv",
					ResultFilePath: "test/result.csv",
					Status:         "COMPLETED",
					CreatedBy:      "test",
					Reasons: sql.NullString{
						String: "success",
						Valid:  true,
					},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*reconciliation.Reconciliation"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Run(func(args mock.Arguments) {
					arg := args.Get(1).(*reconciliation.Reconciliation)
					*arg = *expectedRecon
				}).Return(nil)
			},
			want:    &reconciliation.Reconciliation{},
			wantErr: nil,
		},
		{
			name: "ERROR: Record not found",
			uuid: uuid.NewString(),
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*reconciliation.Reconciliation"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)
			},
			want:    nil,
			wantErr: constant.ErrDataNotFound,
		},
		{
			name: "ERROR: Database error",
			uuid: uuid.NewString(),
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*reconciliation.Reconciliation"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))

			},
			want:    &reconciliation.Reconciliation{},
			wantErr: errors.New("database error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)

			got, err := repo.GetByUUID(context.Background(), tc.uuid)
			if tc.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, got)
		})
	}
}
func TestReconciliationRepositoryGetAll(t *testing.T) {
	testCases := []struct {
		name      string
		input     *reconciliation.ReconciliationFilterRequest
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  *commonModel.PaginationResponse
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get all reconciliation records with pagination",
			input: &reconciliation.ReconciliationFilterRequest{
				Page:    1,
				PerPage: 10,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*reconciliation.Reconciliation"),
					mock.AnythingOfType("string"),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					totalItems := args.Get(1).(*int64)
					*totalItems = 15
				})
			},
			expected: &commonModel.PaginationResponse{
				Data: make([]*reconciliation.Reconciliation, 0),
				Meta: commonModel.Meta{
					Page:       1,
					PerPage:    10,
					TotalItems: 15,
					TotalPages: 2,
				},
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Empty result with zero total items",
			input: &reconciliation.ReconciliationFilterRequest{
				Page:    1,
				PerPage: 10,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*reconciliation.Reconciliation"),
					mock.AnythingOfType("string"),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					totalItems := args.Get(1).(*int64)
					*totalItems = 0
				})
			},
			expected: &commonModel.PaginationResponse{
				Data: make([]*reconciliation.Reconciliation, 0),
				Meta: commonModel.Meta{
					Page:       1,
					PerPage:    10,
					TotalItems: 0,
					TotalPages: 0,
				},
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database error in SelectContext",
			input: &reconciliation.ReconciliationFilterRequest{
				Page:    1,
				PerPage: 10,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*reconciliation.Reconciliation"),
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "SUCCESS: Handle page less than 1",
			input: &reconciliation.ReconciliationFilterRequest{
				Page:    0,
				PerPage: 10,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*reconciliation.Reconciliation"),
					mock.AnythingOfType("string"),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					totalItems := args.Get(1).(*int64)
					*totalItems = 5
				})
			},
			expected: &commonModel.PaginationResponse{
				Data: make([]*reconciliation.Reconciliation, 0),
				Meta: commonModel.Meta{
					Page:       1,
					PerPage:    10,
					TotalItems: 5,
					TotalPages: 1,
				},
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Handle GetContext error - totalItems set to 0",
			input: &reconciliation.ReconciliationFilterRequest{
				Page:    1,
				PerPage: 10,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*reconciliation.Reconciliation"),
					mock.AnythingOfType("string"),
				).Return(nil)

				// GetContext returns an error to trigger the error handling path
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("string"),
				).Return(errors.New("count query failed"))
			},
			expected: &commonModel.PaginationResponse{
				Data: make([]*reconciliation.Reconciliation, 0),
				Meta: commonModel.Meta{
					Page:       1,
					PerPage:    10,
					TotalItems: 0,
					TotalPages: 0,
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			result, err := repo.GetAll(context.Background(), tc.input)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected.Meta, result.Meta)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}
