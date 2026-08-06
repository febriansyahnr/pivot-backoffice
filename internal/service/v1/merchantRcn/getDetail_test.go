package merchantRcn

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchantRcn"
	encryptionMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/encryption"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetRcnDetail(t *testing.T) {
	merchantID := uuid.New()
	rcnID := uuid.New()
	encryptedCardNumber := []byte("encrypted-card-number")
	encodedCardNumber := base64.StdEncoding.EncodeToString(encryptedCardNumber)
	decryptedCardNumber := "1234567890123456"

	validMerchantRcn := &merchantRcn.MerchantRcn{
		ID:                rcnID,
		MerchantID:        merchantID,
		PrincipalIssuer:   "CIMB_NIAGA",
		RealCardNumber:    encodedCardNumber,
		EncryptKMSVersion: "1",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	testCases := []struct {
		name      string
		rcnID     string
		merchantID string
		mockSetup func(
			repo *repoMocks.IMerchantRcnRepository,
			gcs *encryptionMocks.GCSClient,
		)
		wantErr bool
		assertFn func(t *testing.T, result *merchantRcn.MerchantRcnDetail, err error)
	}{
		{
			name:       "Success - Valid RCN decryption",
			rcnID:      rcnID.String(),
			merchantID: merchantID.String(),
			mockSetup: func(repo *repoMocks.IMerchantRcnRepository, gcs *encryptionMocks.GCSClient) {
				repo.On("FindByIDAndMerchantID", mock.Anything, rcnID.String(), merchantID.String()).
					Return(validMerchantRcn, nil).Once()

				gcs.On("DecryptSymmetric", mock.Anything, encryptedCardNumber).
					Return(decryptedCardNumber, nil).Once()
			},
			wantErr: false,
			assertFn: func(t *testing.T, result *merchantRcn.MerchantRcnDetail, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, rcnID, result.ID)
				assert.Equal(t, merchantID, result.MerchantID)
				assert.Equal(t, "CIMB_NIAGA", result.PrincipalIssuer)
				assert.Equal(t, decryptedCardNumber, result.CardNumber)
				assert.NotZero(t, result.CreatedAt)
				assert.NotZero(t, result.UpdatedAt)
			},
		},
		{
			name:       "Failure - Repository error",
			rcnID:      rcnID.String(),
			merchantID: merchantID.String(),
			mockSetup: func(repo *repoMocks.IMerchantRcnRepository, gcs *encryptionMocks.GCSClient) {
				repo.On("FindByIDAndMerchantID", mock.Anything, rcnID.String(), merchantID.String()).
					Return(nil, errors.New("database error")).Once()
			},
			wantErr: true,
			assertFn: func(t *testing.T, result *merchantRcn.MerchantRcnDetail, err error) {
				require.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "database error")
			},
		},
		{
			name:       "Failure - Merchant RCN not found",
			rcnID:      rcnID.String(),
			merchantID: merchantID.String(),
			mockSetup: func(repo *repoMocks.IMerchantRcnRepository, gcs *encryptionMocks.GCSClient) {
				repo.On("FindByIDAndMerchantID", mock.Anything, rcnID.String(), merchantID.String()).
					Return(nil, nil).Once()
			},
			wantErr: true,
			assertFn: func(t *testing.T, result *merchantRcn.MerchantRcnDetail, err error) {
				require.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), constant.ErrMerchantRcnNotFound.Error())
			},
		},
		{
			name:       "Failure - Invalid base64 encoding",
			rcnID:      rcnID.String(),
			merchantID: merchantID.String(),
			mockSetup: func(repo *repoMocks.IMerchantRcnRepository, gcs *encryptionMocks.GCSClient) {
				invalidRcn := &merchantRcn.MerchantRcn{
					ID:                rcnID,
					MerchantID:        merchantID,
					PrincipalIssuer:   "CIMB_NIAGA",
					RealCardNumber:    "invalid-base64!!!",
					EncryptKMSVersion: "1",
					CreatedAt:         time.Now(),
					UpdatedAt:         time.Now(),
				}
				repo.On("FindByIDAndMerchantID", mock.Anything, rcnID.String(), merchantID.String()).
					Return(invalidRcn, nil).Once()
			},
			wantErr: true,
			assertFn: func(t *testing.T, result *merchantRcn.MerchantRcnDetail, err error) {
				require.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), constant.ErrDecodeMerchantRcn.Error())
			},
		},
		{
			name:       "Failure - Decryption error",
			rcnID:      rcnID.String(),
			merchantID: merchantID.String(),
			mockSetup: func(repo *repoMocks.IMerchantRcnRepository, gcs *encryptionMocks.GCSClient) {
				repo.On("FindByIDAndMerchantID", mock.Anything, rcnID.String(), merchantID.String()).
					Return(validMerchantRcn, nil).Once()

				gcs.On("DecryptSymmetric", mock.Anything, encryptedCardNumber).
					Return("", errors.New("decryption failed")).Once()
			},
			wantErr: true,
			assertFn: func(t *testing.T, result *merchantRcn.MerchantRcnDetail, err error) {
				require.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), constant.ErrDecryptMerchantRcn.Error())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockRepo := repoMocks.NewIMerchantRcnRepository(t)
			mockCimbProcessor := repoMocks.NewICimbProcessorRepository(t)
			mockGCS := encryptionMocks.NewGCSClient(t)

			tc.mockSetup(mockRepo, mockGCS)

			// Create service
			svc := New(mockRepo, mockCimbProcessor, mockGCS, mockLogger)

			// Execute
			result, err := svc.GetRcnDetail(context.Background(), tc.rcnID, tc.merchantID)

			// Assert
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tc.assertFn != nil {
				tc.assertFn(t, result, err)
			}
		})
	}
}
