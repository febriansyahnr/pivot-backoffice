package routingprocessorService

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/accountInquiry"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	rabbitmqExt "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcessAccountInquiryCallback(t *testing.T) {
	type mocker struct {
		cfg *config.Config

		routingProcessorRepo      *repositoryMocks.IRoutingProcessorRepository
		accountInquiryRepo        *repositoryMocks.IAccountInquiriesRepository
		outboundRepo              *repositoryMocks.IOutboundRepository
		rabbitMq                  *rabbitmqExt.RabbitMQExt
		requestAccountInquiryRepo *repositoryMocks.IRequestAccountInquiryRepository

		payloadRequest *routingProcessorModel.InquiryAccountResponseData
	}

	dataClient, _ := json.Marshal(outbound.Client{
		RequestId:      uuid.NewString(),
		From:           "test",
		OriginId:       uuid.NewString(),
		ReferenceId:    uuid.NewString(),
		ReplyToAddress: uuid.NewString(),
	})

	dataHeader, _ := json.Marshal(map[string]string{
		"Content-Type": "application/json",
	})

	validOutbound := &outbound.Outbound{
		Id:      uuid.NewString(),
		Client:  dataClient,
		Method:  "POST",
		Date:    time.Now(),
		URL:     "/v2/flip/bank-account-inquiry",
		Headers: dataHeader,
	}

	testCases := []struct {
		desc      string
		wantErr   bool
		mockSetup func(m *mocker)
	}{
		{
			desc:    "error when find by id",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.outboundRepo.On("FindByID", mock.Anything, mock.AnythingOfType("string")).Return(nil, assert.AnError)
			},
		},
		{
			desc:    "error when no data found",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.outboundRepo.On("FindByID", mock.Anything, mock.AnythingOfType("string")).Return(nil, nil)
			},
		},
		{
			desc:    "error when reply to is empty",
			wantErr: true,
			mockSetup: func(m *mocker) {
				invalidClient := validOutbound
				invalidClient.Client, _ = json.Marshal(outbound.Client{})

				m.outboundRepo.On("FindByID", mock.Anything, mock.AnythingOfType("string")).Return(invalidClient, nil)
				m.requestAccountInquiryRepo.On("FindByID", mock.Anything, mock.AnythingOfType("string")).Return(nil, assert.AnError)
			},
		},
		{
			desc:    "error when unmarshalling data client",
			wantErr: true,
			mockSetup: func(m *mocker) {
				invalidClient := validOutbound
				invalidClient.Client = []byte("invalid")

				m.outboundRepo.On("FindByID", mock.Anything, mock.AnythingOfType("string")).Return(invalidClient, nil)
			},
		},
		{
			desc:    "error when publish message to Reply queue",
			wantErr: true,
			mockSetup: func(m *mocker) {
				validOutbound.Client, _ = json.Marshal(outbound.Client{
					ReplyToAddress: uuid.NewString(),
				})
				m.outboundRepo.On("FindByID", mock.Anything, mock.AnythingOfType("string")).Return(validOutbound, nil)
				m.requestAccountInquiryRepo.On("FindByID", mock.Anything, mock.AnythingOfType("string")).Return(nil, assert.AnError)
				m.rabbitMq.On("PublishToReplyQueue", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(assert.AnError)
			},
		},
		{
			desc:    "success when publish message to Reply queue",
			wantErr: false,
			mockSetup: func(m *mocker) {
				validOutbound.Client, _ = json.Marshal(outbound.Client{
					ReplyToAddress: uuid.NewString(),
				})
				m.requestAccountInquiryRepo.On("FindByID", mock.Anything, mock.AnythingOfType("string")).Return(nil, assert.AnError)
				m.outboundRepo.On("FindByID", mock.Anything, mock.AnythingOfType("string")).Return(validOutbound, nil)
				m.rabbitMq.On("PublishToReplyQueue", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			m := &mocker{
				cfg: &config.Config{},

				routingProcessorRepo:      repositoryMocks.NewIRoutingProcessorRepository(t),
				accountInquiryRepo:        repositoryMocks.NewIAccountInquiriesRepository(t),
				requestAccountInquiryRepo: repositoryMocks.NewIRequestAccountInquiryRepository(t),
				outboundRepo:              repositoryMocks.NewIOutboundRepository(t),
				rabbitMq:                  rabbitmqExt.NewRabbitMQExt(t),

				payloadRequest: &routingProcessorModel.InquiryAccountResponseData{
					ResponseCode:    "2001800",
					ResponseMessage: "Success",
				},
			}

			tc.mockSetup(m)
			pdkLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			svc := New(m.cfg, pdkLogger, map[string]repository.IRoutingProcessorRepository{
				constant.FlipPGProcessor: m.routingProcessorRepo,
			},
				WithOutboundRepository(m.outboundRepo),
				WithRabbitMqExt(m.rabbitMq),
				WithRequestAccountInquiryRepository(m.requestAccountInquiryRepo),
			)

			err := svc.ProcessAccountInquiryCallback(context.Background(), m.payloadRequest)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

		})
	}
}

