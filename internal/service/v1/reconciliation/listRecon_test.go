package reconciliation

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/reconciliation"
	gcsMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	rabbitmqExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	xlsxMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/xlsx"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type Mocker struct {
	ReconRepo          *repoMocks.IReconciliationRepository
	Gcs                *gcsMock.GCSService
	Xlsx               *xlsxMock.Exceler
	RabbitMq           *rabbitmqExtMock.RabbitMQExt
	SnapCore           *repoMocks.ISnapCoreRepository
	AccountTransaction *repoMocks.IAccountTransactionRepository
}

func TestListRecon(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{}

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func(*Mocker)
	}{
		{
			name:    "success get all recon",
			wantErr: false,
			setupMock: func(m *Mocker) {
				m.ReconRepo.On("GetAll", mock.Anything, &reconciliation.ReconciliationFilterRequest{}).Return(&commonModel.PaginationResponse{
					Data: []map[string]string{},
				}, nil)
			},
		},
		{
			name:    "success get all recon with data transformation",
			wantErr: false,
			setupMock: func(m *Mocker) {
				mockRecons := []*reconciliation.Reconciliation{
					{
						UUID:           "01979b00-a2ba-7217-976b-f9651af470c1",
						OriginalName:   "RECON_VA.xlsx",
						FilePath:       "reconciliations/uploads/uploaded_reconciliation_20250623041639504.xlsx",
						ResultFilePath: "reconciliations/results/result_reconciliation_20250623041648367.xlsx",
						Status:         "SUCCESS",
						Reasons:        sql.NullString{String: "there is multiple errors, check at result file for more details", Valid: true},
						CreatedBy:      "febriansyah@harsya.com",
						CreatedAt:      time.Date(2025, 6, 23, 4, 16, 40, 0, time.UTC),
						UpdatedAt:      time.Date(2025, 6, 23, 4, 16, 48, 0, time.UTC),
					},
				}
				m.ReconRepo.On("GetAll", mock.Anything, &reconciliation.ReconciliationFilterRequest{}).Return(&commonModel.PaginationResponse{
					Data: mockRecons,
				}, nil)
			},
		},
		{
			name:    "failed get all recon",
			wantErr: true,
			setupMock: func(m *Mocker) {
				m.ReconRepo.On("GetAll", mock.Anything, &reconciliation.ReconciliationFilterRequest{}).Return(&commonModel.PaginationResponse{
					Data: []map[string]string{},
				}, errors.New("error"))

			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
			reconRepo := repoMocks.NewIReconciliationRepository(t)

			mock := &Mocker{

				ReconRepo: reconRepo,
			}

			tc.setupMock(mock)

			service := New(cfg, logger, reconRepo)

			result, err := service.ListRecon(ctx, &reconciliation.ReconciliationFilterRequest{})
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				
				// Test specific transformation for the data transformation test case
				if tc.name == "success get all recon with data transformation" {
					require.NotNil(t, result)
					require.NotNil(t, result.Data)
					
					// Assert that the data has been transformed to []*reconciliation.ReconciliationResponse
					responseData, ok := result.Data.([]*reconciliation.ReconciliationResponse)
					require.True(t, ok, "Data should be transformed to []*reconciliation.ReconciliationResponse")
					require.Len(t, responseData, 1)
					
					// Check that the reasons field is now a string instead of sql.NullString
					reconResponse := responseData[0]
					assert.Equal(t, "there is multiple errors, check at result file for more details", reconResponse.Reasons)
					
					// Verify other fields are correctly transformed
					assert.Equal(t, "01979b00-a2ba-7217-976b-f9651af470c1", reconResponse.UUID)
					assert.Equal(t, "RECON_VA.xlsx", reconResponse.OriginalName)
					assert.Equal(t, "SUCCESS", reconResponse.Status)
					assert.Equal(t, "febriansyah@harsya.com", reconResponse.CreatedBy)
					assert.Equal(t, time.Date(2025, 6, 23, 4, 16, 40, 0, time.UTC), reconResponse.CreatedAt)
					assert.Equal(t, time.Date(2025, 6, 23, 4, 16, 48, 0, time.UTC), reconResponse.UpdatedAt)
				}
			}

		})
	}
}
