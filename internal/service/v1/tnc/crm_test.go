package tnc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	tncModel "github.com/paper-indonesia/pivot-backoffice/internal/model/tnc"
	"github.com/paper-indonesia/pivot-backoffice/internal/service/v1/tnc"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateTNCVersion(t *testing.T) {
	req := &tncModel.CreateTNCVersionRequest{
		Version:     "1.0.0",
		Title:       "Terms",
		HTMLContent: "<p>terms</p>",
		CreatedBy:   "admin-1",
	}
	existing := &tncModel.TNC{UUID: "tnc-existing", Version: "1.0.0"}

	testCases := []struct {
		name         string
		setupTNCRepo func(*repoMocks.ITNCRepository)
		wantErr      bool
		wantErrMsg   string
	}{
		{
			name: "SUCCESS: create new tnc version",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetTNCVersionByVersion", mock.Anything, req.Version).Return(nil, nil)
				repo.On("CreateTNCVersion", mock.Anything, mock.AnythingOfType("*tnc.TNC")).Return(nil)
			},
		},
		{
			name: "ERROR: GetTNCVersionByVersion fails",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetTNCVersionByVersion", mock.Anything, req.Version).Return(nil, errors.New("db connection failed"))
			},
			wantErr:    true,
			wantErrMsg: "db connection failed",
		},
		{
			name: "ERROR: version already exists (pre-check)",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetTNCVersionByVersion", mock.Anything, req.Version).Return(existing, nil)
			},
			wantErr:    true,
			wantErrMsg: constant.ErrTNCVersionAlreadyExists.Error(),
		},
		{
			name: "ERROR: CreateTNCVersion duplicate entry (1062)",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetTNCVersionByVersion", mock.Anything, req.Version).Return(nil, nil)
				repo.On("CreateTNCVersion", mock.Anything, mock.AnythingOfType("*tnc.TNC")).
					Return(errors.New("Error 1062: Duplicate entry '1.0.0' for key 'version'"))
			},
			wantErr:    true,
			wantErrMsg: constant.ErrTNCVersionAlreadyExists.Error(),
		},
		{
			name: "ERROR: CreateTNCVersion fails",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetTNCVersionByVersion", mock.Anything, req.Version).Return(nil, nil)
				repo.On("CreateTNCVersion", mock.Anything, mock.AnythingOfType("*tnc.TNC")).Return(errors.New("insert failed"))
			},
			wantErr:    true,
			wantErrMsg: "insert failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := repoMocks.NewITNCRepository(t)
			mockMerchantRepo := repoMocks.NewIMerchantRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.setupTNCRepo(mockRepo)

			svc := tnc.New(mockRepo, mockMerchantRepo, mockLogger)

			result, err := svc.CreateTNCVersion(context.Background(), req)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tc.wantErrMsg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, req.Version, result.Version)
				assert.Equal(t, req.Title, result.Title)
				assert.Equal(t, req.HTMLContent, result.HTMLContent)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestActivateTNCVersion(t *testing.T) {
	current := &tncModel.TNC{UUID: "tnc-1", Version: "1.0.0", IsActive: false}

	testCases := []struct {
		name         string
		id           string
		setupTNCRepo func(*repoMocks.ITNCRepository)
		wantErr      bool
		wantErrMsg   string
	}{
		{
			name: "SUCCESS: activate tnc version",
			id:   "tnc-1",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetTNCVersionByID", mock.Anything, "tnc-1").Return(current, nil)
				repo.On("DeactivateAllTNCVersions", mock.Anything).Return(nil)
				repo.On("UpdateTNCVersion", mock.Anything, mock.AnythingOfType("*tnc.TNC")).Return(nil)
			},
		},
		{
			name: "ERROR: GetTNCVersionByID fails",
			id:   "tnc-1",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetTNCVersionByID", mock.Anything, "tnc-1").Return(nil, errors.New("db down"))
			},
			wantErr:    true,
			wantErrMsg: "db down",
		},
		{
			name: "ERROR: version not found",
			id:   "missing",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetTNCVersionByID", mock.Anything, "missing").Return(nil, nil)
			},
			wantErr:    true,
			wantErrMsg: constant.ErrTNCVersionNotFound.Error(),
		},
		{
			name: "ERROR: DeactivateAllTNCVersions fails",
			id:   "tnc-1",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetTNCVersionByID", mock.Anything, "tnc-1").Return(current, nil)
				repo.On("DeactivateAllTNCVersions", mock.Anything).Return(errors.New("deactivate failed"))
			},
			wantErr:    true,
			wantErrMsg: "deactivate failed",
		},
		{
			name: "ERROR: UpdateTNCVersion fails",
			id:   "tnc-1",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetTNCVersionByID", mock.Anything, "tnc-1").Return(current, nil)
				repo.On("DeactivateAllTNCVersions", mock.Anything).Return(nil)
				repo.On("UpdateTNCVersion", mock.Anything, mock.AnythingOfType("*tnc.TNC")).Return(errors.New("update failed"))
			},
			wantErr:    true,
			wantErrMsg: "update failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := repoMocks.NewITNCRepository(t)
			mockMerchantRepo := repoMocks.NewIMerchantRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.setupTNCRepo(mockRepo)

			svc := tnc.New(mockRepo, mockMerchantRepo, mockLogger)

			result, err := svc.ActivateTNCVersion(context.Background(), tc.id)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tc.wantErrMsg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.True(t, result.IsActive)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeactivateTNCVersion(t *testing.T) {
	current := &tncModel.TNC{UUID: "tnc-1", Version: "1.0.0", IsActive: true}

	testCases := []struct {
		name         string
		id           string
		setupTNCRepo func(*repoMocks.ITNCRepository)
		wantErr      bool
		wantErrMsg   string
	}{
		{
			name: "SUCCESS: deactivate tnc version",
			id:   "tnc-1",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetTNCVersionByID", mock.Anything, "tnc-1").Return(current, nil)
				repo.On("UpdateTNCVersion", mock.Anything, mock.AnythingOfType("*tnc.TNC")).Return(nil)
			},
		},
		{
			name: "ERROR: GetTNCVersionByID fails",
			id:   "tnc-1",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetTNCVersionByID", mock.Anything, "tnc-1").Return(nil, errors.New("db down"))
			},
			wantErr:    true,
			wantErrMsg: "db down",
		},
		{
			name: "ERROR: version not found",
			id:   "missing",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetTNCVersionByID", mock.Anything, "missing").Return(nil, nil)
			},
			wantErr:    true,
			wantErrMsg: constant.ErrTNCVersionNotFound.Error(),
		},
		{
			name: "ERROR: UpdateTNCVersion fails",
			id:   "tnc-1",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetTNCVersionByID", mock.Anything, "tnc-1").Return(current, nil)
				repo.On("UpdateTNCVersion", mock.Anything, mock.AnythingOfType("*tnc.TNC")).Return(errors.New("update failed"))
			},
			wantErr:    true,
			wantErrMsg: "update failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := repoMocks.NewITNCRepository(t)
			mockMerchantRepo := repoMocks.NewIMerchantRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.setupTNCRepo(mockRepo)

			svc := tnc.New(mockRepo, mockMerchantRepo, mockLogger)

			result, err := svc.DeactivateTNCVersion(context.Background(), tc.id)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tc.wantErrMsg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.False(t, result.IsActive)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestListTNCVersions(t *testing.T) {
	list := []*tncModel.TNC{
		{UUID: "tnc-1", Version: "1.0.0", Title: "Terms"},
		{UUID: "tnc-2", Version: "2.0.0", Title: "Terms v2", IsActive: true},
	}

	testCases := []struct {
		name         string
		query        *tncModel.TNCVersionQuery
		setupTNCRepo func(*repoMocks.ITNCRepository)
		wantErr      bool
		wantErrMsg   string
	}{
		{
			name:  "SUCCESS: returns paginated tnc versions",
			query: &tncModel.TNCVersionQuery{Page: 1, PageSize: 10},
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("ListTNCVersions", mock.Anything, mock.AnythingOfType("*tnc.TNCVersionQuery")).Return(list, 2, nil)
			},
		},
		{
			name:  "SUCCESS: empty result",
			query: &tncModel.TNCVersionQuery{Page: 1, PageSize: 10},
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("ListTNCVersions", mock.Anything, mock.AnythingOfType("*tnc.TNCVersionQuery")).Return([]*tncModel.TNC{}, 0, nil)
			},
		},
		{
			name:  "ERROR: ListTNCVersions fails",
			query: &tncModel.TNCVersionQuery{Page: 1, PageSize: 10},
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("ListTNCVersions", mock.Anything, mock.AnythingOfType("*tnc.TNCVersionQuery")).Return(nil, 0, errors.New("query failed"))
			},
			wantErr:    true,
			wantErrMsg: constant.ErrGetTNCVersionList.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := repoMocks.NewITNCRepository(t)
			mockMerchantRepo := repoMocks.NewIMerchantRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.setupTNCRepo(mockRepo)

			svc := tnc.New(mockRepo, mockMerchantRepo, mockLogger)

			result, err := svc.ListTNCVersions(context.Background(), tc.query)

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

