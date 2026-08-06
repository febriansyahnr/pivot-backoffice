package beneficiaryAccountService

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetList(t *testing.T) {
	data := make([]beneficiaryAccountModel.BeneficiaryAccount, 0)
	data = append(data, beneficiaryAccountModel.BeneficiaryAccount{
		UUID: uuid.NewString(),
	})
	expectedResponse := commonModel.PaginationResponse{
		Data: data,
		Meta: commonModel.Meta{
			Page:       1,
			PerPage:    20,
			TotalItems: 1,
			TotalPages: 1,
		},
	}

	testCases := []struct {
		Name           string
		WantErr        bool
		Context        context.Context
		ExpectedResult *commonModel.PaginationResponse
		ExpectedError  string
		MockSetup      func(mockRepo *mocks.IBeneficiaryAccountRepository)
	}{
		{
			Name:           "SUCCESS: getList for regular merchant",
			WantErr:        false,
			Context:        context.Background(),
			ExpectedResult: &expectedResponse,
			MockSetup: func(mockRepo *mocks.IBeneficiaryAccountRepository) {
				mockRepo.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(&expectedResponse, nil)
			},
		},
		{
			Name:           "SUCCESS: getList for derived merchant",
			WantErr:        false,
			Context:        context.WithValue(context.Background(), constant.CtxDerivedMerchantID, "derived-merchant-123"),
			ExpectedResult: &expectedResponse,
			MockSetup: func(mockRepo *mocks.IBeneficiaryAccountRepository) {
				mockRepo.On(
					"GetListOfDerived",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(&expectedResponse, nil)
			},
		},
		{
			Name:           "SUCCESS: getList with empty derived merchant ID falls back to regular",
			WantErr:        false,
			Context:        context.WithValue(context.Background(), constant.CtxDerivedMerchantID, ""),
			ExpectedResult: &expectedResponse,
			MockSetup: func(mockRepo *mocks.IBeneficiaryAccountRepository) {
				mockRepo.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(&expectedResponse, nil)
			},
		},
		{
			Name:           "FAILED: got error on repository when getList for regular merchant",
			WantErr:        true,
			Context:        context.Background(),
			ExpectedResult: &expectedResponse,
			ExpectedError:  "failed to getList",
			MockSetup: func(mockRepo *mocks.IBeneficiaryAccountRepository) {
				mockRepo.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(nil, errors.New("failed to getList"))
			},
		},
		{
			Name:           "FAILED: got error on repository when getList for derived merchant",
			WantErr:        true,
			Context:        context.WithValue(context.Background(), constant.CtxDerivedMerchantID, "derived-merchant-123"),
			ExpectedResult: &expectedResponse,
			ExpectedError:  "failed to getListOfDerived",
			MockSetup: func(mockRepo *mocks.IBeneficiaryAccountRepository) {
				mockRepo.On(
					"GetListOfDerived",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(nil, errors.New("failed to getListOfDerived"))
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			benMock := mocks.NewIBeneficiaryAccountRepository(t)
			accountInquiryRepoMock := mocks.NewIAccountInquiriesRepository(t)
			snapMock := mocks.NewISnapCoreRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			tc.MockSetup(benMock)

			svc := New(loggerMock, benMock, accountInquiryRepoMock, snapMock)
			response, err := svc.GetList(tc.Context, &beneficiaryAccountModel.GetBeneficiaryAccountFilterRequest{}, 1, 20)
			if tc.WantErr {
				require.Error(t, err)
				require.Empty(t, response)
				require.True(t, strings.Contains(err.Error(), tc.ExpectedError))
			} else {
				assert.NoError(t, err)
				require.NotEmpty(t, response)
			}

			benMock.AssertExpectations(t)
			accountInquiryRepoMock.AssertExpectations(t)
			snapMock.AssertExpectations(t)

		})
	}
}
