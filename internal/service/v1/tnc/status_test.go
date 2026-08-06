package tnc_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	tncModel "github.com/paper-indonesia/pivot-backoffice/internal/model/tnc"
	"github.com/paper-indonesia/pivot-backoffice/internal/service/v1/tnc"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetTNCSigningStatus(t *testing.T) {
	signedAt := time.Now().UTC()
	active := &tncModel.TNC{UUID: "tnc-uuid", Version: "1.2.0", IsActive: true, MarkdownContent: "Test TNC Content"}
	latest := &tncModel.MerchantTNCSigningHistory{
		UUID:          "hist-1",
		Version:       "1.2.0",
		SignedBy:      "user-1",
		SignedByEmail: "user@merchant.com",
		SignedAt:      signedAt,
		DocumentPath:  "tnc-documents/merchant-1/1.2.0/hist-1.pdf",
	}

	validMerchant := &merchantModel.Merchant{
		UUID:     "merchant-1",
		Name:     "Test Merchant",
		ParentID: sql.NullString{Valid: false, String: ""},
	}

	subMerchant := &merchantModel.Merchant{
		UUID:     "submerchant-1",
		Name:     "Test Submerchant",
		ParentID: sql.NullString{Valid: true, String: "parent-1"},
	}

	testCases := []struct {
		name          string
		merchantID    string
		setupTNCRepo  func(*repoMocks.ITNCRepository)
		setupMerchant func(*repoMocks.IMerchantRepository)
		wantErr       bool
		wantErrMsg    string
		wantSigned    bool
		wantNilStatus bool
		assertStatus  func(*testing.T, *tncModel.TNCSigningStatus)
	}{
		{
			name:       "SUCCESS: merchant has signed active version",
			merchantID: "merchant-1",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetActiveTNCVersion", mock.Anything).Return(active, nil)
				repo.On("GetLatestSigningByMerchant", mock.Anything, "merchant-1").Return(latest, nil)
			},
			setupMerchant: func(merchantRepo *repoMocks.IMerchantRepository) {
				merchantRepo.On("FindMerchantByID", mock.Anything, "merchant-1").Return(validMerchant, nil)
			},
			wantSigned: true,
			assertStatus: func(t *testing.T, s *tncModel.TNCSigningStatus) {
				assert.Equal(t, "1.2.0", s.ActiveVersion)
				assert.Equal(t, "1.2.0", s.SignedVersion)
				assert.Equal(t, "user-1", s.SignedBy)
				assert.Equal(t, "user@merchant.com", s.SignedByEmail)
				assert.Equal(t, signedAt, s.SignedAt)
				// With no GCS wired in the test setup, signedDocumentURL returns "".
				assert.Empty(t, s.DocumentURL)
				// When signed, markdown content should not be included
				assert.Empty(t, s.MarkdownContent)
			},
		},
		{
			name:       "SUCCESS: signed version differs from active (not signed)",
			merchantID: "merchant-1",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetActiveTNCVersion", mock.Anything).Return(active, nil)
				stale := &tncModel.MerchantTNCSigningHistory{UUID: "hist-old", Version: "1.0.0", SignedBy: "user-1", SignedAt: signedAt}
				repo.On("GetLatestSigningByMerchant", mock.Anything, "merchant-1").Return(stale, nil)
			},
			setupMerchant: func(merchantRepo *repoMocks.IMerchantRepository) {
				merchantRepo.On("FindMerchantByID", mock.Anything, "merchant-1").Return(validMerchant, nil)
			},
			wantSigned: false,
			assertStatus: func(t *testing.T, s *tncModel.TNCSigningStatus) {
				assert.Equal(t, "1.2.0", s.ActiveVersion)
				assert.Equal(t, "1.0.0", s.SignedVersion)
				// When not signed, markdown content should be included
				assert.Equal(t, "Test TNC Content", s.MarkdownContent)
			},
		},
		{
			name:       "SUCCESS: active version exists but merchant never signed",
			merchantID: "merchant-1",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetActiveTNCVersion", mock.Anything).Return(active, nil)
				repo.On("GetLatestSigningByMerchant", mock.Anything, "merchant-1").Return(nil, nil)
			},
			setupMerchant: func(merchantRepo *repoMocks.IMerchantRepository) {
				merchantRepo.On("FindMerchantByID", mock.Anything, "merchant-1").Return(validMerchant, nil)
			},
			wantSigned: false,
			assertStatus: func(t *testing.T, s *tncModel.TNCSigningStatus) {
				// When there's no signing history, ActiveVersion won't be set
				assert.Empty(t, s.ActiveVersion)
				assert.Empty(t, s.SignedVersion)
				assert.Empty(t, s.SignedBy)
				// When not signed, markdown content should be included
				assert.Equal(t, "Test TNC Content", s.MarkdownContent)
			},
		},
		{
			name:       "SUCCESS: no active version and no signing history",
			merchantID: "merchant-1",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetActiveTNCVersion", mock.Anything).Return(nil, nil)
			},
			setupMerchant: func(merchantRepo *repoMocks.IMerchantRepository) {
				// No merchant repo call expected since function returns early when no active version
			},
			wantNilStatus: true,
		},
		{
			name:       "SUCCESS: merchant is a submerchant - returns nil",
			merchantID: "submerchant-1",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetActiveTNCVersion", mock.Anything).Return(active, nil)
			},
			setupMerchant: func(merchantRepo *repoMocks.IMerchantRepository) {
				merchantRepo.On("FindMerchantByID", mock.Anything, "submerchant-1").Return(subMerchant, nil)
			},
			wantNilStatus: true,
		},
		{
			name:       "ERROR: GetActiveTNCVersion fails",
			merchantID: "merchant-1",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetActiveTNCVersion", mock.Anything).Return(nil, errors.New("db down"))
			},
			setupMerchant: func(merchantRepo *repoMocks.IMerchantRepository) {},
			wantErr:       true,
			wantErrMsg:    "db down",
		},
		{
			name:       "ERROR: FindMerchantByID fails",
			merchantID: "merchant-1",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetActiveTNCVersion", mock.Anything).Return(active, nil)
			},
			setupMerchant: func(merchantRepo *repoMocks.IMerchantRepository) {
				merchantRepo.On("FindMerchantByID", mock.Anything, "merchant-1").Return(nil, errors.New("merchant lookup failed"))
			},
			wantErr:    true,
			wantErrMsg: "merchant lookup failed",
		},
		{
			name:       "ERROR: merchant not found",
			merchantID: "nonexistent-merchant",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetActiveTNCVersion", mock.Anything).Return(active, nil)
			},
			setupMerchant: func(merchantRepo *repoMocks.IMerchantRepository) {
				merchantRepo.On("FindMerchantByID", mock.Anything, "nonexistent-merchant").Return(nil, nil)
			},
			wantErr:    true,
			wantErrMsg: constant.ErrMerchantNotFound.Error(),
		},
		{
			name:       "ERROR: GetLatestSigningByMerchant fails",
			merchantID: "merchant-1",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetActiveTNCVersion", mock.Anything).Return(active, nil)
				repo.On("GetLatestSigningByMerchant", mock.Anything, "merchant-1").Return(nil, errors.New("lookup failed"))
			},
			setupMerchant: func(merchantRepo *repoMocks.IMerchantRepository) {
				merchantRepo.On("FindMerchantByID", mock.Anything, "merchant-1").Return(validMerchant, nil)
			},
			wantErr:    true,
			wantErrMsg: "lookup failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := repoMocks.NewITNCRepository(t)
			mockMerchantRepo := repoMocks.NewIMerchantRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.setupTNCRepo(mockRepo)
			tc.setupMerchant(mockMerchantRepo)

			svc := tnc.New(mockRepo, mockMerchantRepo, mockLogger)

			status, err := svc.GetTNCSigningStatus(context.Background(), tc.merchantID)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, status)
				assert.Contains(t, err.Error(), tc.wantErrMsg)
			} else {
				require.NoError(t, err)
				if tc.wantNilStatus {
					assert.Nil(t, status)
				} else {
					require.NotNil(t, status)
					assert.Equal(t, tc.wantSigned, status.IsSigned)
					if tc.assertStatus != nil {
						tc.assertStatus(t, status)
					}
				}
			}

			mockRepo.AssertExpectations(t)
			mockMerchantRepo.AssertExpectations(t)
		})
	}
}

