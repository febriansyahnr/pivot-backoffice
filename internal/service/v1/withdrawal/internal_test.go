package withdrawalService

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestGetCacheDownloadHistory(t *testing.T) {
	rdb, rdbMock := redismock.NewClientMock()

	service := &withdrawalService{
		redis: redisExt.WrapRedisClient(rdb, nil),
	}

	hash := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
	redisKey := fmt.Sprintf(c.RedisKeyDownloadWithdrawalHistoryFmt, hash)

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult string
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				rdbMock.ExpectGet(redisKey).SetErr(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS:Cache not found",
			setupMock: func() {
				rdbMock.ExpectGet(redisKey).SetErr(redisExt.ErrNil)
			},
		},
		{
			name: "SUCCESS:Cache found",
			setupMock: func() {
				rdbMock.ExpectGet(redisKey).SetVal("https://cache-found")
			},
			wantResult: "https://cache-found",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rdbMock.ClearExpect()

			test.setupMock()
			result, err := service.GetCacheDownloadHistory(context.Background(), hash)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestExportToExcel(t *testing.T) {
	service := &withdrawalService{}

	request := &withdrawal.WithdrawalListRequest{
		StartDate: time.Date(2024, 10, 31, 17, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 11, 14, 16, 59, 59, 0, time.UTC),
	}
	sheetName := "Sheet1"
	headers := map[string]string{
		"A4": "Created Date",
		"B4": "Last Update",
		"C4": "Created By",
		"D4": "Amount",
		"E4": "Status",
		"F4": "Transaction ID",
		"G4": "Bank Reference",
		"H4": "Bank Name",
		"I4": "Account Number",
		"J4": "Beneficiary Name",
	}

	tests := []struct {
		name             string
		transactions     []withdrawal.WithdrawalHistoryResponse
		wantExcelContent map[string]string
	}{
		{
			name:         "No transactions",
			transactions: []withdrawal.WithdrawalHistoryResponse{},
			wantExcelContent: map[string]string{
				"A1": "Payment Balance Withdrawal History",
				"A2": "01 November 2024 - 14 November 2024",
				"A3": "",
			},
		},
		{
			name: "Transactions found",
			transactions: []withdrawal.WithdrawalHistoryResponse{
				{
					Id:                     "c6a03dd1-d73d-4aae-882b-b592c41a8b64",
					Date:                   time.Date(2024, 11, 8, 3, 55, 27, 0, time.UTC),
					Amount:                 10_000,
					BeneficiaryBankName:    "Bank Rakyat Indonesia",
					BeneficiaryAccountName: "Dummy Simulation",
					Status:                 "SUCCESS",
					CreatedBy:              "John Doe",
					UpdatedAt:              time.Date(2024, 11, 8, 3, 56, 01, 0, time.UTC),
					BankReference:          "20220323150251228",
					BeneficiaryAccountNo:   "999966660001",
				},
			},
			wantExcelContent: map[string]string{
				"A1": "Payment Balance Withdrawal History",
				"A2": "01 November 2024 - 14 November 2024",
				"A3": "",
				"A5": "08 Nov 2024, 10:55 AM",
				"B5": "08 Nov 2024, 10:56 AM",
				"C5": "John Doe",
				"D5": "Rp 10,000",
				"E5": "Success",
				"F5": "c6a03dd1-d73d-4aae-882b-b592c41a8b64",
				"G5": "20220323150251228",
				"H5": "Bank Rakyat Indonesia",
				"I5": "999966660001",
				"J5": "Dummy Simulation",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := &bytes.Buffer{}

			assert.NoError(t, service.ExportToExcel(context.Background(), request, test.transactions, buf))

			f, err := excelize.OpenReader(buf)
			require.NoError(t, err)

			idx, err := f.GetSheetIndex(sheetName)
			require.NoError(t, err)
			require.Greater(t, idx, -1)

			maps.Copy(test.wantExcelContent, headers)

			for cell, want := range test.wantExcelContent {

				val, err := f.GetCellValue(sheetName, cell)
				assert.NoError(t, err)
				assert.Equal(t, want, val)
			}
		})
	}
}

func TestGetTimeDurationPublishAutoWithdrawal(t *testing.T) {
	local, err := time.LoadLocation(c.TimeLoc)
	require.NoError(t, err)

	config := &config.WithdrawalConfig{
		AutoWithdrawalDefaultSchedulingTime:        "23:00:00",
		AutoWithdrawalBankCodeToBeExecAfterTrigger: []string{"008"},
	}

	service := withdrawalService{
		config: config,
	}

	tests := []struct {
		name           string
		request        merchantModel.MerchantWithActiveAutoWithdrawalStatus
		schedulingTime string
		assertError    func(t *testing.T, err error)
		assertDuration func(t *testing.T, duration time.Duration)
	}{
		{
			name: "No delay",
			request: merchantModel.MerchantWithActiveAutoWithdrawalStatus{
				BeneficiaryBankCode: "008", // NOSONAR
			},
			assertError:    func(t *testing.T, err error) { /* Empty */ },
			assertDuration: func(t *testing.T, duration time.Duration) { assert.Equal(t, time.Duration(0), duration) },
		},
		{
			name: "Error parse scheduling time",
			request: merchantModel.MerchantWithActiveAutoWithdrawalStatus{
				BeneficiaryBankCode: "014", // NOSONAR
			},
			schedulingTime: "XXX", // NOSONAR
			assertError: func(t *testing.T, err error) {
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), `parsing time "XXX" as "15:04:05"`)
				}
			},
			assertDuration: func(t *testing.T, duration time.Duration) { assert.Equal(t, time.Duration(0), duration) },
		},
		{
			name: "Scheduled today",
			request: merchantModel.MerchantWithActiveAutoWithdrawalStatus{
				BeneficiaryBankCode: "014", // NOSONAR
			},
			schedulingTime: func() string {
				t1 := time.Now().In(local)
				t2 := t1.Add(time.Hour)
				if t1.Day() == t2.Day() {
					return fmt.Sprintf("%02d:00:00", t2.Hour())
				}
				return fmt.Sprintf("%02d:00:00", t1.Hour())
			}(),
			assertError: func(t *testing.T, err error) { assert.NoError(t, err) },
			assertDuration: func(t *testing.T, duration time.Duration) {
				assert.Greater(t, duration, time.Minute)
				now := time.Now().In(local)
				if now.Hour() == 23 {
					assert.Greater(t, duration, time.Hour+(30*time.Minute))
				} else {
					assert.Less(t, duration, time.Hour+(30*time.Minute))
				}
			},
		},
		{
			name: "Scheduled for tomorrow",
			request: merchantModel.MerchantWithActiveAutoWithdrawalStatus{
				BeneficiaryBankCode: "014", // NOSONAR
			},
			schedulingTime: func() string {
				now := time.Now().In(local).Add(-time.Hour)
				return fmt.Sprintf("%02d:00:00", now.Hour())
			}(),
			assertError:    func(t *testing.T, err error) { assert.NoError(t, err) },
			assertDuration: func(t *testing.T, duration time.Duration) { assert.Greater(t, duration, time.Hour+(30*time.Minute)) },
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config.AutoWithdrawalDefaultSchedulingTime = test.schedulingTime

			duration, err := service.getTimeDurationPublishAutoWithdrawal(test.request)

			test.assertError(t, err)
			test.assertDuration(t, duration)
		})
	}
}
