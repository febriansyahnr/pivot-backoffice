package reconciliation

import (
	"context"
	"database/sql"
	"errors"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/reconciliation"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcessPayoutRecon(t *testing.T) {
	ctx := context.Background()
	mockUUID := uuid.New()

	testCases := []struct {
		desc          string
		input         *reconciliation.ReconciliationPayout
		mock          func(m *Mocker)
		expectedError error
	}{
		{
			desc: "SUCCESS: ProcessPayoutRecon with amount",
			input: &reconciliation.ReconciliationPayout{
				UUID:                   "test-uuid",
				ExternalID:             "ext-123",
				PartnerReferenceNo:     "ref-123",
				Acquirer:               "test-acquirer",
				Status:                 "TRUE",
				Amount:                 &commonModel.Amount{Value: "10000"},
				ProcessorReferenceName: "test-processor",
			},
			mock: func(m *Mocker) {
				// Mock FindByID
				m.AccountTransaction.On("FindByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(
					&orchestrator_model.AccountTransactionWithUseCase{
						UUID: mockUUID,
						Type: constant.TypeDisbursement,
					},
					nil,
				)

				// Mock SetAdditionalInfoReconciliation
				m.AccountTransaction.On("SetAdditionalInfoReconciliation",
					mock.Anything,
					mockUUID.String(),
					mock.Anything,
				).Return(nil)
			},
			expectedError: nil,
		},
		{
			desc: "SUCCESS: ProcessPayoutRecon without amount",
			input: &reconciliation.ReconciliationPayout{
				UUID:                   "test-uuid",
				ExternalID:             "ext-456",
				PartnerReferenceNo:     "ref-456",
				Acquirer:               "test-acquirer",
				Status:                 "REVIEW",
				Amount:                 nil,
				ProcessorReferenceName: "test-processor",
			},
			mock: func(m *Mocker) {
				// Mock FindByID
				m.AccountTransaction.On("FindByID",
					mock.Anything,
					"ext-456",
				).Return(
					&orchestrator_model.AccountTransactionWithUseCase{
						UUID: mockUUID,
						Type: constant.TypeDisbursement,
					},
					nil,
				)

				// Mock SetAdditionalInfoReconciliation
				m.AccountTransaction.On("SetAdditionalInfoReconciliation",
					mock.Anything,
					mockUUID.String(),
					mock.MatchedBy(func(detail *reconciliation.ReconDetail) bool {
						return detail.Status == "REVIEW" && detail.Amount == "0"
					}),
				).Return(nil)
			},
			expectedError: nil,
		},
		{
			desc: "ERROR: Transaction not found (ErrNoRows)",
			input: &reconciliation.ReconciliationPayout{
				UUID:                   "test-uuid",
				ExternalID:             "ext-789",
				PartnerReferenceNo:     "ref-789",
				Acquirer:               "test-acquirer",
				Status:                 "TRUE",
				Amount:                 &commonModel.Amount{Value: "10000"},
				ProcessorReferenceName: "test-processor",
			},
			mock: func(m *Mocker) {
				// Mock FindByID returning ErrNoRows
				m.AccountTransaction.On("FindByID",
					mock.Anything,
					"ext-789",
				).Return(nil, sql.ErrNoRows)
			},
			expectedError: nil, // Function returns nil for ErrNoRows
		},
		{
			desc: "ERROR: Transaction is nil",
			input: &reconciliation.ReconciliationPayout{
				UUID:                   "test-uuid",
				ExternalID:             "ext-999",
				PartnerReferenceNo:     "ref-999",
				Acquirer:               "test-acquirer",
				Status:                 "TRUE",
				Amount:                 &commonModel.Amount{Value: "10000"},
				ProcessorReferenceName: "test-processor",
			},
			mock: func(m *Mocker) {
				// Mock FindByID returning nil without error
				m.AccountTransaction.On("FindByID",
					mock.Anything,
					"ext-999",
				).Return(nil, nil)
			},
			expectedError: nil, // Function returns nil for nil transaction
		},
		{
			desc: "ERROR: Database error on FindByID",
			input: &reconciliation.ReconciliationPayout{
				UUID:                   "test-uuid",
				ExternalID:             "ext-error",
				PartnerReferenceNo:     "ref-error",
				Acquirer:               "test-acquirer",
				Status:                 "TRUE",
				Amount:                 &commonModel.Amount{Value: "10000"},
				ProcessorReferenceName: "test-processor",
			},
			mock: func(m *Mocker) {
				// Mock FindByID returning database error
				dbErr := errors.New("database error")
				m.AccountTransaction.On("FindByID",
					mock.Anything,
					"ext-error",
				).Return(nil, dbErr)
			},
			expectedError: errors.New("database error"),
		},
		{
			desc: "ERROR: SetAdditionalInfoReconciliation error",
			input: &reconciliation.ReconciliationPayout{
				UUID:                   "test-uuid",
				ExternalID:             "ext-update-error",
				PartnerReferenceNo:     "ref-update-error",
				Acquirer:               "test-acquirer",
				Status:                 "TRUE",
				Amount:                 &commonModel.Amount{Value: "10000"},
				ProcessorReferenceName: "test-processor",
			},
			mock: func(m *Mocker) {
				// Mock FindByID
				m.AccountTransaction.On("FindByID",
					mock.Anything,
					"ext-update-error",
				).Return(&orchestrator_model.AccountTransactionWithUseCase{
					UUID: mockUUID,
				}, nil)

				// Mock SetAdditionalInfoReconciliation with error
				updateErr := errors.New("update error")
				m.AccountTransaction.On("SetAdditionalInfoReconciliation",
					mock.Anything,
					mockUUID.String(),
					mock.Anything,
				).Return(updateErr)
			},
			expectedError: errors.New("update error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Setup mocks
			m := &Mocker{
				ReconRepo:          repoMocks.NewIReconciliationRepository(t),
				AccountTransaction: repoMocks.NewIAccountTransactionRepository(t),
			}
			tc.mock(m)

			logger, _ := logger.NewZapLogger(logger.Config{})

			// Create service
			cfg := &config.Config{}
			svc := New(cfg, logger, m.ReconRepo,
				WithAccountTransactionRepository(m.AccountTransaction))

			// Call the function
			err := svc.ProcessPayoutRecon(ctx, tc.input)

			// Assert results
			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			m.AccountTransaction.AssertExpectations(t)
		})
	}
}
