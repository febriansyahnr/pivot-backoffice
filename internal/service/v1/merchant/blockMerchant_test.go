package merchant

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMerchantService_BlockMerchant(t *testing.T) {
	merchantID := uuid.NewString()

	merchant := &merchantModel.Merchant{
		UUID:     merchantID,
		Name:     "Test Merchant",
		ParentID: sql.NullString{Valid: false}, // Parent merchant
		Status:   constant.MerchantStatusActive,
	}

	testCases := []struct {
		name         string
		merchantID   string
		setupMocks   func(*repositoryMocks.IMerchantRepository, *repositoryMocks.ISnapCoreRepository, redismock.ClientMock)
		expectError  bool
		expectResult bool
	}{
		{
			name:       "SUCCESS: Block merchant successfully with cache deletion",
			merchantID: merchantID,
			setupMocks: func(mockRepo *repositoryMocks.IMerchantRepository, mockSnapCore *repositoryMocks.ISnapCoreRepository, redisMock redismock.ClientMock) {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).
					Return(merchant, nil)
				mockRepo.On("GetSubMerchantsByParentID", mock.Anything, merchantID).
					Return([]*merchantModel.Merchant{}, nil)
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).
					Return(merchant, nil).Maybe()
				mockRepo.On("Update", mock.Anything, mock.Anything).
					Return(nil)
				mockSnapCore.On("BlockVirtualAccount", mock.Anything, &snapCoreModel.BlockVirtualAccountRequest{
					MerchantID: merchantID,
				}).Return(nil, nil)

				// Mock Redis cache deletion (uses subMerchant.UUID, which is merchantID in this case)
				cacheKey := fmt.Sprintf(constant.MerchantStatusByIDCacheKey, merchantID)
				redisMock.ExpectDel(cacheKey).SetVal(1)
			},
			expectError:  false,
			expectResult: true,
		},
		{
			name:       "SUCCESS: Block merchant with sub-merchants and cache deletion",
			merchantID: merchantID,
			setupMocks: func(mockRepo *repositoryMocks.IMerchantRepository, mockSnapCore *repositoryMocks.ISnapCoreRepository, redisMock redismock.ClientMock) {
				subMerchantID := uuid.NewString()
				subMerchant := &merchantModel.Merchant{
					UUID:     subMerchantID,
					Name:     "Sub Merchant",
					ParentID: sql.NullString{Valid: true, String: merchantID},
					Status:   constant.MerchantStatusActive,
				}

				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).
					Return(merchant, nil)
				mockRepo.On("GetSubMerchantsByParentID", mock.Anything, merchantID).
					Return([]*merchantModel.Merchant{subMerchant}, nil)
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).
					Return(merchant, nil).Maybe()
				mockRepo.On("FindMerchantByID", mock.Anything, subMerchantID).
					Return(subMerchant, nil).Maybe()
				mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(m *merchantModel.Merchant) bool {
					return m.UUID == merchantID || m.UUID == subMerchantID
				})).Return(nil)

				mockSnapCore.On("BlockVirtualAccount", mock.Anything, mock.MatchedBy(func(req *snapCoreModel.BlockVirtualAccountRequest) bool {
					return req.MerchantID == merchantID || req.MerchantID == subMerchantID
				})).Return(nil, nil)

				// Mock Redis cache deletion for both parent and sub-merchant
				// First call for parent merchant (subMerchant.UUID = merchantID)
				parentCacheKey := fmt.Sprintf(constant.MerchantStatusByIDCacheKey, merchantID)
				redisMock.ExpectDel(parentCacheKey).SetVal(1)

				// Second call for sub-merchant (subMerchant.UUID = subMerchantID)
				subCacheKey := fmt.Sprintf(constant.MerchantStatusByIDCacheKey, subMerchantID)
				redisMock.ExpectDel(subCacheKey).SetVal(1)
			},
			expectError:  false,
			expectResult: true,
		},
		{
			name:       "ERROR: Redis cache deletion fails but continues processing",
			merchantID: merchantID,
			setupMocks: func(mockRepo *repositoryMocks.IMerchantRepository, mockSnapCore *repositoryMocks.ISnapCoreRepository, redisMock redismock.ClientMock) {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).
					Return(merchant, nil)
				mockRepo.On("GetSubMerchantsByParentID", mock.Anything, merchantID).
					Return([]*merchantModel.Merchant{}, nil)
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).
					Return(merchant, nil).Maybe()
				mockRepo.On("Update", mock.Anything, mock.Anything).
					Return(nil)
				mockSnapCore.On("BlockVirtualAccount", mock.Anything, mock.Anything).
					Return(nil, nil).Maybe()

				// Mock Redis cache deletion failure (uses subMerchant.UUID, which is merchantID in this case)
				cacheKey := fmt.Sprintf(constant.MerchantStatusByIDCacheKey, merchantID)
				redisMock.ExpectDel(cacheKey).SetErr(errors.New("redis error"))
			},
			expectError:  false, // Should continue processing despite cache deletion failure
			expectResult: false, // But result should be empty due to skip logic
		},
		{
			name:       "ERROR: Merchant not found",
			merchantID: "invalid-id",
			setupMocks: func(mockRepo *repositoryMocks.IMerchantRepository, mockSnapCore *repositoryMocks.ISnapCoreRepository, redisMock redismock.ClientMock) {
				mockRepo.On("FindMerchantByID", mock.Anything, "invalid-id").
					Return(nil, nil)
			},
			expectError:  true,
			expectResult: false,
		},
		{
			name:       "ERROR: Database error",
			merchantID: merchantID,
			setupMocks: func(mockRepo *repositoryMocks.IMerchantRepository, mockSnapCore *repositoryMocks.ISnapCoreRepository, redisMock redismock.ClientMock) {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).
					Return(nil, errors.New("database error"))
			},
			expectError:  true,
			expectResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := repositoryMocks.NewIMerchantRepository(t)
			mockSnapCore := repositoryMocks.NewISnapCoreRepository(t)
			redisClient, redisMock := redismock.NewClientMock()
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.setupMocks(mockRepo, mockSnapCore, redisMock)

			// Create service instance
			service := &MerchantService{
				repo:         mockRepo,
				snapCoreRepo: mockSnapCore,
				redis:        redisExt.WrapRedisClient(redisClient, nil),
				logger:       mockLogger,
			}

			// Execute test
			ctx := context.Background()
			result, err := service.BlockMerchant(ctx, tc.merchantID)

			// Assertions
			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)

				if tc.expectResult {
					assert.Equal(t, tc.merchantID, result.MerchantId)
					assert.Equal(t, "Test Merchant", result.MerchantName)
					assert.NotZero(t, result.BlockedAt)
				}
			}

			// Verify all mock expectations
			mockRepo.AssertExpectations(t)
			mockSnapCore.AssertExpectations(t)
			assert.NoError(t, redisMock.ExpectationsWereMet())
		})
	}
}

