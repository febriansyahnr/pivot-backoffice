package reconciliation

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/reconciliation"
	gcsMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	rabbitmqExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	xlsxMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/xlsx"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateReconDetail(t *testing.T) {
	ctx := context.Background()
	id := "test-id"

	tests := []struct {
		name      string
		payload   *reconciliation.ReconDetail
		wantErr   bool
		setupMock func(*Mocker)
	}{
		{
			name: "success: update recon detail",
			payload: &reconciliation.ReconDetail{
				Status: constant.ReconStatusSuccess,
				Reason: "approved",
			},
			setupMock: func(m *Mocker) {
				m.AccountTransaction.On("FindByID", mock.Anything, id).Return(&orchestratorModel.AccountTransactionWithUseCase{UUID: uuid.New()}, nil)
				m.AccountTransaction.On("UpdateReconDetail", mock.Anything, id, mock.AnythingOfType("*reconciliation.ReconDetail")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "failed: account transaction not found",
			payload: &reconciliation.ReconDetail{
				Status: constant.ReconStatusSuccess,
				Reason: "approved",
			},
			setupMock: func(m *Mocker) {
				m.AccountTransaction.On("FindByID", mock.Anything, id).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "failed: error finding account transaction",
			payload: &reconciliation.ReconDetail{
				Status: constant.ReconStatusSuccess,
				Reason: "approved",
			},
			setupMock: func(m *Mocker) {
				m.AccountTransaction.On("FindByID", mock.Anything, id).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "failed: error updating recon detail",
			payload: &reconciliation.ReconDetail{
				Status: constant.ReconStatusSuccess,
				Reason: "approved",
			},
			setupMock: func(m *Mocker) {
				m.AccountTransaction.On("FindByID", mock.Anything, id).Return(&orchestratorModel.AccountTransactionWithUseCase{UUID: uuid.New()}, nil)
				m.AccountTransaction.On("UpdateReconDetail", mock.Anything, id, mock.AnythingOfType("*reconciliation.ReconDetail")).Return(errors.New("update error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
			accountTransactionRepo := repoMocks.NewIAccountTransactionRepository(t)
			gcs := gcsMock.NewGCSService(t)
			xlsx := xlsxMock.NewExceler(t)
			rabbitMq := rabbitmqExtMock.NewRabbitMQExt(t)

			mock := &Mocker{
				AccountTransaction: accountTransactionRepo,
				Gcs:                gcs,
				Xlsx:               xlsx,
				RabbitMq:           rabbitMq,
			}

			tc.setupMock(mock)

			service := New(
				&config.Config{},
				logger,
				nil,
				WithAccountTransactionRepository(accountTransactionRepo),
				WithExcelService(xlsx),
				WithGCSService(gcs),
				WithRabbitMQClient(rabbitMq),
			)

			err := service.UpdateReconDetail(ctx, id, tc.payload)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
