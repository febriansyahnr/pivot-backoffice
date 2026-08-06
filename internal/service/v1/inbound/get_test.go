package inboundService

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	inboundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/inbound"
	mocksRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetList(t *testing.T) {
	filter := &inboundModel.GetInboundFilterRequest{
		Page:    1,
		PerPage: 10,
	}

	expectedResponse := &commonModel.PaginationResponse{
		Data: []inboundModel.InboundResponse{},
		Meta: commonModel.Meta{
			Page:       1,
			PerPage:    10,
			TotalItems: 0,
			TotalPages: 0,
		},
	}

	testCases := []struct {
		name           string
		setup          func(repo *mocksRepo.IInboundRepository)
		expectedResult *commonModel.PaginationResponse
		wantErr        bool
		expectedError  string
	}{
		{
			name: "SUCCESS: get list",
			setup: func(repo *mocksRepo.IInboundRepository) {
				repo.On("GetList", mock.Anything, filter).Return(expectedResponse, nil)
			},
			expectedResult: expectedResponse,
			wantErr:        false,
		},
		{
			name: "ERROR: repository returns error",
			setup: func(repo *mocksRepo.IInboundRepository) {
				repo.On("GetList", mock.Anything, filter).Return(nil, errors.New("database error"))
			},
			expectedResult: nil,
			wantErr:        true,
			expectedError:  response.HttpErrDatabase,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocksRepo.NewIInboundRepository(t)
			log, _ := logger.NewZapLogger(logger.Config{})

			tc.setup(repo)
			svc := New(nil, log, repo)

			result, err := svc.GetList(context.Background(), filter)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, result)
				require.True(t, strings.Contains(err.Error(), tc.expectedError), "expected error to contain %q, got %q", tc.expectedError, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestGetByID(t *testing.T) {
	id := "01923456-7890-7abc-def0-123456789abc"

	expectedResponse := &inboundModel.InboundResponse{
		ID:                id,
		IP:                "127.0.0.1",
		Method:            "POST",
		URL:               "/open-api/v2/payments",
		StatusCode:        200,
		ResponseTimeMs:    150.5,
		SnapCompatibility: false,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	testCases := []struct {
		name           string
		setup          func(repo *mocksRepo.IInboundRepository)
		expectedResult *inboundModel.InboundResponse
		wantErr        bool
		expectedError  string
	}{
		{
			name: "SUCCESS: get by id",
			setup: func(repo *mocksRepo.IInboundRepository) {
				repo.On("GetByID", mock.Anything, id).Return(expectedResponse, nil)
			},
			expectedResult: expectedResponse,
			wantErr:        false,
		},
		{
			name: "ERROR: repository returns error",
			setup: func(repo *mocksRepo.IInboundRepository) {
				repo.On("GetByID", mock.Anything, id).Return(nil, errors.New("database error"))
			},
			expectedResult: nil,
			wantErr:        true,
			expectedError:  response.HttpErrDatabase,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocksRepo.NewIInboundRepository(t)
			log, _ := logger.NewZapLogger(logger.Config{})

			tc.setup(repo)
			svc := New(nil, log, repo)

			result, err := svc.GetByID(context.Background(), id)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, result)
				require.True(t, strings.Contains(err.Error(), tc.expectedError), "expected error to contain %q, got %q", tc.expectedError, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestGetSnapVersionByID(t *testing.T) {
	id := "01923456-7890-7abc-def0-123456789abc"

	snapCompatibleResponse := &inboundModel.InboundResponse{
		ID:                id,
		IP:                "127.0.0.1",
		Method:            "POST",
		URL:               "/open-api/v2/payments",
		StatusCode:        200,
		ResponseTimeMs:    150.5,
		SnapCompatibility: true,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
		Client: &inboundModel.Client{
			Feature:     "AUTH",
			TraceId:     "trace-123",
			OriginId:    "origin-123",
			ReferenceId: "ref-123",
		},
		Feature: "AUTH",
		Headers: types.JSONText(`{"Content-Type":["application/json"]}`),
		Body: types.NullJSONText{
			JSONText: []byte(`{"grantType":"client_credentials"}`),
			Valid:    true,
		},
		ResponseBody: types.NullJSONText{
			JSONText: []byte(`{"data":{"accessToken":"token","expiresIn":"3600","tokenType":"Bearer"}}`),
			Valid:    true,
		},
	}

	nonSnapCompatibleResponse := &inboundModel.InboundResponse{
		ID:                id,
		SnapCompatibility: false,
	}

	testCases := []struct {
		name           string
		setup          func(repo *mocksRepo.IInboundRepository)
		expectedResult *inboundModel.InboundSnapVersionResponse
		wantErr        bool
		expectedError  string
	}{
		{
			name: "SUCCESS: get snap version by id",
			setup: func(repo *mocksRepo.IInboundRepository) {
				repo.On("GetByID", mock.Anything, id).Return(snapCompatibleResponse, nil)
			},
			expectedResult: snapCompatibleResponse.ToSnapVersionResponse(),
			wantErr:        false,
		},
		{
			name: "ERROR: repository returns error",
			setup: func(repo *mocksRepo.IInboundRepository) {
				repo.On("GetByID", mock.Anything, id).Return(nil, errors.New("database error"))
			},
			wantErr:       true,
			expectedError: response.HttpErrDatabase,
		},
		{
			name: "ERROR: inbound not found (nil detail)",
			setup: func(repo *mocksRepo.IInboundRepository) {
				repo.On("GetByID", mock.Anything, id).Return(nil, nil)
			},
			wantErr:       true,
			expectedError: response.HttpErrNotFound,
		},
		{
			name: "ERROR: not snap compatible",
			setup: func(repo *mocksRepo.IInboundRepository) {
				repo.On("GetByID", mock.Anything, id).Return(nonSnapCompatibleResponse, nil)
			},
			wantErr:       true,
			expectedError: response.HttpErrUnprocessableContent,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocksRepo.NewIInboundRepository(t)
			log, _ := logger.NewZapLogger(logger.Config{})

			tc.setup(repo)
			svc := New(nil, log, repo)

			result, err := svc.GetSnapVersionByID(context.Background(), id)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, result)
				require.True(t, strings.Contains(err.Error(), tc.expectedError), "expected error to contain %q, got %q", tc.expectedError, err.Error())
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tc.expectedResult.ID, result.ID)
				require.Equal(t, tc.expectedResult.SnapCompatibility, result.SnapCompatibility)
			}
		})
	}
}

func TestGetSnapVersionByIDErrorWrapping(t *testing.T) {
	t.Run("error contains original database error wrapped with HttpErrDatabase", func(t *testing.T) {
		repo := mocksRepo.NewIInboundRepository(t)
		log, _ := logger.NewZapLogger(logger.Config{})

		dbErr := errors.New("connection refused")
		repo.On("GetByID", mock.Anything, "id").Return(nil, dbErr)

		svc := New(nil, log, repo)
		_, err := svc.GetSnapVersionByID(context.Background(), "id")

		require.Error(t, err)
		// Verify that pkgErr.New wraps with format "errType | originalErr"
		wrappedErr, ok := err.(interface{ Unwrap() error })
		require.True(t, ok, "error should support Unwrap")
		require.ErrorIs(t, wrappedErr.Unwrap(), dbErr)
		require.True(t, strings.Contains(err.Error(), response.HttpErrDatabase))
	})

	_ = pkgErr.New // ensure pkgErr is referenced
}
