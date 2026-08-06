package merchant_test

import (
	"context"
	"errors"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/merchant"
	rabbitMqExt "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	s "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDepositSetting(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	repo := repoMocks.NewIMerchantRepository(t)

	service := New(repo, logger, nil, nil, nil, nil)

	response := &merchant.DepositSettingResponse{
		AutoWithdrawal: "OFF",
	}
	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *merchant.DepositSettingResponse
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				repo.On(
					"GetDepositSetting", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				repo.On("GetDepositSetting", c.ValueCtxMockType(), c.StringMockType()).Return(response, nil)
			},
			wantResult: response,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.GetDepositSetting(context.Background(), "12345")
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestSetAutoWithdrawal(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	rmq := rabbitMqExt.NewRabbitMQExt(t)
	repo := repoMocks.NewIMerchantRepository(t)

	bankAccountRepo := repoMocks.NewIBankAccountRepository(t)

	service := New(repo, logger, nil, nil, rmq, nil, WithBankAccountRepository(bankAccountRepo))

	rmq.On(
		"PublishActivity", c.ValueCtxMockType(), mock.Anything, mock.Anything, c.StringMockType(), c.StringMockType(), mock.Anything,
	).Return(nil)
	requestMockType := mock.AnythingOfType("*merchant.AutoWithdrawalSettingRequest")

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Validate bank account",
			setupMock: func() {
				bankAccountRepo.On(
					"BankAccountHasBeenPrepared", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(false, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:No bank account",
			setupMock: func() {
				bankAccountRepo.On(
					"BankAccountHasBeenPrepared", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(false, nil)
			},
			wantErr: pkgErrs.New(s.HttpErrUnprocessableContent, errors.New("bank account not found")),
		},
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				bankAccountRepo.On(
					"BankAccountHasBeenPrepared", c.ValueCtxMockType(), c.StringMockType(),
				).Return(true, nil)
				repo.On("SetAutoWithdrawal", c.ValueCtxMockType(), requestMockType).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				repo.On("SetAutoWithdrawal", c.ValueCtxMockType(), requestMockType).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantErr, service.SetAutoWithdrawal(context.Background(), &merchant.AutoWithdrawalSettingRequest{Status: "ON"}))
		})
	}
}
