package qris_test

import (
	"context"
	"fmt"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/qris"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistrationList(t *testing.T) {

	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})

	repo := repoMocks.NewIQrisRepository(t)

	service := New(logger, repo, nil, nil)

	traceId := uuid.NewString()
	ctx := context.WithValue(context.Background(), pdkConst.CtxTraceIdKey, traceId)

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    string
		wantResult []qris.RegistrationListResp
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				repo.On(
					"RegistrationList", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: fmt.Sprintf(c.InternalErrorFmt, traceId),
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				repo.On(
					"RegistrationList", c.ValueCtxMockType(), c.StringMockType(),
				).Return([]qris.RegistrationListResp{{Id: "ID", ExternalId: "EX", MerchantType: "MT"}}, nil)
			},
			wantResult: []qris.RegistrationListResp{{Id: "ID", ExternalId: "EX", MerchantType: "MT"}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			if resp, err := service.RegistrationList(ctx, uuid.NewString()); test.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, resp)
			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
