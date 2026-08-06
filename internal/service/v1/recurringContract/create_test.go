package recurringContractService_test

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/recurringContract"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/recurringContract"
	logMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	log := logMock.NewILogger(t)
	customerSvc := serviceMocks.NewICustomerService(t)
	repo := repoMocks.NewIRecurringContractRepository(t)

	service := New(log, repo, customerSvc)

	customerID := "2ac93f16-93d8-4c2c-a0f2-27c48887617b"
	merchantID := "12f513ca-d538-412a-92a2-6a02344d9b6c"

	requestWithCustomerID := model.CreateRecurringContractRequest{
		MerchantID: merchantID,
		CustomerID: util.ValueToPtr(customerID),
	}
	requestWithCustomerObject := model.CreateRecurringContractRequest{
		MerchantID: merchantID,
		Customer: &unifiedPaymentModel.CustomerInformation{
			PhoneNumber: &unifiedPaymentModel.UnifiedPaymentPhoneNumber{},
		},
	}

	tests := []struct {
		name      string
		request   model.CreateRecurringContractRequest
		setupMock func()
		wantError error
	}{
		{
			name:    "ERROR:Get customer by id",
			request: requestWithCustomerID,
			setupMock: func() {
				customerSvc.On(
					"GetCustomerById", mock.Anything, customerID, merchantID,
				).Once().Return(nil, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name:    "ERROR:Upsert customer data",
			request: requestWithCustomerObject,
			setupMock: func() {
				customerSvc.On("CreateUnfiedPaymentCustomer", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name:    "ERROR:Insert recurring contract",
			request: requestWithCustomerObject,
			setupMock: func() {
				customerSvc.On(
					"CreateUnfiedPaymentCustomer", mock.Anything, mock.Anything,
				).Once().Return(&customerModel.GeneralCustomerResponse{}, nil)
				repo.On("Insert", mock.Anything, mock.Anything).Once().Return(assert.AnError)
				log.On("Error", mock.Anything, "Failed when inserting recurring contract data", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name:    "ERROR:Client reference id already exists",
			request: requestWithCustomerObject,
			setupMock: func() {
				customerSvc.On("CreateUnfiedPaymentCustomer", mock.Anything, mock.Anything).Once().Return(&customerModel.GeneralCustomerResponse{}, nil)
				repo.On("Insert", mock.Anything, mock.Anything).Once().Return(constant.ErrClientReferenceIDAlreadyExist)
			},
			wantError: pkgErrs.New(response.HttpErrDupCheck, constant.ErrClientReferenceIDAlreadyExist),
		},
		{
			name:    "SUCCESS",
			request: requestWithCustomerID,
			setupMock: func() {
				customerSvc.On("GetCustomerById", mock.Anything, customerID, merchantID).Once().Return(nil, nil)
				repo.On("Insert", mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.Create(t.Context(), test.request)

			require.Equal(t, test.wantError, err)
			if err == nil {
				require.NotNil(t, result)
				require.NotEmpty(t, result.RecurringID)
			}

			log.AssertExpectations(t)
			repo.AssertExpectations(t)
			customerSvc.AssertExpectations(t)
		})
	}
}
