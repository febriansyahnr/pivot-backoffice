package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	tncModel "github.com/paper-indonesia/pivot-backoffice/internal/model/tnc"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/repository/user"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"

	"github.com/go-redis/redismock/v9"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidateInvitationToken(t *testing.T) {
	invitationData := &userModel.ValidateInvitationResponse{
		UserID:       "user-123",
		UserName:     "John Doe",
		Email:        "mail@mail.com",
		MerchantName: "Test Merchant",
		MerchantID:   "merchant-123",
	}
	invitationDataJSON, _ := json.Marshal(invitationData)

	signedStatus := &tncModel.TNCSigningStatus{IsSigned: true, ActiveVersion: "1.2.0", SignedVersion: "1.2.0"}

	testCases := []struct {
		name         string
		token        string
		mocksSetup   func(r redismock.ClientMock, tokenKey string)
		tncMockSetup func(*mockService.ITNCService)
		wantErr      bool
		wantTNCMeta  bool
	}{
		{
			name:  "SUCCESS: attaches merchant tnc status",
			token: "valid-token",
			mocksSetup: func(r redismock.ClientMock, tokenKey string) {
				r.ExpectGet(tokenKey).SetVal(string(invitationDataJSON))
			},
			tncMockSetup: func(svc *mockService.ITNCService) {
				svc.On("GetTNCSigningStatus", mock.Anything, invitationData.MerchantID).Return(signedStatus, nil)
			},
			wantErr:     false,
			wantTNCMeta: true,
		},
		{
			name:  "SUCCESS: tnc service returns nil status",
			token: "valid-token-no-tnc",
			mocksSetup: func(r redismock.ClientMock, tokenKey string) {
				r.ExpectGet(tokenKey).SetVal(string(invitationDataJSON))
			},
			tncMockSetup: func(svc *mockService.ITNCService) {
				svc.On("GetTNCSigningStatus", mock.Anything, invitationData.MerchantID).Return(nil, nil)
			},
			wantErr:     false,
			wantTNCMeta: false,
		},
		{
			name:  "FAILURE - Key does not exist",
			token: "nonexistent-token",
			mocksSetup: func(r redismock.ClientMock, tokenKey string) {
				r.ExpectGet(tokenKey).SetErr(redis.Nil)
			},
			tncMockSetup: func(svc *mockService.ITNCService) {},
			wantErr:      true,
		},
		{
			name:  "FAILURE - Redis error",
			token: "error-token",
			mocksSetup: func(r redismock.ClientMock, tokenKey string) {
				r.ExpectGet(tokenKey).SetErr(redis.ErrClosed)
			},
			tncMockSetup: func(svc *mockService.ITNCService) {},
			wantErr:      true,
		},
		{
			name:  "FAILURE - Invalid JSON data",
			token: "invalid-json-token",
			mocksSetup: func(r redismock.ClientMock, tokenKey string) {
				r.ExpectGet(tokenKey).SetVal("invalid-json")
			},
			tncMockSetup: func(svc *mockService.ITNCService) {},
			wantErr:      true,
		},
		{
			name:  "FAILURE - TNC service error",
			token: "valid-token-tnc-err",
			mocksSetup: func(r redismock.ClientMock, tokenKey string) {
				r.ExpectGet(tokenKey).SetVal(string(invitationDataJSON))
			},
			tncMockSetup: func(svc *mockService.ITNCService) {
				svc.On("GetTNCSigningStatus", mock.Anything, invitationData.MerchantID).
					Return(nil, errors.New("tnc service unavailable"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				ServiceName: "testing",
			}

			secret := &config.Secret{
				JWTSignatureKey: config.JWTSignatureKey{
					UserKey: "testing",
				},
			}

			userRepo := mockUser.NewIUserRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			db, redisClientMock := redismock.NewClientMock()
			redisMock := redisExt.WrapRedisClient(db, nil)

			tncSvc := mockService.NewITNCService(t)

			tokenKey := fmt.Sprintf("backend-portal:users:user-invitation:token:%s", tc.token)
			tc.mocksSetup(redisClientMock, tokenKey)
			tc.tncMockSetup(tncSvc)

			trxSvc := New(
				cfg, secret, loggerMock, userRepo, nil,
				WithRedisClient(redisMock), WithTNCService(tncSvc),
			)

			ctx := context.Background()
			result, err := trxSvc.ValidateInvitationToken(ctx, tc.token)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, invitationData.UserID, result.UserID)
				assert.Equal(t, invitationData.UserName, result.UserName)
				assert.Equal(t, invitationData.Email, result.Email)
				assert.Equal(t, invitationData.MerchantName, result.MerchantName)
				assert.Equal(t, invitationData.MerchantID, result.MerchantID)
				if tc.wantTNCMeta {
					assert.NotNil(t, result.TNCMetadata)
					assert.Equal(t, signedStatus, result.TNCMetadata)
				} else {
					assert.Nil(t, result.TNCMetadata)
				}
			}

			userRepo.AssertExpectations(t)
			tncSvc.AssertExpectations(t)
		})
	}
}
