package reconciliation

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/reconciliation"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore"
	gcsMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	xlsxMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/xlsx"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestProcessFile(t *testing.T) {
	ctx := context.Background()
	mockFileContent := []byte("test content")
	testId := uuid.NewString()
	validRows := [][]string{
		{
			bulkUploadHeaders[columnTransactionDatetime],
			bulkUploadHeaders[columnTransactionReference],
			bulkUploadHeaders[columnTransactionReference2],
			bulkUploadHeaders[columnAmount],
			bulkUploadHeaders[columnBank],
			bulkUploadHeaders[columnChannel],
		},
		{
			"2023-01-01 10:00:00",
			"REF123",
			"REF123",
			"100000",
			"BCA",
			"VA",
		},
	}
	invalidRows := [][]string{
		{
			bulkUploadHeaders[columnTransactionDatetime],
			bulkUploadHeaders[columnTransactionReference],
			bulkUploadHeaders[columnTransactionReference2],
			bulkUploadHeaders[columnAmount],
			bulkUploadHeaders[columnBank],
			bulkUploadHeaders[columnChannel],
		},
		{
			"2023-01-0a",
			"REF123",
			"REF123",
			"100000",
			"BCA",
			"VA",
		},
	}
	amount := decimal.NewFromInt(100000)

	tests := []struct {
		name      string
		id        string
		wantErr   bool
		setupMock func(*Mocker)
	}{
		{
			name:    "success process file",
			id:      testId,
			wantErr: false,
			setupMock: func(m *Mocker) {
				m.ReconRepo.On("GetByUUID", mock.Anything, testId).Return(&reconciliation.Reconciliation{FilePath: "test/path"}, nil)
				m.Gcs.On("ReadAll", mock.Anything, mock.Anything, "test/path").Return(mockFileContent, nil)

				file := xlsxMock.NewFiler(t)
				file.On("GetRows", sheetNameToUpload, mock.Anything).Return(validRows, nil)
				file.On("Close").Return(nil)
				file.On("WriteToBuffer").Return(bytes.NewBuffer(mockFileContent), nil)
				file.On("NewSheet", mock.Anything).Return(1, nil)
				file.On("GetCellValue", mock.Anything, mock.Anything, mock.Anything).Return("March 20, 2025 05:26 AM", nil)
				file.On("SetSheetRow", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				m.Xlsx.On("OpenReader", mock.Anything).Return(file, nil)

				// Add transaction mocks
				m.AccountTransaction.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				m.AccountTransaction.On("CommitTransaction", mock.Anything).Return(nil)

				m.AccountTransaction.
					On("GetTransactionForRecon", mock.Anything, mock.Anything).
					Return(&reconciliation.ReconTransactionModel{
						Status: constant.StatusSuccess,
						Amount: amount,
					}, nil)

				m.AccountTransaction.On("UpdateReconDetail", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				m.AccountTransaction.On("UpdateBulkReconStatus", mock.Anything, mock.Anything).Return(nil)

				m.Gcs.On("UploadFile", mock.Anything, mock.Anything, mock.Anything, true).Return(&gcs.UploadMultipart{ObjectName: "result/path"}, nil)
				m.ReconRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "success process file - with va static",
			id:      testId,
			wantErr: false,
			setupMock: func(m *Mocker) {
				m.ReconRepo.On("GetByUUID", mock.Anything, testId).Return(&reconciliation.Reconciliation{FilePath: "test/path"}, nil)
				m.Gcs.On("ReadAll", mock.Anything, mock.Anything, "test/path").Return(mockFileContent, nil)

				file := xlsxMock.NewFiler(t)
				file.On("GetRows", sheetNameToUpload, mock.Anything).Return(validRows, nil)
				file.On("Close").Return(nil)
				file.On("WriteToBuffer").Return(bytes.NewBuffer(mockFileContent), nil)
				file.On("NewSheet", mock.Anything).Return(1, nil)
				file.On("GetCellValue", mock.Anything, mock.Anything, mock.Anything).Return("March 20, 2025 05:26 AM", nil)
				file.On("SetSheetRow", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				m.Xlsx.On("OpenReader", mock.Anything).Return(file, nil)

				// Add transaction mocks
				m.AccountTransaction.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				m.AccountTransaction.On("CommitTransaction", mock.Anything).Return(nil)

				m.AccountTransaction.
					On("GetTransactionForRecon", mock.Anything, mock.Anything).
					Return(&reconciliation.ReconTransactionModel{
						Status:      constant.StatusSuccess,
						Amount:      amount,
						PaymentType: constant.UnifiedPaymentTypeMultiple,
					}, nil)

				m.AccountTransaction.On("UpdateReconDetail", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				totalAmount := reconciliation.PaymentTotalAmountResult{}
				totalAmount.Add("REF123", amount)
				m.AccountTransaction.On("GetTotalPaymentAmount", mock.Anything, mock.Anything).Return(&totalAmount, nil)
				m.AccountTransaction.On("UpdateBulkReconStatus", mock.Anything, mock.Anything).Return(nil)

				m.Gcs.On("UploadFile", mock.Anything, mock.Anything, mock.Anything, true).Return(&gcs.UploadMultipart{ObjectName: "result/path"}, nil)
				m.ReconRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "success process file - with invalid total amount va static",
			id:      testId,
			wantErr: false,
			setupMock: func(m *Mocker) {
				m.ReconRepo.On("GetByUUID", mock.Anything, testId).Return(&reconciliation.Reconciliation{FilePath: "test/path"}, nil)
				m.Gcs.On("ReadAll", mock.Anything, mock.Anything, "test/path").Return(mockFileContent, nil)

				file := xlsxMock.NewFiler(t)
				file.On("GetRows", sheetNameToUpload, mock.Anything).Return(validRows, nil)
				file.On("Close").Return(nil)
				file.On("WriteToBuffer").Return(bytes.NewBuffer(mockFileContent), nil)
				file.On("NewSheet", mock.Anything).Return(1, nil)
				file.On("GetCellValue", mock.Anything, mock.Anything, mock.Anything).Return("March 20, 2025 05:26 AM", nil)
				file.On("SetSheetRow", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				m.Xlsx.On("OpenReader", mock.Anything).Return(file, nil)

				// Add transaction mocks
				m.AccountTransaction.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				m.AccountTransaction.On("CommitTransaction", mock.Anything).Return(nil)

				m.AccountTransaction.
					On("GetTransactionForRecon", mock.Anything, mock.Anything).
					Return(&reconciliation.ReconTransactionModel{
						Status:      constant.StatusSuccess,
						Amount:      amount,
						PaymentType: constant.UnifiedPaymentTypeMultiple,
					}, nil)

				totalAmount := reconciliation.PaymentTotalAmountResult{}
				totalAmount.Add("REF123", amount.Add(amount))
				m.AccountTransaction.On("GetTotalPaymentAmount", mock.Anything, mock.Anything).Return(&totalAmount, nil)
				m.AccountTransaction.On("UpdateBulkReconStatus", mock.Anything, mock.Anything).Return(nil)

				m.Gcs.On("UploadFile", mock.Anything, mock.Anything, mock.Anything, true).Return(&gcs.UploadMultipart{ObjectName: "result/path"}, nil)
				m.ReconRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "success process file - with failed recon status",
			id:      testId,
			wantErr: false,
			setupMock: func(m *Mocker) {
				m.ReconRepo.On("GetByUUID", mock.Anything, testId).Return(&reconciliation.Reconciliation{FilePath: "test/path"}, nil)
				m.Gcs.On("ReadAll", mock.Anything, mock.Anything, "test/path").Return(mockFileContent, nil)

				file := xlsxMock.NewFiler(t)
				file.On("GetRows", sheetNameToUpload, mock.Anything).Return(validRows, nil)
				file.On("Close").Return(nil)
				file.On("WriteToBuffer").Return(bytes.NewBuffer(mockFileContent), nil)
				file.On("NewSheet", mock.Anything).Return(1, nil)
				file.On("SetSheetRow", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				file.On("GetCellValue", mock.Anything, mock.Anything, mock.Anything).Return("March 20, 2025 05:26 AM", nil)
				m.Xlsx.On("OpenReader", mock.Anything).Return(file, nil)

				// Add transaction mocks
				m.AccountTransaction.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				m.AccountTransaction.On("CommitTransaction", mock.Anything).Return(nil)

				m.AccountTransaction.
					On("GetTransactionForRecon", mock.Anything, mock.Anything).
					Return(&reconciliation.ReconTransactionModel{
						Status: constant.StatusPending,
						Amount: amount,
					}, nil)

				m.SnapCore.
					On("CheckReconTransaction", mock.Anything, mock.Anything).
					Return(&snapCoreModel.AutoReconTrxResponse{
						Status: constant.StatusFailed,
						UUID:   uuid.NewString(),
					}, nil)

				m.AccountTransaction.On("UpdateBulkReconStatus", mock.Anything, mock.Anything).Return(nil)
				m.Gcs.On("UploadFile", mock.Anything, mock.Anything, mock.Anything, true).Return(&gcs.UploadMultipart{ObjectName: "result/path"}, nil)
				m.ReconRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "success process file - with invalid amount",
			id:      testId,
			wantErr: false,
			setupMock: func(m *Mocker) {
				m.ReconRepo.On("GetByUUID", mock.Anything, testId).Return(&reconciliation.Reconciliation{FilePath: "test/path"}, nil)
				m.Gcs.On("ReadAll", mock.Anything, mock.Anything, "test/path").Return(mockFileContent, nil)

				file := xlsxMock.NewFiler(t)
				file.On("GetRows", sheetNameToUpload, mock.Anything).Return(validRows, nil)
				file.On("Close").Return(nil)
				file.On("WriteToBuffer").Return(bytes.NewBuffer(mockFileContent), nil)
				file.On("NewSheet", mock.Anything).Return(1, nil)
				file.On("SetSheetRow", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				file.On("GetCellValue", mock.Anything, mock.Anything, mock.Anything).Return("March 20, 2025 05:26 AM", nil)
				m.Xlsx.On("OpenReader", mock.Anything).Return(file, nil)

				// Add transaction mocks
				m.AccountTransaction.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				m.AccountTransaction.On("CommitTransaction", mock.Anything).Return(nil)

				m.AccountTransaction.
					On("GetTransactionForRecon", mock.Anything, mock.Anything).
					Return(nil, nil)

				m.SnapCore.
					On("CheckReconTransaction", mock.Anything, mock.Anything).
					Return(&snapCoreModel.AutoReconTrxResponse{
						Status:               constant.ReconSnapStatusValid,
						Code:                 constant.ReconCodeOk,
						UUID:                 uuid.NewString(),
						ProcessorReferenceID: uuid.NewString(),
					}, nil)

				m.AccountTransaction.
					On("GetTransactionByProcessorID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(&reconciliation.ReconTransactionModel{
						Status: constant.StatusSuccess,
						Amount: decimal.NewFromInt(12000),
					}, nil)

				m.AccountTransaction.On("SetAdditionalInfoReconciliation", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				m.AccountTransaction.On("UpdateBulkReconStatus", mock.Anything, mock.Anything).Return(nil)

				m.Gcs.On("UploadFile", mock.Anything, mock.Anything, mock.Anything, true).Return(&gcs.UploadMultipart{ObjectName: "result/path"}, nil)
				m.ReconRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "success process file - transaction not found",
			id:      testId,
			wantErr: false,
			setupMock: func(m *Mocker) {
				m.ReconRepo.On("GetByUUID", mock.Anything, testId).Return(&reconciliation.Reconciliation{FilePath: "test/path"}, nil)
				m.Gcs.On("ReadAll", mock.Anything, mock.Anything, "test/path").Return(mockFileContent, nil)

				file := xlsxMock.NewFiler(t)
				file.On("GetRows", sheetNameToUpload, mock.Anything).Return(validRows, nil)
				file.On("Close").Return(nil)
				file.On("WriteToBuffer").Return(bytes.NewBuffer(mockFileContent), nil)
				file.On("NewSheet", mock.Anything).Return(1, nil)
				file.On("SetSheetRow", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				file.On("GetCellValue", mock.Anything, mock.Anything, mock.Anything).Return("March 20, 2025 05:26 AM", nil)
				m.Xlsx.On("OpenReader", mock.Anything).Return(file, nil)

				// Add transaction mocks
				m.AccountTransaction.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				m.AccountTransaction.On("CommitTransaction", mock.Anything).Return(nil)

				m.AccountTransaction.
					On("GetTransactionForRecon", mock.Anything, mock.Anything).
					Return(nil, nil)

				m.SnapCore.
					On("CheckReconTransaction", mock.Anything, mock.Anything).
					Return(&snapCoreModel.AutoReconTrxResponse{
						Status: constant.StatusFailed,
						UUID:   uuid.NewString(),
					}, nil)

				m.AccountTransaction.On("UpdateBulkReconStatus", mock.Anything, mock.Anything).Return(nil)
				m.Gcs.On("UploadFile", mock.Anything, mock.Anything, mock.Anything, true).Return(&gcs.UploadMultipart{ObjectName: "result/path"}, nil)
				m.ReconRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "success process file - snapcore va with status success",
			id:      testId,
			wantErr: false,
			setupMock: func(m *Mocker) {
				m.ReconRepo.On("GetByUUID", mock.Anything, testId).Return(&reconciliation.Reconciliation{FilePath: "test/path"}, nil)
				m.Gcs.On("ReadAll", mock.Anything, mock.Anything, "test/path").Return(mockFileContent, nil)

				file := xlsxMock.NewFiler(t)
				file.On("GetRows", sheetNameToUpload, mock.Anything).Return(validRows, nil)
				file.On("Close").Return(nil)
				file.On("WriteToBuffer").Return(bytes.NewBuffer(mockFileContent), nil)
				file.On("NewSheet", mock.Anything).Return(1, nil)
				file.On("SetSheetRow", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				file.On("GetCellValue", mock.Anything, mock.Anything, mock.Anything).Return("March 20, 2025 05:26 AM", nil)
				m.Xlsx.On("OpenReader", mock.Anything).Return(file, nil)

				// Add transaction mocks
				m.AccountTransaction.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				m.AccountTransaction.On("CommitTransaction", mock.Anything).Return(nil)

				trx := &reconciliation.ReconTransactionModel{
					Status:      constant.StatusPending,
					Processor:   constant.SnapCoreProcessor,
					ProcessorID: uuid.NewString(),
					Channel:     constant.ChannelVirtualAccount,
					Amount:      amount,
				}
				m.AccountTransaction.
					On("GetTransactionForRecon", mock.Anything, mock.Anything).
					Return(nil, nil)
				m.SnapCore.
					On("CheckReconTransaction", mock.Anything, mock.Anything).
					Return(&snapCoreModel.AutoReconTrxResponse{
						Status:               constant.ReconSnapStatusValid,
						Code:                 constant.ReconCodeOk,
						ProcessorReferenceID: uuid.NewString(),
					}, nil)
				m.AccountTransaction.
					On("GetTransactionByProcessorID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(trx, nil)

				m.AccountTransaction.On("SetAdditionalInfoReconciliation", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				m.AccountTransaction.On("UpdateBulkReconStatus", mock.Anything, mock.Anything).Return(nil)

				m.Gcs.On("UploadFile", mock.Anything, mock.Anything, mock.Anything, true).Return(&gcs.UploadMultipart{ObjectName: "result/path"}, nil)
				m.ReconRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "success process file - snapcore va with status failed",
			id:      testId,
			wantErr: false,
			setupMock: func(m *Mocker) {
				m.ReconRepo.On("GetByUUID", mock.Anything, testId).Return(&reconciliation.Reconciliation{FilePath: "test/path"}, nil)
				m.Gcs.On("ReadAll", mock.Anything, mock.Anything, "test/path").Return(mockFileContent, nil)

				file := xlsxMock.NewFiler(t)
				file.On("GetRows", sheetNameToUpload, mock.Anything).Return(validRows, nil)
				file.On("Close").Return(nil)
				file.On("WriteToBuffer").Return(bytes.NewBuffer(mockFileContent), nil)
				file.On("NewSheet", mock.Anything).Return(1, nil)
				file.On("SetSheetRow", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				file.On("GetCellValue", mock.Anything, mock.Anything, mock.Anything).Return("March 20, 2025 05:26 AM", nil)
				m.Xlsx.On("OpenReader", mock.Anything).Return(file, nil)

				// Add transaction mocks
				m.AccountTransaction.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				m.AccountTransaction.On("CommitTransaction", mock.Anything).Return(nil)

				m.AccountTransaction.
					On("GetTransactionForRecon", mock.Anything, mock.Anything).
					Return(nil, nil)
				m.SnapCore.
					On("CheckReconTransaction", mock.Anything, mock.Anything).
					Return(&snapCoreModel.AutoReconTrxResponse{
						Status: constant.StatusFailed,
					}, nil)

				m.AccountTransaction.On("UpdateBulkReconStatus", mock.Anything, mock.Anything).Return(nil)
				m.Gcs.On("UploadFile", mock.Anything, mock.Anything, mock.Anything, true).Return(&gcs.UploadMultipart{ObjectName: "result/path"}, nil)
				m.ReconRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "success process file - snapcore qris with error when request to snapcore",
			id:      testId,
			wantErr: false,
			setupMock: func(m *Mocker) {
				m.ReconRepo.On("GetByUUID", mock.Anything, testId).Return(&reconciliation.Reconciliation{FilePath: "test/path"}, nil)
				m.Gcs.On("ReadAll", mock.Anything, mock.Anything, "test/path").Return(mockFileContent, nil)

				file := xlsxMock.NewFiler(t)
				file.On("GetRows", sheetNameToUpload, mock.Anything).Return(validRows, nil)
				file.On("Close").Return(nil)
				file.On("WriteToBuffer").Return(bytes.NewBuffer(mockFileContent), nil)
				file.On("NewSheet", mock.Anything).Return(1, nil)
				file.On("SetSheetRow", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				file.On("GetCellValue", mock.Anything, mock.Anything, mock.Anything).Return("March 20, 2025 05:26 AM", nil)
				m.Xlsx.On("OpenReader", mock.Anything).Return(file, nil)

				// Add transaction mocks
				m.AccountTransaction.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				m.AccountTransaction.On("CommitTransaction", mock.Anything).Return(nil)

				m.AccountTransaction.
					On("GetTransactionForRecon", mock.Anything, mock.Anything).
					Return(nil, nil)
				m.SnapCore.
					On("CheckReconTransaction", mock.Anything, mock.Anything).
					Return(nil, errors.New("error"))

				m.AccountTransaction.On("UpdateBulkReconStatus", mock.Anything, mock.Anything).Return(nil)
				m.Gcs.On("UploadFile", mock.Anything, mock.Anything, mock.Anything, true).Return(&gcs.UploadMultipart{ObjectName: "result/path"}, nil)
				m.ReconRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "failed process file - error when fillsheet",
			id:      testId,
			wantErr: true,
			setupMock: func(m *Mocker) {
				m.ReconRepo.On("GetByUUID", mock.Anything, testId).Return(&reconciliation.Reconciliation{FilePath: "test/path"}, nil)
				m.Gcs.On("ReadAll", mock.Anything, mock.Anything, "test/path").Return(mockFileContent, nil)

				file := xlsxMock.NewFiler(t)
				m.Xlsx.On("OpenReader", mock.Anything).Return(file, nil)
				file.On("GetRows", sheetNameToUpload, mock.Anything).Return(validRows, nil)
				file.On("GetCellValue", mock.Anything, mock.Anything, mock.Anything).Return("March 20, 2025 05:26 AM", nil)
				file.On("Close").Return(nil)

				// Add transaction mocks
				m.AccountTransaction.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				m.AccountTransaction.On("UpdateBulkReconStatus", mock.Anything, mock.Anything).Return(nil)
				m.AccountTransaction.On("CommitTransaction", mock.Anything).Return(nil)

				m.AccountTransaction.
					On("GetTransactionForRecon", mock.Anything, mock.Anything).
					Return(&reconciliation.ReconTransactionModel{
						Status: constant.StatusSuccess,
						Amount: amount,
					}, nil)
				m.AccountTransaction.On("UpdateReconDetail", mock.Anything, mock.Anything, mock.Anything).Return(nil)

				file.On("NewSheet", mock.Anything).Return(1, nil)
				file.On("SetSheetRow", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error"))
			},
		},
		{
			name:    "failed process file - error when create new sheet",
			id:      testId,
			wantErr: true,
			setupMock: func(m *Mocker) {
				m.ReconRepo.On("GetByUUID", mock.Anything, testId).Return(&reconciliation.Reconciliation{FilePath: "test/path"}, nil)
				m.Gcs.On("ReadAll", mock.Anything, mock.Anything, "test/path").Return(mockFileContent, nil)

				file := xlsxMock.NewFiler(t)
				m.Xlsx.On("OpenReader", mock.Anything).Return(file, nil)
				file.On("GetRows", sheetNameToUpload, mock.Anything).Return(validRows, nil)
				file.On("GetCellValue", mock.Anything, mock.Anything, mock.Anything).Return("March 20, 2025 05:26 AM", nil)
				file.On("Close").Return(nil)

				// Add transaction mocks
				m.AccountTransaction.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				m.AccountTransaction.On("UpdateBulkReconStatus", mock.Anything, mock.Anything).Return(nil)
				m.AccountTransaction.On("CommitTransaction", mock.Anything).Return(nil)

				m.AccountTransaction.
					On("GetTransactionForRecon", mock.Anything, mock.Anything).
					Return(&reconciliation.ReconTransactionModel{
						Status: constant.StatusSuccess,
						Amount: amount,
					}, nil)
				m.AccountTransaction.On("UpdateReconDetail", mock.Anything, mock.Anything, mock.Anything).Return(nil)

				file.On("NewSheet", mock.Anything).Return(-1, errors.New("error"))
				// file.On("SetSheetRow", mock.Anything, mock.Anything, mock.Anything).Return()
			},
		},
		{
			name:    "failed process file - error when upload result to gcs",
			id:      testId,
			wantErr: true,
			setupMock: func(m *Mocker) {
				m.ReconRepo.On("GetByUUID", mock.Anything, testId).Return(&reconciliation.Reconciliation{FilePath: "test/path"}, nil)
				m.Gcs.On("ReadAll", mock.Anything, mock.Anything, "test/path").Return(mockFileContent, nil)

				file := xlsxMock.NewFiler(t)
				file.On("GetRows", sheetNameToUpload, mock.Anything).Return(validRows, nil)
				file.On("Close").Return(nil)
				file.On("WriteToBuffer").Return(bytes.NewBuffer(mockFileContent), nil)
				file.On("NewSheet", mock.Anything).Return(1, nil)
				file.On("SetSheetRow", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				file.On("GetCellValue", mock.Anything, mock.Anything, mock.Anything).Return("March 20, 2025 05:26 AM", nil)
				m.Xlsx.On("OpenReader", mock.Anything).Return(file, nil)

				// Add transaction mocks
				m.AccountTransaction.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				m.AccountTransaction.On("CommitTransaction", mock.Anything).Return(nil)

				m.AccountTransaction.
					On("GetTransactionForRecon", mock.Anything, mock.Anything).
					Return(&reconciliation.ReconTransactionModel{
						Status: constant.StatusSuccess,
						Amount: amount,
					}, nil)

				m.AccountTransaction.On("UpdateReconDetail", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				m.AccountTransaction.On("UpdateBulkReconStatus", mock.Anything, mock.Anything).Return(nil)

				m.Gcs.On("UploadFile", mock.Anything, mock.Anything, mock.Anything, true).Return(&gcs.UploadMultipart{ObjectName: "result/path"}, errors.New("error"))
			},
		},
		{
			name:    "failed process file - error when update recon table",
			id:      testId,
			wantErr: true,
			setupMock: func(m *Mocker) {
				m.ReconRepo.On("GetByUUID", mock.Anything, testId).Return(&reconciliation.Reconciliation{FilePath: "test/path"}, nil)
				m.Gcs.On("ReadAll", mock.Anything, mock.Anything, "test/path").Return(mockFileContent, nil)

				file := xlsxMock.NewFiler(t)
				file.On("GetRows", sheetNameToUpload, mock.Anything).Return(validRows, nil)
				file.On("Close").Return(nil)
				file.On("WriteToBuffer").Return(bytes.NewBuffer(mockFileContent), nil)
				file.On("NewSheet", mock.Anything).Return(1, nil)
				file.On("SetSheetRow", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				file.On("GetCellValue", mock.Anything, mock.Anything, mock.Anything).Return("March 20, 2025 05:26 AM", nil)
				m.Xlsx.On("OpenReader", mock.Anything).Return(file, nil)

				// Add transaction mocks
				m.AccountTransaction.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				m.AccountTransaction.On("CommitTransaction", mock.Anything).Return(nil)

				m.AccountTransaction.
					On("GetTransactionForRecon", mock.Anything, mock.Anything).
					Return(&reconciliation.ReconTransactionModel{
						Status: constant.StatusSuccess,
						Amount: amount,
					}, nil)

				m.AccountTransaction.On("UpdateReconDetail", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				m.AccountTransaction.On("UpdateBulkReconStatus", mock.Anything, mock.Anything).Return(nil)

				m.Gcs.On("UploadFile", mock.Anything, mock.Anything, mock.Anything, true).Return(&gcs.UploadMultipart{ObjectName: "result/path"}, nil)
				m.ReconRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("error"))
			},
		},
		{
			name:    "failed process file - error when create buffer for result file",
			id:      testId,
			wantErr: true,
			setupMock: func(m *Mocker) {
				m.ReconRepo.On("GetByUUID", mock.Anything, testId).Return(&reconciliation.Reconciliation{FilePath: "test/path"}, nil)
				m.Gcs.On("ReadAll", mock.Anything, mock.Anything, "test/path").Return(mockFileContent, nil)

				file := xlsxMock.NewFiler(t)
				file.On("GetRows", sheetNameToUpload, mock.Anything).Return(validRows, nil)
				file.On("Close").Return(nil)
				file.On("NewSheet", mock.Anything).Return(1, nil)
				file.On("SetSheetRow", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				file.On("GetCellValue", mock.Anything, mock.Anything, mock.Anything).Return("March 20, 2025 05:26 AM", nil)
				m.Xlsx.On("OpenReader", mock.Anything).Return(file, nil)

				// Add transaction mocks
				m.AccountTransaction.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				m.AccountTransaction.On("CommitTransaction", mock.Anything).Return(nil)

				m.AccountTransaction.
					On("GetTransactionForRecon", mock.Anything, mock.Anything).
					Return(&reconciliation.ReconTransactionModel{
						Status: constant.StatusSuccess,
						Amount: amount,
					}, nil)

				m.AccountTransaction.On("UpdateReconDetail", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				m.AccountTransaction.On("UpdateBulkReconStatus", mock.Anything, mock.Anything).Return(nil)

				file.On("WriteToBuffer").Return(bytes.NewBuffer(mockFileContent), errors.New("error"))
			},
		},
		{
			name:    "failed process file - error when open reader",
			id:      testId,
			wantErr: true,
			setupMock: func(m *Mocker) {
				m.ReconRepo.On("GetByUUID", mock.Anything, testId).Return(&reconciliation.Reconciliation{FilePath: "test/path"}, nil)
				m.Gcs.On("ReadAll", mock.Anything, mock.Anything, "test/path").Return(mockFileContent, nil)

				file := xlsxMock.NewFiler(t)
				m.Xlsx.On("OpenReader", mock.Anything).Return(file, errors.New("error"))
			},
		},
		{
			name:    "failed process file - error when read all file",
			id:      testId,
			wantErr: true,
			setupMock: func(m *Mocker) {
				m.ReconRepo.On("GetByUUID", mock.Anything, testId).Return(&reconciliation.Reconciliation{FilePath: "test/path"}, nil)
				m.Gcs.On("ReadAll", mock.Anything, mock.Anything, "test/path").Return(mockFileContent, errors.New("error"))
			},
		},
		{
			name:    "failed process file - error when get recon by uuid",
			id:      testId,
			wantErr: true,
			setupMock: func(m *Mocker) {
				m.ReconRepo.On("GetByUUID", mock.Anything, testId).Return(&reconciliation.Reconciliation{FilePath: "test/path"}, errors.New("error"))
			},
		},
		{
			name:    "success process file - but recon detail not found",
			id:      testId,
			wantErr: false,
			setupMock: func(m *Mocker) {
				m.ReconRepo.On("GetByUUID", mock.Anything, testId).Return(nil, constant.ErrDataNotFound)
			},
		},
		{
			name:    "success process file - with invalid date format",
			id:      testId,
			wantErr: false,
			setupMock: func(m *Mocker) {
				m.ReconRepo.On("GetByUUID", mock.Anything, testId).Return(&reconciliation.Reconciliation{FilePath: "test/path"}, nil)
				m.Gcs.On("ReadAll", mock.Anything, mock.Anything, "test/path").Return(mockFileContent, nil)

				file := xlsxMock.NewFiler(t)
				file.On("GetRows", sheetNameToUpload, mock.Anything).Return(invalidRows, nil)
				file.On("Close").Return(nil)
				m.Xlsx.On("OpenReader", mock.Anything).Return(file, nil)
				file.On("GetCellValue", mock.Anything, mock.Anything, mock.Anything).Return("2006-01-0a", nil)
				file.On("NewSheet", mock.Anything).Return(1, nil)
				file.On("SetSheetRow", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				file.On("WriteToBuffer").Return(bytes.NewBuffer(mockFileContent), nil)

				// Add transaction mocks
				m.AccountTransaction.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				m.AccountTransaction.On("CommitTransaction", mock.Anything).Return(nil)
				m.Gcs.On("UploadFile", mock.Anything, mock.Anything, mock.Anything, true).Return(&gcs.UploadMultipart{ObjectName: "result/path"}, nil)
				m.ReconRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
			reconRepo := repoMocks.NewIReconciliationRepository(t)
			gcs := gcsMock.NewGCSService(t)
			xlsx := xlsxMock.NewExceler(t)
			snapCore := repoMocks.NewISnapCoreRepository(t)
			accountTrx := repoMocks.NewIAccountTransactionRepository(t)

			mock := &Mocker{
				ReconRepo:          reconRepo,
				Gcs:                gcs,
				Xlsx:               xlsx,
				SnapCore:           snapCore,
				AccountTransaction: accountTrx,
			}

			tc.setupMock(mock)

			service := New(
				&config.Config{},
				logger,
				reconRepo,
				WithExcelService(xlsx),
				WithGCSService(gcs),
				WithAccountTransactionRepository(accountTrx),
				WithSnapCoreRepository(snapCore),
			)

			err := service.ProcessFile(ctx, tc.id)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
