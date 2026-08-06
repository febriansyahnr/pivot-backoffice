package reconciliation

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paper-indonesia/pivot-backoffice/config"
	gcsMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	rabbitmqExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	xlsxMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/xlsx"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
)

func TestUploadFile(t *testing.T) {
	ctx := context.Background()
	mockFileContent := []byte("test content")
	mockReader := bytes.NewReader(mockFileContent)
	validRows := [][]string{
		{
			bulkUploadHeaders[columnTransactionDatetime],
			bulkUploadHeaders[columnTransactionReference],
			bulkUploadHeaders[columnTransactionReference2],
			bulkUploadHeaders[columnAmount],
			bulkUploadHeaders[columnBank],
			bulkUploadHeaders[columnChannel],
			bulkUploadHeaders[10],
		},
		{
			"2023-01-01 10:00:00",
			"REF123",
			"100000",
			"BCA",
			"VA",
		},
	}
	validFiler := xlsxMock.NewFiler(t)
	validFiler.On("GetRows", sheetNameToUpload, mock.Anything).Return(validRows, nil)
	validFiler.On("Close").Return(nil)

	invalidFiler := xlsxMock.NewFiler(t)
	invalidFiler.On("GetRows", sheetNameToUpload, mock.Anything).Return([][]string{}, errors.New("error"))
	invalidFiler.On("Close").Return(nil)

	tests := []struct {
		name      string
		createdBy string
		wantErr   bool
		setupMock func(*Mocker)
	}{
		{
			name:      "success upload file",
			createdBy: "test",
			wantErr:   false,
			setupMock: func(m *Mocker) {
				m.Xlsx.On("OpenReader", mock.Anything).Return(validFiler, nil)
				m.Gcs.On("UploadFile", c.ValueCtxMockType(), c.StringMockType(), c.PtrBytesBufferMockType(), c.BoolMockType(), mock.Anything).Return(&gcs.UploadMultipart{PublicURL: "publicURL"}, nil)
				m.ReconRepo.On("Create", mock.Anything, mock.AnythingOfType("*reconciliation.Reconciliation")).Return(nil)
				m.RabbitMq.On("Publish", mock.Anything, rabbitMqExt.ReconProcessRoutingKey, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:      "success upload file, with failed publish message",
			createdBy: "test",
			wantErr:   false,
			setupMock: func(m *Mocker) {
				m.Xlsx.On("OpenReader", mock.Anything).Return(validFiler, nil)
				m.Gcs.On("UploadFile", c.ValueCtxMockType(), c.StringMockType(), c.PtrBytesBufferMockType(), c.BoolMockType(), mock.Anything).Return(&gcs.UploadMultipart{PublicURL: "publicURL"}, nil)
				m.ReconRepo.On("Create", mock.Anything, mock.AnythingOfType("*reconciliation.Reconciliation")).Return(nil)
				m.RabbitMq.On("Publish", mock.Anything, rabbitMqExt.ReconProcessRoutingKey, mock.Anything, mock.Anything).Return(errors.New("error"))

			},
		},
		{
			name:      "failed: error when validate",
			createdBy: "test",
			wantErr:   true,
			setupMock: func(m *Mocker) {
				m.Xlsx.On("OpenReader", mock.Anything).Return(invalidFiler, nil)
			},
		},
		{
			name:      "failed: error when upload file",
			createdBy: "test",
			wantErr:   true,
			setupMock: func(m *Mocker) {
				m.Xlsx.On("OpenReader", mock.Anything).Return(validFiler, nil)
				m.Gcs.On("UploadFile", c.ValueCtxMockType(), c.StringMockType(), c.PtrBytesBufferMockType(), c.BoolMockType(), mock.Anything).Return(nil, errors.New("error"))
			},
		},
		{
			name:      "failed: error when create reconciliation",
			createdBy: "test",
			wantErr:   true,
			setupMock: func(m *Mocker) {
				m.Xlsx.On("OpenReader", mock.Anything).Return(validFiler, nil)
				m.Gcs.On("UploadFile", c.ValueCtxMockType(), c.StringMockType(), c.PtrBytesBufferMockType(), c.BoolMockType(), mock.Anything).Return(&gcs.UploadMultipart{PublicURL: "publicURL"}, nil)
				m.ReconRepo.On("Create", mock.Anything, mock.AnythingOfType("*reconciliation.Reconciliation")).Return(errors.New("error"))
			},
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

			uuid, err := service.UploadFile(ctx, c.TypePayment, tc.createdBy, io.NopCloser(mockReader), &multipart.FileHeader{
				Filename: "test_absdzxcasdqweutqwomqwekajsdzmnxcjaacjaasdqwsdqweqweasdasdqskdmqwekajszmnxcjaasdqwesajszmnxcjaasd.xlsx",
			})
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, uuid)
			}
		})
	}
}
