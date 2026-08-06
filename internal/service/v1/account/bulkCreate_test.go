package accountService

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	repoMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBulkCreateAccount(t *testing.T) {
	merchants := []*merchant.Merchant{
		{UUID: uuid.NewString()},
	}
	customers := []*customerModel.Customer{
		{UUID: uuid.NewString()},
	}

	input := &account_model.BulkCreateAccountRequest{
		Currency: "IDR",
		Usecase:  constant.ReferenceWallet,
	}

	testCases := []struct {
		Name      string
		MockSetup func(accRepo *repoMock.IAccountRepository)
		WantErr   bool
		Input     *account_model.BulkCreateAccountRequest
	}{
		{
			Name: "SUCCESS: Create bulk Merchant Account",
			MockSetup: func(accRepo *repoMock.IAccountRepository) {
				accRepo.On("GetMerchantsWithoutAccount", mock.Anything, mock.Anything).Once().Return(merchants, nil)
				accRepo.On("GetMerchantsWithoutAccount", mock.Anything, mock.Anything).Return(nil, nil)
				accRepo.On("GetCustomersWithoutAccount", mock.Anything, mock.Anything).Once().Return(customers, nil)
				accRepo.On("GetCustomersWithoutAccount", mock.Anything, mock.Anything).Return(nil, nil)
				accRepo.On("BulkInsert", mock.Anything, mock.Anything).Return(nil)

				accRepo.On("BulkInsert", mock.Anything, mock.Anything).Return(nil)
			},
			WantErr: false,
			Input:   input,
		},
		{
			Name: "SUCCESS: Create bulk Merchant Account",
			MockSetup: func(accRepo *repoMock.IAccountRepository) {
				accRepo.On("GetMerchantsWithoutAccount", mock.Anything, mock.Anything).Once().Return(merchants, nil)
				accRepo.On("GetMerchantsWithoutAccount", mock.Anything, mock.Anything).Return(nil, nil)
				accRepo.On("GetCustomersWithoutAccount", mock.Anything, mock.Anything).Return(nil, errors.New("error"))

				accRepo.On("BulkInsert", mock.Anything, mock.Anything).Return(nil)
			},
			WantErr: true,
			Input:   input,
		},
		{
			Name: "ERROR: Create bulk Merchant Account",
			MockSetup: func(accRepo *repoMock.IAccountRepository) {
				accRepo.On("GetMerchantsWithoutAccount", mock.Anything, mock.Anything).Once().Return(merchants, nil)
				accRepo.On("GetMerchantsWithoutAccount", mock.Anything, mock.Anything).Return(nil, nil)
				accRepo.On("BulkInsert", mock.Anything, mock.Anything).Return(errors.New("error"))
			},
			WantErr: true,
			Input:   input,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			accRepo := &repoMock.IAccountRepository{}
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.MockSetup(accRepo)

			svc := New(logger, nil, accRepo, nil)
			err := svc.BulkCreateAccount(context.Background(), tc.Input)
			if tc.WantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}

}

func TestBulkCreateMerchantAccount(t *testing.T) {
	merchants := []*merchant.Merchant{
		{UUID: uuid.NewString()},
	}
	input := &account_model.BulkCreateAccountRequest{
		Currency: "IDR",
		Usecase:  constant.ReferenceWallet,
	}

	testCases := []struct {
		Name      string
		MockSetup func(accRepo *repoMock.IAccountRepository)
		WantErr   bool
		Input     *account_model.BulkCreateAccountRequest
	}{
		{
			Name: "SUCCESS: Create bulk Merchant Account",
			MockSetup: func(accRepo *repoMock.IAccountRepository) {
				accRepo.On("GetMerchantsWithoutAccount", mock.Anything, mock.Anything).Once().Return(merchants, nil)
				accRepo.On("GetMerchantsWithoutAccount", mock.Anything, mock.Anything).Return(nil, nil)
				accRepo.On("BulkInsert", mock.Anything, mock.Anything).Return(nil)
			},
			WantErr: false,
			Input:   input,
		},
		{
			Name: "SUCCESS: Create bulk SubMerchant Account",
			MockSetup: func(accRepo *repoMock.IAccountRepository) {
				accRepo.On("GetMerchantsWithoutAccount", mock.Anything, mock.Anything).Once().Return([]*merchant.Merchant{
					{
						UUID: uuid.NewString(),
						ParentID: sql.NullString{
							String: uuid.NewString(),
							Valid:  true,
						},
					},
				}, nil)
				accRepo.On("GetMerchantsWithoutAccount", mock.Anything, mock.Anything).Return(nil, nil)
				accRepo.On("BulkInsert", mock.Anything, mock.Anything).Return(nil)
			},
			WantErr: false,
			Input:   input,
		},
		{
			Name: "ERROR: Create bulk Merchant Account",
			MockSetup: func(accRepo *repoMock.IAccountRepository) {
				accRepo.On("GetMerchantsWithoutAccount", mock.Anything, mock.Anything).Once().Return(merchants, nil)
				accRepo.On("GetMerchantsWithoutAccount", mock.Anything, mock.Anything).Return(nil, nil)
				accRepo.On("BulkInsert", mock.Anything, mock.Anything).Return(errors.New("error"))
			},
			WantErr: true,
			Input:   input,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			accRepo := &repoMock.IAccountRepository{}
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.MockSetup(accRepo)

			svc := account{
				logger:      logger,
				accountRepo: accRepo,
			}

			err := svc.bulkCreateMerchantAccount(context.Background(), tc.Input)
			if tc.WantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestBulkCreateCustomerAccount(t *testing.T) {
	customers := []*customerModel.Customer{
		{UUID: uuid.NewString()},
	}
	input := &account_model.BulkCreateAccountRequest{
		Currency:   "IDR",
		Usecase:    constant.ReferenceWallet,
		MerchantID: uuid.NewString(),
	}

	testCases := []struct {
		Name      string
		MockSetup func(accRepo *repoMock.IAccountRepository)
		WantErr   bool
		Input     *account_model.BulkCreateAccountRequest
	}{
		{
			Name: "SUCCESS: Create bulk Customer Account",
			MockSetup: func(accRepo *repoMock.IAccountRepository) {
				accRepo.On("GetCustomersWithoutAccount", mock.Anything, mock.Anything).Once().Return(customers, nil)
				accRepo.On("GetCustomersWithoutAccount", mock.Anything, mock.Anything).Return(nil, nil)
				accRepo.On("BulkInsert", mock.Anything, mock.Anything).Return(nil)
			},
			WantErr: false,
			Input:   input,
		},
		{
			Name: "ERROR: Create bulk Customer Account",
			MockSetup: func(accRepo *repoMock.IAccountRepository) {
				accRepo.On("GetCustomersWithoutAccount", mock.Anything, mock.Anything).Once().Return(customers, nil)
				accRepo.On("GetCustomersWithoutAccount", mock.Anything, mock.Anything).Return(nil, nil)
				accRepo.On("BulkInsert", mock.Anything, mock.Anything).Return(errors.New("error"))
			},
			WantErr: true,
			Input:   input,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			accRepo := &repoMock.IAccountRepository{}
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.MockSetup(accRepo)

			svc := account{
				logger:      logger,
				accountRepo: accRepo,
			}

			err := svc.bulkCreateCustomerAccount(context.Background(), tc.Input)
			if tc.WantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