func TestGetTNCVersion(t *testing.T) {
	version := &tncModel.TNC{UUID: "tnc-1", Version: "1.0.0", Title: "Terms", IsActive: true}

	testCases := []struct {
		name         string
		id           string
		setupTNCRepo func(*repoMocks.ITNCRepository)
		wantErr      bool
		wantErrMsg   string
	}{
		{
			name: "SUCCESS: returns tnc version",
			id:   "tnc-1",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetTNCVersionByID", mock.Anything, "tnc-1").Return(version, nil)
			},
		},
		{
			name: "ERROR: GetTNCVersionByID fails",
			id:   "tnc-1",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetTNCVersionByID", mock.Anything, "tnc-1").Return(nil, errors.New("db down"))
			},
			wantErr:    true,
			wantErrMsg: "db down",
		},
		{
			name: "ERROR: version not found",
			id:   "missing",
			setupTNCRepo: func(repo *repoMocks.ITNCRepository) {
				repo.On("GetTNCVersionByID", mock.Anything, "missing").Return(nil, nil)
			},
			wantErr:    true,
			wantErrMsg: constant.ErrTNCVersionNotFound.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := repoMocks.NewITNCRepository(t)
			mockMerchantRepo := repoMocks.NewIMerchantRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.setupTNCRepo(mockRepo)

			svc := tnc.New(mockRepo, mockMerchantRepo, mockLogger)

			result, err := svc.GetTNCVersion(context.Background(), tc.id)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tc.wantErrMsg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tc.id, result.ID)
				assert.Equal(t, "1.0.0", result.Version)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
