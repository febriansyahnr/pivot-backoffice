package internalPayoutController_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	requestAccountInquiries "github.com/paper-indonesia/pivot-backoffice/internal/model/requestAccountInquiry"
	mockMQ "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	redisMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	internalPayoutController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/payout"

	chi "github.com/go-chi/chi/v5"
	validator "github.com/go-playground/validator/v10"
	redismock "github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	invalidReferenceID           = `{"payouts":[{"channelCode":"C1","channelInformation":{"accountNumber":"123456","accountName":"NAME"},"amount":{"currency":"IDR","value":"10000"},"description":"DESC"}]}`
	invalidAccountName           = `{"payouts":[{"referenceId":"R1","channelCode":"C1","channelInformation":{"accountNumber":"123456","accountName":"Alexander Maximillian Theodore Valentino Sebastian Aurelio Christopher Emmanuela Montgomery Wellington"},"amount":{"currency":"IDR","value":"10000"},"description":"D1"}]}`
	invalidAmount                = `{"payouts":[{"referenceId":"R1","channelCode":"C1","channelInformation":{"accountNumber":"123456","accountName":"NAME"},"amount":{"currency":"IDR","value":"INVALID"},"description":"D1"}]}`
	invalidAmountWithDecimal     = `{"payouts":[{"referenceId":"R1","channelCode":"C1","channelInformation":{"accountNumber":"123456","accountName":"NAME"},"amount":{"currency":"IDR","value":"12000.22"},"description":"D1"}]}`
	invalidAmountIsBelowMinLimit = `{"payouts":[{"referenceId":"R1","channelCode":"C1","channelInformation":{"accountNumber":"123456","accountName":"NAME"},"amount":{"currency":"IDR","value":"1000"},"description":"D1"}]}`
	invalidJSONFormat            = `{"payouts": [{invalid json}]}`
	ValidReqBody                 = `{"payouts":[{"referenceId":"R1","inquiryId":"ABC","channelCode":"BRI","channelInformation":{"accountNumber":"123456","accountName":"NAME"},"amount":{"currency":"IDR","value":"11000"},"description":"D1"}]}`
	ValidOverbookingReqBody      = `{"payouts":[{"referenceId":"R1","inquiryId":"ABC","channelCode":"MANDIRI","channelInformation":{"accountNumber":"123456","accountName":"NAME"},"amount":{"currency":"IDR","value":"400000"},"description":"D1"}]}`
)

