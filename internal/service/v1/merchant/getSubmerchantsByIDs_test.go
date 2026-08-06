package merchant

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	logger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSubmerchantsByIDs(t *testing.T) {

	testCases := []struct {
		name    string
		setup   func(merchantRepo *repositoryMock.IMerchantRepository)
		wantErr bool
	}{
		{
			name: "SUCCESS: Get merchant customer",
			setup: func(merchantRepo *repositoryMock.IMerchantRepository) {
				merchantRepo.On(
					"GetSubmerchantsByIDs",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("[]string"),
				).Return(
					[]*merchant.Merchant{},
					nil,
				)
			},
		},
		{
			name: "ERROR: Error get merchant customer",
			setup: func(merchantRepo *repositoryMock.IMerchantRepository) {
				merchantRepo.On(
					"GetSubmerchantsByIDs",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("[]string"),
				).Return(
					nil,
					errors.New("error"),
				)
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := logger.NewZapLogger(logger.Config{})

			merchantRepo := repositoryMock.NewIMerchantRepository(t)
			tc.setup(merchantRepo)

			service := New(merchantRepo, mockLogger, nil, nil, nil, nil)
			_, err := service.GetSubmerchantsByIDs(context.Background(), uuid.NewString(), []string{uuid.NewString()})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

}