func TestMerchantService_UnblockMerchant(t *testing.T) {
	merchantID := uuid.NewString()

	merchant := &merchantModel.Merchant{
		UUID:     merchantID,
		Name:     "Test Merchant",
		ParentID: sql.NullString{Valid: false},
		Status:   constant.MerchantStatusBlocked,
	}

	testCases := []struct {
		name        string
		merchantID  string
		setupMocks  func(*repositoryMocks.IMerchantRepository, *repositoryMocks.ISnapCoreRepository, *repositoryMocks.IMerchantTopUpRepository, redismock.ClientMock)
		expectError bool
	}{
		{
			name:       "SUCCESS: Unblock parent merchant without sub-merchants and active top-up",
			merchantID: merchantID,
			setupMocks: func(mockRepo *repositoryMocks.IMerchantRepository, mockSnapCore *repositoryMocks.ISnapCoreRepository, mockTopUp *repositoryMocks.IMerchantTopUpRepository, redisMock redismock.ClientMock) {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
				mockRepo.On("GetSubMerchantsByParentID", mock.Anything, merchantID).Return([]*merchantModel.Merchant{}, nil)
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
				redisMock.ExpectDel(fmt.Sprintf(constant.MerchantStatusByIDCacheKey, merchantID)).SetVal(1)
				mockTopUp.On("CountActiveMerchantTopUpReferences", mock.Anything, mock.Anything).Return(1, nil)
				mockSnapCore.On("UnblockVirtualAccount", mock.Anything, &snapCoreModel.UnblockVirtualAccountRequest{
					MerchantID: merchantID,
				}).Return(nil, nil)
			},
			expectError: false,
		},
		{
			name:       "SUCCESS: Unblock merchant with sub-merchants",
			merchantID: merchantID,
			setupMocks: func(mockRepo *repositoryMocks.IMerchantRepository, mockSnapCore *repositoryMocks.ISnapCoreRepository, mockTopUp *repositoryMocks.IMerchantTopUpRepository, redisMock redismock.ClientMock) {
				subID := uuid.NewString()
				subMerchant := &merchantModel.Merchant{
					UUID:     subID,
					Name:     "Sub Merchant",
					ParentID: sql.NullString{Valid: true, String: merchantID},
					Status:   constant.MerchantStatusBlocked,
				}

				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
				mockRepo.On("GetSubMerchantsByParentID", mock.Anything, merchantID).Return([]*merchantModel.Merchant{subMerchant}, nil)
				mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(m *merchantModel.Merchant) bool {
					return m.UUID == merchantID || m.UUID == subID
				})).Return(nil)
				redisMock.ExpectDel(fmt.Sprintf(constant.MerchantStatusByIDCacheKey, merchantID)).SetVal(1)
				redisMock.ExpectDel(fmt.Sprintf(constant.MerchantStatusByIDCacheKey, subID)).SetVal(1)
				mockTopUp.On("CountActiveMerchantTopUpReferences", mock.Anything, mock.Anything).Return(1, nil)
				mockSnapCore.On("UnblockVirtualAccount", mock.Anything, mock.MatchedBy(func(req *snapCoreModel.UnblockVirtualAccountRequest) bool {
					return req.MerchantID == merchantID || req.MerchantID == subID
				})).Return(nil, nil)
			},
			expectError: false,
		},
		{
			name:       "SUCCESS: No UnblockVirtualAccount when zero active top-up references",
			merchantID: merchantID,
			setupMocks: func(mockRepo *repositoryMocks.IMerchantRepository, mockSnapCore *repositoryMocks.ISnapCoreRepository, mockTopUp *repositoryMocks.IMerchantTopUpRepository, redisMock redismock.ClientMock) {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
				mockRepo.On("GetSubMerchantsByParentID", mock.Anything, merchantID).Return([]*merchantModel.Merchant{}, nil)
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
				redisMock.ExpectDel(fmt.Sprintf(constant.MerchantStatusByIDCacheKey, merchantID)).SetVal(1)
				mockTopUp.On("CountActiveMerchantTopUpReferences", mock.Anything, mock.Anything).Return(0, nil)
			},
			expectError: false,
		},
		{
			name:       "ERROR: Merchant not found",
			merchantID: "invalid-id",
			setupMocks: func(mockRepo *repositoryMocks.IMerchantRepository, mockSnapCore *repositoryMocks.ISnapCoreRepository, mockTopUp *repositoryMocks.IMerchantTopUpRepository, redisMock redismock.ClientMock) {
				mockRepo.On("FindMerchantByID", mock.Anything, "invalid-id").Return(nil, nil)
			},
			expectError: true,
		},
		{
			name:       "ERROR: FindMerchantByID database error",
			merchantID: merchantID,
			setupMocks: func(mockRepo *repositoryMocks.IMerchantRepository, mockSnapCore *repositoryMocks.ISnapCoreRepository, mockTopUp *repositoryMocks.IMerchantTopUpRepository, redisMock redismock.ClientMock) {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(nil, errors.New("database error"))
			},
			expectError: true,
		},
		{
			name:       "ERROR: GetSubMerchantsByParentID database error",
			merchantID: merchantID,
			setupMocks: func(mockRepo *repositoryMocks.IMerchantRepository, mockSnapCore *repositoryMocks.ISnapCoreRepository, mockTopUp *repositoryMocks.IMerchantTopUpRepository, redisMock redismock.ClientMock) {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
				mockRepo.On("GetSubMerchantsByParentID", mock.Anything, merchantID).Return(nil, errors.New("database error"))
			},
			expectError: true,
		},
		{
			name:       "SUCCESS: Continue when Update fails",
			merchantID: merchantID,
			setupMocks: func(mockRepo *repositoryMocks.IMerchantRepository, mockSnapCore *repositoryMocks.ISnapCoreRepository, mockTopUp *repositoryMocks.IMerchantTopUpRepository, redisMock redismock.ClientMock) {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
				mockRepo.On("GetSubMerchantsByParentID", mock.Anything, merchantID).Return([]*merchantModel.Merchant{}, nil)
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update error"))
			},
			expectError: false,
		},
		{
			name:       "SUCCESS: Continue when Redis delete fails",
			merchantID: merchantID,
			setupMocks: func(mockRepo *repositoryMocks.IMerchantRepository, mockSnapCore *repositoryMocks.ISnapCoreRepository, mockTopUp *repositoryMocks.IMerchantTopUpRepository, redisMock redismock.ClientMock) {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
				mockRepo.On("GetSubMerchantsByParentID", mock.Anything, merchantID).Return([]*merchantModel.Merchant{}, nil)
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
				redisMock.ExpectDel(fmt.Sprintf(constant.MerchantStatusByIDCacheKey, merchantID)).SetErr(errors.New("redis error"))
			},
			expectError: false,
		},
		{
			name:       "SUCCESS: Continue when UnblockVirtualAccount fails",
			merchantID: merchantID,
			setupMocks: func(mockRepo *repositoryMocks.IMerchantRepository, mockSnapCore *repositoryMocks.ISnapCoreRepository, mockTopUp *repositoryMocks.IMerchantTopUpRepository, redisMock redismock.ClientMock) {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
				mockRepo.On("GetSubMerchantsByParentID", mock.Anything, merchantID).Return([]*merchantModel.Merchant{}, nil)
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
				redisMock.ExpectDel(fmt.Sprintf(constant.MerchantStatusByIDCacheKey, merchantID)).SetVal(1)
				mockTopUp.On("CountActiveMerchantTopUpReferences", mock.Anything, mock.Anything).Return(1, nil)
				mockSnapCore.On("UnblockVirtualAccount", mock.Anything, &snapCoreModel.UnblockVirtualAccountRequest{
					MerchantID: merchantID,
				}).Return(nil, errors.New("snap core error"))
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := repositoryMocks.NewIMerchantRepository(t)
			mockSnapCore := repositoryMocks.NewISnapCoreRepository(t)
			mockTopUp := repositoryMocks.NewIMerchantTopUpRepository(t)
			redisClient, redisMock := redismock.NewClientMock()
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.setupMocks(mockRepo, mockSnapCore, mockTopUp, redisMock)

			service := &MerchantService{
				repo:              mockRepo,
				snapCoreRepo:      mockSnapCore,
				merchantTopUpRepo: mockTopUp,
				redis:             redisExt.WrapRedisClient(redisClient, nil),
				logger:            mockLogger,
			}

			result, err := service.UnblockMerchant(context.Background(), tc.merchantID)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if result != nil {
					assert.Equal(t, merchantID, result.UnblockedMerchantDetails.MerchantId)
					assert.Equal(t, "Test Merchant", result.UnblockedMerchantDetails.MerchantName)
					assert.NotZero(t, result.UnblockedMerchantDetails.UnblockedAt)
				}
			}

			mockRepo.AssertExpectations(t)
			mockSnapCore.AssertExpectations(t)
			mockTopUp.AssertExpectations(t)
			assert.NoError(t, redisMock.ExpectationsWereMet())
		})
	}
}
