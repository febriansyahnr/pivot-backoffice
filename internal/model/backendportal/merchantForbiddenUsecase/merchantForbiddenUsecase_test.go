package merchantForbiddenUsecase

import (
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/stretchr/testify/assert"
)

func TestNewMerchantForbiddenUseCase(t *testing.T) {
	merchantID := uuid.New().String()
	testCases := []struct {
		Name             string
		Input            *NewMerchantForbiddenUseCaseRequest
		ExpectedResponse *MerchantForbiddenUseCase
	}{
		{
			Name: "Create new merchant forbidden use case",
			Input: &NewMerchantForbiddenUseCaseRequest{
				MerchantID: merchantID,
				UseCase:    constant.UseCaseDisbursement,
			},
			ExpectedResponse: &MerchantForbiddenUseCase{
				MerchantID: merchantID,
				UseCase:    constant.UseCaseDisbursement,
			},
		},
		{
			Name: "not specify merchant id",
			Input: &NewMerchantForbiddenUseCaseRequest{
				MerchantID: "",
				UseCase:    constant.UseCaseDisbursement,
			},
			ExpectedResponse: &MerchantForbiddenUseCase{
				MerchantID: "",
				UseCase:    constant.UseCaseDisbursement,
			},
		},
		{
			Name: "not specify use case",
			Input: &NewMerchantForbiddenUseCaseRequest{
				MerchantID: merchantID,
				UseCase:    "",
			},
			ExpectedResponse: &MerchantForbiddenUseCase{
				MerchantID: merchantID,
				UseCase:    "",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			output := NewMerchantForbiddenUseCase(test.Input)
			assert.Equal(t, output.MerchantID, test.ExpectedResponse.MerchantID)
			assert.Equal(t, output.UseCase, test.ExpectedResponse.UseCase)
			assert.NotNil(t, output.UUID)
			assert.NotNil(t, output.CreatedAt)
			assert.NotNil(t, output.UpdatedAt)
		})
	}
}

func TestIsUseCaseExists(t *testing.T) {
	useCase := "disbursement"
	assert.True(t, IsUseCaseExists(useCase))
	useCase = "DISBURSEMENT"
	assert.True(t, IsUseCaseExists(useCase))
	useCase = "disbursem"
	assert.False(t, IsUseCaseExists(useCase))

	useCase = "withdrawal"
	assert.True(t, IsUseCaseExists(useCase))
	useCase = "WITHDRAWAL"
	assert.True(t, IsUseCaseExists(useCase))
	useCase = "withd"
	assert.False(t, IsUseCaseExists(useCase))
}
