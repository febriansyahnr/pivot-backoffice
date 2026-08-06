package merchant

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMerchantServiceUpdateStatusByID(t *testing.T) {
	redisClient, redisClientMock := redismock.NewClientMock()

	testCases := []struct {
		name         string
		status       string
		reasonStatus string
		id           string
		mockSetup    func(mockRepo *mocks.IMerchantRepository, redisMock redismock.ClientMock)
		wantErr      bool
	}{
		{
			name:         "SUCCESS: update merchant status",
			status:       "ACTIVE",
			reasonStatus: "Approved by admin",
			id:           "merchant-123",
			mockSetup: func(mockRepo *mocks.IMerchantRepository, redisMock redismock.ClientMock) {
				mockRepo.On("UpdateStatusByID", mock.Anything, "ACTIVE", "Approved by admin", "merchant-123").Return(nil)
				cacheKey := fmt.Sprintf(constant.MerchantStatusByIDCacheKey, "merchant-123")
				redisMock.ExpectDel(cacheKey).SetVal(1)
			},
			wantErr: false,
		},
		{
			name:         "SUCCESS: update merchant status with empty reason",
			status:       "INACTIVE",
			reasonStatus: "",
			id:           "merchant-456",
			mockSetup: func(mockRepo *mocks.IMerchantRepository, redisMock redismock.ClientMock) {
				mockRepo.On("UpdateStatusByID", mock.Anything, "INACTIVE", "", "merchant-456").Return(nil)
				cacheKey := fmt.Sprintf(constant.MerchantStatusByIDCacheKey, "merchant-456")
				redisMock.ExpectDel(cacheKey).SetVal(1)
			},
			wantErr: false,
		},
		{
			name:         "ERROR: repository update fails",
			status:       "ACTIVE",
			reasonStatus: "Approved by admin",
			id:           "merchant-789",
			mockSetup: func(mockRepo *mocks.IMerchantRepository, redisMock redismock.ClientMock) {
				mockRepo.On("UpdateStatusByID", mock.Anything, "ACTIVE", "Approved by admin", "merchant-789").Return(errors.New("database error"))
			},
			wantErr: true,
		},
		{
			name:         "SUCCESS: update status with suspended status",
			status:       "SUSPENDED",
			reasonStatus: "Violation of terms",
			id:           "merchant-suspended",
			mockSetup: func(mockRepo *mocks.IMerchantRepository, redisMock redismock.ClientMock) {
				mockRepo.On("UpdateStatusByID", mock.Anything, "SUSPENDED", "Violation of terms", "merchant-suspended").Return(nil)
				cacheKey := fmt.Sprintf(constant.MerchantStatusByIDCacheKey, "merchant-suspended")
				redisMock.ExpectDel(cacheKey).SetVal(1)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewIMerchantRepository(t)
			mockLogger, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			tc.mockSetup(mockRepo, redisClientMock)

			svc := New(mockRepo, mockLogger, nil, nil, nil, nil,
				WithRedisClient(redisExt.WrapRedisClient(redisClient, nil)),
			)

			err := svc.UpdateStatusByID(context.Background(), tc.status, tc.reasonStatus, tc.id)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
