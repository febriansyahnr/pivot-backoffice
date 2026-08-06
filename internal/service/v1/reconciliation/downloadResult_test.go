package reconciliation

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/reconciliation"
	gcsMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	rabbitmqExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	xlsxMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/xlsx"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDownloadResult(t *testing.T) {
	ctx := context.Background()
	id := uuid.NewString()
	tests := []struct {
		name      string
		wantErr   bool
		setupMock func(*Mocker)
	}{
		{
			name: "success: get url download",
			setupMock: func(m *Mocker) {
				m.ReconRepo.On("GetByUUID", mock.Anything, id).Return(&reconciliation.Reconciliation{
					UUID: id,
				}, nil)
				m.Gcs.On("CreateSignedURL", mock.Anything, mock.Anything, mock.Anything).Return("url", nil)
			},
			wantErr: false,
		},
		{
			name: "failed: error when create signed url",
			setupMock: func(m *Mocker) {
				m.ReconRepo.On("GetByUUID", mock.Anything, id).Return(&reconciliation.Reconciliation{
					UUID: id,
				}, nil)
				m.Gcs.On("CreateSignedURL", mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "failed: error when get detail from db",
			setupMock: func(m *Mocker) {
				m.ReconRepo.On("GetByUUID", mock.Anything, id).Return(nil, errors.New("error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
			reconRepo := repoMocks.NewIReconciliationRepository(t)
			gcs := gcsMock.NewGCSService(t)
			xlsx := xlsxMock.NewExceler(t)
			rabbitMq := rabbitmqExtMock.NewRabbitMQExt(t)

			mock := &Mocker{
				ReconRepo: reconRepo,
				Gcs:       gcs,
				Xlsx:      xlsx,
				RabbitMq:  rabbitMq,
			}

			tc.setupMock(mock)

			service := New(
				&config.Config{},
				logger,
				reconRepo,
				WithExcelService(xlsx),
				WithGCSService(gcs),
				WithRabbitMQClient(rabbitMq),
			)

			_, err := service.DownloadResult(ctx, id)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
