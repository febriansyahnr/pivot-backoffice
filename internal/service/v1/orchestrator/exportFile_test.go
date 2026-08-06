package orchestrator_service_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/orchestrator"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestGenExcelForTransactionHistories(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot, _ := util.FindProjectRoot(cwd, "backend-portal")
	targetPath := filepath.Join(projectRoot, "test", "consul", "backend-portal", "feature-flag.yaml")

	_ = ffclient.Init(ffclient.Config{
		Retriever:    &fileretriever.Retriever{Path: targetPath},
		DataExporter: ffclient.DataExporter{},
	})
	defer ffclient.Close()

	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	reportingRepo := repoMocks.NewIReportingRepository(t)
	accountTransactionRepo := repoMocks.NewIAccountTransactionRepository(t)

	service := New(logger, nil, accountTransactionRepo, nil)

	WithReportingRepository(service, reportingRepo)

	traceId := uuid.NewString()
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId)

	startDate := time.Date(2024, 6, 30, 17, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		merchantId string
		setupMock  func()
		wantErr    string
		wantResult map[string]string
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				accountTransactionRepo.On("GetListTransactionHistories", c.ValueCtxMockType(), mock.Anything).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf(c.InternalErrorFmt, traceId),
		},
		{
			name:       "ERROR:Some error on reporting",
			merchantId: "79f52ad0-4820-46fd-8d38-10e5a0a8514c", // NOSONAR
			setupMock: func() {
				reportingRepo.On("ExportBalanceHistory", c.ValueCtxMockType(), mock.Anything).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf(c.InternalErrorFmt, traceId),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				accountTransactionRepo.On(
					"GetListTransactionHistories", c.ValueCtxMockType(), mock.Anything,
				).Return([]model.TransactionHistory{
					{
						Id:                  traceId,
						MerchantReferenceID: "ucup/BRI/0013",
						Type:                c.TypeDisbursement,
						BalanceType:         "Payout Balance",
						Amount:              125_325_450.6666,
						Fee:                 0,  // Explicitly set Fee
						Channel:             "", // Explicitly set Channel (empty for this test)
						CreatedBy:           "-",
						ApprovedBy:          "Agus",
						SettlementAt:        &startDate,
						Status:              "FAILED",
					},
				}, nil)
			},
			wantResult: map[string]string{
				"A1": "Settlement Report - Harsya",
				"A2": "01 July 2024 - 31 July 2024",
				"A3": "",
				"A4": "Reference ID",
				"B4": "Balance Type",
				"C4": "Transaction Type",
				"D4": "Channel",
				"E4": "Created Date",
				"F4": "Settlement Date",
				"G4": "Created By",
				"H4": "Transaction ID",
				"I4": "Amount",
				"J4": "Fee",
				"K4": "Bank Reference",
				"L4": "Transaction Status",
				"M4": "Reason Type",
				"N4": "Reason Description",
				"O4": "Remarks",
				"P4": "Destination Bank",
				"Q4": "Destination Account",
				"R4": "Beneficiary Name",
				// Data row (row 5)
				"A5": "ucup/BRI/0013",            // Reference ID = MerchantReferenceID
				"B5": "Payout Balance",           // Balance Type
				"C5": "Single Payout",            // Transaction Type
				"D5": "",                         // Channel (empty)
				"E5": "01 January 1 07:07:12 AM", // Created Date
				"F5": "01 July 2024 12:00:00 AM", // Settlement Date
				"G5": "-",                        // Created By
				"H5": traceId,                    // Transaction ID = Id
				"I5": "Rp 125,325,450.67",        // Amount
				"J5": "Rp 0.00",                  // Fee
				"K5": "",                         // Bank Reference
				"L5": "FAILED",                   // Transaction Status
				"M5": "",                         // Reason Type
				"N5": "",                         // Reason Description
				"O5": "",                         // Remarks
				"P5": "",                         // Destination Bank
				"Q5": "",                         // Destination Account
				"R5": "",                         // Beneficiary Name
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			buff := new(bytes.Buffer)
			defer buff.Reset()

			wr := &model.FileWriter{Writer: buff}
			filter := &model.TransactionHistoryFilterRequest{
				StartDate:  startDate,
				EndDate:    time.Date(2024, 7, 31, 6, 0, 0, 0, time.UTC),
				MerchantID: test.merchantId,
			}
			if err := service.GenExcelForTransactionHistories(ctx, wr, filter); test.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
				return
			}

			f, err := excelize.OpenReader(buff)
			require.NoError(t, err)
			defer assert.NoError(t, f.Close())

			sheetName := f.GetSheetName(0)
			require.Equal(t, "Sheet1", sheetName)

			for cell, want := range test.wantResult {
				val, err := f.GetCellValue(sheetName, cell)

				require.NoError(t, err)
				assert.Equal(t, want, val)
			}

			reportingRepo.AssertExpectations(t)
			accountTransactionRepo.AssertExpectations(t)
		})
	}
}
