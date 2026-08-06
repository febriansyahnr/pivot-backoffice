package xbPayoutService_test

import (
	"context"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/xbPayout"
)

func TestGetFxRate(t *testing.T) {
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	xbCoreProcessorRepo := repositoryMock.NewIXbCoreProcessorRepository(t)

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func()
	}{
		{
			name:    "ERROR: XbCore GetFxRate",
			wantErr: true,
			setupMock: func() {
				xbCoreProcessorRepo.On("GetFxRate",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.GetFxRateRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "SUCCESS",
			wantErr: false,
			setupMock: func() {
				xbCoreProcessorRepo.On("GetFxRate",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbCoreProcessorModel.GetFxRateRequest"),
				).Return(&xbCoreProcessorModel.GetFxRateResponseData{}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			svc := New(log, nil, nil, xbCoreProcessorRepo)
			_, err := svc.GetFxRate(context.Background(), &xbModel.GetFxRateRequest{})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
