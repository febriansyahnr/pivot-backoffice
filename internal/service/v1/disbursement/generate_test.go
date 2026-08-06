package disbursementService

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	gcsMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	bulkID := uuid.NewString()
	validRowData := []*disbursementModel.BulkPreviewResponse{{}}
	conf := config.Config{
		Environment: constant.EnvironmentStaging,
	}

	testCases := []struct {
		name      string
		wantErr   bool
		mockSetup func(mockRepo *repositoryMocks.IDisbursementRepository, mockGcs *gcsMocks.GCSService)
		rowData   []*disbursementModel.BulkPreviewResponse
	}{
		{
			name:    "SUCCESS",
			wantErr: false,
			mockSetup: func(mockRepo *repositoryMocks.IDisbursementRepository, mockGcs *gcsMocks.GCSService) {
				mockGcs.
					On(
						"UploadFile",
						constant.ValueCtxMockType(), constant.StringMockType(), constant.PtrBytesBufferMockType(), constant.BoolMockType(),
					).Return(&gcs.UploadMultipart{PublicURL: "publicURL"}, nil)
				mockGcs.
					On(
						"CreateSignedURL",
						constant.ValueCtxMockType(), constant.StringMockType(), constant.DurationMockType(),
					).Return("signedURL", nil)

				mockRepo.
					On(
						"UpdateBulkDisbursementFailedFileByID",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string"),
					).Return(nil)
			},
			rowData: validRowData,
		},
		{
			name:    "ERROR:UploadBulkDisbursementFile",
			wantErr: true,
			mockSetup: func(mockRepo *repositoryMocks.IDisbursementRepository, mockGcs *gcsMocks.GCSService) {
				mockGcs.
					On(
						"UploadFile",
						constant.ValueCtxMockType(), constant.StringMockType(), constant.PtrBytesBufferMockType(), constant.BoolMockType(),
					).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			rowData: validRowData,
		},
		{
			name:    "ERROR:UpdateBulkDisbursementFailedFileByID",
			wantErr: true,
			mockSetup: func(mockRepo *repositoryMocks.IDisbursementRepository, mockGcs *gcsMocks.GCSService) {
				mockGcs.
					On(
						"UploadFile",
						constant.ValueCtxMockType(), constant.StringMockType(), constant.PtrBytesBufferMockType(), constant.BoolMockType(),
					).Return(&gcs.UploadMultipart{PublicURL: "publicURL"}, nil)
				mockGcs.
					On(
						"CreateSignedURL",
						constant.ValueCtxMockType(), constant.StringMockType(), constant.DurationMockType(),
					).Return("signedURL", nil)

				mockRepo.
					On(
						"UpdateBulkDisbursementFailedFileByID",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string"),
					).Return(constant.ErrSomeErrorForUnitTest)
			},
			rowData: validRowData,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockGcs := gcsMocks.NewGCSService(t)
			merchantRepo := repositoryMocks.NewIMerchantRepository(t)
			mockRepo := repositoryMocks.NewIDisbursementRepository(t)

			tc.mockSetup(mockRepo, mockGcs)

			svc := New(&conf, pdkLoggerMock, merchantRepo, mockRepo, nil, nil, WithGoogleCloudStorage(mockGcs))

			response, err := svc.GenerateExcelAndUpdateInvalidBulkDisbursement(context.Background(), bulkID, tc.rowData)
			if tc.wantErr {
				require.Error(t, err)
				require.Empty(t, response)

			} else {
				assert.NoError(t, err)
				require.NotEmpty(t, response)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
