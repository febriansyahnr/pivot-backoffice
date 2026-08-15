package disbursementService

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	redisMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	xlsxMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/xlsx"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckExistReferenceID(t *testing.T) {
	disbursementRepo := repoMocks.NewIDisbursementRepository(t)

	service := &DisbursementService{
		disbursementRepo: disbursementRepo,
	}
	referenceList := map[string]bool{}

	tests := []struct {
		referenceId string
		setupMock   func()
		wantResult  bool
	}{
		{
			referenceId: "REF001",
			setupMock: func() {
				disbursementRepo.On(
					"CountByMerchantAndReference", mock.Anything, c.StringMockType(), c.StringMockType(),
				).Once().Return(0, nil)
			},
			wantResult: false,
		},
		{
			referenceId: "REF002",
			setupMock: func() {
				disbursementRepo.On(
					"CountByMerchantAndReference", mock.Anything, c.StringMockType(), c.StringMockType(),
				).Once().Return(1, nil)
			},
			wantResult: true,
		},
	}
	for _, test := range tests {
		t.Run(test.referenceId, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantResult, service.isExistReferenceID(context.Background(), "1234", test.referenceId, referenceList))
			referenceList[test.referenceId] = true
		})
	}
}

func TestGetRowsAndValidateBulkUpload(t *testing.T) {
	file := xlsxMock.NewFiler(t)

	service := &DisbursementService{}
	validResult := [][]string{
		{"Reference ID", "Amount", "Channel Code", "Account Number", "Account Name", "Remarks", "", ""}, {},
	}
	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult [][]string
	}{
		{
			name: "ERROR:Template sheetname not found",
			setupMock: func() {
				file.On("GetRows", c.StringMockType(), c.XlsxOptionsMockType()).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("sheet to upload not found")), // NOSONAR
		},
		{
			name: "ERROR:Data not found",
			setupMock: func() {
				file.On("GetRows", c.StringMockType(), c.XlsxOptionsMockType()).Once().Return(nil, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("empty data to upload")), // NOSONAR
		},
		{
			name: "ERROR:Max data upload",
			setupMock: func() {
				file.On("GetRows", c.StringMockType(), c.XlsxOptionsMockType()).Once().Return(make([][]string, 1002), nil)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("max row data is 1000")), // NOSONAR
		},
		{
			name: "ERROR:Header name not match",
			setupMock: func() {
				file.On("GetRows", c.StringMockType(), c.XlsxOptionsMockType()).Once().Return([][]string{
					{"", ""}, {},
				}, nil)
			},
			wantErr: c.ErrHeaderColumnDoesNotMatchWithTemplate,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				file.On("GetRows", c.StringMockType(), c.XlsxOptionsMockType()).Once().Return(validResult, nil)
			},
			wantResult: validResult,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.getRowsAndValidateBulkUpload(file)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestSingleRowValidation(t *testing.T) {
	disbursementRepo := repoMocks.NewIDisbursementRepository(t)
	disbursementRepo.On(
		"CountByMerchantAndReference", mock.Anything, c.StringMockType(), c.StringMockType(),
	).Return(0, nil)
	rdb := redisMock.NewIRedisExt(t)
	result := &redis.StringCmd{}
	result.SetErr(c.ErrSomeErrorForUnitTest)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	service := &DisbursementService{
		disbursementRepo: disbursementRepo,
		redisExt:         rdb,
		logger:           logger,
	}

	trxConfig := &disbursementModel.TransactionConfig{
		MinAmount: 10_000,
		MaxAmount: 100_000,
	}
	referenceList := map[string]bool{"reff001": true}

	tests := []struct {
		name       string
		row        []string
		wantResult *disbursementModel.BulkPreviewResponse
	}{
		{
			name: "Number of rows below limit",
			row:  []string{},
		},
		{
			name: "Empty data",
			row:  make([]string, 6),
		},
		{
			name: "Duplicate reference and empty data",
			row: []string{
				"REFF001", "", "", "", "", "",
			},
			wantResult: &disbursementModel.BulkPreviewResponse{
				ReferenceID: "REFF001",
				Error:       "Channel code is required, Account number is required, Account name is required, Amount is required, reference ID already exist",
				Result:      "INVALID",
			},
		},
		{
			name: "Bank code not found and invalid amount",
			row: []string{
				"", "A", "XXX", "XXX", "01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789", "TEST.1",
			},
			wantResult: &disbursementModel.BulkPreviewResponse{
				ChannelCode:            "XXX",
				BeneficiaryAccountNo:   "XXX",
				BeneficiaryAccountName: "01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789",
				Amount:                 "A",
				Remark:                 "TEST.1",
				Error:                  "Reference ID is required, Channel code not found, Account number must be numeric value, Decimal amounts are not allowed, Max account name 100",
				Result:                 "INVALID",
			},
		},
		{
			name: "Bank code not found and invalid decimal amount",
			row: []string{
				"", "120000.24", "XXX", "XXX", "01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789", "TEST.1",
			},
			wantResult: &disbursementModel.BulkPreviewResponse{
				ChannelCode:            "XXX",
				BeneficiaryAccountNo:   "XXX",
				BeneficiaryAccountName: "01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789",
				Amount:                 "120000.24",
				Remark:                 "TEST.1",
				Error:                  "Reference ID is required, Channel code not found, Account number must be numeric value, Decimal amounts are not allowed, Max account name 100",
				Result:                 "INVALID",
			},
		},
		{
			name: "Large amount with decimal point should be rejected (float32 precision bug regression test)",
			row: []string{
				"REF-LARGE-DECIMAL", "4918185123321.2", "BRI", "1234567890", "JOHN DOE", "Test large decimal",
			},
			wantResult: &disbursementModel.BulkPreviewResponse{
				ReferenceID:            "REF-LARGE-DECIMAL",
				ChannelCode:            "BRI",
				BeneficiaryBankCode:    "002",
				BeneficiaryBankName:    "BANK RAKYAT INDONESIA",
				BeneficiaryAccountNo:   "1234567890",
				BeneficiaryAccountName: "JOHN DOE",
				Amount:                 "4918185123321.2",
				Remark:                 "Test large decimal",
				Error:                  "Decimal amounts are not allowed",
				Result:                 "INVALID",
			},
		},
		{
			name: "Transactions less than the minimum amount",
			row: []string{
				"REFF/TRX/001", "9000", "BRI", "0001", "JOHN", "01234567890123456789xxxx",
			},
			wantResult: &disbursementModel.BulkPreviewResponse{
				ReferenceID:            "REFF/TRX/001",
				ChannelCode:            "BRI",
				BeneficiaryBankCode:    "002",
				BeneficiaryBankName:    "BANK RAKYAT INDONESIA",
				BeneficiaryAccountNo:   "0001",
				BeneficiaryAccountName: "JOHN",
				Amount:                 "9000",
				Remark:                 "01234567890123456789xxxx",
				Error:                  "Min amount is Rp 10.000",
				Result:                 "INVALID",
			},
		},
		{
			name: "SUCCESS",
			row: []string{
				"REFF/TRX/001", "75000", "BRI", "0001", "JOHN", "TEST.3",
			},
			wantResult: &disbursementModel.BulkPreviewResponse{
				ReferenceID:            "REFF/TRX/001",
				ChannelCode:            "BRI",
				BeneficiaryBankCode:    "002",
				BeneficiaryBankName:    "BANK RAKYAT INDONESIA",
				BeneficiaryAccountNo:   "0001",
				BeneficiaryAccountName: "JOHN",
				Amount:                 "75000",
				Remark:                 "TEST.3",
				Result:                 "VALID",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.wantResult, service.singleRowValidation(context.Background(), "123", trxConfig, referenceList, test.row))
		})
	}
}

func TestBeneficiaryAccountValidation(t *testing.T) {
	beneficiaryAccountSvc := serviceMocks.NewIBeneficiaryAccountService(t)
	merchantRepo := repoMocks.NewIMerchantRepository(t)
	redisClient := redisMock.NewIRedisExt(t)
	disbursementRepo := repoMocks.NewIDisbursementRepository(t)
	snapCoreRepo := repoMocks.NewISnapCoreRepository(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	// General merchant mock for all test cases
	merchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).
		Return(&merchant.Merchant{
			UUID: "123",
			Name: "Test Merchant",
		}, nil).Maybe()

	getResult := &redis.StringCmd{}
	getResult.SetErr(redis.Nil)
	redisClient.On("Get", mock.Anything, mock.Anything).Return(getResult).Maybe()

	setResult := &redis.StatusCmd{}
	redisClient.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(setResult).Maybe()

	disbursementRepo.On("GetTransactionConfig", mock.Anything, mock.Anything).
		Return(&disbursementModel.TransactionConfig{
			MinAmount: 10000,
			MaxAmount: 100000,
		}, nil).Maybe()

	redisClient.On("Result", mock.Anything).Return("", redis.Nil).Maybe()

	snapCoreRepo.On("GetBankCodeList", mock.Anything, mock.Anything).Return(nil, nil).Maybe()

	service := DisbursementService{
		beneficiaryAccountSvc: beneficiaryAccountSvc,
		merchantRepo:          merchantRepo,
		redisExt:              redisClient,
		disbursementRepo:      disbursementRepo,
		snapCoreRepo:          snapCoreRepo,
		logger:                logger,
		config: &config.Config{
			DisbursementConfig: config.DisbursementConfig{
				OverbookingBankMaxAmount:              25000000,
				OverbookingBankMaxAmountForCustomRule: 50000000,
			},
		},
	}

	trxConfig := &disbursementModel.TransactionConfig{
		MinAmount: 10000,
		MaxAmount: 100000,
	}

	tests := []struct {
		preview    *disbursementModel.BulkPreviewResponse
		setupMock  func()
		wantResult *disbursementModel.BulkPreviewResponse
	}{
		{
			preview: &disbursementModel.BulkPreviewResponse{
				Error: "reference ID already exist",
			},
			setupMock: func() {
				beneficiaryAccountSvc.On(
					"FindByBankCodeAndAccountNo", c.BackgroundCtxMockType(), mock.AnythingOfType("*beneficiaryAccountModel.CheckAccountRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantResult: &disbursementModel.BulkPreviewResponse{
				Error:  "reference ID already exist, Account number is invalid",
				Result: "INVALID",
			},
		},
		{
			preview: &disbursementModel.BulkPreviewResponse{},
			setupMock: func() {
				beneficiaryAccountSvc.On(
					"FindByBankCodeAndAccountNo", c.BackgroundCtxMockType(), mock.AnythingOfType("*beneficiaryAccountModel.CheckAccountRequest"),
				).Once().Return(nil, nil)
			},
			wantResult: &disbursementModel.BulkPreviewResponse{
				Error:  "Account number is invalid",
				Result: "INVALID",
			},
		},
		{
			preview: &disbursementModel.BulkPreviewResponse{
				BeneficiaryAccountName: "JON",
				Error:                  "Maximum amount is Rp 100.000",
			},
			setupMock: func() {
				beneficiaryAccountSvc.On(
					"FindByBankCodeAndAccountNo", c.BackgroundCtxMockType(), mock.AnythingOfType("*beneficiaryAccountModel.CheckAccountRequest"),
				).Once().Return(&beneficiaryAccountModel.Account{BeneficiaryAccountName: "JOHN"}, nil)
			},
			wantResult: &disbursementModel.BulkPreviewResponse{
				BeneficiaryAccountName: "JOHN",
				Error:                  "Maximum amount is Rp 100.000, Incorrect beneficiary name. Before : <b>JON</b>",
				Result:                 "INVALID",
			},
		},
		{
			preview: &disbursementModel.BulkPreviewResponse{
				BeneficiaryAccountName: "WIK",
			},
			setupMock: func() {
				beneficiaryAccountSvc.On(
					"FindByBankCodeAndAccountNo", c.BackgroundCtxMockType(), mock.AnythingOfType("*beneficiaryAccountModel.CheckAccountRequest"),
				).Once().Return(&beneficiaryAccountModel.Account{BeneficiaryAccountName: "WICK"}, nil)
			},
			wantResult: &disbursementModel.BulkPreviewResponse{
				BeneficiaryAccountName: "WICK",
				Error:                  "Incorrect beneficiary name. Before : <b>WIK</b>",
				Result:                 "WARNING",
			},
		},
		{
			preview: &disbursementModel.BulkPreviewResponse{
				BeneficiaryBankCode:    "002",
				BeneficiaryAccountNo:   "123450000001",
				BeneficiaryAccountName: "TEST",
			},
			setupMock: func() {
				beneficiaryAccountSvc.On(
					"FindByBankCodeAndAccountNo", mock.Anything, mock.Anything,
				).Once().Return(&beneficiaryAccountModel.Account{
					BeneficiaryBankCode:    "002",
					BeneficiaryAccountNo:   "123450000001",
					BeneficiaryAccountName: "TEST 1",
					MetadataObj: beneficiaryAccountModel.Metadata{
						IsVirtualAccount: true,
					},
				}, nil)
			},
			wantResult: &disbursementModel.BulkPreviewResponse{
				BeneficiaryBankCode:    "002",
				BeneficiaryAccountNo:   "123450000001",
				BeneficiaryAccountName: "TEST 1",
				Result:                 "INVALID",
				Error:                  "Incorrect beneficiary name. Before : <b>TEST</b>, Destination account is not eligible for payout",
			},
		},
		{
			preview: &disbursementModel.BulkPreviewResponse{
				BeneficiaryAccountName: "HENDRU",
				Result:                 "VALID",
			},
			setupMock: func() {
				beneficiaryAccountSvc.On(
					"FindByBankCodeAndAccountNo", c.BackgroundCtxMockType(), mock.AnythingOfType("*beneficiaryAccountModel.CheckAccountRequest"),
				).Return(&beneficiaryAccountModel.Account{BeneficiaryAccountName: "HENDRU"}, nil)
			},
			wantResult: &disbursementModel.BulkPreviewResponse{
				BeneficiaryAccountName: "HENDRU",
				Result:                 "VALID",
			},
		},
		{
			preview: &disbursementModel.BulkPreviewResponse{
				BeneficiaryAccountName: "",
				Result:                 "VALID",
			},
			setupMock: func() {
				beneficiaryAccountSvc.On(
					"FindByBankCodeAndAccountNo",
					c.BackgroundCtxMockType(),
					mock.AnythingOfType("*beneficiaryAccountModel.CheckAccountRequest"),
				).Return(&beneficiaryAccountModel.Account{BeneficiaryAccountName: ""}, nil)
			},
			wantResult: &disbursementModel.BulkPreviewResponse{
				BeneficiaryAccountName: "",
				Result:                 "INVALID",
				Error:                  "Account name is required",
			},
		},
		{
			preview: &disbursementModel.BulkPreviewResponse{
				BeneficiaryAccountName: "",
				Result:                 "VALID",
				Error:                  "Previous error",
			},
			setupMock: func() {
				beneficiaryAccountSvc.On(
					"FindByBankCodeAndAccountNo",
					c.BackgroundCtxMockType(),
					mock.AnythingOfType("*beneficiaryAccountModel.CheckAccountRequest"),
				).Return(&beneficiaryAccountModel.Account{BeneficiaryAccountName: ""}, nil)
			},
			wantResult: &disbursementModel.BulkPreviewResponse{
				BeneficiaryAccountName: "",
				Result:                 "INVALID",
				Error:                  "Account name is required",
			},
		},
	}
	for _, test := range tests {
		test.setupMock()

		assert.Equal(t, test.wantResult, service.beneficiaryAccountValidation(
			context.Background(), "123", trxConfig, test.preview,
		))
	}
}

func TestValidateAmountAndLimits(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	redisClient := redisMock.NewIRedisExt(t)
	snapCoreRepo := repoMocks.NewISnapCoreRepository(t)

	// Mock Redis Get for overbooking check
	getResult := &redis.StringCmd{}
	getResult.SetErr(redis.Nil) // Cache miss
	redisClient.On("Get", mock.Anything, mock.Anything).Return(getResult).Maybe()

	// Mock Redis Set for caching overbooking bank list
	setResult := &redis.StatusCmd{}
	redisClient.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(setResult).Maybe()

	trxConfig := &disbursementModel.TransactionConfig{
		MinAmount: 10000,
		MaxAmount: 100000,
	}

	tests := []struct {
		name               string
		merchantID         string
		previewResponse    *disbursementModel.BulkPreviewResponse
		beneficiaryAccount *beneficiaryAccountModel.Account
		setupMocks         func()
		wantErr            bool
		wantErrMsg         string
	}{
		{
			name:       "SUCCESS: Valid amount within limits",
			merchantID: "MERCHANT001",
			previewResponse: &disbursementModel.BulkPreviewResponse{
				Amount:              "50000",
				BeneficiaryBankCode: "BRI",
			},
			beneficiaryAccount: &beneficiaryAccountModel.Account{
				BeneficiaryAccountName: "JOHN DOE",
				MetadataObj: beneficiaryAccountModel.Metadata{
					BeneficiaryPayoutLimitRule: nil,
				},
			},
			setupMocks: func() {
				snapCoreRepo.On("GetBankCodeList", mock.Anything, mock.Anything).Return(nil, nil).Maybe()
			},
			wantErr: false,
		},
		{
			name:       "ERROR: Amount exceeds max limit",
			merchantID: "MERCHANT001",
			previewResponse: &disbursementModel.BulkPreviewResponse{
				Amount:              "150000",
				BeneficiaryBankCode: "BRI",
			},
			beneficiaryAccount: &beneficiaryAccountModel.Account{
				BeneficiaryAccountName: "JOHN DOE",
				MetadataObj: beneficiaryAccountModel.Metadata{
					BeneficiaryPayoutLimitRule: nil,
				},
			},
			setupMocks: func() {
				snapCoreRepo.On("GetBankCodeList", mock.Anything, mock.Anything).Return(nil, nil).Maybe()
			},
			wantErr:    true,
			wantErrMsg: "Maximum amount is Rp 100.000",
		},
		{
			name:       "ERROR: Overbooking amount exceeds overbooking limit",
			merchantID: "MERCHANT001",
			previewResponse: &disbursementModel.BulkPreviewResponse{
				Amount:              "50000000",
				BeneficiaryBankCode: "BRI",
			},
			beneficiaryAccount: &beneficiaryAccountModel.Account{
				BeneficiaryAccountName: "JOHN DOE",
				MetadataObj: beneficiaryAccountModel.Metadata{
					BeneficiaryPayoutLimitRule: nil,
				},
			},
			setupMocks: func() {
				snapCoreRepo.On("GetBankCodeList", mock.Anything, mock.Anything).Return(nil, nil).Maybe()
			},
			wantErr:    true,
			wantErrMsg: "Maximum amount is Rp",
		},
		{
			name:       "SUCCESS: Valid amount at exact max limit",
			merchantID: "MERCHANT001",
			previewResponse: &disbursementModel.BulkPreviewResponse{
				Amount:              "100000",
				BeneficiaryBankCode: "BRI",
			},
			beneficiaryAccount: &beneficiaryAccountModel.Account{
				BeneficiaryAccountName: "JOHN DOE",
				MetadataObj: beneficiaryAccountModel.Metadata{
					BeneficiaryPayoutLimitRule: nil,
				},
			},
			setupMocks: func() {
				snapCoreRepo.On("GetBankCodeList", mock.Anything, mock.Anything).Return(nil, nil).Maybe()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			service := &DisbursementService{
				logger:       logger,
				redisExt:     redisClient,
				snapCoreRepo: snapCoreRepo,
				config: &config.Config{
					DisbursementConfig: config.DisbursementConfig{
						OverbookingBankMaxAmount:              25000000,
						OverbookingBankMaxAmountForCustomRule: 50000000,
					},
				},
			}

			err := service.validateAmountAndLimits(
				context.Background(),
				tt.merchantID,
				tt.previewResponse,
				tt.beneficiaryAccount,
				trxConfig,
			)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
