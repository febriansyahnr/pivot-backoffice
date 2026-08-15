package accounttransaction_repository

import (
	"context"
	"testing"
	"time"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/reconciliation"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateBulkReconStatus(t *testing.T) {
	testCases := []struct {
		name      string
		input     *reconciliation.BulkUpatedStatus
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update bulk recon status within 3 days",
			input: &reconciliation.BulkUpatedStatus{
				StartTime:    time.Now().Add(-48 * time.Hour),
				EndTime:      time.Now(),
				Status:       constant.ReconStatusSuccess,
				TrxReference: constant.ReferencePayment,
				TrxType:      constant.TypePayment,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Update bulk recon status with duration > 3 days",
			input: &reconciliation.BulkUpatedStatus{
				StartTime:    time.Now().Add(-96 * time.Hour),
				EndTime:      time.Now(),
				Status:       constant.ReconStatusReview,
				TrxReference: constant.ReferencePayment,
				TrxType:      constant.TypePayment,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: No rows affected",
			input: &reconciliation.BulkUpatedStatus{
				StartTime:    time.Now().Add(-24 * time.Hour),
				EndTime:      time.Now(),
				Status:       constant.ReconStatusReview,
				TrxReference: constant.ReferencePayment,
				TrxType:      constant.TypePayment,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(false, constant.ErrNoRowsAffected)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database error",
			input: &reconciliation.BulkUpatedStatus{
				StartTime:    time.Now().Add(-24 * time.Hour),
				EndTime:      time.Now(),
				Status:       constant.ReconStatusReview,
				TrxReference: constant.ReferencePayment,
				TrxType:      constant.TypePayment,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: Update with empty status",
			input: &reconciliation.BulkUpatedStatus{
				StartTime:    time.Now().Add(-24 * time.Hour),
				EndTime:      time.Now(),
				Status:       "",
				TrxReference: constant.ReferencePayment,
				TrxType:      constant.TypePayment,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(true, nil)
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
			err := repo.UpdateBulkReconStatus(context.Background(), tc.input)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}

func TestUpdateReconDetail(t *testing.T) {
	testCases := []struct {
		name      string
		id        string
		input     *reconciliation.ReconDetail
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update recon detail",
			id:   "transaction-id-1",
			input: &reconciliation.ReconDetail{
				Status: constant.ReconStatusSuccess,
				Reason: "Transaction matched with processor",
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					constant.ReconStatusSuccess,
					"Transaction matched with processor",
					"transaction-id-1",
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: No rows affected",
			id:   "transaction-id-2",
			input: &reconciliation.ReconDetail{
				Status: constant.ReconStatusReview,
				Reason: "Needs further verification",
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					constant.ReconStatusReview,
					"Needs further verification",
					"transaction-id-2",
				).Return(false, constant.ErrNoRowsAffected)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database error",
			id:   "transaction-id-3",
			input: &reconciliation.ReconDetail{
				Status: constant.ReconStatusReview,
				Reason: "Missing in processor records",
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					constant.ReconStatusReview,
					"Missing in processor records",
					"transaction-id-3",
				).Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: Update recon detail with DateTime",
			id:   "transaction-id-4",
			input: &reconciliation.ReconDetail{
				Status:   constant.ReconStatusSuccess,
				Reason:   "Transaction matched",
				DateTime: "2025-01-15 10:30:00",
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					constant.ReconStatusSuccess,
					"Transaction matched",
					"transaction-id-4",
				).Return(true, nil)
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
			err := repo.UpdateReconDetail(context.Background(), tc.id, tc.input)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}

func TestSetAdditionalInfoReconciliation(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		input     *reconciliation.ReconDetail
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Set additional info reconciliation",
			id:   "transaction-id-1",
			input: &reconciliation.ReconDetail{
				Status: constant.ReconStatusSuccess,
				Reason: "Transaction matched with processor",
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: No rows affected",
			id:   "transaction-id-2",
			input: &reconciliation.ReconDetail{
				Status: constant.ReconStatusReview,
				Reason: "Needs further verification",
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(false, constant.ErrNoRowsAffected)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database error",
			id:   "transaction-id-3",
			input: &reconciliation.ReconDetail{
				Status: constant.ReconStatusReview,
				Reason: "Missing in processor records",
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(false, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			err := repo.SetAdditionalInfoReconciliation(context.Background(), tc.id, tc.input)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}