func TestGetSigningHistory(t *testing.T) {
	list := []*tncModel.MerchantTNCSigningHistory{
		{UUID: "h1", MerchantID: "merchant-1", Version: "1.2.0"},
		{UUID: "h2", MerchantID: "merchant-1", Version: "1.0.0"},
	}

	testCases := []struct {
		name         string
		query        *tncModel.SigningHistoryQuery
		setupTNCRepo func(*repoMocks.ITNCRepository)
		wantErr      bool
		wantErrMsg   string
	}{
		{
			name:  "SUCCESS: returns paginated signing history",
			query: &tncModel.SigningHistoryQuery{MerchantID: "merchant-1", Page: 1, PageSize: 10},
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("ListSigningHistories", mock.Anything, mock.AnythingOfType("*tnc.SigningHistoryQuery")).Return(list, 2, nil)
			},
		},
		{
			name:  "SUCCESS: empty result",
			query: &tncModel.SigningHistoryQuery{MerchantID: "merchant-1", Page: 1, PageSize: 10},
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("ListSigningHistories", mock.Anything, mock.AnythingOfType("*tnc.SigningHistoryQuery")).Return([]*tncModel.MerchantTNCSigningHistory{}, 0, nil)
			},
		},
		{
			name:         "ERROR: empty merchant id",
			query:        &tncModel.SigningHistoryQuery{MerchantID: "", Page: 1, PageSize: 10},
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {},
			wantErr:      true,
			wantErrMsg:   constant.ErrInvalidMerchantID.Error(),
		},
		{
			name:  "ERROR: ListSigningHistories fails",
			query: &tncModel.SigningHistoryQuery{MerchantID: "merchant-1", Page: 1, PageSize: 10},
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("ListSigningHistories", mock.Anything, mock.AnythingOfType("*tnc.SigningHistoryQuery")).Return(nil, 0, errors.New("query failed"))
			},
			wantErr:    true,
			wantErrMsg: constant.ErrGetTNCSigningHistory.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := repoMocks.NewITNCRepository(t)
			mockMerchantRepo := repoMocks.NewIMerchantRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.setupTNCRepo(mockRepo)

			svc := tnc.New(mockRepo, mockMerchantRepo, mockLogger)

			result, err := svc.GetSigningHistory(context.Background(), tc.query)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tc.wantErrMsg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tc.query.Page, result.Meta.Page)
				assert.Equal(t, tc.query.PageSize, result.Meta.PerPage)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
