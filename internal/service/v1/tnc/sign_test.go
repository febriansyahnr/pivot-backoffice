package tnc_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	tncModel "github.com/paper-indonesia/pivot-backoffice/internal/model/tnc"
	"github.com/paper-indonesia/pivot-backoffice/internal/service/v1/tnc"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSignTNC(t *testing.T) {
	req := &tncModel.SignTNCRequest{
		MerchantID: "merchant-1",
		SignedBy:   "user-1",
		Email:      "user@merchant.com",
		IPAddress:  "203.0.113.10",
		UserAgent:  "Mozilla/5.0",
	}
	active := &tncModel.TNC{
		UUID:            "tnc-uuid",
		Version:         "1.2.0",
		Title:           "Terms",
		MarkdownContent: "# Terms and Conditions\n\nThese are the terms.",
		IsActive:        true,
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
		name              string
		request           *tncModel.SignTNCRequest
		setupTNCRepo      func(*repoMocks.ITNCRepository)
		setupMerchantRepo func(*repoMocks.IMerchantRepository)
		wantErr           bool
		wantErrMsg        string
		assertResponse    func(*testing.T, *tncModel.MerchantTNCSigningHistoryResponse)
	}{
		{
			name:    "SUCCESS: sign active tnc",
			request: req,
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetActiveTNCVersion", mock.Anything).Return(active, nil)
				repo.On("GetSigningByMerchantAndVersion", mock.Anything, req.MerchantID, active.Version).Return(nil, nil)
				repo.On("InsertSigningHistory", mock.Anything, mock.AnythingOfType("*tnc.MerchantTNCSigningHistory")).Return(nil)
			},
			setupMerchantRepo: func(repo *repoMocks.IMerchantRepository) {
				repo.On("FindMerchantByID", mock.Anything, req.MerchantID).Return(validMerchant, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*merchant.Merchant")).Return(nil)
			},
			assertResponse: func(t *testing.T, resp *tncModel.MerchantTNCSigningHistoryResponse) {
				assert.Equal(t, req.MerchantID, resp.MerchantID)
				assert.Equal(t, active.Version, resp.Version)
				assert.Equal(t, req.SignedBy, resp.SignedBy)
				assert.Equal(t, req.Email, resp.SignedByEmail)
				// With no GCS in test, DocumentURL will be empty
				assert.Empty(t, resp.DocumentURL)
			},
		},
		{
			name:    "ERROR: GetActiveTNCVersion fails",
			request: req,
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetActiveTNCVersion", mock.Anything).Return(nil, errors.New("db connection failed"))
			},
			setupMerchantRepo: func(repo *repoMocks.IMerchantRepository) {},
			wantErr:           true,
			wantErrMsg:        "db connection failed",
		},
		{
			name:    "ERROR: no active tnc version",
			request: req,
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetActiveTNCVersion", mock.Anything).Return(nil, nil)
			},
			setupMerchantRepo: func(repo *repoMocks.IMerchantRepository) {},
			wantErr:           true,
			wantErrMsg:        constant.ErrNoActiveTNCVersion.Error(),
		},
		{
			name:    "ERROR: GetSigningByMerchantAndVersion fails",
			request: req,
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetActiveTNCVersion", mock.Anything).Return(active, nil)
				repo.On("GetSigningByMerchantAndVersion", mock.Anything, req.MerchantID, active.Version).Return(nil, errors.New("lookup failed"))
			},
			setupMerchantRepo: func(repo *repoMocks.IMerchantRepository) {},
			wantErr:           true,
			wantErrMsg:        "lookup failed",
		},
		{
			name:    "ERROR: merchant already signed active version",
			request: req,
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetActiveTNCVersion", mock.Anything).Return(active, nil)
				repo.On("GetSigningByMerchantAndVersion", mock.Anything, req.MerchantID, active.Version).Return(&tncModel.MerchantTNCSigningHistory{Version: active.Version}, nil)
			},
			setupMerchantRepo: func(repo *repoMocks.IMerchantRepository) {},
			wantErr:           true,
			wantErrMsg:        constant.ErrMerchantAlreadySignedTNC.Error(),
		},
		{
			name:    "ERROR: FindMerchantByID returns error",
			request: req,
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetActiveTNCVersion", mock.Anything).Return(active, nil)
				repo.On("GetSigningByMerchantAndVersion", mock.Anything, req.MerchantID, active.Version).Return(nil, nil)
			},
			setupMerchantRepo: func(repo *repoMocks.IMerchantRepository) {
				repo.On("FindMerchantByID", mock.Anything, req.MerchantID).Return(nil, errors.New("merchant lookup failed"))
			},
			wantErr:    true,
			wantErrMsg: "merchant lookup failed",
		},
		{
			name:    "ERROR: merchant not found (nil)",
			request: req,
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetActiveTNCVersion", mock.Anything).Return(active, nil)
				repo.On("GetSigningByMerchantAndVersion", mock.Anything, req.MerchantID, active.Version).Return(nil, nil)
			},
			setupMerchantRepo: func(repo *repoMocks.IMerchantRepository) {
				repo.On("FindMerchantByID", mock.Anything, req.MerchantID).Return(nil, nil)
			},
			wantErr:    true,
			wantErrMsg: constant.ErrInvalidMerchantID.Error(),
		},
		{
			name: "ERROR: merchant is a submerchant cannot sign",
			request: func() *tncModel.SignTNCRequest {
				subReq := *req
				subReq.MerchantID = "submerchant-1"
				return &subReq
			}(),
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetActiveTNCVersion", mock.Anything).Return(active, nil)
				repo.On("GetSigningByMerchantAndVersion", mock.Anything, "submerchant-1", active.Version).Return(nil, nil)
			},
			setupMerchantRepo: func(repo *repoMocks.IMerchantRepository) {
				repo.On("FindMerchantByID", mock.Anything, "submerchant-1").Return(subMerchant, nil)
			},
			wantErr:    true,
			wantErrMsg: constant.ErrSubmerchantCannotSignTNC.Error(),
		},
		{
			name:    "ERROR: InsertSigningHistory fails",
			request: req,
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetActiveTNCVersion", mock.Anything).Return(active, nil)
				repo.On("GetSigningByMerchantAndVersion", mock.Anything, req.MerchantID, active.Version).Return(nil, nil)
				repo.On("InsertSigningHistory", mock.Anything, mock.AnythingOfType("*tnc.MerchantTNCSigningHistory")).Return(errors.New("insert failed"))
			},
			setupMerchantRepo: func(repo *repoMocks.IMerchantRepository) {
				repo.On("FindMerchantByID", mock.Anything, req.MerchantID).Return(validMerchant, nil)
			},
			wantErr:    true,
			wantErrMsg: "insert failed",
		},
		{
			name:    "ERROR: merchant update fails",
			request: req,
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetActiveTNCVersion", mock.Anything).Return(active, nil)
				repo.On("GetSigningByMerchantAndVersion", mock.Anything, req.MerchantID, active.Version).Return(nil, nil)
				repo.On("InsertSigningHistory", mock.Anything, mock.AnythingOfType("*tnc.MerchantTNCSigningHistory")).Return(nil)
			},
			setupMerchantRepo: func(repo *repoMocks.IMerchantRepository) {
				repo.On("FindMerchantByID", mock.Anything, req.MerchantID).Return(validMerchant, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*merchant.Merchant")).Return(errors.New("merchant update failed"))
			},
			wantErr:    true,
			wantErrMsg: "merchant update failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := repoMocks.NewITNCRepository(t)
			mockMerchantRepo := repoMocks.NewIMerchantRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.setupTNCRepo(mockRepo)
			tc.setupMerchantRepo(mockMerchantRepo)

			svc := tnc.New(mockRepo, mockMerchantRepo, mockLogger)

			result, err := svc.SignTNC(context.Background(), tc.request)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tc.wantErrMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tc.assertResponse != nil {
					tc.assertResponse(t, result)
				}
			}

			mockRepo.AssertExpectations(t)
			mockMerchantRepo.AssertExpectations(t)
		})
	}
}
