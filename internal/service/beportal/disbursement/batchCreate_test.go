package disbursementService

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/notification"
	beneficiaryAccountRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/beneficiaryAccount"
	disbursementRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/disbursement"
	feeRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/fee"
	merchantRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/merchant"
	statusHistoriesRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/statusHistories"
	beneficiaryAccountService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/beneficiaryAccount"
	feeService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/fee"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	redisMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/test"
	"github.com/paper-indonesia/pivot-backoffice/test/schemas"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/amqp"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBatchCreateDisbursement(t *testing.T) {
	redismock := redisMock.NewIRedisExt(t)

	// Make all Redis operations optional with .Maybe()
	redismock.On(
		"Del", c.ValueCtxMockType(), c.StringMockType(),
	).Return(&redis.IntCmd{}).Maybe()

	redismock.On(
		"SetNX", c.ValueCtxMockType(), c.StringMockType(), true, mock.AnythingOfType("time.Duration"),
	).Return(redis.NewBoolResult(true, nil)).Maybe()

	redismock.On(
		"Set", c.ValueCtxMockType(), mock.AnythingOfType("string"), true, mock.AnythingOfType("time.Duration"),
	).Return(redis.NewStatusResult("OK", nil)).Maybe()

	validData := make([]disbursementModel.CreateSingleRequest, 0, 3)
	for i := 0; i < 3; i++ {
		validData = append(validData, disbursementModel.CreateSingleRequest{
			ReferenceID:            fmt.Sprintf("client-ref %d", i),
			BeneficiaryBankCode:    "013",
			BeneficiaryBankName:    "Bank Permata",
			BeneficiaryAccountNo:   "123412341234",
			BeneficiaryAccountName: "John Doe",
			Amount:                 decimal.NewFromInt(10000),
			Remark:                 "this is remark",
		})
	}

	validRequest := &disbursementModel.BatchCreateDisbursementRequest{
		BulkID:       uuid.NewString(),
		MerchantID:   uuid.NewString(),
		MerchantName: "test",
		CreatedBy:    uuid.NewString(),
		CreatedFrom:  c.DisbursementCreatedFromMerchantPortal,
		TotalTrx:     len(validData),
		Data:         validData,
	}

	validAutoApproveRequest := &disbursementModel.BatchCreateDisbursementRequest{
		BulkID:       uuid.NewString(),
		MerchantID:   uuid.NewString(),
		MerchantName: "test",
		CreatedBy:    uuid.NewString(),
		CreatedFrom:  c.DisbursementCreatedFromMerchantPortal,
		TotalTrx:     len(validData),
		Data:         validData,
		AutoApprove:  true,
	}

	feeDecimal := 1000.0
	disbursementIntSvc := serviceMocks.NewIDisbursementInternalService(t)

	merchant := &merchantModel.Merchant{}

	testCases := []struct {
		name       string
		wantErr    bool
		mocksSetup func(
			mockRepo *repositoryMocks.IDisbursementRepository,
			rmqExt *mockRabbitMq.RabbitMQExt,
			feeSvc *serviceMocks.IFeeService,
			merchantRepo *repositoryMocks.IMerchantRepository,
		)
		input *disbursementModel.BatchCreateDisbursementRequest
	}{
		{
			name:    "SUCCESS",
			wantErr: false,
			mocksSetup: func(mockRepo *repositoryMocks.IDisbursementRepository, rmqExt *mockRabbitMq.RabbitMQExt, feeSvc *serviceMocks.IFeeService, merchantRepo *repositoryMocks.IMerchantRepository) {
				merchantRepo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).
					Return(merchant, nil)

				mockRepo.On("CountByBulkID", mock.Anything, c.StringMockType()).Return(len(validData))

				mockRepo.On(
					"Insert", mock.Anything, DisbursementMockType,
				).Return(nil)

				mockRepo.On(
					"UpdateBulkDisbursementStatusByID", mock.Anything, c.StringMockType(), c.StringMockType(),
				).Return(nil)
				rmqExt.On(
					"PushNotification", mock.Anything, NotificationMockType,
				).Return(nil)

				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).
					Return(feeDecimal, &feeModel.FeeMetadataObject{}, nil)
			},
			input: validRequest,
		},
		{
			name:    "SUCCESS: With auto approve",
			wantErr: false,
			mocksSetup: func(mockRepo *repositoryMocks.IDisbursementRepository, rmqExt *mockRabbitMq.RabbitMQExt, feeSvc *serviceMocks.IFeeService, merchantRepo *repositoryMocks.IMerchantRepository) {
				merchantRepo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).
					Return(merchant, nil)

				mockRepo.On(
					"CountByBulkID", mock.Anything, c.StringMockType(),
				).Return(len(validData))

				mockRepo.On(
					"Insert", mock.Anything, DisbursementMockType,
				).Return(nil)

				mockRepo.On(
					"UpdateBulkDisbursementStatusByID", mock.Anything, c.StringMockType(), c.StringMockType(),
				).Return(nil)

				mockRepo.On(
					"GetAllDisbursementByBulkID", mock.Anything, c.StringMockType(),
				).Return([]*disbursementModel.DisbursementWithTransaction{
					{
						Disbursement: disbursementModel.Disbursement{UUID: uuid.NewString(), Fee: util.ValueToPtr(decimal.NewFromFloat(feeDecimal))},
					},
				}, nil)

				rmqExt.On("PushNotification", mock.Anything, NotificationMockType).Return(nil)
				disbursementIntSvc.On("Approve", c.ValueCtxMockType(), mock.Anything).Once().Return(nil)
				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).Return(feeDecimal, &feeModel.FeeMetadataObject{}, nil)
			},
			input: validAutoApproveRequest,
		},
		{
			name:    "SUCCESS: But got error in CreateSingle service",
			wantErr: false,
			mocksSetup: func(mockRepo *repositoryMocks.IDisbursementRepository, rmqExt *mockRabbitMq.RabbitMQExt, feeSvc *serviceMocks.IFeeService, merchantRepo *repositoryMocks.IMerchantRepository) {
				merchantRepo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).
					Return(merchant, nil)

				mockRepo.On(
					"CountByBulkID", mock.Anything, c.StringMockType(),
				).Return(len(validData))

				mockRepo.On(
					"Insert", mock.Anything, DisbursementMockType,
				).Return(c.ErrSomeErrorForUnitTest)

				mockRepo.On(
					"UpdateBulkDisbursementStatusByID", mock.Anything, c.StringMockType(), c.StringMockType(),
				).Return(nil)

				rmqExt.On(
					"PushNotification", mock.Anything, NotificationMockType,
				).Return(nil)

				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).
					Return(feeDecimal, &feeModel.FeeMetadataObject{}, nil)
			},
			input: validRequest,
		},
		{
			name:    "SUCCESS: With skip UpdateBulkDisbursementStatusByID",
			wantErr: false,
			mocksSetup: func(mockRepo *repositoryMocks.IDisbursementRepository, rmqExt *mockRabbitMq.RabbitMQExt, feeSvc *serviceMocks.IFeeService, merchantRepo *repositoryMocks.IMerchantRepository) {
				merchantRepo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).
					Return(merchant, nil)

				mockRepo.On(
					"CountByBulkID", mock.Anything, c.StringMockType(),
				).Return(0)

				mockRepo.On(
					"Insert", mock.Anything, DisbursementMockType,
				).Return(nil)

				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).
					Return(feeDecimal, &feeModel.FeeMetadataObject{}, nil)
			},
			input: validRequest,
		},
		{
			name:    "ERROR: UpdateBulkDisbursementStatusByID",
			wantErr: true,
			mocksSetup: func(mockRepo *repositoryMocks.IDisbursementRepository, rmqExt *mockRabbitMq.RabbitMQExt, feeSvc *serviceMocks.IFeeService, merchantRepo *repositoryMocks.IMerchantRepository) {
				merchantRepo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).
					Return(merchant, nil)

				mockRepo.On(
					"CountByBulkID",
					mock.Anything,
					c.StringMockType(),
				).Return(len(validData))

				mockRepo.On(
					"Insert",
					mock.Anything,
					DisbursementMockType,
				).Return(nil)

				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).
					Return(feeDecimal, &feeModel.FeeMetadataObject{}, nil)

				mockRepo.On(
					"UpdateBulkDisbursementStatusByID",
					mock.Anything,
					c.StringMockType(),
					c.StringMockType(),
				).Return(c.ErrSomeErrorForUnitTest)
			},
			input: validRequest,
		},
		{
			name:    "ERROR: With auto approve and got error GetAllDisbursementByBulkID",
			wantErr: true,
			mocksSetup: func(mockRepo *repositoryMocks.IDisbursementRepository, rmqExt *mockRabbitMq.RabbitMQExt, feeSvc *serviceMocks.IFeeService, merchantRepo *repositoryMocks.IMerchantRepository) {
				merchantRepo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).
					Return(merchant, nil)

				mockRepo.On(
					"CountByBulkID", mock.Anything, c.StringMockType(),
				).Return(len(validData))

				mockRepo.On(
					"Insert", mock.Anything, DisbursementMockType,
				).Return(nil)

				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).
					Return(feeDecimal, &feeModel.FeeMetadataObject{}, nil)

				mockRepo.On(
					"UpdateBulkDisbursementStatusByID", mock.Anything, c.StringMockType(), c.StringMockType(),
				).Return(nil)

				mockRepo.On(
					"GetAllDisbursementByBulkID", mock.Anything, c.StringMockType(),
				).Return([]*disbursementModel.DisbursementWithTransaction{}, c.ErrSomeErrorForUnitTest)

				rmqExt.On(
					"PushNotification", mock.Anything, NotificationMockType,
				).Return(errors.New("exchange name not found"))
			},
			input: validAutoApproveRequest,
		},
		{
			name:    "ERROR: With auto approve and got error service",
			wantErr: true,
			mocksSetup: func(mockRepo *repositoryMocks.IDisbursementRepository, rmqExt *mockRabbitMq.RabbitMQExt, feeSvc *serviceMocks.IFeeService, merchantRepo *repositoryMocks.IMerchantRepository) {
				merchantRepo.On("FindMerchantByID", c.ValueCtxMockType(), c.StringMockType()).
					Return(merchant, nil)

				mockRepo.On(
					"CountByBulkID", mock.Anything, c.StringMockType(),
				).Return(len(validData))

				mockRepo.On(
					"Insert", mock.Anything, DisbursementMockType,
				).Return(nil)

				feeSvc.On("GetFeeCalculationAndDetail", c.ValueCtxMockType(), c.PtrGetFeeRequestMockType()).
					Return(feeDecimal, &feeModel.FeeMetadataObject{}, nil)

				mockRepo.On(
					"UpdateBulkDisbursementStatusByID", mock.Anything, c.StringMockType(), c.StringMockType(),
				).Return(nil)

				mockRepo.On(
					"GetAllDisbursementByBulkID", mock.Anything, c.StringMockType(),
				).Return([]*disbursementModel.DisbursementWithTransaction{
					{
						Disbursement: disbursementModel.Disbursement{UUID: uuid.NewString()},
					},
				}, nil)

				rmqExt.On("PushNotification", mock.Anything, NotificationMockType).Return(nil)
				disbursementIntSvc.On("Approve", c.ValueCtxMockType(), mock.Anything).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			input: validAutoApproveRequest,
		},
		{
			name:    "ERROR:Get merchant by id",
			wantErr: true,
			mocksSetup: func(_ *repositoryMocks.IDisbursementRepository, _ *mockRabbitMq.RabbitMQExt, _ *serviceMocks.IFeeService, merchantRepo *repositoryMocks.IMerchantRepository) {
				merchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			input: validAutoApproveRequest,
		},
		{
			name:    "ERROR:Merchant not found",
			wantErr: true,
			mocksSetup: func(_ *repositoryMocks.IDisbursementRepository, _ *mockRabbitMq.RabbitMQExt, _ *serviceMocks.IFeeService, merchantRepo *repositoryMocks.IMerchantRepository) {
				merchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
			input: validAutoApproveRequest,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			feeSvc := serviceMocks.NewIFeeService(t)
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)
			disbursementRepoMock := repositoryMocks.NewIDisbursementRepository(t)
			statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
			merchantRepo := repositoryMocks.NewIMerchantRepository(t)
			beneficiaryAccountSvc := serviceMocks.NewIBeneficiaryAccountService(t)

			conf := config.Config{
				Environment: c.EnvironmentStaging,
			}

			// General status history mock that will handle any calls
			statusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Maybe()

			// General beneficiary account mock (returns nil = not a VA)
			beneficiaryAccountSvc.On("FindByBankCodeAndAccountNo", c.ValueCtxMockType(), CheckAccountReqMockType).
				Return(nil, nil).Maybe()

			tc.mocksSetup(disbursementRepoMock, mockRmq, feeSvc, merchantRepo)

			svc := New(
				&conf, pdkLoggerMock, merchantRepo, disbursementRepoMock, nil, nil,
				WithStatusHistoriesRepository(statusHistoriesRepo), WithFeeService(feeSvc), WithBeneficiaryAccService(beneficiaryAccountSvc), WithRedisClient(redismock), WithRabbitMQClient(mockRmq), WithDisbursementInternalService(disbursementIntSvc),
			)

			if err := svc.BatchCreateDisbursement(context.Background(), tc.input); tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			feeSvc.AssertExpectations(t)
			mockRmq.AssertExpectations(t)
			disbursementRepoMock.AssertExpectations(t)
		})
	}
}

