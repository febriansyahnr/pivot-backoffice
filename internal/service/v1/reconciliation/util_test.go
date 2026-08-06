package reconciliation

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	xlsxMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/xlsx"
)

func TestGetRowsAndValidateBulkUpload(t *testing.T) {
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
	invalidRows := [][]string{
		{
			bulkUploadHeaders[columnTransactionReference],
			bulkUploadHeaders[columnTransactionDatetime],
			bulkUploadHeaders[columnAmount],
			bulkUploadHeaders[columnBank],
			bulkUploadHeaders[columnChannel],
		},
		{
			"REF123",
			"2023-01-01 10:00:00",
			"100000",
			"BCA",
			"VA",
		},
	}
	tests := []struct {
		name      string
		wantErr   bool
		setupMock func(*xlsxMock.Filer)
	}{
		{
			name:    "success get rows",
			wantErr: false,
			setupMock: func(f *xlsxMock.Filer) {
				f.On("GetRows", sheetNameToUpload, mock.Anything).Return(validRows, nil)
			},
		},
		{
			name:    "failed, error get rows",
			wantErr: true,
			setupMock: func(f *xlsxMock.Filer) {
				f.On("GetRows", sheetNameToUpload, mock.Anything).Return([][]string{}, errors.New("error"))
			},
		},
		{
			name:    "failed, empty rows",
			wantErr: true,
			setupMock: func(f *xlsxMock.Filer) {
				f.On("GetRows", sheetNameToUpload, mock.Anything).Return([][]string{}, nil)
			},
		},
		{
			name:    "failed, invalid headers",
			wantErr: true,
			setupMock: func(f *xlsxMock.Filer) {
				f.On("GetRows", sheetNameToUpload, mock.Anything).Return(invalidRows, nil)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := &ReconciliationService{}
			filer := xlsxMock.NewFiler(t)

			tc.setupMock(filer)

			rows, err := service.getRowsAndValidateBulkUpload(filer)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, rows)
			}
		})
	}
}