func TestCreate(pt *testing.T) {
	db, clientMock := redismock.NewClientMock()
	redisMock := redisExt.WrapRedisClient(db, nil)

	mockRabbitMq := mockMQ.NewRabbitMQExt(pt)
	mockAccountInquirySvc := serviceMocks.NewIAccountInquiryService(pt)
	mockMerchant := serviceMocks.NewIMerchantService(pt)
	mockDisbursement := serviceMocks.NewIDisbursementService(pt)
	mockTrxCloser := &serviceMocks.ITransactionCloser{}
	mockTrxCloser.On(
		"Close", mock.Anything, mock.Anything,
	).Return(nil)

	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	defer buf.Reset()

	logger := logger.NewSlogger(logger.Config{}, logger.WithSlogOutput(buf))

	conf := config.Config{
		Environment: "development",
		DisbursementConfig: config.DisbursementConfig{
			CutOffTimeWindow: config.DisbursementCutOffTimeWindow{
				TransactionInfo: "The transactions approved bla bla bla",
			},
			MaxAmount:                100_000,
			OverbookingBankMaxAmount: 500_000,
		},
		WorkerPoolConfig: config.WorkerPoolConfig{
			Disbursement: 10,
		},
	}
	merchantId := "3916e602-4e9d-4bc7-b895-f31bb81c6944"
	subMerchantId := "c05cedbb-2311-4473-a753-6a23c5b2e591"
	redisSeq, redisDefer := []func(){}, []func(){}
	queueKey := fmt.Sprintf(constant.BulkDisbursementQueueLockFmt, merchantId, "R1")

	router := chi.NewRouter()
	router.Post(
		"/payouts", internalPayoutController.New(
			&conf, validator.New(), mockDisbursement, mockMerchant, mockAccountInquirySvc, mockRabbitMq,
			internalPayoutController.WithLogger(logger), internalPayoutController.WithRedisClient(redisMock),
		).Create,
	)

	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: merchantId,
		}))
		req.Header.Set("Content-Type", "application/json")
	}

	trxConfig := &disbursementModel.TransactionConfig{
		MinAmount: 10_000, MaxAmount: 110_000,
	}

	mockTrxCloser.On(
		"Close", mock.MatchedBy(func(ctx context.Context) bool { return true }), mock.AnythingOfType("bool"),
	).Return(nil)

	tests := []struct {
		name           string
		reqBody        string
		reqSetting     func(req *http.Request)
		mockSetup      func(merchant *serviceMocks.IMerchantService, disbursement *serviceMocks.IDisbursementService, mq *mockMQ.RabbitMQExt)
		setHeaders     func(req *http.Request)
		wantStatusCode int
		wantRespBody   string
		skip           bool
	}{
		{
			name:           "ERROR:Invalid merchant info ",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   wrapErrOpenApiNonSnap(41, "merchant not found", "ERROR_UNAUTHORIZED"),
		},
		{
			name: "ERROR:Merchant ID not found",
			reqSetting: func(req *http.Request) {
				validRequestID(req)
				req.Header.Set(constant.HeaderXSubMerchantID, subMerchantId)
			},
			mockSetup: func(merchant *serviceMocks.IMerchantService, _ *serviceMocks.IDisbursementService, _ *mockMQ.RabbitMQExt) {
				merchant.On("FindMerchantByID", mock.Anything, subMerchantId).Once().Return(nil, constant.ErrMerchantNotFound)
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   wrapErrOpenApiNonSnap(40, constant.ErrMerchantNotFound.Error()),
		},
		{
			name: "ERROR:Invalid request body",
			reqSetting: func(req *http.Request) {
				validRequestID(req)
				req.Header.Set(constant.HeaderXSubMerchantID, subMerchantId)
			},
			reqBody: "{1}",
			mockSetup: func(merchant *serviceMocks.IMerchantService, _ *serviceMocks.IDisbursementService, _ *mockMQ.RabbitMQExt) {
				merchant.On(
					"FindMerchantByID", mock.Anything, subMerchantId,
				).Once().Return(&merchantModel.Merchant{
					UUID:      subMerchantId,
					KYCStatus: sql.NullString{String: constant.KYCStatusApproved},
				}, nil)
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody: `
				{
				"code":"format_invalid",
				"message":"Payouts format is invalid",
				"error":{
				  "type":"API_ERROR",
				  "details":[{
					"field":"",
					"message":"Make sure Payout request format is correct"
				  }],
				  "traceId":""
				}
		    }
 			`,
		},
		{
			name: "ERROR:Invalid JSON format",
			reqSetting: func(req *http.Request) {
				*req = *req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
					MerchantId: subMerchantId,
				}))
			},
			reqBody: invalidJSONFormat,
			mockSetup: func(merchant *serviceMocks.IMerchantService, _ *serviceMocks.IDisbursementService, _ *mockMQ.RabbitMQExt) {
				merchant.On(
					"FindMerchantByID", mock.Anything, subMerchantId,
				).Once().Return(&merchantModel.Merchant{
					UUID:      subMerchantId,
					ParentID:  sql.NullString{String: merchantId},
					KYCStatus: sql.NullString{String: constant.KYCStatusNotRequired},
				}, nil)
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody: `{
				"code":"format_invalid",
				"message":"Payouts format is invalid",
				"error":{
				  "type":"API_ERROR",
				  "details":[{
					"field":"",
					"message":"Make sure Payout request format is correct"
				  }],
				  "traceId":""
				}
			  }`,
		},
		{
			name:       "ERROR:Empty payouts data",
			reqSetting: validRequestID,
			reqBody:    `{"payouts": []}`,
			mockSetup: func(merchant *serviceMocks.IMerchantService, _ *serviceMocks.IDisbursementService, _ *mockMQ.RabbitMQExt) {
				merchant.On("FindMerchantByID", mock.Anything, merchantId).Return(&merchantModel.Merchant{UUID: merchantId}, nil)
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   wrapErrOpenApiNonSnap(40, "Key: 'CreateDisbursementFromOpenApiRequest.Payouts' Error:Field validation for 'Payouts' failed on the 'min' tag"),
		},
		{
			name:       "ERROR:Get transaction config",
			reqSetting: validRequestID,
			reqBody:    ValidReqBody,
			mockSetup: func(_ *serviceMocks.IMerchantService, disbursement *serviceMocks.IDisbursementService, _ *mockMQ.RabbitMQExt) {
				disbursement.On(
					"GetTransactionConfig", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   wrapErrOpenApiNonSnap(99, "some error", response.ErrTypeUnknown),
		},
		{
			name:       "ERROR:Invalid reference ID",
			reqSetting: validRequestID,
			reqBody:    invalidReferenceID,
			mockSetup: func(_ *serviceMocks.IMerchantService, disbursement *serviceMocks.IDisbursementService, _ *mockMQ.RabbitMQExt) {
				disbursement.On(
					"GetTransactionConfig", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Return(trxConfig, nil)
				disbursement.On("IsBankcodeOverbookingChannelAllowed", mock.AnythingOfType(constant.MockTypeValueContextReference), "002", constant.StringMockType()).Return(false)
				disbursement.On("IsMerchantAllowedExcludeBeneficiaryRules", mock.AnythingOfType(constant.MockTypeValueContextReference), merchantId, 0.0).Return(0.0, false)
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   wrapErrOpenApiNonSnap(40, `Key: 'PayoutObjectForCreate.ReferenceID' Error:Field validation for 'ReferenceID' failed on the 'required' tag`),
		},
		{
			name:           "ERROR:Account name exceeds maximum limit",
			reqSetting:     validRequestID,
			reqBody:        invalidAccountName,
			wantStatusCode: http.StatusBadRequest,
			mockSetup: func(merchant *serviceMocks.IMerchantService, disbursement *serviceMocks.IDisbursementService, _ *mockMQ.RabbitMQExt) {
				merchant.On(
					"FindMerchantByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Return(&merchantModel.Merchant{}, nil)
			},
			wantRespBody: wrapErrOpenApiNonSnap(40, constant.ErrBeneficiaryNameLengthExceeded.Error()),
		},
		{
			name:           "ERROR:Invalid transaction amount",
			reqSetting:     validRequestID,
			reqBody:        invalidAmount,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   wrapErrOpenApiNonSnap(40, constant.ErrInvalidDisbursementAmount.Error()),
		},
		{
			name:           "ERROR:Invalid transaction amount of decimal point",
			reqSetting:     validRequestID,
			reqBody:        invalidAmountWithDecimal,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   wrapErrOpenApiNonSnap(40, constant.ErrInvalidDisbursementAmount.Error()),
		},
		{
			name:           "ERROR:Transaction amount is below the minimum limit",
			reqSetting:     validRequestID,
			reqBody:        invalidAmountIsBelowMinLimit,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   wrapErrOpenApiNonSnap(40, "min amount 10.000"),
			mockSetup: func(merchant *serviceMocks.IMerchantService, disbursement *serviceMocks.IDisbursementService, _ *mockMQ.RabbitMQExt) {
				merchant.On(
					"FindMerchantByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Return(&merchantModel.Merchant{}, nil)
			},
		},
		{
			name:       "ERROR:Reference ID is already in use",
			reqSetting: validRequestID,
			reqBody:    ValidReqBody,
			mockSetup: func(_ *serviceMocks.IMerchantService, disbursement *serviceMocks.IDisbursementService, _ *mockMQ.RabbitMQExt) {
				disbursement.On(
					"IsExistReferenceID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Once().Return(true)
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   wrapErrOpenApiNonSnap(40, "reference ID already exist"),
		},
		{
			name: "ERROR:Total payout data exceeds the maximum",
			reqSetting: func(req *http.Request) {
				validRequestID(req)

				payload := disbursementModel.CreateDisbursementFromOpenApiRequest{}
				_ = json.Unmarshal([]byte(ValidReqBody), &payload)

				for i := 1; i <= constant.BulkDisbursementMaxDataRequest; i++ {
					payload.Payouts = append(payload.Payouts, payload.Payouts[0])
					payload.Payouts[i].ReferenceID = fmt.Sprintf("REF-%06d", i)
				}

				buf := new(bytes.Buffer)
				_ = json.NewEncoder(buf).Encode(&payload)

				req.Body = io.NopCloser(buf)
				req.ContentLength = int64(buf.Len())
			},
			mockSetup: func(_ *serviceMocks.IMerchantService, disbursement *serviceMocks.IDisbursementService, _ *mockMQ.RabbitMQExt) {
				disbursement.On(
					"IsExistReferenceID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Return(false)
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   wrapErrOpenApiNonSnap(40, "max bulk disbursement request is 1000"),
		},
		{
			name: "ERROR:Using magic account number",
			reqSetting: func(req *http.Request) {
				validRequestID(req)

				payload := disbursementModel.CreateDisbursementFromOpenApiRequest{}
				_ = json.Unmarshal([]byte(ValidReqBody), &payload)

				for i := 1; i <= 1; i++ {
					payload.Payouts = append(payload.Payouts, payload.Payouts[0])
					payload.Payouts[i].ReferenceID = fmt.Sprintf("REF-%06d", i)
					payload.Payouts[i].ChannelInformation.AccountNumber = "999966660004"
				}

				buf := new(bytes.Buffer)
				_ = json.NewEncoder(buf).Encode(&payload)

				req.Body = io.NopCloser(buf)
				req.ContentLength = int64(buf.Len())
			},
			mockSetup: func(_ *serviceMocks.IMerchantService, disbursement *serviceMocks.IDisbursementService, _ *mockMQ.RabbitMQExt) {
				disbursement.On(
					"IsExistReferenceID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Return(false)
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   wrapErrOpenApiNonSnap(40, constant.ErrDuplicateDisbursementReferenceId.Error()),
		},
		{
			name: "ERROR:Using timeout magic account number",
			reqSetting: func(req *http.Request) {
				validRequestID(req)

				payload := disbursementModel.CreateDisbursementFromOpenApiRequest{}
				_ = json.Unmarshal([]byte(ValidReqBody), &payload)

				for i := 1; i <= 1; i++ {
					payload.Payouts = append(payload.Payouts, payload.Payouts[0])
					payload.Payouts[i].ReferenceID = fmt.Sprintf("REF-%06d", i)
					payload.Payouts[i].ChannelInformation.AccountNumber = "999966660008"
				}

				buf := new(bytes.Buffer)
				_ = json.NewEncoder(buf).Encode(&payload)

				req.Body = io.NopCloser(buf)
				req.ContentLength = int64(buf.Len())
			},
			mockSetup: func(_ *serviceMocks.IMerchantService, disbursement *serviceMocks.IDisbursementService, _ *mockMQ.RabbitMQExt) {
				disbursement.On(
					"IsExistReferenceID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Return(false)
			},
			wantStatusCode: http.StatusGatewayTimeout,
			wantRespBody:   wrapErrOpenApiNonSnap(54, constant.ErrTimeout.Error(), "PARTNER_ERROR"),
		},
		{
			name:       "ERROR:Timeout by inquiry id",
			reqSetting: validRequestID,
			reqBody:    ValidReqBody,
			mockSetup: func(_ *serviceMocks.IMerchantService, _ *serviceMocks.IDisbursementService, _ *mockMQ.RabbitMQExt) {
				for _, fn := range append(redisSeq, redisDefer...) {
					fn()
				}

				mockAccountInquirySvc.On(
					"FindLatestByInquiryID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Once().Return(&requestAccountInquiries.RequestAccountInquiryWithMaster{
					RequestAccountInquiries: requestAccountInquiries.RequestAccountInquiries{
						BeneficiaryAccountNo: sql.NullString{
							Valid:  true,
							String: "999966660008",
						},
					},
				}, nil)
			},
			wantStatusCode: http.StatusGatewayTimeout,
			wantRespBody:   wrapErrOpenApiNonSnap(54, constant.ErrTimeout.Error(), "PARTNER_ERROR"),
		},
		{
			name:       "ERROR:Find latest by inquiry id",
			reqSetting: validRequestID,
			reqBody:    ValidReqBody,
			mockSetup: func(_ *serviceMocks.IMerchantService, _ *serviceMocks.IDisbursementService, _ *mockMQ.RabbitMQExt) {
				for _, fn := range append(redisSeq, redisDefer...) {
					fn()
				}

				mockAccountInquirySvc.On(
					"FindLatestByInquiryID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Once().Return(nil, pkgErrors.New(response.HttpErrRequest, errors.New("FIND: some error")))
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   wrapErrOpenApiNonSnap(40, "FIND: some error"),
		},
		{
			name:       "ERROR:Invalid beneficiary account",
			reqSetting: validRequestID,
			reqBody:    ValidReqBody,
			mockSetup: func(_ *serviceMocks.IMerchantService, _ *serviceMocks.IDisbursementService, _ *mockMQ.RabbitMQExt) {
				for _, fn := range append(redisSeq, redisDefer...) {
					fn()
				}

				mockAccountInquirySvc.On(
					"FindLatestByInquiryID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Once().Return(&requestAccountInquiries.RequestAccountInquiryWithMaster{
					RequestAccountInquiries: requestAccountInquiries.RequestAccountInquiries{
						Status: sql.NullString{String: constant.RequestAccountInquiryStatusInvalid},
					},
				}, nil)
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   wrapErrOpenApiNonSnap(40, "invalid status inquiry"),
		},
		{
			name:       "ERROR:Invalid beneficiary bank code",
			reqSetting: validRequestID,
			reqBody:    ValidReqBody,
			mockSetup: func(_ *serviceMocks.IMerchantService, _ *serviceMocks.IDisbursementService, _ *mockMQ.RabbitMQExt) {
				for _, fn := range append(redisSeq, redisDefer...) {
					fn()
				}

				mockAccountInquirySvc.On(
					"FindLatestByInquiryID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Once().Return(&requestAccountInquiries.RequestAccountInquiryWithMaster{}, nil)
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   wrapErrOpenApiNonSnap(40, "invalid beneficiary bank code"),
		},
		{
			name:       "ERROR:Set queue lock",
			reqSetting: validRequestID,
			reqBody:    ValidReqBody,
			mockSetup: func(_ *serviceMocks.IMerchantService, _ *serviceMocks.IDisbursementService, _ *mockMQ.RabbitMQExt) {
				mockAccountInquirySvc.On(
					"FindLatestByInquiryID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Return(&requestAccountInquiries.RequestAccountInquiryWithMaster{
					MasterBeneficiaryAccountName: "NAME",
					RequestAccountInquiries: requestAccountInquiries.RequestAccountInquiries{
						BeneficiaryBankCode:  "002",
						BeneficiaryAccountNo: sql.NullString{String: "ACC"},
					},
				}, nil)

				for _, fn := range redisSeq {
					fn()
				}

				clientMock.ExpectSetNX(queueKey, true, 0).RedisNil()

				for _, fn := range redisDefer {
					fn()
				}
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   wrapErrOpenApiNonSnap(99, "QUEUE: there is an internal error with the id ", response.ErrTypeUnknown),
		},
		{
			name:       "ERROR:Reference ID is in queue",
			reqSetting: validRequestID,
			reqBody:    ValidReqBody,
			mockSetup: func(_ *serviceMocks.IMerchantService, _ *serviceMocks.IDisbursementService, _ *mockMQ.RabbitMQExt) {
				for _, fn := range redisSeq {
					fn()
				}

				clientMock.ExpectSetNX(queueKey, true, 0).SetVal(false)

				for _, fn := range redisDefer {
					fn()
				}
				redisSeq = append(redisSeq, func() {
					clientMock.ExpectSetNX(queueKey, true, 0).SetVal(true)
				})
			},
			wantStatusCode: http.StatusConflict,
			wantRespBody:   wrapErrOpenApiNonSnap(49, constant.ErrPayoutsInProcess.Error()),
		},
		{
			name:       "ERROR:Get payout cut-off time status",
			reqSetting: validRequestID,
			reqBody:    ValidReqBody,
			mockSetup: func(_ *serviceMocks.IMerchantService, disbursement *serviceMocks.IDisbursementService, _ *mockMQ.RabbitMQExt) {
				for _, fn := range append(redisSeq, redisDefer...) {
					fn()
				}

				disbursement.On(
					"GetCutOffTimeStatus", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("time.Time"), mock.AnythingOfType("string"), mock.Anything,
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   wrapErrOpenApiNonSnap(99, "some error", response.ErrTypeUnknown),
		},
		{
			name:       "ERROR:Daily transaction limit reached",
			reqSetting: validRequestID,
			reqBody:    ValidReqBody,
			mockSetup: func(_ *serviceMocks.IMerchantService, disbursement *serviceMocks.IDisbursementService, _ *mockMQ.RabbitMQExt) {
				for _, fn := range append(redisSeq, redisDefer...) {
					fn()
				}

				disbursement.On(
					"GetCutOffTimeStatus", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("time.Time"), mock.AnythingOfType("string"), mock.Anything,
				).Once().Return(&disbursementModel.CutOffTimeStatusResponse{}, nil)
				disbursement.On(
					"ValidateDailyTransactionLimit", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("float64"),
				).Once().Return(nil, pkgErrors.New(response.HttpErrDailyLimitReached, constant.ErrDailyLimitReached))
			},
			wantStatusCode: http.StatusTooManyRequests,
			wantRespBody:   wrapErrOpenApiNonSnap(46, "you have exceeded your daily transaction limit"),
		},
		{
			name:       "ERROR:Fail create data in bulk",
			reqSetting: validRequestID,
			reqBody:    ValidReqBody,
			mockSetup: func(_ *serviceMocks.IMerchantService, disbursement *serviceMocks.IDisbursementService, _ *mockMQ.RabbitMQExt) {
				for _, fn := range append(redisSeq, redisDefer...) {
					fn()
				}
				disbursement.On(
					"GetCutOffTimeStatus", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("time.Time"), mock.AnythingOfType("string"), mock.Anything,
				).Once().Return(&disbursementModel.CutOffTimeStatusResponse{}, nil)
				disbursement.On(
					"ValidateDailyTransactionLimit", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("float64"),
				).Return(mockTrxCloser, nil)

				disbursement.On(
					"CreateBulk", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("*disbursementModel.CreateBulkDisbursementRequest"),
				).Once().Return(nil, pkgErrors.New(response.HttpErrRequest, errors.New("failed on create data in bulk")))
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   wrapErrOpenApiNonSnap(40, "failed on create data in bulk"),
		},
		{
			name:       "SUCCESS",
			reqSetting: validRequestID,
			reqBody:    ValidReqBody,
			mockSetup: func(_ *serviceMocks.IMerchantService, disbursement *serviceMocks.IDisbursementService, mq *mockMQ.RabbitMQExt) {
				// Mock FindMerchantByID
				mockMerchant.On(
					"FindMerchantByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Return(&merchantModel.Merchant{UUID: merchantId}, nil)

				// Mock IsExistReferenceID
				disbursement.On(
					"IsExistReferenceID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Return(false, nil)

				// Setup account inquiry mock
				mockAccountInquirySvc.On(
					"FindLatestByInquiryID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Return(&requestAccountInquiries.RequestAccountInquiryWithMaster{
					MasterBeneficiaryAccountName: "NAME",
					RequestAccountInquiries: requestAccountInquiries.RequestAccountInquiries{
						BeneficiaryBankCode:  "002",
						BeneficiaryAccountNo: sql.NullString{String: "ACC"},
					},
				}, nil)

				// Setup Redis mocks for locks
				clientMock.ExpectSetNX(queueKey, true, 0).SetVal(true)
				bulkQueueKey := fmt.Sprintf(constant.BulkDisbursementInProgressQueueLockFmt, merchantId, "")
				clientMock.ExpectSetNX(bulkQueueKey, true, time.Minute*30).SetVal(true)

				// Add whitelist check mock
				whitelistKey := fmt.Sprintf("backend-portal:trx-merchant-whitelist:%s", "R1")
				clientMock.ExpectGet(whitelistKey).SetVal("true")
				clientMock.ExpectDel(whitelistKey).SetVal(1)

				// Mock GetTransactionConfig
				disbursement.On(
					"GetTransactionConfig", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Return(trxConfig, nil)

				// Mock IsBankcodeOverbookingChannelAllowed
				disbursement.On("IsBankcodeOverbookingChannelAllowed", mock.AnythingOfType(constant.MockTypeValueContextReference), "002", constant.StringMockType()).Return(false)

				disbursement.On(
					"GetCutOffTimeStatus", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("time.Time"), mock.AnythingOfType("string"), mock.Anything,
				).Once().Return(&disbursementModel.CutOffTimeStatusResponse{}, nil)

				disbursement.On(
					"ValidateDailyTransactionLimit", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("float64"),
				).Return(mockTrxCloser, nil)

				disbursement.On(
					"CreateBulk", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("*disbursementModel.CreateBulkDisbursementRequest"),
				).Return(&disbursementModel.BulkDisbursement{}, nil)

				disbursement.On("IsMerchantAllowedExcludeBeneficiaryRules", mock.AnythingOfType(constant.MockTypeValueContextReference), merchantId, 0.0).Return(0.0, false)

				mq.On(
					"Publish",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Maybe().Return(nil)

				mq.On(
					"PublishActivity",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Maybe().Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"Success","data":{"uuid":"","merchantId":"12345","payouts":[{"referenceId":"R1","inquiryId":"ABC", "channelCode":"BRI","channelInformation":{"accountNumber":"ACC","accountName":"NAME"},"amount":{"currency":"IDR","value":"11000"},"description":"D1"}],"status":"","created":"0001-01-01T00:00:00Z","updated":"0001-01-01T00:00:00Z"}}`,
			skip:           true,
		},
		{
			name:       "SUCCESS_on_behalf_of_submerchant",
			reqSetting: validRequestID,
			reqBody:    ValidReqBody,
			setHeaders: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.Max.String())
			},
			mockSetup: func(_ *serviceMocks.IMerchantService, disbursement *serviceMocks.IDisbursementService, mq *mockMQ.RabbitMQExt) {
				// Mock FindMerchantByID
				mockMerchant.On(
					"FindMerchantByID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Return(&merchantModel.Merchant{UUID: merchantId}, nil)

				// Mock IsExistReferenceID
				disbursement.On(
					"IsExistReferenceID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Return(false, nil)

				// Setup account inquiry mock
				mockAccountInquirySvc.On(
					"FindLatestByInquiryID", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Return(&requestAccountInquiries.RequestAccountInquiryWithMaster{
					MasterBeneficiaryAccountName: "NAME",
					RequestAccountInquiries: requestAccountInquiries.RequestAccountInquiries{
						BeneficiaryBankCode:  "002",
						BeneficiaryAccountNo: sql.NullString{String: "ACC"},
					},
				}, nil)

				// Setup Redis mocks for locks
				queueKeySubmerchant := fmt.Sprintf(constant.BulkDisbursementQueueLockFmt, uuid.Max.String(), "R1")
				clientMock.ExpectSetNX(queueKeySubmerchant, true, 0).SetVal(true)
				bulkQueueKeySubmerchant := fmt.Sprintf(constant.BulkDisbursementInProgressQueueLockFmt, uuid.Max.String(), "")
				clientMock.ExpectSetNX(bulkQueueKeySubmerchant, true, time.Minute*30).SetVal(true)

				// Add whitelist check mock
				whitelistKey := fmt.Sprintf("backend-portal:trx-merchant-whitelist:%s", "R1")
				clientMock.ExpectGet(whitelistKey).SetVal("true")
				clientMock.ExpectDel(whitelistKey).SetVal(1)

				// Mock GetTransactionConfig
				disbursement.On(
					"GetTransactionConfig", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"),
				).Return(trxConfig, nil)

				// Mock IsBankcodeOverbookingChannelAllowed
				disbursement.On("IsBankcodeOverbookingChannelAllowed", mock.AnythingOfType(constant.MockTypeValueContextReference), "002", constant.StringMockType()).Return(false)

				disbursement.On(
					"GetCutOffTimeStatus", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("time.Time"), mock.AnythingOfType("string"), mock.Anything,
				).Return(&disbursementModel.CutOffTimeStatusResponse{
					Status: constant.DisbursementCutOffTimeStatusOngoing,
				}, nil)

				disbursement.On(
					"ValidateDailyTransactionLimit", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("float64"),
				).Return(mockTrxCloser, nil)

				disbursement.On(
					"CreateBulk", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("*disbursementModel.CreateBulkDisbursementRequest"),
				).Return(&disbursementModel.BulkDisbursement{}, nil)

				mq.On(
					"Publish",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Maybe().Return(nil)

				mq.On(
					"PublishActivity",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Maybe().Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"The transactions approved bla bla bla","data":{"uuid":"","merchantId":"ffffffff-ffff-ffff-ffff-ffffffffffff","payouts":[{"referenceId":"R1","inquiryId":"ABC", "channelCode":"BRI","channelInformation":{"accountNumber":"ACC","accountName":"NAME"},"amount":{"currency":"IDR","value":"11000"},"description":"D1"}],"status":"","created":"0001-01-01T00:00:00Z","updated":"0001-01-01T00:00:00Z"}}`,
			skip:           true,
		},
	}

	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {
			// Skip the test if skip is true
			if test.skip {
				t.Skip("Skipping test due to Redis mock issues")
			}

			buf.Reset()
			rec := httptest.NewRecorder()

			var body io.Reader
			if test.reqBody != "" {
				body = strings.NewReader(test.reqBody)
			}
			req := httptest.NewRequest(http.MethodPost, "/payouts", body)
			req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantIDKey, merchantPlatformWhitelistedOldResponseFormat))

			clientMock.ClearExpect()

			if test.reqSetting != nil {
				test.reqSetting(req)
			}
			if test.mockSetup != nil {
				test.mockSetup(mockMerchant, mockDisbursement, mockRabbitMq)
			}
			if test.setHeaders != nil {
				test.setHeaders(req)
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}

func TestCreatePayoutOverbooking(t *testing.T) {
	redisExt := redisMocks.NewIRedisExt(t)
	merchantSvc := serviceMocks.NewIMerchantService(t)
	disbursementSvc := serviceMocks.NewIDisbursementService(t)
	beneficiaryAccountSvc := serviceMocks.NewIBeneficiaryAccountService(t)

	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	defer buf.Reset()

	logger := logger.NewSlogger(logger.Config{}, logger.WithSlogOutput(buf))

	handler := internalPayoutController.New(
		&config.Config{
			WorkerPoolConfig: config.WorkerPoolConfig{
				Disbursement: 10,
			},
		}, validator.New(), disbursementSvc, merchantSvc, nil, nil,
		internalPayoutController.WithLogger(logger),
		internalPayoutController.WithRedisClient(redisExt),
		internalPayoutController.WithBeneficiaryAccountService(beneficiaryAccountSvc),
	)

	router := chi.NewRouter()
	router.Post("/payouts", handler.Create)

	merchantID := "e414d2f4-816b-42fa-b1d8-a198ac6446be"
	validWithChannelInfo := `{"payouts":[{"referenceId":"R1","channelCode":"BRI","channelInformation":{"accountNumber":"123456","accountName":"NAME"},"amount":{"currency":"IDR","value":"11000"},"description":""}]}`

	merchantSvc.On(
		"FindMerchantByID", mock.Anything, merchantID,
	).Return(&merchantModel.Merchant{UUID: merchantID}, nil)
	disbursementSvc.On(
		"GetTransactionConfig", mock.Anything, merchantID,
	).Return(&disbursementModel.TransactionConfig{}, nil)
	disbursementSvc.On("IsMerchantAllowedExcludeBeneficiaryRules", mock.Anything, merchantID, 0.0).Return(0.0, false)
	disbursementSvc.On("IsExistReferenceID", mock.Anything, merchantID, "R1").Return(false)
	disbursementSvc.On("IsBankcodeOverbookingChannelAllowed", mock.Anything, mock.Anything, merchantID).Return(true)

	boolCmd := &redis.BoolCmd{}
	boolCmd.SetVal(true)
	redisExt.On("SetNX", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(boolCmd)

	intCmd := &redis.IntCmd{}
	redisExt.On("Del", mock.Anything, mock.Anything).Return(intCmd)

	tests := []struct {
		name           string
		requestBody    string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:        "ERROR: Find beneficiary account",
			requestBody: validWithChannelInfo,
			setupMock: func() {
				beneficiaryAccountSvc.On(
					"FindByBankCodeAndAccountNo", mock.Anything, mock.Anything,
				).Once().Return(nil, assert.AnError)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"general_error","message":"General error","error":{"type":"API_ERROR","details":[{"field":"","message":"Please contact our representative team"}],"traceId":""}}`,
		},
		{
			name:        "ERROR: Payout not eligible",
			requestBody: validWithChannelInfo,
			setupMock: func() {
				disbursementSvc.On(
					"IsMerchantAllowedExcludeBeneficiaryRules", mock.Anything, merchantID, mock.Anything,
				).Return(1_000_000.00, true)
				beneficiaryAccountSvc.On(
					"FindByBankCodeAndAccountNo", mock.Anything, mock.Anything,
				).Once().Return(&beneficiaryAccountModel.Account{
					BeneficiaryBankCode:    "002",
					BeneficiaryAccountNo:   "123450000001",
					BeneficiaryAccountName: "TEST",
					MetadataObj: beneficiaryAccountModel.Metadata{
						IsVirtualAccount: true,
					},
				}, nil)
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"payout_not_eligible","message":"Destination account is not eligible for payout","error":{"type":"API_ERROR","details":[{"field":"destination account","message":"Destination account is not eligible for payout. bankCode=002 accountNumber=123450000001"}],"traceId":""}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/payouts", bytes.NewBufferString(tt.requestBody))

			req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{MerchantId: merchantID}))

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, tt.wantRespBody, rec.Body.String()) {
				t.Log("Actual:", rec.Body.String())
			}

			redisExt.AssertExpectations(t)
			merchantSvc.AssertExpectations(t)
			disbursementSvc.AssertExpectations(t)
			beneficiaryAccountSvc.AssertExpectations(t)
		})
	}
}
