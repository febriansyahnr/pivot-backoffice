package qris_test

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/qris"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindQrRegistrationByExternalID(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	tests := []struct {
		name       string
		setupMock  func(repo *repoMocks.IQrisRepository)
		wantErr    string
		wantResult string
	}{
		{
			name: "ERROR:FindQrRegistrationByExternalID error",
			setupMock: func(qrisRepo *repoMocks.IQrisRepository) {
				qrisRepo.On(
					"FindQrRegistrationByExternalID", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR:FindQrRegistrationByExternalID data not found",
			setupMock: func(qrisRepo *repoMocks.IQrisRepository) {
				qrisRepo.On(
					"FindQrRegistrationByExternalID", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: c.ErrDataNotFound.Error(),
		},
		{
			name: "SUCCESS:FindQrRegistrationByExternalID",
			setupMock: func(qrisRepo *repoMocks.IQrisRepository) {
				qrisRepo.On(
					"FindQrRegistrationByExternalID", c.ValueCtxMockType(), c.StringMockType(),
				).Return(&qris.Registration{Status: c.QrRegistrationStatusSuccess}, nil)
			},
			wantErr: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := repoMocks.NewIQrisRepository(t)
			service := New(logger, repo, nil, nil)

			test.setupMock(repo)

			if _, err := service.FindQrRegistrationByExternalID(context.Background(), util.GenerateULID()); test.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestFindQrRegistrationByExternalIDAndAcquirer(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	tests := []struct {
		name       string
		setupMock  func(repo *repoMocks.IQrisRepository)
		wantErr    string
		wantResult string
	}{
		{
			name: "ERROR:FindQrRegistrationByExternalIDAndAcquirer error",
			setupMock: func(qrisRepo *repoMocks.IQrisRepository) {
				qrisRepo.On(
					"FindQrRegistrationByExternalIDAndAcquirer", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR:FindQrRegistrationByExternalIDAndAcquirer data not found",
			setupMock: func(qrisRepo *repoMocks.IQrisRepository) {
				qrisRepo.On(
					"FindQrRegistrationByExternalIDAndAcquirer", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: c.ErrDataNotFound.Error(),
		},
		{
			name: "ERROR:QR registration status is not success",
			setupMock: func(qrisRepo *repoMocks.IQrisRepository) {
				qrisRepo.On(
					"FindQrRegistrationByExternalIDAndAcquirer", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&qris.Registration{Status: "IN_PROGRESS"}, nil)
			},
			wantErr: c.ErrQrRegistrationIsNotCompleted.Error(),
		},
		{
			name: "SUCCESS:FindQrRegistrationByExternalIDAndAcquirer",
			setupMock: func(qrisRepo *repoMocks.IQrisRepository) {
				qrisRepo.On(
					"FindQrRegistrationByExternalIDAndAcquirer", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Return(&qris.Registration{Status: c.QrRegistrationStatusSuccess}, nil)
			},
			wantErr: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := repoMocks.NewIQrisRepository(t)
			service := New(logger, repo, nil, nil)

			test.setupMock(repo)

			if _, err := service.FindQrRegistrationByExternalIDAndAcquirer(context.Background(), "external-id", "acquirer"); test.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