func TestIntegrationBatchCreateDisbursement(t *testing.T) {
	if os.Getenv(c.IntegrationTestEnv) != "1" {
		t.Skip(c.SkipIntegrationTest)
	}

	ctx := context.Background()

	cfg := &config.Config{
		Environment: c.EnvironmentStaging,
		DisbursementConfig: config.DisbursementConfig{
			DailyLimitMerchant: 10_000_000,
		},
	}

	bulkID := uuid.NewString()
	merchantID := "922e39ab-7565-49f6-b84f-fb56122821ae"
	createdBy := "7b73232a-afa6-4f63-bd21-52d1d0f8619b"
	request := &disbursementModel.BatchCreateDisbursementRequest{
		BulkID:       bulkID,
		MerchantID:   merchantID,
		MerchantName: "TerterPay",
		CreatedBy:    createdBy,
		CreatedFrom:  "Integration-Test",
		TotalTrx:     1,
		AutoApprove:  false,
		Data: []disbursementModel.CreateSingleRequest{
			{
				ReferenceID:            "REF001",
				BeneficiaryBankCode:    "002",
				BeneficiaryBankName:    "BANK RAKYAT INDONESIA",
				BeneficiaryAccountNo:   "999966660001",
				BeneficiaryAccountName: "Dummy Simulation",
				Amount:                 decimal.NewFromInt(12_500),
				BulkID:                 &bulkID,
				MerchantID:             merchantID,
				MerchantName:           "TerterPay",
				CreatedBy:              &createdBy,
				CreatedFrom:            "Integration-Test",
			},
		},
	}

	schemaMigrations := []string{
		schemas.MerchantFee(), schemas.Disbursement(), schemas.BulkDisbursement(), schemas.Merchant(), schemas.BeneficiaryAccount(),
		schemas.StatusHistory(), schemas.District(), schemas.City(), schemas.Province(), schemas.User(),
	}
	for _, raw := range schemaMigrations {
		_, err := db.ExecContext(ctx, raw)
		require.NoErrorf(t, err, "run.schema-migration")
	}

	_, err := db.ExecContext(
		ctx,
		`INSERT INTO bulk_disbursements(uuid, merchant_id, file, status, created_by, created_at, updated_at) VALUES(?, ?, 'none', 'WAITING', ?, NOW(), NOW());`,
		bulkID, merchantID, createdBy,
	)
	require.NoErrorf(t, err, "insert.bulk_disbursements")

	_, err = db.ExecContext(
		ctx,
		`INSERT INTO merchant_fees (uuid, merchant_id, amount, amount_type, percentage, reference, deduction_type, tax_type, tax_percentage, created_at, updated_at) VALUES
		('b5845993-454b-4e98-a26c-358843256ed8', '922e39ab-7565-49f6-b84f-fb56122821ae', 3500.00, 'AMOUNT', 0, 'DISBURSEMENT', 'DIRECT', 'EXCLUSIVE', 0, NOW(), NOW());`,
	)
	require.NoErrorf(t, err, "insert.merchant_fees")

	_, err = db.ExecContext(
		ctx,
		`INSERT INTO merchants(uuid, name, short_name, external_id, address, district_id, postcode, logo, merchant_email, merchant_phone, pic_email, pic_phone, created_at, description, updated_at) 
			VALUES('922e39ab-7565-49f6-b84f-fb56122821ae', 'Testing', 'TT', '', '', 0, 0, '', '', '', '', '', NOW(), '', NOW());`,
	)
	require.NoErrorf(t, err, "insert.merchants")

	_, err = db.ExecContext(
		ctx,
		`INSERT INTO beneficiary_accounts(uuid, merchant_id, beneficiary_bank_code, beneficiary_bank_name, beneficiary_account_no, beneficiary_account_name, metadata, created_at, updated_at, deleted_at)
			VALUES('ff5241b3-5b39-4649-856c-538aaef0de29', '922e39ab-7565-49f6-b84f-fb56122821ae', '002', 'BANK RAKYAT INDONESIA', '999966660001', 'Dummy Simulation', '{"isXb": false, "maxAmount": "0", "isOverbooking": false, "requestInquiryStatus": "VALID"}', NOW(), NOW(), NULL);`,
	)
	require.NoErrorf(t, err, "insert.beneficiary_accounts")

	conn, err := test.OpenConnectionRabbitMQ(ctx, rmqContainer)
	require.NoErrorf(t, err, "amqp.DialConfig")
	defer conn.Close()

	ch, err := conn.Channel()
	require.NoErrorf(t, err, "conn.Channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		rabbitMqExt.NotificationExchange, amqp.ExchangeDirect, true, false, false, false, amqp.Table{"alternate-exchange": rabbitMqExt.UnroutedNotificationExchange},
	)
	require.NoError(t, err)

	q, err := ch.QueueDeclare("", false, true, false, false, nil)
	require.NoError(t, err)
	defer ch.QueueDelete(q.Name, false, false, false)

	err = ch.QueueBind(q.Name, fmt.Sprintf(c.NotificationRoutingKeyFmt, createdBy), rabbitMqExt.NotificationExchange, false, nil)
	require.NoError(t, err)

	consumer, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	require.NoErrorf(t, err, "ch.Consume")

	feeRepo := feeRepository.New(db, pdkLoggerMock)
	merchantRepo := merchantRepository.New(db, pdkLoggerMock)
	beneficiaryAccountRepo := beneficiaryAccountRepository.New(db, pdkLoggerMock)
	statusHistoryRepo := statusHistoriesRepository.New(db)
	feeSvc := feeService.New(pdkLoggerMock, feeRepo, merchantRepo, feeService.WithRedisClient(redisClient))
	beneficiaryAccountSvc := beneficiaryAccountService.New(pdkLoggerMock, beneficiaryAccountRepo, nil, nil)
	service := New(
		cfg, pdkLoggerMock, merchantRepository.New(db, pdkLoggerMock), disbursementRepository.New(db, pdkLoggerMock), nil, nil,
		WithRabbitMQClient(publisher), WithFeeService(feeSvc), WithRedisClient(redisClient), WithBeneficiaryAccService(beneficiaryAccountSvc), WithStatusHistoriesRepository(statusHistoryRepo),
	)
	require.NoErrorf(t, service.BatchCreateDisbursement(ctx, request), "service.BatchCreateDisbursement")

	payload := notification.PushNotificationPayload{}
	select {
	case <-time.After(10 * time.Second):
	case d := <-consumer.Delivery:
		require.NoErrorf(t, json.Unmarshal(d.Body, &payload), "json.Unmarshal")
	}
	assert.Equal(t, fmt.Sprintf("Your batch transaction <b>%s</b> has been successfully uploaded.", bulkID[len(bulkID)-12:]), payload.Message)
}
