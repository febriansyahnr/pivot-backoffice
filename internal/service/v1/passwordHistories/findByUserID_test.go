package passwordHistories

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/passwordHistories"
	mockPh "github.com/paper-indonesia/pivot-backoffice/mocks/repository/passwordHistories"
	mocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFindByUserID(t *testing.T) {
	history := &passwordHistories.PasswordHistories{
		UUID:           "uuid-uuid-uuid",
		UserID:         "uuid-uuid-uuid",
		PasswordHashes: "password-hashes",
		CreatedAt:      time.Now(),
	}
	expectedHistories := []*passwordHistories.PasswordHistories{history}

	testCases := []struct {
		Name           string
		IsSuccess      bool
		UserID         string
		ExpectedResult []*passwordHistories.PasswordHistories
		ExpectedError  string
		MockSetup      func(mockRepo *mockPh.IPasswordHistoriesRepository)
	}{
		{
			Name:           "SUCCESS: find history by user id",
			IsSuccess:      true,
			UserID:         "uuid-uuid-uuid",
			ExpectedResult: expectedHistories,
			MockSetup: func(mockRepo *mockPh.IPasswordHistoriesRepository) {
				mockRepo.On(
					"FindByUserID",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*int"),
				).Return(expectedHistories, nil)
			},
		},
		{
			Name:          "ERROR: error find history by user id",
			IsSuccess:     false,
			UserID:        "user-error",
			ExpectedError: "error find history by user id",
			MockSetup: func(mockRepo *mockPh.IPasswordHistoriesRepository) {
				mockRepo.On(
					"FindByUserID",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*int"),
				).Return(nil, errors.New("error find history by user id"))
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

			response, err := userService.FindByUserID(ctx, tc.UserID)
			if tc.IsSuccess {
				assert.NoError(t, err)
				require.NotEmpty(t, response)
			} else {
				require.Error(t, err)
				require.Empty(t, response)
				require.True(t, strings.Contains(err.Error(), tc.ExpectedError))
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