func TestAddressingReplyToAccountInquiry(t *testing.T) {
	type mocker struct {
		cfg                  *config.Config
		routingProcessorRepo *repositoryMocks.IRoutingProcessorRepository
		accountInquiryRepo   *repositoryMocks.IAccountInquiriesRepository
		outboundRepo         *repositoryMocks.IOutboundRepository
		rabbitMq             *rabbitmqExt.RabbitMQExt

		ctx            context.Context
		payloadRequest *routingProcessorModel.InquiryAccountResponseData
	}

	dataClient, _ := json.Marshal(outbound.Client{
		RequestId:      uuid.NewString(),
		From:           "test",
		OriginId:       uuid.NewString(),
		ReferenceId:    uuid.NewString(),
		ReplyToAddress: uuid.NewString(),
	})

	dataHeader, _ := json.Marshal(map[string]string{
		"Content-Type": "application/json",
	})

	validOutbound := &outbound.Outbound{
		Id:      uuid.NewString(),
		Client:  dataClient,
		Method:  "POST",
		Date:    time.Now(),
		URL:     "/v2/flip/bank-account-inquiry",
		Headers: dataHeader,
	}

	testCases := []struct {
		desc      string
		wantErr   bool
		mockSetup func(m *mocker)
	}{
		{
			desc:    "error when find by id",
			wantErr: true,
			mockSetup: func(m *mocker) {
				m.outboundRepo.On("FindByID", mock.Anything, mock.AnythingOfType("string")).Return(nil, assert.AnError)
			},
		},
		{
			desc:    "error when unmarshalling data client",
			wantErr: true,
			mockSetup: func(m *mocker) {
				invalidClient := validOutbound
				invalidClient.Client = []byte("invalid")

				m.outboundRepo.On("FindByID", mock.Anything, mock.AnythingOfType("string")).Return(invalidClient, nil)
			},
		},
		{
			desc:    "error when get reply to address from ctx",
			wantErr: true,
			mockSetup: func(m *mocker) {
				validOutbound.Client, _ = json.Marshal(outbound.Client{
					ReplyToAddress: uuid.NewString(),
				})
				m.outboundRepo.On("FindByID", mock.Anything, mock.AnythingOfType("string")).Return(validOutbound, nil)
			},
		},
		{
			desc:    "error when update client data",
			wantErr: true,
			mockSetup: func(m *mocker) {
				validOutbound.Client, _ = json.Marshal(outbound.Client{
					ReplyToAddress: uuid.NewString(),
				})
				m.ctx = context.WithValue(m.ctx, constant.CtxRabbitMQReplyTo, uuid.NewString())
				m.outboundRepo.On("FindByID", mock.Anything, mock.AnythingOfType("string")).Return(validOutbound, nil)
				m.outboundRepo.On("UpdateClient", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*outbound.Client")).Return(assert.AnError)

			},
		},
		{
			desc:    "success when update client data",
			wantErr: false,
			mockSetup: func(m *mocker) {
				validOutbound.Client, _ = json.Marshal(outbound.Client{
					ReplyToAddress: uuid.NewString(),
				})
				m.ctx = context.WithValue(m.ctx, constant.CtxRabbitMQReplyTo, uuid.NewString())
				m.outboundRepo.On("FindByID", mock.Anything, mock.AnythingOfType("string")).Return(validOutbound, nil)
				m.outboundRepo.On("UpdateClient", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*outbound.Client")).Return(nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			m := &mocker{
				cfg:                  &config.Config{},
				routingProcessorRepo: repositoryMocks.NewIRoutingProcessorRepository(t),
				accountInquiryRepo:   repositoryMocks.NewIAccountInquiriesRepository(t),
				outboundRepo:         repositoryMocks.NewIOutboundRepository(t),
				rabbitMq:             rabbitmqExt.NewRabbitMQExt(t),

				ctx: context.Background(),
				payloadRequest: &routingProcessorModel.InquiryAccountResponseData{
					ResponseCode:       "2001800",
					ResponseMessage:    "Success",
					PartnerReferenceNo: "BT-123",
				},
			}

			tc.mockSetup(m)
			pdkLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			svc := New(m.cfg, pdkLogger, map[string]repository.IRoutingProcessorRepository{
				constant.FlipPGProcessor: m.routingProcessorRepo,
			},
				WithOutboundRepository(m.outboundRepo),
				WithRabbitMqExt(m.rabbitMq),
			)

			err := svc.AddressingReplyToAccountInquiry(m.ctx, m.payloadRequest)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
