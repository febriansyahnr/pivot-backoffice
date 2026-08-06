package creditcard_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/creditcard"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRemoveCardTokenization(t *testing.T) {
	log := loggerMock.NewILogger(t)
	paymentRepo := repoMocks.NewIPaymentRepository(t)
	customerRepo := repoMocks.NewICustomerRepository(t)

	service := New(nil, log, nil, paymentRepo, nil, nil, WithCustomerRepo(customerRepo))

	errInternalService := pkgErrors.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)

	request := unifiedPaymentModel.RemoveCardTokenizationRequest{
		MerchantID: "7b771512-02c3-4cfc-a7ce-f69aebc5bfb5",
		CustomerID: "019b1750-33cf-4a7e-8b4d-e9a40f91f17a",
		TokenID:    "0acb85d2-fb02-4e6e-bf93-e831f1bc0933",
		PaymentID:  "a4f760d9-60e1-47b7-90a5-7b698665412c",
	}

	tests := []struct {
		name      string
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR:Get payment details", // NOSONAR
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed to get payment details based on payment id while removing card tokenization", logger.Error(assert.AnError)).Once().Return()
			},
			wantError: errInternalService,
		},
		{
			name: "ERROR:Payment not found", // NOSONAR
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
			wantError: pkgErrors.New(response.HttpErrNotFound, constant.ErrPaymentNotFound),
		},
		{
			name: "ERROR:Action not permitted", // NOSONAR
			setupMock: func() {
				paymentRepo.On("GetPaymentById", mock.Anything, mock.Anything).Once().Return(&paymentModel.Payment{}, nil)
			},
			wantError: pkgErrors.New(response.HttpErrForbidden, errors.New("action not permitted")),
		},
		{
			name: "ERROR:Get customer details", // NOSONAR
			setupMock: func() {
				paymentRepo.On(
					"GetPaymentById", mock.Anything, mock.Anything,
				).Return(&paymentModel.Payment{
					MerchantID: request.MerchantID,
					CustomerID: request.CustomerID,
				}, nil)
				customerRepo.On("FindCustomerById", mock.Anything, request.CustomerID).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed to get customer details based on customer id while removing card tokenization", logger.Error(assert.AnError)).Once().Return()
			},
			wantError: errInternalService,
		},
		{
			name: "ERROR:Parse payment methods", // NOSONAR
			setupMock: func() {
				customerRepo.On("FindCustomerById", mock.Anything, request.CustomerID).Once().Return(&customerModel.Customer{Metadata: map[string]any{}}, nil)
				log.On("Error", mock.Anything, "Failed while convert map value to struct", mock.Anything).Once().Return()
			},
			wantError: errInternalService,
		},
		{
			name: "ERROR:Token ID not found", // NOSONAR
			setupMock: func() {
				customerRepo.On(
					"FindCustomerById", mock.Anything, request.CustomerID,
				).Return(&customerModel.Customer{Metadata: map[string]any{"paymentMethods": `[]`}}, nil)
				customerRepo.On(
					"RemovePaymentMethodFromCustomerByIDAndTokenID", mock.Anything, request.CustomerID, request.TokenID, mock.Anything,
				).Once().Return(constant.ErrDataNotFound)
			},
			wantError: pkgErrors.New(response.HttpErrNotFound, fmt.Errorf("payment method with token id %s not found", request.TokenID)),
		},
		{
			name: "ERROR:Remove card tokenization", // NOSONAR
			setupMock: func() {
				customerRepo.On(
					"RemovePaymentMethodFromCustomerByIDAndTokenID", mock.Anything, request.CustomerID, request.TokenID, mock.Anything,
				).Once().Return(assert.AnError)
				log.On("Error", mock.Anything, "Failed to remove payment method from customer", logger.Error(assert.AnError)).Once().Return()
			},
			wantError: errInternalService,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				customerRepo.On(
					"RemovePaymentMethodFromCustomerByIDAndTokenID", mock.Anything, request.CustomerID, request.TokenID, mock.Anything,
				).Once().Return(nil)
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantError, service.RemoveCardTokenization(t.Context(), request))
		})
	}
}
