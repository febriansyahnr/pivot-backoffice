package passwordHistories

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	mockPh "github.com/paper-indonesia/pivot-backoffice/mocks/repository/passwordHistories"
	mocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPasswordHistoriesService_InsertByUserID(t *testing.T) {
	testCases := []struct {
		Name      string
		IsSuccess bool
		MockSetup func(mockRepo *mockPh.IPasswordHistoriesRepository)
	}{
		{
			Name:      "SUCCESS: insert password history",
			IsSuccess: true,
			MockSetup: func(mockRepo *mockPh.IPasswordHistoriesRepository) {
				mockRepo.On(
					"Insert",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockRepo := mockPh.NewIPasswordHistoriesRepository(t)
			mockLogger, _ := mocks.NewZapLogger(mocks.Config{})

			ctx := context.Background()

			tc.MockSetup(mockRepo)

			userService := New(mockRepo, mockLogger)

			err := userService.InsertByUserID(ctx, uuid.NewString(), "random")
			if tc.IsSuccess {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
